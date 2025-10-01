package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/services"
	"github.com/projectdiscovery/wappalyzergo/internal/cache"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MetricsHandler handles metrics-related HTTP requests
type MetricsHandler struct {
	metricsService services.MetricsService
	cacheService   *cache.CacheService
	logger         *logrus.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(metricsService services.MetricsService, cacheService *cache.CacheService, logger *logrus.Logger) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
		cacheService:   cacheService,
		logger:         logger,
	}
}

// GetMetrics handles GET /api/v1/metrics requests with filtering and pagination
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse and validate metrics request from query parameters
	req, err := h.parseMetricsRequest(r)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse metrics request")
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Generate cache key for the request
	cacheKey := h.generateMetricsCacheKey(req)

	// Try to get cached result if caching is enabled
	if h.cacheService != nil {
		var cachedResult models.MetricsResponse
		if err := h.cacheService.GetWithConfig(r.Context(), cacheKey, &cachedResult); err == nil {
			// Add freshness metadata to cached response
			response := h.addFreshnessMetadata(&cachedResult, true)
			
			h.logger.WithFields(logrus.Fields{
				"workspace_id": req.WorkspaceID,
				"granularity":  req.Granularity,
				"cache_key":    cacheKey,
				"duration":     time.Since(startTime),
			}).Info("Returning cached metrics result")

			h.writeJSONResponse(w, http.StatusOK, response)
			return
		}
	}

	// Get metrics from service
	result, err := h.metricsService.GetMetrics(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"workspace_id": req.WorkspaceID,
			"start_date":   req.StartDate,
			"end_date":     req.EndDate,
			"granularity":  req.Granularity,
		}).Error("Failed to get metrics")

		// Return appropriate error based on error type
		if err.Error() == "invalid metrics request: workspace_id is required" ||
			err.Error() == "invalid metrics request: start_date is required" ||
			err.Error() == "invalid metrics request: end_date is required" ||
			err.Error() == "invalid metrics request: granularity is required" {
			h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		} else if err.Error() == "invalid metrics request: end_date must be after start_date" {
			h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_DATE_RANGE", "End date must be after start date", map[string]interface{}{
				"start_date": req.StartDate,
				"end_date":   req.EndDate,
			})
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve metrics", map[string]interface{}{
				"workspace_id": req.WorkspaceID,
			})
		}
		return
	}

	// Cache the result if caching is enabled
	if h.cacheService != nil {
		// Use dynamic TTL based on granularity
		config := cache.MetricsConfig
		switch req.Granularity {
		case "hourly":
			config.TTL = 5 * time.Minute
		case "daily":
			config.TTL = 30 * time.Minute
		case "weekly":
			config.TTL = 2 * time.Hour
		case "monthly":
			config.TTL = 6 * time.Hour
		}
		
		if err := h.cacheService.SetWithConfig(r.Context(), cacheKey, result, config); err != nil {
			h.logger.WithError(err).WithField("cache_key", cacheKey).Warn("Failed to cache metrics result")
		}
	}

	// Add freshness metadata to response
	response := h.addFreshnessMetadata(result, false)

	// Log successful metrics retrieval
	h.logger.WithFields(logrus.Fields{
		"workspace_id":   req.WorkspaceID,
		"start_date":     req.StartDate,
		"end_date":       req.EndDate,
		"granularity":    req.Granularity,
		"metrics_count":  len(result.Metrics),
		"kpis_count":     len(result.KPIs),
		"anomalies_count": len(result.Anomalies),
		"duration":       time.Since(startTime),
	}).Info("Metrics retrieved successfully")

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RegisterRoutes registers metrics routes with the router
func (h *MetricsHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/metrics", h.GetMetrics).Methods("GET")
}

// parseMetricsRequest parses metrics request from query parameters
func (h *MetricsHandler) parseMetricsRequest(r *http.Request) (*models.MetricsRequest, error) {
	req := &models.MetricsRequest{}

	// Parse workspace_id (required)
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		return nil, fmt.Errorf("workspace_id query parameter is required")
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace_id format: %s", workspaceIDStr)
	}
	req.WorkspaceID = workspaceID

	// Parse start_date (required)
	startDateStr := r.URL.Query().Get("start_date")
	if startDateStr == "" {
		return nil, fmt.Errorf("start_date query parameter is required")
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, expected RFC3339: %s", startDateStr)
	}
	req.StartDate = startDate

	// Parse end_date (required)
	endDateStr := r.URL.Query().Get("end_date")
	if endDateStr == "" {
		return nil, fmt.Errorf("end_date query parameter is required")
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format, expected RFC3339: %s", endDateStr)
	}
	req.EndDate = endDate

	// Parse granularity (required)
	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		return nil, fmt.Errorf("granularity query parameter is required")
	}

	// Validate granularity values
	validGranularities := map[string]bool{
		"hourly":  true,
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}

	if !validGranularities[granularity] {
		return nil, fmt.Errorf("invalid granularity, must be one of: hourly, daily, weekly, monthly")
	}
	req.Granularity = granularity

	// Validate date range
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	// Validate date range limits based on granularity
	maxDuration := h.getMaxDurationForGranularity(granularity)
	if req.EndDate.Sub(req.StartDate) > maxDuration {
		return nil, fmt.Errorf("date range too large for %s granularity, maximum allowed: %v", granularity, maxDuration)
	}

	return req, nil
}

// getMaxDurationForGranularity returns the maximum allowed duration for each granularity
func (h *MetricsHandler) getMaxDurationForGranularity(granularity string) time.Duration {
	switch granularity {
	case "hourly":
		return 7 * 24 * time.Hour // 7 days
	case "daily":
		return 90 * 24 * time.Hour // 90 days
	case "weekly":
		return 365 * 24 * time.Hour // 1 year
	case "monthly":
		return 2 * 365 * 24 * time.Hour // 2 years
	default:
		return 30 * 24 * time.Hour // 30 days default
	}
}

// generateMetricsCacheKey generates a cache key for the metrics request
func (h *MetricsHandler) generateMetricsCacheKey(req *models.MetricsRequest) string {
	return fmt.Sprintf("metrics:%s:%s:%s:%s",
		req.WorkspaceID.String(),
		req.StartDate.Format("2006-01-02T15:04:05Z07:00"),
		req.EndDate.Format("2006-01-02T15:04:05Z07:00"),
		req.Granularity,
	)
}



// addFreshnessMetadata adds freshness metadata to the metrics response
func (h *MetricsHandler) addFreshnessMetadata(result *models.MetricsResponse, fromCache bool) map[string]interface{} {
	response := map[string]interface{}{
		"metrics":   result.Metrics,
		"kpis":      result.KPIs,
		"anomalies": result.Anomalies,
		"metadata": map[string]interface{}{
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"from_cache":  fromCache,
			"data_source": "real_time",
		},
	}

	// Add data freshness indicators
	if len(result.Metrics) > 0 {
		// Find the most recent data point across all metrics
		var mostRecentTimestamp time.Time
		for _, metric := range result.Metrics {
			for _, dataPoint := range metric.DataPoints {
				if dataPoint.Timestamp.After(mostRecentTimestamp) {
					mostRecentTimestamp = dataPoint.Timestamp
				}
			}
		}

		if !mostRecentTimestamp.IsZero() {
			response["metadata"].(map[string]interface{})["most_recent_data"] = mostRecentTimestamp.Format(time.RFC3339)
			
			// Calculate data age
			dataAge := time.Since(mostRecentTimestamp)
			response["metadata"].(map[string]interface{})["data_age_minutes"] = int(dataAge.Minutes())
			
			// Determine freshness status
			var freshnessStatus string
			if dataAge < time.Hour {
				freshnessStatus = "fresh"
			} else if dataAge < 24*time.Hour {
				freshnessStatus = "recent"
			} else {
				freshnessStatus = "stale"
			}
			response["metadata"].(map[string]interface{})["freshness_status"] = freshnessStatus
		}
	}

	return response
}

// writeJSONResponse writes a JSON response
func (h *MetricsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func (h *MetricsHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, details map[string]interface{}) {
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":      code,
			"message":   message,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
	if details != nil {
		errorResponse["error"].(map[string]interface{})["details"] = details
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}