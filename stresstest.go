package main

import (
	//"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	Execute()
}

type RunEFunc func(cmd *cobra.Command, args []string) error

var rootCmd = &cobra.Command{
	Use:   "stresstest",
	Short: "Teste de stress para uma URL",
	Long:  `Programa para realizar teste de stress em uma URL`,

	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		requests, _ := cmd.Flags().GetInt("requests")
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		StressTest(url, requests, concurrency)
	},
}

type Requests struct {
	requisicoes_feitas   int
	requisicoes_com_erro int
	mu                   sync.Mutex
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("url", "u", "", "URL para stress test")
	rootCmd.Flags().IntP("requests", "r", 100, "Número de requisições a serem enviadas")
	rootCmd.Flags().IntP("concurrency", "c", 10, "Número de requisições concorrentes")
	rootCmd.MarkFlagRequired("url")
	rootCmd.MarkFlagRequired("requests")
	rootCmd.MarkFlagRequired("concurrency")
}

func StressTest(url string, requests int, concurrency int) {
	// Record the start time
	log.Println("Iniciando teste de stress...")
	startTime := time.Now()

	// Create a channel to communicate with the workers
	jobs := make(chan int, requests)
	results := make(chan int, requests)
	controle := Requests{requisicoes_feitas: 0, requisicoes_com_erro: 0}

	// Create a wait group to wait for all the workers to finish
	var wg sync.WaitGroup

	// Workers de acordo com quantidade de concorrência
	for i := 0; i < concurrency; i++ {
		go worker(url, jobs, results, &wg, &controle)
	}

	wg.Add(requests)

	for i := 0; i < requests; i++ {
		jobs <- i
	}

	close(jobs)
	wg.Wait()

	// Tempo final do teste
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Dicionário status code e quantidade
	statusCodes := make(map[int]int)

	// Consumindo canal de resultados
	for range requests {
		statusCode := <-results
		statusCodes[statusCode]++
	}

	// Resultados
	log.Println("Resultados:")
	log.Printf("Tempo total: %s", duration)
	log.Printf("Total de requisições feitas: %d", controle.requisicoes_feitas)
	log.Printf("Total de requisições não concluídas: %d", controle.requisicoes_com_erro)
	for statusCode, count := range statusCodes {
		log.Printf("Número de requisições com status %d: %d", statusCode, count)
	}

	close(results)

}

func worker(url string, jobs <-chan int, results chan<- int, wg *sync.WaitGroup, controle *Requests) {

	// Criando cliente http
	client := &http.Client{}

	for range jobs {
		// Criando requisição
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		// Enviando requisição
		resp, err := client.Do(req)
		if err != nil {
			controle.mu.Lock()
			controle.requisicoes_com_erro++
			controle.mu.Unlock()
			results <- 0
		} else {
			controle.mu.Lock()
			controle.requisicoes_feitas++
			controle.mu.Unlock()
			results <- resp.StatusCode
			resp.Body.Close()
		}
		wg.Done()
	}
}
