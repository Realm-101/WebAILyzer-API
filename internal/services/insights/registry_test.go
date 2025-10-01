package insights

import (
	"context"
	"errors"
	"testing"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuleRegistry(t *testing.T) {
	registry := NewRuleRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, 0, registry.GetRuleCount())
}

func TestRegisterRule(t *testing.T) {
	registry := NewRuleRegistry()
	
	rule := NewMockRule(
		"test-rule",
		"Test rule",
		models.PriorityHigh,
		models.InsightTypePerformanceBottleneck,
	)
	
	err := registry.RegisterRule(rule)
	assert.NoError(t, err)
	assert.Equal(t, 1, registry.GetRuleCount())
}

func TestRegisterRuleDuplicate(t *testing.T) {
	registry := NewRuleRegistry()
	
	rule1 := NewMockRule(
		"test-rule",
		"Test rule 1",
		models.PriorityHigh,
		models.InsightTypePerformanceBottleneck,
	)
	
	rule2 := NewMockRule(
		"test-rule", // Same name
		"Test rule 2",
		models.PriorityMedium,
		models.InsightTypeSEOOptimization,
	)
	
	err := registry.RegisterRule(rule1)
	assert.NoError(t, err)
	
	err = registry.RegisterRule(rule2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Equal(t, 1, registry.GetRuleCount())
}

func TestRegisterRuleEmptyName(t *testing.T) {
	registry := NewRuleRegistry()
	
	rule := NewMockRule(
		"", // Empty name
		"Test rule",
		models.PriorityHigh,
		models.InsightTypePerformanceBottleneck,
	)
	
	err := registry.RegisterRule(rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
	assert.Equal(t, 0, registry.GetRuleCount())
}

func TestUnregisterRule(t *testing.T) {
	registry := NewRuleRegistry()
	
	rule := NewMockRule(
		"test-rule",
		"Test rule",
		models.PriorityHigh,
		models.InsightTypePerformanceBottleneck,
	)
	
	err := registry.RegisterRule(rule)
	require.NoError(t, err)
	
	err = registry.UnregisterRule("test-rule")
	assert.NoError(t, err)
	assert.Equal(t, 0, registry.GetRuleCount())
}

func TestUnregisterRuleNotFound(t *testing.T) {
	registry := NewRuleRegistry()
	
	err := registry.UnregisterRule("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetRule(t *testing.T) {
	registry := NewRuleRegistry()
	
	rule := NewMockRule(
		"test-rule",
		"Test rule",
		models.PriorityHigh,
		models.InsightTypePerformanceBottleneck,
	)
	
	err := registry.RegisterRule(rule)
	require.NoError(t, err)
	
	retrieved, err := registry.GetRule("test-rule")
	assert.NoError(t, err)
	assert.Equal(t, rule, retrieved)
}

func TestGetRuleNotFound(t *testing.T) {
	registry := NewRuleRegistry()
	
	_, err := registry.GetRule("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListRules(t *testing.T) {
	registry := NewRuleRegistry()
	
	// Register rules with different priorities
	rules := []*MockRule{
		NewMockRule("low-rule", "Low priority", models.PriorityLow, models.InsightTypePerformanceBottleneck),
		NewMockRule("critical-rule", "Critical priority", models.PriorityCritical, models.InsightTypeSEOOptimization),
		NewMockRule("medium-rule", "Medium priority", models.PriorityMedium, models.InsightTypeAccessibilityIssue),
		NewMockRule("high-rule", "High priority", models.PriorityHigh, models.InsightTypeConversionFunnel),
	}
	
	for _, rule := range rules {
		err := registry.RegisterRule(rule)
		require.NoError(t, err)
	}
	
	listedRules := registry.ListRules()
	assert.Len(t, listedRules, 4)
	
	// Check that rules are sorted by priority (critical, high, medium, low)
	expectedOrder := []string{"critical-rule", "high-rule", "medium-rule", "low-rule"}
	for i, rule := range listedRules {
		assert.Equal(t, expectedOrder[i], rule.Name())
	}
}

func TestExecuteRules(t *testing.T) {
	registry := NewRuleRegistry()
	workspaceID := uuid.New()
	
	// Register multiple rules
	rule1 := NewMockRule("rule1", "Rule 1", models.PriorityHigh, models.InsightTypePerformanceBottleneck)
	rule2 := NewMockRule("rule2", "Rule 2", models.PriorityMedium, models.InsightTypeSEOOptimization)
	
	err := registry.RegisterRule(rule1)
	require.NoError(t, err)
	err = registry.RegisterRule(rule2)
	require.NoError(t, err)
	
	data := &AnalysisData{
		WorkspaceID: workspaceID,
	}
	
	insights, err := registry.ExecuteRules(context.Background(), data)
	assert.NoError(t, err)
	assert.Len(t, insights, 2)
	
	// Check that insights are from both rules
	insightTypes := make(map[models.InsightType]bool)
	for _, insight := range insights {
		insightTypes[insight.InsightType] = true
		assert.Equal(t, workspaceID, insight.WorkspaceID)
	}
	
	assert.True(t, insightTypes[models.InsightTypePerformanceBottleneck])
	assert.True(t, insightTypes[models.InsightTypeSEOOptimization])
}

func TestExecuteRulesWithError(t *testing.T) {
	registry := NewRuleRegistry()
	workspaceID := uuid.New()
	
	// Create a rule that returns an error
	errorRule := NewMockRule("error-rule", "Error rule", models.PriorityHigh, models.InsightTypePerformanceBottleneck)
	errorRule.evaluateFunc = func(ctx context.Context, data *AnalysisData) ([]*models.Insight, error) {
		return nil, errors.New("evaluation failed")
	}
	
	// Create a successful rule
	successRule := NewMockRule("success-rule", "Success rule", models.PriorityMedium, models.InsightTypeSEOOptimization)
	
	err := registry.RegisterRule(errorRule)
	require.NoError(t, err)
	err = registry.RegisterRule(successRule)
	require.NoError(t, err)
	
	data := &AnalysisData{
		WorkspaceID: workspaceID,
	}
	
	insights, err := registry.ExecuteRules(context.Background(), data)
	assert.NoError(t, err) // Should not error if at least one rule succeeds
	assert.Len(t, insights, 1)
	assert.Equal(t, models.InsightTypeSEOOptimization, insights[0].InsightType)
}

func TestExecuteRulesAllFail(t *testing.T) {
	registry := NewRuleRegistry()
	workspaceID := uuid.New()
	
	// Create rules that all return errors
	errorRule1 := NewMockRule("error-rule1", "Error rule 1", models.PriorityHigh, models.InsightTypePerformanceBottleneck)
	errorRule1.evaluateFunc = func(ctx context.Context, data *AnalysisData) ([]*models.Insight, error) {
		return nil, errors.New("evaluation failed 1")
	}
	
	errorRule2 := NewMockRule("error-rule2", "Error rule 2", models.PriorityMedium, models.InsightTypeSEOOptimization)
	errorRule2.evaluateFunc = func(ctx context.Context, data *AnalysisData) ([]*models.Insight, error) {
		return nil, errors.New("evaluation failed 2")
	}
	
	err := registry.RegisterRule(errorRule1)
	require.NoError(t, err)
	err = registry.RegisterRule(errorRule2)
	require.NoError(t, err)
	
	data := &AnalysisData{
		WorkspaceID: workspaceID,
	}
	
	insights, err := registry.ExecuteRules(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all rules failed")
	assert.Nil(t, insights)
}

func TestExecuteRulesByType(t *testing.T) {
	registry := NewRuleRegistry()
	workspaceID := uuid.New()
	
	// Register rules of different types
	perfRule := NewMockRule("perf-rule", "Performance rule", models.PriorityHigh, models.InsightTypePerformanceBottleneck)
	seoRule := NewMockRule("seo-rule", "SEO rule", models.PriorityMedium, models.InsightTypeSEOOptimization)
	accessRule := NewMockRule("access-rule", "Accessibility rule", models.PriorityLow, models.InsightTypeAccessibilityIssue)
	
	err := registry.RegisterRule(perfRule)
	require.NoError(t, err)
	err = registry.RegisterRule(seoRule)
	require.NoError(t, err)
	err = registry.RegisterRule(accessRule)
	require.NoError(t, err)
	
	data := &AnalysisData{
		WorkspaceID: workspaceID,
	}
	
	// Execute only performance rules
	insights, err := registry.ExecuteRulesByType(context.Background(), data, models.InsightTypePerformanceBottleneck)
	assert.NoError(t, err)
	assert.Len(t, insights, 1)
	assert.Equal(t, models.InsightTypePerformanceBottleneck, insights[0].InsightType)
}

func TestGetRulesByType(t *testing.T) {
	registry := NewRuleRegistry()
	
	// Register rules of different types
	perfRule1 := NewMockRule("perf-rule1", "Performance rule 1", models.PriorityHigh, models.InsightTypePerformanceBottleneck)
	perfRule2 := NewMockRule("perf-rule2", "Performance rule 2", models.PriorityMedium, models.InsightTypePerformanceBottleneck)
	seoRule := NewMockRule("seo-rule", "SEO rule", models.PriorityLow, models.InsightTypeSEOOptimization)
	
	err := registry.RegisterRule(perfRule1)
	require.NoError(t, err)
	err = registry.RegisterRule(perfRule2)
	require.NoError(t, err)
	err = registry.RegisterRule(seoRule)
	require.NoError(t, err)
	
	perfRules := registry.GetRulesByType(models.InsightTypePerformanceBottleneck)
	assert.Len(t, perfRules, 2)
	
	seoRules := registry.GetRulesByType(models.InsightTypeSEOOptimization)
	assert.Len(t, seoRules, 1)
	
	accessRules := registry.GetRulesByType(models.InsightTypeAccessibilityIssue)
	assert.Len(t, accessRules, 0)
}

func TestDeduplicateInsights(t *testing.T) {
	registry := NewRuleRegistry()
	workspaceID := uuid.New()
	
	// Create duplicate insights (same type and title)
	insight1 := &models.Insight{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		InsightType: models.InsightTypePerformanceBottleneck,
		Priority:    models.PriorityHigh,
		Title:       "Slow Page Load",
		ImpactScore: &[]int{90}[0],
	}
	
	insight2 := &models.Insight{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		InsightType: models.InsightTypePerformanceBottleneck,
		Priority:    models.PriorityMedium,
		Title:       "Slow Page Load", // Same title and type
		ImpactScore: &[]int{80}[0],
	}
	
	insight3 := &models.Insight{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		InsightType: models.InsightTypeSEOOptimization,
		Priority:    models.PriorityLow,
		Title:       "Missing Meta Tags",
		ImpactScore: &[]int{70}[0],
	}
	
	insights := []*models.Insight{insight1, insight2, insight3}
	deduplicated := registry.deduplicateInsights(insights)
	
	assert.Len(t, deduplicated, 2) // Should remove one duplicate
	
	// Should keep the higher priority/impact one
	titles := make(map[string]bool)
	for _, insight := range deduplicated {
		key := insight.Title
		assert.False(t, titles[key], "Duplicate title found: %s", key)
		titles[key] = true
	}
}

func TestGetPriorityWeight(t *testing.T) {
	tests := []struct {
		priority models.Priority
		expected int
	}{
		{models.PriorityCritical, 4},
		{models.PriorityHigh, 3},
		{models.PriorityMedium, 2},
		{models.PriorityLow, 1},
		{models.Priority("unknown"), 0},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			result := getPriorityWeight(tt.priority)
			assert.Equal(t, tt.expected, result)
		})
	}
}