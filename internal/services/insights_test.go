package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockInsightRepository is a mock implementation of InsightRepository
type MockInsightRepository struct {
	mock.Mock
}

func (m *MockInsightRepository) Create(ctx context.Context, insight *models.Insight) error {
	args := m.Called(ctx, insight)
	return args.Error(0)
}

func (m *MockInsightRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Insight, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Insight), args.Error(1)
}

func (m *MockInsightRepository) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, status *models.InsightStatus, limit, offset int) ([]*models.Insight, error) {
	args := m.Called(ctx, workspaceID, status, limit, offset)
	return args.Get(0).([]*models.Insight), args.Error(1)
}

func (m *MockInsightRepository) GetByFilters(ctx context.Context, workspaceID uuid.UUID, status *models.InsightStatus, insightType *models.InsightType, priority *models.Priority, limit, offset int) ([]*models.Insight, error) {
	args := m.Called(ctx, workspaceID, status, insightType, priority, limit, offset)
	return args.Get(0).([]*models.Insight), args.Error(1)
}

func (m *MockInsightRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.InsightStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockInsightRepository) Update(ctx context.Context, insight *models.Insight) error {
	args := m.Called(ctx, insight)
	return args.Error(0)
}

func (m *MockInsightRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestInsightsService_GetInsights(t *testing.T) {
	tests := []struct {
		name          string
		workspaceID   uuid.UUID
		filters       *InsightFilters
		mockSetup     func(*MockInsightRepository)
		expectedError string
		validateFunc  func([]*models.Insight) bool
	}{
		{
			name:        "successful retrieval with default filters",
			workspaceID: uuid.New(),
			filters:     nil,
			mockSetup: func(m *MockInsightRepository) {
				insights := []*models.Insight{
					{
						ID:          uuid.New(),
						WorkspaceID: uuid.New(),
						InsightType: models.InsightTypePerformanceBottleneck,
						Priority:    models.PriorityHigh,
						Title:       "Test Insight",
						Status:      models.InsightStatusPending,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				m.On("GetByFilters", mock.Anything, mock.Anything, (*models.InsightStatus)(nil), (*models.InsightType)(nil), (*models.Priority)(nil), 50, 0).Return(insights, nil)
			},
			validateFunc: func(insights []*models.Insight) bool {
				return len(insights) == 1 && insights[0].Title == "Test Insight"
			},
		},
		{
			name:        "successful retrieval with custom filters",
			workspaceID: uuid.New(),
			filters: &InsightFilters{
				Status:   &[]models.InsightStatus{models.InsightStatusPending}[0],
				Type:     &[]models.InsightType{models.InsightTypePerformanceBottleneck}[0],
				Priority: &[]models.Priority{models.PriorityHigh}[0],
				Limit:    10,
				Offset:   5,
			},
			mockSetup: func(m *MockInsightRepository) {
				insights := []*models.Insight{}
				m.On("GetByFilters", mock.Anything, mock.Anything, mock.MatchedBy(func(status *models.InsightStatus) bool {
					return status != nil && *status == models.InsightStatusPending
				}), mock.MatchedBy(func(insightType *models.InsightType) bool {
					return insightType != nil && *insightType == models.InsightTypePerformanceBottleneck
				}), mock.MatchedBy(func(priority *models.Priority) bool {
					return priority != nil && *priority == models.PriorityHigh
				}), 10, 5).Return(insights, nil)
			},
			validateFunc: func(insights []*models.Insight) bool {
				return len(insights) == 0
			},
		},
		{
			name:        "repository error",
			workspaceID: uuid.New(),
			filters:     nil,
			mockSetup: func(m *MockInsightRepository) {
				m.On("GetByFilters", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*models.Insight{}, fmt.Errorf("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "zero limit gets set to default",
			workspaceID: uuid.New(),
			filters: &InsightFilters{
				Limit:  0,
				Offset: 0,
			},
			mockSetup: func(m *MockInsightRepository) {
				insights := []*models.Insight{}
				m.On("GetByFilters", mock.Anything, mock.Anything, (*models.InsightStatus)(nil), (*models.InsightType)(nil), (*models.Priority)(nil), 50, 0).Return(insights, nil)
			},
			validateFunc: func(insights []*models.Insight) bool {
				return len(insights) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInsightRepository)
			tt.mockSetup(mockRepo)

			service := NewInsightsService(mockRepo)

			insights, err := service.GetInsights(context.Background(), tt.workspaceID, tt.filters)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.validateFunc(insights))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInsightsService_UpdateInsightStatus(t *testing.T) {
	tests := []struct {
		name          string
		insightID     uuid.UUID
		status        models.InsightStatus
		mockSetup     func(*MockInsightRepository)
		expectedError string
	}{
		{
			name:      "successful status update to applied",
			insightID: uuid.New(),
			status:    models.InsightStatusApplied,
			mockSetup: func(m *MockInsightRepository) {
				m.On("UpdateStatus", mock.Anything, mock.Anything, models.InsightStatusApplied).Return(nil)
			},
		},
		{
			name:      "successful status update to dismissed",
			insightID: uuid.New(),
			status:    models.InsightStatusDismissed,
			mockSetup: func(m *MockInsightRepository) {
				m.On("UpdateStatus", mock.Anything, mock.Anything, models.InsightStatusDismissed).Return(nil)
			},
		},
		{
			name:      "successful status update to pending",
			insightID: uuid.New(),
			status:    models.InsightStatusPending,
			mockSetup: func(m *MockInsightRepository) {
				m.On("UpdateStatus", mock.Anything, mock.Anything, models.InsightStatusPending).Return(nil)
			},
		},
		{
			name:          "invalid status",
			insightID:     uuid.New(),
			status:        models.InsightStatus("invalid"),
			mockSetup:     func(m *MockInsightRepository) {},
			expectedError: "invalid insight status: invalid",
		},
		{
			name:      "repository error",
			insightID: uuid.New(),
			status:    models.InsightStatusApplied,
			mockSetup: func(m *MockInsightRepository) {
				m.On("UpdateStatus", mock.Anything, mock.Anything, models.InsightStatusApplied).Return(fmt.Errorf("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInsightRepository)
			tt.mockSetup(mockRepo)

			service := NewInsightsService(mockRepo)

			err := service.UpdateInsightStatus(context.Background(), tt.insightID, tt.status)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInsightsService_GenerateInsights(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID uuid.UUID
		mockSetup   func(*MockInsightRepository)
	}{
		{
			name:        "successful generation",
			workspaceID: uuid.New(),
			mockSetup:   func(m *MockInsightRepository) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInsightRepository)
			tt.mockSetup(mockRepo)

			service := NewInsightsService(mockRepo)

			insights, err := service.GenerateInsights(context.Background(), tt.workspaceID)

			assert.NoError(t, err)
			assert.NotNil(t, insights)
			// For now, the service returns an empty slice
			assert.Len(t, insights, 0)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNewInsightsService(t *testing.T) {
	mockRepo := new(MockInsightRepository)
	service := NewInsightsService(mockRepo)

	assert.NotNil(t, service)
	assert.Implements(t, (*InsightsService)(nil), service)
}