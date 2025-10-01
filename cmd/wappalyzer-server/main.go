package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

var (
	port    = flag.Int("port", 8080, "Server port")
	timeout = flag.Duration("timeout", 10*time.Second, "HTTP timeout for analysis requests")
)

type Server struct {
	wappalyzer *wappalyzer.Wappalyze
	client     *http.Client
}

type AnalysisRequest struct {
	URL        string `json:"url"`
	UserAgent  string `json:"user_agent,omitempty"`
	WithInfo   bool   `json:"with_info,omitempty"`
	WithCats   bool   `json:"with_cats,omitempty"`
}

type AnalysisResponse struct {
	URL          string                 `json:"url"`
	Title        string                 `json:"title,omitempty"`
	Technologies map[string]interface{} `json:"technologies"`
	Timestamp    time.Time              `json:"timestamp"`
	Duration     time.Duration          `json:"duration"`
	Error        string                 `json:"error,omitempty"`
}

func main() {
	flag.Parse()

	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		log.Fatalf("Failed to initialize wappalyzer: %v", err)
	}

	server := &Server{
		wappalyzer: wappalyzerClient,
		client:     &http.Client{Timeout: *timeout},
	}

	http.HandleFunc("/", server.handleHome)
	http.HandleFunc("/api/analyze", server.handleAnalyze)
	http.HandleFunc("/api/health", server.handleHealth)
	http.HandleFunc("/api/stats", server.handleStats)

	log.Printf("Starting server on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>Wappalyzer Technology Detection</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input[type="url"], input[type="text"] { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        button { background: #007cba; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #005a87; }
        .results { margin-top: 20px; padding: 20px; background: #f5f5f5; border-radius: 4px; }
        .tech-item { margin: 10px 0; padding: 10px; background: white; border-radius: 4px; }
        .loading { display: none; color: #666; }
        .error { color: #d32f2f; }
        .success { color: #388e3c; }
    </style>
</head>
<body>
    <h1>Wappalyzer Technology Detection</h1>
    <form id="analyzeForm">
        <div class="form-group">
            <label for="url">URL to analyze:</label>
            <input type="url" id="url" name="url" required placeholder="https://example.com">
        </div>
        <div class="form-group">
            <label for="userAgent">User Agent (optional):</label>
            <input type="text" id="userAgent" name="userAgent" placeholder="Custom user agent">
        </div>
        <div class="form-group">
            <label>
                <input type="checkbox" id="withInfo" name="withInfo"> Include detailed information
            </label>
        </div>
        <div class="form-group">
            <label>
                <input type="checkbox" id="withCats" name="withCats"> Include categories
            </label>
        </div>
        <button type="submit">Analyze</button>
        <div class="loading" id="loading">Analyzing...</div>
    </form>

    <div id="results" class="results" style="display: none;">
        <h2>Results</h2>
        <div id="resultsContent"></div>
    </div>

    <script>
        document.getElementById('analyzeForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const loading = document.getElementById('loading');
            const results = document.getElementById('results');
            const resultsContent = document.getElementById('resultsContent');
            
            loading.style.display = 'block';
            results.style.display = 'none';
            
            const formData = new FormData(e.target);
            const data = {
                url: formData.get('url'),
                user_agent: formData.get('userAgent') || '',
                with_info: formData.has('withInfo'),
                with_cats: formData.has('withCats')
            };
            
            try {
                const response = await fetch('/api/analyze', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                loading.style.display = 'none';
                
                if (result.error) {
                    resultsContent.innerHTML = '<div class="error">Error: ' + result.error + '</div>';
                } else {
                    let html = '<div class="success">Analysis completed in ' + result.duration + '</div>';
                    if (result.title) {
                        html += '<p><strong>Title:</strong> ' + result.title + '</p>';
                    }
                    html += '<h3>Technologies Detected (' + Object.keys(result.technologies).length + '):</h3>';
                    
                    for (const [tech, info] of Object.entries(result.technologies)) {
                        html += '<div class="tech-item">';
                        html += '<strong>' + tech + '</strong>';
                        if (info.description) {
                            html += '<br><small>' + info.description + '</small>';
                        }
                        if (info.website) {
                            html += '<br><a href="' + info.website + '" target="_blank">Website</a>';
                        }
                        if (info.categories && info.categories.length > 0) {
                            html += '<br><em>Categories: ' + info.categories.join(', ') + '</em>';
                        }
                        html += '</div>';
                    }
                    
                    resultsContent.innerHTML = html;
                }
                
                results.style.display = 'block';
            } catch (error) {
                loading.style.display = 'none';
                resultsContent.innerHTML = '<div class="error">Error: ' + error.message + '</div>';
                results.style.display = 'block';
            }
        });
    </script>
</body>
</html>`

	t, _ := template.New("home").Parse(tmpl)
	t.Execute(w, nil)
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Validate URL
	if _, err := url.Parse(req.URL); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	start := time.Now()
	response := s.analyzeURL(req)
	response.Duration = time.Since(start)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) analyzeURL(req AnalysisRequest) *AnalysisResponse {
	response := &AnalysisResponse{
		URL:       req.URL,
		Timestamp: time.Now(),
	}

	httpReq, err := http.NewRequest("GET", req.URL, nil)
	if err != nil {
		response.Error = fmt.Sprintf("Failed to create request: %v", err)
		return response
	}

	userAgent := req.UserAgent
	if userAgent == "" {
		userAgent = "wappalyzer-server/1.0"
	}
	httpReq.Header.Set("User-Agent", userAgent)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		response.Error = fmt.Sprintf("Failed to fetch URL: %v", err)
		return response
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = fmt.Sprintf("Failed to read response: %v", err)
		return response
	}

	response.Technologies = make(map[string]interface{})

	if req.WithInfo {
		technologies := s.wappalyzer.FingerprintWithInfo(resp.Header, body)
		for tech, info := range technologies {
			response.Technologies[tech] = info
		}
	} else if req.WithCats {
		technologies := s.wappalyzer.FingerprintWithCats(resp.Header, body)
		for tech, cats := range technologies {
			response.Technologies[tech] = cats
		}
	} else {
		technologies, title := s.wappalyzer.FingerprintWithTitle(resp.Header, body)
		response.Title = title
		for tech := range technologies {
			response.Technologies[tech] = struct{}{}
		}
	}

	return response
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	fingerprints := s.wappalyzer.GetFingerprints()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_apps":      len(fingerprints.Apps),
		"server_version":  "1.0.0",
		"last_updated":    time.Now().Format(time.RFC3339),
	})
}