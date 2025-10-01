package integration

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/webailyzer/webailyzer-lite-api/internal/config"
	"github.com/webailyzer/webailyzer-lite-api/internal/database"
	"github.com/webailyzer/webailyzer-lite-api/internal/handlers"
	"github.com/webailyzer/webailyzer-lite-api/internal/middleware"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/repositories/postgres"
	"github.com/webailyzer/webailyzer-lite-api/internal/services"
)

// TestE2EAnalysisWorkflow tests the complete analysis workflow from request to insights
func TestE2EAnalysisWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Setup test environment
	testEnv := setupTestEnvironment(t)
	defer testEnv.cleanup()

	workspaceID := uuid.New()
	sessionID := uuid.New()

	// Create test workspace
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "E2E Test Workspace",
		APIKey:    "e2e-test-api-key",
		IsActive:  true,
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(t, err)

	t.Run("Complete Analysis Workflow", func(t *testing.T) {
		// Step 1: Analyze a URL
		analysisReq := models.AnalysisRequest{
			URL:         "https://example.com",
			WorkspaceID: workspaceID,
			SessionID:   &sessionID,
			Options: models.AnalysisOptions{
				IncludePerformance:   true,
				IncludeSEO:          true,
				IncludeAccessibility: true,
				IncludeSecurity:     true,
			},
		}

		reqBody, _ := json.Marshal(analysisReq)
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer e2e-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var analysisResult models.AnalysisResult
		err := json.Unmarshal(rr.Body.Bytes(), &analysisResult)
		require.NoError(t, err)

		assert.Equal(t, analysisReq.URL, analysisResult.URL)
		assert.Equal(t, workspaceID, analysisResult.WorkspaceID)
		assert.NotNil(t, analysisResult.Technologies)
		assert.NotNil(t, analysisResult.PerformanceMetrics)

		// Step 2: Generate insights from the analysis
		insightsReq := map[string]string{
			"workspace_id": workspaceID.String(),
		}
		reqBody, _ = json.Marshal(insightsReq)
		req = httptest.NewRequest("POST", "/api/v1/insights/generate", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer e2e-test-api-key")

		rr = httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var insightsResponse map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &insightsResponse)
		require.NoError(t, err)

		assert.Equal(t, true, insightsResponse["success"])
		assert.Equal(t, workspaceID.String(), insightsResponse["workspace_id"])

		// Step 3: Retrieve generated insights
		req = httptest.NewRequest("GET", "/api/v1/insights?workspace_id="+workspaceID.String(), nil)
		req.Header.Set("Authorization", "Bearer e2e-test-api-key")

		rr = httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var getInsightsResponse map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &getInsightsResponse)
		require.NoError(t, err)

		insights := getInsightsResponse["insights"].([]interface{})
		assert.GreaterOrEqual(t, len(insights), 0)

		// Step 4: Get metrics for the workspace
		startDate := time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
		endDate := time.Now().Format(time.RFC3339)
		metricsURL := fmt.Sprintf("/api/v1/metrics?workspace_id=%s&start_date=%s&end_date=%s&granularity=daily",
			workspaceID.String(), startDate, endDate)

		req = httptest.NewRequest("GET", metricsURL, nil)
		req.Header.Set("Authorization", "Bearer e2e-test-api-key")

		rr = httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var metricsResponse map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &metricsResponse)
		require.NoError(t, err)

		assert.Contains(t, metricsResponse, "metrics")
		assert.Contains(t, metricsResponse, "kpis")
		assert.Contains(t, metricsResponse, "metadata")
	})
}

// TestE2EBatchProcessing tests batch analysis functionality
func TestE2EBatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	testEnv := setupTestEnvironment(t)
	defer testEnv.cleanup()

	workspaceID := uuid.New()

	// Create test workspace
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "Batch Test Workspace",
		APIKey:    "batch-test-api-key",
		IsActive:  true,
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(t, err)

	t.Run("Batch Analysis Processing", func(t *testing.T) {
		// Submit batch analysis request
		batchReq := services.BatchAnalysisRequest{
			URLs: []string{
				"https://example.com",
				"https://google.com",
				"https://github.com",
			},
			WorkspaceID: workspaceID,
			Options: models.AnalysisOptions{
				IncludePerformance: true,
				IncludeSEO:        true,
			},
		}

		reqBody, _ := json.Marshal(batchReq)
		req := httptest.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer batch-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var batchResult services.BatchAnalysisResult
		err := json.Unmarshal(rr.Body.Bytes(), &batchResult)
		require.NoError(t, err)

		assert.NotEmpty(t, batchResult.BatchID)
		assert.Equal(t, "completed", batchResult.Status)
		assert.Len(t, batchResult.Results, 3)
		assert.Equal(t, 3, batchResult.Progress.Total)
		assert.Equal(t, 3, batchResult.Progress.Completed)

		// Verify each result
		for i, result := range batchResult.Results {
			assert.Equal(t, batchReq.URLs[i], result.URL)
			assert.Equal(t, workspaceID, result.WorkspaceID)
			assert.NotNil(t, result.Technologies)
			assert.NotNil(t, result.PerformanceMetrics)
		}
	})
}

// TestE2EErrorScenarios tests error handling and recovery behavior
func TestE2EErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	testEnv := setupTestEnvironment(t)
	defer testEnv.cleanup()

	t.Run("Authentication Errors", func(t *testing.T) {
		// Test without API key
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "INVALID_API_KEY", errorObj["code"])
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		workspaceID := uuid.New()

		// Create workspace with low rate limit
		workspace := &models.Workspace{
			ID:        workspaceID,
			Name:      "Rate Limit Test",
			APIKey:    "rate-limit-test-key",
			IsActive:  true,
			RateLimit: 2, // Very low limit
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := testEnv.workspaceRepo.Create(context.Background(), workspace)
		require.NoError(t, err)

		analysisReq := models.AnalysisRequest{
			URL:         "https://example.com",
			WorkspaceID: workspaceID,
		}
		reqBody, _ := json.Marshal(analysisReq)

		// Make requests up to the limit
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer rate-limit-test-key")

			rr := httptest.NewRecorder()
			testEnv.router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
		}

		// Next request should be rate limited
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer rate-limit-test-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusTooManyRequests, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "RATE_LIMIT_EXCEEDED", errorObj["code"])
	})

	t.Run("Invalid Request Data", func(t *testing.T) {
		workspaceID := uuid.New()

		workspace := &models.Workspace{
			ID:        workspaceID,
			Name:      "Invalid Data Test",
			APIKey:    "invalid-data-test-key",
			IsActive:  true,
			RateLimit: 1000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := testEnv.workspaceRepo.Create(context.Background(), workspace)
		require.NoError(t, err)

		// Test invalid URL
		invalidReq := models.AnalysisRequest{
			URL:         "not-a-valid-url",
			WorkspaceID: workspaceID,
		}
		reqBody, _ := json.Marshal(invalidReq)

		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer invalid-data-test-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "INVALID_REQUEST", errorObj["code"])
	})
}

// TestEnvironment holds test dependencies
type TestEnvironment struct {
	router        *mux.Router
	dbConn        *database.Connection
	workspaceRepo *postgres.WorkspaceRepository
	cleanup       func()
}

// setupTestEnvironment creates a test environment with all dependencies
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	// Setup test database
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "webailyzer_test",
		SSLMode:  "disable",
	}

	dbConn, err := database.NewConnection(cfg, logrus.New())
	if err != nil {
		t.Skipf("Could not connect to test database: %v", err)
	}

	// Run migrations
	migrationManager := database.NewMigrationManager(dbConn.Pool, logrus.New())
	err = migrationManager.Migrate(context.Background())
	require.NoError(t, err)

	// Setup repositories
	workspaceRepo := postgres.NewWorkspaceRepository(dbConn.Pool)
	analysisRepo := postgres.NewAnalysisRepository(dbConn.Pool)
	insightRepo := postgres.NewInsightRepository(dbConn.Pool)
	metricsRepo := postgres.NewMetricsRepository(dbConn.Pool)

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	// Setup services
	analysisService, err := services.NewAnalysisService(analysisRepo, logger)
	require.NoError(t, err)
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
		DefaultLimit:    1000,
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

	return &TestEnvironment{
		router:        router,
		dbConn:        dbConn,
		workspaceRepo: workspaceRepo.(*postgres.WorkspaceRepository),
		cleanup:       cleanup,
	}
}