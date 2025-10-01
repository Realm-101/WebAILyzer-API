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

// analysisRepository implements the AnalysisRepository interface for PostgreSQL
type analysisRepository struct {
	db *pgxpool.Pool
}

// NewAnalysisRepository creates a new PostgreSQL analysis repository
func NewAnalysisRepository(db *pgxpool.Pool) repositories.AnalysisRepository {
	return &analysisRepository{db: db}
}

// Create inserts a new analysis result into the database
func (r *analysisRepository) Create(ctx context.Context, analysis *models.AnalysisResult) error {
	query := `
		INSERT INTO analysis_results (
			id, workspace_id, session_id, url, technologies, 
			performance_metrics, seo_metrics, accessibility_metrics, 
			security_metrics, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(ctx, query,
		analysis.ID,
		analysis.WorkspaceID,
		analysis.SessionID,
		analysis.URL,
		analysis.Technologies,
		analysis.PerformanceMetrics,
		analysis.SEOMetrics,
		analysis.AccessibilityMetrics,
		analysis.SecurityMetrics,
		analysis.CreatedAt,
		analysis.UpdatedAt,
	)
	return err
}

// GetByID retrieves an analysis result by its ID
func (r *analysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AnalysisResult, error) {
	query := `
		SELECT id, workspace_id, session_id, url, technologies,
			   performance_metrics, seo_metrics, accessibility_metrics,
			   security_metrics, created_at, updated_at
		FROM analysis_results WHERE id = $1`

	var analysis models.AnalysisResult
	err := r.db.QueryRow(ctx, query, id).Scan(
		&analysis.ID,
		&analysis.WorkspaceID,
		&analysis.SessionID,
		&analysis.URL,
		&analysis.Technologies,
		&analysis.PerformanceMetrics,
		&analysis.SEOMetrics,
		&analysis.AccessibilityMetrics,
		&analysis.SecurityMetrics,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &analysis, nil
}

// GetByWorkspace retrieves analysis results for a workspace with pagination
func (r *analysisRepository) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.AnalysisResult, error) {
	query := `
		SELECT id, workspace_id, session_id, url, technologies,
			   performance_metrics, seo_metrics, accessibility_metrics,
			   security_metrics, created_at, updated_at
		FROM analysis_results 
		WHERE workspace_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.AnalysisResult
	for rows.Next() {
		var analysis models.AnalysisResult
		err := rows.Scan(
			&analysis.ID,
			&analysis.WorkspaceID,
			&analysis.SessionID,
			&analysis.URL,
			&analysis.Technologies,
			&analysis.PerformanceMetrics,
			&analysis.SEOMetrics,
			&analysis.AccessibilityMetrics,
			&analysis.SecurityMetrics,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, &analysis)
	}
	return results, rows.Err()
}

// GetBySession retrieves analysis results for a session
func (r *analysisRepository) GetBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AnalysisResult, error) {
	query := `
		SELECT id, workspace_id, session_id, url, technologies,
			   performance_metrics, seo_metrics, accessibility_metrics,
			   security_metrics, created_at, updated_at
		FROM analysis_results 
		WHERE session_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.AnalysisResult
	for rows.Next() {
		var analysis models.AnalysisResult
		err := rows.Scan(
			&analysis.ID,
			&analysis.WorkspaceID,
			&analysis.SessionID,
			&analysis.URL,
			&analysis.Technologies,
			&analysis.PerformanceMetrics,
			&analysis.SEOMetrics,
			&analysis.AccessibilityMetrics,
			&analysis.SecurityMetrics,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, &analysis)
	}
	return results, rows.Err()
}

// Update updates an existing analysis result
func (r *analysisRepository) Update(ctx context.Context, analysis *models.AnalysisResult) error {
	query := `
		UPDATE analysis_results SET
			technologies = $2,
			performance_metrics = $3,
			seo_metrics = $4,
			accessibility_metrics = $5,
			security_metrics = $6,
			updated_at = $7
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		analysis.ID,
		analysis.Technologies,
		analysis.PerformanceMetrics,
		analysis.SEOMetrics,
		analysis.AccessibilityMetrics,
		analysis.SecurityMetrics,
		analysis.UpdatedAt,
	)
	return err
}

// Delete removes an analysis result from the database
func (r *analysisRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM analysis_results WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// GetByFilters retrieves analysis results based on provided filters
func (r *analysisRepository) GetByFilters(ctx context.Context, filters *repositories.AnalysisFilters) ([]*models.AnalysisResult, error) {
	query := `
		SELECT id, workspace_id, session_id, url, technologies,
			   performance_metrics, seo_metrics, accessibility_metrics,
			   security_metrics, created_at, updated_at
		FROM analysis_results 
		WHERE workspace_id = $1`
	
	args := []interface{}{filters.WorkspaceID}
	argIndex := 2

	// Add optional filters
	if filters.SessionID != nil {
		query += ` AND session_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.SessionID)
		argIndex++
	}

	if filters.StartDate != nil {
		query += ` AND created_at >= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		query += ` AND created_at <= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filters.EndDate)
		argIndex++
	}

	query += ` ORDER BY created_at DESC`

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

	var results []*models.AnalysisResult
	for rows.Next() {
		var analysis models.AnalysisResult
		err := rows.Scan(
			&analysis.ID,
			&analysis.WorkspaceID,
			&analysis.SessionID,
			&analysis.URL,
			&analysis.Technologies,
			&analysis.PerformanceMetrics,
			&analysis.SEOMetrics,
			&analysis.AccessibilityMetrics,
			&analysis.SecurityMetrics,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, &analysis)
	}
	return results, rows.Err()
}

// GetByWorkspaceID retrieves analysis results for a workspace with filters
func (r *analysisRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.AnalysisFilters) ([]*models.AnalysisResult, error) {
	// Use the existing GetByFilters method with workspace ID set
	if filters == nil {
		filters = &repositories.AnalysisFilters{
			WorkspaceID: workspaceID,
		}
	} else {
		filters.WorkspaceID = workspaceID
	}
	
	return r.GetByFilters(ctx, filters)
}