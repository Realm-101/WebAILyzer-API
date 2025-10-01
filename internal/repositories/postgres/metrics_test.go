package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsRepository_CreateDailyMetrics(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewMetricsRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	bounceRate := 35.5
	avgSessionDuration := 180
	conversionRate := 2.8
	avgLoadTime := 1250

	metrics := &models.DailyMetrics{
		ID:                 uuid.New(),
		WorkspaceID:        workspaceID,
		Date:               time.Now().Truncate(24 * time.Hour),
		TotalSessions:      100,
		TotalPageViews:     250,
		UniqueVisitors:     85,
		BounceRate:         &bounceRate,
		AvgSessionDuration: &avgSessionDuration,
		ConversionRate:     &conversionRate,
		AvgLoadTime:        &avgLoadTime,
		CreatedAt:          time.Now(),
	}

	err := repo.CreateDailyMetrics(ctx, metrics)
	assert.NoError(t, err)
}

func TestMetricsRepository_GetDailyMetrics(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewMetricsRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	baseDate := time.Now().Truncate(24 * time.Hour)

	// Create multiple daily metrics
	for i := 0; i < 3; i++ {
		bounceRate := float64(30 + i*5)
		avgSessionDuration := 150 + i*30
		conversionRate := float64(2 + i)
		avgLoadTime := 1200 + i*100

		metrics := &models.DailyMetrics{
			ID:                 uuid.New(),
			WorkspaceID:        workspaceID,
			Date:               baseDate.AddDate(0, 0, i),
			TotalSessions:      100 + i*10,
			TotalPageViews:     250 + i*25,
			UniqueVisitors:     85 + i*5,
			BounceRate:         &bounceRate,
			AvgSessionDuration: &avgSessionDuration,
			ConversionRate:     &conversionRate,
			AvgLoadTime:        &avgLoadTime,
			CreatedAt:          time.Now(),
		}

		err := repo.CreateDailyMetrics(ctx, metrics)
		require.NoError(t, err)
	}

	// Test GetDailyMetrics
	startDate := baseDate
	endDate := baseDate.AddDate(0, 0, 2)
	
	retrievedMetrics, err := repo.GetDailyMetrics(ctx, workspaceID, startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, retrievedMetrics, 3)

	// Verify ordering (should be ASC by date)
	assert.True(t, retrievedMetrics[0].Date.Before(retrievedMetrics[1].Date) || retrievedMetrics[0].Date.Equal(retrievedMetrics[1].Date))
	assert.True(t, retrievedMetrics[1].Date.Before(retrievedMetrics[2].Date) || retrievedMetrics[1].Date.Equal(retrievedMetrics[2].Date))
}

func TestMetricsRepository_UpdateDailyMetrics(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewMetricsRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	date := time.Now().Truncate(24 * time.Hour)
	bounceRate := 35.5
	avgSessionDuration := 180
	conversionRate := 2.8
	avgLoadTime := 1250

	// Create initial metrics
	metrics := &models.DailyMetrics{
		ID:                 uuid.New(),
		WorkspaceID:        workspaceID,
		Date:               date,
		TotalSessions:      100,
		TotalPageViews:     250,
		UniqueVisitors:     85,
		BounceRate:         &bounceRate,
		AvgSessionDuration: &avgSessionDuration,
		ConversionRate:     &conversionRate,
		AvgLoadTime:        &avgLoadTime,
		CreatedAt:          time.Now(),
	}

	err := repo.CreateDailyMetrics(ctx, metrics)
	require.NoError(t, err)

	// Update metrics
	newBounceRate := 40.0
	newConversionRate := 3.2
	metrics.TotalSessions = 150
	metrics.TotalPageViews = 400
	metrics.BounceRate = &newBounceRate
	metrics.ConversionRate = &newConversionRate

	err = repo.UpdateDailyMetrics(ctx, metrics)
	assert.NoError(t, err)

	// Verify update by retrieving the metrics
	retrievedMetrics, err := repo.GetDailyMetrics(ctx, workspaceID, date, date)
	assert.NoError(t, err)
	assert.Len(t, retrievedMetrics, 1)
	
	retrieved := retrievedMetrics[0]
	assert.Equal(t, 150, retrieved.TotalSessions)
	assert.Equal(t, 400, retrieved.TotalPageViews)
	assert.Equal(t, newBounceRate, *retrieved.BounceRate)
	assert.Equal(t, newConversionRate, *retrieved.ConversionRate)
}

func TestMetricsRepository_GetMetricsByWorkspace(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewMetricsRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	baseDate := time.Now().Truncate(24 * time.Hour)

	// Create metrics for current and previous periods
	bounceRate := 35.5
	avgSessionDuration := 180
	conversionRate := 2.8
	avgLoadTime := 1250

	currentMetrics := &models.DailyMetrics{
		ID:                 uuid.New(),
		WorkspaceID:        workspaceID,
		Date:               baseDate,
		TotalSessions:      100,
		TotalPageViews:     250,
		UniqueVisitors:     85,
		BounceRate:         &bounceRate,
		AvgSessionDuration: &avgSessionDuration,
		ConversionRate:     &conversionRate,
		AvgLoadTime:        &avgLoadTime,
		CreatedAt:          time.Now(),
	}

	err := repo.CreateDailyMetrics(ctx, currentMetrics)
	require.NoError(t, err)

	// Test GetMetricsByWorkspace
	startTime := baseDate
	endTime := baseDate.Add(24 * time.Hour)
	
	response, err := repo.GetMetricsByWorkspace(ctx, workspaceID, startTime, endTime)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Metrics)
	
	// Check that metrics are present
	assert.Contains(t, response.Metrics, "conversion_rate")
	assert.Contains(t, response.Metrics, "bounce_rate")
	assert.Contains(t, response.Metrics, "avg_session_duration")
	assert.Contains(t, response.Metrics, "avg_load_time")
}