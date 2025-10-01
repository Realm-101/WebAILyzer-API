package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	
	"github.com/projectdiscovery/wappalyzergo/internal/monitoring"
)

func TestMetricsMiddleware_CollectMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	collector := monitoring.NewMetricsCollector()
	middleware := NewMetricsMiddleware(collector, logger)
	
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate some processing time
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Wrap with metrics middleware
	wrappedHandler := middleware.CollectMetrics(testHandler)
	
	// Create a test request
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	
	// Execute the request
	wrappedHandler.ServeHTTP(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
	
	// Note: In a real test, we would verify that metrics were recorded,
	// but since we're using the global Prometheus registry, we can't easily
	// verify the exact values without interfering with other tests.
	// In production, you might want to use a custom registry for testing.
}

func TestMetricsMiddleware_GetEndpointPattern(t *testing.T) {
	logger := logrus.New()
	// Don't create a new collector to avoid registration conflicts
	middleware := &MetricsMiddleware{
		collector: nil, // We don't need the collector for this test
		logger:    logger,
	}
	
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "analyze endpoint",
			path:     "/api/v1/analyze",
			expected: "/api/v1/analyze",
		},
		{
			name:     "batch endpoint",
			path:     "/api/v1/batch",
			expected: "/api/v1/batch",
		},
		{
			name:     "insights endpoint",
			path:     "/api/v1/insights",
			expected: "/api/v1/insights",
		},
		{
			name:     "events endpoint",
			path:     "/api/v1/events",
			expected: "/api/v1/events",
		},
		{
			name:     "metrics endpoint",
			path:     "/api/v1/metrics",
			expected: "/api/v1/metrics",
		},
		{
			name:     "health endpoint",
			path:     "/api/health",
			expected: "/api/health",
		},
		{
			name:     "ready endpoint",
			path:     "/api/ready",
			expected: "/api/ready",
		},
		{
			name:     "live endpoint",
			path:     "/api/live",
			expected: "/api/live",
		},
		{
			name:     "prometheus metrics endpoint",
			path:     "/metrics",
			expected: "/metrics",
		},
		{
			name:     "unknown endpoint",
			path:     "/some/unknown/path",
			expected: "unknown",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			result := middleware.getEndpointPattern(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetricsMiddleware_WithMuxRouter(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	// Don't create a new collector to avoid registration conflicts
	middleware := &MetricsMiddleware{
		collector: nil, // We don't need the collector for this test
		logger:    logger,
	}
	
	// Create a mux router with a route
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/test/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	// Wrap with metrics middleware
	wrappedRouter := middleware.CollectMetrics(router)
	
	// Create a test request
	req := httptest.NewRequest("GET", "/api/v1/test/123", nil)
	w := httptest.NewRecorder()
	
	// Execute the request
	wrappedRouter.ServeHTTP(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestMetricsMiddleware_ResponseWriter(t *testing.T) {
	// Test the custom response writer
	w := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
	
	// Test WriteHeader
	rw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, rw.statusCode)
	assert.True(t, rw.written)
	
	// Test that subsequent WriteHeader calls don't change the status
	rw.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusCreated, rw.statusCode) // Should remain unchanged
}

func TestMetricsMiddleware_ResponseWriterWrite(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
	
	// Test Write without explicit WriteHeader
	n, err := rw.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, http.StatusOK, rw.statusCode)
	assert.True(t, rw.written)
}

func TestMetricsMiddleware_GetMetricsCollector(t *testing.T) {
	logger := logrus.New()
	// Create a simple test without using the global registry
	middleware := &MetricsMiddleware{
		collector: nil,
		logger:    logger,
	}
	
	result := middleware.GetMetricsCollector()
	assert.Nil(t, result) // Since we set it to nil
}

func TestMetricsMiddleware_SlowRequestLogging(t *testing.T) {
	// This test would require capturing log output to verify slow request logging
	// For now, we'll just ensure the middleware doesn't panic with slow requests
	
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise
	
	// Don't create a new collector to avoid registration conflicts
	middleware := &MetricsMiddleware{
		collector: nil,
		logger:    logger,
	}
	
	// Create a slow test handler
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't actually sleep for 6 seconds in tests, just verify the middleware works
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Wrap with metrics middleware
	wrappedHandler := middleware.CollectMetrics(slowHandler)
	
	// Create a test request
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	
	// Execute the request
	wrappedHandler.ServeHTTP(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}