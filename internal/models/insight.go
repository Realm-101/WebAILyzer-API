package models

import (
	"time"
	"github.com/google/uuid"
)

// InsightType represents the type of insight
type InsightType string

const (
	InsightTypePerformanceBottleneck InsightType = "performance_bottleneck"
	InsightTypeConversionFunnel      InsightType = "conversion_funnel"
	InsightTypeSEOOptimization       InsightType = "seo_optimization"
	InsightTypeAccessibilityIssue    InsightType = "accessibility_issue"
	InsightTypeSecurityVulnerability InsightType = "security_vulnerability"
)

// Priority represents the priority level of an insight
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityCritical Priority = "critical"
)

// InsightStatus represents the status of an insight
type InsightStatus string

const (
	InsightStatusPending   InsightStatus = "pending"
	InsightStatusApplied   InsightStatus = "applied"
	InsightStatusDismissed InsightStatus = "dismissed"
)

// Insight represents an AI-generated insight
type Insight struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	WorkspaceID     uuid.UUID              `json:"workspace_id" db:"workspace_id"`
	InsightType     InsightType            `json:"insight_type" db:"insight_type"`
	Priority        Priority               `json:"priority" db:"priority"`
	Title           string                 `json:"title" db:"title"`
	Description     *string                `json:"description,omitempty" db:"description"`
	ImpactScore     *int                   `json:"impact_score,omitempty" db:"impact_score"`
	EffortScore     *int                   `json:"effort_score,omitempty" db:"effort_score"`
	Recommendations map[string]interface{} `json:"recommendations" db:"recommendations"`
	DataSource      map[string]interface{} `json:"data_source" db:"data_source"`
	Status          InsightStatus          `json:"status" db:"status"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// Recommendation represents a specific recommendation within an insight
type Recommendation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Effort      string `json:"effort"`
}