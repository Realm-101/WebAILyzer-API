package monitoring

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.RequestDuration)
	assert.NotNil(t, collector.RequestCount)
	assert.NotNil(t, collector.ActiveConnections)
	assert.NotNil(t, collector.AnalysisOperations)
	assert.NotNil(t, collector.AnalysisDuration)
	assert.NotNil(t, collector.AnalysisErrors)
	assert.NotNil(t, collector.DatabaseConnections)
	assert.NotNil(t, collector.DatabaseOperations)
	assert.NotNil(t, collector.DatabaseDuration)
	assert.NotNil(t, collector.CacheOperations)
	assert.NotNil(t, collector.CacheHitRatio)
	assert.NotNil(t, collector.WorkspaceCount)
	assert.NotNil(t, collector.SessionCount)
	assert.NotNil(t, collector.InsightGeneration)
}

func TestMetricsCollector_RecordHTTPRequest(t *testing.T) {
	// Create a new registry for this test to avoid conflicts
	registry := prometheus.NewRegistry()
	
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)
	
	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)
	
	registry.MustRegister(requestDuration, requestCount)
	
	collector := &MetricsCollector{
		RequestDuration: requestDuration,
		RequestCount:    requestCount,
	}
	
	// Record a request
	collector.RecordHTTPRequest("GET", "/api/health", 200, 100*time.Millisecond)
	
	// Verify the metrics were recorded
	assert.Equal(t, 1.0, testutil.ToFloat64(requestCount.WithLabelValues("GET", "/api/health", "200")))
	assert.Equal(t, 1, testutil.CollectAndCount(requestDuration))
}

func TestMetricsCollector_ActiveConnections(t *testing.T) {
	registry := prometheus.NewRegistry()
	
	activeConnections := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)
	
	registry.MustRegister(activeConnections)
	
	collector := &MetricsCollector{
		ActiveConnections: activeConnections,
	}
	
	// Test increment and decrement
	collector.IncrementActiveConnections()
	assert.Equal(t, 1.0, testutil.ToFloat64(activeConnections))
	
	collector.IncrementActiveConnections()
	assert.Equal(t, 2.0, testutil.ToFloat64(activeConnections))
	
	collector.DecrementActiveConnections()
	assert.Equal(t, 1.0, testutil.ToFloat64(activeConnections))
}

func TestMetricsCollector_RecordAnalysisOperation(t *testing.T) {
	registry := prometheus.NewRegistry()
	
	analysisOperations := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_analysis_operations_total",
			Help: "Total number of analysis operations",
		},
		[]string{"type", "status"},
	)
	
	analysisDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_analysis_duration_seconds",
			Help:    "Duration of analysis operations in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
		[]string{"type"},
	)
	
	registry.MustRegister(analysisOperations, analysisDuration)
	
	collector := &MetricsCollector{
		AnalysisOperations: analysisOperations,
		AnalysisDuration:   analysisDuration,
	}
	
	// Record an analysis operation
	collector.RecordAnalysisOperation("technology", "success", 500*time.Millisecond)
	
	// Verify the metrics were recorded
	assert.Equal(t, 1.0, testutil.ToFloat64(analysisOperations.WithLabelValues("technology", "success")))
	assert.Equal(t, 1, testutil.CollectAndCount(analysisDuration))
}

func TestMetricsCollector_RecordAnalysisError(t *testing.T) {
	registry := prometheus.NewRegistry()
	
	analysisErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_analysis_errors_total",
			Help: "Total number of analysis errors",
		},
		[]string{"type", "error_type"},
	)
	
	registry.MustRegister(analysisErrors)
	
	collector := &MetricsCollector{
		AnalysisErrors: analysisErrors,
	}
	
	// Record an analysis error
	collector.RecordAnalysisError("seo", "timeout")
	
	// Verify the metric was recorded
	assert.Equal(t, 1.0, testutil.ToFloat64(analysisErrors.WithLabelValues("seo", "timeout")))
}

func TestMetricsCollector_DatabaseMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	
	dbConnections := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_database_connections",
			Help: "Number of active database connections",
		},
	)
	
	dbOperations := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_database_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"},
	)
	
	dbDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_database_operation_duration_seconds",
			Help:    "Duration of database operations in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
		},
		[]string{"operation", "table"},
	)
	
	registry.MustRegister(dbConnections, dbOperations, dbDuration)
	
	collector := &MetricsCollector{
		DatabaseConnections: dbConnections,
		DatabaseOperations:  dbOperations,
		DatabaseDuration:    dbDuration,
	}
	
	// Test database connection count
	collector.SetDatabaseConnections(5)
	assert.Equal(t, 5.0, testutil.ToFloat64(dbConnections))
	
	// Test database operation recording
	collector.RecordDatabaseOperation("select", "analysis_results", "success", 10*time.Millisecond)
	assert.Equal(t, 1.0, testutil.ToFloat64(dbOperations.WithLabelValues("select", "analysis_results", "success")))
	assert.Equal(t, 1, testutil.CollectAndCount(dbDuration))
}

func TestMetricsCollector_CacheMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	
	cacheOperations := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"},
	)
	
	cacheHitRatio := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_cache_hit_ratio",
			Help: "Cache hit ratio (0-1)",
		},
	)
	
	registry.MustRegister(cacheOperations, cacheHitRatio)
	
	collector := &MetricsCollector{
		CacheOperations: cacheOperations,
		CacheHitRatio:   cacheHitRatio,
	}
	
	// Test cache operations
	collector.RecordCacheOperation("get", "hit")
	collector.RecordCacheOperation("get", "miss")
	collector.RecordCacheOperation("set", "success")
	
	assert.Equal(t, 1.0, testutil.ToFloat64(cacheOperations.WithLabelValues("get", "hit")))
	assert.Equal(t, 1.0, testutil.ToFloat64(cacheOperations.WithLabelValues("get", "miss")))
	assert.Equal(t, 1.0, testutil.ToFloat64(cacheOperations.WithLabelValues("set", "success")))
	
	// Test cache hit ratio
	collector.SetCacheHitRatio(0.75)
	assert.Equal(t, 0.75, testutil.ToFloat64(cacheHitRatio))
}

func TestMetricsCollector_BusinessMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	
	workspaceCount := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_workspaces_total",
			Help: "Total number of active workspaces",
		},
	)
	
	sessionCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_sessions_total",
			Help: "Total number of user sessions",
		},
		[]string{"workspace_id"},
	)
	
	insightGeneration := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_insights_generated_total",
			Help: "Total number of insights generated",
		},
		[]string{"type", "priority"},
	)
	
	registry.MustRegister(workspaceCount, sessionCount, insightGeneration)
	
	collector := &MetricsCollector{
		WorkspaceCount:    workspaceCount,
		SessionCount:      sessionCount,
		InsightGeneration: insightGeneration,
	}
	
	// Test workspace count
	collector.SetWorkspaceCount(10)
	assert.Equal(t, 10.0, testutil.ToFloat64(workspaceCount))
	
	// Test session recording
	collector.RecordSession("workspace-123")
	assert.Equal(t, 1.0, testutil.ToFloat64(sessionCount.WithLabelValues("workspace-123")))
	
	// Test insight generation
	collector.RecordInsightGeneration("performance", "high")
	assert.Equal(t, 1.0, testutil.ToFloat64(insightGeneration.WithLabelValues("performance", "high")))
}