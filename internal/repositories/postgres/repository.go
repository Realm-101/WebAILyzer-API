package postgres

import (
	"github.com/webailyzer/webailyzer-lite-api/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories holds all repository implementations
type Repositories struct {
	Analysis repositories.AnalysisRepository
	Session  repositories.SessionRepository
	Event    repositories.EventRepository
	Insight  repositories.InsightRepository
	Metrics  repositories.MetricsRepository
}

// NewRepositories creates a new set of PostgreSQL repositories
func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Analysis: NewAnalysisRepository(db),
		Session:  NewSessionRepository(db),
		Event:    NewEventRepository(db),
		Insight:  NewInsightRepository(db),
		Metrics:  NewMetricsRepository(db),
	}
}