package postgres

import (
	"context"
	"fmt"
	"time"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// eventRepository implements the EventRepository interface for PostgreSQL
type eventRepository struct {
	db *pgxpool.Pool
}

// NewEventRepository creates a new PostgreSQL event repository
func NewEventRepository(db *pgxpool.Pool) repositories.EventRepository {
	return &eventRepository{db: db}
}

// CreateEvent inserts a new event into the database
func (r *eventRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (
			id, session_id, workspace_id, event_type, url, 
			timestamp, properties, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		event.ID,
		event.SessionID,
		event.WorkspaceID,
		event.EventType,
		event.URL,
		event.Timestamp,
		event.Properties,
		event.CreatedAt,
	)
	return err
}

// CreateEvents inserts multiple events into the database within a transaction
func (r *eventRepository) CreateEvents(ctx context.Context, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO events (
			id, session_id, workspace_id, event_type, url, 
			timestamp, properties, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	batch := &pgx.Batch{}
	for _, event := range events {
		batch.Queue(query,
			event.ID,
			event.SessionID,
			event.WorkspaceID,
			event.EventType,
			event.URL,
			event.Timestamp,
			event.Properties,
			event.CreatedAt,
		)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(events); i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetEventsBySession retrieves events for a specific session
func (r *eventRepository) GetEventsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.Event, error) {
	query := `
		SELECT id, session_id, workspace_id, event_type, url,
			   timestamp, properties, created_at
		FROM events 
		WHERE session_id = $1 
		ORDER BY timestamp ASC`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.SessionID,
			&event.WorkspaceID,
			&event.EventType,
			&event.URL,
			&event.Timestamp,
			&event.Properties,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, rows.Err()
}

// GetEventsByWorkspace retrieves events for a workspace within a time range
func (r *eventRepository) GetEventsByWorkspace(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) ([]*models.Event, error) {
	query := `
		SELECT id, session_id, workspace_id, event_type, url,
			   timestamp, properties, created_at
		FROM events 
		WHERE workspace_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC`

	rows, err := r.db.Query(ctx, query, workspaceID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.SessionID,
			&event.WorkspaceID,
			&event.EventType,
			&event.URL,
			&event.Timestamp,
			&event.Properties,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, rows.Err()
}

// GetByWorkspaceID retrieves events for a workspace with filters
func (r *eventRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.EventFilters) ([]*models.Event, error) {
	query := `
		SELECT id, session_id, workspace_id, event_type, url,
			   timestamp, properties, created_at
		FROM events 
		WHERE workspace_id = $1`
	
	args := []interface{}{workspaceID}
	argIndex := 2

	// Add optional filters
	if filters.SessionID != nil {
		query += ` AND session_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.SessionID)
		argIndex++
	}

	if filters.EventType != nil {
		query += ` AND event_type = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.EventType)
		argIndex++
	}

	if filters.StartTime != nil {
		query += ` AND timestamp >= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.StartTime)
		argIndex++
	}

	if filters.EndTime != nil {
		query += ` AND timestamp <= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.EndTime)
		argIndex++
	}

	query += ` ORDER BY timestamp ASC`

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

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.SessionID,
			&event.WorkspaceID,
			&event.EventType,
			&event.URL,
			&event.Timestamp,
			&event.Properties,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, rows.Err()
}