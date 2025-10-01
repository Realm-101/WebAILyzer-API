package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/projectdiscovery/wappalyzergo/internal/models"
)

func TestWorkspaceRepository_Create(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewWorkspaceRepository(testDB.Pool)

	workspace := &models.Workspace{
		ID:        uuid.New(),
		Name:      "Test Workspace",
		APIKey:    "test-api-key-" + uuid.New().String(),
		IsActive:  true,
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	// Verify the workspace was created
	retrieved, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, workspace.ID, retrieved.ID)
	assert.Equal(t, workspace.Name, retrieved.Name)
	assert.Equal(t, workspace.APIKey, retrieved.APIKey)
	assert.Equal(t, workspace.IsActive, retrieved.IsActive)
	assert.Equal(t, workspace.RateLimit, retrieved.RateLimit)
}

func TestWorkspaceRepository_GetByID(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewWorkspaceRepository(testDB.Pool)

	// Test getting non-existent workspace
	nonExistentID := uuid.New()
	workspace, err := repo.GetByID(context.Background(), nonExistentID)
	require.NoError(t, err)
	assert.Nil(t, workspace)

	// Create a workspace
	testWorkspace := &models.Workspace{
		ID:        uuid.New(),
		Name:      "Test Workspace",
		APIKey:    "test-api-key-" + uuid.New().String(),
		IsActive:  true,
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(context.Background(), testWorkspace)
	require.NoError(t, err)

	// Test getting existing workspace
	retrieved, err := repo.GetByID(context.Background(), testWorkspace.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, testWorkspace.ID, retrieved.ID)
	assert.Equal(t, testWorkspace.Name, retrieved.Name)
	assert.Equal(t, testWorkspace.APIKey, retrieved.APIKey)
}

func TestWorkspaceRepository_GetByAPIKey(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewWorkspaceRepository(testDB.Pool)

	// Test getting non-existent API key
	workspace, err := repo.GetByAPIKey(context.Background(), "non-existent-key")
	require.NoError(t, err)
	assert.Nil(t, workspace)

	// Create a workspace
	testWorkspace := &models.Workspace{
		ID:        uuid.New(),
		Name:      "Test Workspace",
		APIKey:    "test-api-key-" + uuid.New().String(),
		IsActive:  true,
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(context.Background(), testWorkspace)
	require.NoError(t, err)

	// Test getting existing API key
	retrieved, err := repo.GetByAPIKey(context.Background(), testWorkspace.APIKey)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, testWorkspace.ID, retrieved.ID)
	assert.Equal(t, testWorkspace.Name, retrieved.Name)
	assert.Equal(t, testWorkspace.APIKey, retrieved.APIKey)
}

func TestWorkspaceRepository_Update(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewWorkspaceRepository(testDB.Pool)

	// Create a workspace
	workspace := &models.Workspace{
		ID:        uuid.New(),
		Name:      "Test Workspace",
		APIKey:    "test-api-key-" + uuid.New().String(),
		IsActive:  true,
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	// Update the workspace
	workspace.Name = "Updated Workspace"
	workspace.IsActive = false
	workspace.RateLimit = 2000
	workspace.UpdatedAt = time.Now()

	err = repo.Update(context.Background(), workspace)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, "Updated Workspace", retrieved.Name)
	assert.False(t, retrieved.IsActive)
	assert.Equal(t, 2000, retrieved.RateLimit)
}

func TestWorkspaceRepository_Delete(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewWorkspaceRepository(testDB.Pool)

	// Create a workspace
	workspace := &models.Workspace{
		ID:        uuid.New(),
		Name:      "Test Workspace",
		APIKey:    "test-api-key-" + uuid.New().String(),
		IsActive:  true,
		RateLimit: 1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	// Delete the workspace
	err = repo.Delete(context.Background(), workspace.ID)
	require.NoError(t, err)

	// Verify the workspace is deleted
	retrieved, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	assert.Nil(t, retrieved)

	// Test deleting non-existent workspace
	err = repo.Delete(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace not found")
}

func TestWorkspaceRepository_List(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewWorkspaceRepository(testDB.Pool)

	// Create multiple workspaces
	workspaces := []*models.Workspace{
		{
			ID:        uuid.New(),
			Name:      "Workspace 1",
			APIKey:    "api-key-1-" + uuid.New().String(),
			IsActive:  true,
			RateLimit: 1000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Workspace 2",
			APIKey:    "api-key-2-" + uuid.New().String(),
			IsActive:  false,
			RateLimit: 2000,
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
		},
		{
			ID:        uuid.New(),
			Name:      "Workspace 3",
			APIKey:    "api-key-3-" + uuid.New().String(),
			IsActive:  true,
			RateLimit: 3000,
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	for _, workspace := range workspaces {
		err := repo.Create(context.Background(), workspace)
		require.NoError(t, err)
	}

	// Test listing with limit
	retrieved, err := repo.List(context.Background(), 2, 0)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// Verify ordering (should be by created_at DESC)
	assert.Equal(t, "Workspace 1", retrieved[0].Name) // Most recent
	assert.Equal(t, "Workspace 2", retrieved[1].Name) // Second most recent

	// Test listing with offset
	retrieved, err = repo.List(context.Background(), 2, 1)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)
	assert.Equal(t, "Workspace 2", retrieved[0].Name)
	assert.Equal(t, "Workspace 3", retrieved[1].Name)

	// Test listing all
	retrieved, err = repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(retrieved), 3) // At least our 3 workspaces
}