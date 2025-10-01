package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/services"
)

// TestErrorRecoveryScenarios tests various error conditions and recovery behavior
func TestErrorRecoveryScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error scenario tests in short mode")
	}

	testEnv := setupTestEnvironment(t)
	defer testEnv.cleanup()

	workspaceID := uuid.New()
	workspace := &models.Workspace{
		ID:        workspaceID,
		Name:      "Error Test Workspace",
		APIKey:    "error-test-api-key",
		IsActive:  true,
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := testEnv.workspaceRepo.Create(context.Background(), workspace)
	require.NoError(t, err)

	t.Run("Network Timeout Scenarios", func(t *testing.T) {
		// Test analysis of unreachable URL
		analysisReq := models.AnalysisRequest{
			URL:         "https://unreachable-domain-12345.com",
			WorkspaceID: workspaceID,
			Options: models.AnalysisOptions{
				IncludePerformance: true,
			},
		}

		reqBody, _ := json.Marshal(analysisReq)
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer error-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		// Should handle gracefully and return appropriate error
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		errorObj := response["error"].(map[string]interface{})
		assert.Contains(t, errorObj, "code")
		assert.Contains(t, errorObj, "message")
	})

	t.Run("Invalid SSL Certificate Handling", func(t *testing.T) {
		// Test analysis of site with invalid SSL
		analysisReq := models.AnalysisRequest{
			URL:         "https://self-signed.badssl.com",
			WorkspaceID: workspaceID,
			Options: models.AnalysisOptions{
				IncludeSecurity: true,
			},
		}

		reqBody, _ := json.Marshal(analysisReq)
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer error-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		// Should handle SSL errors gracefully
		if rr.Code == http.StatusOK {
			var result models.AnalysisResult
			err := json.Unmarshal(rr.Body.Bytes(), &result)
			require.NoError(t, err)

			// Security metrics should indicate SSL issues
			if result.SecurityMetrics != nil {
				if sslInfo, exists := result.SecurityMetrics["ssl"]; exists {
					if ssl, ok := sslInfo.(map[string]interface{}); ok {
						assert.Contains(t, ssl, "valid")
						assert.Equal(t, false, ssl["valid"])
					}
				}
			}
		}
	})

	t.Run("Large Response Handling", func(t *testing.T) {
		// Test analysis of very large page
		analysisReq := models.AnalysisRequest{
			URL:         "https://httpbin.org/bytes/10485760", // 10MB response
			WorkspaceID: workspaceID,
			Options: models.AnalysisOptions{
				IncludePerformance: true,
			},
		}

		reqBody, _ := json.Marshal(analysisReq)
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer error-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		// Should handle large responses appropriately
		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusRequestEntityTooLarge}, rr.Code)
	})

	t.Run("Malformed HTML Handling", func(t *testing.T) {
		// Test analysis of page with malformed HTML
		analysisReq := models.AnalysisRequest{
			URL:         "https://httpbin.org/html",
			WorkspaceID: workspaceID,
			Options: models.AnalysisOptions{
				IncludeSEO:          true,
				IncludeAccessibility: true,
			},
		}

		reqBody, _ := json.Marshal(analysisReq)
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer error-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		// Should handle malformed HTML gracefully
		if rr.Code == http.StatusOK {
			var result models.AnalysisResult
			err := json.Unmarshal(rr.Body.Bytes(), &result)
			require.NoError(t, err)

			// Should still provide some analysis results
			assert.NotNil(t, result.Technologies)
		}
	})

	t.Run("Batch Analysis Partial Failures", func(t *testing.T) {
		// Test batch analysis with mix of valid and invalid URLs
		batchReq := services.BatchAnalysisRequest{
			URLs: []string{
				"https://httpbin.org/get",           // Valid
				"https://unreachable-domain.com",   // Invalid
				"https://httpbin.org/status/200",   // Valid
				"https://httpbin.org/status/500",   // Server error
				"not-a-url",                        // Malformed
			},
			WorkspaceID: workspaceID,
			Options: models.AnalysisOptions{
				IncludePerformance: true,
			},
		}

		reqBody, _ := json.Marshal(batchReq)
		req := httptest.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer error-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var batchResult services.BatchAnalysisResult
		err := json.Unmarshal(rr.Body.Bytes(), &batchResult)
		require.NoError(t, err)

		// Should have some successful results and some failures
		assert.Greater(t, len(batchResult.Results), 0)
		assert.Greater(t, len(batchResult.FailedURLs), 0)
		assert.Equal(t, 5, batchResult.Progress.Total)
		assert.Equal(t, len(batchResult.Results)+len(batchResult.FailedURLs), batchResult.Progress.Total)
	})

	t.Run("Database Connection Recovery", func(t *testing.T) {
		// This test would require more complex setup to simulate database failures
		// For now, we'll test that the system handles database errors gracefully

		// Test with non-existent workspace ID
		nonExistentWorkspaceID := uuid.New()
		analysisReq := models.AnalysisRequest{
			URL:         "https://httpbin.org/get",
			WorkspaceID: nonExistentWorkspaceID,
		}

		reqBody, _ := json.Marshal(analysisReq)
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer error-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		// Should handle gracefully (might succeed if workspace validation is not strict)
		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusNotFound}, rr.Code)
	})

	t.Run("Rate Limit Recovery", func(t *testing.T) {
		// Create workspace with very low rate limit
		lowLimitWorkspaceID := uuid.New()
		lowLimitWorkspace := &models.Workspace{
			ID:        lowLimitWorkspaceID,
			Name:      "Low Limit Workspace",
			APIKey:    "low-limit-api-key",
			IsActive:  true,
			RateLimit: 1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := testEnv.workspaceRepo.Create(context.Background(), lowLimitWorkspace)
		require.NoError(t, err)

		analysisReq := models.AnalysisRequest{
			URL:         "https://httpbin.org/get",
			WorkspaceID: lowLimitWorkspaceID,
		}
		reqBody, _ := json.Marshal(analysisReq)

		// First request should succeed
		req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer low-limit-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		// Second request should be rate limited
		req = httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer low-limit-api-key")

		rr = httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "RATE_LIMIT_EXCEEDED", errorObj["code"])
		assert.Contains(t, errorObj, "retry_after")
	})

	t.Run("Concurrent Request Handling", func(t *testing.T) {
		// Test system behavior under concurrent load
		analysisReq := models.AnalysisRequest{
			URL:         "https://httpbin.org/delay/1", // 1 second delay
			WorkspaceID: workspaceID,
		}
		reqBody, _ := json.Marshal(analysisReq)

		// Launch multiple concurrent requests
		results := make(chan int, 10)
		for i := 0; i < 10; i++ {
			go func() {
				req := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer error-test-api-key")

				rr := httptest.NewRecorder()
				testEnv.router.ServeHTTP(rr, req)
				results <- rr.Code
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < 10; i++ {
			select {
			case code := <-results:
				if code == http.StatusOK {
					successCount++
				}
			case <-time.After(30 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		// Should handle most requests successfully
		assert.GreaterOrEqual(t, successCount, 5, "Should handle at least half of concurrent requests successfully")
	})

	t.Run("Memory Pressure Handling", func(t *testing.T) {
		// Test system behavior under memory pressure by creating large batch requests
		urls := make([]string, 50)
		for i := 0; i < 50; i++ {
			urls[i] = "https://httpbin.org/get"
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
		req := httptest.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer error-test-api-key")

		rr := httptest.NewRecorder()
		testEnv.router.ServeHTTP(rr, req)

		// Should handle large batch requests appropriately
		assert.Contains(t, []int{http.StatusOK, http.StatusRequestEntityTooLarge, http.StatusInternalServerError}, rr.Code)

		if rr.Code == http.StatusOK {
			var batchResult services.BatchAnalysisResult
			err := json.Unmarshal(rr.Body.Bytes(), &batchResult)
			require.NoError(t, err)

			// Should process most URLs successfully
			assert.GreaterOrEqual(t, len(batchResult.Results), 25, "Should process at least half of the URLs")
		}
	})
}