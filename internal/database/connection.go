package database

import (
	"context"
	"fmt"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projectdiscovery/wappalyzergo/internal/config"
	"github.com/sirupsen/logrus"
)

// Connection manages database connections
type Connection struct {
	Pool              *pgxpool.Pool
	config            *config.DatabaseConfig
	logger            *logrus.Logger
	migrationManager  *MigrationManager
	performanceService *PerformanceService
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.DatabaseConfig, logger *logrus.Logger) (*Connection, error) {
	conn := &Connection{
		config: cfg,
		logger: logger,
	}

	if err := conn.Connect(); err != nil {
		return nil, err
	}

	return conn, nil
}

// Connect establishes a connection to the database
func (c *Connection) Connect() error {
	poolConfig, err := pgxpool.ParseConfig(c.config.GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(c.config.MaxOpenConns)
	poolConfig.MinConns = int32(c.config.MaxIdleConns)
	poolConfig.MaxConnLifetime = c.config.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.Pool = pool
	c.migrationManager = NewMigrationManager(pool, c.logger)
	c.performanceService = NewPerformanceService(pool, c.logger)
	c.logger.Info("Successfully connected to database")
	return nil
}

// Close closes the database connection
func (c *Connection) Close() {
	if c.Pool != nil {
		c.Pool.Close()
		c.logger.Info("Database connection closed")
	}
}

// HealthCheck performs a health check on the database connection
func (c *Connection) HealthCheck(ctx context.Context) error {
	if c.Pool == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Pool.Ping(ctx)
}

// Migrate runs all pending database migrations
func (c *Connection) Migrate(ctx context.Context) error {
	if c.migrationManager == nil {
		return fmt.Errorf("migration manager is not initialized")
	}
	return c.migrationManager.Migrate(ctx)
}

// GetMigrationStatus returns the status of all migrations
func (c *Connection) GetMigrationStatus(ctx context.Context) ([]Migration, error) {
	if c.migrationManager == nil {
		return nil, fmt.Errorf("migration manager is not initialized")
	}
	return c.migrationManager.GetMigrationStatus(ctx)
}

// RollbackMigration rolls back the last applied migration
func (c *Connection) RollbackMigration(ctx context.Context) error {
	if c.migrationManager == nil {
		return fmt.Errorf("migration manager is not initialized")
	}
	return c.migrationManager.Rollback(ctx)
}

// GetPerformanceService returns the performance service
func (c *Connection) GetPerformanceService() *PerformanceService {
	return c.performanceService
}

// OptimizeDatabase performs database optimization tasks
func (c *Connection) OptimizeDatabase(ctx context.Context) error {
	if c.performanceService == nil {
		return fmt.Errorf("performance service is not initialized")
	}

	// Optimize common tables
	tables := []string{"analysis_results", "sessions", "events", "insights", "daily_metrics"}
	for _, table := range tables {
		if err := c.performanceService.OptimizeTable(ctx, table); err != nil {
			c.logger.WithError(err).WithField("table", table).Warn("Failed to optimize table")
		}
	}

	// Refresh workspace stats
	if err := c.performanceService.RefreshWorkspaceStats(ctx); err != nil {
		c.logger.WithError(err).Warn("Failed to refresh workspace stats")
	}

	c.logger.Info("Database optimization completed")
	return nil
}