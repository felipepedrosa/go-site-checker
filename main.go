package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var httpClient = &http.Client{
	Timeout: 2 * time.Second,
}

type StatusResponse struct {
	URL    string
	Status int
}

type FlagsUsed struct {
	Timeout   bool
	Addresses bool
}

var urlRegex = regexp.MustCompile(
	`^(https?|ftp)://` +
		`(([a-zA-Z0-9\-]+\.)+[a-zA-Z]{2,}` +
		`|localhost` +
		`|(\d{1,3}\.){3}\d{1,3})` +
		`(:\d+)?` +
		`(/[^\s]*)?$`,
)

func main() {
	timeout := flag.Int("timeout", 2, "HTTP request timeout in seconds")
	addresses := flag.String("addresses", "", `Addresses to be checked, if multiple use ';' as separator ("http://www.google.com.br;https://facebook.com")`)

	flag.Parse()

	flagsUsed := &FlagsUsed{}

	if timeout != nil && *timeout != 2 {
		flagsUsed.Timeout = true
		configureHttpClientTimeout(*timeout)
	}

	if *addresses != "" {
		flagsUsed.Addresses = true
	}

	addressesSlice := strings.Split(*addresses, ";")
	fmt.Printf("address slice len = %d\n", len(addressesSlice))

	if len(addressesSlice) >= 1 && addressesSlice[0] == "" {
		fmt.Fprintln(os.Stderr, `usage: app --addresses "address1;address2;..."`)
		os.Exit(1)
	}

	for i, addr := range addressesSlice {
		addressesSlice[i] = strings.TrimSpace(addr)
	}

	statuses := checkAddresses(addressesSlice)
	validStatuses := filterProcessed(statuses)
	printStatuses(validStatuses)
}

func configureHttpClientTimeout(timeout int) {
	httpClient.Timeout = time.Duration(timeout) * time.Second
}

func checkAddresses(addresses []string) []StatusResponse {
	statuses := make([]StatusResponse, len(addresses))

	var wg sync.WaitGroup
	for i, addr := range addresses {
		wg.Go(func() {
			resp, err := getUrlStatus(addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", addr, err)
				return
			}
			statuses[i] = resp
		})
	}
	wg.Wait()

	return statuses
}

func filterProcessed(responses []StatusResponse) []StatusResponse {
	result := make([]StatusResponse, 0, len(responses))
	for _, r := range responses {
		if r.URL != "" {
			result = append(result, r)
		}
	}
	return result
}

func printStatuses(responses []StatusResponse) {
	if len(responses) == 0 {
		fmt.Println("no responses to print")
		return
	}

	maxAddrLen := len("Address")
	for _, r := range responses {
		if l := len(r.URL); l > maxAddrLen {
			maxAddrLen = l
		}
	}

	fmt.Printf("%-*s | Status\n", maxAddrLen, "Address")
	fmt.Printf("%s-+--------\n", strings.Repeat("-", maxAddrLen))
	for _, r := range responses {
		fmt.Printf("%-*s | %d\n", maxAddrLen, r.URL, r.Status)
	}
}

func isValidURL(url string) bool {
	return urlRegex.MatchString(url)
}

func getUrlStatus(addr string) (StatusResponse, error) {
	if !isValidURL(addr) {
		return StatusResponse{}, fmt.Errorf("invalid URL: %s", addr)
	}

	resp, err := httpClient.Get(addr)
	if err != nil {
		return StatusResponse{}, err
	}
	defer resp.Body.Close()

	return StatusResponse{
		URL:    addr,
		Status: resp.StatusCode,
	}, nil
}
