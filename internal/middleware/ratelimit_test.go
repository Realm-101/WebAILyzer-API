package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/webailyzer/webailyzer-lite-api/internal/config"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
)

func TestRateLimitMiddleware_RateLimit(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	cfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}

	rl := NewRateLimitMiddleware(cfg, logger)
	defer rl.Stop()

	workspaceID := uuid.New()
	authContext := &models.AuthContext{
		WorkspaceID: workspaceID,
		APIKey:      "test-api-key",
		RateLimit:   5, // Lower limit for testing
	}

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test successful requests within limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		rl.RateLimit(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "5", rr.Header().Get("X-RateLimit-Limit"))
		
		remaining, err := strconv.Atoi(rr.Header().Get("X-RateLimit-Remaining"))
		require.NoError(t, err)
		assert.Equal(t, 4-i, remaining)
		
		assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Reset"))
	}

	// Test rate limit exceeded
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	rl.RateLimit(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Equal(t, "5", rr.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rr.Header().Get("Retry-After"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", errorObj["code"])
}

func TestRateLimitMiddleware_NoAuthContext(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	cfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}

	rl := NewRateLimitMiddleware(cfg, logger)
	defer rl.Stop()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	rl.RateLimit(testHandler).ServeHTTP(rr, req)

	// Should pass through without rate limiting
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Header().Get("X-RateLimit-Limit"))
}

func TestRateLimitMiddleware_WindowReset(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	cfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  100 * time.Millisecond, // Short window for testing
		CleanupInterval: time.Minute,
	}

	rl := NewRateLimitMiddleware(cfg, logger)
	defer rl.Stop()

	workspaceID := uuid.New()
	authContext := &models.AuthContext{
		WorkspaceID: workspaceID,
		APIKey:      "test-api-key",
		RateLimit:   2,
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Use up the rate limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		rl.RateLimit(testHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	rl.RateLimit(testHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	req = httptest.NewRequest("GET", "/test", nil)
	ctx = context.WithValue(req.Context(), AuthContextKeyValue, authContext)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	rl.RateLimit(testHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "1", rr.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimitMiddleware_MultipleWorkspaces(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	cfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}

	rl := NewRateLimitMiddleware(cfg, logger)
	defer rl.Stop()

	workspace1 := uuid.New()
	workspace2 := uuid.New()

	authContext1 := &models.AuthContext{
		WorkspaceID: workspace1,
		APIKey:      "test-api-key-1",
		RateLimit:   2,
	}

	authContext2 := &models.AuthContext{
		WorkspaceID: workspace2,
		APIKey:      "test-api-key-2",
		RateLimit:   3,
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test workspace 1 - use up its limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext1)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		rl.RateLimit(testHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Workspace 1 should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext1)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	rl.RateLimit(testHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// Workspace 2 should still work
	req = httptest.NewRequest("GET", "/test", nil)
	ctx = context.WithValue(req.Context(), AuthContextKeyValue, authContext2)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	rl.RateLimit(testHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "2", rr.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimitMiddleware_DefaultLimit(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	cfg := &config.RateLimitConfig{
		DefaultLimit:    5,
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}

	rl := NewRateLimitMiddleware(cfg, logger)
	defer rl.Stop()

	workspaceID := uuid.New()
	authContext := &models.AuthContext{
		WorkspaceID: workspaceID,
		APIKey:      "test-api-key",
		RateLimit:   0, // Should use default limit
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	rl.RateLimit(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Limit")) // Should show 0 (workspace limit)
	// But internally should use default limit of 5
}

func TestRateLimitMiddleware_GetStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	cfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  time.Hour,
		CleanupInterval: time.Minute,
	}

	rl := NewRateLimitMiddleware(cfg, logger)
	defer rl.Stop()

	workspaceID := uuid.New()
	authContext := &models.AuthContext{
		WorkspaceID: workspaceID,
		APIKey:      "test-api-key",
		RateLimit:   5,
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Make a request to create an entry
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	rl.RateLimit(testHandler).ServeHTTP(rr, req)

	// Get stats
	stats := rl.GetRateLimitStats()

	assert.Equal(t, 1, stats["total_workspaces"])
	assert.Equal(t, time.Hour.String(), stats["window_duration"])
	assert.Equal(t, 10, stats["default_limit"])

	workspaces, ok := stats["workspaces"].(map[string]interface{})
	require.True(t, ok)

	workspaceStats, ok := workspaces[workspaceID.String()].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, workspaceStats["count"])
	assert.Equal(t, 5, workspaceStats["limit"])
	assert.Equal(t, 4, workspaceStats["remaining"])
}

func TestRateLimitMiddleware_Cleanup(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	cfg := &config.RateLimitConfig{
		DefaultLimit:    10,
		WindowDuration:  50 * time.Millisecond, // Very short for testing
		CleanupInterval: 25 * time.Millisecond, // Frequent cleanup
	}

	rl := NewRateLimitMiddleware(cfg, logger)
	defer rl.Stop()

	workspaceID := uuid.New()
	authContext := &models.AuthContext{
		WorkspaceID: workspaceID,
		APIKey:      "test-api-key",
		RateLimit:   5,
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Make a request to create an entry
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), AuthContextKeyValue, authContext)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	rl.RateLimit(testHandler).ServeHTTP(rr, req)

	// Verify entry exists
	stats := rl.GetRateLimitStats()
	assert.Equal(t, 1, stats["total_workspaces"])

	// Wait for cleanup to occur
	time.Sleep(100 * time.Millisecond)

	// Entry should be cleaned up
	stats = rl.GetRateLimitStats()
	assert.Equal(t, 0, stats["total_workspaces"])
}