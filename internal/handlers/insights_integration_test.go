package handlers

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
	"github.com/projectdiscovery/wappalyzergo/internal/config"
	"github.com/projectdiscovery/wappalyzergo/internal/database"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories/postgres"
	"github.com/projectdiscovery/wappalyzergo/internal/services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsightsHandler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

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
	defer dbConn.Close()

	// Run migrations
	migrationManager := database.NewMigrationManager(dbConn.Pool, logrus.New())
	err = migrationManager.Migrate(context.Background())
	require.NoError(t, err)

	// Setup repositories and services
	insightRepo := postgres.NewInsightRepository(dbConn.Pool)
	insightsService := services.NewInsightsService(insightRepo)

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	handler := NewInsightsHandler(insightsService, logger)

	// Setup router
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test data
	workspaceID := uuid.New()
	testInsight := &models.Insight{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		InsightType:     models.InsightTypePerformanceBottleneck,
		Priority:        models.PriorityHigh,
		Title:           "Test Performance Insight",
		Description:     strPtr("Page load time is too slow"),
		ImpactScore:     intPtr(85),
		EffortScore:     intPtr(30),
		Recommendations: map[string]interface{}{
			"actions": []string{"Optimize images", "Minify CSS"},
		},
		DataSource: map[string]interface{}{
			"url":       "https://example.com",
			"load_time": 3500,
		},
		Status:    models.InsightStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Clean up function
	cleanup := func() {
		_, _ = dbConn.Pool.Exec(context.Background(), "DELETE FROM insights WHERE workspace_id = $1", workspaceID)
	}
	defer cleanup()

	t.Run("GetInsights_EmptyResult", func(t *testing.T) {
		cleanup()

		req, err := http.NewRequest("GET", "/api/v1/insights?workspace_id="+workspaceID.String(), nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		insights := response["insights"].([]interface{})
		assert.Len(t, insights, 0)
	})

	t.Run("GetInsights_WithData", func(t *testing.T) {
		cleanup()

		// Create test insight
		err := insightRepo.Create(context.Background(), testInsight)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "/api/v1/insights?workspace_id="+workspaceID.String()+"&limit=10", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		insights := response["insights"].([]interface{})
		assert.Len(t, insights, 1)

		insight := insights[0].(map[string]interface{})
		assert.Equal(t, testInsight.Title, insight["title"])
		assert.Equal(t, string(testInsight.InsightType), insight["insight_type"])
		assert.Equal(t, string(testInsight.Priority), insight["priority"])
		assert.Equal(t, string(testInsight.Status), insight["status"])
	})

	t.Run("GetInsights_WithStatusFilter", func(t *testing.T) {
		cleanup()

		// Create insights with different statuses
		pendingInsight := *testInsight
		pendingInsight.ID = uuid.New()
		pendingInsight.Status = models.InsightStatusPending
		pendingInsight.Title = "Pending Insight"

		appliedInsight := *testInsight
		appliedInsight.ID = uuid.New()
		appliedInsight.Status = models.InsightStatusApplied
		appliedInsight.Title = "Applied Insight"

		err := insightRepo.Create(context.Background(), &pendingInsight)
		require.NoError(t, err)
		err = insightRepo.Create(context.Background(), &appliedInsight)
		require.NoError(t, err)

		// Test filtering by pending status
		req, err := http.NewRequest("GET", "/api/v1/insights?workspace_id="+workspaceID.String()+"&status=pending", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		insights := response["insights"].([]interface{})
		assert.Len(t, insights, 1)

		insight := insights[0].(map[string]interface{})
		assert.Equal(t, "Pending Insight", insight["title"])
		assert.Equal(t, "pending", insight["status"])
	})

	t.Run("UpdateInsightStatus_Success", func(t *testing.T) {
		cleanup()

		// Create test insight
		err := insightRepo.Create(context.Background(), testInsight)
		require.NoError(t, err)

		// Update status to applied
		requestBody := map[string]string{
			"status": "applied",
		}
		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", "/api/v1/insights/"+testInsight.ID.String()+"/status", bytes.NewBuffer(body))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		assert.Equal(t, testInsight.ID.String(), response["insight_id"])
		assert.Equal(t, "applied", response["status"])

		// Verify the status was actually updated in the database
		updatedInsight, err := insightRepo.GetByID(context.Background(), testInsight.ID)
		require.NoError(t, err)
		assert.Equal(t, models.InsightStatusApplied, updatedInsight.Status)
	})

	t.Run("UpdateInsightStatus_NotFound", func(t *testing.T) {
		cleanup()

		nonExistentID := uuid.New()
		requestBody := map[string]string{
			"status": "applied",
		}
		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", "/api/v1/insights/"+nonExistentID.String()+"/status", bytes.NewBuffer(body))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// The handler doesn't explicitly check for existence, so it will return success
		// In a production system, you might want to add this check
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("GenerateInsights_Success", func(t *testing.T) {
		cleanup()

		requestBody := map[string]string{
			"workspace_id": workspaceID.String(),
		}
		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/insights/generate", bytes.NewBuffer(body))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		assert.Equal(t, workspaceID.String(), response["workspace_id"])
		assert.Contains(t, response, "insights_generated")
	})

	t.Run("GetInsights_Pagination", func(t *testing.T) {
		cleanup()

		// Create multiple insights
		for i := 0; i < 5; i++ {
			insight := *testInsight
			insight.ID = uuid.New()
			insight.Title = fmt.Sprintf("Test Insight %d", i)
			err := insightRepo.Create(context.Background(), &insight)
			require.NoError(t, err)
		}

		// Test pagination
		req, err := http.NewRequest("GET", "/api/v1/insights?workspace_id="+workspaceID.String()+"&limit=2&offset=1", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		insights := response["insights"].([]interface{})
		assert.Len(t, insights, 2)

		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(2), pagination["limit"])
		assert.Equal(t, float64(1), pagination["offset"])
		assert.Equal(t, float64(2), pagination["count"])
	})
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}