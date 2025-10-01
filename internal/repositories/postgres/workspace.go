package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
)

// WorkspaceRepository implements the workspace repository interface for PostgreSQL
type WorkspaceRepository struct {
	pool *pgxpool.Pool
}

// NewWorkspaceRepository creates a new workspace repository
func NewWorkspaceRepository(pool *pgxpool.Pool) repositories.WorkspaceRepository {
	return &WorkspaceRepository{
		pool: pool,
	}
}

// Create creates a new workspace
func (r *WorkspaceRepository) Create(ctx context.Context, workspace *models.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, api_key, is_active, rate_limit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.APIKey,
		workspace.IsActive,
		workspace.RateLimit,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	return nil
}

// GetByID retrieves a workspace by ID
func (r *WorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	query := `
		SELECT id, name, api_key, is_active, rate_limit, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`

	var workspace models.Workspace
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.APIKey,
		&workspace.IsActive,
		&workspace.RateLimit,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace by ID: %w", err)
	}

	return &workspace, nil
}

// GetByAPIKey retrieves a workspace by API key
func (r *WorkspaceRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Workspace, error) {
	query := `
		SELECT id, name, api_key, is_active, rate_limit, created_at, updated_at
		FROM workspaces
		WHERE api_key = $1
	`

	var workspace models.Workspace
	err := r.pool.QueryRow(ctx, query, apiKey).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.APIKey,
		&workspace.IsActive,
		&workspace.RateLimit,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace by API key: %w", err)
	}

	return &workspace, nil
}

// Update updates a workspace
func (r *WorkspaceRepository) Update(ctx context.Context, workspace *models.Workspace) error {
	query := `
		UPDATE workspaces
		SET name = $2, api_key = $3, is_active = $4, rate_limit = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.APIKey,
		workspace.IsActive,
		workspace.RateLimit,
		workspace.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("workspace not found")
	}

	return nil
}

// Delete deletes a workspace
func (r *WorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspaces WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("workspace not found")
	}

	return nil
}

// List retrieves workspaces with pagination
func (r *WorkspaceRepository) List(ctx context.Context, limit, offset int) ([]*models.Workspace, error) {
	query := `
		SELECT id, name, api_key, is_active, rate_limit, created_at, updated_at
		FROM workspaces
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		var workspace models.Workspace
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.APIKey,
			&workspace.IsActive,
			&workspace.RateLimit,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspace rows: %w", err)
	}

	return workspaces, nil
}