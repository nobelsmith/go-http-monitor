package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config for the colors used in the tool
const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

var Cfg = Config{}

func addEntry(results []CheckOutput, url string, active bool, elapsed time.Duration, err error) []CheckOutput {
	check := &CheckOutput{
		Resource: url,
		Status:   strconv.FormatBool(!active),
		Elapsed:  elapsed.String(),
		Error:    err.Error(),
	}
	results = append(results, *check)
	return results
}

func main() {
	// File flag setup and parse
	filenamePtr := flag.String("file", "monitor.yml", "Monitoring file")
	flag.Parse()

	// Open file for reading. Close when main closes
	file, err := os.Open(*filenamePtr)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read file into data
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Unmarshal data into config struct
	err = yaml.Unmarshal([]byte(data), &Cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	hostUnreachable := false
	// Make new results struct
	results := &JsonOutput{}

	// Make new http client
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: Cfg.Insecure}
	client := http.Client{
		Timeout: time.Duration(Cfg.TimeoutRequest) * time.Second,
	}

	for _, plugin := range Cfg.Checks {
		tmpString := ""

		start := time.Now()

		// Send Get Request to specified url
		if strings.Contains(plugin.URL, "http") {
			resp, err := client.Get(plugin.URL)
			elapsed := time.Since(start)

			// if we fail connecting to the host
			if err != nil {
				tmpString = "[NOK] " + plugin.URL + "\n"
				fmt.Printf(ErrorColor, tmpString)
				hostUnreachable = true

				results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed, err)
				continue
			}

			// Read content from response body
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				tmpString = "[NOK] " + plugin.URL + "\n"
				fmt.Printf(ErrorColor, tmpString)
				hostUnreachable = true

				results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed, err)
				continue
			}

			// If Verbose log response body content
			if Cfg.Verbose {
				tmpString = "URL : " + plugin.URL + "\n"
				tmpString += "Body : " + string(content)
				fmt.Println(tmpString)
			}

			// if the status code does not correspond
			if plugin.StatusCode != nil && *plugin.StatusCode != resp.StatusCode {
				tmpString = "[NOK] " + plugin.URL + "\n"
				fmt.Printf(ErrorColor, tmpString)
				hostUnreachable = true

				results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed, err)
				continue
			}

			// if your search string does not appear in the response body
			if plugin.Match != nil && !strings.Contains(string(content), *plugin.Match) {
				tmpString = "[NOK] " + plugin.URL + "\n"
				fmt.Printf(ErrorColor, tmpString)
				hostUnreachable = true

				results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed, err)
				continue
			}

			// if http response takes more time than expected
			if plugin.ResponseTime != nil {
				responseTimeDuration := time.Duration(*plugin.ResponseTime) * time.Millisecond
				if responseTimeDuration-elapsed < 0 {
					responseTime := strconv.Itoa(*plugin.ResponseTime)
					tmpString = "[NOK]  " + plugin.URL + ", Elapsed time: " + elapsed.String() + " instead of " + responseTime + "\n"
					fmt.Printf(ErrorColor, tmpString)
					hostUnreachable = true

					results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed, err)
					continue
				}
			}

		} else if plugin.TCP != "" {
			servAddr := plugin.TCP + ":" + strconv.Itoa(*plugin.Port)
			tcpAddr, _ := net.ResolveTCPAddr("tcp", servAddr)
			_, err := net.DialTCP("tcp", nil, tcpAddr)

			elapsed := time.Since(start)
			if err != nil { // error on tcp connect
				hostUnreachable = true
				tmpString = "[NOK] TCP:" + servAddr + "\n"
				fmt.Printf(ErrorColor, tmpString)
				results.Results = addEntry(results.Results, servAddr, hostUnreachable, elapsed, err)
				continue
			} else if plugin.ResponseTime != nil { // error on connection
				responseTimeDuration := time.Duration(*plugin.ResponseTime) * time.Millisecond
				if responseTimeDuration-elapsed < 0 {
					responseTime := strconv.Itoa(*plugin.ResponseTime)
					tmpString = "[NOK] TCP:" + servAddr + ", Elapsed time: " + elapsed.String() + " instead of " + responseTime + "\n"
					fmt.Printf(ErrorColor, tmpString)
					hostUnreachable = true
					results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed, err)
					continue
				}
			}
		}
	}

	PostSlack(results)
}
