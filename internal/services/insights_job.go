package services

import (
	"context"
	"sync"
	"time"

	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// InsightsJobManager manages background insight generation jobs
type InsightsJobManager struct {
	insightsService InsightsService
	logger          *logrus.Logger
	jobs            map[uuid.UUID]*InsightJob
	mutex           sync.RWMutex
}

// InsightJob represents a background insight generation job
type InsightJob struct {
	ID          uuid.UUID                `json:"id"`
	WorkspaceID uuid.UUID                `json:"workspace_id"`
	Status      InsightJobStatus         `json:"status"`
	Progress    int                      `json:"progress"`
	Results     []*models.Insight        `json:"results,omitempty"`
	Error       string                   `json:"error,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	CompletedAt *time.Time               `json:"completed_at,omitempty"`
}

// InsightJobStatus represents the status of an insight generation job
type InsightJobStatus string

const (
	InsightJobStatusPending    InsightJobStatus = "pending"
	InsightJobStatusRunning    InsightJobStatus = "running"
	InsightJobStatusCompleted  InsightJobStatus = "completed"
	InsightJobStatusFailed     InsightJobStatus = "failed"
)

// NewInsightsJobManager creates a new insights job manager
func NewInsightsJobManager(insightsService InsightsService, logger *logrus.Logger) *InsightsJobManager {
	return &InsightsJobManager{
		insightsService: insightsService,
		logger:          logger,
		jobs:            make(map[uuid.UUID]*InsightJob),
	}
}

// StartInsightGeneration starts a background insight generation job
func (m *InsightsJobManager) StartInsightGeneration(ctx context.Context, workspaceID uuid.UUID) (*InsightJob, error) {
	job := &InsightJob{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Status:      InsightJobStatusPending,
		Progress:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.mutex.Lock()
	m.jobs[job.ID] = job
	m.mutex.Unlock()

	// Start the job in a goroutine
	go m.runInsightGeneration(ctx, job)

	return job, nil
}

// GetJob retrieves a job by ID
func (m *InsightsJobManager) GetJob(jobID uuid.UUID) (*InsightJob, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	job, exists := m.jobs[jobID]
	return job, exists
}

// GetJobsByWorkspace retrieves all jobs for a workspace
func (m *InsightsJobManager) GetJobsByWorkspace(workspaceID uuid.UUID) []*InsightJob {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var jobs []*InsightJob
	for _, job := range m.jobs {
		if job.WorkspaceID == workspaceID {
			jobs = append(jobs, job)
		}
	}
	
	return jobs
}

// runInsightGeneration executes the insight generation job
func (m *InsightsJobManager) runInsightGeneration(ctx context.Context, job *InsightJob) {
	m.updateJobStatus(job.ID, InsightJobStatusRunning, 10, "", nil)

	m.logger.WithFields(logrus.Fields{
		"job_id":       job.ID,
		"workspace_id": job.WorkspaceID,
	}).Info("Starting insight generation job")

	// Simulate progress updates
	progressSteps := []int{25, 50, 75, 90}
	for _, progress := range progressSteps {
		select {
		case <-ctx.Done():
			m.updateJobStatus(job.ID, InsightJobStatusFailed, progress, "Job cancelled", nil)
			return
		default:
			m.updateJobStatus(job.ID, InsightJobStatusRunning, progress, "", nil)
			time.Sleep(100 * time.Millisecond) // Simulate work
		}
	}

	// Generate insights
	insights, err := m.insightsService.GenerateInsights(ctx, job.WorkspaceID)
	if err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"job_id":       job.ID,
			"workspace_id": job.WorkspaceID,
		}).Error("Failed to generate insights")
		
		m.updateJobStatus(job.ID, InsightJobStatusFailed, 100, err.Error(), nil)
		return
	}

	// Job completed successfully
	now := time.Now()
	m.updateJobStatus(job.ID, InsightJobStatusCompleted, 100, "", insights)
	
	m.mutex.Lock()
	if job, exists := m.jobs[job.ID]; exists {
		job.CompletedAt = &now
	}
	m.mutex.Unlock()

	m.logger.WithFields(logrus.Fields{
		"job_id":         job.ID,
		"workspace_id":   job.WorkspaceID,
		"insights_count": len(insights),
	}).Info("Insight generation job completed successfully")
}

// updateJobStatus updates the status of a job
func (m *InsightsJobManager) updateJobStatus(jobID uuid.UUID, status InsightJobStatus, progress int, errorMsg string, results []*models.Insight) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if job, exists := m.jobs[jobID]; exists {
		job.Status = status
		job.Progress = progress
		job.UpdatedAt = time.Now()
		
		if errorMsg != "" {
			job.Error = errorMsg
		}
		
		if results != nil {
			job.Results = results
		}
	}
}

// CleanupCompletedJobs removes completed jobs older than the specified duration
func (m *InsightsJobManager) CleanupCompletedJobs(maxAge time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	
	for jobID, job := range m.jobs {
		if (job.Status == InsightJobStatusCompleted || job.Status == InsightJobStatusFailed) &&
			job.UpdatedAt.Before(cutoff) {
			delete(m.jobs, jobID)
		}
	}
}