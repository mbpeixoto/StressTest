package main

import (
	//"fmt"
	"io"
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
	startTime := time.Now()

	// Create a channel to communicate with the workers
	jobs := make(chan int, requests)
	results := make(chan int, requests)

	// Create a wait group to wait for all the workers to finish
	var wg sync.WaitGroup

	// Create the workers
	for i := 0; i < concurrency; i++ {
		go worker(url, jobs, results, &wg, i)
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
	qtdRequests := 0
	for i := 0; i < requests; i++ {
		statusCode := <-results
		statusCodes[statusCode]++
		qtdRequests++
	}

	// Print the results
	log.Println("Resultados:")
	log.Printf("Tempo total: %s", duration)
	log.Printf("Total de requisições: %d", qtdRequests)
	for statusCode, count := range statusCodes {
		log.Printf("Número de requisições com status %d: %d", statusCode, count)
	}

	// Close the results channel
	close(results)

}

func worker(url string, jobs <-chan int, results chan<- int, wg *sync.WaitGroup, workeID int) {

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
	for i := range jobs {

		log.Printf("Worker %d enviando requisição %d", workeID, i)

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		// Read the response body
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Close the response body
		resp.Body.Close()

		// Send the result to the results channel
		results <- resp.StatusCode
		wg.Done()
	}
}
