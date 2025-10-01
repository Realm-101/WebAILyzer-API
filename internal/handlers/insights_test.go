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

	"github.com/gorilla/mux"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockInsightsService is a mock implementation of InsightsService
type MockInsightsService struct {
	mock.Mock
}

func (m *MockInsightsService) GenerateInsights(ctx context.Context, workspaceID uuid.UUID) ([]*models.Insight, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]*models.Insight), args.Error(1)
}

func (m *MockInsightsService) GetInsights(ctx context.Context, workspaceID uuid.UUID, filters *services.InsightFilters) ([]*models.Insight, error) {
	args := m.Called(ctx, workspaceID, filters)
	return args.Get(0).([]*models.Insight), args.Error(1)
}

func (m *MockInsightsService) UpdateInsightStatus(ctx context.Context, insightID uuid.UUID, status models.InsightStatus) error {
	args := m.Called(ctx, insightID, status)
	return args.Error(0)
}

func TestInsightsHandler_GetInsights(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockInsightsService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful insights retrieval",
			queryParams: "workspace_id=550e8400-e29b-41d4-a716-446655440000&limit=10&offset=0",
			mockSetup: func(m *MockInsightsService) {
				workspaceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				insights := []*models.Insight{
					{
						ID:          uuid.New(),
						WorkspaceID: workspaceID,
						InsightType: models.InsightTypePerformanceBottleneck,
						Priority:    models.PriorityHigh,
						Title:       "Test Insight",
						Status:      models.InsightStatusPending,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				m.On("GetInsights", mock.Anything, workspaceID, mock.MatchedBy(func(filters *services.InsightFilters) bool {
					return filters.Limit == 10 && filters.Offset == 0
				})).Return(insights, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful insights retrieval with status filter",
			queryParams: "workspace_id=550e8400-e29b-41d4-a716-446655440000&status=pending&limit=5",
			mockSetup: func(m *MockInsightsService) {
				workspaceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				insights := []*models.Insight{}
				m.On("GetInsights", mock.Anything, workspaceID, mock.MatchedBy(func(filters *services.InsightFilters) bool {
					return filters.Status != nil && *filters.Status == models.InsightStatusPending && filters.Limit == 5
				})).Return(insights, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing workspace_id",
			queryParams:    "",
			mockSetup:      func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace_id query parameter is required",
		},
		{
			name:           "invalid workspace_id format",
			queryParams:    "workspace_id=invalid-uuid",
			mockSetup:      func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid workspace_id format",
		},
		{
			name:        "invalid status filter",
			queryParams: "workspace_id=550e8400-e29b-41d4-a716-446655440000&status=invalid",
			mockSetup:   func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid status: must be one of 'pending', 'applied', or 'dismissed'",
		},
		{
			name:        "invalid limit",
			queryParams: "workspace_id=550e8400-e29b-41d4-a716-446655440000&limit=0",
			mockSetup:   func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid limit: must be a positive integer between 1 and 100",
		},
		{
			name:        "service error",
			queryParams: "workspace_id=550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(m *MockInsightsService) {
				workspaceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				m.On("GetInsights", mock.Anything, workspaceID, mock.Anything).Return([]*models.Insight{}, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to retrieve insights",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockInsightsService)
			tt.mockSetup(mockService)

			handler := NewInsightsHandler(mockService, logger)

			req, err := http.NewRequest("GET", "/api/v1/insights?"+tt.queryParams, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.GetInsights(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"].(map[string]interface{})["message"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "insights")
				assert.Contains(t, response, "pagination")
				assert.Contains(t, response, "metadata")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInsightsHandler_UpdateInsightStatus(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	tests := []struct {
		name           string
		insightID      string
		requestBody    interface{}
		mockSetup      func(*MockInsightsService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful status update",
			insightID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody: map[string]string{
				"status": "applied",
			},
			mockSetup: func(m *MockInsightsService) {
				insightID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				m.On("UpdateInsightStatus", mock.Anything, insightID, models.InsightStatusApplied).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "invalid insight ID",
			insightID: "invalid-uuid",
			requestBody: map[string]string{
				"status": "applied",
			},
			mockSetup:      func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid insight ID format",
		},
		{
			name:      "invalid status",
			insightID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody: map[string]string{
				"status": "invalid",
			},
			mockSetup:      func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid status: must be one of 'pending', 'applied', or 'dismissed'",
		},
		{
			name:           "invalid JSON",
			insightID:      "550e8400-e29b-41d4-a716-446655440000",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON in request body",
		},
		{
			name:      "service error",
			insightID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody: map[string]string{
				"status": "applied",
			},
			mockSetup: func(m *MockInsightsService) {
				insightID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				m.On("UpdateInsightStatus", mock.Anything, insightID, models.InsightStatusApplied).Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to update insight status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockInsightsService)
			tt.mockSetup(mockService)

			handler := NewInsightsHandler(mockService, logger)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("PUT", "/api/v1/insights/"+tt.insightID+"/status", bytes.NewBuffer(body))
			assert.NoError(t, err)

			// Set up mux vars
			req = mux.SetURLVars(req, map[string]string{"id": tt.insightID})

			rr := httptest.NewRecorder()
			handler.UpdateInsightStatus(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"].(map[string]interface{})["message"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, true, response["success"])
				assert.Contains(t, response, "insight_id")
				assert.Contains(t, response, "status")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInsightsHandler_GenerateInsights(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockInsightsService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful insights generation",
			requestBody: map[string]string{
				"workspace_id": "550e8400-e29b-41d4-a716-446655440000",
			},
			mockSetup: func(m *MockInsightsService) {
				workspaceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				insights := []*models.Insight{
					{
						ID:          uuid.New(),
						WorkspaceID: workspaceID,
						InsightType: models.InsightTypePerformanceBottleneck,
						Priority:    models.PriorityHigh,
						Title:       "Generated Insight",
						Status:      models.InsightStatusPending,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				m.On("GenerateInsights", mock.Anything, workspaceID).Return(insights, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON in request body",
		},
		{
			name: "missing workspace_id",
			requestBody: map[string]string{
				"other_field": "value",
			},
			mockSetup:      func(m *MockInsightsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace_id is required",
		},
		{
			name: "service error",
			requestBody: map[string]string{
				"workspace_id": "550e8400-e29b-41d4-a716-446655440000",
			},
			mockSetup: func(m *MockInsightsService) {
				workspaceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				m.On("GenerateInsights", mock.Anything, workspaceID).Return([]*models.Insight{}, fmt.Errorf("generation error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to generate insights",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockInsightsService)
			tt.mockSetup(mockService)

			handler := NewInsightsHandler(mockService, logger)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/api/v1/insights/generate", bytes.NewBuffer(body))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.GenerateInsights(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"].(map[string]interface{})["message"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, true, response["success"])
				assert.Contains(t, response, "workspace_id")
				assert.Contains(t, response, "insights_generated")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInsightsHandler_parseInsightFilters(t *testing.T) {
	logger := logrus.New()
	handler := NewInsightsHandler(nil, logger)

	tests := []struct {
		name          string
		queryParams   string
		expectedError string
		validateFunc  func(*services.InsightFilters) bool
	}{
		{
			name:        "default filters",
			queryParams: "",
			validateFunc: func(f *services.InsightFilters) bool {
				return f.Limit == 50 && f.Offset == 0 && f.Status == nil && f.Type == nil && f.Priority == nil
			},
		},
		{
			name:        "all filters",
			queryParams: "status=pending&type=performance_bottleneck&priority=high&limit=10&offset=5",
			validateFunc: func(f *services.InsightFilters) bool {
				return f.Limit == 10 && f.Offset == 5 &&
					f.Status != nil && *f.Status == models.InsightStatusPending &&
					f.Type != nil && *f.Type == models.InsightTypePerformanceBottleneck &&
					f.Priority != nil && *f.Priority == models.PriorityHigh
			},
		},
		{
			name:          "invalid status",
			queryParams:   "status=invalid",
			expectedError: "invalid status",
		},
		{
			name:          "invalid type",
			queryParams:   "type=invalid",
			expectedError: "invalid insight type",
		},
		{
			name:          "invalid priority",
			queryParams:   "priority=invalid",
			expectedError: "invalid priority",
		},
		{
			name:          "invalid limit",
			queryParams:   "limit=0",
			expectedError: "invalid limit",
		},
		{
			name:          "limit too high",
			queryParams:   "limit=101",
			expectedError: "invalid limit",
		},
		{
			name:          "invalid offset",
			queryParams:   "offset=-1",
			expectedError: "invalid offset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/v1/insights?"+tt.queryParams, nil)
			assert.NoError(t, err)

			filters, err := handler.parseInsightFilters(req)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.validateFunc(filters))
			}
		})
	}
}

func TestInsightsHandler_RegisterRoutes(t *testing.T) {
	logger := logrus.New()
	mockService := new(MockInsightsService)
	handler := NewInsightsHandler(mockService, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that routes are registered
	req, _ := http.NewRequest("GET", "/api/v1/insights", nil)
	match := &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))

	req, _ = http.NewRequest("PUT", "/api/v1/insights/123/status", nil)
	match = &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))

	req, _ = http.NewRequest("POST", "/api/v1/insights/generate", nil)
	match = &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))
}