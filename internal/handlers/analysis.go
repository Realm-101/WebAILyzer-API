package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	
	"github.com/gorilla/mux"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/services"
	"github.com/webailyzer/webailyzer-lite-api/internal/cache"
	"github.com/webailyzer/webailyzer-lite-api/internal/errors"
	"github.com/webailyzer/webailyzer-lite-api/internal/middleware"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AnalysisHandler handles analysis-related HTTP requests
type AnalysisHandler struct {
	analysisService services.AnalysisService
	cacheService    *cache.CacheService
	logger          *logrus.Logger
	errorHandler    *middleware.ErrorHandler
}

// NewAnalysisHandler creates a new analysis handler
func NewAnalysisHandler(analysisService services.AnalysisService, cacheService *cache.CacheService, logger *logrus.Logger) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
		cacheService:    cacheService,
		logger:          logger,
		errorHandler:    middleware.NewErrorHandler(logger),
	}
}

// AnalyzeURL handles POST /api/v1/analyze requests with enhanced validation and caching
func (h *AnalysisHandler) AnalyzeURL(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Parse and validate request
	var req models.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		h.errorHandler.HandleError(w, r, errors.BadRequestWithDetails("Invalid JSON in request body", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// Validate request fields
	if err := h.validateAnalysisRequest(&req); err != nil {
		h.logger.WithError(err).WithField("url", req.URL).Error("Request validation failed")
		h.errorHandler.HandleError(w, r, errors.ValidationFailed(err.Error(), map[string]interface{}{
			"url": req.URL,
		}))
		return
	}

	// Generate cache key for the request
	cacheKey := h.generateCacheKey(&req)
	
	// Try to get cached result if caching is enabled
	if h.cacheService != nil {
		var cachedResult models.AnalysisResult
		if err := h.cacheService.GetWithConfig(r.Context(), cacheKey, &cachedResult); err == nil {
			h.logger.WithFields(logrus.Fields{
				"url":       req.URL,
				"cache_key": cacheKey,
				"duration":  time.Since(startTime),
			}).Info("Returning cached analysis result")
			
			h.writeJSONResponse(w, http.StatusOK, &cachedResult)
			return
		}
	}

	// Perform analysis
	result, err := h.analysisService.AnalyzeURL(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"url":          req.URL,
			"workspace_id": req.WorkspaceID,
			"session_id":   req.SessionID,
		}).Error("Failed to analyze URL")
		
		// Return appropriate error based on error type
		errMsg := err.Error()
		if strings.Contains(errMsg, "invalid URL") || strings.Contains(errMsg, "malformed") {
			h.errorHandler.HandleError(w, r, errors.InvalidURL(req.URL, err))
		} else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded") {
			h.errorHandler.HandleError(w, r, errors.AnalysisTimeout(req.URL))
		} else if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "no such host") {
			h.errorHandler.HandleError(w, r, errors.ConnectionError(req.URL, err))
		} else {
			h.errorHandler.HandleError(w, r, errors.InternalErrorWithCause("Failed to analyze URL", err))
		}
		return
	}

	// Cache the result if caching is enabled
	if h.cacheService != nil {
		if err := h.cacheService.SetWithConfig(r.Context(), cacheKey, result, cache.AnalysisResultConfig); err != nil {
			h.logger.WithError(err).WithField("cache_key", cacheKey).Warn("Failed to cache analysis result")
		}
	}

	// Log successful analysis
	h.logger.WithFields(logrus.Fields{
		"analysis_id":  result.ID,
		"url":          req.URL,
		"workspace_id": req.WorkspaceID,
		"session_id":   req.SessionID,
		"duration":     time.Since(startTime),
	}).Info("Analysis completed successfully")

	h.writeJSONResponse(w, http.StatusOK, result)
}

// BatchAnalyze handles POST /api/v1/batch requests with enhanced validation
func (h *AnalysisHandler) BatchAnalyze(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Parse and validate request
	var req services.BatchAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode batch request body")
		h.errorHandler.HandleError(w, r, errors.BadRequestWithDetails("Invalid JSON in request body", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// Validate batch request
	if err := h.validateBatchAnalysisRequest(&req); err != nil {
		h.logger.WithError(err).WithField("workspace_id", req.WorkspaceID).Error("Batch request validation failed")
		h.errorHandler.HandleError(w, r, errors.ValidationFailed(err.Error(), map[string]interface{}{
			"workspace_id": req.WorkspaceID,
			"url_count":    len(req.URLs),
		}))
		return
	}

	// Perform batch analysis
	result, err := h.analysisService.BatchAnalyze(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"workspace_id": req.WorkspaceID,
			"url_count":    len(req.URLs),
		}).Error("Failed to perform batch analysis")
		
		if strings.Contains(err.Error(), "context canceled") || strings.Contains(err.Error(), "deadline exceeded") {
			h.errorHandler.HandleError(w, r, errors.NewAPIErrorWithDetails(errors.ErrCodeRequestTimeout, "Batch analysis request timed out", http.StatusRequestTimeout, map[string]interface{}{
				"url_count": len(req.URLs),
			}))
		} else {
			h.errorHandler.HandleError(w, r, errors.InternalErrorWithCause("Failed to perform batch analysis", err))
		}
		return
	}

	// Log successful batch analysis
	h.logger.WithFields(logrus.Fields{
		"batch_id":     result.BatchID,
		"workspace_id": req.WorkspaceID,
		"url_count":    len(req.URLs),
		"successful":   len(result.Results),
		"failed":       len(result.FailedURLs),
		"status":       result.Status,
		"duration":     time.Since(startTime),
	}).Info("Batch analysis completed")

	h.writeJSONResponse(w, http.StatusOK, result)
}

// GetAnalysisHistory handles GET /api/v1/analysis requests with enhanced filtering
func (h *AnalysisHandler) GetAnalysisHistory(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Parse and validate workspace ID
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		h.errorHandler.HandleError(w, r, errors.BadRequest("workspace_id query parameter is required"))
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		h.errorHandler.HandleError(w, r, errors.BadRequestWithDetails("Invalid workspace_id format", map[string]interface{}{
			"workspace_id": workspaceIDStr,
		}))
		return
	}

	// Parse filters from query parameters
	filters, err := h.parseAnalysisFilters(r)
	if err != nil {
		h.errorHandler.HandleError(w, r, errors.BadRequest(err.Error()))
		return
	}

	// Get analysis history
	results, err := h.analysisService.GetAnalysisHistory(r.Context(), workspaceID, filters)
	if err != nil {
		h.logger.WithError(err).WithField("workspace_id", workspaceID).Error("Failed to get analysis history")
		h.errorHandler.HandleError(w, r, errors.InternalErrorWithCause("Failed to retrieve analysis history", err))
		return
	}

	// Log successful retrieval
	h.logger.WithFields(logrus.Fields{
		"workspace_id":  workspaceID,
		"result_count":  len(results),
		"limit":         filters.Limit,
		"offset":        filters.Offset,
		"duration":      time.Since(startTime),
	}).Info("Analysis history retrieved successfully")

	// Prepare response with metadata
	response := map[string]interface{}{
		"results": results,
		"metadata": map[string]interface{}{
			"count":        len(results),
			"limit":        filters.Limit,
			"offset":       filters.Offset,
			"workspace_id": workspaceID,
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RegisterRoutes registers analysis routes with the router
func (h *AnalysisHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/analyze", h.AnalyzeURL).Methods("POST")
	router.HandleFunc("/api/v1/batch", h.BatchAnalyze).Methods("POST")
	router.HandleFunc("/api/v1/analysis", h.GetAnalysisHistory).Methods("GET")
}

// writeJSONResponse writes a JSON response
func (h *AnalysisHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// validateBatchAnalysisRequest validates the batch analysis request
func (h *AnalysisHandler) validateBatchAnalysisRequest(req *services.BatchAnalysisRequest) error {
	// Validate workspace ID
	if req.WorkspaceID == uuid.Nil {
		return fmt.Errorf("workspace_id is required")
	}
	
	// Validate URLs array
	if len(req.URLs) == 0 {
		return fmt.Errorf("at least one URL is required")
	}
	
	// Check batch size limits
	const maxBatchSize = 100
	if len(req.URLs) > maxBatchSize {
		return fmt.Errorf("batch size cannot exceed %d URLs", maxBatchSize)
	}
	
	// Validate each URL
	for i, urlStr := range req.URLs {
		if urlStr == "" {
			return fmt.Errorf("URL at index %d cannot be empty", i)
		}
		
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("invalid URL format at index %d: %w", i, err)
		}
		
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return fmt.Errorf("URL at index %d must use http or https scheme", i)
		}
		
		if parsedURL.Host == "" {
			return fmt.Errorf("URL at index %d must have a valid host", i)
		}
	}
	
	// Validate user agent length if provided
	if req.Options.UserAgent != "" && len(req.Options.UserAgent) > 500 {
		return fmt.Errorf("user_agent cannot exceed 500 characters")
	}
	
	return nil
}

// validateAnalysisRequest validates the analysis request
func (h *AnalysisHandler) validateAnalysisRequest(req *models.AnalysisRequest) error {
	// Validate URL
	if req.URL == "" {
		return fmt.Errorf("URL is required")
	}
	
	// Parse and validate URL format
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	// Check URL scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}
	
	// Check URL host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}
	
	// Validate workspace ID
	if req.WorkspaceID == uuid.Nil {
		return fmt.Errorf("workspace_id is required")
	}
	
	// Validate session ID if provided
	if req.SessionID != nil && *req.SessionID == uuid.Nil {
		return fmt.Errorf("session_id cannot be empty UUID")
	}
	
	// Validate user agent length if provided
	if req.Options.UserAgent != "" && len(req.Options.UserAgent) > 500 {
		return fmt.Errorf("user_agent cannot exceed 500 characters")
	}
	
	return nil
}

// generateCacheKey generates a cache key for the analysis request
func (h *AnalysisHandler) generateCacheKey(req *models.AnalysisRequest) string {
	// Create a deterministic cache key based on request parameters
	key := fmt.Sprintf("analysis:%s:%s:%t:%t:%t:%t:%s",
		req.URL,
		req.WorkspaceID.String(),
		req.Options.IncludePerformance,
		req.Options.IncludeSEO,
		req.Options.IncludeAccessibility,
		req.Options.IncludeSecurity,
		req.Options.UserAgent,
	)
	return key
}



// parseAnalysisFilters parses analysis filters from query parameters
func (h *AnalysisHandler) parseAnalysisFilters(r *http.Request) (*services.AnalysisFilters, error) {
	filters := &services.AnalysisFilters{
		Limit:  50, // Default limit
		Offset: 0,  // Default offset
	}

	// Parse session_id if provided
	if sessionIDStr := r.URL.Query().Get("session_id"); sessionIDStr != "" {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid session_id format: %s", sessionIDStr)
		}
		filters.SessionID = &sessionID
	}

	// Parse start_date if provided
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format, expected RFC3339: %s", startDateStr)
		}
		filters.StartDate = &startDate
	}

	// Parse end_date if provided
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format, expected RFC3339: %s", endDateStr)
		}
		filters.EndDate = &endDate
	}

	// Parse limit if provided
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid limit format: %s", limitStr)
		}
		if limit < 1 || limit > 1000 {
			return nil, fmt.Errorf("limit must be between 1 and 1000")
		}
		filters.Limit = limit
	}

	// Parse offset if provided
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, fmt.Errorf("invalid offset format: %s", offsetStr)
		}
		if offset < 0 {
			return nil, fmt.Errorf("offset cannot be negative")
		}
		filters.Offset = offset
	}

	return filters, nil
}

