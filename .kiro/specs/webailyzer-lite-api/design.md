# Design Document

## Overview

The WebAIlyzer Lite API is a minimal, stateless HTTP service that provides web technology detection using the wappalyzer engine. The design focuses on simplicity, reliability, and low operational costs by eliminating complex dependencies while maintaining the core functionality of website technology fingerprinting.

The system is designed to be deployed as a single binary or Docker container with no external dependencies, making it suitable for cost-effective hosting on basic infrastructure.

## Architecture

### High-Level Architecture

```
┌─────────────────┐
│   Client Apps   │
│                 │
└─────────┬───────┘
          │
          │ HTTP Requests
          │
┌─────────▼───────┐
│ WebAIlyzer Lite │
│      API        │
│                 │
│ ┌─────────────┐ │
│ │HTTP Handlers│ │
│ └─────────────┘ │
│                 │
│ ┌─────────────┐ │
│ │ Wappalyzer  │ │
│ │   Engine    │ │
│ └─────────────┘ │
│                 │
│ ┌─────────────┐ │
│ │HTTP Client  │ │
│ └─────────────┘ │
└─────────────────┘
```

### Component Architecture

The system follows a simple, stateless architecture:

1. **HTTP Server**: Gorilla Mux router handling incoming requests
2. **Request Handlers**: Simple handlers for /v1/analyze and /health endpoints
3. **Wappalyzer Integration**: Direct integration with wappalyzer engine for technology detection
4. **HTTP Client**: Safe HTTP client for fetching external URLs with timeouts

## Components and Interfaces

### API Endpoints

#### Analysis Endpoint
```go
POST /v1/analyze
Content-Type: application/json

{
  "url": "https://example.com"
}

Response:
{
  "url": "https://example.com",
  "detected": {
    "WordPress": {
      "categories": ["CMS"],
      "version": "6.0",
      "confidence": 100
    },
    "jQuery": {
      "categories": ["JavaScript libraries"],
      "version": "3.6.0",
      "confidence": 100
    }
  },
  "content_type": "text/html; charset=utf-8"
}
```

#### Health Check Endpoint
```go
GET /health

Response:
{
  "status": "ok"
}
```

### Core Components

#### HTTP Server
```go
func main() {
    r := mux.NewRouter()
    
    // Health endpoint
    r.HandleFunc("/health", healthHandler).Methods("GET")
    
    // Analysis endpoint
    r.HandleFunc("/v1/analyze", analyzeHandler).Methods("POST")
    
    srv := &http.Server{
        Addr:         ":8080",
        Handler:      r,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    srv.ListenAndServe()
}
```

#### Request/Response Types
```go
type AnalyzeRequest struct {
    URL string `json:"url"`
}

type AnalyzeResponse struct {
    URL         string                 `json:"url"`
    Detected    map[string]map[string]any `json:"detected"`
    ContentType string                 `json:"content_type,omitempty"`
}

type HealthResponse struct {
    Status string `json:"status"`
}
```

#### HTTP Client Configuration
```go
func createHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 15 * time.Second,
        Transport: &http.Transport{
            DialContext: (&net.Dialer{
                Timeout: 5 * time.Second,
            }).DialContext,
            MaxIdleConns:        10,
            IdleConnTimeout:     30 * time.Second,
            TLSHandshakeTimeout: 5 * time.Second,
        },
    }
}
```

### Analysis Pipeline

#### Wappalyzer Integration
```go
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
    var req AnalyzeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
        http.Error(w, `{"error":"invalid request: url required"}`, http.StatusBadRequest)
        return
    }

    // Fetch URL content
    client := createHTTPClient()
    resp, err := client.Get(req.URL)
    if err != nil {
        http.Error(w, `{"error":"fetch failed"}`, http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    // Initialize wappalyzer
    wc, err := wappalyzer.New()
    if err != nil {
        http.Error(w, `{"error":"wappalyzer init failed"}`, http.StatusInternalServerError)
        return
    }

    // Perform fingerprinting
    detected := wc.Fingerprint(resp.Header, body)

    // Return results
    result := AnalyzeResponse{
        URL:         req.URL,
        Detected:    detected,
        ContentType: resp.Header.Get("Content-Type"),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

## Data Models

### In-Memory Data Structures

Since this is a stateless API, all data is processed in memory without persistence:

```go
// Request structure
type AnalyzeRequest struct {
    URL string `json:"url" validate:"required,url"`
}

// Response structure  
type AnalyzeResponse struct {
    URL         string                 `json:"url"`
    Detected    map[string]map[string]any `json:"detected"`
    ContentType string                 `json:"content_type,omitempty"`
}

// Health check response
type HealthResponse struct {
    Status string `json:"status"`
}

// Error response
type ErrorResponse struct {
    Error string `json:"error"`
}
```

## Error Handling

### Error Response Strategy
```go
// Simple error handling with appropriate HTTP status codes
func handleError(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// Error scenarios:
// 400 Bad Request - Invalid JSON or missing URL
// 502 Bad Gateway - Cannot fetch the provided URL  
// 500 Internal Server Error - Wappalyzer initialization failure
// 408 Request Timeout - HTTP client timeout
```

### Timeout Configuration
```go
const (
    HTTPClientTimeout     = 15 * time.Second
    ConnectionTimeout     = 5 * time.Second
    TLSHandshakeTimeout   = 5 * time.Second
    ServerReadTimeout     = 10 * time.Second
    ServerWriteTimeout    = 30 * time.Second
    ServerIdleTimeout     = 60 * time.Second
)
```

## Testing Strategy

### Unit Testing
```go
func TestAnalyzeHandler(t *testing.T) {
    // Test valid requests
    // Test invalid JSON
    // Test missing URL
    // Test HTTP client errors
    // Test wappalyzer initialization errors
}

func TestHealthHandler(t *testing.T) {
    // Test health endpoint returns 200 OK
    // Test response format
}
```

### Integration Testing
```go
func TestEndToEnd(t *testing.T) {
    // Start test server
    // Make real HTTP requests
    // Verify response format and content
    // Test with various website types
}
```

### Manual Testing
```bash
# Health check
curl -s http://localhost:8080/health

# Technology detection
curl -s -X POST http://localhost:8080/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}' | jq .
```

## Performance Considerations

### Resource Management
```go
// HTTP client with connection pooling
client := &http.Client{
    Timeout: 15 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        10,
        IdleConnTimeout:     30 * time.Second,
        MaxIdleConnsPerHost: 2,
    },
}
```

### Memory Management
- No persistent storage means no memory leaks from database connections
- HTTP response bodies are read and closed properly
- Wappalyzer engine is initialized once per request (stateless)
- Garbage collection handles temporary objects automatically

### Concurrency
- Go's built-in HTTP server handles concurrent requests efficiently
- Each request is processed in its own goroutine
- No shared state between requests eliminates race conditions
- HTTP client connection pooling reduces connection overhead

### Deployment Optimization
```dockerfile
# Multi-stage build for minimal image size
FROM golang:1.24-alpine AS builder
# ... build steps ...

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
# ... minimal runtime image ...
```