package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

var (
	url        = flag.String("url", "", "URL to analyze")
	output     = flag.String("output", "json", "Output format: json, table, csv")
	timeout    = flag.Duration("timeout", 10*time.Second, "HTTP timeout")
	userAgent  = flag.String("user-agent", "wappalyzer-cli/1.0", "User agent string")
	verbose    = flag.Bool("verbose", false, "Verbose output")
	categories = flag.Bool("categories", false, "Include category information")
	info       = flag.Bool("info", false, "Include detailed app information")
)

type Result struct {
	URL          string                     `json:"url"`
	Title        string                     `json:"title,omitempty"`
	Technologies map[string]interface{}     `json:"technologies"`
	Timestamp    time.Time                  `json:"timestamp"`
	Duration     time.Duration              `json:"duration"`
}

func main() {
	flag.Parse()

	if *url == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -url <URL>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	client := &http.Client{Timeout: *timeout}
	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		log.Fatalf("Failed to initialize wappalyzer: %v", err)
	}

	start := time.Now()
	result, err := analyzeURL(*url, client, wappalyzerClient)
	if err != nil {
		log.Fatalf("Failed to analyze URL: %v", err)
	}
	result.Duration = time.Since(start)

	switch *output {
	case "json":
		outputJSON(result)
	case "table":
		outputTable(result)
	case "csv":
		outputCSV(result)
	default:
		log.Fatalf("Unknown output format: %s", *output)
	}
}

func analyzeURL(targetURL string, client *http.Client, wappalyzerClient *wappalyzer.Wappalyze) (*Result, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", *userAgent)

	if *verbose {
		log.Printf("Analyzing URL: %s", targetURL)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	result := &Result{
		URL:       targetURL,
		Timestamp: time.Now(),
	}

	if *info {
		technologies := wappalyzerClient.FingerprintWithInfo(resp.Header, body)
		result.Technologies = make(map[string]interface{})
		for tech, appInfo := range technologies {
			result.Technologies[tech] = appInfo
		}
	} else if *categories {
		technologies := wappalyzerClient.FingerprintWithCats(resp.Header, body)
		result.Technologies = make(map[string]interface{})
		for tech, catInfo := range technologies {
			result.Technologies[tech] = catInfo
		}
	} else {
		technologies, title := wappalyzerClient.FingerprintWithTitle(resp.Header, body)
		result.Title = title
		result.Technologies = make(map[string]interface{})
		for tech := range technologies {
			result.Technologies[tech] = struct{}{}
		}
	}

	return result, nil
}

func outputJSON(result *Result) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}
}

func outputTable(result *Result) {
	fmt.Printf("URL: %s\n", result.URL)
	if result.Title != "" {
		fmt.Printf("Title: %s\n", result.Title)
	}
	fmt.Printf("Analysis Duration: %v\n", result.Duration)
	fmt.Printf("Timestamp: %s\n\n", result.Timestamp.Format(time.RFC3339))

	fmt.Println("Technologies Detected:")
	fmt.Println(strings.Repeat("-", 50))
	
	for tech, data := range result.Technologies {
		fmt.Printf("• %s", tech)
		if *info {
			if appInfo, ok := data.(wappalyzer.AppInfo); ok {
				if appInfo.Description != "" {
					fmt.Printf("\n  Description: %s", appInfo.Description)
				}
				if appInfo.Website != "" {
					fmt.Printf("\n  Website: %s", appInfo.Website)
				}
				if len(appInfo.Categories) > 0 {
					fmt.Printf("\n  Categories: %s", strings.Join(appInfo.Categories, ", "))
				}
			}
		}
		fmt.Println()
	}
}

func outputCSV(result *Result) {
	fmt.Println("Technology,Description,Website,Categories")
	for tech, data := range result.Technologies {
		if *info {
			if appInfo, ok := data.(wappalyzer.AppInfo); ok {
				fmt.Printf("%s,\"%s\",\"%s\",\"%s\"\n",
					tech,
					strings.ReplaceAll(appInfo.Description, "\"", "\"\""),
					appInfo.Website,
					strings.Join(appInfo.Categories, "; "))
			}
		} else {
			fmt.Printf("%s,,,\n", tech)
		}
	}
}