package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMetricsHandler_Integration_GetMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	// Create a mock service that returns realistic data
	mockService := new(MockMetricsService)
	
	// Setup realistic mock response
	workspaceID := uuid.New()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	mockResponse := &models.MetricsResponse{
		Metrics: map[string]models.MetricData{
			"conversion_rate": {
				Current:  3.2,
				Previous: 2.8,
				Trend:    "up",
				DataPoints: []models.DataPoint{
					{Timestamp: startDate, Value: 2.5},
					{Timestamp: startDate.Add(24 * time.Hour), Value: 2.8},
					{Timestamp: startDate.Add(48 * time.Hour), Value: 3.0},
					{Timestamp: startDate.Add(72 * time.Hour), Value: 3.2},
					{Timestamp: startDate.Add(96 * time.Hour), Value: 3.1},
					{Timestamp: startDate.Add(120 * time.Hour), Value: 3.4},
					{Timestamp: startDate.Add(144 * time.Hour), Value: 3.2},
				},
			},
			"bounce_rate": {
				Current:  45.2,
				Previous: 48.1,
				Trend:    "down",
				DataPoints: []models.DataPoint{
					{Timestamp: startDate, Value: 48.5},
					{Timestamp: startDate.Add(24 * time.Hour), Value: 47.2},
					{Timestamp: startDate.Add(48 * time.Hour), Value: 46.8},
					{Timestamp: startDate.Add(72 * time.Hour), Value: 45.2},
					{Timestamp: startDate.Add(96 * time.Hour), Value: 44.9},
					{Timestamp: startDate.Add(120 * time.Hour), Value: 43.8},
					{Timestamp: startDate.Add(144 * time.Hour), Value: 45.2},
				},
			},
			"avg_session_duration": {
				Current:  185.5,
				Previous: 172.3,
				Trend:    "up",
				DataPoints: []models.DataPoint{
					{Timestamp: startDate, Value: 165.0},
					{Timestamp: startDate.Add(24 * time.Hour), Value: 172.3},
					{Timestamp: startDate.Add(48 * time.Hour), Value: 178.9},
					{Timestamp: startDate.Add(72 * time.Hour), Value: 185.5},
					{Timestamp: startDate.Add(96 * time.Hour), Value: 189.2},
					{Timestamp: startDate.Add(120 * time.Hour), Value: 192.1},
					{Timestamp: startDate.Add(144 * time.Hour), Value: 185.5},
				},
			},
			"avg_load_time": {
				Current:  1850.0,
				Previous: 2100.0,
				Trend:    "down",
				DataPoints: []models.DataPoint{
					{Timestamp: startDate, Value: 2200.0},
					{Timestamp: startDate.Add(24 * time.Hour), Value: 2100.0},
					{Timestamp: startDate.Add(48 * time.Hour), Value: 1950.0},
					{Timestamp: startDate.Add(72 * time.Hour), Value: 1850.0},
					{Timestamp: startDate.Add(96 * time.Hour), Value: 1800.0},
					{Timestamp: startDate.Add(120 * time.Hour), Value: 1750.0},
					{Timestamp: startDate.Add(144 * time.Hour), Value: 1850.0},
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
			{
				Name:        "Bounce Rate",
				Value:       45.2,
				Target:      40.0,
				Status:      "warning",
				Description: "Percentage of single-page sessions",
			},
			{
				Name:        "Average Session Duration",
				Value:       185.5,
				Target:      180.0,
				Status:      "good",
				Description: "Average time users spend on the site (seconds)",
			},
			{
				Name:        "Page Load Time",
				Value:       1850.0,
				Target:      2000.0,
				Status:      "good",
				Description: "Average page load time (milliseconds)",
			},
		},
		Anomalies: []models.Anomaly{
			{
				Metric:      "conversion_rate",
				Timestamp:   startDate.Add(72 * time.Hour),
				Expected:    2.9,
				Actual:      3.2,
				Severity:    "medium",
				Description: "Conversion rate spike detected",
			},
		},
	}

	mockService.On("GetMetrics", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Create handler and router
	handler := NewMetricsHandler(mockService, nil, logger)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create test request
	queryParams := fmt.Sprintf("workspace_id=%s&start_date=%s&end_date=%s&granularity=daily",
		workspaceID.String(),
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339),
	)

	req := httptest.NewRequest("GET", "/api/v1/metrics?"+queryParams, nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "metrics")
	assert.Contains(t, response, "kpis")
	assert.Contains(t, response, "anomalies")
	assert.Contains(t, response, "metadata")

	// Verify metrics data
	metrics := response["metrics"].(map[string]interface{})
	assert.Len(t, metrics, 4)

	// Verify conversion rate metric
	conversionRate := metrics["conversion_rate"].(map[string]interface{})
	assert.Equal(t, 3.2, conversionRate["current"])
	assert.Equal(t, 2.8, conversionRate["previous"])
	assert.Equal(t, "up", conversionRate["trend"])
	
	dataPoints := conversionRate["data_points"].([]interface{})
	assert.Len(t, dataPoints, 7)

	// Verify KPIs
	kpis := response["kpis"].([]interface{})
	assert.Len(t, kpis, 4)

	firstKPI := kpis[0].(map[string]interface{})
	assert.Equal(t, "Conversion Rate", firstKPI["name"])
	assert.Equal(t, 3.2, firstKPI["value"])
	assert.Equal(t, 5.0, firstKPI["target"])
	assert.Equal(t, "warning", firstKPI["status"])

	// Verify anomalies
	anomalies := response["anomalies"].([]interface{})
	assert.Len(t, anomalies, 1)

	firstAnomaly := anomalies[0].(map[string]interface{})
	assert.Equal(t, "conversion_rate", firstAnomaly["metric"])
	assert.Equal(t, 2.9, firstAnomaly["expected"])
	assert.Equal(t, 3.2, firstAnomaly["actual"])
	assert.Equal(t, "medium", firstAnomaly["severity"])

	// Verify metadata
	metadata := response["metadata"].(map[string]interface{})
	assert.Equal(t, false, metadata["from_cache"])
	assert.Equal(t, "real_time", metadata["data_source"])
	assert.Contains(t, metadata, "timestamp")
	assert.Contains(t, metadata, "most_recent_data")
	assert.Contains(t, metadata, "data_age_minutes")
	assert.Contains(t, metadata, "freshness_status")

	// Verify mock expectations
	mockService.AssertExpectations(t)
}

func TestMetricsHandler_Integration_ErrorScenarios(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing all parameters",
			url:            "/api/v1/metrics",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace_id query parameter is required",
		},
		{
			name:           "invalid date range",
			url:            "/api/v1/metrics?workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-02T00:00:00Z&end_date=2024-01-01T00:00:00Z&granularity=daily",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "end_date must be after start_date",
		},
		{
			name:           "unsupported granularity",
			url:            "/api/v1/metrics?workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=2024-01-02T00:00:00Z&granularity=minutely",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid granularity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			handler := NewMetricsHandler(mockService, nil, logger)
			router := mux.NewRouter()
			handler.RegisterRoutes(router)

			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Contains(t, response, "error")
			errorObj := response["error"].(map[string]interface{})
			assert.Contains(t, errorObj, "message")

			message := errorObj["message"].(string)
			assert.Contains(t, message, tt.expectedError)
		})
	}
}

func TestMetricsHandler_Integration_DifferentGranularities(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	granularities := []struct {
		name        string
		granularity string
		startDate   string
		endDate     string
	}{
		{
			name:        "hourly metrics",
			granularity: "hourly",
			startDate:   "2024-01-01T00:00:00Z",
			endDate:     "2024-01-01T12:00:00Z",
		},
		{
			name:        "daily metrics",
			granularity: "daily",
			startDate:   "2024-01-01T00:00:00Z",
			endDate:     "2024-01-07T00:00:00Z",
		},
		{
			name:        "weekly metrics",
			granularity: "weekly",
			startDate:   "2024-01-01T00:00:00Z",
			endDate:     "2024-02-01T00:00:00Z",
		},
		{
			name:        "monthly metrics",
			granularity: "monthly",
			startDate:   "2024-01-01T00:00:00Z",
			endDate:     "2024-06-01T00:00:00Z",
		},
	}

	for _, tt := range granularities {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			mockResponse := &models.MetricsResponse{
				Metrics:   make(map[string]models.MetricData),
				KPIs:      []models.KPI{},
				Anomalies: []models.Anomaly{},
			}
			mockService.On("GetMetrics", mock.Anything, mock.Anything).Return(mockResponse, nil)

			handler := NewMetricsHandler(mockService, nil, logger)
			router := mux.NewRouter()
			handler.RegisterRoutes(router)

			workspaceID := uuid.New()
			queryParams := fmt.Sprintf("workspace_id=%s&start_date=%s&end_date=%s&granularity=%s",
				workspaceID.String(), tt.startDate, tt.endDate, tt.granularity)

			req := httptest.NewRequest("GET", "/api/v1/metrics?"+queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Contains(t, response, "metrics")
			assert.Contains(t, response, "metadata")

			mockService.AssertExpectations(t)
		})
	}
}