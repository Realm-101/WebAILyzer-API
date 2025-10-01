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

// insightRepository implements the InsightRepository interface for PostgreSQL
type insightRepository struct {
	db *pgxpool.Pool
}

// NewInsightRepository creates a new PostgreSQL insight repository
func NewInsightRepository(db *pgxpool.Pool) repositories.InsightRepository {
	return &insightRepository{db: db}
}

// Create inserts a new insight into the database
func (r *insightRepository) Create(ctx context.Context, insight *models.Insight) error {
	query := `
		INSERT INTO insights (
			id, workspace_id, insight_type, priority, title, 
			description, impact_score, effort_score, recommendations, 
			data_source, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(ctx, query,
		insight.ID,
		insight.WorkspaceID,
		insight.InsightType,
		insight.Priority,
		insight.Title,
		insight.Description,
		insight.ImpactScore,
		insight.EffortScore,
		insight.Recommendations,
		insight.DataSource,
		insight.Status,
		insight.CreatedAt,
		insight.UpdatedAt,
	)
	return err
}

// GetByID retrieves an insight by its ID
func (r *insightRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Insight, error) {
	query := `
		SELECT id, workspace_id, insight_type, priority, title,
			   description, impact_score, effort_score, recommendations,
			   data_source, status, created_at, updated_at
		FROM insights WHERE id = $1`

	var insight models.Insight
	err := r.db.QueryRow(ctx, query, id).Scan(
		&insight.ID,
		&insight.WorkspaceID,
		&insight.InsightType,
		&insight.Priority,
		&insight.Title,
		&insight.Description,
		&insight.ImpactScore,
		&insight.EffortScore,
		&insight.Recommendations,
		&insight.DataSource,
		&insight.Status,
		&insight.CreatedAt,
		&insight.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &insight, nil
}

// GetByWorkspace retrieves insights for a workspace with optional status filter and pagination
func (r *insightRepository) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, status *models.InsightStatus, limit, offset int) ([]*models.Insight, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
			SELECT id, workspace_id, insight_type, priority, title,
				   description, impact_score, effort_score, recommendations,
				   data_source, status, created_at, updated_at
			FROM insights 
			WHERE workspace_id = $1 AND status = $2
			ORDER BY created_at DESC 
			LIMIT $3 OFFSET $4`
		args = []interface{}{workspaceID, *status, limit, offset}
	} else {
		query = `
			SELECT id, workspace_id, insight_type, priority, title,
				   description, impact_score, effort_score, recommendations,
				   data_source, status, created_at, updated_at
			FROM insights 
			WHERE workspace_id = $1
			ORDER BY created_at DESC 
			LIMIT $2 OFFSET $3`
		args = []interface{}{workspaceID, limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*models.Insight
	for rows.Next() {
		var insight models.Insight
		err := rows.Scan(
			&insight.ID,
			&insight.WorkspaceID,
			&insight.InsightType,
			&insight.Priority,
			&insight.Title,
			&insight.Description,
			&insight.ImpactScore,
			&insight.EffortScore,
			&insight.Recommendations,
			&insight.DataSource,
			&insight.Status,
			&insight.CreatedAt,
			&insight.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		insights = append(insights, &insight)
	}
	return insights, rows.Err()
}

// GetByFilters retrieves insights for a workspace with comprehensive filtering
func (r *insightRepository) GetByFilters(ctx context.Context, workspaceID uuid.UUID, status *models.InsightStatus, insightType *models.InsightType, priority *models.Priority, limit, offset int) ([]*models.Insight, error) {
	query := `
		SELECT id, workspace_id, insight_type, priority, title,
			   description, impact_score, effort_score, recommendations,
			   data_source, status, created_at, updated_at
		FROM insights 
		WHERE workspace_id = $1`
	
	args := []interface{}{workspaceID}
	argIndex := 2

	// Add status filter
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	// Add type filter
	if insightType != nil {
		query += fmt.Sprintf(" AND insight_type = $%d", argIndex)
		args = append(args, *insightType)
		argIndex++
	}

	// Add priority filter
	if priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argIndex)
		args = append(args, *priority)
		argIndex++
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*models.Insight
	for rows.Next() {
		var insight models.Insight
		err := rows.Scan(
			&insight.ID,
			&insight.WorkspaceID,
			&insight.InsightType,
			&insight.Priority,
			&insight.Title,
			&insight.Description,
			&insight.ImpactScore,
			&insight.EffortScore,
			&insight.Recommendations,
			&insight.DataSource,
			&insight.Status,
			&insight.CreatedAt,
			&insight.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		insights = append(insights, &insight)
	}
	return insights, rows.Err()
}

// UpdateStatus updates the status of an insight
func (r *insightRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.InsightStatus) error {
	query := `
		UPDATE insights SET
			status = $2,
			updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id, status)
	return err
}

// Update updates an existing insight
func (r *insightRepository) Update(ctx context.Context, insight *models.Insight) error {
	query := `
		UPDATE insights SET
			insight_type = $2,
			priority = $3,
			title = $4,
			description = $5,
			impact_score = $6,
			effort_score = $7,
			recommendations = $8,
			data_source = $9,
			status = $10,
			updated_at = $11
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		insight.ID,
		insight.InsightType,
		insight.Priority,
		insight.Title,
		insight.Description,
		insight.ImpactScore,
		insight.EffortScore,
		insight.Recommendations,
		insight.DataSource,
		insight.Status,
		insight.UpdatedAt,
	)
	return err
}

// Delete removes an insight from the database
func (r *insightRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM insights WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}