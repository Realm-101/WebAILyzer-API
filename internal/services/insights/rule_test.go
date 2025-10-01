package insights

import (
	"context"
	"testing"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRule implements InsightRule for testing
type MockRule struct {
	*BaseRule
	evaluateFunc func(ctx context.Context, data *AnalysisData) ([]*models.Insight, error)
}

func NewMockRule(name, description string, priority models.Priority, insightType models.InsightType) *MockRule {
	return &MockRule{
		BaseRule: NewBaseRule(name, description, priority, insightType),
	}
}

func (m *MockRule) Evaluate(ctx context.Context, data *AnalysisData) ([]*models.Insight, error) {
	if m.evaluateFunc != nil {
		return m.evaluateFunc(ctx, data)
	}
	
	// Default implementation returns a single insight
	insight := m.CreateInsight(
		data.WorkspaceID,
		"Test Insight",
		"Test insight description",
		75,
		25,
		map[string]interface{}{"action": "test"},
		map[string]interface{}{"source": "test"},
	)
	
	return []*models.Insight{insight}, nil
}

func TestBaseRule(t *testing.T) {
	rule := NewBaseRule(
		"test-rule",
		"Test rule description",
		models.PriorityHigh,
		models.InsightTypePerformanceBottleneck,
	)
	
	assert.Equal(t, "test-rule", rule.Name())
	assert.Equal(t, "Test rule description", rule.Description())
	assert.Equal(t, models.PriorityHigh, rule.Priority())
	assert.Equal(t, models.InsightTypePerformanceBottleneck, rule.Type())
}

func TestCreateInsight(t *testing.T) {
	rule := NewBaseRule(
		"test-rule",
		"Test rule description",
		models.PriorityHigh,
		models.InsightTypePerformanceBottleneck,
	)
	
	workspaceID := uuid.New()
	recommendations := map[string]interface{}{
		"action": "optimize images",
		"priority": 1,
	}
	dataSource := map[string]interface{}{
		"analysis_id": "test-123",
		"metric": "load_time",
	}
	
	insight := rule.CreateInsight(
		workspaceID,
		"Slow Page Load",
		"Page loads too slowly",
		85,
		30,
		recommendations,
		dataSource,
	)
	
	assert.NotEqual(t, uuid.Nil, insight.ID)
	assert.Equal(t, workspaceID, insight.WorkspaceID)
	assert.Equal(t, models.InsightTypePerformanceBottleneck, insight.InsightType)
	assert.Equal(t, models.PriorityHigh, insight.Priority)
	assert.Equal(t, "Slow Page Load", insight.Title)
	assert.Equal(t, "Page loads too slowly", *insight.Description)
	assert.Equal(t, 85, *insight.ImpactScore)
	assert.Equal(t, 30, *insight.EffortScore)
	assert.Equal(t, recommendations, insight.Recommendations)
	assert.Equal(t, dataSource, insight.DataSource)
	assert.Equal(t, models.InsightStatusPending, insight.Status)
}

func TestCalculateImpactScore(t *testing.T) {
	tests := []struct {
		name     string
		factors  map[string]float64
		weights  map[string]float64
		expected int
	}{
		{
			name: "balanced factors",
			factors: map[string]float64{
				"performance": 0.8,
				"user_impact": 0.9,
				"frequency":   0.6,
			},
			weights: map[string]float64{
				"performance": 0.4,
				"user_impact": 0.4,
				"frequency":   0.2,
			},
			expected: 80, // (0.8*0.4 + 0.9*0.4 + 0.6*0.2) * 100 = 0.8 * 100
		},
		{
			name: "high impact factors",
			factors: map[string]float64{
				"performance": 1.0,
				"user_impact": 1.0,
			},
			weights: map[string]float64{
				"performance": 0.5,
				"user_impact": 0.5,
			},
			expected: 100,
		},
		{
			name: "low impact factors",
			factors: map[string]float64{
				"performance": 0.1,
				"user_impact": 0.2,
			},
			weights: map[string]float64{
				"performance": 0.5,
				"user_impact": 0.5,
			},
			expected: 15,
		},
		{
			name:     "no factors",
			factors:  map[string]float64{},
			weights:  map[string]float64{},
			expected: 0,
		},
		{
			name: "mismatched factors and weights",
			factors: map[string]float64{
				"performance": 0.8,
				"unknown":     0.9,
			},
			weights: map[string]float64{
				"performance": 1.0,
			},
			expected: 80,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateImpactScore(tt.factors, tt.weights)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateEffortScore(t *testing.T) {
	tests := []struct {
		name              string
		complexity        string
		resourcesRequired int
		timeEstimate      int
		expectedMin       int
		expectedMax       int
	}{
		{
			name:              "low complexity",
			complexity:        "low",
			resourcesRequired: 1,
			timeEstimate:      2,
			expectedMin:       20,
			expectedMax:       30,
		},
		{
			name:              "medium complexity",
			complexity:        "medium",
			resourcesRequired: 3,
			timeEstimate:      5,
			expectedMin:       50,
			expectedMax:       70,
		},
		{
			name:              "high complexity",
			complexity:        "high",
			resourcesRequired: 5,
			timeEstimate:      10,
			expectedMin:       80,
			expectedMax:       100,
		},
		{
			name:              "unknown complexity defaults to medium",
			complexity:        "unknown",
			resourcesRequired: 2,
			timeEstimate:      3,
			expectedMin:       45,
			expectedMax:       65,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateEffortScore(tt.complexity, tt.resourcesRequired, tt.timeEstimate)
			assert.GreaterOrEqual(t, result, tt.expectedMin)
			assert.LessOrEqual(t, result, tt.expectedMax)
			assert.GreaterOrEqual(t, result, 1)
			assert.LessOrEqual(t, result, 100)
		})
	}
}

func TestMockRule(t *testing.T) {
	rule := NewMockRule(
		"mock-rule",
		"Mock rule for testing",
		models.PriorityMedium,
		models.InsightTypeSEOOptimization,
	)
	
	workspaceID := uuid.New()
	data := &AnalysisData{
		WorkspaceID: workspaceID,
	}
	
	insights, err := rule.Evaluate(context.Background(), data)
	require.NoError(t, err)
	require.Len(t, insights, 1)
	
	insight := insights[0]
	assert.Equal(t, workspaceID, insight.WorkspaceID)
	assert.Equal(t, models.InsightTypeSEOOptimization, insight.InsightType)
	assert.Equal(t, models.PriorityMedium, insight.Priority)
	assert.Equal(t, "Test Insight", insight.Title)
}

func TestMockRuleWithCustomEvaluate(t *testing.T) {
	rule := NewMockRule(
		"custom-mock-rule",
		"Custom mock rule",
		models.PriorityHigh,
		models.InsightTypeAccessibilityIssue,
	)
	
	// Set custom evaluate function
	rule.evaluateFunc = func(ctx context.Context, data *AnalysisData) ([]*models.Insight, error) {
		insights := make([]*models.Insight, 2)
		for i := 0; i < 2; i++ {
			insights[i] = rule.CreateInsight(
				data.WorkspaceID,
				"Custom Insight",
				"Custom insight description",
				90,
				20,
				map[string]interface{}{"custom": true},
				map[string]interface{}{"index": i},
			)
		}
		return insights, nil
	}
	
	workspaceID := uuid.New()
	data := &AnalysisData{
		WorkspaceID: workspaceID,
	}
	
	insights, err := rule.Evaluate(context.Background(), data)
	require.NoError(t, err)
	require.Len(t, insights, 2)
	
	for i, insight := range insights {
		assert.Equal(t, workspaceID, insight.WorkspaceID)
		assert.Equal(t, models.InsightTypeAccessibilityIssue, insight.InsightType)
		assert.Equal(t, models.PriorityHigh, insight.Priority)
		assert.Equal(t, "Custom Insight", insight.Title)
		assert.Equal(t, i, insight.DataSource["index"])
	}
}