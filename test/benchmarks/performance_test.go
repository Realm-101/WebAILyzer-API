package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/webailyzer/webailyzer-lite-api/internal/config"
	"github.com/webailyzer/webailyzer-lite-api/internal/database"
	"github.com/webailyzer/webailyzer-lite-api/internal/handlers"
	"github.com/webailyzer/webailyzer-lite-api/internal/middleware"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/repositories/postgres"
	"github.com/webailyzer/webailyzer-lite-api/internal/services"
)

// BenchmarkAnalysisEndpoint benchmarks the analysis endpoint performance
func BenchmarkAnalysisEndpoint(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	testEnv := setupBenchmarkEnvironment(b)
	defer testEnv.cleanup()

	workspaceID := uuid.New()
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "Benchmark Workspace",
		APIKey:    "benchmark-api-key",
		IsActive:  true,
		RateLimit: 10000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(b, err)

	analysisReq := models.AnalysisRequest{
		URL:         "https://example.com",
		WorkspaceID: workspaceID,
		Options: models.AnalysisOptions{
			IncludePerformance:   true,
			IncludeSEO:          true,
			IncludeAccessibility: true,
			IncludeSecurity:     true,
		},
	}

	reqBody, _ := json.Marshal(analysisReq)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer benchmark-api-key")

			rr := httptest.NewRecorder()
			testEnv.router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", rr.Code)
			}
		}
	})
}

// BenchmarkBatchAnalysis benchmarks batch analysis performance
func BenchmarkBatchAnalysis(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	testEnv := setupBenchmarkEnvironment(b)
	defer testEnv.cleanup()

	workspaceID := uuid.New()
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "Batch Benchmark Workspace",
		APIKey:    "batch-benchmark-api-key",
		IsActive:  true,
		RateLimit: 10000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(b, err)

	// Test different batch sizes
	batchSizes := []int{5, 10, 25, 50}

	for _, size := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", size), func(b *testing.B) {
			urls := make([]string, size)
			for i := 0; i < size; i++ {
				urls[i] = fmt.Sprintf("https://example%d.com", i)
			}

			batchReq := services.BatchAnalysisRequest{
				URLs:        urls,
				WorkspaceID: workspaceID,
				Options: models.AnalysisOptions{
					IncludePerformance: true,
					IncludeSEO:        true,
				},
			}

			reqBody, _ := json.Marshal(batchReq)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer(reqBody))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer batch-benchmark-api-key")

				rr := httptest.NewRecorder()
				testEnv.router.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					b.Errorf("Expected status 200, got %d", rr.Code)
				}
			}
		})
	}
}

// BenchmarkInsightsGeneration benchmarks insights generation performance
func BenchmarkInsightsGeneration(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	testEnv := setupBenchmarkEnvironment(b)
	defer testEnv.cleanup()

	workspaceID := uuid.New()
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "Insights Benchmark Workspace",
		APIKey:    "insights-benchmark-api-key",
		IsActive:  true,
		RateLimit: 10000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(b, err)

	insightsReq := map[string]string{
		"workspace_id": workspaceID.String(),
	}
	reqBody, _ := json.Marshal(insightsReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/v1/insights/generate", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer insights-benchmark-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", rr.Code)
		}
	}
}

// BenchmarkMetricsRetrieval benchmarks metrics retrieval performance
func BenchmarkMetricsRetrieval(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	testEnv := setupBenchmarkEnvironment(b)
	defer testEnv.cleanup()

	workspaceID := uuid.New()
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "Metrics Benchmark Workspace",
		APIKey:    "metrics-benchmark-api-key",
		IsActive:  true,
		RateLimit: 10000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(b, err)

	startDate := time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().Format(time.RFC3339)

	granularities := []string{"hourly", "daily", "weekly"}

	for _, granularity := range granularities {
		b.Run(fmt.Sprintf("Granularity_%s", granularity), func(b *testing.B) {
			metricsURL := fmt.Sprintf("/api/v1/metrics?workspace_id=%s&start_date=%s&end_date=%s&granularity=%s",
				workspaceID.String(), startDate, endDate, granularity)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest("GET", metricsURL, nil)
				req.Header.Set("Authorization", "Bearer metrics-benchmark-api-key")

				rr := httptest.NewRecorder()
				testEnv.router.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					b.Errorf("Expected status 200, got %d", rr.Code)
				}
			}
		})
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	testEnv := setupBenchmarkEnvironment(b)
	defer testEnv.cleanup()

	workspaceID := uuid.New()
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "Concurrent Benchmark Workspace",
		APIKey:    "concurrent-benchmark-api-key",
		IsActive:  true,
		RateLimit: 100000, // High limit for concurrent testing
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(b, err)

	analysisReq := models.AnalysisRequest{
		URL:         "https://example.com",
		WorkspaceID: workspaceID,
		Options: models.AnalysisOptions{
			IncludePerformance: true,
		},
	}
	reqBody, _ := json.Marshal(analysisReq)

	concurrencyLevels := []int{1, 10, 50, 100}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer concurrent-benchmark-api-key")

					rr := httptest.NewRecorder()
					testEnv.router.ServeHTTP(rr, req)

					if rr.Code != http.StatusOK {
						b.Errorf("Expected status 200, got %d", rr.Code)
					}
				}
			})
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	testEnv := setupBenchmarkEnvironment(b)
	defer testEnv.cleanup()

	workspaceID := uuid.New()
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "Memory Benchmark Workspace",
		APIKey:    "memory-benchmark-api-key",
		IsActive:  true,
		RateLimit: 10000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(b, err)

	// Large batch request to test memory usage
	urls := make([]string, 100)
	for i := 0; i < 100; i++ {
		urls[i] = fmt.Sprintf("https://example%d.com", i)
	}

	batchReq := services.BatchAnalysisRequest{
		URLs:        urls,
		WorkspaceID: workspaceID,
		Options: models.AnalysisOptions{
			IncludePerformance:   true,
			IncludeSEO:          true,
			IncludeAccessibility: true,
			IncludeSecurity:     true,
		},
	}
	reqBody, _ := json.Marshal(batchReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer memory-benchmark-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", rr.Code)
		}
	}
}

// BenchmarkEnvironment holds benchmark test dependencies
type BenchmarkEnvironment struct {
	router        *mux.Router
	dbConn        *database.Connection
	workspaceRepo *postgres.WorkspaceRepository
	cleanup       func()
}

// setupBenchmarkEnvironment creates a benchmark test environment
func setupBenchmarkEnvironment(b *testing.B) *BenchmarkEnvironment {
	// Setup test database
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "webailyzer_benchmark",
		SSLMode:  "disable",
	}

	dbConn, err := database.NewConnection(cfg, logrus.New())
	if err != nil {
		b.Skipf("Could not connect to benchmark database: %v", err)
	}

	// Run migrations
	migrationManager := database.NewMigrationManager(dbConn.Pool, logrus.New())
	err = migrationManager.Migrate(context.Background())
	require.NoError(b, err)

	// Setup repositories
	workspaceRepo := postgres.NewWorkspaceRepository(dbConn.Pool)
	analysisRepo := postgres.NewAnalysisRepository(dbConn.Pool)
	insightRepo := postgres.NewInsightRepository(dbConn.Pool)
	metricsRepo := postgres.NewMetricsRepository(dbConn.Pool)

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during benchmarks

	// Setup services
	analysisService, err := services.NewAnalysisService(analysisRepo, logger)
	require.NoError(b, err)
	insightsService := services.NewInsightsService(insightRepo)
	eventRepo := postgres.NewEventRepository(dbConn.Pool)
	sessionRepo := postgres.NewSessionRepository(dbConn.Pool)
	metricsService := services.NewMetricsService(metricsRepo, sessionRepo, eventRepo, analysisRepo)

	// Setup handlers
	analysisHandler := handlers.NewAnalysisHandler(analysisService, nil, logger)
	insightsHandler := handlers.NewInsightsHandler(insightsService, logger)
	metricsHandler := handlers.NewMetricsHandler(metricsService, nil, logger)

	// Setup middleware
	authMiddleware := middleware.NewAuthMiddleware(workspaceRepo, logger)
	rateLimitCfg := &config.RateLimitConfig{
		DefaultLimit:    100000, // High limit for benchmarks
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimitCfg, logger)

	// Setup router
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.Authenticate)
	apiRouter.Use(rateLimitMiddleware.RateLimit)

	analysisHandler.RegisterRoutes(apiRouter)
	insightsHandler.RegisterRoutes(apiRouter)
	metricsHandler.RegisterRoutes(apiRouter)

	cleanup := func() {
		rateLimitMiddleware.Stop()
		dbConn.Close()
	}

	return &BenchmarkEnvironment{
		router:        router,
		dbConn:        dbConn,
		workspaceRepo: workspaceRepo.(*postgres.WorkspaceRepository),
		cleanup:       cleanup,
	}
}