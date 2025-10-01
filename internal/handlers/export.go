package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/webailyzer/webailyzer-lite-api/internal/middleware"
	"github.com/webailyzer/webailyzer-lite-api/internal/services"
	"github.com/sirupsen/logrus"
)

// ExportHandler handles data export requests
type ExportHandler struct {
	exportService services.ExportService
	logger        *logrus.Logger
}

// NewExportHandler creates a new export handler
func NewExportHandler(exportService services.ExportService, logger *logrus.Logger) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
		logger:        logger,
	}
}

// RegisterRoutes registers the export routes
func (h *ExportHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v1/export/analysis/csv", h.ExportAnalysisCSV).Methods("GET")
	router.HandleFunc("/v1/export/analysis/json", h.ExportAnalysisJSON).Methods("GET")
	router.HandleFunc("/v1/export/metrics/csv", h.ExportMetricsCSV).Methods("GET")
	router.HandleFunc("/v1/export/metrics/json", h.ExportMetricsJSON).Methods("GET")
	router.HandleFunc("/v1/export/sessions/csv", h.ExportSessionsCSV).Methods("GET")
	router.HandleFunc("/v1/export/sessions/json", h.ExportSessionsJSON).Methods("GET")
	router.HandleFunc("/v1/export/events/csv", h.ExportEventsCSV).Methods("GET")
	router.HandleFunc("/v1/export/events/json", h.ExportEventsJSON).Methods("GET")
}

// ExportAnalysisCSV exports analysis results in CSV format
func (h *ExportHandler) ExportAnalysisCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceIDStr, ok := middleware.GetWorkspaceID(ctx)
	if !ok {
		h.logger.Error("Workspace ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid workspace ID format")
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	req, err := h.parseAnalysisExportRequest(r, workspaceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse analysis export request")
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"analysis_export_%s.csv\"", time.Now().Format("2006-01-02")))

	if err := h.exportService.ExportAnalysisResultsCSV(ctx, req, w); err != nil {
		h.logger.WithError(err).Error("Failed to export analysis results as CSV")
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}
}

// ExportAnalysisJSON exports analysis results in JSON format
func (h *ExportHandler) ExportAnalysisJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceIDStr, ok := middleware.GetWorkspaceID(ctx)
	if !ok {
		h.logger.Error("Workspace ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid workspace ID format")
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	req, err := h.parseAnalysisExportRequest(r, workspaceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse analysis export request")
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// Set JSON headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"analysis_export_%s.json\"", time.Now().Format("2006-01-02")))

	if err := h.exportService.ExportAnalysisResultsJSON(ctx, req, w); err != nil {
		h.logger.WithError(err).Error("Failed to export analysis results as JSON")
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}
}

// ExportMetricsCSV exports metrics in CSV format
func (h *ExportHandler) ExportMetricsCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID, err := h.getWorkspaceID(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workspace ID")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req, err := h.parseMetricsExportRequest(r, workspaceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse metrics export request")
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"metrics_export_%s.csv\"", time.Now().Format("2006-01-02")))

	if err := h.exportService.ExportMetricsCSV(ctx, req, w); err != nil {
		h.logger.WithError(err).Error("Failed to export metrics as CSV")
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}
}

// ExportMetricsJSON exports metrics in JSON format
func (h *ExportHandler) ExportMetricsJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID, err := h.getWorkspaceID(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workspace ID")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req, err := h.parseMetricsExportRequest(r, workspaceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse metrics export request")
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// Set JSON headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"metrics_export_%s.json\"", time.Now().Format("2006-01-02")))

	if err := h.exportService.ExportMetricsJSON(ctx, req, w); err != nil {
		h.logger.WithError(err).Error("Failed to export metrics as JSON")
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}
}

// ExportSessionsCSV exports sessions in CSV format
func (h *ExportHandler) ExportSessionsCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID, err := h.getWorkspaceID(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workspace ID")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req, err := h.parseSessionExportRequest(r, workspaceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse session export request")
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sessions_export_%s.csv\"", time.Now().Format("2006-01-02")))

	if err := h.exportService.ExportSessionsCSV(ctx, req, w); err != nil {
		h.logger.WithError(err).Error("Failed to export sessions as CSV")
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}
}

// ExportSessionsJSON exports sessions in JSON format
func (h *ExportHandler) ExportSessionsJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID, err := h.getWorkspaceID(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workspace ID")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req, err := h.parseSessionExportRequest(r, workspaceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse session export request")
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// Set JSON headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sessions_export_%s.json\"", time.Now().Format("2006-01-02")))

	if err := h.exportService.ExportSessionsJSON(ctx, req, w); err != nil {
		h.logger.WithError(err).Error("Failed to export sessions as JSON")
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}
}

// ExportEventsCSV exports events in CSV format
func (h *ExportHandler) ExportEventsCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID, err := h.getWorkspaceID(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workspace ID")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req, err := h.parseEventExportRequest(r, workspaceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse event export request")
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"events_export_%s.csv\"", time.Now().Format("2006-01-02")))

	if err := h.exportService.ExportEventsCSV(ctx, req, w); err != nil {
		h.logger.WithError(err).Error("Failed to export events as CSV")
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}
}

// ExportEventsJSON exports events in JSON format
func (h *ExportHandler) ExportEventsJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID, err := h.getWorkspaceID(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workspace ID")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req, err := h.parseEventExportRequest(r, workspaceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse event export request")
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// Set JSON headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"events_export_%s.json\"", time.Now().Format("2006-01-02")))

	if err := h.exportService.ExportEventsJSON(ctx, req, w); err != nil {
		h.logger.WithError(err).Error("Failed to export events as JSON")
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}
}

// Helper methods for parsing request parameters

func (h *ExportHandler) parseAnalysisExportRequest(r *http.Request, workspaceID uuid.UUID) (*services.ExportRequest, error) {
	req := &services.ExportRequest{
		WorkspaceID: workspaceID,
	}

	// Parse optional parameters
	if sessionIDStr := r.URL.Query().Get("session_id"); sessionIDStr != "" {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid session_id: %w", err)
		}
		req.SessionID = &sessionID
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format, use RFC3339: %w", err)
		}
		req.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format, use RFC3339: %w", err)
		}
		req.EndDate = &endDate
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return nil, fmt.Errorf("invalid limit: must be a positive integer")
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return nil, fmt.Errorf("invalid offset: must be a positive integer")
		}
		req.Offset = offset
	}

	return req, nil
}

func (h *ExportHandler) parseMetricsExportRequest(r *http.Request, workspaceID uuid.UUID) (*services.MetricsExportRequest, error) {
	req := &services.MetricsExportRequest{
		WorkspaceID: workspaceID,
	}

	// Parse required parameters
	startDateStr := r.URL.Query().Get("start_date")
	if startDateStr == "" {
		return nil, fmt.Errorf("start_date is required")
	}
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, use RFC3339: %w", err)
	}
	req.StartDate = startDate

	endDateStr := r.URL.Query().Get("end_date")
	if endDateStr == "" {
		return nil, fmt.Errorf("end_date is required")
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format, use RFC3339: %w", err)
	}
	req.EndDate = endDate

	return req, nil
}

func (h *ExportHandler) parseSessionExportRequest(r *http.Request, workspaceID uuid.UUID) (*services.SessionExportRequest, error) {
	req := &services.SessionExportRequest{
		WorkspaceID: workspaceID,
	}

	// Parse optional parameters
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		req.UserID = &userID
	}

	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format, use RFC3339: %w", err)
		}
		req.StartTime = &startTime
	}

	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format, use RFC3339: %w", err)
		}
		req.EndTime = &endTime
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return nil, fmt.Errorf("invalid limit: must be a positive integer")
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return nil, fmt.Errorf("invalid offset: must be a positive integer")
		}
		req.Offset = offset
	}

	return req, nil
}

func (h *ExportHandler) parseEventExportRequest(r *http.Request, workspaceID uuid.UUID) (*services.EventExportRequest, error) {
	req := &services.EventExportRequest{
		WorkspaceID: workspaceID,
	}

	// Parse optional parameters
	if sessionIDStr := r.URL.Query().Get("session_id"); sessionIDStr != "" {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid session_id: %w", err)
		}
		req.SessionID = &sessionID
	}

	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		req.EventType = &eventType
	}

	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format, use RFC3339: %w", err)
		}
		req.StartTime = &startTime
	}

	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format, use RFC3339: %w", err)
		}
		req.EndTime = &endTime
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return nil, fmt.Errorf("invalid limit: must be a positive integer")
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return nil, fmt.Errorf("invalid offset: must be a positive integer")
		}
		req.Offset = offset
	}

	return req, nil
}

// Helper method to extract and validate workspace ID from context
func (h *ExportHandler) getWorkspaceID(ctx context.Context) (uuid.UUID, error) {
	workspaceIDStr, ok := middleware.GetWorkspaceID(ctx)
	if !ok {
		return uuid.Nil, fmt.Errorf("workspace ID not found in context")
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid workspace ID format: %w", err)
	}

	return workspaceID, nil
}