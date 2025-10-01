package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projectdiscovery/wappalyzergo/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestDB holds the test database connection and helper methods
type TestDB struct {
	Pool   *pgxpool.Pool
	conn   *database.Connection
	logger *logrus.Logger
}

// setupTestDB creates a test database connection and runs migrations
func setupTestDB(t *testing.T) *TestDB {
	// Skip if no test database URL is provided
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping database tests")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	// Parse the database URL
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err)

	// Configure for testing
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err)

	// Test the connection
	err = pool.Ping(ctx)
	require.NoError(t, err)

	// Create connection wrapper
	conn := &database.Connection{
		Pool: pool,
	}

	// Run migrations
	migrationManager := database.NewMigrationManager(pool, logger)
	err = migrationManager.Migrate(ctx)
	require.NoError(t, err)

	return &TestDB{
		Pool:   pool,
		conn:   conn,
		logger: logger,
	}
}

// cleanup cleans up test data and closes the database connection
func (tdb *TestDB) cleanup(t *testing.T) {
	ctx := context.Background()

	// Clean up test data in reverse dependency order
	tables := []string{
		"events",
		"analysis_results", 
		"sessions",
		"insights",
		"daily_metrics",
		"workspaces",
		"schema_migrations",
	}

	for _, table := range tables {
		_, err := tdb.Pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}

	tdb.Pool.Close()
}

// createTestWorkspaceID creates a test workspace UUID
func createTestWorkspaceID() string {
	return "550e8400-e29b-41d4-a716-446655440000"
}

// createTestSessionID creates a test session UUID  
func createTestSessionID() string {
	return "550e8400-e29b-41d4-a716-446655440001"
}

// createTestAnalysisID creates a test analysis UUID
func createTestAnalysisID() string {
	return "550e8400-e29b-41d4-a716-446655440002"
}