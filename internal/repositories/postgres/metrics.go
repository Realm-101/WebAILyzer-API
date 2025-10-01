package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// metricsRepository implements the MetricsRepository interface for PostgreSQL
type metricsRepository struct {
	db *pgxpool.Pool
}

// NewMetricsRepository creates a new PostgreSQL metrics repository
func NewMetricsRepository(db *pgxpool.Pool) repositories.MetricsRepository {
	return &metricsRepository{db: db}
}

// CreateDailyMetrics inserts new daily metrics into the database
func (r *metricsRepository) CreateDailyMetrics(ctx context.Context, metrics *models.DailyMetrics) error {
	query := `
		INSERT INTO daily_metrics (
			id, workspace_id, date, total_sessions, total_page_views,
			unique_visitors, bounce_rate, avg_session_duration,
			conversion_rate, avg_load_time, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (workspace_id, date) 
		DO UPDATE SET
			total_sessions = EXCLUDED.total_sessions,
			total_page_views = EXCLUDED.total_page_views,
			unique_visitors = EXCLUDED.unique_visitors,
			bounce_rate = EXCLUDED.bounce_rate,
			avg_session_duration = EXCLUDED.avg_session_duration,
			conversion_rate = EXCLUDED.conversion_rate,
			avg_load_time = EXCLUDED.avg_load_time`

	_, err := r.db.Exec(ctx, query,
		metrics.ID,
		metrics.WorkspaceID,
		metrics.Date,
		metrics.TotalSessions,
		metrics.TotalPageViews,
		metrics.UniqueVisitors,
		metrics.BounceRate,
		metrics.AvgSessionDuration,
		metrics.ConversionRate,
		metrics.AvgLoadTime,
		metrics.CreatedAt,
	)
	return err
}

// GetDailyMetrics retrieves daily metrics for a workspace within a date range
func (r *metricsRepository) GetDailyMetrics(ctx context.Context, workspaceID uuid.UUID, startDate, endDate time.Time) ([]*models.DailyMetrics, error) {
	query := `
		SELECT id, workspace_id, date, total_sessions, total_page_views,
			   unique_visitors, bounce_rate, avg_session_duration,
			   conversion_rate, avg_load_time, created_at
		FROM daily_metrics 
		WHERE workspace_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC`

	rows, err := r.db.Query(ctx, query, workspaceID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*models.DailyMetrics
	for rows.Next() {
		var metric models.DailyMetrics
		err := rows.Scan(
			&metric.ID,
			&metric.WorkspaceID,
			&metric.Date,
			&metric.TotalSessions,
			&metric.TotalPageViews,
			&metric.UniqueVisitors,
			&metric.BounceRate,
			&metric.AvgSessionDuration,
			&metric.ConversionRate,
			&metric.AvgLoadTime,
			&metric.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, &metric)
	}
	return metrics, rows.Err()
}

// UpdateDailyMetrics updates existing daily metrics
func (r *metricsRepository) UpdateDailyMetrics(ctx context.Context, metrics *models.DailyMetrics) error {
	query := `
		UPDATE daily_metrics SET
			total_sessions = $3,
			total_page_views = $4,
			unique_visitors = $5,
			bounce_rate = $6,
			avg_session_duration = $7,
			conversion_rate = $8,
			avg_load_time = $9
		WHERE workspace_id = $1 AND date = $2`

	_, err := r.db.Exec(ctx, query,
		metrics.WorkspaceID,
		metrics.Date,
		metrics.TotalSessions,
		metrics.TotalPageViews,
		metrics.UniqueVisitors,
		metrics.BounceRate,
		metrics.AvgSessionDuration,
		metrics.ConversionRate,
		metrics.AvgLoadTime,
	)
	return err
}

// GetMetricsByWorkspace calculates and returns metrics for a workspace within a time range
func (r *metricsRepository) GetMetricsByWorkspace(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) (*models.MetricsResponse, error) {
	// Calculate current period metrics
	currentMetrics, err := r.calculatePeriodMetrics(ctx, workspaceID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate current metrics: %w", err)
	}

	// Calculate previous period metrics for comparison
	duration := endTime.Sub(startTime)
	prevStartTime := startTime.Add(-duration)
	prevEndTime := startTime
	
	previousMetrics, err := r.calculatePeriodMetrics(ctx, workspaceID, prevStartTime, prevEndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate previous metrics: %w", err)
	}

	// Build metrics response
	response := &models.MetricsResponse{
		Metrics: make(map[string]models.MetricData),
		KPIs:    []models.KPI{},
		Anomalies: []models.Anomaly{},
	}

	// Conversion rate
	response.Metrics["conversion_rate"] = models.MetricData{
		Current:  currentMetrics.ConversionRate,
		Previous: previousMetrics.ConversionRate,
		Trend:    calculateTrend(currentMetrics.ConversionRate, previousMetrics.ConversionRate),
	}

	// Bounce rate
	response.Metrics["bounce_rate"] = models.MetricData{
		Current:  currentMetrics.BounceRate,
		Previous: previousMetrics.BounceRate,
		Trend:    calculateTrend(previousMetrics.BounceRate, currentMetrics.BounceRate), // Inverted for bounce rate
	}

	// Average session duration
	response.Metrics["avg_session_duration"] = models.MetricData{
		Current:  float64(currentMetrics.AvgSessionDuration),
		Previous: float64(previousMetrics.AvgSessionDuration),
		Trend:    calculateTrend(float64(currentMetrics.AvgSessionDuration), float64(previousMetrics.AvgSessionDuration)),
	}

	// Average load time
	response.Metrics["avg_load_time"] = models.MetricData{
		Current:  float64(currentMetrics.AvgLoadTime),
		Previous: float64(previousMetrics.AvgLoadTime),
		Trend:    calculateTrend(float64(previousMetrics.AvgLoadTime), float64(currentMetrics.AvgLoadTime)), // Inverted for load time
	}

	return response, nil
}

// calculatePeriodMetrics calculates aggregated metrics for a specific time period
func (r *metricsRepository) calculatePeriodMetrics(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) (*periodMetrics, error) {
	query := `
		SELECT 
			COALESCE(SUM(total_sessions), 0) as total_sessions,
			COALESCE(SUM(total_page_views), 0) as total_page_views,
			COALESCE(SUM(unique_visitors), 0) as unique_visitors,
			COALESCE(AVG(bounce_rate), 0) as avg_bounce_rate,
			COALESCE(AVG(avg_session_duration), 0) as avg_session_duration,
			COALESCE(AVG(conversion_rate), 0) as avg_conversion_rate,
			COALESCE(AVG(avg_load_time), 0) as avg_load_time
		FROM daily_metrics 
		WHERE workspace_id = $1 AND date >= $2 AND date <= $3`

	var metrics periodMetrics
	err := r.db.QueryRow(ctx, query, workspaceID, startTime.Format("2006-01-02"), endTime.Format("2006-01-02")).Scan(
		&metrics.TotalSessions,
		&metrics.TotalPageViews,
		&metrics.UniqueVisitors,
		&metrics.BounceRate,
		&metrics.AvgSessionDuration,
		&metrics.ConversionRate,
		&metrics.AvgLoadTime,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &metrics, nil
}

// periodMetrics represents aggregated metrics for a time period
type periodMetrics struct {
	TotalSessions      int     `db:"total_sessions"`
	TotalPageViews     int     `db:"total_page_views"`
	UniqueVisitors     int     `db:"unique_visitors"`
	BounceRate         float64 `db:"avg_bounce_rate"`
	AvgSessionDuration int     `db:"avg_session_duration"`
	ConversionRate     float64 `db:"avg_conversion_rate"`
	AvgLoadTime        int     `db:"avg_load_time"`
}

// calculateTrend determines if a metric is trending up, down, or stable
func calculateTrend(current, previous float64) string {
	if previous == 0 {
		if current > 0 {
			return "up"
		}
		return "stable"
	}

	change := ((current - previous) / previous) * 100
	if change > 5 {
		return "up"
	} else if change < -5 {
		return "down"
	}
	return "stable"
}