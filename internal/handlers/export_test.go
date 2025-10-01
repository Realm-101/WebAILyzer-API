package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/projectdiscovery/wappalyzergo/internal/middleware"
	"github.com/projectdiscovery/wappalyzergo/internal/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock export service
type mockExportService struct {
	mock.Mock
}

func (m *mockExportService) ExportAnalysisResultsCSV(ctx context.Context, req *services.ExportRequest, writer io.Writer) error {
	args := m.Called(ctx, req, writer)
	
	// Write sample CSV data for testing
	if args.Error(0) == nil {
		csvWriter := csv.NewWriter(writer)
		csvWriter.Write([]string{"id", "url", "created_at"})
		csvWriter.Write([]string{"123", "https://example.com", "2024-01-01T00:00:00Z"})
		csvWriter.Flush()
	}
	
	return args.Error(0)
}

func (m *mockExportService) ExportAnalysisResultsJSON(ctx context.Context, req *services.ExportRequest, writer io.Writer) error {
	args := m.Called(ctx, req, writer)
	
	// Write sample JSON data for testing
	if args.Error(0) == nil {
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"export_type":  "analysis_results",
				"record_count": 1,
			},
			"data": []map[string]interface{}{
				{
					"id":  "123",
					"url": "https://example.com",
				},
			},
		}
		json.NewEncoder(writer).Encode(data)
	}
	
	return args.Error(0)
}

func (m *mockExportService) ExportMetricsCSV(ctx context.Context, req *services.MetricsExportRequest, writer io.Writer) error {
	args := m.Called(ctx, req, writer)
	
	// Write sample CSV data for testing
	if args.Error(0) == nil {
		csvWriter := csv.NewWriter(writer)
		csvWriter.Write([]string{"date", "total_sessions"})
		csvWriter.Write([]string{"2024-01-01", "100"})
		csvWriter.Flush()
	}
	
	return args.Error(0)
}

func (m *mockExportService) ExportMetricsJSON(ctx context.Context, req *services.MetricsExportRequest, writer io.Writer) error {
	args := m.Called(ctx, req, writer)
	
	// Write sample JSON data for testing
	if args.Error(0) == nil {
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"export_type":  "metrics",
				"record_count": 1,
			},
			"data": []map[string]interface{}{
				{
					"date":           "2024-01-01",
					"total_sessions": 100,
				},
			},
		}
		json.NewEncoder(writer).Encode(data)
	}
	
	return args.Error(0)
}

func (m *mockExportService) ExportSessionsCSV(ctx context.Context, req *services.SessionExportRequest, writer io.Writer) error {
	args := m.Called(ctx, req, writer)
	
	// Write sample CSV data for testing
	if args.Error(0) == nil {
		csvWriter := csv.NewWriter(writer)
		csvWriter.Write([]string{"id", "user_id", "started_at"})
		csvWriter.Write([]string{"123", "user1", "2024-01-01T00:00:00Z"})
		csvWriter.Flush()
	}
	
	return args.Error(0)
}

func (m *mockExportService) ExportSessionsJSON(ctx context.Context, req *services.SessionExportRequest, writer io.Writer) error {
	args := m.Called(ctx, req, writer)
	
	// Write sample JSON data for testing
	if args.Error(0) == nil {
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"export_type":  "sessions",
				"record_count": 1,
			},
			"data": []map[string]interface{}{
				{
					"id":         "123",
					"user_id":    "user1",
					"started_at": "2024-01-01T00:00:00Z",
				},
			},
		}
		json.NewEncoder(writer).Encode(data)
	}
	
	return args.Error(0)
}

func (m *mockExportService) ExportEventsCSV(ctx context.Context, req *services.EventExportRequest, writer io.Writer) error {
	args := m.Called(ctx, req, writer)
	
	// Write sample CSV data for testing
	if args.Error(0) == nil {
		csvWriter := csv.NewWriter(writer)
		csvWriter.Write([]string{"id", "event_type", "timestamp"})
		csvWriter.Write([]string{"123", "pageview", "2024-01-01T00:00:00Z"})
		csvWriter.Flush()
	}
	
	return args.Error(0)
}

func (m *mockExportService) ExportEventsJSON(ctx context.Context, req *services.EventExportRequest, writer io.Writer) error {
	args := m.Called(ctx, req, writer)
	
	// Write sample JSON data for testing
	if args.Error(0) == nil {
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"export_type":  "events",
				"record_count": 1,
			},
			"data": []map[string]interface{}{
				{
					"id":         "123",
					"event_type": "pageview",
					"timestamp":  "2024-01-01T00:00:00Z",
				},
			},
		}
		json.NewEncoder(writer).Encode(data)
	}
	
	return args.Error(0)
}

func TestExportHandler_ExportAnalysisCSV(t *testing.T) {
	// Setup
	mockService := &mockExportService{}
	logger := logrus.New()
	handler := NewExportHandler(mockService, logger)

	workspaceID := uuid.New()

	// Setup expectations
	mockService.On("ExportAnalysisResultsCSV", mock.Anything, mock.AnythingOfType("*services.ExportRequest"), mock.Anything).Return(nil)

	// Create request
	req := httptest.NewRequest("GET", "/v1/export/analysis/csv", nil)
	
	// Add workspace ID to context
	ctx := context.WithValue(req.Context(), middleware.WorkspaceIDKey, workspaceID.String())
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ExportAnalysisCSV(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/csv", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(t, rr.Header().Get("Content-Disposition"), "analysis_export_")
	assert.Contains(t, rr.Header().Get("Content-Disposition"), ".csv")

	// Verify CSV content
	csvReader := csv.NewReader(rr.Body)
	records, err := csvReader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 2) // Header + 1 data row
	assert.Equal(t, []string{"id", "url", "created_at"}, records[0])
	assert.Equal(t, []string{"123", "https://example.com", "2024-01-01T00:00:00Z"}, records[1])

	mockService.AssertExpectations(t)
}

func TestExportHandler_ExportAnalysisJSON(t *testing.T) {
	// Setup
	mockService := &mockExportService{}
	logger := logrus.New()
	handler := NewExportHandler(mockService, logger)

	workspaceID := uuid.New()

	// Setup expectations
	mockService.On("ExportAnalysisResultsJSON", mock.Anything, mock.AnythingOfType("*services.ExportRequest"), mock.Anything).Return(nil)

	// Create request
	req := httptest.NewRequest("GET", "/v1/export/analysis/json", nil)
	
	// Add workspace ID to context
	ctx := context.WithValue(req.Context(), middleware.WorkspaceIDKey, workspaceID.String())
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ExportAnalysisJSON(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(t, rr.Header().Get("Content-Disposition"), "analysis_export_")
	assert.Contains(t, rr.Header().Get("Content-Disposition"), ".json")

	// Verify JSON content
	var data map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &data)
	assert.NoError(t, err)
	
	metadata, ok := data["metadata"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "analysis_results", metadata["export_type"])

	mockService.AssertExpectations(t)
}

func TestExportHandler_ExportAnalysisCSV_WithFilters(t *testing.T) {
	// Setup
	mockService := &mockExportService{}
	logger := logrus.New()
	handler := NewExportHandler(mockService, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	// Setup expectations with specific request validation
	mockService.On("ExportAnalysisResultsCSV", mock.Anything, mock.MatchedBy(func(req *services.ExportRequest) bool {
		return req.WorkspaceID == workspaceID &&
			req.SessionID != nil &&
			*req.SessionID == sessionID &&
			req.Limit == 100 &&
			req.Offset == 50
	}), mock.Anything).Return(nil)

	// Create request with query parameters
	reqURL := fmt.Sprintf("/v1/export/analysis/csv?session_id=%s&limit=100&offset=50", sessionID.String())
	req := httptest.NewRequest("GET", reqURL, nil)
	
	// Add workspace ID to context
	ctx := context.WithValue(req.Context(), middleware.WorkspaceIDKey, workspaceID.String())
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ExportAnalysisCSV(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	mockService.AssertExpectations(t)
}

func TestExportHandler_ExportMetricsCSV(t *testing.T) {
	// Setup
	mockService := &mockExportService{}
	logger := logrus.New()
	handler := NewExportHandler(mockService, logger)

	workspaceID := uuid.New()
	startDate := "2024-01-01T00:00:00Z"
	endDate := "2024-01-31T23:59:59Z"

	// Setup expectations
	mockService.On("ExportMetricsCSV", mock.Anything, mock.AnythingOfType("*services.MetricsExportRequest"), mock.Anything).Return(nil)

	// Create request with required parameters
	reqURL := fmt.Sprintf("/v1/export/metrics/csv?start_date=%s&end_date=%s", 
		url.QueryEscape(startDate), url.QueryEscape(endDate))
	req := httptest.NewRequest("GET", reqURL, nil)
	
	// Add workspace ID to context
	ctx := context.WithValue(req.Context(), middleware.WorkspaceIDKey, workspaceID.String())
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ExportMetricsCSV(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/csv", rr.Header().Get("Content-Type"))

	mockService.AssertExpectations(t)
}

func TestExportHandler_ExportMetricsCSV_MissingRequiredParams(t *testing.T) {
	// Setup
	mockService := &mockExportService{}
	logger := logrus.New()
	handler := NewExportHandler(mockService, logger)

	workspaceID := uuid.New()

	// Create request without required parameters
	req := httptest.NewRequest("GET", "/v1/export/metrics/csv", nil)
	
	// Add workspace ID to context
	ctx := context.WithValue(req.Context(), middleware.WorkspaceIDKey, workspaceID.String())
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ExportMetricsCSV(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request parameters")

	// Ensure service was not called
	mockService.AssertNotCalled(t, "ExportMetricsCSV")
}

func TestExportHandler_ExportSessionsJSON(t *testing.T) {
	// Setup
	mockService := &mockExportService{}
	logger := logrus.New()
	handler := NewExportHandler(mockService, logger)

	workspaceID := uuid.New()
	userID := "user123"

	// Setup expectations
	mockService.On("ExportSessionsJSON", mock.Anything, mock.MatchedBy(func(req *services.SessionExportRequest) bool {
		return req.WorkspaceID == workspaceID &&
			req.UserID != nil &&
			*req.UserID == userID
	}), mock.Anything).Return(nil)

	// Create request with user_id filter
	reqURL := fmt.Sprintf("/v1/export/sessions/json?user_id=%s", userID)
	req := httptest.NewRequest("GET", reqURL, nil)
	
	// Add workspace ID to context
	ctx := context.WithValue(req.Context(), middleware.WorkspaceIDKey, workspaceID.String())
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ExportSessionsJSON(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	mockService.AssertExpectations(t)
}

func TestExportHandler_ExportEventsCSV(t *testing.T) {
	// Setup
	mockService := &mockExportService{}
	logger := logrus.New()
	handler := NewExportHandler(mockService, logger)

	workspaceID := uuid.New()
	eventType := "pageview"

	// Setup expectations
	mockService.On("ExportEventsCSV", mock.Anything, mock.MatchedBy(func(req *services.EventExportRequest) bool {
		return req.WorkspaceID == workspaceID &&
			req.EventType != nil &&
			*req.EventType == eventType
	}), mock.Anything).Return(nil)

	// Create request with event_type filter
	reqURL := fmt.Sprintf("/v1/export/events/csv?event_type=%s", eventType)
	req := httptest.NewRequest("GET", reqURL, nil)
	
	// Add workspace ID to context
	ctx := context.WithValue(req.Context(), middleware.WorkspaceIDKey, workspaceID.String())
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ExportEventsCSV(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/csv", rr.Header().Get("Content-Type"))

	mockService.AssertExpectations(t)
}

func TestExportHandler_ParseAnalysisExportRequest(t *testing.T) {
	handler := &ExportHandler{}
	workspaceID := uuid.New()

	tests := []struct {
		name        string
		queryParams string
		expectError bool
		validate    func(*services.ExportRequest) bool
	}{
		{
			name:        "no parameters",
			queryParams: "",
			expectError: false,
			validate: func(req *services.ExportRequest) bool {
				return req.WorkspaceID == workspaceID &&
					req.SessionID == nil &&
					req.StartDate == nil &&
					req.EndDate == nil &&
					req.Limit == 0 &&
					req.Offset == 0
			},
		},
		{
			name:        "with session_id",
			queryParams: "session_id=" + uuid.New().String(),
			expectError: false,
			validate: func(req *services.ExportRequest) bool {
				return req.SessionID != nil
			},
		},
		{
			name:        "with dates",
			queryParams: "start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z",
			expectError: false,
			validate: func(req *services.ExportRequest) bool {
				return req.StartDate != nil && req.EndDate != nil
			},
		},
		{
			name:        "with pagination",
			queryParams: "limit=100&offset=50",
			expectError: false,
			validate: func(req *services.ExportRequest) bool {
				return req.Limit == 100 && req.Offset == 50
			},
		},
		{
			name:        "invalid session_id",
			queryParams: "session_id=invalid-uuid",
			expectError: true,
			validate:    nil,
		},
		{
			name:        "invalid date format",
			queryParams: "start_date=invalid-date",
			expectError: true,
			validate:    nil,
		},
		{
			name:        "invalid limit",
			queryParams: "limit=invalid",
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqURL := "/test?" + tt.queryParams
			req := httptest.NewRequest("GET", reqURL, nil)

			result, err := handler.parseAnalysisExportRequest(req, workspaceID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					assert.True(t, tt.validate(result))
				}
			}
		})
	}
}

func TestExportHandler_ParseMetricsExportRequest(t *testing.T) {
	handler := &ExportHandler{}
	workspaceID := uuid.New()

	tests := []struct {
		name        string
		queryParams string
		expectError bool
	}{
		{
			name:        "valid parameters",
			queryParams: "start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z",
			expectError: false,
		},
		{
			name:        "missing start_date",
			queryParams: "end_date=2024-01-31T23:59:59Z",
			expectError: true,
		},
		{
			name:        "missing end_date",
			queryParams: "start_date=2024-01-01T00:00:00Z",
			expectError: true,
		},
		{
			name:        "invalid date format",
			queryParams: "start_date=invalid&end_date=2024-01-31T23:59:59Z",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqURL := "/test?" + tt.queryParams
			req := httptest.NewRequest("GET", reqURL, nil)

			result, err := handler.parseMetricsExportRequest(req, workspaceID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, workspaceID, result.WorkspaceID)
			}
		})
	}
}
