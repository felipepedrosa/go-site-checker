package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var httpClient = &http.Client{
	Timeout: defaultTimeout * time.Second,
}

type StatusResponse struct {
	URL    string
	Status int
}

type FlagsUsed struct {
	Timeout   bool
	Addresses bool
	JsonFile  bool
}

const (
	addressFlag    = "addresses"
	timeoutFlag    = "timeout"
	fileFlag       = "file"
	filePathFlag   = "filepath"
	defaultTimeout = 2
	minTableWidth  = 15
)

func main() {
	timeout := flag.Int(timeoutFlag, defaultTimeout, "HTTP request timeout in seconds")
	addresses := flag.String(addressFlag, "", `Addresses to be checked, if multiple use ';' as separator ("http://www.google.com.br;https://facebook.com")`)
	useFile := flag.Bool(fileFlag, false, "Set to true to read addresses from a JSON file")
	filePath := flag.String(filePathFlag, "addresses.json", "Path to the JSON file containing addresses (used with --file)")
	flag.Parse()

	flagsUsed := &FlagsUsed{}
	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "timeout must be greater than 0")
		os.Exit(1)
	}

	if *timeout != 2 {
		flagsUsed.Timeout = true
		configureHttpClientTimeout(*timeout)
	}

	if *addresses != "" {
		flagsUsed.Addresses = true
	}

	if *useFile {
		flagsUsed.JsonFile = true
	}

	checkAtLeastOneFlagUsed(flagsUsed)
	checkConflictingFlags(flagsUsed)

	var addressesSlice []string
	if flagsUsed.JsonFile {
		addressesSlice = getAddressesFromFile(*filePath)
	} else {
		addressesSlice = getAddressesFromCommandLine(*addresses)
	}

	statuses := checkAddresses(addressesSlice)
	validStatuses := filterProcessed(statuses)
	printStatuses(validStatuses)
}

func getAddressesFromFile(path string) []string {
	fileExtension := filepath.Ext(path)
	if fileExtension != ".json" {
		fmt.Fprintf(os.Stderr, "file %s is not a JSON file\n", path)
		os.Exit(1)
	}

	file, err := os.Open(path)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file %s: %v\n", path, err)
		os.Exit(1)
	}

	defer file.Close()

	var addresses []string
	reader := bufio.NewReader(file)
	err = json.NewDecoder(reader).Decode(&addresses)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error decoding JSON from file %s: %v\n", path, err)
		os.Exit(1)
	}

	if len(addresses) == 0 {
		fmt.Fprintf(os.Stderr, "no addresses found in file %s\n", path)
		os.Exit(1)
	}

	for i, addr := range addresses {
		addresses[i] = strings.TrimSpace(addr)
	}

	return addresses
}

func getAddressesFromCommandLine(addresses string) []string {
	addressesSlice := strings.Split(addresses, ";")

	if len(addressesSlice) >= 1 && addressesSlice[0] == "" {
		fmt.Fprintln(os.Stderr, `usage: app --addresses "address1;address2;..."`)
		os.Exit(1)
	}

	for i, addr := range addressesSlice {
		addressesSlice[i] = strings.TrimSpace(addr)
	}

	return addressesSlice
}

func checkAtLeastOneFlagUsed(flags *FlagsUsed) {
	if !flags.Addresses && !flags.JsonFile {
		fmt.Fprintln(os.Stderr, "either --addresses or --file flag must be provided")
		os.Exit(1)
	}
}

func checkConflictingFlags(flags *FlagsUsed) {
	if flags.Addresses && flags.JsonFile {
		fmt.Fprintln(os.Stderr, "cannot use --addresses and --file flags together")
		os.Exit(1)
	}
}

func configureHttpClientTimeout(timeout int) {
	httpClient.Timeout = time.Duration(timeout) * time.Second
}

func checkAddresses(addresses []string) []StatusResponse {
	statuses := make([]StatusResponse, len(addresses))

	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := range addresses {
		wg.Go(func() {
			resp, err := getUrlStatus(addresses[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", addresses[i], err)
				return
			}
			mu.Lock()
			statuses[i] = resp
			mu.Unlock()
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
		fmt.Println("no responses to show")
		return
	}

	maxAddrLen := minTableWidth
	for _, r := range responses {
		if l := len(r.URL); l > maxAddrLen {
			maxAddrLen = l
		}
	}

	fmt.Printf("%-*s | Status\n", maxAddrLen, "URL")
	fmt.Printf("%s-+--------\n", strings.Repeat("-", maxAddrLen))
	for _, r := range responses {
		fmt.Printf("%-*s | %d\n", maxAddrLen, r.URL, r.Status)
	}
	fmt.Printf("%s-+--------\n", strings.Repeat("-", maxAddrLen))
}

func isValidURL(addr string) bool {
	u, err := url.ParseRequestURI(addr)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
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
