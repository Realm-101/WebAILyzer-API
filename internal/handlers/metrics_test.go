package handlers

import (
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

// MockMetricsService is a mock implementation of MetricsService
type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) GetMetrics(ctx context.Context, req *models.MetricsRequest) (*models.MetricsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MetricsResponse), args.Error(1)
}

func (m *MockMetricsService) GetKPIs(ctx context.Context, workspaceID uuid.UUID, timeRange services.TimeRange) (*services.KPIResponse, error) {
	args := m.Called(ctx, workspaceID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.KPIResponse), args.Error(1)
}

func (m *MockMetricsService) DetectAnomalies(ctx context.Context, workspaceID uuid.UUID) ([]*models.Anomaly, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Anomaly), args.Error(1)
}

func TestMetricsHandler_GetMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	tests := []struct {
		name           string
		queryParams    string
		mockResponse   *models.MetricsResponse
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful metrics retrieval",
			queryParams: "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=2024-01-02T00:00:00Z&granularity=daily",
			mockResponse: &models.MetricsResponse{
				Metrics: map[string]models.MetricData{
					"conversion_rate": {
						Current:  3.2,
						Previous: 2.8,
						Trend:    "up",
						DataPoints: []models.DataPoint{
							{
								Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
								Value:     3.2,
							},
						},
					},
				},
				KPIs: []models.KPI{
					{
						Name:        "Conversion Rate",
						Value:       3.2,
						Target:      5.0,
						Status:      "warning",
						Description: "Percentage of sessions that result in conversions",
					},
				},
				Anomalies: []models.Anomaly{},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing workspace_id",
			queryParams:    "start_date=2024-01-01T00:00:00Z&end_date=2024-01-02T00:00:00Z&granularity=daily",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace_id query parameter is required",
		},
		{
			name:           "missing start_date",
			queryParams:    "workspace_id=550e8400-e29b-41d4-a716-446655440000&end_date=2024-01-02T00:00:00Z&granularity=daily",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "start_date query parameter is required",
		},
		{
			name:           "missing end_date",
			queryParams:    "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&granularity=daily",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "end_date query parameter is required",
		},
		{
			name:           "missing granularity",
			queryParams:    "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=2024-01-02T00:00:00Z",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "granularity query parameter is required",
		},
		{
			name:           "invalid workspace_id format",
			queryParams:    "workspace_id=invalid-uuid&start_date=2024-01-01T00:00:00Z&end_date=2024-01-02T00:00:00Z&granularity=daily",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid workspace_id format",
		},
		{
			name:           "invalid start_date format",
			queryParams:    "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=invalid-date&end_date=2024-01-02T00:00:00Z&granularity=daily",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid start_date format",
		},
		{
			name:           "invalid end_date format",
			queryParams:    "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=invalid-date&granularity=daily",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid end_date format",
		},
		{
			name:           "invalid granularity",
			queryParams:    "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=2024-01-02T00:00:00Z&granularity=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid granularity",
		},
		{
			name:           "end_date before start_date",
			queryParams:    "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-02T00:00:00Z&end_date=2024-01-01T00:00:00Z&granularity=daily",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "end_date must be after start_date",
		},
		{
			name:           "date range too large for hourly granularity",
			queryParams:    "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=2024-01-15T00:00:00Z&granularity=hourly",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "date range too large for hourly granularity",
		},
		{
			name:        "service error",
			queryParams: "workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=2024-01-02T00:00:00Z&granularity=daily",
			mockError:   fmt.Errorf("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to retrieve metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(MockMetricsService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("GetMetrics", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			handler := NewMetricsHandler(mockService, nil, logger)

			// Create request
			req := httptest.NewRequest("GET", "/api/v1/metrics?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Execute request
			handler.GetMetrics(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				// Check successful response structure
				assert.Contains(t, response, "metrics")
				assert.Contains(t, response, "kpis")
				assert.Contains(t, response, "anomalies")
				assert.Contains(t, response, "metadata")

				// Check metadata structure
				metadata := response["metadata"].(map[string]interface{})
				assert.Contains(t, metadata, "timestamp")
				assert.Contains(t, metadata, "from_cache")
				assert.Contains(t, metadata, "data_source")
				assert.Equal(t, false, metadata["from_cache"])
				assert.Equal(t, "real_time", metadata["data_source"])

				// Check metrics structure
				metrics := response["metrics"].(map[string]interface{})
				if len(metrics) > 0 {
					conversionRate := metrics["conversion_rate"].(map[string]interface{})
					assert.Contains(t, conversionRate, "current")
					assert.Contains(t, conversionRate, "previous")
					assert.Contains(t, conversionRate, "trend")
					assert.Contains(t, conversionRate, "data_points")
				}
			} else {
				// Check error response structure
				assert.Contains(t, response, "error")
				errorObj := response["error"].(map[string]interface{})
				assert.Contains(t, errorObj, "code")
				assert.Contains(t, errorObj, "message")
				assert.Contains(t, errorObj, "timestamp")

				if tt.expectedError != "" {
					message := errorObj["message"].(string)
					assert.Contains(t, message, tt.expectedError)
				}
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricsHandler_GetMetrics_WithDifferentGranularities(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	granularities := []struct {
		name        string
		granularity string
		startDate   string
		endDate     string
		expectError bool
	}{
		{
			name:        "hourly granularity - valid range",
			granularity: "hourly",
			startDate:   "2024-01-01T00:00:00Z",
			endDate:     "2024-01-02T00:00:00Z",
			expectError: false,
		},
		{
			name:        "daily granularity - valid range",
			granularity: "daily",
			startDate:   "2024-01-01T00:00:00Z",
			endDate:     "2024-01-31T00:00:00Z",
			expectError: false,
		},
		{
			name:        "weekly granularity - valid range",
			granularity: "weekly",
			startDate:   "2024-01-01T00:00:00Z",
			endDate:     "2024-06-01T00:00:00Z",
			expectError: false,
		},
		{
			name:        "monthly granularity - valid range",
			granularity: "monthly",
			startDate:   "2024-01-01T00:00:00Z",
			endDate:     "2024-12-01T00:00:00Z",
			expectError: false,
		},
	}

	for _, tt := range granularities {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(MockMetricsService)
			mockResponse := &models.MetricsResponse{
				Metrics:   make(map[string]models.MetricData),
				KPIs:      []models.KPI{},
				Anomalies: []models.Anomaly{},
			}
			mockService.On("GetMetrics", mock.Anything, mock.Anything).Return(mockResponse, nil)

			// Create handler
			handler := NewMetricsHandler(mockService, nil, logger)

			// Create request
			queryParams := fmt.Sprintf("workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=%s&end_date=%s&granularity=%s",
				tt.startDate, tt.endDate, tt.granularity)
			req := httptest.NewRequest("GET", "/api/v1/metrics?"+queryParams, nil)
			w := httptest.NewRecorder()

			// Execute request
			handler.GetMetrics(w, req)

			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricsHandler_RegisterRoutes(t *testing.T) {
	logger := logrus.New()
	mockService := new(MockMetricsService)
	handler := NewMetricsHandler(mockService, nil, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that the route is registered
	req := httptest.NewRequest("GET", "/api/v1/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 (bad request) due to missing parameters, not 404 (not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestMetricsHandler_CacheKeyGeneration(t *testing.T) {
	logger := logrus.New()
	mockService := new(MockMetricsService)
	handler := NewMetricsHandler(mockService, nil, logger)

	workspaceID := uuid.New()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	req := &models.MetricsRequest{
		WorkspaceID: workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: "daily",
	}

	cacheKey := handler.generateMetricsCacheKey(req)

	// Verify cache key format
	expectedKey := fmt.Sprintf("metrics:%s:%s:%s:daily",
		workspaceID.String(),
		startDate.Format("2006-01-02T15:04:05Z07:00"),
		endDate.Format("2006-01-02T15:04:05Z07:00"),
	)

	assert.Equal(t, expectedKey, cacheKey)
}

func TestMetricsHandler_MaxDurationValidation(t *testing.T) {
	logger := logrus.New()
	mockService := new(MockMetricsService)
	handler := NewMetricsHandler(mockService, nil, logger)

	tests := []struct {
		granularity string
		expected    time.Duration
	}{
		{"hourly", 7 * 24 * time.Hour},
		{"daily", 90 * 24 * time.Hour},
		{"weekly", 365 * 24 * time.Hour},
		{"monthly", 2 * 365 * 24 * time.Hour},
		{"unknown", 30 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.granularity, func(t *testing.T) {
			duration := handler.getMaxDurationForGranularity(tt.granularity)
			assert.Equal(t, tt.expected, duration)
		})
	}
}

func TestMetricsHandler_FreshnessMetadata(t *testing.T) {
	logger := logrus.New()
	mockService := new(MockMetricsService)
	handler := NewMetricsHandler(mockService, nil, logger)

	// Create test data with recent timestamp
	recentTime := time.Now().Add(-30 * time.Minute)
	result := &models.MetricsResponse{
		Metrics: map[string]models.MetricData{
			"conversion_rate": {
				DataPoints: []models.DataPoint{
					{
						Timestamp: recentTime,
						Value:     3.2,
					},
				},
			},
		},
		KPIs:      []models.KPI{},
		Anomalies: []models.Anomaly{},
	}

	response := handler.addFreshnessMetadata(result, false)

	// Check metadata structure
	assert.Contains(t, response, "metadata")
	metadata := response["metadata"].(map[string]interface{})

	assert.Contains(t, metadata, "timestamp")
	assert.Contains(t, metadata, "from_cache")
	assert.Contains(t, metadata, "data_source")
	assert.Contains(t, metadata, "most_recent_data")
	assert.Contains(t, metadata, "data_age_minutes")
	assert.Contains(t, metadata, "freshness_status")

	assert.Equal(t, false, metadata["from_cache"])
	assert.Equal(t, "real_time", metadata["data_source"])
	assert.Equal(t, "fresh", metadata["freshness_status"])
	assert.Equal(t, 30, metadata["data_age_minutes"])
}