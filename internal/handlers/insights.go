package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// InsightsHandler handles insights-related HTTP requests
type InsightsHandler struct {
	insightsService services.InsightsService
	jobManager      *services.InsightsJobManager
	logger          *logrus.Logger
}

// NewInsightsHandler creates a new insights handler
func NewInsightsHandler(insightsService services.InsightsService, logger *logrus.Logger) *InsightsHandler {
	jobManager := services.NewInsightsJobManager(insightsService, logger)
	return &InsightsHandler{
		insightsService: insightsService,
		jobManager:      jobManager,
		logger:          logger,
	}
}

// GetInsights handles GET /api/v1/insights requests
func (h *InsightsHandler) GetInsights(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse workspace_id from query parameters
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "workspace_id query parameter is required", nil)
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid workspace_id format", map[string]interface{}{
			"workspace_id": workspaceIDStr,
		})
		return
	}

	// Parse filters from query parameters
	filters, err := h.parseInsightFilters(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Get insights from service
	insights, err := h.insightsService.GetInsights(r.Context(), workspaceID, filters)
	if err != nil {
		h.logger.WithError(err).WithField("workspace_id", workspaceID).Error("Failed to get insights")
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve insights", map[string]interface{}{
			"workspace_id": workspaceID,
		})
		return
	}

	// Log successful retrieval
	h.logger.WithFields(logrus.Fields{
		"workspace_id": workspaceID,
		"result_count": len(insights),
		"limit":        filters.Limit,
		"offset":       filters.Offset,
		"duration":     time.Since(startTime),
	}).Info("Insights retrieved successfully")

	// Prepare response with metadata
	response := map[string]interface{}{
		"insights": insights,
		"pagination": map[string]interface{}{
			"limit":  filters.Limit,
			"offset": filters.Offset,
			"count":  len(insights),
		},
		"metadata": map[string]interface{}{
			"workspace_id": workspaceID,
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateInsightStatus handles PUT /api/v1/insights/{id}/status requests
func (h *InsightsHandler) UpdateInsightStatus(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse insight ID from URL path
	vars := mux.Vars(r)
	insightIDStr := vars["id"]
	if insightIDStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "insight ID is required", nil)
		return
	}

	insightID, err := uuid.Parse(insightIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid insight ID format", map[string]interface{}{
			"insight_id": insightIDStr,
		})
		return
	}

	// Parse request body
	var req struct {
		Status models.InsightStatus `json:"status" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode status update request")
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON in request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Validate status
	if err := h.validateInsightStatus(req.Status); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), map[string]interface{}{
			"status": req.Status,
		})
		return
	}

	// Update insight status
	err = h.insightsService.UpdateInsightStatus(r.Context(), insightID, req.Status)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"insight_id": insightID,
			"status":     req.Status,
		}).Error("Failed to update insight status")
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update insight status", map[string]interface{}{
			"insight_id": insightID,
		})
		return
	}

	// Log successful update
	h.logger.WithFields(logrus.Fields{
		"insight_id": insightID,
		"status":     req.Status,
		"duration":   time.Since(startTime),
	}).Info("Insight status updated successfully")

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"insight_id": insightID,
		"status": req.Status,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GenerateInsights handles POST /api/v1/insights/generate requests
func (h *InsightsHandler) GenerateInsights(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse request body
	var req struct {
		WorkspaceID uuid.UUID `json:"workspace_id" validate:"required"`
		Async       bool      `json:"async,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode insights generation request")
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON in request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Validate workspace ID
	if req.WorkspaceID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "workspace_id is required", nil)
		return
	}

	// Handle async generation
	if req.Async {
		job, err := h.jobManager.StartInsightGeneration(r.Context(), req.WorkspaceID)
		if err != nil {
			h.logger.WithError(err).WithField("workspace_id", req.WorkspaceID).Error("Failed to start insight generation job")
			h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start insight generation", map[string]interface{}{
				"workspace_id": req.WorkspaceID,
			})
			return
		}

		// Return job information
		response := map[string]interface{}{
			"success":      true,
			"job_id":       job.ID,
			"workspace_id": req.WorkspaceID,
			"status":       job.Status,
			"created_at":   job.CreatedAt.UTC().Format(time.RFC3339),
		}

		h.writeJSONResponse(w, http.StatusAccepted, response)
		return
	}

	// Synchronous generation
	insights, err := h.insightsService.GenerateInsights(r.Context(), req.WorkspaceID)
	if err != nil {
		h.logger.WithError(err).WithField("workspace_id", req.WorkspaceID).Error("Failed to generate insights")
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate insights", map[string]interface{}{
			"workspace_id": req.WorkspaceID,
		})
		return
	}

	// Log successful generation
	h.logger.WithFields(logrus.Fields{
		"workspace_id":   req.WorkspaceID,
		"insights_count": len(insights),
		"duration":       time.Since(startTime),
	}).Info("Insights generated successfully")

	// Return response
	response := map[string]interface{}{
		"success":            true,
		"workspace_id":       req.WorkspaceID,
		"insights_generated": len(insights),
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetJobStatus handles GET /api/v1/insights/jobs/{id} requests
func (h *InsightsHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	// Parse job ID from URL path
	vars := mux.Vars(r)
	jobIDStr := vars["id"]
	if jobIDStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "job ID is required", nil)
		return
	}

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid job ID format", map[string]interface{}{
			"job_id": jobIDStr,
		})
		return
	}

	// Get job status
	job, exists := h.jobManager.GetJob(jobID)
	if !exists {
		h.writeErrorResponse(w, http.StatusNotFound, "JOB_NOT_FOUND", "Job not found", map[string]interface{}{
			"job_id": jobID,
		})
		return
	}

	h.writeJSONResponse(w, http.StatusOK, job)
}

// RegisterRoutes registers insights routes with the router
func (h *InsightsHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/insights", h.GetInsights).Methods("GET")
	router.HandleFunc("/api/v1/insights/{id}/status", h.UpdateInsightStatus).Methods("PUT")
	router.HandleFunc("/api/v1/insights/generate", h.GenerateInsights).Methods("POST")
	router.HandleFunc("/api/v1/insights/jobs/{id}", h.GetJobStatus).Methods("GET")
}

// parseInsightFilters parses insight filters from query parameters
func (h *InsightsHandler) parseInsightFilters(r *http.Request) (*services.InsightFilters, error) {
	filters := &services.InsightFilters{
		Limit:  50, // default limit
		Offset: 0,  // default offset
	}

	// Parse status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.InsightStatus(statusStr)
		if err := h.validateInsightStatus(status); err != nil {
			return nil, err
		}
		filters.Status = &status
	}

	// Parse type filter
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		insightType := models.InsightType(typeStr)
		if err := h.validateInsightType(insightType); err != nil {
			return nil, err
		}
		filters.Type = &insightType
	}

	// Parse priority filter
	if priorityStr := r.URL.Query().Get("priority"); priorityStr != "" {
		priority := models.Priority(priorityStr)
		if err := h.validatePriority(priority); err != nil {
			return nil, err
		}
		filters.Priority = &priority
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			return nil, fmt.Errorf("invalid limit: must be a positive integer between 1 and 100")
		}
		filters.Limit = limit
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return nil, fmt.Errorf("invalid offset: must be a non-negative integer")
		}
		filters.Offset = offset
	}

	return filters, nil
}

// validateInsightStatus validates an insight status
func (h *InsightsHandler) validateInsightStatus(status models.InsightStatus) error {
	switch status {
	case models.InsightStatusPending, models.InsightStatusApplied, models.InsightStatusDismissed:
		return nil
	default:
		return fmt.Errorf("invalid status: must be one of 'pending', 'applied', or 'dismissed'")
	}
}

// validateInsightType validates an insight type
func (h *InsightsHandler) validateInsightType(insightType models.InsightType) error {
	switch insightType {
	case models.InsightTypePerformanceBottleneck, models.InsightTypeConversionFunnel,
		 models.InsightTypeSEOOptimization, models.InsightTypeAccessibilityIssue,
		 models.InsightTypeSecurityVulnerability:
		return nil
	default:
		return fmt.Errorf("invalid insight type: must be one of 'performance_bottleneck', 'conversion_funnel', 'seo_optimization', 'accessibility_issue', or 'security_vulnerability'")
	}
}

// validatePriority validates a priority
func (h *InsightsHandler) validatePriority(priority models.Priority) error {
	switch priority {
	case models.PriorityLow, models.PriorityMedium, models.PriorityHigh, models.PriorityCritical:
		return nil
	default:
		return fmt.Errorf("invalid priority: must be one of 'low', 'medium', 'high', or 'critical'")
	}
}

// writeJSONResponse writes a JSON response
func (h *InsightsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func (h *InsightsHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, details map[string]interface{}) {
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