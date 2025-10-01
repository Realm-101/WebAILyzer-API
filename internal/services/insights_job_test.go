package services

import (
	"context"
	"testing"
	"time"

	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockInsightsServiceForJob is a mock implementation of InsightsService for job tests
type MockInsightsServiceForJob struct {
	mock.Mock
}

func (m *MockInsightsServiceForJob) GenerateInsights(ctx context.Context, workspaceID uuid.UUID) ([]*models.Insight, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]*models.Insight), args.Error(1)
}

func (m *MockInsightsServiceForJob) GetInsights(ctx context.Context, workspaceID uuid.UUID, filters *InsightFilters) ([]*models.Insight, error) {
	args := m.Called(ctx, workspaceID, filters)
	return args.Get(0).([]*models.Insight), args.Error(1)
}

func (m *MockInsightsServiceForJob) UpdateInsightStatus(ctx context.Context, insightID uuid.UUID, status models.InsightStatus) error {
	args := m.Called(ctx, insightID, status)
	return args.Error(0)
}

func TestInsightsJobManager_StartInsightGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	mockService := new(MockInsightsServiceForJob)
	jobManager := NewInsightsJobManager(mockService, logger)

	workspaceID := uuid.New()

	// Mock the service to return some insights
	insights := []*models.Insight{
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			InsightType: models.InsightTypePerformanceBottleneck,
			Priority:    models.PriorityHigh,
			Title:       "Test Insight",
			Status:      models.InsightStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	mockService.On("GenerateInsights", mock.Anything, workspaceID).Return(insights, nil)

	// Start insight generation
	job, err := jobManager.StartInsightGeneration(context.Background(), workspaceID)

	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, workspaceID, job.WorkspaceID)
	// Job might already be running by the time we check, so just verify it's not nil
	assert.NotEqual(t, uuid.Nil, job.ID)

	// Wait for the job to complete with timeout
	timeout := time.After(1 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var updatedJob *InsightJob
	var exists bool

	for {
		select {
		case <-timeout:
			t.Fatal("Job did not complete within timeout")
		case <-ticker.C:
			updatedJob, exists = jobManager.GetJob(job.ID)
			if exists && updatedJob.Status == InsightJobStatusCompleted {
				goto jobCompleted
			}
		}
	}

jobCompleted:
	assert.True(t, exists)
	assert.Equal(t, InsightJobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 100, updatedJob.Progress)
	assert.Len(t, updatedJob.Results, 1)
	assert.NotNil(t, updatedJob.CompletedAt)

	mockService.AssertExpectations(t)
}

func TestInsightsJobManager_GetJob(t *testing.T) {
	logger := logrus.New()
	mockService := new(MockInsightsServiceForJob)
	jobManager := NewInsightsJobManager(mockService, logger)

	workspaceID := uuid.New()
	mockService.On("GenerateInsights", mock.Anything, workspaceID).Return([]*models.Insight{}, nil)

	// Start a job
	job, err := jobManager.StartInsightGeneration(context.Background(), workspaceID)
	assert.NoError(t, err)

	// Test getting existing job
	retrievedJob, exists := jobManager.GetJob(job.ID)
	assert.True(t, exists)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, job.WorkspaceID, retrievedJob.WorkspaceID)

	// Test getting non-existent job
	nonExistentID := uuid.New()
	_, exists = jobManager.GetJob(nonExistentID)
	assert.False(t, exists)
}

func TestInsightsJobManager_GetJobsByWorkspace(t *testing.T) {
	logger := logrus.New()
	mockService := new(MockInsightsServiceForJob)
	jobManager := NewInsightsJobManager(mockService, logger)

	workspaceID1 := uuid.New()
	workspaceID2 := uuid.New()

	mockService.On("GenerateInsights", mock.Anything, workspaceID1).Return([]*models.Insight{}, nil)
	mockService.On("GenerateInsights", mock.Anything, workspaceID2).Return([]*models.Insight{}, nil)

	// Start jobs for different workspaces
	job1, err := jobManager.StartInsightGeneration(context.Background(), workspaceID1)
	assert.NoError(t, err)

	job2, err := jobManager.StartInsightGeneration(context.Background(), workspaceID1)
	assert.NoError(t, err)

	job3, err := jobManager.StartInsightGeneration(context.Background(), workspaceID2)
	assert.NoError(t, err)

	// Get jobs for workspace1
	jobs1 := jobManager.GetJobsByWorkspace(workspaceID1)
	assert.Len(t, jobs1, 2)

	jobIDs := []uuid.UUID{jobs1[0].ID, jobs1[1].ID}
	assert.Contains(t, jobIDs, job1.ID)
	assert.Contains(t, jobIDs, job2.ID)

	// Get jobs for workspace2
	jobs2 := jobManager.GetJobsByWorkspace(workspaceID2)
	assert.Len(t, jobs2, 1)
	assert.Equal(t, job3.ID, jobs2[0].ID)

	// Get jobs for non-existent workspace
	jobs3 := jobManager.GetJobsByWorkspace(uuid.New())
	assert.Len(t, jobs3, 0)
}

func TestInsightsJobManager_CleanupCompletedJobs(t *testing.T) {
	logger := logrus.New()
	mockService := new(MockInsightsServiceForJob)
	jobManager := NewInsightsJobManager(mockService, logger)

	workspaceID := uuid.New()
	mockService.On("GenerateInsights", mock.Anything, workspaceID).Return([]*models.Insight{}, nil)

	// Start a job
	job, err := jobManager.StartInsightGeneration(context.Background(), workspaceID)
	assert.NoError(t, err)

	// Wait for job to complete with timeout
	timeout := time.After(1 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Job did not complete within timeout")
		case <-ticker.C:
			if updatedJob, exists := jobManager.GetJob(job.ID); exists && updatedJob.Status == InsightJobStatusCompleted {
				goto jobCompleted
			}
		}
	}

jobCompleted:
	// Verify job exists
	_, exists := jobManager.GetJob(job.ID)
	assert.True(t, exists)

	// Cleanup jobs older than 1 hour (should not affect our recent job)
	jobManager.CleanupCompletedJobs(1 * time.Hour)
	_, exists = jobManager.GetJob(job.ID)
	assert.True(t, exists)

	// Cleanup jobs older than 0 seconds (should remove our job)
	jobManager.CleanupCompletedJobs(0)
	_, exists = jobManager.GetJob(job.ID)
	assert.False(t, exists)
}

func TestInsightsJobManager_JobFailure(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	mockService := new(MockInsightsServiceForJob)
	jobManager := NewInsightsJobManager(mockService, logger)

	workspaceID := uuid.New()

	// Mock the service to return an error
	mockService.On("GenerateInsights", mock.Anything, workspaceID).Return([]*models.Insight{}, assert.AnError)

	// Start insight generation
	job, err := jobManager.StartInsightGeneration(context.Background(), workspaceID)
	assert.NoError(t, err)

	// Wait for job to fail with timeout
	timeout := time.After(1 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var updatedJob *InsightJob
	var exists bool

	for {
		select {
		case <-timeout:
			t.Fatal("Job did not fail within timeout")
		case <-ticker.C:
			updatedJob, exists = jobManager.GetJob(job.ID)
			if exists && updatedJob.Status == InsightJobStatusFailed {
				goto jobFailed
			}
		}
	}

jobFailed:
	assert.True(t, exists)
	assert.Equal(t, InsightJobStatusFailed, updatedJob.Status)
	assert.Equal(t, 100, updatedJob.Progress)
	assert.NotEmpty(t, updatedJob.Error)

	mockService.AssertExpectations(t)
}

func TestInsightsJobManager_JobCancellation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	mockService := new(MockInsightsServiceForJob)
	jobManager := NewInsightsJobManager(mockService, logger)

	workspaceID := uuid.New()

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Start insight generation
	job, err := jobManager.StartInsightGeneration(ctx, workspaceID)
	assert.NoError(t, err)

	// Cancel the context immediately
	cancel()

	// Wait for job to be cancelled with timeout
	timeout := time.After(1 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var updatedJob *InsightJob
	var exists bool

	for {
		select {
		case <-timeout:
			t.Fatal("Job was not cancelled within timeout")
		case <-ticker.C:
			updatedJob, exists = jobManager.GetJob(job.ID)
			if exists && updatedJob.Status == InsightJobStatusFailed {
				goto jobCancelled
			}
		}
	}

jobCancelled:
	assert.True(t, exists)
	assert.Equal(t, InsightJobStatusFailed, updatedJob.Status)
	assert.Contains(t, updatedJob.Error, "cancelled")
}