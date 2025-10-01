package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type mockAnalysisRepository struct {
	mock.Mock
}

func (m *mockAnalysisRepository) Create(ctx context.Context, analysis *models.AnalysisResult) error {
	args := m.Called(ctx, analysis)
	return args.Error(0)
}

func (m *mockAnalysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AnalysisResult, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.AnalysisResult), args.Error(1)
}

func (m *mockAnalysisRepository) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func (m *mockAnalysisRepository) GetBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func (m *mockAnalysisRepository) GetByFilters(ctx context.Context, filters *repositories.AnalysisFilters) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func (m *mockAnalysisRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.AnalysisFilters) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, workspaceID, filters)
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func (m *mockAnalysisRepository) Update(ctx context.Context, analysis *models.AnalysisResult) error {
	args := m.Called(ctx, analysis)
	return args.Error(0)
}

func (m *mockAnalysisRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockMetricsRepository struct {
	mock.Mock
}

func (m *mockMetricsRepository) CreateDailyMetrics(ctx context.Context, metrics *models.DailyMetrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *mockMetricsRepository) GetDailyMetrics(ctx context.Context, workspaceID uuid.UUID, startDate, endDate time.Time) ([]*models.DailyMetrics, error) {
	args := m.Called(ctx, workspaceID, startDate, endDate)
	return args.Get(0).([]*models.DailyMetrics), args.Error(1)
}

func (m *mockMetricsRepository) UpdateDailyMetrics(ctx context.Context, metrics *models.DailyMetrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *mockMetricsRepository) GetMetricsByWorkspace(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) (*models.MetricsResponse, error) {
	args := m.Called(ctx, workspaceID, startTime, endTime)
	return args.Get(0).(*models.MetricsResponse), args.Error(1)
}

type mockSessionRepository struct {
	mock.Mock
}

func (m *mockSessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *mockSessionRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) GetSessionsByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.Session, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *mockSessionRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.SessionFilters) ([]*models.Session, error) {
	args := m.Called(ctx, workspaceID, filters)
	return args.Get(0).([]*models.Session), args.Error(1)
}

type mockEventRepository struct {
	mock.Mock
}

func (m *mockEventRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventRepository) CreateEvents(ctx context.Context, events []*models.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *mockEventRepository) GetEventsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.Event, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*models.Event), args.Error(1)
}

func (m *mockEventRepository) GetEventsByWorkspace(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) ([]*models.Event, error) {
	args := m.Called(ctx, workspaceID, startTime, endTime)
	return args.Get(0).([]*models.Event), args.Error(1)
}

func (m *mockEventRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.EventFilters) ([]*models.Event, error) {
	args := m.Called(ctx, workspaceID, filters)
	return args.Get(0).([]*models.Event), args.Error(1)
}

func TestExportService_ExportAnalysisResultsCSV(t *testing.T) {
	// Setup
	mockAnalysisRepo := &mockAnalysisRepository{}
	mockMetricsRepo := &mockMetricsRepository{}
	mockSessionRepo := &mockSessionRepository{}
	mockEventRepo := &mockEventRepository{}
	logger := logrus.New()

	service := NewExportService(mockAnalysisRepo, mockMetricsRepo, mockSessionRepo, mockEventRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()
	analysisID := uuid.New()

	// Mock data
	analysisResults := []*models.AnalysisResult{
		{
			ID:          analysisID,
			WorkspaceID: workspaceID,
			SessionID:   &sessionID,
			URL:         "https://example.com",
			Technologies: map[string]interface{}{
				"React": map[string]interface{}{"version": "18.0.0"},
			},
			PerformanceMetrics: map[string]interface{}{
				"load_time_ms": 1200,
				"score":        85,
			},
			SEOMetrics: map[string]interface{}{
				"title": "Example Site",
				"score": 90,
			},
			AccessibilityMetrics: map[string]interface{}{
				"score": 78,
			},
			SecurityMetrics: map[string]interface{}{
				"score": 92,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	req := &ExportRequest{
		WorkspaceID: workspaceID,
		SessionID:   &sessionID,
	}

	// Setup expectations
	mockAnalysisRepo.On("GetByWorkspaceID", mock.Anything, workspaceID, mock.AnythingOfType("*repositories.AnalysisFilters")).Return(analysisResults, nil)

	// Execute
	var buf bytes.Buffer
	err := service.ExportAnalysisResultsCSV(context.Background(), req, &buf)

	// Assert
	assert.NoError(t, err)
	
	// Parse CSV and verify content
	csvReader := csv.NewReader(&buf)
	records, err := csvReader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 2) // Header + 1 data row
	
	// Verify header
	expectedHeader := []string{
		"id", "workspace_id", "session_id", "url", "created_at", "updated_at",
		"technologies", "performance_load_time", "performance_score",
		"seo_title", "seo_score", "accessibility_score", "security_score",
	}
	assert.Equal(t, expectedHeader, records[0])
	
	// Verify data row
	assert.Equal(t, analysisID.String(), records[1][0])
	assert.Equal(t, workspaceID.String(), records[1][1])
	assert.Equal(t, sessionID.String(), records[1][2])
	assert.Equal(t, "https://example.com", records[1][3])
	assert.Equal(t, "1200", records[1][7]) // performance_load_time
	assert.Equal(t, "85", records[1][8])   // performance_score
	assert.Equal(t, "Example Site", records[1][9]) // seo_title
	assert.Equal(t, "90", records[1][10])  // seo_score
	assert.Equal(t, "78", records[1][11])  // accessibility_score
	assert.Equal(t, "92", records[1][12])  // security_score

	mockAnalysisRepo.AssertExpectations(t)
}

func TestExportService_ExportAnalysisResultsJSON(t *testing.T) {
	// Setup
	mockAnalysisRepo := &mockAnalysisRepository{}
	mockMetricsRepo := &mockMetricsRepository{}
	mockSessionRepo := &mockSessionRepository{}
	mockEventRepo := &mockEventRepository{}
	logger := logrus.New()

	service := NewExportService(mockAnalysisRepo, mockMetricsRepo, mockSessionRepo, mockEventRepo, logger)

	workspaceID := uuid.New()
	analysisID := uuid.New()

	// Mock data
	analysisResults := []*models.AnalysisResult{
		{
			ID:          analysisID,
			WorkspaceID: workspaceID,
			URL:         "https://example.com",
			Technologies: map[string]interface{}{
				"React": map[string]interface{}{"version": "18.0.0"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	req := &ExportRequest{
		WorkspaceID: workspaceID,
	}

	// Setup expectations
	mockAnalysisRepo.On("GetByWorkspaceID", mock.Anything, workspaceID, mock.AnythingOfType("*repositories.AnalysisFilters")).Return(analysisResults, nil)

	// Execute
	var buf bytes.Buffer
	err := service.ExportAnalysisResultsJSON(context.Background(), req, &buf)

	// Assert
	assert.NoError(t, err)
	
	// Parse JSON and verify content
	var exportData map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &exportData)
	assert.NoError(t, err)
	
	// Verify metadata
	metadata, ok := exportData["metadata"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, workspaceID.String(), metadata["workspace_id"])
	assert.Equal(t, float64(1), metadata["record_count"])
	assert.Equal(t, "analysis_results", metadata["export_type"])
	assert.Equal(t, "json", metadata["format"])
	
	// Verify data
	data, ok := exportData["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)

	mockAnalysisRepo.AssertExpectations(t)
}

func TestExportService_ExportMetricsCSV(t *testing.T) {
	// Setup
	mockAnalysisRepo := &mockAnalysisRepository{}
	mockMetricsRepo := &mockMetricsRepository{}
	mockSessionRepo := &mockSessionRepository{}
	mockEventRepo := &mockEventRepository{}
	logger := logrus.New()

	service := NewExportService(mockAnalysisRepo, mockMetricsRepo, mockSessionRepo, mockEventRepo, logger)

	workspaceID := uuid.New()
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	// Mock data
	bounceRate := 25.5
	avgSessionDuration := 180
	conversionRate := 3.2
	avgLoadTime := 1200

	metrics := []*models.DailyMetrics{
		{
			ID:                 uuid.New(),
			WorkspaceID:        workspaceID,
			Date:               startDate,
			TotalSessions:      100,
			TotalPageViews:     250,
			UniqueVisitors:     80,
			BounceRate:         &bounceRate,
			AvgSessionDuration: &avgSessionDuration,
			ConversionRate:     &conversionRate,
			AvgLoadTime:        &avgLoadTime,
			CreatedAt:          time.Now(),
		},
	}

	req := &MetricsExportRequest{
		WorkspaceID: workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Setup expectations
	mockMetricsRepo.On("GetDailyMetrics", mock.Anything, workspaceID, startDate, endDate).Return(metrics, nil)

	// Execute
	var buf bytes.Buffer
	err := service.ExportMetricsCSV(context.Background(), req, &buf)

	// Assert
	assert.NoError(t, err)
	
	// Parse CSV and verify content
	csvReader := csv.NewReader(&buf)
	records, err := csvReader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 2) // Header + 1 data row
	
	// Verify header
	expectedHeader := []string{
		"date", "workspace_id", "total_sessions", "total_page_views",
		"unique_visitors", "bounce_rate", "avg_session_duration",
		"conversion_rate", "avg_load_time",
	}
	assert.Equal(t, expectedHeader, records[0])
	
	// Verify data row
	assert.Equal(t, startDate.Format("2006-01-02"), records[1][0])
	assert.Equal(t, workspaceID.String(), records[1][1])
	assert.Equal(t, "100", records[1][2])
	assert.Equal(t, "250", records[1][3])
	assert.Equal(t, "80", records[1][4])
	assert.Equal(t, "25.50", records[1][5])
	assert.Equal(t, "180", records[1][6])
	assert.Equal(t, "3.20", records[1][7])
	assert.Equal(t, "1200", records[1][8])

	mockMetricsRepo.AssertExpectations(t)
}

func TestExportService_ExportSessionsCSV(t *testing.T) {
	// Setup
	mockAnalysisRepo := &mockAnalysisRepository{}
	mockMetricsRepo := &mockMetricsRepository{}
	mockSessionRepo := &mockSessionRepository{}
	mockEventRepo := &mockEventRepository{}
	logger := logrus.New()

	service := NewExportService(mockAnalysisRepo, mockMetricsRepo, mockSessionRepo, mockEventRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()
	userID := "user123"
	durationSeconds := 300
	deviceType := "desktop"
	browser := "Chrome"
	country := "US"
	referrer := "https://google.com"

	// Mock data
	sessions := []*models.Session{
		{
			ID:              sessionID,
			WorkspaceID:     workspaceID,
			UserID:          &userID,
			StartedAt:       time.Now().Add(-time.Hour),
			EndedAt:         &time.Time{},
			DurationSeconds: &durationSeconds,
			PageViews:       5,
			EventsCount:     12,
			DeviceType:      &deviceType,
			Browser:         &browser,
			Country:         &country,
			Referrer:        &referrer,
		},
	}

	req := &SessionExportRequest{
		WorkspaceID: workspaceID,
		UserID:      &userID,
	}

	// Setup expectations
	mockSessionRepo.On("GetByWorkspaceID", mock.Anything, workspaceID, mock.AnythingOfType("*repositories.SessionFilters")).Return(sessions, nil)

	// Execute
	var buf bytes.Buffer
	err := service.ExportSessionsCSV(context.Background(), req, &buf)

	// Assert
	assert.NoError(t, err)
	
	// Parse CSV and verify content
	csvReader := csv.NewReader(&buf)
	records, err := csvReader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 2) // Header + 1 data row
	
	// Verify header
	expectedHeader := []string{
		"id", "workspace_id", "user_id", "started_at", "ended_at",
		"duration_seconds", "page_views", "events_count", "device_type",
		"browser", "country", "referrer",
	}
	assert.Equal(t, expectedHeader, records[0])
	
	// Verify data row
	assert.Equal(t, sessionID.String(), records[1][0])
	assert.Equal(t, workspaceID.String(), records[1][1])
	assert.Equal(t, userID, records[1][2])
	assert.Equal(t, "300", records[1][5])
	assert.Equal(t, "5", records[1][6])
	assert.Equal(t, "12", records[1][7])
	assert.Equal(t, deviceType, records[1][8])
	assert.Equal(t, browser, records[1][9])
	assert.Equal(t, country, records[1][10])
	assert.Equal(t, referrer, records[1][11])

	mockSessionRepo.AssertExpectations(t)
}

func TestExportService_ExportEventsCSV(t *testing.T) {
	// Setup
	mockAnalysisRepo := &mockAnalysisRepository{}
	mockMetricsRepo := &mockMetricsRepository{}
	mockSessionRepo := &mockSessionRepository{}
	mockEventRepo := &mockEventRepository{}
	logger := logrus.New()

	service := NewExportService(mockAnalysisRepo, mockMetricsRepo, mockSessionRepo, mockEventRepo, logger)

	workspaceID := uuid.New()
	sessionID := uuid.New()
	eventID := uuid.New()
	url := "https://example.com/page"

	// Mock data
	events := []*models.Event{
		{
			ID:          eventID,
			SessionID:   sessionID,
			WorkspaceID: workspaceID,
			EventType:   "pageview",
			URL:         &url,
			Timestamp:   time.Now(),
			Properties: map[string]interface{}{
				"page_title": "Example Page",
				"referrer":   "https://google.com",
			},
			CreatedAt: time.Now(),
		},
	}

	req := &EventExportRequest{
		WorkspaceID: workspaceID,
		SessionID:   &sessionID,
	}

	// Setup expectations
	mockEventRepo.On("GetByWorkspaceID", mock.Anything, workspaceID, mock.AnythingOfType("*repositories.EventFilters")).Return(events, nil)

	// Execute
	var buf bytes.Buffer
	err := service.ExportEventsCSV(context.Background(), req, &buf)

	// Assert
	assert.NoError(t, err)
	
	// Parse CSV and verify content
	csvReader := csv.NewReader(&buf)
	records, err := csvReader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 2) // Header + 1 data row
	
	// Verify header
	expectedHeader := []string{
		"id", "session_id", "workspace_id", "event_type", "url",
		"timestamp", "properties", "created_at",
	}
	assert.Equal(t, expectedHeader, records[0])
	
	// Verify data row
	assert.Equal(t, eventID.String(), records[1][0])
	assert.Equal(t, sessionID.String(), records[1][1])
	assert.Equal(t, workspaceID.String(), records[1][2])
	assert.Equal(t, "pageview", records[1][3])
	assert.Equal(t, url, records[1][4])
	
	// Verify properties JSON
	assert.True(t, strings.Contains(records[1][6], "page_title"))
	assert.True(t, strings.Contains(records[1][6], "Example Page"))

	mockEventRepo.AssertExpectations(t)
}

func TestExportService_ExtractJSONField(t *testing.T) {
	service := &exportService{}

	tests := []struct {
		name     string
		data     map[string]interface{}
		field    string
		expected string
	}{
		{
			name:     "existing field",
			data:     map[string]interface{}{"score": 85},
			field:    "score",
			expected: "85",
		},
		{
			name:     "missing field",
			data:     map[string]interface{}{"other": "value"},
			field:    "score",
			expected: "",
		},
		{
			name:     "nil data",
			data:     nil,
			field:    "score",
			expected: "",
		},
		{
			name:     "string field",
			data:     map[string]interface{}{"title": "Example Title"},
			field:    "title",
			expected: "Example Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractJSONField(tt.data, tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}