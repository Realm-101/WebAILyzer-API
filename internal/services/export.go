package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
	"github.com/sirupsen/logrus"
)

// ExportService defines the interface for data export operations
type ExportService interface {
	ExportAnalysisResultsCSV(ctx context.Context, req *ExportRequest, writer io.Writer) error
	ExportAnalysisResultsJSON(ctx context.Context, req *ExportRequest, writer io.Writer) error
	ExportMetricsCSV(ctx context.Context, req *MetricsExportRequest, writer io.Writer) error
	ExportMetricsJSON(ctx context.Context, req *MetricsExportRequest, writer io.Writer) error
	ExportSessionsCSV(ctx context.Context, req *SessionExportRequest, writer io.Writer) error
	ExportSessionsJSON(ctx context.Context, req *SessionExportRequest, writer io.Writer) error
	ExportEventsCSV(ctx context.Context, req *EventExportRequest, writer io.Writer) error
	ExportEventsJSON(ctx context.Context, req *EventExportRequest, writer io.Writer) error
}

// ExportRequest represents a request for exporting analysis results
type ExportRequest struct {
	WorkspaceID uuid.UUID  `json:"workspace_id" validate:"required"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	SessionID   *uuid.UUID `json:"session_id,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// MetricsExportRequest represents a request for exporting metrics
type MetricsExportRequest struct {
	WorkspaceID uuid.UUID `json:"workspace_id" validate:"required"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
}

// SessionExportRequest represents a request for exporting sessions
type SessionExportRequest struct {
	WorkspaceID uuid.UUID  `json:"workspace_id" validate:"required"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	UserID      *string    `json:"user_id,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// EventExportRequest represents a request for exporting events
type EventExportRequest struct {
	WorkspaceID uuid.UUID  `json:"workspace_id" validate:"required"`
	SessionID   *uuid.UUID `json:"session_id,omitempty"`
	EventType   *string    `json:"event_type,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// ExportMetadata contains metadata about the exported data
type ExportMetadata struct {
	ExportedAt      time.Time `json:"exported_at"`
	WorkspaceID     uuid.UUID `json:"workspace_id"`
	RecordCount     int       `json:"record_count"`
	DataFreshness   string    `json:"data_freshness"`
	CollectionStart *time.Time `json:"collection_start,omitempty"`
	CollectionEnd   *time.Time `json:"collection_end,omitempty"`
	ExportType      string    `json:"export_type"`
	Format          string    `json:"format"`
}

// exportService implements the ExportService interface
type exportService struct {
	analysisRepo repositories.AnalysisRepository
	metricsRepo  repositories.MetricsRepository
	sessionRepo  repositories.SessionRepository
	eventRepo    repositories.EventRepository
	logger       *logrus.Logger
}

// NewExportService creates a new export service
func NewExportService(
	analysisRepo repositories.AnalysisRepository,
	metricsRepo repositories.MetricsRepository,
	sessionRepo repositories.SessionRepository,
	eventRepo repositories.EventRepository,
	logger *logrus.Logger,
) ExportService {
	return &exportService{
		analysisRepo: analysisRepo,
		metricsRepo:  metricsRepo,
		sessionRepo:  sessionRepo,
		eventRepo:    eventRepo,
		logger:       logger,
	}
}

// ExportAnalysisResultsCSV exports analysis results in CSV format
func (s *exportService) ExportAnalysisResultsCSV(ctx context.Context, req *ExportRequest, writer io.Writer) error {
	filters := &repositories.AnalysisFilters{
		SessionID: req.SessionID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	results, err := s.analysisRepo.GetByWorkspaceID(ctx, req.WorkspaceID, filters)
	if err != nil {
		return fmt.Errorf("failed to get analysis results: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV header
	header := []string{
		"id", "workspace_id", "session_id", "url", "created_at", "updated_at",
		"technologies", "performance_load_time", "performance_score",
		"seo_title", "seo_score", "accessibility_score", "security_score",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, result := range results {
		row := s.analysisResultToCSVRow(result)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportAnalysisResultsJSON exports analysis results in JSON format
func (s *exportService) ExportAnalysisResultsJSON(ctx context.Context, req *ExportRequest, writer io.Writer) error {
	filters := &repositories.AnalysisFilters{
		SessionID: req.SessionID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	results, err := s.analysisRepo.GetByWorkspaceID(ctx, req.WorkspaceID, filters)
	if err != nil {
		return fmt.Errorf("failed to get analysis results: %w", err)
	}

	metadata := &ExportMetadata{
		ExportedAt:      time.Now(),
		WorkspaceID:     req.WorkspaceID,
		RecordCount:     len(results),
		DataFreshness:   "real-time",
		CollectionStart: req.StartDate,
		CollectionEnd:   req.EndDate,
		ExportType:      "analysis_results",
		Format:          "json",
	}

	exportData := map[string]interface{}{
		"metadata": metadata,
		"data":     results,
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(exportData)
}

// ExportMetricsCSV exports metrics in CSV format
func (s *exportService) ExportMetricsCSV(ctx context.Context, req *MetricsExportRequest, writer io.Writer) error {
	metrics, err := s.metricsRepo.GetDailyMetrics(ctx, req.WorkspaceID, req.StartDate, req.EndDate)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV header
	header := []string{
		"date", "workspace_id", "total_sessions", "total_page_views",
		"unique_visitors", "bounce_rate", "avg_session_duration",
		"conversion_rate", "avg_load_time",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, metric := range metrics {
		row := s.metricsToCSVRow(metric)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportMetricsJSON exports metrics in JSON format
func (s *exportService) ExportMetricsJSON(ctx context.Context, req *MetricsExportRequest, writer io.Writer) error {
	metrics, err := s.metricsRepo.GetDailyMetrics(ctx, req.WorkspaceID, req.StartDate, req.EndDate)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	metadata := &ExportMetadata{
		ExportedAt:      time.Now(),
		WorkspaceID:     req.WorkspaceID,
		RecordCount:     len(metrics),
		DataFreshness:   "daily-aggregated",
		CollectionStart: &req.StartDate,
		CollectionEnd:   &req.EndDate,
		ExportType:      "metrics",
		Format:          "json",
	}

	exportData := map[string]interface{}{
		"metadata": metadata,
		"data":     metrics,
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(exportData)
}

// ExportSessionsCSV exports sessions in CSV format
func (s *exportService) ExportSessionsCSV(ctx context.Context, req *SessionExportRequest, writer io.Writer) error {
	filters := &repositories.SessionFilters{
		WorkspaceID: req.WorkspaceID,
		UserID:      req.UserID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	sessions, err := s.sessionRepo.GetByWorkspaceID(ctx, req.WorkspaceID, filters)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV header
	header := []string{
		"id", "workspace_id", "user_id", "started_at", "ended_at",
		"duration_seconds", "page_views", "events_count", "device_type",
		"browser", "country", "referrer",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, session := range sessions {
		row := s.sessionToCSVRow(session)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportSessionsJSON exports sessions in JSON format
func (s *exportService) ExportSessionsJSON(ctx context.Context, req *SessionExportRequest, writer io.Writer) error {
	filters := &repositories.SessionFilters{
		WorkspaceID: req.WorkspaceID,
		UserID:      req.UserID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	sessions, err := s.sessionRepo.GetByWorkspaceID(ctx, req.WorkspaceID, filters)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	metadata := &ExportMetadata{
		ExportedAt:      time.Now(),
		WorkspaceID:     req.WorkspaceID,
		RecordCount:     len(sessions),
		DataFreshness:   "real-time",
		CollectionStart: req.StartTime,
		CollectionEnd:   req.EndTime,
		ExportType:      "sessions",
		Format:          "json",
	}

	exportData := map[string]interface{}{
		"metadata": metadata,
		"data":     sessions,
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(exportData)
}

// ExportEventsCSV exports events in CSV format
func (s *exportService) ExportEventsCSV(ctx context.Context, req *EventExportRequest, writer io.Writer) error {
	filters := &repositories.EventFilters{
		WorkspaceID: req.WorkspaceID,
		SessionID:   req.SessionID,
		EventType:   req.EventType,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	events, err := s.eventRepo.GetByWorkspaceID(ctx, req.WorkspaceID, filters)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV header
	header := []string{
		"id", "session_id", "workspace_id", "event_type", "url",
		"timestamp", "properties", "created_at",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, event := range events {
		row := s.eventToCSVRow(event)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportEventsJSON exports events in JSON format
func (s *exportService) ExportEventsJSON(ctx context.Context, req *EventExportRequest, writer io.Writer) error {
	filters := &repositories.EventFilters{
		WorkspaceID: req.WorkspaceID,
		SessionID:   req.SessionID,
		EventType:   req.EventType,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	events, err := s.eventRepo.GetByWorkspaceID(ctx, req.WorkspaceID, filters)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	metadata := &ExportMetadata{
		ExportedAt:      time.Now(),
		WorkspaceID:     req.WorkspaceID,
		RecordCount:     len(events),
		DataFreshness:   "real-time",
		CollectionStart: req.StartTime,
		CollectionEnd:   req.EndTime,
		ExportType:      "events",
		Format:          "json",
	}

	exportData := map[string]interface{}{
		"metadata": metadata,
		"data":     events,
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(exportData)
}

// Helper methods for CSV conversion

func (s *exportService) analysisResultToCSVRow(result *models.AnalysisResult) []string {
	sessionID := ""
	if result.SessionID != nil {
		sessionID = result.SessionID.String()
	}

	// Extract key metrics from JSON fields
	performanceLoadTime := s.extractJSONField(result.PerformanceMetrics, "load_time_ms")
	performanceScore := s.extractJSONField(result.PerformanceMetrics, "score")
	seoTitle := s.extractJSONField(result.SEOMetrics, "title")
	seoScore := s.extractJSONField(result.SEOMetrics, "score")
	accessibilityScore := s.extractJSONField(result.AccessibilityMetrics, "score")
	securityScore := s.extractJSONField(result.SecurityMetrics, "score")

	technologiesJSON, _ := json.Marshal(result.Technologies)

	return []string{
		result.ID.String(),
		result.WorkspaceID.String(),
		sessionID,
		result.URL,
		result.CreatedAt.Format(time.RFC3339),
		result.UpdatedAt.Format(time.RFC3339),
		string(technologiesJSON),
		performanceLoadTime,
		performanceScore,
		seoTitle,
		seoScore,
		accessibilityScore,
		securityScore,
	}
}

func (s *exportService) metricsToCSVRow(metric *models.DailyMetrics) []string {
	bounceRate := ""
	if metric.BounceRate != nil {
		bounceRate = fmt.Sprintf("%.2f", *metric.BounceRate)
	}

	avgSessionDuration := ""
	if metric.AvgSessionDuration != nil {
		avgSessionDuration = strconv.Itoa(*metric.AvgSessionDuration)
	}

	conversionRate := ""
	if metric.ConversionRate != nil {
		conversionRate = fmt.Sprintf("%.2f", *metric.ConversionRate)
	}

	avgLoadTime := ""
	if metric.AvgLoadTime != nil {
		avgLoadTime = strconv.Itoa(*metric.AvgLoadTime)
	}

	return []string{
		metric.Date.Format("2006-01-02"),
		metric.WorkspaceID.String(),
		strconv.Itoa(metric.TotalSessions),
		strconv.Itoa(metric.TotalPageViews),
		strconv.Itoa(metric.UniqueVisitors),
		bounceRate,
		avgSessionDuration,
		conversionRate,
		avgLoadTime,
	}
}

func (s *exportService) sessionToCSVRow(session *models.Session) []string {
	userID := ""
	if session.UserID != nil {
		userID = *session.UserID
	}

	endedAt := ""
	if session.EndedAt != nil {
		endedAt = session.EndedAt.Format(time.RFC3339)
	}

	durationSeconds := ""
	if session.DurationSeconds != nil {
		durationSeconds = strconv.Itoa(*session.DurationSeconds)
	}

	deviceType := ""
	if session.DeviceType != nil {
		deviceType = *session.DeviceType
	}

	browser := ""
	if session.Browser != nil {
		browser = *session.Browser
	}

	country := ""
	if session.Country != nil {
		country = *session.Country
	}

	referrer := ""
	if session.Referrer != nil {
		referrer = *session.Referrer
	}

	return []string{
		session.ID.String(),
		session.WorkspaceID.String(),
		userID,
		session.StartedAt.Format(time.RFC3339),
		endedAt,
		durationSeconds,
		strconv.Itoa(session.PageViews),
		strconv.Itoa(session.EventsCount),
		deviceType,
		browser,
		country,
		referrer,
	}
}

func (s *exportService) eventToCSVRow(event *models.Event) []string {
	url := ""
	if event.URL != nil {
		url = *event.URL
	}

	propertiesJSON, _ := json.Marshal(event.Properties)

	return []string{
		event.ID.String(),
		event.SessionID.String(),
		event.WorkspaceID.String(),
		event.EventType,
		url,
		event.Timestamp.Format(time.RFC3339),
		string(propertiesJSON),
		event.CreatedAt.Format(time.RFC3339),
	}
}

func (s *exportService) extractJSONField(data map[string]interface{}, field string) string {
	if data == nil {
		return ""
	}
	
	if value, exists := data[field]; exists {
		return fmt.Sprintf("%v", value)
	}
	
	return ""
}