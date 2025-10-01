package rules

import (
	"context"
	"fmt"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/services/insights"
	"github.com/google/uuid"
)

// SlowPageLoadRule detects pages with slow load times
type SlowPageLoadRule struct {
	*insights.BaseRule
}

// NewSlowPageLoadRule creates a new slow page load detection rule
func NewSlowPageLoadRule() *SlowPageLoadRule {
	return &SlowPageLoadRule{
		BaseRule: insights.NewBaseRule(
			"slow-page-load",
			"Detects pages with slow load times that impact user experience",
			models.PriorityHigh,
			models.InsightTypePerformanceBottleneck,
		),
	}
}

// Evaluate analyzes load times and generates insights for slow pages
func (r *SlowPageLoadRule) Evaluate(ctx context.Context, data *insights.AnalysisData) ([]*models.Insight, error) {
	var insightsList []*models.Insight
	
	for _, analysis := range data.AnalysisResults {
		if analysis.PerformanceMetrics == nil {
			continue
		}
		
		perfData := analysis.PerformanceMetrics
		
		// Extract load times
		loadTimes, ok := perfData["load_times"].(map[string]interface{})
		if !ok {
			continue
		}
		
		totalTimeMs, ok := loadTimes["total_time_ms"].(float64)
		if !ok {
			continue
		}
		
		// Thresholds for slow page load (in milliseconds)
		const (
			slowThreshold     = 3000  // 3 seconds
			verySlowThreshold = 5000  // 5 seconds
		)
		
		if totalTimeMs > slowThreshold {
			priority := models.PriorityMedium
			if totalTimeMs > verySlowThreshold {
				priority = models.PriorityHigh
			}
			
			// Calculate impact score based on load time
			impactFactors := map[string]float64{
				"load_time_impact": (totalTimeMs - slowThreshold) / slowThreshold,
				"user_experience":  0.9, // High impact on UX
				"seo_impact":       0.8, // High impact on SEO
			}
			
			weights := map[string]float64{
				"load_time_impact": 0.5,
				"user_experience":  0.3,
				"seo_impact":       0.2,
			}
			
			impactScore := insights.CalculateImpactScore(impactFactors, weights)
			effortScore := insights.CalculateEffortScore("medium", 3, 5)
			
			recommendations := map[string]interface{}{
				"optimize_images":     "Compress and optimize images",
				"minify_resources":    "Minify CSS and JavaScript files",
				"enable_compression":  "Enable gzip/brotli compression",
				"cdn_implementation":  "Consider using a CDN",
				"server_optimization": "Optimize server response times",
			}
			
			dataSource := map[string]interface{}{
				"analysis_id":    analysis.ID.String(),
				"url":           analysis.URL,
				"total_time_ms": totalTimeMs,
				"threshold_ms":  slowThreshold,
			}
			
			title := fmt.Sprintf("Slow page load detected (%.0fms)", totalTimeMs)
			description := fmt.Sprintf("Page load time of %.0fms exceeds recommended threshold of %dms, potentially impacting user experience and SEO rankings", totalTimeMs, slowThreshold)
			
			insight := &models.Insight{
				ID:              uuid.New(),
				WorkspaceID:     data.WorkspaceID,
				InsightType:     models.InsightTypePerformanceBottleneck,
				Priority:        priority,
				Title:           title,
				Description:     &description,
				ImpactScore:     &impactScore,
				EffortScore:     &effortScore,
				Recommendations: recommendations,
				DataSource:      dataSource,
				Status:          models.InsightStatusPending,
			}
			
			insightsList = append(insightsList, insight)
		}
	}
	
	return insightsList, nil
}

// LargeResourceSizeRule detects pages with large resource sizes
type LargeResourceSizeRule struct {
	*insights.BaseRule
}

// NewLargeResourceSizeRule creates a new large resource size detection rule
func NewLargeResourceSizeRule() *LargeResourceSizeRule {
	return &LargeResourceSizeRule{
		BaseRule: insights.NewBaseRule(
			"large-resource-size",
			"Detects pages with large resource sizes that slow down loading",
			models.PriorityMedium,
			models.InsightTypePerformanceBottleneck,
		),
	}
}

// Evaluate analyzes resource sizes and generates insights for optimization opportunities
func (r *LargeResourceSizeRule) Evaluate(ctx context.Context, data *insights.AnalysisData) ([]*models.Insight, error) {
	var insightsList []*models.Insight
	
	for _, analysis := range data.AnalysisResults {
		if analysis.PerformanceMetrics == nil {
			continue
		}
		
		perfData := analysis.PerformanceMetrics
		
		// Extract resource sizes
		resourceSizes, ok := perfData["resource_sizes"].(map[string]interface{})
		if !ok {
			continue
		}
		
		estimatedSizeBytes, ok := resourceSizes["estimated_total_size_bytes"].(float64)
		if !ok {
			continue
		}
		
		// Thresholds for large resource sizes (in bytes)
		const (
			largeThreshold     = 1000000  // 1MB
			veryLargeThreshold = 3000000  // 3MB
		)
		
		if estimatedSizeBytes > largeThreshold {
			priority := models.PriorityMedium
			if estimatedSizeBytes > veryLargeThreshold {
				priority = models.PriorityHigh
			}
			
			// Calculate impact score based on resource size
			impactFactors := map[string]float64{
				"size_impact":     (estimatedSizeBytes - largeThreshold) / largeThreshold,
				"mobile_impact":   0.9, // High impact on mobile users
				"bandwidth_cost":  0.7, // Impact on bandwidth costs
			}
			
			weights := map[string]float64{
				"size_impact":     0.5,
				"mobile_impact":   0.3,
				"bandwidth_cost":  0.2,
			}
			
			impactScore := insights.CalculateImpactScore(impactFactors, weights)
			effortScore := insights.CalculateEffortScore("medium", 2, 3)
			
			// Get resource counts for specific recommendations
			totalResources, _ := resourceSizes["total_resources_count"].(float64)
			imageResources, _ := resourceSizes["image_resources_count"].(float64)
			jsResources, _ := resourceSizes["js_resources_count"].(float64)
			cssResources, _ := resourceSizes["css_resources_count"].(float64)
			
			recommendations := map[string]interface{}{
				"image_optimization": fmt.Sprintf("Optimize %d images for better compression", int(imageResources)),
				"resource_bundling":  fmt.Sprintf("Bundle and minify %d JavaScript and %d CSS files", int(jsResources), int(cssResources)),
				"lazy_loading":       "Implement lazy loading for non-critical resources",
				"compression":        "Enable gzip/brotli compression for all text resources",
				"cdn_usage":          "Use a CDN to serve static resources",
			}
			
			dataSource := map[string]interface{}{
				"analysis_id":              analysis.ID.String(),
				"url":                     analysis.URL,
				"estimated_size_bytes":    estimatedSizeBytes,
				"total_resources":         totalResources,
				"image_resources":         imageResources,
				"js_resources":           jsResources,
				"css_resources":          cssResources,
			}
			
			sizeMB := estimatedSizeBytes / 1000000
			title := fmt.Sprintf("Large page size detected (%.1fMB)", sizeMB)
			description := fmt.Sprintf("Page size of %.1fMB exceeds recommended threshold of 1MB, potentially causing slow loading on mobile devices and high bandwidth usage", sizeMB)
			
			insight := &models.Insight{
				ID:              uuid.New(),
				WorkspaceID:     data.WorkspaceID,
				InsightType:     models.InsightTypePerformanceBottleneck,
				Priority:        priority,
				Title:           title,
				Description:     &description,
				ImpactScore:     &impactScore,
				EffortScore:     &effortScore,
				Recommendations: recommendations,
				DataSource:      dataSource,
				Status:          models.InsightStatusPending,
			}
			
			insightsList = append(insightsList, insight)
		}
	}
	
	return insightsList, nil
}

// CoreWebVitalsRule detects poor Core Web Vitals scores
type CoreWebVitalsRule struct {
	*insights.BaseRule
}

// NewCoreWebVitalsRule creates a new Core Web Vitals detection rule
func NewCoreWebVitalsRule() *CoreWebVitalsRule {
	return &CoreWebVitalsRule{
		BaseRule: insights.NewBaseRule(
			"core-web-vitals",
			"Detects poor Core Web Vitals scores that impact user experience and SEO",
			models.PriorityHigh,
			models.InsightTypePerformanceBottleneck,
		),
	}
}

// Evaluate analyzes Core Web Vitals and generates insights for poor scores
func (r *CoreWebVitalsRule) Evaluate(ctx context.Context, data *insights.AnalysisData) ([]*models.Insight, error) {
	var insightsList []*models.Insight
	
	for _, analysis := range data.AnalysisResults {
		if analysis.PerformanceMetrics == nil {
			continue
		}
		
		perfData := analysis.PerformanceMetrics
		
		// Extract Core Web Vitals
		coreWebVitals, ok := perfData["core_web_vitals"].(map[string]interface{})
		if !ok {
			continue
		}
		
		// Check each Core Web Vital
		vitals := []struct {
			name        string
			displayName string
			key         string
		}{
			{"fcp", "First Contentful Paint", "first_contentful_paint"},
			{"lcp", "Largest Contentful Paint", "largest_contentful_paint"},
			{"cls", "Cumulative Layout Shift", "cumulative_layout_shift"},
			{"fid", "First Input Delay", "first_input_delay"},
		}
		
		for _, vital := range vitals {
			vitalData, ok := coreWebVitals[vital.key].(map[string]interface{})
			if !ok {
				continue
			}
			
			rating, ok := vitalData["rating"].(string)
			if !ok {
				continue
			}
			
			value, ok := vitalData["value"].(float64)
			if !ok {
				continue
			}
			
			unit, _ := vitalData["unit"].(string)
			
			// Generate insights for poor ratings
			if rating == "poor" || rating == "needs-improvement" {
				priority := models.PriorityMedium
				if rating == "poor" {
					priority = models.PriorityHigh
				}
				
				// Calculate impact score based on rating and vital type
				impactFactors := map[string]float64{
					"seo_impact":    0.9, // High SEO impact
					"ux_impact":     0.8, // High UX impact
					"rating_impact": r.getRatingImpact(rating),
				}
				
				weights := map[string]float64{
					"seo_impact":    0.4,
					"ux_impact":     0.4,
					"rating_impact": 0.2,
				}
				
				impactScore := insights.CalculateImpactScore(impactFactors, weights)
				effortScore := insights.CalculateEffortScore("medium", 3, 7)
				
				recommendations := r.getVitalRecommendations(vital.name)
				
				dataSource := map[string]interface{}{
					"analysis_id": analysis.ID.String(),
					"url":        analysis.URL,
					"vital_name": vital.name,
					"value":      value,
					"unit":       unit,
					"rating":     rating,
				}
				
				title := fmt.Sprintf("Poor %s score (%.1f%s)", vital.displayName, value, unit)
				description := fmt.Sprintf("%s score of %.1f%s is rated as '%s', which may negatively impact user experience and SEO rankings", vital.displayName, value, unit, rating)
				
				insight := &models.Insight{
					ID:              uuid.New(),
					WorkspaceID:     data.WorkspaceID,
					InsightType:     models.InsightTypePerformanceBottleneck,
					Priority:        priority,
					Title:           title,
					Description:     &description,
					ImpactScore:     &impactScore,
					EffortScore:     &effortScore,
					Recommendations: recommendations,
					DataSource:      dataSource,
					Status:          models.InsightStatusPending,
				}
				
				insightsList = append(insightsList, insight)
			}
		}
	}
	
	return insightsList, nil
}

// getRatingImpact returns impact factor based on rating
func (r *CoreWebVitalsRule) getRatingImpact(rating string) float64 {
	switch rating {
	case "poor":
		return 1.0
	case "needs-improvement":
		return 0.6
	case "good":
		return 0.0
	default:
		return 0.5
	}
}

// getVitalRecommendations returns specific recommendations for each Core Web Vital
func (r *CoreWebVitalsRule) getVitalRecommendations(vital string) map[string]interface{} {
	switch vital {
	case "fcp":
		return map[string]interface{}{
			"optimize_server_response": "Reduce server response times",
			"eliminate_render_blocking": "Eliminate render-blocking resources",
			"optimize_css":             "Optimize CSS delivery",
			"preload_key_resources":    "Preload key resources",
		}
	case "lcp":
		return map[string]interface{}{
			"optimize_images":          "Optimize images and media",
			"improve_server_response":  "Improve server response times",
			"eliminate_render_blocking": "Remove render-blocking JavaScript and CSS",
			"preload_lcp_element":      "Preload the LCP element",
		}
	case "cls":
		return map[string]interface{}{
			"size_images":              "Always include size attributes on images and video elements",
			"reserve_space":            "Reserve space for ad slots",
			"avoid_inserting_content":  "Avoid inserting content above existing content",
			"use_transform_animations": "Use transform animations instead of animating layout properties",
		}
	case "fid":
		return map[string]interface{}{
			"reduce_javascript":        "Reduce JavaScript execution time",
			"break_up_long_tasks":      "Break up long tasks",
			"optimize_interaction":     "Optimize your page for interaction readiness",
			"use_web_worker":          "Use a web worker for heavy computations",
		}
	default:
		return map[string]interface{}{
			"general_optimization": "Follow general performance optimization best practices",
		}
	}
}

