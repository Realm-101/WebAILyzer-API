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

func TestSessionRepository_CreateSession(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewSessionRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	sessionID := uuid.MustParse(createTestSessionID())
	userID := "test-user-123"
	deviceType := "desktop"
	browser := "Chrome"
	country := "US"
	referrer := "https://google.com"

	session := &models.Session{
		ID:          sessionID,
		WorkspaceID: workspaceID,
		UserID:      &userID,
		StartedAt:   time.Now(),
		PageViews:   1,
		EventsCount: 0,
		DeviceType:  &deviceType,
		Browser:     &browser,
		Country:     &country,
		Referrer:    &referrer,
	}

	err := repo.CreateSession(ctx, session)
	assert.NoError(t, err)
}

func TestSessionRepository_GetSessionByID(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewSessionRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	sessionID := uuid.MustParse(createTestSessionID())
	userID := "test-user-123"

	// Create test session
	session := &models.Session{
		ID:          sessionID,
		WorkspaceID: workspaceID,
		UserID:      &userID,
		StartedAt:   time.Now(),
		PageViews:   1,
		EventsCount: 0,
	}

	err := repo.CreateSession(ctx, session)
	require.NoError(t, err)

	// Test GetSessionByID
	retrieved, err := repo.GetSessionByID(ctx, sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, sessionID, retrieved.ID)
	assert.Equal(t, workspaceID, retrieved.WorkspaceID)
	assert.Equal(t, userID, *retrieved.UserID)
	assert.Equal(t, 1, retrieved.PageViews)
}

func TestSessionRepository_UpdateSession(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewSessionRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	sessionID := uuid.MustParse(createTestSessionID())

	// Create test session
	session := &models.Session{
		ID:          sessionID,
		WorkspaceID: workspaceID,
		StartedAt:   time.Now(),
		PageViews:   1,
		EventsCount: 0,
	}

	err := repo.CreateSession(ctx, session)
	require.NoError(t, err)

	// Update the session
	endTime := time.Now().Add(30 * time.Minute)
	duration := 1800 // 30 minutes in seconds
	session.EndedAt = &endTime
	session.DurationSeconds = &duration
	session.PageViews = 5
	session.EventsCount = 10

	err = repo.UpdateSession(ctx, session)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetSessionByID(ctx, sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.NotNil(t, retrieved.EndedAt)
	assert.NotNil(t, retrieved.DurationSeconds)
	assert.Equal(t, 5, retrieved.PageViews)
	assert.Equal(t, 10, retrieved.EventsCount)
	assert.Equal(t, duration, *retrieved.DurationSeconds)
}

func TestSessionRepository_GetSessionsByWorkspace(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	repo := NewSessionRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())

	// Create multiple test sessions
	for i := 0; i < 3; i++ {
		session := &models.Session{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			StartedAt:   time.Now().Add(time.Duration(i) * time.Minute),
			PageViews:   i + 1,
			EventsCount: 0,
		}

		err := repo.CreateSession(ctx, session)
		require.NoError(t, err)
	}

	// Test GetSessionsByWorkspace
	sessions, err := repo.GetSessionsByWorkspace(ctx, workspaceID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Verify ordering (should be DESC by started_at)
	assert.True(t, sessions[0].StartedAt.After(sessions[1].StartedAt))
	assert.True(t, sessions[1].StartedAt.After(sessions[2].StartedAt))

	// Test pagination
	sessions, err = repo.GetSessionsByWorkspace(ctx, workspaceID, 2, 0)
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)

	sessions, err = repo.GetSessionsByWorkspace(ctx, workspaceID, 2, 2)
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
}