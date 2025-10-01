package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalysisRepository_Create(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewAnalysisRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	sessionID := uuid.MustParse(createTestSessionID())

	// First create a session (required for foreign key)
	sessionRepo := NewSessionRepository(testDB.Pool)
	session := &models.Session{
		ID:          sessionID,
		WorkspaceID: workspaceID,
		StartedAt:   time.Now(),
		PageViews:   0,
		EventsCount: 0,
	}
	err := sessionRepo.CreateSession(ctx, session)
	require.NoError(t, err)

	analysis := &models.AnalysisResult{
		ID:          uuid.MustParse(createTestAnalysisID()),
		WorkspaceID: workspaceID,
		SessionID:   &sessionID,
		URL:         "https://example.com",
		Technologies: map[string]interface{}{
			"React": map[string]interface{}{
				"version": "18.0.0",
			},
		},
		PerformanceMetrics: map[string]interface{}{
			"load_time": 1250,
		},
		SEOMetrics: map[string]interface{}{
			"title": "Example Site",
		},
		AccessibilityMetrics: map[string]interface{}{
			"score": 85,
		},
		SecurityMetrics: map[string]interface{}{
			"https": true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, analysis)
	assert.NoError(t, err)
}

func TestAnalysisRepository_GetByID(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewAnalysisRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	analysisID := uuid.MustParse(createTestAnalysisID())

	// Create test analysis
	analysis := &models.AnalysisResult{
		ID:          analysisID,
		WorkspaceID: workspaceID,
		URL:         "https://example.com",
		Technologies: map[string]interface{}{
			"React": map[string]interface{}{
				"version": "18.0.0",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, analysis)
	require.NoError(t, err)

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, analysisID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, analysisID, retrieved.ID)
	assert.Equal(t, workspaceID, retrieved.WorkspaceID)
	assert.Equal(t, "https://example.com", retrieved.URL)
}

func TestAnalysisRepository_GetByWorkspace(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewAnalysisRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())

	// Create multiple test analyses
	for i := 0; i < 3; i++ {
		analysis := &models.AnalysisResult{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			URL:         fmt.Sprintf("https://example%d.com", i),
			Technologies: map[string]interface{}{
				"React": map[string]interface{}{
					"version": "18.0.0",
				},
			},
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
			UpdatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}

		err := repo.Create(ctx, analysis)
		require.NoError(t, err)
	}

	// Test GetByWorkspace
	results, err := repo.GetByWorkspace(ctx, workspaceID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Test pagination
	results, err = repo.GetByWorkspace(ctx, workspaceID, 2, 0)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	results, err = repo.GetByWorkspace(ctx, workspaceID, 2, 2)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestAnalysisRepository_Update(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewAnalysisRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	analysisID := uuid.MustParse(createTestAnalysisID())

	// Create test analysis
	analysis := &models.AnalysisResult{
		ID:          analysisID,
		WorkspaceID: workspaceID,
		URL:         "https://example.com",
		Technologies: map[string]interface{}{
			"React": map[string]interface{}{
				"version": "18.0.0",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, analysis)
	require.NoError(t, err)

	// Update the analysis
	analysis.Technologies = map[string]interface{}{
		"Vue": map[string]interface{}{
			"version": "3.0.0",
		},
	}
	analysis.UpdatedAt = time.Now()

	err = repo.Update(ctx, analysis)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, analysisID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	
	vue, exists := retrieved.Technologies["Vue"]
	assert.True(t, exists)
	assert.Equal(t, "3.0.0", vue.(map[string]interface{})["version"])
}

func TestAnalysisRepository_Delete(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewAnalysisRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	analysisID := uuid.MustParse(createTestAnalysisID())

	// Create test analysis
	analysis := &models.AnalysisResult{
		ID:          analysisID,
		WorkspaceID: workspaceID,
		URL:         "https://example.com",
		Technologies: map[string]interface{}{
			"React": map[string]interface{}{
				"version": "18.0.0",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, analysis)
	require.NoError(t, err)

	// Delete the analysis
	err = repo.Delete(ctx, analysisID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := repo.GetByID(ctx, analysisID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}