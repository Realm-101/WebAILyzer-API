package services

import (
	"context"
	"fmt"

	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
	"github.com/google/uuid"
)

// insightsService implements the InsightsService interface
type insightsService struct {
	insightRepo repositories.InsightRepository
}

// NewInsightsService creates a new insights service
func NewInsightsService(insightRepo repositories.InsightRepository) InsightsService {
	return &insightsService{
		insightRepo: insightRepo,
	}
}

// GenerateInsights generates new insights for a workspace
func (s *insightsService) GenerateInsights(ctx context.Context, workspaceID uuid.UUID) ([]*models.Insight, error) {
	// This would typically involve running analysis rules against workspace data
	// For now, we'll return an empty slice as the actual rule execution
	// is handled by the existing rule framework
	return []*models.Insight{}, nil
}

// GetInsights retrieves insights for a workspace with optional filtering
func (s *insightsService) GetInsights(ctx context.Context, workspaceID uuid.UUID, filters *InsightFilters) ([]*models.Insight, error) {
	if filters == nil {
		filters = &InsightFilters{
			Limit:  50,
			Offset: 0,
		}
	}

	// Set default limit if not specified
	if filters.Limit <= 0 {
		filters.Limit = 50
	}

	// Use the enhanced filtering method
	return s.insightRepo.GetByFilters(ctx, workspaceID, filters.Status, filters.Type, filters.Priority, filters.Limit, filters.Offset)
}

// UpdateInsightStatus updates the status of an insight
func (s *insightsService) UpdateInsightStatus(ctx context.Context, insightID uuid.UUID, status models.InsightStatus) error {
	// Validate the status
	switch status {
	case models.InsightStatusPending, models.InsightStatusApplied, models.InsightStatusDismissed:
		// Valid status
	default:
		return fmt.Errorf("invalid insight status: %s", status)
	}

	return s.insightRepo.UpdateStatus(ctx, insightID, status)
}