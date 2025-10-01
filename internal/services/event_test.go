package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
)

// MockEventRepository is a mock implementation of EventRepository
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) CreateEvents(ctx context.Context, events []*models.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockEventRepository) GetEventsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.Event, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Event), args.Error(1)
}

func (m *MockEventRepository) GetEventsByWorkspace(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) ([]*models.Event, error) {
	args := m.Called(ctx, workspaceID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Event), args.Error(1)
}

func (m *MockEventRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.EventFilters) ([]*models.Event, error) {
	args := m.Called(ctx, workspaceID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Event), args.Error(1)
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSessionsByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.Session, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.SessionFilters) ([]*models.Session, error) {
	args := m.Called(ctx, workspaceID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func TestEventService_TrackEvents_NewSession(t *testing.T) {
	// Setup mocks
	mockEventRepo := new(MockEventRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewEventService(mockEventRepo, mockSessionRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	req := &models.EventTrackingRequest{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		Events: []models.Event{
			{
				ID:        uuid.New(),
				EventType: "pageview",
				URL:       stringPtr("https://example.com"),
				Timestamp: time.Now(),
				Properties: map[string]interface{}{
					"page_title": "Example Page",
				},
			},
			{
				ID:        uuid.New(),
				EventType: "click",
				Timestamp: time.Now(),
				Properties: map[string]interface{}{
					"element": "button",
				},
			},
		},
	}

	// Mock session not found (will create new session)
	mockSessionRepo.On("GetSessionByID", mock.Anything, sessionID).Return(nil, assert.AnError)
	mockSessionRepo.On("CreateSession", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.ID == sessionID && s.WorkspaceID == workspaceID
	})).Return(nil)

	// Mock event creation
	mockEventRepo.On("CreateEvents", mock.Anything, mock.MatchedBy(func(events []*models.Event) bool {
		return len(events) == 2 && events[0].SessionID == sessionID && events[1].SessionID == sessionID
	})).Return(nil)

	// Mock session update
	mockSessionRepo.On("UpdateSession", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.ID == sessionID && s.EventsCount == 2 && s.PageViews == 1
	})).Return(nil)

	// Execute
	err := service.TrackEvents(context.Background(), req)

	// Assert
	require.NoError(t, err)

	mockEventRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestEventService_TrackEvents_ExistingSession(t *testing.T) {
	// Setup mocks
	mockEventRepo := new(MockEventRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewEventService(mockEventRepo, mockSessionRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	existingSession := &models.Session{
		ID:          sessionID,
		WorkspaceID: workspaceID,
		StartedAt:   time.Now().Add(-time.Hour),
		PageViews:   3,
		EventsCount: 5,
	}

	req := &models.EventTrackingRequest{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		Events: []models.Event{
			{
				ID:        uuid.New(),
				EventType: "pageview",
				URL:       stringPtr("https://example.com/page2"),
				Timestamp: time.Now(),
			},
		},
	}

	// Mock existing session found
	mockSessionRepo.On("GetSessionByID", mock.Anything, sessionID).Return(existingSession, nil)

	// Mock event creation
	mockEventRepo.On("CreateEvents", mock.Anything, mock.MatchedBy(func(events []*models.Event) bool {
		return len(events) == 1 && events[0].SessionID == sessionID
	})).Return(nil)

	// Mock session update
	mockSessionRepo.On("UpdateSession", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.ID == sessionID && s.EventsCount == 6 && s.PageViews == 4
	})).Return(nil)

	// Execute
	err := service.TrackEvents(context.Background(), req)

	// Assert
	require.NoError(t, err)

	mockEventRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestEventService_TrackEvents_EventValidation(t *testing.T) {
	// Setup mocks
	mockEventRepo := new(MockEventRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewEventService(mockEventRepo, mockSessionRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	req := &models.EventTrackingRequest{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		Events: []models.Event{
			{
				ID:        uuid.New(),
				EventType: "pageview",
				URL:       stringPtr("https://example.com"),
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New(),
				EventType: "", // Invalid: empty event type
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New(),
				EventType: "invalid_type", // Invalid: not in allowed types
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New(),
				EventType: "pageview",
				// Invalid: pageview without URL
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New(),
				EventType: "click",
				Timestamp: time.Now(),
			},
		},
	}

	// Mock session not found (will create new session)
	mockSessionRepo.On("GetSessionByID", mock.Anything, sessionID).Return(nil, assert.AnError)
	mockSessionRepo.On("CreateSession", mock.Anything, mock.Anything).Return(nil)

	// Mock event creation - should only receive 2 valid events
	mockEventRepo.On("CreateEvents", mock.Anything, mock.MatchedBy(func(events []*models.Event) bool {
		return len(events) == 2 // Only valid events should be processed
	})).Return(nil)

	// Mock session update
	mockSessionRepo.On("UpdateSession", mock.Anything, mock.Anything).Return(nil)

	// Execute
	err := service.TrackEvents(context.Background(), req)

	// Assert
	require.NoError(t, err)

	mockEventRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestEventService_TrackEvents_DuplicateDetection(t *testing.T) {
	// Setup mocks
	mockEventRepo := new(MockEventRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewEventService(mockEventRepo, mockSessionRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()
	duplicateEventID := uuid.New()

	req := &models.EventTrackingRequest{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		Events: []models.Event{
			{
				ID:        duplicateEventID,
				EventType: "pageview",
				URL:       stringPtr("https://example.com"),
				Timestamp: time.Now(),
			},
			{
				ID:        duplicateEventID, // Duplicate ID
				EventType: "click",
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New(),
				EventType: "conversion",
				Timestamp: time.Now(),
			},
		},
	}

	// Mock session not found (will create new session)
	mockSessionRepo.On("GetSessionByID", mock.Anything, sessionID).Return(nil, assert.AnError)
	mockSessionRepo.On("CreateSession", mock.Anything, mock.Anything).Return(nil)

	// Mock event creation - should only receive 2 events (duplicate removed)
	mockEventRepo.On("CreateEvents", mock.Anything, mock.MatchedBy(func(events []*models.Event) bool {
		return len(events) == 2 // Duplicate should be removed
	})).Return(nil)

	// Mock session update
	mockSessionRepo.On("UpdateSession", mock.Anything, mock.Anything).Return(nil)

	// Execute
	err := service.TrackEvents(context.Background(), req)

	// Assert
	require.NoError(t, err)

	mockEventRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestEventService_TrackEvents_NoValidEvents(t *testing.T) {
	// Setup mocks
	mockEventRepo := new(MockEventRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewEventService(mockEventRepo, mockSessionRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	req := &models.EventTrackingRequest{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		Events: []models.Event{
			{
				ID:        uuid.New(),
				EventType: "", // Invalid: empty event type
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New(),
				EventType: "invalid_type", // Invalid: not in allowed types
				Timestamp: time.Now(),
			},
		},
	}

	// Execute
	err := service.TrackEvents(context.Background(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid events to process")

	// No repository calls should be made
	mockEventRepo.AssertNotCalled(t, "CreateEvents")
	mockSessionRepo.AssertNotCalled(t, "CreateSession")
	mockSessionRepo.AssertNotCalled(t, "UpdateSession")
}

func TestEventService_GetEvents_BySession(t *testing.T) {
	// Setup mocks
	mockEventRepo := new(MockEventRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewEventService(mockEventRepo, mockSessionRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	expectedEvents := []*models.Event{
		{
			ID:          uuid.New(),
			SessionID:   sessionID,
			WorkspaceID: workspaceID,
			EventType:   "pageview",
			URL:         stringPtr("https://example.com"),
			Timestamp:   time.Now(),
		},
		{
			ID:          uuid.New(),
			SessionID:   sessionID,
			WorkspaceID: workspaceID,
			EventType:   "click",
			Timestamp:   time.Now(),
		},
	}

	filters := &EventFilters{
		WorkspaceID: workspaceID,
		SessionID:   &sessionID,
		Limit:       50,
		Offset:      0,
	}

	mockEventRepo.On("GetEventsBySession", mock.Anything, sessionID).Return(expectedEvents, nil)

	// Execute
	events, err := service.GetEvents(context.Background(), filters)

	// Assert
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, expectedEvents, events)

	mockEventRepo.AssertExpectations(t)
}

func TestEventService_GetEvents_ByWorkspace(t *testing.T) {
	// Setup mocks
	mockEventRepo := new(MockEventRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewEventService(mockEventRepo, mockSessionRepo, logger)

	workspaceID := uuid.New()

	expectedEvents := []*models.Event{
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			EventType:   "pageview",
			Timestamp:   time.Now(),
		},
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			EventType:   "click",
			Timestamp:   time.Now(),
		},
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			EventType:   "conversion",
			Timestamp:   time.Now(),
		},
	}

	filters := &EventFilters{
		WorkspaceID: workspaceID,
		EventType:   stringPtr("pageview"), // Filter by event type
		Limit:       50,
		Offset:      0,
	}

	mockEventRepo.On("GetEventsByWorkspace", mock.Anything, workspaceID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(expectedEvents, nil)

	// Execute
	events, err := service.GetEvents(context.Background(), filters)

	// Assert
	require.NoError(t, err)
	assert.Len(t, events, 1) // Only pageview events should be returned
	assert.Equal(t, "pageview", events[0].EventType)

	mockEventRepo.AssertExpectations(t)
}

func TestEventService_GetSessions_Success(t *testing.T) {
	// Setup mocks
	mockEventRepo := new(MockEventRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewEventService(mockEventRepo, mockSessionRepo, logger)

	workspaceID := uuid.New()
	userID := "user123"

	expectedSessions := []*models.Session{
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			UserID:      &userID,
			StartedAt:   time.Now(),
			PageViews:   5,
			EventsCount: 10,
		},
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			UserID:      stringPtr("user456"),
			StartedAt:   time.Now().Add(-time.Hour),
			PageViews:   3,
			EventsCount: 8,
		},
	}

	filters := &SessionFilters{
		WorkspaceID: workspaceID,
		UserID:      &userID, // Filter by user ID
		Limit:       50,
		Offset:      0,
	}

	mockSessionRepo.On("GetSessionsByWorkspace", mock.Anything, workspaceID, 50, 0).Return(expectedSessions, nil)

	// Execute
	sessions, err := service.GetSessions(context.Background(), filters)

	// Assert
	require.NoError(t, err)
	assert.Len(t, sessions, 1) // Only sessions for user123 should be returned
	assert.Equal(t, userID, *sessions[0].UserID)

	mockSessionRepo.AssertExpectations(t)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}