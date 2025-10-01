package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	
	"github.com/projectdiscovery/wappalyzergo/internal/database"
	"github.com/projectdiscovery/wappalyzergo/internal/cache"
)

// HealthChecker interface for components that can be health checked
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// HealthHandler handles health check requests
type HealthHandler struct {
	logger    *logrus.Logger
	dbConn    *database.Connection
	cacheService *cache.CacheService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *logrus.Logger, dbConn *database.Connection, cacheService *cache.CacheService) *HealthHandler {
	return &HealthHandler{
		logger:    logger,
		dbConn:    dbConn,
		cacheService: cacheService,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents an individual health check result
type HealthCheck struct {
	Status    string        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration_ms"`
	Timestamp time.Time     `json:"timestamp"`
}

// HealthCheck handles GET /api/health requests
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	checks := make(map[string]HealthCheck)
	overallStatus := "healthy"

	// Check database health
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// Check cache health
	cacheCheck := h.checkCache(ctx)
	checks["cache"] = cacheCheck
	if cacheCheck.Status != "healthy" && overallStatus == "healthy" {
		overallStatus = "degraded"
	}

	response := HealthResponse{
		Status:    overallStatus,
		Service:   "webailyzer-lite-api",
		Version:   "1.0.0",
		Timestamp: time.Now(),
		Checks:    checks,
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		statusCode = http.StatusOK // Still return 200 for degraded state
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// ReadinessCheck handles GET /api/ready requests
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// For readiness, we only check critical dependencies
	dbCheck := h.checkDatabase(ctx)
	
	ready := dbCheck.Status == "healthy"
	
	response := map[string]interface{}{
		"ready":     ready,
		"service":   "webailyzer-lite-api",
		"timestamp": time.Now(),
		"checks": map[string]HealthCheck{
			"database": dbCheck,
		},
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// LivenessCheck handles GET /api/live requests
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	// Liveness check is simple - if we can respond, we're alive
	response := map[string]interface{}{
		"alive":     true,
		"service":   "webailyzer-lite-api",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// checkDatabase performs a database health check
func (h *HealthHandler) checkDatabase(ctx context.Context) HealthCheck {
	start := time.Now()
	
	if h.dbConn == nil {
		return HealthCheck{
			Status:    "unhealthy",
			Message:   "database connection not initialized",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	err := h.dbConn.HealthCheck(ctx)
	duration := time.Since(start)

	if err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		return HealthCheck{
			Status:    "unhealthy",
			Message:   err.Error(),
			Duration:  duration,
			Timestamp: time.Now(),
		}
	}

	return HealthCheck{
		Status:    "healthy",
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// checkCache performs a cache health check
func (h *HealthHandler) checkCache(ctx context.Context) HealthCheck {
	start := time.Now()
	
	if h.cacheService == nil {
		return HealthCheck{
			Status:    "unhealthy",
			Message:   "cache service not initialized",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	err := h.cacheService.HealthCheck(ctx)
	duration := time.Since(start)

	if err != nil {
		h.logger.WithError(err).Warn("Cache health check failed")
		return HealthCheck{
			Status:    "unhealthy",
			Message:   err.Error(),
			Duration:  duration,
			Timestamp: time.Now(),
		}
	}

	return HealthCheck{
		Status:    "healthy",
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// RegisterRoutes registers health routes with the router
func (h *HealthHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/health", h.HealthCheck).Methods("GET")
	router.HandleFunc("/api/ready", h.ReadinessCheck).Methods("GET")
	router.HandleFunc("/api/live", h.LivenessCheck).Methods("GET")
}