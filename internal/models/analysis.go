package models

import (
	"time"
	"github.com/google/uuid"
)

// AnalysisResult represents the result of a comprehensive website analysis
type AnalysisResult struct {
	ID                    uuid.UUID              `json:"id" db:"id"`
	WorkspaceID           uuid.UUID              `json:"workspace_id" db:"workspace_id"`
	SessionID             *uuid.UUID             `json:"session_id,omitempty" db:"session_id"`
	URL                   string                 `json:"url" db:"url"`
	Technologies          map[string]interface{} `json:"technologies" db:"technologies"`
	PerformanceMetrics    map[string]interface{} `json:"performance" db:"performance_metrics"`
	SEOMetrics           map[string]interface{} `json:"seo" db:"seo_metrics"`
	AccessibilityMetrics map[string]interface{} `json:"accessibility" db:"accessibility_metrics"`
	SecurityMetrics      map[string]interface{} `json:"security" db:"security_metrics"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`
}

// AnalysisRequest represents a request for website analysis
type AnalysisRequest struct {
	URL         string           `json:"url" validate:"required,url"`
	SessionID   *uuid.UUID       `json:"session_id,omitempty"`
	WorkspaceID uuid.UUID        `json:"workspace_id" validate:"required"`
	Options     AnalysisOptions  `json:"options"`
}

// AnalysisOptions configures what analysis modules to run
type AnalysisOptions struct {
	IncludePerformance   bool   `json:"include_performance"`
	IncludeSEO          bool   `json:"include_seo"`
	IncludeAccessibility bool   `json:"include_accessibility"`
	IncludeSecurity     bool   `json:"include_security"`
	UserAgent           string `json:"user_agent,omitempty"`
}