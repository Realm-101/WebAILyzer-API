package models

import (
	"time"
	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	WorkspaceID      uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	UserID           *string    `json:"user_id,omitempty" db:"user_id"`
	StartedAt        time.Time  `json:"started_at" db:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	DurationSeconds  *int       `json:"duration_seconds,omitempty" db:"duration_seconds"`
	PageViews        int        `json:"page_views" db:"page_views"`
	EventsCount      int        `json:"events_count" db:"events_count"`
	DeviceType       *string    `json:"device_type,omitempty" db:"device_type"`
	Browser          *string    `json:"browser,omitempty" db:"browser"`
	Country          *string    `json:"country,omitempty" db:"country"`
	Referrer         *string    `json:"referrer,omitempty" db:"referrer"`
}

// Event represents a tracked user event
type Event struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	SessionID   uuid.UUID              `json:"session_id" db:"session_id"`
	WorkspaceID uuid.UUID              `json:"workspace_id" db:"workspace_id"`
	EventType   string                 `json:"event_type" db:"event_type"`
	URL         *string                `json:"url,omitempty" db:"url"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Properties  map[string]interface{} `json:"properties" db:"properties"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// EventTrackingRequest represents a request to track events
type EventTrackingRequest struct {
	SessionID   uuid.UUID `json:"session_id" validate:"required"`
	WorkspaceID uuid.UUID `json:"workspace_id" validate:"required"`
	Events      []Event   `json:"events" validate:"required,min=1"`
}