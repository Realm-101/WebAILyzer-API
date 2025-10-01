package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// sessionRepository implements the SessionRepository interface for PostgreSQL
type sessionRepository struct {
	db *pgxpool.Pool
}

// NewSessionRepository creates a new PostgreSQL session repository
func NewSessionRepository(db *pgxpool.Pool) repositories.SessionRepository {
	return &sessionRepository{db: db}
}

// CreateSession inserts a new session into the database
func (r *sessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (
			id, workspace_id, user_id, started_at, ended_at, 
			duration_seconds, page_views, events_count, device_type, 
			browser, country, referrer
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.WorkspaceID,
		session.UserID,
		session.StartedAt,
		session.EndedAt,
		session.DurationSeconds,
		session.PageViews,
		session.EventsCount,
		session.DeviceType,
		session.Browser,
		session.Country,
		session.Referrer,
	)
	return err
}

// GetSessionByID retrieves a session by its ID
func (r *sessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	query := `
		SELECT id, workspace_id, user_id, started_at, ended_at,
			   duration_seconds, page_views, events_count, device_type,
			   browser, country, referrer
		FROM sessions WHERE id = $1`

	var session models.Session
	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.WorkspaceID,
		&session.UserID,
		&session.StartedAt,
		&session.EndedAt,
		&session.DurationSeconds,
		&session.PageViews,
		&session.EventsCount,
		&session.DeviceType,
		&session.Browser,
		&session.Country,
		&session.Referrer,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// UpdateSession updates an existing session
func (r *sessionRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE sessions SET
			ended_at = $2,
			duration_seconds = $3,
			page_views = $4,
			events_count = $5,
			device_type = $6,
			browser = $7,
			country = $8,
			referrer = $9
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.EndedAt,
		session.DurationSeconds,
		session.PageViews,
		session.EventsCount,
		session.DeviceType,
		session.Browser,
		session.Country,
		session.Referrer,
	)
	return err
}

// GetSessionsByWorkspace retrieves sessions for a workspace with pagination
func (r *sessionRepository) GetSessionsByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.Session, error) {
	query := `
		SELECT id, workspace_id, user_id, started_at, ended_at,
			   duration_seconds, page_views, events_count, device_type,
			   browser, country, referrer
		FROM sessions 
		WHERE workspace_id = $1 
		ORDER BY started_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.ID,
			&session.WorkspaceID,
			&session.UserID,
			&session.StartedAt,
			&session.EndedAt,
			&session.DurationSeconds,
			&session.PageViews,
			&session.EventsCount,
			&session.DeviceType,
			&session.Browser,
			&session.Country,
			&session.Referrer,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}
	return sessions, rows.Err()
}

// GetByWorkspaceID retrieves sessions for a workspace with filters
func (r *sessionRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.SessionFilters) ([]*models.Session, error) {
	query := `
		SELECT id, workspace_id, user_id, started_at, ended_at,
			   duration_seconds, page_views, events_count, device_type,
			   browser, country, referrer
		FROM sessions 
		WHERE workspace_id = $1`
	
	args := []interface{}{workspaceID}
	argIndex := 2

	// Add optional filters
	if filters.UserID != nil {
		query += ` AND user_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.StartTime != nil {
		query += ` AND started_at >= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.StartTime)
		argIndex++
	}

	if filters.EndTime != nil {
		query += ` AND started_at <= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.EndTime)
		argIndex++
	}

	query += ` ORDER BY started_at DESC`

	// Add pagination
	if filters.Limit > 0 {
		query += ` LIMIT $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += ` OFFSET $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.ID,
			&session.WorkspaceID,
			&session.UserID,
			&session.StartedAt,
			&session.EndedAt,
			&session.DurationSeconds,
			&session.PageViews,
			&session.EventsCount,
			&session.DeviceType,
			&session.Browser,
			&session.Country,
			&session.Referrer,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}
	return sessions, rows.Err()
}