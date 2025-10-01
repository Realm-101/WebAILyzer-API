package services

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"io"

	"github.com/google/uuid"
	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/repositories"
	"github.com/webailyzer/webailyzer-lite-api/internal/services/analyzers"
	"github.com/sirupsen/logrus"
)
// AnalysisServiceImpl implements the AnalysisService interface
type AnalysisServiceImpl struct {
	technologyAnalyzer   *analyzers.TechnologyAnalyzer
	performanceAnalyzer  *analyzers.PerformanceAnalyzer
	seoAnalyzer          *analyzers.SEOAnalyzer
	accessibilityAnalyzer *analyzers.AccessibilityAnalyzer
	securityAnalyzer     *analyzers.SecurityAnalyzer
	analysisRepo         repositories.AnalysisRepository
	httpClient           *http.Client
	logger               *logrus.Logger
}

// NewAnalysisService creates a new analysis service instance
func NewAnalysisService(
	analysisRepo repositories.AnalysisRepository,
	logger *logrus.Logger,
) (*AnalysisServiceImpl, error) {
	// Initialize technology analyzer
	techAnalyzer, err := analyzers.NewTechnologyAnalyzer(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize technology analyzer: %w", err)
	}

	// Initialize performance analyzer
	perfAnalyzer := analyzers.NewPerformanceAnalyzer(logger)

	// Initialize SEO analyzer
	seoAnalyzer := analyzers.NewSEOAnalyzer(logger)

	// Initialize accessibility analyzer
	accessibilityAnalyzer := analyzers.NewAccessibilityAnalyzer(logger)

	// Initialize security analyzer
	securityAnalyzer := analyzers.NewSecurityAnalyzer(logger)

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &AnalysisServiceImpl{
		technologyAnalyzer:    techAnalyzer,
		performanceAnalyzer:   perfAnalyzer,
		seoAnalyzer:           seoAnalyzer,
		accessibilityAnalyzer: accessibilityAnalyzer,
		securityAnalyzer:      securityAnalyzer,
		analysisRepo:          analysisRepo,
		httpClient:            httpClient,
		logger:                logger,
	}, nil
}

// AnalyzeURL performs comprehensive analysis of a URL
func (s *AnalysisServiceImpl) AnalyzeURL(ctx context.Context, req *models.AnalysisRequest) (*models.AnalysisResult, error) {
	startTime := time.Now()
	
	// Create analysis result with metadata
	result := &models.AnalysisResult{
		ID:          uuid.New(),
		WorkspaceID: req.WorkspaceID,
		SessionID:   req.SessionID,
		URL:         req.URL,
		CreatedAt:   startTime,
		UpdatedAt:   startTime,
	}

	// Fetch the webpage
	resp, body, err := s.fetchWebpage(ctx, req.URL, req.Options.UserAgent)
	if err != nil {
		s.logger.WithError(err).WithField("url", req.URL).Error("Failed to fetch webpage")
		return nil, fmt.Errorf("failed to fetch webpage: %w", err)
	}
	defer resp.Body.Close()

	// Perform technology detection using the new technology analyzer
	techResult, err := s.technologyAnalyzer.Analyze(ctx, resp.Header, body, s.getUserAgent(req.Options.UserAgent), resp.StatusCode)
	if err != nil {
		s.logger.WithError(err).WithField("url", req.URL).Error("Failed to analyze technologies")
		return nil, fmt.Errorf("failed to analyze technologies: %w", err)
	}

	// Store technology results with enhanced metadata
	result.Technologies = map[string]interface{}{
		"detected":         techResult.Technologies,
		"technology_info":  techResult.TechnologyInfo,
		"categories":       techResult.Categories,
		"metadata":         techResult.Metadata,
	}

	// Perform performance analysis if requested
	if req.Options.IncludePerformance {
		// Create load time metrics from the HTTP request timing
		loadTimes := analyzers.LoadTimeMetrics{
			TotalTime: time.Since(startTime),
			// Other timing metrics would be populated from actual HTTP client metrics
		}
		
		perfResult, err := s.performanceAnalyzer.Analyze(ctx, req.URL, resp.Header, body, loadTimes, s.getUserAgent(req.Options.UserAgent))
		if err != nil {
			s.logger.WithError(err).WithField("url", req.URL).Warn("Failed to analyze performance")
			result.PerformanceMetrics = make(map[string]interface{})
		} else {
			result.PerformanceMetrics = map[string]interface{}{
				"load_times":      perfResult.LoadTimes,
				"resource_sizes":  perfResult.ResourceSizes,
				"core_web_vitals": perfResult.CoreWebVitals,
				"optimization":    perfResult.OptimizationScore,
				"metadata":        perfResult.Metadata,
			}
		}
	} else {
		result.PerformanceMetrics = make(map[string]interface{})
	}

	// Perform SEO analysis if requested
	if req.Options.IncludeSEO {
		seoResult, err := s.seoAnalyzer.Analyze(ctx, req.URL, resp.Header, body, s.getUserAgent(req.Options.UserAgent))
		if err != nil {
			s.logger.WithError(err).WithField("url", req.URL).Warn("Failed to analyze SEO")
			result.SEOMetrics = make(map[string]interface{})
		} else {
			result.SEOMetrics = map[string]interface{}{
				"meta_tags":         seoResult.MetaTags,
				"content_structure": seoResult.ContentStructure,
				"structured_data":   seoResult.StructuredData,
				"seo_score":         seoResult.SEOScore,
				"metadata":          seoResult.Metadata,
			}
		}
	} else {
		result.SEOMetrics = make(map[string]interface{})
	}

	// Perform accessibility analysis if requested
	if req.Options.IncludeAccessibility {
		accessibilityResult, err := s.accessibilityAnalyzer.Analyze(ctx, req.URL, resp.Header, body, s.getUserAgent(req.Options.UserAgent))
		if err != nil {
			s.logger.WithError(err).WithField("url", req.URL).Warn("Failed to analyze accessibility")
			result.AccessibilityMetrics = make(map[string]interface{})
		} else {
			result.AccessibilityMetrics = map[string]interface{}{
				"wcag_compliance":     accessibilityResult.WCAGCompliance,
				"color_contrast":      accessibilityResult.ColorContrast,
				"alt_tag_analysis":    accessibilityResult.AltTagAnalysis,
				"keyboard_navigation": accessibilityResult.KeyboardNav,
				"form_accessibility":  accessibilityResult.FormAccessibility,
				"accessibility_score": accessibilityResult.AccessibilityScore,
				"issues":              accessibilityResult.Issues,
				"metadata":            accessibilityResult.Metadata,
			}
		}
	} else {
		result.AccessibilityMetrics = make(map[string]interface{})
	}

	// Perform security analysis if requested
	if req.Options.IncludeSecurity {
		securityResult, err := s.securityAnalyzer.Analyze(ctx, req.URL, resp.Header, body, s.getUserAgent(req.Options.UserAgent))
		if err != nil {
			s.logger.WithError(err).WithField("url", req.URL).Warn("Failed to analyze security")
			result.SecurityMetrics = make(map[string]interface{})
		} else {
			result.SecurityMetrics = map[string]interface{}{
				"https_configuration": securityResult.HTTPSConfiguration,
				"security_headers":    securityResult.SecurityHeaders,
				"security_score":      securityResult.SecurityScore,
				"vulnerabilities":     securityResult.Vulnerabilities,
				"metadata":            securityResult.Metadata,
			}
		}
	} else {
		result.SecurityMetrics = make(map[string]interface{})
	}

	// Store the analysis result in the database
	if err := s.analysisRepo.Create(ctx, result); err != nil {
		s.logger.WithError(err).WithField("analysis_id", result.ID).Error("Failed to store analysis result")
		return nil, fmt.Errorf("failed to store analysis result: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"analysis_id":    result.ID,
		"url":           req.URL,
		"technologies":  len(techResult.Technologies),
		"duration_ms":   time.Since(startTime).Milliseconds(),
	}).Info("Analysis completed successfully")

	return result, nil
}

// BatchAnalyze performs batch analysis of multiple URLs
func (s *AnalysisServiceImpl) BatchAnalyze(ctx context.Context, req *BatchAnalysisRequest) (*BatchAnalysisResult, error) {
	batchID := uuid.New()
	startTime := time.Now()
	
	result := &BatchAnalysisResult{
		BatchID:    batchID,
		Status:     "processing",
		Results:    make([]*models.AnalysisResult, 0, len(req.URLs)),
		FailedURLs: make([]string, 0),
		Progress: BatchProgress{
			Completed: 0,
			Total:     len(req.URLs),
		},
	}

	s.logger.WithFields(logrus.Fields{
		"batch_id":     batchID,
		"url_count":    len(req.URLs),
		"workspace_id": req.WorkspaceID,
	}).Info("Starting batch analysis")

	// Use concurrent processing with worker pool
	const maxConcurrency = 10 // Configurable concurrency limit
	concurrency := maxConcurrency
	if len(req.URLs) < concurrency {
		concurrency = len(req.URLs)
	}

	// Create channels for work distribution
	urlChan := make(chan string, len(req.URLs))
	resultChan := make(chan *batchAnalysisResult, len(req.URLs))
	
	// Send URLs to work channel
	for _, url := range req.URLs {
		urlChan <- url
	}
	close(urlChan)

	// Start worker goroutines
	for i := 0; i < concurrency; i++ {
		go s.batchWorker(ctx, req, urlChan, resultChan)
	}

	// Collect results
	var results []*models.AnalysisResult
	var failedURLs []string
	completed := 0

	for completed < len(req.URLs) {
		select {
		case batchResult := <-resultChan:
			completed++
			if batchResult.Error != nil {
				s.logger.WithError(batchResult.Error).WithField("url", batchResult.URL).Error("Failed to analyze URL in batch")
				failedURLs = append(failedURLs, batchResult.URL)
			} else {
				results = append(results, batchResult.Result)
			}
			
			// Update progress
			result.Progress.Completed = completed
			
		case <-ctx.Done():
			result.Status = "cancelled"
			result.Results = results
			result.FailedURLs = failedURLs
			return result, ctx.Err()
		}
	}

	result.Results = results
	result.FailedURLs = failedURLs
	
	// Determine final status
	if len(result.FailedURLs) == 0 {
		result.Status = "completed"
	} else if len(result.Results) == 0 {
		result.Status = "failed"
	} else {
		result.Status = "partial"
	}
	
	s.logger.WithFields(logrus.Fields{
		"batch_id":     batchID,
		"status":       result.Status,
		"url_count":    len(req.URLs),
		"successful":   len(result.Results),
		"failed":       len(result.FailedURLs),
		"workspace_id": req.WorkspaceID,
		"duration_ms":  time.Since(startTime).Milliseconds(),
		"concurrency":  concurrency,
	}).Info("Batch analysis completed")

	return result, nil
}

// GetAnalysisHistory retrieves historical analysis data
func (s *AnalysisServiceImpl) GetAnalysisHistory(ctx context.Context, workspaceID uuid.UUID, filters *AnalysisFilters) ([]*models.AnalysisResult, error) {
	// Convert service filters to repository filters
	repoFilters := &repositories.AnalysisFilters{
		WorkspaceID: workspaceID,
		SessionID:   filters.SessionID,
		StartDate:   filters.StartDate,
		EndDate:     filters.EndDate,
		Limit:       filters.Limit,
		Offset:      filters.Offset,
	}

	results, err := s.analysisRepo.GetByFilters(ctx, repoFilters)
	if err != nil {
		s.logger.WithError(err).WithField("workspace_id", workspaceID).Error("Failed to get analysis history")
		return nil, fmt.Errorf("failed to get analysis history: %w", err)
	}

	return results, nil
}

// fetchWebpage fetches a webpage and returns the response and body
func (s *AnalysisServiceImpl) fetchWebpage(ctx context.Context, url, userAgent string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", s.getUserAgent(userAgent))
	
	// Set other common headers
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp, body, nil
}



// getUserAgent returns the user agent to use for requests
func (s *AnalysisServiceImpl) getUserAgent(customUA string) string {
	if customUA != "" {
		return customUA
	}
	return "WebAIlyzer-Lite/1.0 (Website Analysis Bot)"
}

// batchAnalysisResult represents the result of a single URL analysis in a batch
type batchAnalysisResult struct {
	URL    string
	Result *models.AnalysisResult
	Error  error
}

// batchWorker processes URLs from the work channel
func (s *AnalysisServiceImpl) batchWorker(ctx context.Context, req *BatchAnalysisRequest, urlChan <-chan string, resultChan chan<- *batchAnalysisResult) {
	for url := range urlChan {
		analysisReq := &models.AnalysisRequest{
			URL:         url,
			WorkspaceID: req.WorkspaceID,
			Options:     req.Options,
		}

		analysisResult, err := s.AnalyzeURL(ctx, analysisReq)
		
		resultChan <- &batchAnalysisResult{
			URL:    url,
			Result: analysisResult,
			Error:  err,
		}
		
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}