package models

import (
	"time"
	"github.com/google/uuid"
)

// DailyMetrics represents aggregated daily metrics for a workspace
type DailyMetrics struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	WorkspaceID          uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Date                 time.Time  `json:"date" db:"date"`
	TotalSessions        int        `json:"total_sessions" db:"total_sessions"`
	TotalPageViews       int        `json:"total_page_views" db:"total_page_views"`
	UniqueVisitors       int        `json:"unique_visitors" db:"unique_visitors"`
	BounceRate           *float64   `json:"bounce_rate,omitempty" db:"bounce_rate"`
	AvgSessionDuration   *int       `json:"avg_session_duration,omitempty" db:"avg_session_duration"`
	ConversionRate       *float64   `json:"conversion_rate,omitempty" db:"conversion_rate"`
	AvgLoadTime          *int       `json:"avg_load_time,omitempty" db:"avg_load_time"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}

// MetricsRequest represents a request for metrics data
type MetricsRequest struct {
	WorkspaceID uuid.UUID  `json:"workspace_id" validate:"required"`
	StartDate   time.Time  `json:"start_date" validate:"required"`
	EndDate     time.Time  `json:"end_date" validate:"required"`
	Granularity string     `json:"granularity" validate:"required,oneof=hourly daily weekly monthly"`
}

// MetricsResponse represents the response containing metrics data
type MetricsResponse struct {
	Metrics    map[string]MetricData `json:"metrics"`
	KPIs       []KPI                 `json:"kpis"`
	Anomalies  []Anomaly             `json:"anomalies"`
}

// MetricData represents a single metric with trend information
type MetricData struct {
	Current    float64     `json:"current"`
	Previous   float64     `json:"previous"`
	Trend      string      `json:"trend"`
	DataPoints []DataPoint `json:"data_points"`
}

// DataPoint represents a single data point in a time series
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// KPI represents a key performance indicator
type KPI struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Target      float64 `json:"target"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}

// Anomaly represents a detected anomaly in the data
type Anomaly struct {
	Metric      string    `json:"metric"`
	Timestamp   time.Time `json:"timestamp"`
	Expected    float64   `json:"expected"`
	Actual      float64   `json:"actual"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
}