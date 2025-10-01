package database

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migration represents a database migration
type Migration struct {
	Version     string
	Name        string
	SQL         string
	AppliedAt   *time.Time
}

// MigrationManager handles database migrations
type MigrationManager struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(pool *pgxpool.Pool, logger *logrus.Logger) *MigrationManager {
	return &MigrationManager{
		pool:   pool,
		logger: logger,
	}
}

// Migrate runs all pending migrations
func (m *MigrationManager) Migrate(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all migration files
	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if _, applied := appliedMigrations[migration.Version]; !applied {
			if err := m.applyMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
			}
			m.logger.Infof("Applied migration: %s - %s", migration.Version, migration.Name)
		}
	}

	m.logger.Info("All migrations applied successfully")
	return nil
}

// createMigrationsTable creates the migrations tracking table
func (m *MigrationManager) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	_, err := m.pool.Exec(ctx, query)
	return err
}

// loadMigrations loads all migration files from the embedded filesystem
func (m *MigrationManager) loadMigrations() ([]Migration, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := migrationFiles.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		// Extract version and name from filename (e.g., "001_initial_schema.sql")
		parts := strings.SplitN(strings.TrimSuffix(entry.Name(), ".sql"), "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration filename format: %s", entry.Name())
		}

		migration := Migration{
			Version: parts[0],
			Name:    strings.ReplaceAll(parts[1], "_", " "),
			SQL:     string(content),
		}

		migrations = append(migrations, migration)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations returns a map of applied migration versions
func (m *MigrationManager) getAppliedMigrations(ctx context.Context) (map[string]*time.Time, error) {
	query := "SELECT version, applied_at FROM schema_migrations"
	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]*time.Time)
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = &appliedAt
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration within a transaction
func (m *MigrationManager) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Execute the migration SQL
	if _, err := tx.Exec(ctx, migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration as applied
	insertQuery := "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)"
	if _, err := tx.Exec(ctx, insertQuery, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit(ctx)
}

// GetMigrationStatus returns the status of all migrations
func (m *MigrationManager) GetMigrationStatus(ctx context.Context) ([]Migration, error) {
	migrations, err := m.loadMigrations()
	if err != nil {
		return nil, err
	}

	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	// Set applied_at for applied migrations
	for i := range migrations {
		if appliedAt, applied := appliedMigrations[migrations[i].Version]; applied {
			migrations[i].AppliedAt = appliedAt
		}
	}

	return migrations, nil
}

// Rollback rolls back the last applied migration (use with caution)
func (m *MigrationManager) Rollback(ctx context.Context) error {
	// Get the last applied migration
	query := "SELECT version, name FROM schema_migrations ORDER BY applied_at DESC LIMIT 1"
	var version, name string
	err := m.pool.QueryRow(ctx, query).Scan(&version, &name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("no migrations to rollback")
		}
		return err
	}

	// Remove from migrations table
	deleteQuery := "DELETE FROM schema_migrations WHERE version = $1"
	if _, err := m.pool.Exec(ctx, deleteQuery, version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	m.logger.Warnf("Rolled back migration: %s - %s (manual cleanup may be required)", version, name)
	return nil
}