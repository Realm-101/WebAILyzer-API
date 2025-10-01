package services

import (
	"context"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/google/uuid"
	"time"
)

// AnalysisService defines the interface for analysis operations
type AnalysisService interface {
	AnalyzeURL(ctx context.Context, req *models.AnalysisRequest) (*models.AnalysisResult, error)
	BatchAnalyze(ctx context.Context, req *BatchAnalysisRequest) (*BatchAnalysisResult, error)
	GetAnalysisHistory(ctx context.Context, workspaceID uuid.UUID, filters *AnalysisFilters) ([]*models.AnalysisResult, error)
}

// InsightsService defines the interface for insights operations
type InsightsService interface {
	GenerateInsights(ctx context.Context, workspaceID uuid.UUID) ([]*models.Insight, error)
	GetInsights(ctx context.Context, workspaceID uuid.UUID, filters *InsightFilters) ([]*models.Insight, error)
	UpdateInsightStatus(ctx context.Context, insightID uuid.UUID, status models.InsightStatus) error
}

// EventService defines the interface for event tracking operations
type EventService interface {
	TrackEvents(ctx context.Context, req *models.EventTrackingRequest) error
	GetEvents(ctx context.Context, filters *EventFilters) ([]*models.Event, error)
	GetSessions(ctx context.Context, filters *SessionFilters) ([]*models.Session, error)
}

// MetricsService defines the interface for metrics operations
type MetricsService interface {
	GetMetrics(ctx context.Context, req *models.MetricsRequest) (*models.MetricsResponse, error)
	GetKPIs(ctx context.Context, workspaceID uuid.UUID, timeRange TimeRange) (*KPIResponse, error)
	DetectAnomalies(ctx context.Context, workspaceID uuid.UUID) ([]*models.Anomaly, error)
}

// BatchAnalysisRequest represents a request for batch analysis
type BatchAnalysisRequest struct {
	URLs        []string               `json:"urls" validate:"required,min=1"`
	WorkspaceID uuid.UUID              `json:"workspace_id" validate:"required"`
	Options     models.AnalysisOptions `json:"options"`
}

// BatchAnalysisResult represents the result of batch analysis
type BatchAnalysisResult struct {
	BatchID    uuid.UUID                `json:"batch_id"`
	Status     string                   `json:"status"`
	Results    []*models.AnalysisResult `json:"results"`
	FailedURLs []string                 `json:"failed_urls"`
	Progress   BatchProgress            `json:"progress"`
}

// BatchProgress represents the progress of batch processing
type BatchProgress struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

// AnalysisFilters represents filters for analysis queries
type AnalysisFilters struct {
	SessionID *uuid.UUID `json:"session_id,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

// InsightFilters represents filters for insight queries
type InsightFilters struct {
	Status   *models.InsightStatus `json:"status,omitempty"`
	Type     *models.InsightType   `json:"type,omitempty"`
	Priority *models.Priority      `json:"priority,omitempty"`
	Limit    int                   `json:"limit"`
	Offset   int                   `json:"offset"`
}

// EventFilters represents filters for event queries
type EventFilters struct {
	SessionID   *uuid.UUID `json:"session_id,omitempty"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	EventType   *string    `json:"event_type,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// SessionFilters represents filters for session queries
type SessionFilters struct {
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	UserID      *string    `json:"user_id,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// TimeRange represents a time range for metrics queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// KPIResponse represents the response containing KPIs
type KPIResponse struct {
	KPIs      []models.KPI `json:"kpis"`
	Timestamp time.Time    `json:"timestamp"`
}