package insights

import (
	"context"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/google/uuid"
)

// InsightRule defines the interface for insight generation rules
type InsightRule interface {
	// Evaluate analyzes the provided data and generates insights if applicable
	Evaluate(ctx context.Context, data *AnalysisData) ([]*models.Insight, error)
	
	// Priority returns the priority level of this rule
	Priority() models.Priority
	
	// Type returns the type of insights this rule generates
	Type() models.InsightType
	
	// Name returns a unique identifier for this rule
	Name() string
	
	// Description returns a human-readable description of what this rule does
	Description() string
}

// AnalysisData contains all the data needed for insight generation
type AnalysisData struct {
	WorkspaceID      uuid.UUID
	AnalysisResults  []*models.AnalysisResult
	Sessions         []*models.Session
	Events           []*models.Event
	Metrics          *models.MetricsResponse
	ExistingInsights []*models.Insight
}

// BaseRule provides common functionality for all insight rules
type BaseRule struct {
	name        string
	description string
	priority    models.Priority
	insightType models.InsightType
}

// NewBaseRule creates a new base rule with the specified parameters
func NewBaseRule(name, description string, priority models.Priority, insightType models.InsightType) *BaseRule {
	return &BaseRule{
		name:        name,
		description: description,
		priority:    priority,
		insightType: insightType,
	}
}

// Name returns the rule name
func (r *BaseRule) Name() string {
	return r.name
}

// Description returns the rule description
func (r *BaseRule) Description() string {
	return r.description
}

// Priority returns the rule priority
func (r *BaseRule) Priority() models.Priority {
	return r.priority
}

// Type returns the insight type
func (r *BaseRule) Type() models.InsightType {
	return r.insightType
}

// CreateInsight is a helper method to create a new insight with common fields
func (r *BaseRule) CreateInsight(workspaceID uuid.UUID, title, description string, impactScore, effortScore int, recommendations map[string]interface{}, dataSource map[string]interface{}) *models.Insight {
	return &models.Insight{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		InsightType:     r.insightType,
		Priority:        r.priority,
		Title:           title,
		Description:     &description,
		ImpactScore:     &impactScore,
		EffortScore:     &effortScore,
		Recommendations: recommendations,
		DataSource:      dataSource,
		Status:          models.InsightStatusPending,
	}
}

// CalculateImpactScore calculates impact score based on various factors
func CalculateImpactScore(factors map[string]float64, weights map[string]float64) int {
	var totalScore float64
	var totalWeight float64
	
	for factor, value := range factors {
		if weight, exists := weights[factor]; exists {
			totalScore += value * weight
			totalWeight += weight
		}
	}
	
	if totalWeight == 0 {
		return 0
	}
	
	score := (totalScore / totalWeight) * 100
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}
	
	return int(score)
}

// CalculateEffortScore calculates effort score based on implementation complexity
func CalculateEffortScore(complexity string, resourcesRequired int, timeEstimate int) int {
	baseScore := 0
	
	switch complexity {
	case "low":
		baseScore = 20
	case "medium":
		baseScore = 50
	case "high":
		baseScore = 80
	default:
		baseScore = 50
	}
	
	// Adjust based on resources and time
	resourceMultiplier := float64(resourcesRequired) / 5.0 // Normalize to 1-5 scale
	timeMultiplier := float64(timeEstimate) / 10.0        // Normalize to days
	
	finalScore := float64(baseScore) * (1 + resourceMultiplier*0.2 + timeMultiplier*0.1)
	
	if finalScore > 100 {
		finalScore = 100
	} else if finalScore < 1 {
		finalScore = 1
	}
	
	return int(finalScore)
}