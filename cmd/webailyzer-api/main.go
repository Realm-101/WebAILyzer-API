package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

// Logger instance for structured logging
var logger = logrus.New()

func main() {
	// Initialize logger
	initLogger()

	// Create router
	r := mux.NewRouter()

	// Add error handling middleware
	r.Use(errorHandlingMiddleware)
	r.Use(loggingMiddleware)

	// Add CORS middleware
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(r)

	// Register routes
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/v1/analyze", analyzeHandler).Methods("POST")

	// Create server with appropriate timeouts
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      corsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting WebAIlyzer Lite API server on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server shutdown complete")
	}
}

// Error types for structured error handling
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation_error"
	ErrorTypeNetwork       ErrorType = "network_error"
	ErrorTypeTimeout       ErrorType = "timeout_error"
	ErrorTypeInternal      ErrorType = "internal_error"
	ErrorTypeNotFound      ErrorType = "not_found_error"
	ErrorTypeUnauthorized  ErrorType = "unauthorized_error"
)

// APIError represents a structured API error
type APIError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
	RequestID  string    `json:"request_id,omitempty"`
}

// Error implements the error interface
func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// healthHandler handles GET /health requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	requestID := ""
	if id := r.Context().Value("request_id"); id != nil {
		requestID = id.(string)
	}
	
	logger.WithField("request_id", requestID).Debug("Health check requested")
	
	response := HealthResponse{Status: "ok"}
	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err,
		}).Error("Failed to encode health response")
		
		sendErrorResponse(w, APIError{
			Type:       ErrorTypeInternal,
			Message:    "Failed to generate response",
			StatusCode: http.StatusInternalServerError,
			RequestID:  requestID,
		})
	}
}

// AnalyzeRequest represents the request structure for analysis
type AnalyzeRequest struct {
	URL string `json:"url"`
}

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Error     string    `json:"error"`
	Type      ErrorType `json:"type"`
	Details   string    `json:"details,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp string    `json:"timestamp"`
}

// AnalyzeResponse represents the analysis response structure
type AnalyzeResponse struct {
	URL         string                 `json:"url"`
	Detected    map[string]interface{} `json:"detected"`
	ContentType string                 `json:"content_type,omitempty"`
}

// initLogger initializes the structured logger
func initLogger() {
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(os.Stdout)
}

// generateRequestID generates a simple request ID for tracking
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// errorHandlingMiddleware provides consistent error handling across all endpoints
func errorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add request ID to context
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		// Set default headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", requestID)

		// Recover from panics and convert to internal server error
		defer func() {
			if err := recover(); err != nil {
				logger.WithFields(logrus.Fields{
					"request_id": requestID,
					"panic":      err,
					"method":     r.Method,
					"path":       r.URL.Path,
				}).Error("Panic recovered")

				sendErrorResponse(w, APIError{
					Type:       ErrorTypeInternal,
					Message:    "Internal server error",
					Details:    "An unexpected error occurred",
					StatusCode: http.StatusInternalServerError,
					RequestID:  requestID,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Context().Value("request_id").(string)

		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_agent": r.UserAgent(),
			"remote_ip":  getClientIP(r),
		}).Info("Request started")

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logger.WithFields(logrus.Fields{
			"request_id":  requestID,
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": wrapped.statusCode,
			"duration_ms": duration.Milliseconds(),
		}).Info("Request completed")
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// sendErrorResponse sends a structured error response
func sendErrorResponse(w http.ResponseWriter, apiErr APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)
	
	response := ErrorResponse{
		Error:     apiErr.Message,
		Type:      apiErr.Type,
		Details:   apiErr.Details,
		RequestID: apiErr.RequestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("Failed to encode error response")
	}
}

// validateURL validates if the provided URL is valid and safe
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL is required")
	}
	
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}
	
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS URLs are supported")
	}
	
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}
	
	return nil
}

// createHTTPClient creates a safe HTTP client with appropriate timeouts
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
			MaxIdleConnsPerHost: 2,
		},
	}
}

// analyzeHandler handles POST /v1/analyze requests
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	requestID := ""
	if id := r.Context().Value("request_id"); id != nil {
		requestID = id.(string)
	}
	
	logger.WithField("request_id", requestID).Debug("Analysis request started")
	
	// Parse JSON request
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err,
		}).Warn("Invalid JSON in request body")
		
		sendErrorResponse(w, APIError{
			Type:       ErrorTypeValidation,
			Message:    "Invalid JSON format",
			Details:    "Request body must be valid JSON",
			StatusCode: http.StatusBadRequest,
			RequestID:  requestID,
		})
		return
	}
	
	// Validate URL field
	if err := validateURL(req.URL); err != nil {
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"url":        req.URL,
			"error":      err,
		}).Warn("URL validation failed")
		
		sendErrorResponse(w, APIError{
			Type:       ErrorTypeValidation,
			Message:    "Invalid URL",
			Details:    err.Error(),
			StatusCode: http.StatusBadRequest,
			RequestID:  requestID,
		})
		return
	}
	
	logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"url":        req.URL,
	}).Info("Starting URL analysis")
	
	// Create HTTP client and fetch URL
	client := createHTTPClient()
	resp, err := client.Get(req.URL)
	if err != nil {
		// Determine error type based on error details
		var apiErr APIError
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			apiErr = APIError{
				Type:       ErrorTypeTimeout,
				Message:    "Request timeout",
				Details:    "The URL took too long to respond",
				StatusCode: http.StatusGatewayTimeout,
				RequestID:  requestID,
			}
		} else if strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "connection refused") {
			apiErr = APIError{
				Type:       ErrorTypeNetwork,
				Message:    "Network error",
				Details:    "Unable to connect to the specified URL",
				StatusCode: http.StatusBadGateway,
				RequestID:  requestID,
			}
		} else {
			apiErr = APIError{
				Type:       ErrorTypeNetwork,
				Message:    "Failed to fetch URL",
				Details:    "Network error occurred while fetching the URL",
				StatusCode: http.StatusBadGateway,
				RequestID:  requestID,
			}
		}
		
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"url":        req.URL,
			"error":      err,
			"error_type": apiErr.Type,
		}).Error("Failed to fetch URL")
		
		sendErrorResponse(w, apiErr)
		return
	}
	defer resp.Body.Close()
	
	// Check HTTP status code
	if resp.StatusCode >= 400 {
		logger.WithFields(logrus.Fields{
			"request_id":  requestID,
			"url":         req.URL,
			"status_code": resp.StatusCode,
		}).Warn("URL returned error status code")
		
		var apiErr APIError
		if resp.StatusCode == 404 {
			apiErr = APIError{
				Type:       ErrorTypeNotFound,
				Message:    "URL not found",
				Details:    fmt.Sprintf("The URL returned status code %d", resp.StatusCode),
				StatusCode: http.StatusNotFound,
				RequestID:  requestID,
			}
		} else if resp.StatusCode == 401 || resp.StatusCode == 403 {
			apiErr = APIError{
				Type:       ErrorTypeUnauthorized,
				Message:    "Access denied",
				Details:    fmt.Sprintf("The URL returned status code %d", resp.StatusCode),
				StatusCode: http.StatusForbidden,
				RequestID:  requestID,
			}
		} else {
			apiErr = APIError{
				Type:       ErrorTypeNetwork,
				Message:    "URL returned error",
				Details:    fmt.Sprintf("The URL returned status code %d", resp.StatusCode),
				StatusCode: http.StatusBadGateway,
				RequestID:  requestID,
			}
		}
		
		sendErrorResponse(w, apiErr)
		return
	}
	
	// Read response body with size limit
	const maxBodySize = 10 * 1024 * 1024 // 10MB limit
	limitedReader := io.LimitReader(resp.Body, maxBodySize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"url":        req.URL,
			"error":      err,
		}).Error("Failed to read response body")
		
		sendErrorResponse(w, APIError{
			Type:       ErrorTypeNetwork,
			Message:    "Failed to read response",
			Details:    "Error occurred while reading the response body",
			StatusCode: http.StatusBadGateway,
			RequestID:  requestID,
		})
		return
	}
	
	// Initialize wappalyzer engine
	wc, err := wappalyzer.New()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err,
		}).Error("Wappalyzer initialization failed")
		
		sendErrorResponse(w, APIError{
			Type:       ErrorTypeInternal,
			Message:    "Technology detection engine failed",
			Details:    "Unable to initialize the technology detection engine",
			StatusCode: http.StatusInternalServerError,
			RequestID:  requestID,
		})
		return
	}
	
	// Perform technology fingerprinting with detailed information
	detected := wc.FingerprintWithInfo(resp.Header, body)
	
	logger.WithFields(logrus.Fields{
		"request_id":         requestID,
		"url":                req.URL,
		"technologies_found": len(detected),
		"content_type":       resp.Header.Get("Content-Type"),
	}).Info("Analysis completed successfully")
	
	// Create response with detected technologies
	result := AnalyzeResponse{
		URL:         req.URL,
		Detected:    make(map[string]interface{}),
		ContentType: resp.Header.Get("Content-Type"),
	}
	
	// Convert detected technologies to interface{} map
	for tech, info := range detected {
		result.Detected[tech] = info
	}
	
	// Return successful analysis results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err,
		}).Error("Failed to encode analysis response")
		
		sendErrorResponse(w, APIError{
			Type:       ErrorTypeInternal,
			Message:    "Failed to generate response",
			Details:    "Error occurred while encoding the response",
			StatusCode: http.StatusInternalServerError,
			RequestID:  requestID,
		})
	}
}