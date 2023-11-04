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

type ProxyChecker struct {
	Proxies      []string
	ValidProxies []string
}

func NewProxyChecker(proxies []string) *ProxyChecker {
	return &ProxyChecker{
		Proxies:      proxies,
		ValidProxies: make([]string, 0),
	}
}

func (pc *ProxyChecker) CheckProxy(proxy string, validChan chan<- string, httpClient *http.Client) {
	proxyURL, _ := url.Parse("http://" + proxy)
	httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	response, err := httpClient.Get("https://ipinfo.io/ip")
	if err == nil && response.StatusCode == 200 {
		validChan <- proxy
	} else {
		validChan <- ""
		fmt.Println("Invalid proxy:", proxy) // Log the invalid proxy
	}
}

func (pc *ProxyChecker) RunChecks() {
	var wg sync.WaitGroup
	validChan := make(chan string, len(pc.Proxies))
	httpClient := &http.Client{Timeout: 5 * time.Second}

	for _, proxy := range pc.Proxies {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			pc.CheckProxy(p, validChan, httpClient)
		}(proxy)
	}

	go func() {
		wg.Wait()
		close(validChan)
	}()

	for proxy := range validChan {
		if proxy != "" {
			pc.ValidProxies = append(pc.ValidProxies, proxy)
			fmt.Println("Valid proxy found:", proxy)
			file, err := os.OpenFile("valid.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if _, err := file.WriteString(proxy + "\n"); err != nil {
				fmt.Println(err)
			}
			file.Close()
		}
	}
}

func main() {
	file, err := os.Open("proxy.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	var proxies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxies = append(proxies, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return
	}

	proxyChecker := NewProxyChecker(proxies)
	proxyChecker.RunChecks()
}
