package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/repositories"
)



// EventServiceImpl implements the EventService interface
type EventServiceImpl struct {
	eventRepo   repositories.EventRepository
	sessionRepo repositories.SessionRepository
	logger      *logrus.Logger
}

// NewEventService creates a new event service instance
func NewEventService(
	eventRepo repositories.EventRepository,
	sessionRepo repositories.SessionRepository,
	logger *logrus.Logger,
) EventService {
	return &EventServiceImpl{
		eventRepo:   eventRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// TrackEvents processes and stores event tracking data
func (s *EventServiceImpl) TrackEvents(ctx context.Context, req *models.EventTrackingRequest) error {
	startTime := time.Now()

	s.logger.WithFields(logrus.Fields{
		"session_id":   req.SessionID,
		"workspace_id": req.WorkspaceID,
		"event_count":  len(req.Events),
	}).Info("Processing event tracking request")

	// Validate and prepare events
	var validEvents []*models.Event
	duplicateMap := make(map[string]bool)

	for i, event := range req.Events {
		// Validate event
		if err := s.validateEvent(&event, req); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"event_index": i,
				"event_type":  event.EventType,
			}).Warn("Invalid event in batch, skipping")
			continue
		}

		// Check for duplicates based on event ID
		if event.ID != uuid.Nil {
			eventKey := event.ID.String()
			if duplicateMap[eventKey] {
				s.logger.WithFields(logrus.Fields{
					"event_id":   event.ID,
					"event_type": event.EventType,
				}).Warn("Duplicate event detected, skipping")
				continue
			}
			duplicateMap[eventKey] = true
		} else {
			// Generate ID if not provided
			event.ID = uuid.New()
		}

		// Set required fields
		event.SessionID = req.SessionID
		event.WorkspaceID = req.WorkspaceID
		event.CreatedAt = time.Now()

		// Set timestamp if not provided
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now()
		}

		validEvents = append(validEvents, &event)
	}

	if len(validEvents) == 0 {
		return fmt.Errorf("no valid events to process")
	}

	// Ensure session exists or create it
	session, err := s.sessionRepo.GetSessionByID(ctx, req.SessionID)
	if err != nil {
		// Session doesn't exist, create it
		session = &models.Session{
			ID:          req.SessionID,
			WorkspaceID: req.WorkspaceID,
			StartedAt:   time.Now(),
			PageViews:   0,
			EventsCount: 0,
		}

		if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
			s.logger.WithError(err).WithField("session_id", req.SessionID).Error("Failed to create session")
			return fmt.Errorf("failed to create session: %w", err)
		}

		s.logger.WithField("session_id", req.SessionID).Info("Created new session")
	}

	// Store events
	if err := s.eventRepo.CreateEvents(ctx, validEvents); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"session_id":  req.SessionID,
			"event_count": len(validEvents),
		}).Error("Failed to store events")
		return fmt.Errorf("failed to store events: %w", err)
	}

	// Update session statistics
	session.EventsCount += len(validEvents)
	
	// Count page views
	pageViews := 0
	for _, event := range validEvents {
		if event.EventType == "pageview" {
			pageViews++
		}
	}
	session.PageViews += pageViews

	// Update session
	if err := s.sessionRepo.UpdateSession(ctx, session); err != nil {
		s.logger.WithError(err).WithField("session_id", req.SessionID).Warn("Failed to update session statistics")
		// Don't fail the request if session update fails
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":   req.SessionID,
		"workspace_id": req.WorkspaceID,
		"events_stored": len(validEvents),
		"page_views":   pageViews,
		"duration_ms":  time.Since(startTime).Milliseconds(),
	}).Info("Event tracking completed successfully")

	return nil
}

// GetEvents retrieves events based on filters
func (s *EventServiceImpl) GetEvents(ctx context.Context, filters *EventFilters) ([]*models.Event, error) {
	s.logger.WithFields(logrus.Fields{
		"workspace_id": filters.WorkspaceID,
		"session_id":   filters.SessionID,
		"event_type":   filters.EventType,
		"limit":        filters.Limit,
		"offset":       filters.Offset,
	}).Debug("Retrieving events")

	// If session ID is specified, use session-based query
	if filters.SessionID != nil {
		return s.eventRepo.GetEventsBySession(ctx, *filters.SessionID)
	}

	// Use workspace-based query with time range
	startTime := time.Time{}
	endTime := time.Now()
	
	if filters.StartTime != nil {
		startTime = *filters.StartTime
	}
	if filters.EndTime != nil {
		endTime = *filters.EndTime
	}

	events, err := s.eventRepo.GetEventsByWorkspace(ctx, filters.WorkspaceID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve events: %w", err)
	}

	// Apply additional filters (event type, pagination)
	var filteredEvents []*models.Event
	for _, event := range events {
		if filters.EventType != nil && event.EventType != *filters.EventType {
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}

	// Apply pagination
	if filters.Offset > 0 && filters.Offset < len(filteredEvents) {
		filteredEvents = filteredEvents[filters.Offset:]
	}
	if filters.Limit > 0 && filters.Limit < len(filteredEvents) {
		filteredEvents = filteredEvents[:filters.Limit]
	}

	return filteredEvents, nil
}

// GetSessions retrieves sessions based on filters
func (s *EventServiceImpl) GetSessions(ctx context.Context, filters *SessionFilters) ([]*models.Session, error) {
	s.logger.WithFields(logrus.Fields{
		"workspace_id": filters.WorkspaceID,
		"user_id":      filters.UserID,
		"limit":        filters.Limit,
		"offset":       filters.Offset,
	}).Debug("Retrieving sessions")

	sessions, err := s.sessionRepo.GetSessionsByWorkspace(ctx, filters.WorkspaceID, filters.Limit, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sessions: %w", err)
	}

	// Apply additional filters
	var filteredSessions []*models.Session
	for _, session := range sessions {
		// Filter by user ID if specified
		if filters.UserID != nil {
			if session.UserID == nil || *session.UserID != *filters.UserID {
				continue
			}
		}

		// Filter by time range if specified
		if filters.StartTime != nil && session.StartedAt.Before(*filters.StartTime) {
			continue
		}
		if filters.EndTime != nil && session.StartedAt.After(*filters.EndTime) {
			continue
		}

		filteredSessions = append(filteredSessions, session)
	}

	return filteredSessions, nil
}

// validateEvent validates an individual event
func (s *EventServiceImpl) validateEvent(event *models.Event, req *models.EventTrackingRequest) error {
	// Validate event type
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	// Validate event type against allowed types
	allowedTypes := map[string]bool{
		"pageview":   true,
		"click":      true,
		"conversion": true,
		"custom":     true,
		"form_submit": true,
		"scroll":     true,
		"download":   true,
		"video_play": true,
		"video_pause": true,
		"search":     true,
	}

	if !allowedTypes[event.EventType] {
		return fmt.Errorf("invalid event_type: %s", event.EventType)
	}

	// Validate URL for pageview events
	if event.EventType == "pageview" && (event.URL == nil || *event.URL == "") {
		return fmt.Errorf("URL is required for pageview events")
	}

	// Validate properties size (prevent abuse)
	if event.Properties != nil {
		if len(event.Properties) > 50 {
			return fmt.Errorf("too many properties (max 50)")
		}

		// Check property value sizes
		for key, value := range event.Properties {
			if len(key) > 100 {
				return fmt.Errorf("property key too long (max 100 chars): %s", key)
			}
			
			if str, ok := value.(string); ok && len(str) > 1000 {
				return fmt.Errorf("property value too long (max 1000 chars) for key: %s", key)
			}
		}
	}

	return nil
}