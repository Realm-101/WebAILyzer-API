package database

import (
	"context"
	"time"
	"github.com/sirupsen/logrus"
)

// MaintenanceService handles periodic database maintenance tasks
type MaintenanceService struct {
	connection *Connection
	logger     *logrus.Logger
	stopChan   chan struct{}
}

// NewMaintenanceService creates a new maintenance service
func NewMaintenanceService(connection *Connection, logger *logrus.Logger) *MaintenanceService {
	return &MaintenanceService{
		connection: connection,
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

// Start begins the maintenance service with periodic tasks
func (ms *MaintenanceService) Start(ctx context.Context) {
	ms.logger.Info("Starting database maintenance service")
	
	// Run initial maintenance
	go ms.runMaintenance(ctx)
	
	// Schedule periodic maintenance tasks
	go ms.scheduleHourlyTasks(ctx)
	go ms.scheduleDailyTasks(ctx)
}

// Stop stops the maintenance service
func (ms *MaintenanceService) Stop() {
	ms.logger.Info("Stopping database maintenance service")
	close(ms.stopChan)
}

// scheduleHourlyTasks runs tasks every hour
func (ms *MaintenanceService) scheduleHourlyTasks(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ms.runHourlyMaintenance(ctx)
		case <-ms.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// scheduleDailyTasks runs tasks every day
func (ms *MaintenanceService) scheduleDailyTasks(ctx context.Context) {
	// Calculate time until next midnight
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	timeUntilMidnight := nextMidnight.Sub(now)
	
	// Wait until midnight, then run daily tasks
	select {
	case <-time.After(timeUntilMidnight):
		ms.runDailyMaintenance(ctx)
	case <-ms.stopChan:
		return
	case <-ctx.Done():
		return
	}
	
	// Then run daily tasks every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ms.runDailyMaintenance(ctx)
		case <-ms.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// runMaintenance performs initial maintenance tasks
func (ms *MaintenanceService) runMaintenance(ctx context.Context) {
	ms.logger.Info("Running initial database maintenance")
	
	perfService := ms.connection.GetPerformanceService()
	if perfService == nil {
		ms.logger.Error("Performance service not available")
		return
	}
	
	// Refresh workspace statistics
	if err := perfService.RefreshWorkspaceStats(ctx); err != nil {
		ms.logger.WithError(err).Error("Failed to refresh workspace stats during initial maintenance")
	}
	
	// Aggregate metrics for the current hour
	currentHour := time.Now().Truncate(time.Hour)
	if err := perfService.AggregateHourlyMetrics(ctx, currentHour); err != nil {
		ms.logger.WithError(err).Error("Failed to aggregate hourly metrics during initial maintenance")
	}
}

// runHourlyMaintenance performs hourly maintenance tasks
func (ms *MaintenanceService) runHourlyMaintenance(ctx context.Context) {
	ms.logger.Info("Running hourly database maintenance")
	
	perfService := ms.connection.GetPerformanceService()
	if perfService == nil {
		ms.logger.Error("Performance service not available")
		return
	}
	
	// Aggregate metrics for the previous hour
	previousHour := time.Now().Add(-1 * time.Hour).Truncate(time.Hour)
	if err := perfService.AggregateHourlyMetrics(ctx, previousHour); err != nil {
		ms.logger.WithError(err).Error("Failed to aggregate hourly metrics")
	}
	
	// Refresh workspace statistics
	if err := perfService.RefreshWorkspaceStats(ctx); err != nil {
		ms.logger.WithError(err).Error("Failed to refresh workspace stats")
	}
	
	// Log connection pool statistics
	stats := perfService.GetConnectionStats()
	ms.logger.WithFields(logrus.Fields{
		"max_conns":          stats.MaxConns,
		"total_conns":        stats.TotalConns,
		"idle_conns":         stats.IdleConns,
		"acquired_conns":     stats.AcquiredConns,
		"constructing_conns": stats.ConstructingConns,
	}).Info("Connection pool statistics")
}

// runDailyMaintenance performs daily maintenance tasks
func (ms *MaintenanceService) runDailyMaintenance(ctx context.Context) {
	ms.logger.Info("Running daily database maintenance")
	
	perfService := ms.connection.GetPerformanceService()
	if perfService == nil {
		ms.logger.Error("Performance service not available")
		return
	}
	
	// Optimize all tables
	if err := ms.connection.OptimizeDatabase(ctx); err != nil {
		ms.logger.WithError(err).Error("Failed to optimize database")
	}
	
	// Get and log table statistics
	tableStats, err := perfService.GetTableStats(ctx)
	if err != nil {
		ms.logger.WithError(err).Error("Failed to get table statistics")
	} else {
		for _, stat := range tableStats {
			ms.logger.WithFields(logrus.Fields{
				"table":      stat.TableName,
				"rows":       stat.RowCount,
				"table_size": stat.TableSize,
				"index_size": stat.IndexSize,
				"total_size": stat.TotalSize,
			}).Info("Table statistics")
		}
	}
	
	// Get and log slow queries
	slowQueries, err := perfService.GetSlowQueries(ctx)
	if err != nil {
		ms.logger.WithError(err).Error("Failed to get slow queries")
	} else if len(slowQueries) > 0 {
		ms.logger.WithField("slow_queries_count", len(slowQueries)).Warn("Slow queries detected")
		for i, query := range slowQueries {
			if i >= 5 { // Log only top 5 slow queries
				break
			}
			ms.logger.WithFields(logrus.Fields{
				"query":      query.Query[:min(100, len(query.Query))], // Truncate long queries
				"calls":      query.Calls,
				"total_time": query.TotalTime,
				"mean_time":  query.MeanTime,
				"rows":       query.Rows,
			}).Warn("Slow query detected")
		}
	}
	
	// Get and log index usage
	indexUsage, err := perfService.GetIndexUsage(ctx)
	if err != nil {
		ms.logger.WithError(err).Error("Failed to get index usage")
	} else {
		unusedIndexes := 0
		for _, usage := range indexUsage {
			if scans, ok := usage["scans"].(int64); ok && scans == 0 {
				unusedIndexes++
			}
		}
		if unusedIndexes > 0 {
			ms.logger.WithField("unused_indexes", unusedIndexes).Warn("Unused indexes detected")
		}
	}
	
	// Clean up old data (older than 90 days)
	ms.cleanupOldData(ctx)
}

// cleanupOldData removes old data to keep the database size manageable
func (ms *MaintenanceService) cleanupOldData(ctx context.Context) {
	ms.logger.Info("Starting old data cleanup")
	
	cutoffDate := time.Now().AddDate(0, 0, -90) // 90 days ago
	
	// Clean up old events (keep only last 90 days)
	eventsQuery := `DELETE FROM events WHERE created_at < $1`
	result, err := ms.connection.Pool.Exec(ctx, eventsQuery, cutoffDate)
	if err != nil {
		ms.logger.WithError(err).Error("Failed to clean up old events")
	} else {
		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			ms.logger.WithField("rows_deleted", rowsAffected).Info("Cleaned up old events")
		}
	}
	
	// Clean up old analysis results (keep only last 90 days)
	analysisQuery := `DELETE FROM analysis_results WHERE created_at < $1`
	result, err = ms.connection.Pool.Exec(ctx, analysisQuery, cutoffDate)
	if err != nil {
		ms.logger.WithError(err).Error("Failed to clean up old analysis results")
	} else {
		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			ms.logger.WithField("rows_deleted", rowsAffected).Info("Cleaned up old analysis results")
		}
	}
	
	// Clean up old sessions that have no associated events or analysis results
	sessionsQuery := `
		DELETE FROM sessions 
		WHERE started_at < $1 
		AND id NOT IN (
			SELECT DISTINCT session_id FROM events WHERE session_id IS NOT NULL
			UNION
			SELECT DISTINCT session_id FROM analysis_results WHERE session_id IS NOT NULL
		)
	`
	result, err = ms.connection.Pool.Exec(ctx, sessionsQuery, cutoffDate)
	if err != nil {
		ms.logger.WithError(err).Error("Failed to clean up old sessions")
	} else {
		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			ms.logger.WithField("rows_deleted", rowsAffected).Info("Cleaned up old sessions")
		}
	}
	
	// Clean up old daily metrics (keep only last 2 years)
	metricsQuery := `DELETE FROM daily_metrics WHERE date < $1`
	result, err = ms.connection.Pool.Exec(ctx, metricsQuery, time.Now().AddDate(-2, 0, 0))
	if err != nil {
		ms.logger.WithError(err).Error("Failed to clean up old daily metrics")
	} else {
		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			ms.logger.WithField("rows_deleted", rowsAffected).Info("Cleaned up old daily metrics")
		}
	}
	
	// Clean up old hourly metrics (keep only last 6 months)
	hourlyQuery := `DELETE FROM hourly_metrics WHERE hour_timestamp < $1`
	result, err = ms.connection.Pool.Exec(ctx, hourlyQuery, time.Now().AddDate(0, -6, 0))
	if err != nil {
		ms.logger.WithError(err).Error("Failed to clean up old hourly metrics")
	} else {
		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			ms.logger.WithField("rows_deleted", rowsAffected).Info("Cleaned up old hourly metrics")
		}
	}
	
	ms.logger.Info("Old data cleanup completed")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}