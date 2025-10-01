package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	
	"github.com/webailyzer/webailyzer-lite-api/internal/monitoring"
)

// MetricsMiddleware collects HTTP request metrics
type MetricsMiddleware struct {
	collector *monitoring.MetricsCollector
	logger    *logrus.Logger
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(collector *monitoring.MetricsCollector, logger *logrus.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		collector: collector,
		logger:    logger,
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// CollectMetrics returns a middleware function that collects HTTP metrics
func (m *MetricsMiddleware) CollectMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Increment active connections (if collector is available)
		if m.collector != nil {
			m.collector.IncrementActiveConnections()
			defer m.collector.DecrementActiveConnections()
		}
		
		// Wrap response writer to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Call the next handler
		next.ServeHTTP(rw, r)
		
		// Calculate duration
		duration := time.Since(start)
		
		// Get endpoint pattern from mux router
		endpoint := m.getEndpointPattern(r)
		
		// Record metrics (if collector is available)
		if m.collector != nil {
			m.collector.RecordHTTPRequest(r.Method, endpoint, rw.statusCode, duration)
		}
		
		// Log slow requests
		if duration > 5*time.Second {
			m.logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"endpoint":    endpoint,
				"duration":    duration,
				"status_code": rw.statusCode,
			}).Warn("Slow HTTP request detected")
		}
	})
}

// getEndpointPattern extracts the route pattern from the request
func (m *MetricsMiddleware) getEndpointPattern(r *http.Request) string {
	// Try to get the route pattern from mux
	if route := mux.CurrentRoute(r); route != nil {
		if template, err := route.GetPathTemplate(); err == nil {
			return template
		}
	}
	
	// Fallback to path normalization
	path := r.URL.Path
	
	// Normalize common patterns to reduce cardinality
	if strings.HasPrefix(path, "/api/v1/analyze") {
		return "/api/v1/analyze"
	}
	if strings.HasPrefix(path, "/api/v1/batch") {
		return "/api/v1/batch"
	}
	if strings.HasPrefix(path, "/api/v1/insights") {
		return "/api/v1/insights"
	}
	if strings.HasPrefix(path, "/api/v1/events") {
		return "/api/v1/events"
	}
	if strings.HasPrefix(path, "/api/v1/metrics") {
		return "/api/v1/metrics"
	}
	if strings.HasPrefix(path, "/api/health") {
		return "/api/health"
	}
	if strings.HasPrefix(path, "/api/ready") {
		return "/api/ready"
	}
	if strings.HasPrefix(path, "/api/live") {
		return "/api/live"
	}
	if strings.HasPrefix(path, "/metrics") {
		return "/metrics"
	}
	
	// For unknown paths, use a generic label to avoid high cardinality
	return "unknown"
}

// GetMetricsCollector returns the metrics collector
func (m *MetricsMiddleware) GetMetricsCollector() *monitoring.MetricsCollector {
	return m.collector
}