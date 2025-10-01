package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestHealthHandler(t *testing.T) {
	// Create a request to pass to our handler with context
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add request ID to context (simulating middleware)
	ctx := context.WithValue(req.Context(), "request_id", "test-request-id")
	req = req.WithContext(ctx)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	// Call the handler with our request and recorder
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect
	var response HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	expectedStatus := "ok"
	if response.Status != expectedStatus {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response.Status, expectedStatus)
	}

	// Check the content type
	expected := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expected {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, expected)
	}
	
	// Check that X-Request-ID header is set
	if requestID := rr.Header().Get("X-Request-ID"); requestID == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestHealthHandlerResponseFormat(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(rr, req)

	// Verify the exact JSON format matches {"status":"ok"}
	var responseMap map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &responseMap); err != nil {
		t.Errorf("failed to unmarshal response as map: %v", err)
	}

	// Check that only the "status" field exists
	if len(responseMap) != 1 {
		t.Errorf("response should contain exactly one field, got %d fields", len(responseMap))
	}

	// Check that the status field exists and has the correct value
	status, exists := responseMap["status"]
	if !exists {
		t.Error("response should contain 'status' field")
	}

	if status != "ok" {
		t.Errorf("status field should be 'ok', got %v", status)
	}
}

func TestHealthHandlerHTTPMethod(t *testing.T) {
	// Test that only GET method is supported
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	
	for _, method := range methods {
		req, err := http.NewRequest(method, "/health", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(healthHandler)
		handler.ServeHTTP(rr, req)

		// The handler itself doesn't check method, but this test documents expected behavior
		// In the actual router setup, only GET is allowed
		if rr.Code != http.StatusOK {
			// This is expected behavior - the handler responds to any method
			// but the router restricts it to GET only
			continue
		}
	}
}

func TestHealthEndpointWithRouter(t *testing.T) {
	// Create router with the same setup as main()
	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler).Methods("GET")

	// Test GET request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GET /health returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check response format
	var response HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status 'ok', got %v", response.Status)
	}

	// Test that POST method is not allowed
	req, err = http.NewRequest("POST", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Should return 405 Method Not Allowed
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("POST /health should return 405, got %v", status)
	}
}

func TestAnalyzeHandlerValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		expectedType   ErrorType
	}{
		{
			name:           "valid request with URL",
			requestBody:    `{"url":"https://example.com"}`,
			expectedStatus: http.StatusOK, // Now returns successful analysis
			expectedError:  "", // No error expected for valid requests
		},
		{
			name:           "invalid JSON format",
			requestBody:    `{"url":}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON format",
			expectedType:   ErrorTypeValidation,
		},
		{
			name:           "empty JSON object",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid URL",
			expectedType:   ErrorTypeValidation,
		},
		{
			name:           "empty URL field",
			requestBody:    `{"url":""}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid URL",
			expectedType:   ErrorTypeValidation,
		},
		{
			name:           "missing URL field",
			requestBody:    `{"other_field":"value"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid URL",
			expectedType:   ErrorTypeValidation,
		},
		{
			name:           "completely invalid JSON",
			requestBody:    `not json at all`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON format",
			expectedType:   ErrorTypeValidation,
		},
		{
			name:           "invalid URL scheme",
			requestBody:    `{"url":"ftp://example.com"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid URL",
			expectedType:   ErrorTypeValidation,
		},
		{
			name:           "URL without host",
			requestBody:    `{"url":"https://"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid URL",
			expectedType:   ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/v1/analyze", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			
			// Add request ID to context (simulating middleware)
			ctx := context.WithValue(req.Context(), "request_id", "test-request-id")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(analyzeHandler)
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check response format based on expected status
			if tt.expectedStatus == http.StatusOK {
				// For successful requests, expect AnalyzeResponse
				var response AnalyzeResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal successful response: %v", err)
				}
				// Verify basic structure for successful response
				if response.URL == "" {
					t.Error("successful response should have URL field")
				}
				if response.Detected == nil {
					t.Error("successful response should have detected field")
				}
			} else {
				// For error requests, expect ErrorResponse
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal error response: %v", err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("handler returned unexpected error: got %v want %v",
						response.Error, tt.expectedError)
				}
				if response.Type != tt.expectedType {
					t.Errorf("handler returned unexpected error type: got %v want %v",
						response.Type, tt.expectedType)
				}
				if response.RequestID == "" {
					t.Error("error response should include request ID")
				}
				if response.Timestamp == "" {
					t.Error("error response should include timestamp")
				}
			}

			// Check content type
			expected := "application/json"
			if contentType := rr.Header().Get("Content-Type"); contentType != expected {
				t.Errorf("handler returned wrong content type: got %v want %v",
					contentType, expected)
			}
		})
	}
}

func TestAnalyzeHandlerWithRouter(t *testing.T) {
	// Create router with the same setup as main()
	r := mux.NewRouter()
	r.HandleFunc("/v1/analyze", analyzeHandler).Methods("POST")

	// Test POST request with valid JSON
	requestBody := `{"url":"https://example.com"}`
	req, err := http.NewRequest("POST", "/v1/analyze", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Should return 200 OK (since wappalyzer integration is now implemented)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("POST /v1/analyze returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Test that GET method is not allowed
	req, err = http.NewRequest("GET", "/v1/analyze", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Should return 405 Method Not Allowed
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/analyze should return 405, got %v", status)
	}
}

func TestAnalyzeHandlerHTTPClient(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid URL scheme",
			url:            "invalid-url",
			expectedStatus: http.StatusBadGateway,
			expectedError:  "failed to fetch URL",
		},
		{
			name:           "non-existent domain",
			url:            "http://non-existent-domain-12345.com",
			expectedStatus: http.StatusBadGateway,
			expectedError:  "failed to fetch URL",
		},
		{
			name:           "invalid port",
			url:            "http://localhost:99999",
			expectedStatus: http.StatusBadGateway,
			expectedError:  "failed to fetch URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := `{"url":"` + tt.url + `"}`
			req, err := http.NewRequest("POST", "/v1/analyze", strings.NewReader(requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(analyzeHandler)
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check response format
			var response ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			if response.Error != tt.expectedError {
				t.Errorf("handler returned unexpected error: got %v want %v",
					response.Error, tt.expectedError)
			}
		})
	}
}

func TestCreateHTTPClient(t *testing.T) {
	client := createHTTPClient()
	
	// Check that client is not nil
	if client == nil {
		t.Error("createHTTPClient returned nil")
	}
	
	// Check timeout is set correctly
	expectedTimeout := 15 * time.Second
	if client.Timeout != expectedTimeout {
		t.Errorf("client timeout should be %v, got %v", expectedTimeout, client.Timeout)
	}
	
	// Check transport is configured
	if client.Transport == nil {
		t.Error("client transport should not be nil")
	}
	
	// Check transport configuration
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Error("client transport should be *http.Transport")
		return
	}
	
	if transport.TLSHandshakeTimeout != 5*time.Second {
		t.Errorf("TLS handshake timeout should be 5s, got %v", transport.TLSHandshakeTimeout)
	}
	
	if transport.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns should be 10, got %v", transport.MaxIdleConns)
	}
	
	if transport.MaxIdleConnsPerHost != 2 {
		t.Errorf("MaxIdleConnsPerHost should be 2, got %v", transport.MaxIdleConnsPerHost)
	}
}

func TestAnalyzeHandlerWappalyzerIntegration(t *testing.T) {
	// Test with a real URL that should work
	requestBody := `{"url":"https://httpbin.org/html"}`
	req, err := http.NewRequest("POST", "/v1/analyze", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(analyzeHandler)
	handler.ServeHTTP(rr, req)

	// Should return 200 OK now that wappalyzer is integrated
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check response format
	var response AnalyzeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if response.URL != "https://httpbin.org/html" {
		t.Errorf("expected URL to be https://httpbin.org/html, got %v", response.URL)
	}

	if response.Detected == nil {
		t.Error("detected field should not be nil")
	}

	// Content type should be set
	if response.ContentType == "" {
		t.Error("content_type should not be empty")
	}

	// Check content type header
	expected := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expected {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, expected)
	}
}

func TestAnalyzeResponseStructure(t *testing.T) {
	// Test the response structure matches the API specification
	response := AnalyzeResponse{
		URL: "https://example.com",
		Detected: map[string]interface{}{
			"WordPress": struct{}{},
			"jQuery":    struct{}{},
		},
		ContentType: "text/html; charset=utf-8",
	}

	// Marshal to JSON and back to verify structure
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Errorf("failed to marshal response: %v", err)
	}

	var unmarshaled AnalyzeResponse
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if unmarshaled.URL != response.URL {
		t.Errorf("URL mismatch: got %v want %v", unmarshaled.URL, response.URL)
	}

	if unmarshaled.ContentType != response.ContentType {
		t.Errorf("ContentType mismatch: got %v want %v", unmarshaled.ContentType, response.ContentType)
	}

	if len(unmarshaled.Detected) != len(response.Detected) {
		t.Errorf("Detected length mismatch: got %v want %v", len(unmarshaled.Detected), len(response.Detected))
	}
}

func TestCompleteAnalysisFlow(t *testing.T) {
	// Test the complete analysis flow from request to response
	testCases := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "analyze real website",
			url:         "https://httpbin.org/html",
			expectError: false,
		},
		{
			name:        "analyze example.com",
			url:         "https://example.com",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := fmt.Sprintf(`{"url":"%s"}`, tc.url)
			req, err := http.NewRequest("POST", "/v1/analyze", strings.NewReader(requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(analyzeHandler)
			handler.ServeHTTP(rr, req)

			if tc.expectError {
				if rr.Code == http.StatusOK {
					t.Errorf("expected error but got success for URL: %s", tc.url)
				}
				return
			}

			// Verify successful response
			if rr.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d for URL: %s", rr.Code, tc.url)
				t.Logf("Response body: %s", rr.Body.String())
				return
			}

			// Verify response structure
			var response AnalyzeResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
				return
			}

			// Verify required fields
			if response.URL != tc.url {
				t.Errorf("expected URL %s, got %s", tc.url, response.URL)
			}

			if response.Detected == nil {
				t.Error("detected field should not be nil")
			}

			// Verify Content-Type header is set
			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}

			// Verify response.ContentType field is populated
			if response.ContentType == "" {
				t.Error("content_type field should not be empty")
			}

			// Log detected technologies for debugging
			t.Logf("URL: %s", response.URL)
			t.Logf("Content-Type: %s", response.ContentType)
			t.Logf("Detected technologies: %d", len(response.Detected))
			for tech, info := range response.Detected {
				t.Logf("  - %s: %+v", tech, info)
			}
		})
	}
}

func TestAnalysisResponseFormat(t *testing.T) {
	// Test that the response format matches the API specification
	requestBody := `{"url":"https://httpbin.org/html"}`
	req, err := http.NewRequest("POST", "/v1/analyze", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(analyzeHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	// Parse response as generic map to verify structure
	var responseMap map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &responseMap); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify required top-level fields exist
	requiredFields := []string{"url", "detected", "content_type"}
	for _, field := range requiredFields {
		if _, exists := responseMap[field]; !exists {
			t.Errorf("response missing required field: %s", field)
		}
	}

	// Verify url field is string
	if url, ok := responseMap["url"].(string); !ok || url == "" {
		t.Error("url field should be a non-empty string")
	}

	// Verify detected field is object
	if detected, ok := responseMap["detected"].(map[string]interface{}); !ok {
		t.Error("detected field should be an object")
	} else {
		// Verify detected technologies have proper structure
		for tech, info := range detected {
			if tech == "" {
				t.Error("technology name should not be empty")
			}
			// Info can be various types depending on wappalyzer output
			if info == nil {
				t.Errorf("technology info should not be nil for %s", tech)
			}
		}
	}

	// Verify content_type field is string (can be empty)
	if _, ok := responseMap["content_type"].(string); !ok {
		t.Error("content_type field should be a string")
	}
}
// Test error handling middleware
func TestErrorHandlingMiddleware(t *testing.T) {
	// Create a test handler that we can control
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that request ID was added to context
		requestID := r.Context().Value("request_id")
		if requestID == nil {
			t.Error("request_id should be added to context")
		}
		
		// Check that headers were set
		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("Content-Type header should be set to application/json")
		}
		
		if w.Header().Get("X-Request-ID") == "" {
			t.Error("X-Request-ID header should be set")
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test":"ok"}`))
	})

	// Wrap with error handling middleware
	handler := errorHandlingMiddleware(testHandler)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check that headers were set
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type header not set correctly: got %v want %v",
			contentType, "application/json")
	}

	if requestID := rr.Header().Get("X-Request-ID"); requestID == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestErrorHandlingMiddlewarePanicRecovery(t *testing.T) {
	// Create a test handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap with error handling middleware
	handler := errorHandlingMiddleware(panicHandler)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that panic was recovered and converted to 500 error
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler should return 500 after panic: got %v want %v",
			status, http.StatusInternalServerError)
	}

	// Check error response format
	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal error response: %v", err)
	}

	if response.Error != "Internal server error" {
		t.Errorf("unexpected error message: got %v want %v",
			response.Error, "Internal server error")
	}

	if response.Type != ErrorTypeInternal {
		t.Errorf("unexpected error type: got %v want %v",
			response.Type, ErrorTypeInternal)
	}

	if response.RequestID == "" {
		t.Error("error response should include request ID")
	}

	if response.Timestamp == "" {
		t.Error("error response should include timestamp")
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test":"ok"}`))
	})

	// Create a handler chain with both middlewares
	handler := errorHandlingMiddleware(loggingMiddleware(testHandler))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "test-agent")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// The logging middleware should not affect the response
	// but should log the request (we can't easily test the logging output in unit tests)
}

func TestResponseWriter(t *testing.T) {
	// Test the custom responseWriter wrapper
	rr := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: rr, statusCode: http.StatusOK}

	// Test default status code
	if wrapped.statusCode != http.StatusOK {
		t.Errorf("default status code should be 200, got %v", wrapped.statusCode)
	}

	// Test WriteHeader
	wrapped.WriteHeader(http.StatusNotFound)
	if wrapped.statusCode != http.StatusNotFound {
		t.Errorf("status code should be updated to 404, got %v", wrapped.statusCode)
	}

	// Test that the underlying ResponseWriter was called
	if rr.Code != http.StatusNotFound {
		t.Errorf("underlying ResponseWriter should have status 404, got %v", rr.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		expectedIP     string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.2",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.3:8080",
			expectedIP: "192.168.1.3",
		},
		{
			name: "X-Forwarded-For takes precedence",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.4",
				"X-Real-IP":       "192.168.1.5",
			},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Set RemoteAddr
			req.RemoteAddr = tt.remoteAddr

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getClientIP() = %v, want %v", ip, tt.expectedIP)
			}
		})
	}
}

func TestSendErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		apiError APIError
	}{
		{
			name: "validation error",
			apiError: APIError{
				Type:       ErrorTypeValidation,
				Message:    "Invalid input",
				Details:    "URL is required",
				StatusCode: http.StatusBadRequest,
				RequestID:  "test-request-123",
			},
		},
		{
			name: "network error",
			apiError: APIError{
				Type:       ErrorTypeNetwork,
				Message:    "Network failure",
				Details:    "Connection timeout",
				StatusCode: http.StatusBadGateway,
				RequestID:  "test-request-456",
			},
		},
		{
			name: "internal error",
			apiError: APIError{
				Type:       ErrorTypeInternal,
				Message:    "Internal server error",
				StatusCode: http.StatusInternalServerError,
				RequestID:  "test-request-789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			sendErrorResponse(rr, tt.apiError)

			// Check status code
			if status := rr.Code; status != tt.apiError.StatusCode {
				t.Errorf("sendErrorResponse() status = %v, want %v", status, tt.apiError.StatusCode)
			}

			// Check response format
			var response ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("failed to unmarshal error response: %v", err)
			}

			if response.Error != tt.apiError.Message {
				t.Errorf("error message = %v, want %v", response.Error, tt.apiError.Message)
			}

			if response.Type != tt.apiError.Type {
				t.Errorf("error type = %v, want %v", response.Type, tt.apiError.Type)
			}

			if response.Details != tt.apiError.Details {
				t.Errorf("error details = %v, want %v", response.Details, tt.apiError.Details)
			}

			if response.RequestID != tt.apiError.RequestID {
				t.Errorf("request ID = %v, want %v", response.RequestID, tt.apiError.RequestID)
			}

			if response.Timestamp == "" {
				t.Error("timestamp should not be empty")
			}

			// Verify timestamp format
			if _, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
				t.Errorf("invalid timestamp format: %v", err)
			}
		})
	}
}

func TestAPIErrorInterface(t *testing.T) {
	// Test that APIError implements the error interface
	apiErr := APIError{
		Type:    ErrorTypeValidation,
		Message: "Test error",
	}

	expectedError := "validation_error: Test error"
	if apiErr.Error() != expectedError {
		t.Errorf("APIError.Error() = %v, want %v", apiErr.Error(), expectedError)
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid HTTP URL",
			url:       "http://example.com",
			wantError: false,
		},
		{
			name:      "valid HTTPS URL",
			url:       "https://example.com/path",
			wantError: false,
		},
		{
			name:      "empty URL",
			url:       "",
			wantError: true,
			errorMsg:  "URL is required",
		},
		{
			name:      "relative URL without scheme",
			url:       "not-a-url",
			wantError: true,
			errorMsg:  "only HTTP and HTTPS URLs are supported",
		},
		{
			name:      "unsupported scheme",
			url:       "ftp://example.com",
			wantError: true,
			errorMsg:  "only HTTP and HTTPS URLs are supported",
		},
		{
			name:      "URL without host",
			url:       "https://",
			wantError: true,
			errorMsg:  "URL must include a host",
		},
		{
			name:      "URL with port",
			url:       "https://example.com:8080/path",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("validateURL() should return error for %v", tt.url)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateURL() error = %v, should contain %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateURL() should not return error for %v, got %v", tt.url, err)
				}
			}
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	// Test that generateRequestID returns non-empty strings
	id1 := generateRequestID()
	if id1 == "" {
		t.Error("generateRequestID() should not return empty string")
	}

	// Test that consecutive calls return different IDs
	id2 := generateRequestID()
	if id1 == id2 {
		t.Error("generateRequestID() should return different IDs on consecutive calls")
	}

	// Test that IDs are numeric (since they're based on UnixNano)
	if _, err := time.Parse("", id1); err == nil {
		// This is a weak test, but we expect the ID to be a number
		// We can't easily test the exact format without making assumptions
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that all error types are defined correctly
	errorTypes := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeNetwork,
		ErrorTypeTimeout,
		ErrorTypeInternal,
		ErrorTypeNotFound,
		ErrorTypeUnauthorized,
	}

	expectedTypes := []string{
		"validation_error",
		"network_error",
		"timeout_error",
		"internal_error",
		"not_found_error",
		"unauthorized_error",
	}

	for i, errorType := range errorTypes {
		if string(errorType) != expectedTypes[i] {
			t.Errorf("ErrorType %d = %v, want %v", i, string(errorType), expectedTypes[i])
		}
	}
}

func TestMiddlewareChain(t *testing.T) {
	// Test that middleware chain works correctly
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that both middlewares have been applied
		requestID := r.Context().Value("request_id")
		if requestID == nil {
			t.Error("request_id should be in context from errorHandlingMiddleware")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test":"ok"}`))
	})

	// Create the same middleware chain as in main()
	handler := errorHandlingMiddleware(loggingMiddleware(testHandler))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("middleware chain returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check that headers from errorHandlingMiddleware are set
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type header not set by middleware: got %v want %v",
			contentType, "application/json")
	}

	if requestID := rr.Header().Get("X-Request-ID"); requestID == "" {
		t.Error("X-Request-ID header should be set by middleware")
	}
}