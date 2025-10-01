package monitoring

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector handles Prometheus metrics collection
type MetricsCollector struct {
	// HTTP request metrics
	RequestDuration   *prometheus.HistogramVec
	RequestCount      *prometheus.CounterVec
	ActiveConnections prometheus.Gauge
	
	// Analysis operation metrics
	AnalysisOperations *prometheus.CounterVec
	AnalysisDuration   *prometheus.HistogramVec
	AnalysisErrors     *prometheus.CounterVec
	
	// Database metrics
	DatabaseConnections prometheus.Gauge
	DatabaseOperations  *prometheus.CounterVec
	DatabaseDuration    *prometheus.HistogramVec
	
	// Cache metrics
	CacheOperations *prometheus.CounterVec
	CacheHitRatio   prometheus.Gauge
	
	// Business metrics
	WorkspaceCount    prometheus.Gauge
	SessionCount      *prometheus.CounterVec
	InsightGeneration *prometheus.CounterVec
}

// NewMetricsCollector creates a new metrics collector with all Prometheus metrics
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		// HTTP request metrics
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "webailyzer_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),
		RequestCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webailyzer_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		ActiveConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "webailyzer_http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),
		
		// Analysis operation metrics
		AnalysisOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webailyzer_analysis_operations_total",
				Help: "Total number of analysis operations",
			},
			[]string{"type", "status"},
		),
		AnalysisDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "webailyzer_analysis_duration_seconds",
				Help:    "Duration of analysis operations in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
			[]string{"type"},
		),
		AnalysisErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webailyzer_analysis_errors_total",
				Help: "Total number of analysis errors",
			},
			[]string{"type", "error_type"},
		),
		
		// Database metrics
		DatabaseConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "webailyzer_database_connections",
				Help: "Number of active database connections",
			},
		),
		DatabaseOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webailyzer_database_operations_total",
				Help: "Total number of database operations",
			},
			[]string{"operation", "table", "status"},
		),
		DatabaseDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "webailyzer_database_operation_duration_seconds",
				Help:    "Duration of database operations in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
			},
			[]string{"operation", "table"},
		),
		
		// Cache metrics
		CacheOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webailyzer_cache_operations_total",
				Help: "Total number of cache operations",
			},
			[]string{"operation", "status"},
		),
		CacheHitRatio: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "webailyzer_cache_hit_ratio",
				Help: "Cache hit ratio (0-1)",
			},
		),
		
		// Business metrics
		WorkspaceCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "webailyzer_workspaces_total",
				Help: "Total number of active workspaces",
			},
		),
		SessionCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webailyzer_sessions_total",
				Help: "Total number of user sessions",
			},
			[]string{"workspace_id"},
		),
		InsightGeneration: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webailyzer_insights_generated_total",
				Help: "Total number of insights generated",
			},
			[]string{"type", "priority"},
		),
	}
}

// RecordHTTPRequest records metrics for an HTTP request
func (m *MetricsCollector) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	statusStr := strconv.Itoa(statusCode)
	m.RequestDuration.WithLabelValues(method, endpoint, statusStr).Observe(duration.Seconds())
	m.RequestCount.WithLabelValues(method, endpoint, statusStr).Inc()
}

// IncrementActiveConnections increments the active connections counter
func (m *MetricsCollector) IncrementActiveConnections() {
	m.ActiveConnections.Inc()
}

// DecrementActiveConnections decrements the active connections counter
func (m *MetricsCollector) DecrementActiveConnections() {
	m.ActiveConnections.Dec()
}

// RecordAnalysisOperation records metrics for an analysis operation
func (m *MetricsCollector) RecordAnalysisOperation(analysisType, status string, duration time.Duration) {
	m.AnalysisOperations.WithLabelValues(analysisType, status).Inc()
	m.AnalysisDuration.WithLabelValues(analysisType).Observe(duration.Seconds())
}

// RecordAnalysisError records an analysis error
func (m *MetricsCollector) RecordAnalysisError(analysisType, errorType string) {
	m.AnalysisErrors.WithLabelValues(analysisType, errorType).Inc()
}

// RecordDatabaseOperation records metrics for a database operation
func (m *MetricsCollector) RecordDatabaseOperation(operation, table, status string, duration time.Duration) {
	m.DatabaseOperations.WithLabelValues(operation, table, status).Inc()
	m.DatabaseDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// SetDatabaseConnections sets the current number of database connections
func (m *MetricsCollector) SetDatabaseConnections(count int) {
	m.DatabaseConnections.Set(float64(count))
}

// RecordCacheOperation records metrics for a cache operation
func (m *MetricsCollector) RecordCacheOperation(operation, status string) {
	m.CacheOperations.WithLabelValues(operation, status).Inc()
}

// SetCacheHitRatio sets the current cache hit ratio
func (m *MetricsCollector) SetCacheHitRatio(ratio float64) {
	m.CacheHitRatio.Set(ratio)
}

// SetWorkspaceCount sets the current number of workspaces
func (m *MetricsCollector) SetWorkspaceCount(count int) {
	m.WorkspaceCount.Set(float64(count))
}

// RecordSession records a new session
func (m *MetricsCollector) RecordSession(workspaceID string) {
	m.SessionCount.WithLabelValues(workspaceID).Inc()
}

// RecordInsightGeneration records insight generation
func (m *MetricsCollector) RecordInsightGeneration(insightType, priority string) {
	m.InsightGeneration.WithLabelValues(insightType, priority).Inc()
}