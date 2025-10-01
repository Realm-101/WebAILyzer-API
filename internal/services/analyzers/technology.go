package analyzers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/sirupsen/logrus"
)

// TechnologyAnalyzer handles technology detection using wappalyzer
type TechnologyAnalyzer struct {
	wappalyzer *wappalyzer.Wappalyze
	logger     *logrus.Logger
}

// NewTechnologyAnalyzer creates a new technology analyzer instance
func NewTechnologyAnalyzer(logger *logrus.Logger) (*TechnologyAnalyzer, error) {
	wapp, err := wappalyzer.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wappalyzer: %w", err)
	}

	return &TechnologyAnalyzer{
		wappalyzer: wapp,
		logger:     logger,
	}, nil
}

// TechnologyAnalysisResult represents the result of technology analysis
type TechnologyAnalysisResult struct {
	Technologies    map[string]struct{}            `json:"technologies"`
	TechnologyInfo  map[string]wappalyzer.AppInfo  `json:"technology_info"`
	Categories      map[string]wappalyzer.CatsInfo `json:"categories"`
	Metadata        TechnologyMetadata             `json:"metadata"`
}

// TechnologyMetadata contains analysis metadata and timing information
type TechnologyMetadata struct {
	AnalysisTime        time.Duration `json:"analysis_time_ms"`
	TechnologiesFound   int           `json:"technologies_found"`
	CategoriesFound     int           `json:"categories_found"`
	WappalyzerMetrics   WappalyzerMetrics `json:"wappalyzer_metrics"`
	Timestamp          time.Time     `json:"timestamp"`
	UserAgent          string        `json:"user_agent"`
	ContentType        string        `json:"content_type"`
	StatusCode         int           `json:"status_code"`
}

// WappalyzerMetrics contains internal wappalyzer performance metrics
type WappalyzerMetrics struct {
	TotalRequests      int           `json:"total_requests"`
	AverageDuration    time.Duration `json:"average_duration_ms"`
	TechnologiesFound  int           `json:"technologies_found"`
}

// Analyze performs technology detection on the provided HTTP response and body
func (ta *TechnologyAnalyzer) Analyze(ctx context.Context, headers http.Header, body []byte, userAgent string, statusCode int) (*TechnologyAnalysisResult, error) {
	startTime := time.Now()
	
	ta.logger.WithFields(logrus.Fields{
		"content_length": len(body),
		"status_code":   statusCode,
		"user_agent":    userAgent,
	}).Debug("Starting technology analysis")

	// Convert http.Header to map[string][]string for wappalyzer
	headerMap := ta.convertHeaders(headers)

	// Reset wappalyzer metrics for this analysis
	ta.wappalyzer.ResetMetrics()

	// Perform technology detection
	technologies := ta.wappalyzer.Fingerprint(headerMap, body)
	
	// Get detailed information about detected technologies
	techInfo := ta.wappalyzer.FingerprintWithInfo(headerMap, body)
	
	// Get category information
	categories := ta.wappalyzer.FingerprintWithCats(headerMap, body)
	
	// Get wappalyzer performance metrics
	wappMetrics := ta.wappalyzer.GetMetrics()

	analysisTime := time.Since(startTime)

	result := &TechnologyAnalysisResult{
		Technologies:   technologies,
		TechnologyInfo: techInfo,
		Categories:     categories,
		Metadata: TechnologyMetadata{
			AnalysisTime:      analysisTime,
			TechnologiesFound: len(technologies),
			CategoriesFound:   ta.countUniqueCategories(categories),
			WappalyzerMetrics: WappalyzerMetrics{
				TotalRequests:     int(wappMetrics.TotalRequests),
				AverageDuration:   wappMetrics.AverageDuration,
				TechnologiesFound: int(wappMetrics.TechnologiesFound),
			},
			Timestamp:   startTime,
			UserAgent:   userAgent,
			ContentType: headers.Get("Content-Type"),
			StatusCode:  statusCode,
		},
	}

	ta.logger.WithFields(logrus.Fields{
		"technologies_found": len(technologies),
		"categories_found":   result.Metadata.CategoriesFound,
		"analysis_time_ms":   analysisTime.Milliseconds(),
	}).Debug("Technology analysis completed")

	return result, nil
}

// AnalyzeWithTitle performs technology detection and also extracts the page title
func (ta *TechnologyAnalyzer) AnalyzeWithTitle(ctx context.Context, headers http.Header, body []byte, userAgent string, statusCode int) (*TechnologyAnalysisResult, string, error) {
	// Convert http.Header to map[string][]string for wappalyzer
	headerMap := ta.convertHeaders(headers)

	// Reset wappalyzer metrics for this analysis
	ta.wappalyzer.ResetMetrics()

	// Perform technology detection with title extraction
	technologies, title := ta.wappalyzer.FingerprintWithTitle(headerMap, body)
	
	// Get detailed information about detected technologies
	techInfo := ta.wappalyzer.FingerprintWithInfo(headerMap, body)
	
	// Get category information
	categories := ta.wappalyzer.FingerprintWithCats(headerMap, body)
	
	// Get wappalyzer performance metrics
	wappMetrics := ta.wappalyzer.GetMetrics()

	result := &TechnologyAnalysisResult{
		Technologies:   technologies,
		TechnologyInfo: techInfo,
		Categories:     categories,
		Metadata: TechnologyMetadata{
			TechnologiesFound: len(technologies),
			CategoriesFound:   ta.countUniqueCategories(categories),
			WappalyzerMetrics: WappalyzerMetrics{
				TotalRequests:     int(wappMetrics.TotalRequests),
				AverageDuration:   wappMetrics.AverageDuration,
				TechnologiesFound: int(wappMetrics.TechnologiesFound),
			},
			Timestamp:   time.Now(),
			UserAgent:   userAgent,
			ContentType: headers.Get("Content-Type"),
			StatusCode:  statusCode,
		},
	}

	return result, title, nil
}

// GetFingerprints returns the original wappalyzer fingerprints
func (ta *TechnologyAnalyzer) GetFingerprints() *wappalyzer.Fingerprints {
	return ta.wappalyzer.GetFingerprints()
}

// GetCompiledFingerprints returns the compiled wappalyzer fingerprints
func (ta *TechnologyAnalyzer) GetCompiledFingerprints() *wappalyzer.CompiledFingerprints {
	return ta.wappalyzer.GetCompiledFingerprints()
}

// convertHeaders converts http.Header to the format expected by wappalyzer
func (ta *TechnologyAnalyzer) convertHeaders(headers http.Header) map[string][]string {
	headerMap := make(map[string][]string)
	for key, values := range headers {
		headerMap[strings.ToLower(key)] = values
	}
	return headerMap
}

// countUniqueCategories counts the number of unique categories from the detected technologies
func (ta *TechnologyAnalyzer) countUniqueCategories(categories map[string]wappalyzer.CatsInfo) int {
	uniqueCategories := make(map[int]struct{})
	for _, catInfo := range categories {
		for _, cat := range catInfo.Cats {
			uniqueCategories[cat] = struct{}{}
		}
	}
	return len(uniqueCategories)
}

// GetMetrics returns the current wappalyzer metrics
func (ta *TechnologyAnalyzer) GetMetrics() wappalyzer.PerformanceMetrics {
	return ta.wappalyzer.GetMetrics()
}

// ResetMetrics resets the wappalyzer metrics
func (ta *TechnologyAnalyzer) ResetMetrics() {
	ta.wappalyzer.ResetMetrics()
}