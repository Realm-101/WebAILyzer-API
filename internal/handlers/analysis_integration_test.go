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
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/services"
)

// MockAnalysisService is a mock implementation of AnalysisService
type MockAnalysisService struct {
	mock.Mock
}

func (m *MockAnalysisService) AnalyzeURL(ctx context.Context, req *models.AnalysisRequest) (*models.AnalysisResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalysisResult), args.Error(1)
}

func (m *MockAnalysisService) BatchAnalyze(ctx context.Context, req *services.BatchAnalysisRequest) (*services.BatchAnalysisResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.BatchAnalysisResult), args.Error(1)
}

func (m *MockAnalysisService) GetAnalysisHistory(ctx context.Context, workspaceID uuid.UUID, filters *services.AnalysisFilters) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, workspaceID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func TestAnalysisHandler_AnalyzeURL_Success(t *testing.T) {
	// Setup
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	handler := NewAnalysisHandler(mockService, nil, logger)
	
	workspaceID := uuid.New()
	sessionID := uuid.New()
	analysisID := uuid.New()
	
	// Mock request
	req := models.AnalysisRequest{
		URL:         "https://example.com",
		WorkspaceID: workspaceID,
		SessionID:   &sessionID,
		Options: models.AnalysisOptions{
			IncludePerformance:   true,
			IncludeSEO:          true,
			IncludeAccessibility: true,
			IncludeSecurity:     true,
		},
	}
	
	// Mock response
	expectedResult := &models.AnalysisResult{
		ID:          analysisID,
		WorkspaceID: workspaceID,
		SessionID:   &sessionID,
		URL:         "https://example.com",
		Technologies: map[string]interface{}{
			"detected": []string{"nginx", "php"},
		},
		PerformanceMetrics: map[string]interface{}{
			"load_time_ms": 1250,
		},
		SEOMetrics: map[string]interface{}{
			"title": "Example Site",
		},
		AccessibilityMetrics: map[string]interface{}{
			"score": 85,
		},
		SecurityMetrics: map[string]interface{}{
			"https": true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	mockService.On("AnalyzeURL", mock.Anything, mock.MatchedBy(func(r *models.AnalysisRequest) bool {
		return r.URL == req.URL && r.WorkspaceID == req.WorkspaceID
	})).Return(expectedResult, nil)
	
	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Execute
	handler.AnalyzeURL(rr, httpReq)
	
	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response models.AnalysisResult
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, expectedResult.ID, response.ID)
	assert.Equal(t, expectedResult.URL, response.URL)
	assert.Equal(t, expectedResult.WorkspaceID, response.WorkspaceID)
	
	mockService.AssertExpectations(t)
}

func TestAnalysisHandler_AnalyzeURL_ValidationErrors(t *testing.T) {
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	handler := NewAnalysisHandler(mockService, nil, logger)
	
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
			name: "Missing URL",
			request: models.AnalysisRequest{
				WorkspaceID: uuid.New(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "Invalid URL scheme",
			request: models.AnalysisRequest{
				URL:         "ftp://example.com",
				WorkspaceID: uuid.New(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "Missing workspace ID",
			request: models.AnalysisRequest{
				URL: "https://example.com",
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
			
			httpReq := httptest.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			handler.AnalyzeURL(rr, httpReq)
			
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

func TestAnalysisHandler_BatchAnalyze_Success(t *testing.T) {
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	handler := NewAnalysisHandler(mockService, nil, logger)
	
	workspaceID := uuid.New()
	batchID := uuid.New()
	
	req := services.BatchAnalysisRequest{
		URLs:        []string{"https://example1.com", "https://example2.com"},
		WorkspaceID: workspaceID,
		Options: models.AnalysisOptions{
			IncludePerformance: true,
		},
	}
	
	expectedResult := &services.BatchAnalysisResult{
		BatchID: batchID,
		Status:  "completed",
		Results: []*models.AnalysisResult{
			{
				ID:          uuid.New(),
				WorkspaceID: workspaceID,
				URL:         "https://example1.com",
				CreatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				WorkspaceID: workspaceID,
				URL:         "https://example2.com",
				CreatedAt:   time.Now(),
			},
		},
		FailedURLs: []string{},
		Progress: services.BatchProgress{
			Completed: 2,
			Total:     2,
		},
	}
	
	mockService.On("BatchAnalyze", mock.Anything, mock.MatchedBy(func(r *services.BatchAnalysisRequest) bool {
		return len(r.URLs) == 2 && r.WorkspaceID == workspaceID
	})).Return(expectedResult, nil)
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handler.BatchAnalyze(rr, httpReq)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response services.BatchAnalysisResult
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, expectedResult.BatchID, response.BatchID)
	assert.Equal(t, expectedResult.Status, response.Status)
	assert.Len(t, response.Results, 2)
	
	mockService.AssertExpectations(t)
}

func TestAnalysisHandler_BatchAnalyze_ValidationErrors(t *testing.T) {
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	handler := NewAnalysisHandler(mockService, nil, logger)
	
	testCases := []struct {
		name           string
		request        services.BatchAnalysisRequest
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "Empty URLs array",
			request: services.BatchAnalysisRequest{
				URLs:        []string{},
				WorkspaceID: uuid.New(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "Too many URLs",
			request: services.BatchAnalysisRequest{
				URLs:        make([]string, 101), // Exceeds max batch size
				WorkspaceID: uuid.New(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "Invalid URL in batch",
			request: services.BatchAnalysisRequest{
				URLs:        []string{"https://example.com", "invalid-url"},
				WorkspaceID: uuid.New(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tc.request)
			httpReq := httptest.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			handler.BatchAnalyze(rr, httpReq)
			
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

func TestAnalysisHandler_GetAnalysisHistory_Success(t *testing.T) {
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	handler := NewAnalysisHandler(mockService, nil, logger)
	
	workspaceID := uuid.New()
	
	expectedResults := []*models.AnalysisResult{
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			URL:         "https://example1.com",
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			URL:         "https://example2.com",
			CreatedAt:   time.Now(),
		},
	}
	
	mockService.On("GetAnalysisHistory", mock.Anything, workspaceID, mock.AnythingOfType("*services.AnalysisFilters")).Return(expectedResults, nil)
	
	url := fmt.Sprintf("/api/v1/analysis?workspace_id=%s&limit=10&offset=0", workspaceID.String())
	httpReq := httptest.NewRequest("GET", url, nil)
	
	rr := httptest.NewRecorder()
	handler.GetAnalysisHistory(rr, httpReq)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	results, ok := response["results"].([]interface{})
	require.True(t, ok)
	assert.Len(t, results, 2)
	
	metadata, ok := response["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), metadata["count"])
	assert.Equal(t, workspaceID.String(), metadata["workspace_id"])
	
	mockService.AssertExpectations(t)
}

func TestAnalysisHandler_GetAnalysisHistory_ValidationErrors(t *testing.T) {
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	handler := NewAnalysisHandler(mockService, nil, logger)
	
	testCases := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Missing workspace_id",
			url:            "/api/v1/analysis",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name:           "Invalid workspace_id format",
			url:            "/api/v1/analysis?workspace_id=invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name:           "Invalid limit",
			url:            "/api/v1/analysis?workspace_id=" + uuid.New().String() + "&limit=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name:           "Limit too high",
			url:            "/api/v1/analysis?workspace_id=" + uuid.New().String() + "&limit=2000",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			httpReq := httptest.NewRequest("GET", tc.url, nil)
			
			rr := httptest.NewRecorder()
			handler.GetAnalysisHistory(rr, httpReq)
			
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

func TestAnalysisHandler_Routes(t *testing.T) {
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	handler := NewAnalysisHandler(mockService, nil, logger)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)
	
	// Test that routes are registered
	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/analyze"},
		{"POST", "/api/v1/batch"},
		{"GET", "/api/v1/analysis"},
	}
	
	for _, route := range routes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			rr := httptest.NewRecorder()
			
			router.ServeHTTP(rr, req)
			
			// Should not return 404 (route not found)
			assert.NotEqual(t, http.StatusNotFound, rr.Code)
		})
	}
}