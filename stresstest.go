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

	// Create the workers
	for i := 0; i < concurrency; i++ {
		go worker(url, jobs, results, &wg, &controle)
	}

	// Add the jobs to the jobs channel
	for i := 0; i < requests; i++ {
		wg.Add(1)
		jobs <- i
	}

	// Close the jobs channel
	close(jobs)

	// Wait for all the workers to finish
	wg.Wait()

	// Record the end time
	endTime := time.Now()

	// Calculate the duration
	duration := endTime.Sub(startTime)

	// Create a map to store the results
	statusCodes := make(map[int]int)

	// Loop over the results channel
	for range controle.requisicoes_feitas {
		statusCode := <-results
		statusCodes[statusCode]++
	}

	// Print the results
	log.Println("Resultados:")
	log.Printf("Tempo total: %s", duration)
	log.Printf("Total de requisições feitas: %d", controle.requisicoes_feitas)
	log.Printf("Total de requisições com erro: %d", controle.requisicoes_com_erro)
	for statusCode, count := range statusCodes {
		log.Printf("Número de requisições com status %d: %d", statusCode, count)
	}

	// Close the results channel
	close(results)

}

func worker(url string, jobs <-chan int, results chan<- int, wg *sync.WaitGroup, controle *Requests) {

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", "Test")

	// Loop over the jobs channel
	for range jobs {

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			controle.requisicoes_com_erro = controle.requisicoes_com_erro + 1
		} else {
			controle.requisicoes_feitas = controle.requisicoes_feitas + 1
			results <- resp.StatusCode
		}
		resp.Body.Close()
		wg.Done()
	}
}
