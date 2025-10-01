package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/repositories"
)

// MockAnalysisRepository is a mock implementation of AnalysisRepository
type MockAnalysisRepository struct {
	mock.Mock
}

func (m *MockAnalysisRepository) Create(ctx context.Context, analysis *models.AnalysisResult) error {
	args := m.Called(ctx, analysis)
	return args.Error(0)
}

func (m *MockAnalysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AnalysisResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalysisResult), args.Error(1)
}

func (m *MockAnalysisRepository) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func (m *MockAnalysisRepository) GetBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func (m *MockAnalysisRepository) GetByFilters(ctx context.Context, filters *repositories.AnalysisFilters) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func (m *MockAnalysisRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, filters *repositories.AnalysisFilters) ([]*models.AnalysisResult, error) {
	args := m.Called(ctx, workspaceID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AnalysisResult), args.Error(1)
}

func (m *MockAnalysisRepository) Update(ctx context.Context, analysis *models.AnalysisResult) error {
	args := m.Called(ctx, analysis)
	return args.Error(0)
}

func (m *MockAnalysisRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAnalysisService_AnalyzeURL_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockAnalysisRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service, err := NewAnalysisService(mockRepo, logger)
	require.NoError(t, err)

	workspaceID := uuid.New()
	sessionID := uuid.New()

	req := &models.AnalysisRequest{
		URL:         "https://httpbin.org/html",
		WorkspaceID: workspaceID,
		SessionID:   &sessionID,
		Options: models.AnalysisOptions{
			IncludePerformance:   true,
			IncludeSEO:          true,
			IncludeAccessibility: true,
			IncludeSecurity:     true,
		},
	}

	// Mock repository call
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.AnalysisResult")).Return(nil)

	// Execute
	result, err := service.AnalyzeURL(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.URL, result.URL)
	assert.Equal(t, req.WorkspaceID, result.WorkspaceID)
	assert.Equal(t, req.SessionID, result.SessionID)
	assert.NotEmpty(t, result.Technologies)

	mockRepo.AssertExpectations(t)
}

func TestAnalysisService_BatchAnalyze_ConcurrentProcessing(t *testing.T) {
	// Setup mock repository
	mockRepo := new(MockAnalysisRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service, err := NewAnalysisService(mockRepo, logger)
	require.NoError(t, err)

	workspaceID := uuid.New()
	urls := []string{
		"https://httpbin.org/json",
		"https://httpbin.org/html",
		"https://httpbin.org/xml",
	}

	// Mock repository calls for successful analysis results
	for range urls {
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.AnalysisResult")).Return(nil)
	}

	req := &BatchAnalysisRequest{
		URLs:        urls,
		WorkspaceID: workspaceID,
		Options: models.AnalysisOptions{
			IncludePerformance:   true,
			IncludeSEO:          true,
			IncludeAccessibility: true,
			IncludeSecurity:     true,
		},
	}

	// Execute batch analysis
	result, err := service.BatchAnalyze(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.BatchID)
	assert.Contains(t, []string{"completed", "partial"}, result.Status) // Allow partial due to network issues
	assert.Equal(t, len(urls), result.Progress.Completed)
	assert.Equal(t, len(urls), result.Progress.Total)

	// Verify all URLs were processed (either successfully or failed)
	totalProcessed := len(result.Results) + len(result.FailedURLs)
	assert.Equal(t, len(urls), totalProcessed)

	// Verify repository was called for successful results
	mockRepo.AssertExpectations(t)
}

func TestAnalysisService_BatchAnalyze_PartialFailure(t *testing.T) {
	// Setup mock repository
	mockRepo := new(MockAnalysisRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service, err := NewAnalysisService(mockRepo, logger)
	require.NoError(t, err)

	workspaceID := uuid.New()
	urls := []string{
		"https://httpbin.org/json",
		"https://invalid-domain-that-does-not-exist.com",
		"https://httpbin.org/html",
	}

	// Mock repository calls - only for successful analyses
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.AnalysisResult")).Return(nil).Maybe()

	req := &BatchAnalysisRequest{
		URLs:        urls,
		WorkspaceID: workspaceID,
		Options: models.AnalysisOptions{
			IncludePerformance:   true,
			IncludeSEO:          true,
			IncludeAccessibility: true,
			IncludeSecurity:     true,
		},
	}

	// Execute batch analysis
	result, err := service.BatchAnalyze(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, []string{"completed", "partial", "failed"}, result.Status)
	assert.Equal(t, 3, result.Progress.Completed)
	assert.Equal(t, 3, result.Progress.Total)

	// Verify all URLs were processed (either successfully or failed)
	totalProcessed := len(result.Results) + len(result.FailedURLs)
	assert.Equal(t, len(urls), totalProcessed)

	// Should have at least one failure due to invalid domain
	assert.Greater(t, len(result.FailedURLs), 0)

	// Verify repository was called for successful results
	mockRepo.AssertExpectations(t)
}

func TestAnalysisService_BatchAnalyze_ContextCancellation(t *testing.T) {
	// Setup mock repository
	mockRepo := new(MockAnalysisRepository)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service, err := NewAnalysisService(mockRepo, logger)
	require.NoError(t, err)

	workspaceID := uuid.New()
	urls := []string{
		"https://httpbin.org/json",
		"https://httpbin.org/html",
		"https://httpbin.org/xml",
	}

	req := &BatchAnalysisRequest{
		URLs:        urls,
		WorkspaceID: workspaceID,
		Options:     models.AnalysisOptions{},
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Execute batch analysis with cancelled context
	result, err := service.BatchAnalyze(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.NotNil(t, result)
	assert.Equal(t, "cancelled", result.Status)
}