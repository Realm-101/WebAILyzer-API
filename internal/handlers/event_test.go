package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/services"
)

// MockEventService is a mock implementation of EventService
type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) TrackEvents(ctx context.Context, req *models.EventTrackingRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockEventService) GetEvents(ctx context.Context, filters *services.EventFilters) ([]*models.Event, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Event), args.Error(1)
}

func (m *MockEventService) GetSessions(ctx context.Context, filters *services.SessionFilters) ([]*models.Session, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func TestEventHandler_TrackEvents_Success(t *testing.T) {
	// Setup
	mockService := new(MockEventService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	handler := NewEventHandler(mockService, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()
	eventID := uuid.New()

	// Mock request
	req := models.EventTrackingRequest{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		Events: []models.Event{
			{
				ID:        eventID,
				EventType: "pageview",
				URL:       stringPtr("https://example.com"),
				Timestamp: time.Now(),
				Properties: map[string]interface{}{
					"page_title": "Example Page",
					"referrer":   "https://google.com",
				},
			},
		},
	}

	mockService.On("TrackEvents", mock.Anything, mock.MatchedBy(func(r *models.EventTrackingRequest) bool {
		return r.SessionID == sessionID && r.WorkspaceID == workspaceID && len(r.Events) == 1
	})).Return(nil)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.TrackEvents(rr, httpReq)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, sessionID.String(), response["session_id"].(string))
	assert.Equal(t, float64(1), response["events_count"].(float64))

	mockService.AssertExpectations(t)
}

func TestEventHandler_TrackEvents_ValidationErrors(t *testing.T) {
	mockService := new(MockEventService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := NewEventHandler(mockService, logger)

	testCases := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Invalid JSON",
			request:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "Missing session ID",
			request: models.EventTrackingRequest{
				WorkspaceID: uuid.New(),
				Events: []models.Event{
					{EventType: "pageview"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "Missing workspace ID",
			request: models.EventTrackingRequest{
				SessionID: uuid.New(),
				Events: []models.Event{
					{EventType: "pageview"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "Empty events array",
			request: models.EventTrackingRequest{
				SessionID:   uuid.New(),
				WorkspaceID: uuid.New(),
				Events:      []models.Event{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "Too many events",
			request: models.EventTrackingRequest{
				SessionID:   uuid.New(),
				WorkspaceID: uuid.New(),
				Events:      make([]models.Event, 101), // Exceeds max batch size
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			if str, ok := tc.request.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tc.request)
				require.NoError(t, err)
			}

			httpReq := httptest.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.TrackEvents(rr, httpReq)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			errorObj, ok := response["error"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tc.expectedCode, errorObj["code"])
		})
	}
}

func TestEventHandler_TrackEvents_ServiceError(t *testing.T) {
	mockService := new(MockEventService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := NewEventHandler(mockService, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	req := models.EventTrackingRequest{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		Events: []models.Event{
			{
				EventType: "pageview",
				URL:       stringPtr("https://example.com"),
			},
		},
	}

	// Mock service error
	mockService.On("TrackEvents", mock.Anything, mock.Anything).Return(fmt.Errorf("database error"))

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.TrackEvents(rr, httpReq)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "INTERNAL_ERROR", errorObj["code"])

	mockService.AssertExpectations(t)
}

func TestEventHandler_GetEvents_Success(t *testing.T) {
	mockService := new(MockEventService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := NewEventHandler(mockService, logger)

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
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			SessionID:   sessionID,
			WorkspaceID: workspaceID,
			EventType:   "click",
			Timestamp:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}

	mockService.On("GetEvents", mock.Anything, mock.MatchedBy(func(f *services.EventFilters) bool {
		return f.WorkspaceID == workspaceID
	})).Return(expectedEvents, nil)

	url := fmt.Sprintf("/api/v1/events?workspace_id=%s&limit=10", workspaceID.String())
	httpReq := httptest.NewRequest("GET", url, nil)

	rr := httptest.NewRecorder()
	handler.GetEvents(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	events, ok := response["events"].([]interface{})
	require.True(t, ok)
	assert.Len(t, events, 2)

	metadata, ok := response["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), metadata["count"])
	assert.Equal(t, workspaceID.String(), metadata["workspace_id"])

	mockService.AssertExpectations(t)
}

func TestEventHandler_GetEvents_ValidationErrors(t *testing.T) {
	mockService := new(MockEventService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := NewEventHandler(mockService, logger)

	testCases := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Missing workspace_id",
			url:            "/api/v1/events",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name:           "Invalid workspace_id format",
			url:            "/api/v1/events?workspace_id=invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name:           "Invalid limit",
			url:            "/api/v1/events?workspace_id=" + uuid.New().String() + "&limit=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name:           "Limit too high",
			url:            "/api/v1/events?workspace_id=" + uuid.New().String() + "&limit=2000",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			httpReq := httptest.NewRequest("GET", tc.url, nil)

			rr := httptest.NewRecorder()
			handler.GetEvents(rr, httpReq)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			var response map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			errorObj, ok := response["error"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tc.expectedCode, errorObj["code"])
		})
	}
}

func TestEventHandler_GetSessions_Success(t *testing.T) {
	mockService := new(MockEventService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := NewEventHandler(mockService, logger)

	workspaceID := uuid.New()

	expectedSessions := []*models.Session{
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			StartedAt:   time.Now(),
			PageViews:   5,
			EventsCount: 10,
		},
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			StartedAt:   time.Now().Add(-time.Hour),
			PageViews:   3,
			EventsCount: 8,
		},
	}

	mockService.On("GetSessions", mock.Anything, mock.MatchedBy(func(f *services.SessionFilters) bool {
		return f.WorkspaceID == workspaceID
	})).Return(expectedSessions, nil)

	url := fmt.Sprintf("/api/v1/sessions?workspace_id=%s&limit=10", workspaceID.String())
	httpReq := httptest.NewRequest("GET", url, nil)

	rr := httptest.NewRecorder()
	handler.GetSessions(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	sessions, ok := response["sessions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, sessions, 2)

	metadata, ok := response["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), metadata["count"])
	assert.Equal(t, workspaceID.String(), metadata["workspace_id"])

	mockService.AssertExpectations(t)
}

func TestEventHandler_EventValidation_TimestampValidation(t *testing.T) {
	mockService := new(MockEventService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := NewEventHandler(mockService, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	testCases := []struct {
		name      string
		timestamp time.Time
		shouldErr bool
	}{
		{
			name:      "Valid current timestamp",
			timestamp: time.Now(),
			shouldErr: false,
		},
		{
			name:      "Valid past timestamp (1 hour ago)",
			timestamp: time.Now().Add(-time.Hour),
			shouldErr: false,
		},
		{
			name:      "Invalid future timestamp (2 hours ahead)",
			timestamp: time.Now().Add(2 * time.Hour),
			shouldErr: true,
		},
		{
			name:      "Invalid old timestamp (31 days ago)",
			timestamp: time.Now().Add(-31 * 24 * time.Hour),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := models.EventTrackingRequest{
				SessionID:   sessionID,
				WorkspaceID: workspaceID,
				Events: []models.Event{
					{
						EventType: "pageview",
						URL:       stringPtr("https://example.com"),
						Timestamp: tc.timestamp,
					},
				},
			}

			if !tc.shouldErr {
				mockService.On("TrackEvents", mock.Anything, mock.Anything).Return(nil).Once()
			}

			reqBody, _ := json.Marshal(req)
			httpReq := httptest.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.TrackEvents(rr, httpReq)

			if tc.shouldErr {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code)
			}
		})
	}

	mockService.AssertExpectations(t)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}