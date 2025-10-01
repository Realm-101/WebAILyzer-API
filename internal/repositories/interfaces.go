package repositories

import (
	"context"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/google/uuid"
	"time"
)

// AnalysisFilters represents filters for analysis queries
type AnalysisFilters struct {
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	SessionID   *uuid.UUID `json:"session_id,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// AnalysisRepository defines the interface for analysis data operations
type AnalysisRepository interface {
	Create(ctx context.Context, analysis *models.AnalysisResult) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AnalysisResult, error)
	GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.AnalysisResult, error)
	GetBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AnalysisResult, error)
	GetByFilters(ctx context.Context, filters *AnalysisFilters) ([]*models.AnalysisResult, error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *AnalysisFilters) ([]*models.AnalysisResult, error)
	Update(ctx context.Context, analysis *models.AnalysisResult) error
	Delete(ctx context.Context, id uuid.UUID) error
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

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	GetSessionsByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.Session, error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *SessionFilters) ([]*models.Session, error)
}

// EventFilters represents filters for event queries
type EventFilters struct {
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	SessionID   *uuid.UUID `json:"session_id,omitempty"`
	EventType   *string    `json:"event_type,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// EventRepository defines the interface for event data operations
type EventRepository interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	CreateEvents(ctx context.Context, events []*models.Event) error
	GetEventsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.Event, error)
	GetEventsByWorkspace(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) ([]*models.Event, error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *EventFilters) ([]*models.Event, error)
}

// InsightRepository defines the interface for insight data operations
type InsightRepository interface {
	Create(ctx context.Context, insight *models.Insight) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Insight, error)
	GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, status *models.InsightStatus, limit, offset int) ([]*models.Insight, error)
	GetByFilters(ctx context.Context, workspaceID uuid.UUID, status *models.InsightStatus, insightType *models.InsightType, priority *models.Priority, limit, offset int) ([]*models.Insight, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.InsightStatus) error
	Update(ctx context.Context, insight *models.Insight) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MetricsRepository defines the interface for metrics data operations
type MetricsRepository interface {
	CreateDailyMetrics(ctx context.Context, metrics *models.DailyMetrics) error
	GetDailyMetrics(ctx context.Context, workspaceID uuid.UUID, startDate, endDate time.Time) ([]*models.DailyMetrics, error)
	UpdateDailyMetrics(ctx context.Context, metrics *models.DailyMetrics) error
	GetMetricsByWorkspace(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) (*models.MetricsResponse, error)
}

// WorkspaceRepository defines the interface for workspace data operations
type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *models.Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*models.Workspace, error)
	Update(ctx context.Context, workspace *models.Workspace) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Workspace, error)
}