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

func TestInsightRepository_Create(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewInsightRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	description := "Your mobile pages are loading slowly, affecting user experience."
	impactScore := 85
	effortScore := 30

	insight := &models.Insight{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		InsightType: models.InsightTypePerformanceBottleneck,
		Priority:    models.PriorityHigh,
		Title:       "Mobile page load time exceeds 3 seconds",
		Description: &description,
		ImpactScore: &impactScore,
		EffortScore: &effortScore,
		Recommendations: map[string]interface{}{
			"optimize_images": "Compress and resize images for mobile devices",
			"minify_css":      "Minify CSS files to reduce load time",
		},
		DataSource: map[string]interface{}{
			"avg_load_time": 3200,
			"device_type":   "mobile",
		},
		Status:    models.InsightStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, insight)
	assert.NoError(t, err)
}

func TestInsightRepository_GetByID(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewInsightRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	insightID := uuid.New()

	// Create test insight
	insight := &models.Insight{
		ID:          insightID,
		WorkspaceID: workspaceID,
		InsightType: models.InsightTypePerformanceBottleneck,
		Priority:    models.PriorityHigh,
		Title:       "Test Insight",
		Recommendations: map[string]interface{}{
			"test": "recommendation",
		},
		DataSource: map[string]interface{}{
			"test": "data",
		},
		Status:    models.InsightStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, insight)
	require.NoError(t, err)

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, insightID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, insightID, retrieved.ID)
	assert.Equal(t, workspaceID, retrieved.WorkspaceID)
	assert.Equal(t, "Test Insight", retrieved.Title)
	assert.Equal(t, models.InsightStatusPending, retrieved.Status)
}

func TestInsightRepository_GetByWorkspace(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewInsightRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())

	// Create multiple test insights with different statuses
	insights := []struct {
		title  string
		status models.InsightStatus
	}{
		{"Pending Insight 1", models.InsightStatusPending},
		{"Pending Insight 2", models.InsightStatusPending},
		{"Applied Insight", models.InsightStatusApplied},
	}

	for _, insightData := range insights {
		insight := &models.Insight{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			InsightType: models.InsightTypePerformanceBottleneck,
			Priority:    models.PriorityMedium,
			Title:       insightData.title,
			Recommendations: map[string]interface{}{
				"test": "recommendation",
			},
			DataSource: map[string]interface{}{
				"test": "data",
			},
			Status:    insightData.status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.Create(ctx, insight)
		require.NoError(t, err)
	}

	// Test GetByWorkspace without status filter
	allInsights, err := repo.GetByWorkspace(ctx, workspaceID, nil, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, allInsights, 3)

	// Test GetByWorkspace with status filter
	pendingStatus := models.InsightStatusPending
	pendingInsights, err := repo.GetByWorkspace(ctx, workspaceID, &pendingStatus, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, pendingInsights, 2)

	appliedStatus := models.InsightStatusApplied
	appliedInsights, err := repo.GetByWorkspace(ctx, workspaceID, &appliedStatus, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, appliedInsights, 1)
}

func TestInsightRepository_UpdateStatus(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewInsightRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	insightID := uuid.New()

	// Create test insight
	insight := &models.Insight{
		ID:          insightID,
		WorkspaceID: workspaceID,
		InsightType: models.InsightTypePerformanceBottleneck,
		Priority:    models.PriorityHigh,
		Title:       "Test Insight",
		Recommendations: map[string]interface{}{
			"test": "recommendation",
		},
		DataSource: map[string]interface{}{
			"test": "data",
		},
		Status:    models.InsightStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, insight)
	require.NoError(t, err)

	// Update status
	err = repo.UpdateStatus(ctx, insightID, models.InsightStatusApplied)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, insightID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, models.InsightStatusApplied, retrieved.Status)
}