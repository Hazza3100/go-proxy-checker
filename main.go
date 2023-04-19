package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	start := time.Now()
	inputFile, err := os.Open("proxies.txt")
	if err != nil {
		fmt.Println("Failed to open input file:", err)
		os.Exit(1)
	}
	defer inputFile.Close()
	outputFile, err := os.Create("valid.txt")
	if err != nil {
		fmt.Println("Failed to create output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()
	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(outputFile)

	// 12K max threads, change if u want i just dont feel the need to do more lol
	semaphore := make(chan struct{}, 120000)

	for scanner.Scan() {
		proxy := scanner.Text()

		wg.Add(1)
		go func(proxy string) {
			semaphore <- struct{}{}
			defer func() {
				wg.Done()
				<-semaphore
			}()

			if checkProxy(proxy) {
				_, err := writer.WriteString(proxy + "\n")
				if err != nil {
					fmt.Println("Failed to write to file:", err)
				} else {
					fmt.Println("Valid proxy:", proxy)
				}
			} else {
				fmt.Println("Invalid proxy:", proxy)
			}
		}(proxy)
	}

	wg.Wait()
	writer.Flush()
	fmt.Println("Finished checking proxies!")

	took := time.Since(start)
	fmt.Println("Took:", took)
}

func checkProxy(proxy string) bool {
	proxyURL := &url.URL{
		Host:   proxy,
		Scheme: "http",
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get("https://ipinfo.io/ip")
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}

	return false
}
