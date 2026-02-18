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
	Addresses string
	JSONFile  bool
	Export    bool
	FilePath  string
}

const (
	addressFlag    = "addresses"
	timeoutFlag    = "timeout"
	fileFlag       = "file"
	filePathFlag   = "filepath"
	exportFlag     = "export"
	defaultTimeout = 2
	minTableWidth  = 15
)

func main() {
	flagsUsed := configureFlags()
	addresses := getAddresses(flagsUsed)

	if flagsUsed.Export {
		exportToCSV(addresses)
	} else {
		printStatuses(addresses)
	}
}

func getAddresses(flags *FlagsUsed) []StatusResponse {
	var addressesSlice []string
	if flags.JSONFile {
		addressesSlice = getAddressesFromFile(flags.FilePath)
	} else {
		addressesSlice = getAddressesFromCommandLine(flags.Addresses)
	}

	statuses := checkAddresses(addressesSlice)
	return filterProcessed(statuses)
}

func configureFlags() *FlagsUsed {
	flagsUsed := &FlagsUsed{}
	timeout := flag.Int(timeoutFlag, defaultTimeout, "HTTP request timeout in seconds")
	addresses := flag.String(addressFlag, "", `Addresses to be checked, if multiple use ';' as separator ("http://www.google.com.br;https://facebook.com")`)
	useFile := flag.Bool(fileFlag, false, "Set to true to read addresses from a JSON file")
	filePath := flag.String(filePathFlag, "addresses.json", "Path to the JSON file containing addresses (used with --file)")
	export := flag.Bool(exportFlag, false, "Set to true to export results to a CSV file")
	flag.Parse()

	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "timeout must be greater than 0")
		os.Exit(1)
	}

	if *timeout != defaultTimeout {
		flagsUsed.Timeout = true
		configureHttpClientTimeout(*timeout)
	}

	if *addresses != "" {
		flagsUsed.Addresses = *addresses
	}

	if *useFile {
		flagsUsed.JSONFile = true
	}

	if *export {
		flagsUsed.Export = true
	}

	flagsUsed.FilePath = *filePath

	checkAtLeastOneFlagUsed(flagsUsed)
	checkConflictingFlags(flagsUsed)

	return flagsUsed
}

func exportToCSV(responses []StatusResponse) {
	file, err := os.Create("output.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating CSV file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString("URL;Status\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing to CSV file: %v\n", err)
		os.Exit(1)
	}

	for _, r := range responses {
		line := fmt.Sprintf("%s;%d\n", r.URL, r.Status)
		_, err = writer.WriteString(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing to CSV file: %v\n", err)
			os.Exit(1)
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error flushing CSV file: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, "results exported to output.csv")
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
	if flags.Addresses == "" && !flags.JSONFile {
		fmt.Fprintln(os.Stderr, "either --addresses or --file flag must be provided")
		os.Exit(1)
	}
}

func checkConflictingFlags(flags *FlagsUsed) {
	if flags.Addresses != "" && flags.JSONFile {
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
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			resp, err := getUrlStatus(addresses[index])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", addresses[index], err)
				return
			}
			mu.Lock()
			statuses[index] = resp
			mu.Unlock()
		}(i)
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
