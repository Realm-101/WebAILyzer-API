package analyzers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// PerformanceAnalyzer handles performance analysis including Core Web Vitals
type PerformanceAnalyzer struct {
	logger *logrus.Logger
}

// NewPerformanceAnalyzer creates a new performance analyzer instance
func NewPerformanceAnalyzer(logger *logrus.Logger) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		logger: logger,
	}
}

// PerformanceAnalysisResult represents the result of performance analysis
type PerformanceAnalysisResult struct {
	LoadTimes       LoadTimeMetrics       `json:"load_times"`
	ResourceSizes   ResourceSizeMetrics   `json:"resource_sizes"`
	CoreWebVitals   CoreWebVitalsMetrics  `json:"core_web_vitals"`
	OptimizationScore OptimizationScoring `json:"optimization_score"`
	Metadata        PerformanceMetadata   `json:"metadata"`
}

// LoadTimeMetrics contains timing information
type LoadTimeMetrics struct {
	DNSLookupTime    time.Duration `json:"dns_lookup_time_ms"`
	ConnectionTime   time.Duration `json:"connection_time_ms"`
	TLSHandshakeTime time.Duration `json:"tls_handshake_time_ms"`
	ServerTime       time.Duration `json:"server_time_ms"`
	TransferTime     time.Duration `json:"transfer_time_ms"`
	TotalTime        time.Duration `json:"total_time_ms"`
}

// ResourceSizeMetrics contains resource size information
type ResourceSizeMetrics struct {
	HTMLSize       int64 `json:"html_size_bytes"`
	CSSResources   int   `json:"css_resources_count"`
	JSResources    int   `json:"js_resources_count"`
	ImageResources int   `json:"image_resources_count"`
	TotalResources int   `json:"total_resources_count"`
	EstimatedSize  int64 `json:"estimated_total_size_bytes"`
}

// CoreWebVitalsMetrics contains Core Web Vitals measurements
type CoreWebVitalsMetrics struct {
	FCP CoreWebVitalMetric `json:"first_contentful_paint"`
	LCP CoreWebVitalMetric `json:"largest_contentful_paint"`
	CLS CoreWebVitalMetric `json:"cumulative_layout_shift"`
	FID CoreWebVitalMetric `json:"first_input_delay"`
}

// CoreWebVitalMetric represents a single Core Web Vital measurement
type CoreWebVitalMetric struct {
	Value  float64 `json:"value"`
	Rating string  `json:"rating"` // "good", "needs-improvement", "poor"
	Unit   string  `json:"unit"`
}

// OptimizationScoring contains optimization scores for different resource types
type OptimizationScoring struct {
	OverallScore int                    `json:"overall_score"` // 0-100
	Images       ResourceOptimization   `json:"images"`
	CSS          ResourceOptimization   `json:"css"`
	JavaScript   ResourceOptimization   `json:"javascript"`
	Suggestions  []OptimizationSuggestion `json:"suggestions"`
}

// ResourceOptimization contains optimization details for a resource type
type ResourceOptimization struct {
	Score       int      `json:"score"` // 0-100
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
}

// OptimizationSuggestion represents a performance optimization suggestion
type OptimizationSuggestion struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"` // "high", "medium", "low"
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// PerformanceMetadata contains analysis metadata
type PerformanceMetadata struct {
	AnalysisTime time.Duration `json:"analysis_time_ms"`
	Timestamp    time.Time     `json:"timestamp"`
	UserAgent    string        `json:"user_agent"`
	URL          string        `json:"url"`
}

// Analyze performs comprehensive performance analysis
func (pa *PerformanceAnalyzer) Analyze(ctx context.Context, url string, headers http.Header, body []byte, loadTimes LoadTimeMetrics, userAgent string) (*PerformanceAnalysisResult, error) {
	startTime := time.Now()
	
	pa.logger.WithFields(logrus.Fields{
		"url":            url,
		"content_length": len(body),
		"user_agent":     userAgent,
	}).Debug("Starting performance analysis")

	// Analyze resource sizes
	resourceSizes := pa.analyzeResourceSizes(body)
	
	// Calculate Core Web Vitals (estimated based on content analysis)
	coreWebVitals := pa.calculateCoreWebVitals(body, resourceSizes, loadTimes)
	
	// Generate optimization scores and suggestions
	optimizationScore := pa.generateOptimizationScore(body, resourceSizes, headers)

	analysisTime := time.Since(startTime)

	result := &PerformanceAnalysisResult{
		LoadTimes:       loadTimes,
		ResourceSizes:   resourceSizes,
		CoreWebVitals:   coreWebVitals,
		OptimizationScore: optimizationScore,
		Metadata: PerformanceMetadata{
			AnalysisTime: analysisTime,
			Timestamp:    startTime,
			UserAgent:    userAgent,
			URL:          url,
		},
	}

	pa.logger.WithFields(logrus.Fields{
		"url":                url,
		"total_resources":    resourceSizes.TotalResources,
		"estimated_size":     resourceSizes.EstimatedSize,
		"overall_score":      optimizationScore.OverallScore,
		"analysis_time_ms":   analysisTime.Milliseconds(),
	}).Debug("Performance analysis completed")

	return result, nil
}

// analyzeResourceSizes analyzes the HTML content to estimate resource sizes and counts
func (pa *PerformanceAnalyzer) analyzeResourceSizes(body []byte) ResourceSizeMetrics {
	htmlContent := string(body)
	
	// Count CSS resources
	cssRegex := regexp.MustCompile(`<link[^>]*rel=["']stylesheet["'][^>]*>|<style[^>]*>`)
	cssMatches := cssRegex.FindAllString(htmlContent, -1)
	
	// Count JavaScript resources
	jsRegex := regexp.MustCompile(`<script[^>]*src=["'][^"']*["'][^>]*>|<script[^>]*>`)
	jsMatches := jsRegex.FindAllString(htmlContent, -1)
	
	// Count image resources
	imgRegex := regexp.MustCompile(`<img[^>]*src=["'][^"']*["'][^>]*>`)
	imgMatches := imgRegex.FindAllString(htmlContent, -1)
	
	// Estimate total size (rough calculation)
	estimatedSize := int64(len(body))
	estimatedSize += int64(len(cssMatches) * 15000)  // Estimate 15KB per CSS file
	estimatedSize += int64(len(jsMatches) * 25000)   // Estimate 25KB per JS file
	estimatedSize += int64(len(imgMatches) * 50000)  // Estimate 50KB per image

	return ResourceSizeMetrics{
		HTMLSize:       int64(len(body)),
		CSSResources:   len(cssMatches),
		JSResources:    len(jsMatches),
		ImageResources: len(imgMatches),
		TotalResources: len(cssMatches) + len(jsMatches) + len(imgMatches),
		EstimatedSize:  estimatedSize,
	}
}

// calculateCoreWebVitals estimates Core Web Vitals based on content analysis
func (pa *PerformanceAnalyzer) calculateCoreWebVitals(body []byte, resources ResourceSizeMetrics, loadTimes LoadTimeMetrics) CoreWebVitalsMetrics {
	// These are estimated values based on content analysis
	// In a real implementation, these would come from browser performance APIs
	
	// First Contentful Paint (FCP) - estimate based on HTML size and server time
	fcpValue := float64(loadTimes.ServerTime.Milliseconds()) + (float64(resources.HTMLSize) / 10000)
	fcp := CoreWebVitalMetric{
		Value:  fcpValue,
		Rating: pa.rateFCP(fcpValue),
		Unit:   "ms",
	}
	
	// Largest Contentful Paint (LCP) - estimate based on total resources and load time
	lcpValue := float64(loadTimes.TotalTime.Milliseconds()) + (float64(resources.TotalResources) * 100)
	lcp := CoreWebVitalMetric{
		Value:  lcpValue,
		Rating: pa.rateLCP(lcpValue),
		Unit:   "ms",
	}
	
	// Cumulative Layout Shift (CLS) - estimate based on image count and CSS resources
	clsValue := float64(resources.ImageResources) * 0.05
	if resources.CSSResources == 0 {
		clsValue += 0.1 // Penalty for no CSS (likely unstyled)
	}
	cls := CoreWebVitalMetric{
		Value:  clsValue,
		Rating: pa.rateCLS(clsValue),
		Unit:   "score",
	}
	
	// First Input Delay (FID) - estimate based on JavaScript resources
	fidValue := float64(resources.JSResources) * 20
	fid := CoreWebVitalMetric{
		Value:  fidValue,
		Rating: pa.rateFID(fidValue),
		Unit:   "ms",
	}

	return CoreWebVitalsMetrics{
		FCP: fcp,
		LCP: lcp,
		CLS: cls,
		FID: fid,
	}
}

// generateOptimizationScore analyzes content and generates optimization scores
func (pa *PerformanceAnalyzer) generateOptimizationScore(body []byte, resources ResourceSizeMetrics, headers http.Header) OptimizationScoring {
	htmlContent := string(body)
	
	// Analyze images
	imageOpt := pa.analyzeImageOptimization(htmlContent)
	
	// Analyze CSS
	cssOpt := pa.analyzeCSSOptimization(htmlContent, headers)
	
	// Analyze JavaScript
	jsOpt := pa.analyzeJavaScriptOptimization(htmlContent)
	
	// Calculate overall score
	overallScore := (imageOpt.Score + cssOpt.Score + jsOpt.Score) / 3
	
	// Generate suggestions
	suggestions := pa.generateSuggestions(imageOpt, cssOpt, jsOpt, resources)

	return OptimizationScoring{
		OverallScore: overallScore,
		Images:       imageOpt,
		CSS:          cssOpt,
		JavaScript:   jsOpt,
		Suggestions:  suggestions,
	}
}

// analyzeImageOptimization analyzes image optimization opportunities
func (pa *PerformanceAnalyzer) analyzeImageOptimization(htmlContent string) ResourceOptimization {
	score := 100
	var issues []string
	var suggestions []string
	
	// Check for images without alt attributes (simplified approach)
	imgTags := regexp.MustCompile(`<img[^>]*>`)
	imgWithAlt := regexp.MustCompile(`<img[^>]*alt=`)
	
	imgMatches := imgTags.FindAllString(htmlContent, -1)
	imagesWithoutAlt := 0
	for _, img := range imgMatches {
		if !imgWithAlt.MatchString(img) {
			imagesWithoutAlt++
		}
	}
	
	if imagesWithoutAlt > 0 {
		score -= 10
		issues = append(issues, fmt.Sprintf("%d images without alt attributes found", imagesWithoutAlt))
		suggestions = append(suggestions, "Add alt attributes to all images for accessibility and SEO")
	}
	
	// Check for potentially large images (by file extension)
	largeImageFormats := regexp.MustCompile(`\.(bmp|tiff|tif)["']`)
	if largeImageFormats.MatchString(htmlContent) {
		score -= 20
		issues = append(issues, "Large image formats detected")
		suggestions = append(suggestions, "Convert BMP/TIFF images to WebP or JPEG for better compression")
	}
	
	// Check for modern image formats
	modernFormats := regexp.MustCompile(`\.(webp|avif)["']`)
	if !modernFormats.MatchString(htmlContent) {
		score -= 15
		issues = append(issues, "No modern image formats detected")
		suggestions = append(suggestions, "Use WebP or AVIF formats for better compression")
	}

	return ResourceOptimization{
		Score:       score,
		Issues:      issues,
		Suggestions: suggestions,
	}
}

// analyzeCSSOptimization analyzes CSS optimization opportunities
func (pa *PerformanceAnalyzer) analyzeCSSOptimization(htmlContent string, headers http.Header) ResourceOptimization {
	score := 100
	var issues []string
	var suggestions []string
	
	// Check for inline styles
	inlineStyles := regexp.MustCompile(`style=["'][^"']*["']`)
	inlineCount := len(inlineStyles.FindAllString(htmlContent, -1))
	if inlineCount > 5 {
		score -= 15
		issues = append(issues, fmt.Sprintf("Excessive inline styles found (%d)", inlineCount))
		suggestions = append(suggestions, "Move inline styles to external CSS files")
	}
	
	// Check for CSS compression
	contentEncoding := headers.Get("Content-Encoding")
	if !strings.Contains(contentEncoding, "gzip") && !strings.Contains(contentEncoding, "br") {
		score -= 20
		issues = append(issues, "CSS resources not compressed")
		suggestions = append(suggestions, "Enable gzip or Brotli compression for CSS files")
	}
	
	// Check for CSS minification (rough heuristic)
	if strings.Contains(htmlContent, "  ") && strings.Contains(htmlContent, "\n") {
		score -= 10
		issues = append(issues, "CSS may not be minified")
		suggestions = append(suggestions, "Minify CSS files to reduce file size")
	}

	return ResourceOptimization{
		Score:       score,
		Issues:      issues,
		Suggestions: suggestions,
	}
}

// analyzeJavaScriptOptimization analyzes JavaScript optimization opportunities
func (pa *PerformanceAnalyzer) analyzeJavaScriptOptimization(htmlContent string) ResourceOptimization {
	score := 100
	var issues []string
	var suggestions []string
	
	// Check for blocking scripts in head (simplified approach)
	headScripts := regexp.MustCompile(`<head[^>]*>[\s\S]*?<script[^>]*src=`)
	asyncScripts := regexp.MustCompile(`<script[^>]*async[^>]*src=`)
	deferScripts := regexp.MustCompile(`<script[^>]*defer[^>]*src=`)
	
	headScriptMatches := headScripts.FindAllString(htmlContent, -1)
	blockingScripts := 0
	for _, script := range headScriptMatches {
		if !asyncScripts.MatchString(script) && !deferScripts.MatchString(script) {
			blockingScripts++
		}
	}
	
	if blockingScripts > 0 {
		score -= 25
		issues = append(issues, fmt.Sprintf("%d render-blocking JavaScript files in head", blockingScripts))
		suggestions = append(suggestions, "Add async or defer attributes to non-critical scripts")
	}
	
	// Check for inline JavaScript
	inlineJS := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	inlineCount := len(inlineJS.FindAllString(htmlContent, -1))
	if inlineCount > 3 {
		score -= 15
		issues = append(issues, fmt.Sprintf("Excessive inline JavaScript found (%d blocks)", inlineCount))
		suggestions = append(suggestions, "Move inline JavaScript to external files")
	}
	
	// Check for jQuery (potential optimization opportunity)
	if strings.Contains(htmlContent, "jquery") {
		score -= 5
		issues = append(issues, "jQuery detected")
		suggestions = append(suggestions, "Consider using vanilla JavaScript or lighter alternatives")
	}

	return ResourceOptimization{
		Score:       score,
		Issues:      issues,
		Suggestions: suggestions,
	}
}

// generateSuggestions creates prioritized optimization suggestions
func (pa *PerformanceAnalyzer) generateSuggestions(imageOpt, cssOpt, jsOpt ResourceOptimization, resources ResourceSizeMetrics) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion
	
	// High priority suggestions
	if resources.EstimatedSize > 1000000 { // > 1MB
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "resource_size",
			Priority:    "high",
			Description: "Large page size detected",
			Impact:      "Reduce total page size to improve load times",
		})
	}
	
	if jsOpt.Score < 70 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "javascript",
			Priority:    "high",
			Description: "JavaScript optimization needed",
			Impact:      "Optimize JavaScript to improve page interactivity",
		})
	}
	
	// Medium priority suggestions
	if cssOpt.Score < 80 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "css",
			Priority:    "medium",
			Description: "CSS optimization opportunities",
			Impact:      "Optimize CSS to improve render times",
		})
	}
	
	if resources.ImageResources > 10 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "images",
			Priority:    "medium",
			Description: "Many images detected",
			Impact:      "Implement lazy loading for images",
		})
	}
	
	// Low priority suggestions
	if imageOpt.Score < 90 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "images",
			Priority:    "low",
			Description: "Image optimization opportunities",
			Impact:      "Optimize images for better compression",
		})
	}

	return suggestions
}

// Rating functions for Core Web Vitals
func (pa *PerformanceAnalyzer) rateFCP(value float64) string {
	if value <= 1800 {
		return "good"
	} else if value <= 3000 {
		return "needs-improvement"
	}
	return "poor"
}

func (pa *PerformanceAnalyzer) rateLCP(value float64) string {
	if value <= 2500 {
		return "good"
	} else if value <= 4000 {
		return "needs-improvement"
	}
	return "poor"
}

func (pa *PerformanceAnalyzer) rateCLS(value float64) string {
	if value <= 0.1 {
		return "good"
	} else if value <= 0.25 {
		return "needs-improvement"
	}
	return "poor"
}

func (pa *PerformanceAnalyzer) rateFID(value float64) string {
	if value <= 100 {
		return "good"
	} else if value <= 300 {
		return "needs-improvement"
	}
	return "poor"
}