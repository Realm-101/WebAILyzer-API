package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/webailyzer/webailyzer-lite-api/internal/config"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
)

func TestAuthAndRateLimitIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	// Setup mock workspace repository
	mockRepo := &MockWorkspaceRepository{}
	workspace := &models.Workspace{
		ID:        uuid.New(),
		Name:      "Test Workspace",
		APIKey:    "valid-api-key",
		IsActive:  true,
		RateLimit: 3, // Low limit for testing
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockRepo.On("GetByAPIKey", mock.Anything, "valid-api-key").Return(workspace, nil)

	// Setup middlewares
	authMiddleware := NewAuthMiddleware(mockRepo, logger)
	
	rateLimitCfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}
	rateLimitMiddleware := NewRateLimitMiddleware(rateLimitCfg, logger)
	defer rateLimitMiddleware.Stop()

	// Setup router with middlewares
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.Authenticate)
	apiRouter.Use(rateLimitMiddleware.RateLimit)

	// Test handler
	apiRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Test 1: Request without API key should be rejected by auth middleware
	req := httptest.NewRequest("GET", "/api/test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Empty(t, rr.Header().Get("X-RateLimit-Limit")) // No rate limit headers

	// Test 2: Valid requests within rate limit should succeed
	for i := 0; i < 3; i++ {
		req = httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Bearer valid-api-key")
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "3", rr.Header().Get("X-RateLimit-Limit"))
		
		remaining, err := strconv.Atoi(rr.Header().Get("X-RateLimit-Remaining"))
		require.NoError(t, err)
		assert.Equal(t, 2-i, remaining)
	}

	// Test 3: Request exceeding rate limit should be rejected by rate limit middleware
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer valid-api-key")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Equal(t, "3", rr.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rr.Header().Get("Retry-After"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", errorObj["code"])

	mockRepo.AssertExpectations(t)
}

func TestAuthAndRateLimitWithInvalidAPIKey(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	// Setup mock workspace repository
	mockRepo := &MockWorkspaceRepository{}
	mockRepo.On("GetByAPIKey", mock.Anything, "invalid-api-key").Return(nil, nil)

	// Setup middlewares
	authMiddleware := NewAuthMiddleware(mockRepo, logger)
	
	rateLimitCfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}
	rateLimitMiddleware := NewRateLimitMiddleware(rateLimitCfg, logger)
	defer rateLimitMiddleware.Stop()

	// Setup router with middlewares
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.Authenticate)
	apiRouter.Use(rateLimitMiddleware.RateLimit)

	// Test handler
	apiRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Request with invalid API key should be rejected by auth middleware
	// Rate limiting should not be applied
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-api-key")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Empty(t, rr.Header().Get("X-RateLimit-Limit")) // No rate limit headers

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "INVALID_API_KEY", errorObj["code"])

	mockRepo.AssertExpectations(t)
}

func TestAuthAndRateLimitWithInactiveWorkspace(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	// Setup mock workspace repository
	mockRepo := &MockWorkspaceRepository{}
	workspace := &models.Workspace{
		ID:        uuid.New(),
		Name:      "Inactive Workspace",
		APIKey:    "inactive-api-key",
		IsActive:  false, // Inactive workspace
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockRepo.On("GetByAPIKey", mock.Anything, "inactive-api-key").Return(workspace, nil)

	// Setup middlewares
	authMiddleware := NewAuthMiddleware(mockRepo, logger)
	
	rateLimitCfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}
	rateLimitMiddleware := NewRateLimitMiddleware(rateLimitCfg, logger)
	defer rateLimitMiddleware.Stop()

	// Setup router with middlewares
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.Authenticate)
	apiRouter.Use(rateLimitMiddleware.RateLimit)

	// Test handler
	apiRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Request with inactive workspace should be rejected by auth middleware
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer inactive-api-key")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Empty(t, rr.Header().Get("X-RateLimit-Limit")) // No rate limit headers

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "WORKSPACE_INACTIVE", errorObj["code"])

	mockRepo.AssertExpectations(t)
}