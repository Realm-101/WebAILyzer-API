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

func TestEventRepository_CreateEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	// Create session first (required for foreign key)
	sessionRepo := NewSessionRepository(testDB.Pool)
	eventRepo := NewEventRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	sessionID := uuid.MustParse(createTestSessionID())

	session := &models.Session{
		ID:          sessionID,
		WorkspaceID: workspaceID,
		StartedAt:   time.Now(),
		PageViews:   0,
		EventsCount: 0,
	}

	err := sessionRepo.CreateSession(ctx, session)
	require.NoError(t, err)

	// Create event
	url := "https://example.com/page1"
	event := &models.Event{
		ID:          uuid.New(),
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		EventType:   "pageview",
		URL:         &url,
		Timestamp:   time.Now(),
		Properties: map[string]interface{}{
			"page_title": "Example Page",
			"referrer":   "https://google.com",
		},
		CreatedAt: time.Now(),
	}

	err = eventRepo.CreateEvent(ctx, event)
	assert.NoError(t, err)
}

func TestEventRepository_CreateEvents(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	// Create session first
	sessionRepo := NewSessionRepository(testDB.Pool)
	eventRepo := NewEventRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	sessionID := uuid.MustParse(createTestSessionID())

	session := &models.Session{
		ID:          sessionID,
		WorkspaceID: workspaceID,
		StartedAt:   time.Now(),
		PageViews:   0,
		EventsCount: 0,
	}

	err := sessionRepo.CreateSession(ctx, session)
	require.NoError(t, err)

	// Create multiple events
	events := []*models.Event{}
	for i := 0; i < 3; i++ {
		url := fmt.Sprintf("https://example.com/page%d", i)
		event := &models.Event{
			ID:          uuid.New(),
			SessionID:   sessionID,
			WorkspaceID: workspaceID,
			EventType:   "pageview",
			URL:         &url,
			Timestamp:   time.Now().Add(time.Duration(i) * time.Minute),
			Properties: map[string]interface{}{
				"page_title": fmt.Sprintf("Page %d", i),
			},
			CreatedAt: time.Now(),
		}
		events = append(events, event)
	}

	err = eventRepo.CreateEvents(ctx, events)
	assert.NoError(t, err)
}

func TestEventRepository_GetEventsBySession(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	// Create session first
	sessionRepo := NewSessionRepository(testDB.Pool)
	eventRepo := NewEventRepository(testDB.Pool)
	ctx := context.Background()

	workspaceID := uuid.MustParse(createTestWorkspaceID())
	sessionID := uuid.MustParse(createTestSessionID())

	session := &models.Session{
		ID:          sessionID,
		WorkspaceID: workspaceID,
		StartedAt:   time.Now(),
		PageViews:   0,
		EventsCount: 0,
	}

	err := sessionRepo.CreateSession(ctx, session)
	require.NoError(t, err)

	// Create events
	events := []*models.Event{}
	for i := 0; i < 3; i++ {
		url := fmt.Sprintf("https://example.com/page%d", i)
		event := &models.Event{
			ID:          uuid.New(),
			SessionID:   sessionID,
			WorkspaceID: workspaceID,
			EventType:   "pageview",
			URL:         &url,
			Timestamp:   time.Now().Add(time.Duration(i) * time.Minute),
			Properties: map[string]interface{}{
				"page_title": fmt.Sprintf("Page %d", i),
			},
			CreatedAt: time.Now(),
		}
		events = append(events, event)
	}

	err = eventRepo.CreateEvents(ctx, events)
	require.NoError(t, err)

	// Test GetEventsBySession
	retrievedEvents, err := eventRepo.GetEventsBySession(ctx, sessionID)
	assert.NoError(t, err)
	assert.Len(t, retrievedEvents, 3)

	// Verify ordering (should be ASC by timestamp)
	assert.True(t, retrievedEvents[0].Timestamp.Before(retrievedEvents[1].Timestamp))
	assert.True(t, retrievedEvents[1].Timestamp.Before(retrievedEvents[2].Timestamp))
}