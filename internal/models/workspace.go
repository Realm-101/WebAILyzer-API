package models

import (
	"time"
	"github.com/google/uuid"
)

// Workspace represents a workspace with API access
type Workspace struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	APIKey      string    `json:"api_key" db:"api_key"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	RateLimit   int       `json:"rate_limit" db:"rate_limit"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AuthContext represents the authentication context for a request
type AuthContext struct {
	WorkspaceID uuid.UUID
	APIKey      string
	RateLimit   int
}