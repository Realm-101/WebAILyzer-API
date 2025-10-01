package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_HealthCheck_NilDependencies(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	handler := &HealthHandler{
		logger:       logger,
		dbConn:       nil,
		cacheService: nil,
	}

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "unhealthy", response.Status)
	assert.Equal(t, "webailyzer-lite-api", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
	assert.NotZero(t, response.Timestamp)

	// Verify checks
	assert.Contains(t, response.Checks, "database")
	assert.Contains(t, response.Checks, "cache")

	dbCheck := response.Checks["database"]
	assert.Equal(t, "unhealthy", dbCheck.Status)
	assert.Contains(t, dbCheck.Message, "not initialized")

	cacheCheck := response.Checks["cache"]
	assert.Equal(t, "unhealthy", cacheCheck.Status)
	assert.Contains(t, cacheCheck.Message, "not initialized")

	// Verify duration is recorded (may be 0 for very fast operations)
	assert.GreaterOrEqual(t, dbCheck.Duration, time.Duration(0))
	assert.GreaterOrEqual(t, cacheCheck.Duration, time.Duration(0))
}

func TestHealthHandler_ReadinessCheck_NilDatabase(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := &HealthHandler{
		logger: logger,
		dbConn: nil,
	}

	req := httptest.NewRequest("GET", "/api/ready", nil)
	w := httptest.NewRecorder()

	handler.ReadinessCheck(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["ready"])
	assert.Equal(t, "webailyzer-lite-api", response["service"])
	assert.NotNil(t, response["timestamp"])
}

func TestHealthHandler_LivenessCheck(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := &HealthHandler{
		logger: logger,
	}

	req := httptest.NewRequest("GET", "/api/live", nil)
	w := httptest.NewRecorder()

	handler.LivenessCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["alive"])
	assert.Equal(t, "webailyzer-lite-api", response["service"])
	assert.NotNil(t, response["timestamp"])
}

func TestHealthHandler_RegisterRoutes(t *testing.T) {
	logger := logrus.New()
	handler := NewHealthHandler(logger, nil, nil)
	router := mux.NewRouter()

	handler.RegisterRoutes(router)

	// Test that routes are registered
	req := httptest.NewRequest("GET", "/api/health", nil)
	match := &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))

	req = httptest.NewRequest("GET", "/api/ready", nil)
	match = &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))

	req = httptest.NewRequest("GET", "/api/live", nil)
	match = &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))
}

func TestHealthHandler_ResponseStructure(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := &HealthHandler{
		logger:       logger,
		dbConn:       nil,
		cacheService: nil,
	}

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.NotEmpty(t, response.Status)
	assert.Equal(t, "webailyzer-lite-api", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
	assert.NotZero(t, response.Timestamp)
	assert.NotNil(t, response.Checks)

	// Verify checks structure
	for checkName, check := range response.Checks {
		assert.NotEmpty(t, check.Status, "Check %s should have status", checkName)
		assert.GreaterOrEqual(t, check.Duration, time.Duration(0), "Check %s should have non-negative duration", checkName)
		assert.NotZero(t, check.Timestamp, "Check %s should have timestamp", checkName)
	}
}