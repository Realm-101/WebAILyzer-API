package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/services"
)

// EventHandler handles event-related HTTP requests
type EventHandler struct {
	eventService services.EventService
	logger       *logrus.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(eventService services.EventService, logger *logrus.Logger) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		logger:       logger,
	}
}

// TrackEvents handles POST /api/v1/events requests
func (h *EventHandler) TrackEvents(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse and validate request
	var req models.EventTrackingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode event tracking request")
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON in request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validateEventTrackingRequest(&req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id":   req.SessionID,
			"workspace_id": req.WorkspaceID,
		}).Error("Event tracking request validation failed")
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Apply rate limiting per workspace
	if err := h.checkRateLimit(r.Context(), req.WorkspaceID, len(req.Events)); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"workspace_id": req.WorkspaceID,
			"event_count":  len(req.Events),
		}).Warn("Rate limit exceeded for event tracking")
		h.writeErrorResponse(w, http.StatusTooManyRequests, "RATE_LIMITED", "Rate limit exceeded for event tracking", map[string]interface{}{
			"workspace_id": req.WorkspaceID,
			"retry_after":  60, // seconds
		})
		return
	}

	// Process events
	if err := h.eventService.TrackEvents(r.Context(), &req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id":   req.SessionID,
			"workspace_id": req.WorkspaceID,
			"event_count":  len(req.Events),
		}).Error("Failed to track events")

		// Return appropriate error based on error type
		if err.Error() == "no valid events to process" {
			h.writeErrorResponse(w, http.StatusBadRequest, "NO_VALID_EVENTS", "No valid events to process", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to track events", nil)
		}
		return
	}

	// Log successful event tracking
	h.logger.WithFields(logrus.Fields{
		"session_id":   req.SessionID,
		"workspace_id": req.WorkspaceID,
		"event_count":  len(req.Events),
		"duration_ms":  time.Since(startTime).Milliseconds(),
	}).Info("Events tracked successfully")

	// Return success response
	response := map[string]interface{}{
		"success":      true,
		"session_id":   req.SessionID,
		"events_count": len(req.Events),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetEvents handles GET /api/v1/events requests
func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse and validate workspace ID
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
	filters, err := h.parseEventFilters(r, workspaceID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Get events
	events, err := h.eventService.GetEvents(r.Context(), filters)
	if err != nil {
		h.logger.WithError(err).WithField("workspace_id", workspaceID).Error("Failed to get events")
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve events", map[string]interface{}{
			"workspace_id": workspaceID,
		})
		return
	}

	// Log successful retrieval
	h.logger.WithFields(logrus.Fields{
		"workspace_id": workspaceID,
		"event_count":  len(events),
		"duration_ms":  time.Since(startTime).Milliseconds(),
	}).Info("Events retrieved successfully")

	// Prepare response with metadata
	response := map[string]interface{}{
		"events": events,
		"metadata": map[string]interface{}{
			"count":        len(events),
			"workspace_id": workspaceID,
			"filters":      filters,
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetSessions handles GET /api/v1/sessions requests
func (h *EventHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse and validate workspace ID
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
	filters, err := h.parseSessionFilters(r, workspaceID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Get sessions
	sessions, err := h.eventService.GetSessions(r.Context(), filters)
	if err != nil {
		h.logger.WithError(err).WithField("workspace_id", workspaceID).Error("Failed to get sessions")
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve sessions", map[string]interface{}{
			"workspace_id": workspaceID,
		})
		return
	}

	// Log successful retrieval
	h.logger.WithFields(logrus.Fields{
		"workspace_id":  workspaceID,
		"session_count": len(sessions),
		"duration_ms":   time.Since(startTime).Milliseconds(),
	}).Info("Sessions retrieved successfully")

	// Prepare response with metadata
	response := map[string]interface{}{
		"sessions": sessions,
		"metadata": map[string]interface{}{
			"count":        len(sessions),
			"workspace_id": workspaceID,
			"filters":      filters,
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RegisterRoutes registers event routes with the router
func (h *EventHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/events", h.TrackEvents).Methods("POST")
	router.HandleFunc("/api/v1/events", h.GetEvents).Methods("GET")
	router.HandleFunc("/api/v1/sessions", h.GetSessions).Methods("GET")
}

// validateEventTrackingRequest validates the event tracking request
func (h *EventHandler) validateEventTrackingRequest(req *models.EventTrackingRequest) error {
	// Validate session ID
	if req.SessionID == uuid.Nil {
		return fmt.Errorf("session_id is required")
	}

	// Validate workspace ID
	if req.WorkspaceID == uuid.Nil {
		return fmt.Errorf("workspace_id is required")
	}

	// Validate events array
	if len(req.Events) == 0 {
		return fmt.Errorf("at least one event is required")
	}

	// Check batch size limits
	const maxEventBatchSize = 100
	if len(req.Events) > maxEventBatchSize {
		return fmt.Errorf("event batch size cannot exceed %d events", maxEventBatchSize)
	}

	// Basic validation of each event
	for i, event := range req.Events {
		if event.EventType == "" {
			return fmt.Errorf("event_type is required for event at index %d", i)
		}

		// Validate timestamp if provided
		if !event.Timestamp.IsZero() {
			// Check if timestamp is not too far in the future
			if event.Timestamp.After(time.Now().Add(time.Hour)) {
				return fmt.Errorf("event timestamp cannot be more than 1 hour in the future at index %d", i)
			}

			// Check if timestamp is not too far in the past (30 days)
			if event.Timestamp.Before(time.Now().Add(-30 * 24 * time.Hour)) {
				return fmt.Errorf("event timestamp cannot be more than 30 days in the past at index %d", i)
			}
		}
	}

	return nil
}

// checkRateLimit implements basic rate limiting per workspace
func (h *EventHandler) checkRateLimit(ctx context.Context, workspaceID uuid.UUID, eventCount int) error {
	// Simple rate limiting: max 1000 events per minute per workspace
	// In a production system, this would use Redis or similar
	const maxEventsPerMinute = 1000
	
	if eventCount > maxEventsPerMinute {
		return fmt.Errorf("event batch size exceeds rate limit")
	}

	// TODO: Implement proper rate limiting with Redis
	// For now, just check batch size
	return nil
}

// parseEventFilters parses event filters from query parameters
func (h *EventHandler) parseEventFilters(r *http.Request, workspaceID uuid.UUID) (*services.EventFilters, error) {
	filters := &services.EventFilters{
		WorkspaceID: workspaceID,
		Limit:       50, // Default limit
		Offset:      0,  // Default offset
	}

	// Parse session_id if provided
	if sessionIDStr := r.URL.Query().Get("session_id"); sessionIDStr != "" {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid session_id format: %s", sessionIDStr)
		}
		filters.SessionID = &sessionID
	}

	// Parse event_type if provided
	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		filters.EventType = &eventType
	}

	// Parse start_time if provided
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format, expected RFC3339: %s", startTimeStr)
		}
		filters.StartTime = &startTime
	}

	// Parse end_time if provided
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format, expected RFC3339: %s", endTimeStr)
		}
		filters.EndTime = &endTime
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

// parseSessionFilters parses session filters from query parameters
func (h *EventHandler) parseSessionFilters(r *http.Request, workspaceID uuid.UUID) (*services.SessionFilters, error) {
	filters := &services.SessionFilters{
		WorkspaceID: workspaceID,
		Limit:       50, // Default limit
		Offset:      0,  // Default offset
	}

	// Parse user_id if provided
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filters.UserID = &userID
	}

	// Parse start_time if provided
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format, expected RFC3339: %s", startTimeStr)
		}
		filters.StartTime = &startTime
	}

	// Parse end_time if provided
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format, expected RFC3339: %s", endTimeStr)
		}
		filters.EndTime = &endTime
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

// writeJSONResponse writes a JSON response
func (h *EventHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func (h *EventHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, details map[string]interface{}) {
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