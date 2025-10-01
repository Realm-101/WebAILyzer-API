package rules

import (
	"context"
	"testing"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/services/insights"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlowPageLoadRule(t *testing.T) {
	rule := NewSlowPageLoadRule()
	
	assert.Equal(t, "slow-page-load", rule.Name())
	assert.Equal(t, models.PriorityHigh, rule.Priority())
	assert.Equal(t, models.InsightTypePerformanceBottleneck, rule.Type())
}

func TestSlowPageLoadRule_Evaluate(t *testing.T) {
	rule := NewSlowPageLoadRule()
	workspaceID := uuid.New()
	
	tests := []struct {
		name           string
		totalTimeMs    float64
		expectedCount  int
		expectedPriority models.Priority
	}{
		{
			name:           "fast page load",
			totalTimeMs:    2000, // 2 seconds - below threshold
			expectedCount:  0,
		},
		{
			name:           "slow page load",
			totalTimeMs:    4000, // 4 seconds - above threshold but below very slow
			expectedCount:  1,
			expectedPriority: models.PriorityMedium,
		},
		{
			name:           "very slow page load",
			totalTimeMs:    6000, // 6 seconds - above very slow threshold
			expectedCount:  1,
			expectedPriority: models.PriorityHigh,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysisResult := &models.AnalysisResult{
				ID:  uuid.New(),
				URL: "https://example.com",
				PerformanceMetrics: map[string]interface{}{
					"load_times": map[string]interface{}{
						"total_time_ms": tt.totalTimeMs,
					},
				},
			}
			
			data := &insights.AnalysisData{
				WorkspaceID:     workspaceID,
				AnalysisResults: []*models.AnalysisResult{analysisResult},
			}
			
			insights, err := rule.Evaluate(context.Background(), data)
			require.NoError(t, err)
			assert.Len(t, insights, tt.expectedCount)
			
			if tt.expectedCount > 0 {
				insight := insights[0]
				assert.Equal(t, workspaceID, insight.WorkspaceID)
				assert.Equal(t, models.InsightTypePerformanceBottleneck, insight.InsightType)
				assert.Equal(t, tt.expectedPriority, insight.Priority)
				assert.Contains(t, insight.Title, "Slow page load detected")
				assert.NotNil(t, insight.Description)
				assert.NotNil(t, insight.ImpactScore)
				assert.NotNil(t, insight.EffortScore)
				assert.NotEmpty(t, insight.Recommendations)
				assert.NotEmpty(t, insight.DataSource)
				
				// Check data source contains expected fields
				dataSource := insight.DataSource
				assert.Equal(t, analysisResult.ID.String(), dataSource["analysis_id"])
				assert.Equal(t, analysisResult.URL, dataSource["url"])
				assert.Equal(t, tt.totalTimeMs, dataSource["total_time_ms"])
			}
		})
	}
}

func TestSlowPageLoadRule_EvaluateInvalidData(t *testing.T) {
	rule := NewSlowPageLoadRule()
	workspaceID := uuid.New()
	
	tests := []struct {
		name           string
		analysisResult *models.AnalysisResult
	}{
		{
			name: "nil performance data",
			analysisResult: &models.AnalysisResult{
				ID:                 uuid.New(),
				URL:                "https://example.com",
				PerformanceMetrics: nil,
			},
		},
		{
			name: "invalid performance data type",
			analysisResult: &models.AnalysisResult{
				ID:                 uuid.New(),
				URL:                "https://example.com",
				PerformanceMetrics: map[string]interface{}{"invalid": "data"},
			},
		},
		{
			name: "missing load_times",
			analysisResult: &models.AnalysisResult{
				ID:  uuid.New(),
				URL: "https://example.com",
				PerformanceMetrics: map[string]interface{}{
					"other_data": "value",
				},
			},
		},
		{
			name: "invalid load_times type",
			analysisResult: &models.AnalysisResult{
				ID:  uuid.New(),
				URL: "https://example.com",
				PerformanceMetrics: map[string]interface{}{
					"load_times": "invalid",
				},
			},
		},
		{
			name: "missing total_time_ms",
			analysisResult: &models.AnalysisResult{
				ID:  uuid.New(),
				URL: "https://example.com",
				PerformanceMetrics: map[string]interface{}{
					"load_times": map[string]interface{}{
						"other_time": 1000,
					},
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &insights.AnalysisData{
				WorkspaceID:     workspaceID,
				AnalysisResults: []*models.AnalysisResult{tt.analysisResult},
			}
			
			insights, err := rule.Evaluate(context.Background(), data)
			require.NoError(t, err)
			assert.Len(t, insights, 0)
		})
	}
}

func TestLargeResourceSizeRule(t *testing.T) {
	rule := NewLargeResourceSizeRule()
	
	assert.Equal(t, "large-resource-size", rule.Name())
	assert.Equal(t, models.PriorityMedium, rule.Priority())
	assert.Equal(t, models.InsightTypePerformanceBottleneck, rule.Type())
}

func TestLargeResourceSizeRule_Evaluate(t *testing.T) {
	rule := NewLargeResourceSizeRule()
	workspaceID := uuid.New()
	
	tests := []struct {
		name             string
		estimatedSizeBytes float64
		expectedCount    int
		expectedPriority models.Priority
	}{
		{
			name:             "small page size",
			estimatedSizeBytes: 500000, // 0.5MB - below threshold
			expectedCount:    0,
		},
		{
			name:             "large page size",
			estimatedSizeBytes: 2000000, // 2MB - above threshold but below very large
			expectedCount:    1,
			expectedPriority: models.PriorityMedium,
		},
		{
			name:             "very large page size",
			estimatedSizeBytes: 4000000, // 4MB - above very large threshold
			expectedCount:    1,
			expectedPriority: models.PriorityHigh,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysisResult := &models.AnalysisResult{
				ID:  uuid.New(),
				URL: "https://example.com",
				PerformanceMetrics: map[string]interface{}{
					"resource_sizes": map[string]interface{}{
						"estimated_total_size_bytes": tt.estimatedSizeBytes,
						"total_resources_count":      float64(50),
						"image_resources_count":      float64(20),
						"js_resources_count":         float64(15),
						"css_resources_count":        float64(10),
					},
				},
			}
			
			data := &insights.AnalysisData{
				WorkspaceID:     workspaceID,
				AnalysisResults: []*models.AnalysisResult{analysisResult},
			}
			
			insights, err := rule.Evaluate(context.Background(), data)
			require.NoError(t, err)
			assert.Len(t, insights, tt.expectedCount)
			
			if tt.expectedCount > 0 {
				insight := insights[0]
				assert.Equal(t, workspaceID, insight.WorkspaceID)
				assert.Equal(t, models.InsightTypePerformanceBottleneck, insight.InsightType)
				assert.Equal(t, tt.expectedPriority, insight.Priority)
				assert.Contains(t, insight.Title, "Large page size detected")
				assert.NotNil(t, insight.Description)
				assert.NotNil(t, insight.ImpactScore)
				assert.NotNil(t, insight.EffortScore)
				assert.NotEmpty(t, insight.Recommendations)
				assert.NotEmpty(t, insight.DataSource)
				
				// Check recommendations contain resource-specific advice
				recommendations := insight.Recommendations
				assert.Contains(t, recommendations, "image_optimization")
				assert.Contains(t, recommendations, "resource_bundling")
			}
		})
	}
}

func TestCoreWebVitalsRule(t *testing.T) {
	rule := NewCoreWebVitalsRule()
	
	assert.Equal(t, "core-web-vitals", rule.Name())
	assert.Equal(t, models.PriorityHigh, rule.Priority())
	assert.Equal(t, models.InsightTypePerformanceBottleneck, rule.Type())
}

func TestCoreWebVitalsRule_Evaluate(t *testing.T) {
	rule := NewCoreWebVitalsRule()
	workspaceID := uuid.New()
	
	tests := []struct {
		name             string
		coreWebVitals    map[string]interface{}
		expectedCount    int
		expectedPriorities []models.Priority
	}{
		{
			name: "all good vitals",
			coreWebVitals: map[string]interface{}{
				"first_contentful_paint": map[string]interface{}{
					"value":  1500.0,
					"rating": "good",
					"unit":   "ms",
				},
				"largest_contentful_paint": map[string]interface{}{
					"value":  2000.0,
					"rating": "good",
					"unit":   "ms",
				},
			},
			expectedCount: 0,
		},
		{
			name: "one poor vital",
			coreWebVitals: map[string]interface{}{
				"first_contentful_paint": map[string]interface{}{
					"value":  4000.0,
					"rating": "poor",
					"unit":   "ms",
				},
				"largest_contentful_paint": map[string]interface{}{
					"value":  2000.0,
					"rating": "good",
					"unit":   "ms",
				},
			},
			expectedCount:      1,
			expectedPriorities: []models.Priority{models.PriorityHigh},
		},
		{
			name: "mixed vitals",
			coreWebVitals: map[string]interface{}{
				"first_contentful_paint": map[string]interface{}{
					"value":  3500.0,
					"rating": "needs-improvement",
					"unit":   "ms",
				},
				"cumulative_layout_shift": map[string]interface{}{
					"value":  0.3,
					"rating": "poor",
					"unit":   "score",
				},
			},
			expectedCount:      2,
			expectedPriorities: []models.Priority{models.PriorityMedium, models.PriorityHigh},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysisResult := &models.AnalysisResult{
				ID:  uuid.New(),
				URL: "https://example.com",
				PerformanceMetrics: map[string]interface{}{
					"core_web_vitals": tt.coreWebVitals,
				},
			}
			
			data := &insights.AnalysisData{
				WorkspaceID:     workspaceID,
				AnalysisResults: []*models.AnalysisResult{analysisResult},
			}
			
			insights, err := rule.Evaluate(context.Background(), data)
			require.NoError(t, err)
			assert.Len(t, insights, tt.expectedCount)
			
			for i, insight := range insights {
				assert.Equal(t, workspaceID, insight.WorkspaceID)
				assert.Equal(t, models.InsightTypePerformanceBottleneck, insight.InsightType)
				if i < len(tt.expectedPriorities) {
					assert.Equal(t, tt.expectedPriorities[i], insight.Priority)
				}
				assert.Contains(t, insight.Title, "score")
				assert.NotNil(t, insight.Description)
				assert.NotNil(t, insight.ImpactScore)
				assert.NotNil(t, insight.EffortScore)
				assert.NotEmpty(t, insight.Recommendations)
				assert.NotEmpty(t, insight.DataSource)
			}
		})
	}
}

func TestCoreWebVitalsRule_GetRatingImpact(t *testing.T) {
	rule := NewCoreWebVitalsRule()
	
	tests := []struct {
		rating   string
		expected float64
	}{
		{"poor", 1.0},
		{"needs-improvement", 0.6},
		{"good", 0.0},
		{"unknown", 0.5},
	}
	
	for _, tt := range tests {
		t.Run(tt.rating, func(t *testing.T) {
			result := rule.getRatingImpact(tt.rating)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCoreWebVitalsRule_GetVitalRecommendations(t *testing.T) {
	rule := NewCoreWebVitalsRule()
	
	tests := []struct {
		vital                string
		expectedRecommendations []string
	}{
		{
			vital: "fcp",
			expectedRecommendations: []string{
				"optimize_server_response",
				"eliminate_render_blocking",
				"optimize_css",
				"preload_key_resources",
			},
		},
		{
			vital: "lcp",
			expectedRecommendations: []string{
				"optimize_images",
				"improve_server_response",
				"eliminate_render_blocking",
				"preload_lcp_element",
			},
		},
		{
			vital: "cls",
			expectedRecommendations: []string{
				"size_images",
				"reserve_space",
				"avoid_inserting_content",
				"use_transform_animations",
			},
		},
		{
			vital: "fid",
			expectedRecommendations: []string{
				"reduce_javascript",
				"break_up_long_tasks",
				"optimize_interaction",
				"use_web_worker",
			},
		},
		{
			vital: "unknown",
			expectedRecommendations: []string{
				"general_optimization",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.vital, func(t *testing.T) {
			recommendations := rule.getVitalRecommendations(tt.vital)
			
			for _, expectedRec := range tt.expectedRecommendations {
				assert.Contains(t, recommendations, expectedRec)
			}
		})
	}
}

func TestPerformanceRules_Integration(t *testing.T) {
	// Test all performance rules together
	rules := []insights.InsightRule{
		NewSlowPageLoadRule(),
		NewLargeResourceSizeRule(),
		NewCoreWebVitalsRule(),
	}
	
	workspaceID := uuid.New()
	
	// Create analysis result with multiple performance issues
	analysisResult := &models.AnalysisResult{
		ID:  uuid.New(),
		URL: "https://example.com",
		PerformanceMetrics: map[string]interface{}{
			"load_times": map[string]interface{}{
				"total_time_ms": 4500.0, // Slow
			},
			"resource_sizes": map[string]interface{}{
				"estimated_total_size_bytes": 2500000.0, // Large
				"total_resources_count":      float64(60),
				"image_resources_count":      float64(25),
				"js_resources_count":         float64(20),
				"css_resources_count":        float64(15),
			},
			"core_web_vitals": map[string]interface{}{
				"first_contentful_paint": map[string]interface{}{
					"value":  3500.0,
					"rating": "needs-improvement",
					"unit":   "ms",
				},
				"largest_contentful_paint": map[string]interface{}{
					"value":  5000.0,
					"rating": "poor",
					"unit":   "ms",
				},
			},
		},
	}
	
	data := &insights.AnalysisData{
		WorkspaceID:     workspaceID,
		AnalysisResults: []*models.AnalysisResult{analysisResult},
	}
	
	var allInsights []*models.Insight
	
	for _, rule := range rules {
		insights, err := rule.Evaluate(context.Background(), data)
		require.NoError(t, err)
		allInsights = append(allInsights, insights...)
	}
	
	// Should generate insights from all rules
	assert.GreaterOrEqual(t, len(allInsights), 3) // At least one from each rule
	
	// Check that we have different types of insights
	insightTitles := make(map[string]bool)
	for _, insight := range allInsights {
		insightTitles[insight.Title] = true
		assert.Equal(t, workspaceID, insight.WorkspaceID)
		assert.Equal(t, models.InsightTypePerformanceBottleneck, insight.InsightType)
		assert.NotNil(t, insight.ImpactScore)
		assert.NotNil(t, insight.EffortScore)
	}
	
	// Should have unique insights
	assert.Equal(t, len(allInsights), len(insightTitles))
}