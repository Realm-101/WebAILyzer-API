package analyzers

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPerformanceAnalyzer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer := NewPerformanceAnalyzer(logger)
	assert.NotNil(t, analyzer)
	assert.NotNil(t, analyzer.logger)
}

func TestPerformanceAnalyzer_Analyze(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer := NewPerformanceAnalyzer(logger)

	tests := []struct {
		name              string
		url               string
		headers           http.Header
		body              []byte
		loadTimes         LoadTimeMetrics
		userAgent         string
		expectedMinScore  int
		expectError       bool
	}{
		{
			name: "Optimized HTML page",
			url:  "https://example.com",
			headers: http.Header{
				"Content-Type":     []string{"text/html; charset=UTF-8"},
				"Content-Encoding": []string{"gzip"},
			},
			body: []byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<title>Optimized Page</title>
					<link rel="stylesheet" href="styles.css">
				</head>
				<body>
					<img src="image.webp" alt="Optimized image">
					<script src="script.js" async></script>
				</body>
				</html>
			`),
			loadTimes: LoadTimeMetrics{
				DNSLookupTime:    50 * time.Millisecond,
				ConnectionTime:   100 * time.Millisecond,
				TLSHandshakeTime: 200 * time.Millisecond,
				ServerTime:       300 * time.Millisecond,
				TransferTime:     150 * time.Millisecond,
				TotalTime:        800 * time.Millisecond,
			},
			userAgent:        "TestAgent/1.0",
			expectedMinScore: 70,
			expectError:      false,
		},
		{
			name: "Unoptimized HTML page",
			url:  "https://example.com/slow",
			headers: http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
			body: []byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<title>Unoptimized Page</title>
					<script src="jquery.js"></script>
					<script src="blocking-script.js"></script>
					<style>
						body { margin: 0; }
						.container { padding: 20px; }
					</style>
				</head>
				<body style="background: white;">
					<img src="large-image.bmp">
					<img src="another-image.tiff">
					<div style="color: red;">Inline styles everywhere</div>
					<script>
						console.log('Inline JavaScript');
						// More inline code
					</script>
				</body>
				</html>
			`),
			loadTimes: LoadTimeMetrics{
				DNSLookupTime:    200 * time.Millisecond,
				ConnectionTime:   500 * time.Millisecond,
				TLSHandshakeTime: 800 * time.Millisecond,
				ServerTime:       2000 * time.Millisecond,
				TransferTime:     1000 * time.Millisecond,
				TotalTime:        4500 * time.Millisecond,
			},
			userAgent:        "TestAgent/1.0",
			expectedMinScore: 0, // Should be low due to many issues
			expectError:      false,
		},
		{
			name: "Empty HTML page",
			url:  "https://example.com/empty",
			headers: http.Header{
				"Content-Type": []string{"text/html"},
			},
			body: []byte(`<!DOCTYPE html><html><head><title>Empty</title></head><body></body></html>`),
			loadTimes: LoadTimeMetrics{
				TotalTime: 100 * time.Millisecond,
			},
			userAgent:        "TestAgent/1.0",
			expectedMinScore: 80, // Should score well due to simplicity
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			result, err := analyzer.Analyze(ctx, tt.url, tt.headers, tt.body, tt.loadTimes, tt.userAgent)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.NotNil(t, result)
			
			// Check basic structure
			assert.NotNil(t, result.LoadTimes)
			assert.NotNil(t, result.ResourceSizes)
			assert.NotNil(t, result.CoreWebVitals)
			assert.NotNil(t, result.OptimizationScore)
			assert.NotNil(t, result.Metadata)
			
			// Check metadata
			assert.Equal(t, tt.url, result.Metadata.URL)
			assert.Equal(t, tt.userAgent, result.Metadata.UserAgent)
			assert.True(t, result.Metadata.AnalysisTime > 0)
			
			// Check load times are preserved
			assert.Equal(t, tt.loadTimes.TotalTime, result.LoadTimes.TotalTime)
			
			// Check resource sizes
			assert.Equal(t, int64(len(tt.body)), result.ResourceSizes.HTMLSize)
			assert.GreaterOrEqual(t, result.ResourceSizes.TotalResources, 0)
			
			// Check Core Web Vitals structure
			assert.NotEmpty(t, result.CoreWebVitals.FCP.Rating)
			assert.NotEmpty(t, result.CoreWebVitals.LCP.Rating)
			assert.NotEmpty(t, result.CoreWebVitals.CLS.Rating)
			assert.NotEmpty(t, result.CoreWebVitals.FID.Rating)
			
			// Check optimization score
			assert.GreaterOrEqual(t, result.OptimizationScore.OverallScore, tt.expectedMinScore)
			assert.LessOrEqual(t, result.OptimizationScore.OverallScore, 100)
			
			// Check that suggestions are provided for low scores
			if result.OptimizationScore.OverallScore < 80 {
				assert.Greater(t, len(result.OptimizationScore.Suggestions), 0)
			}
		})
	}
}

func TestPerformanceAnalyzer_analyzeResourceSizes(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer := NewPerformanceAnalyzer(logger)

	tests := []struct {
		name           string
		body           []byte
		expectedCSS    int
		expectedJS     int
		expectedImages int
	}{
		{
			name: "Page with multiple resources",
			body: []byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<link rel="stylesheet" href="style1.css">
					<link rel="stylesheet" href="style2.css">
					<style>body { margin: 0; }</style>
				</head>
				<body>
					<img src="image1.jpg" alt="Image 1">
					<img src="image2.png" alt="Image 2">
					<script src="script1.js"></script>
					<script src="script2.js"></script>
					<script>console.log('inline');</script>
				</body>
				</html>
			`),
			expectedCSS:    3, // 2 external + 1 inline
			expectedJS:     3, // 2 external + 1 inline
			expectedImages: 2,
		},
		{
			name: "Simple page",
			body: []byte(`
				<!DOCTYPE html>
				<html>
				<head><title>Simple</title></head>
				<body><p>Hello World</p></body>
				</html>
			`),
			expectedCSS:    0,
			expectedJS:     0,
			expectedImages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeResourceSizes(tt.body)
			
			assert.Equal(t, int64(len(tt.body)), result.HTMLSize)
			assert.Equal(t, tt.expectedCSS, result.CSSResources)
			assert.Equal(t, tt.expectedJS, result.JSResources)
			assert.Equal(t, tt.expectedImages, result.ImageResources)
			assert.Equal(t, tt.expectedCSS+tt.expectedJS+tt.expectedImages, result.TotalResources)
			assert.Greater(t, result.EstimatedSize, result.HTMLSize)
		})
	}
}

func TestPerformanceAnalyzer_calculateCoreWebVitals(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer := NewPerformanceAnalyzer(logger)

	body := []byte(`<html><body><p>Test</p></body></html>`)
	resources := ResourceSizeMetrics{
		HTMLSize:       int64(len(body)),
		TotalResources: 5,
	}
	loadTimes := LoadTimeMetrics{
		ServerTime: 500 * time.Millisecond,
		TotalTime:  2000 * time.Millisecond,
	}

	result := analyzer.calculateCoreWebVitals(body, resources, loadTimes)

	// Check that all metrics are calculated
	assert.Greater(t, result.FCP.Value, 0.0)
	assert.Greater(t, result.LCP.Value, 0.0)
	assert.GreaterOrEqual(t, result.CLS.Value, 0.0)
	assert.GreaterOrEqual(t, result.FID.Value, 0.0)

	// Check that ratings are valid
	validRatings := []string{"good", "needs-improvement", "poor"}
	assert.Contains(t, validRatings, result.FCP.Rating)
	assert.Contains(t, validRatings, result.LCP.Rating)
	assert.Contains(t, validRatings, result.CLS.Rating)
	assert.Contains(t, validRatings, result.FID.Rating)

	// Check units
	assert.Equal(t, "ms", result.FCP.Unit)
	assert.Equal(t, "ms", result.LCP.Unit)
	assert.Equal(t, "score", result.CLS.Unit)
	assert.Equal(t, "ms", result.FID.Unit)
}

func TestPerformanceAnalyzer_analyzeImageOptimization(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer := NewPerformanceAnalyzer(logger)

	tests := []struct {
		name          string
		htmlContent   string
		expectedScore int
		expectIssues  bool
	}{
		{
			name: "Optimized images",
			htmlContent: `
				<img src="image1.webp" alt="Modern format image">
				<img src="image2.avif" alt="Another modern format">
			`,
			expectedScore: 100,
			expectIssues:  false,
		},
		{
			name: "Unoptimized images",
			htmlContent: `
				<img src="image1.bmp">
				<img src="image2.tiff" alt="Large format">
			`,
			expectedScore: 70, // Should lose points for large formats and missing alt
			expectIssues:  true,
		},
		{
			name: "Mixed optimization",
			htmlContent: `
				<img src="image1.webp" alt="Good image">
				<img src="image2.jpg">
			`,
			expectedScore: 90, // Should lose some points for missing alt
			expectIssues:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeImageOptimization(tt.htmlContent)
			
			assert.LessOrEqual(t, result.Score, tt.expectedScore)
			assert.GreaterOrEqual(t, result.Score, 0)
			assert.LessOrEqual(t, result.Score, 100)
			
			if tt.expectIssues {
				assert.Greater(t, len(result.Issues), 0)
				assert.Greater(t, len(result.Suggestions), 0)
			}
		})
	}
}

func TestPerformanceAnalyzer_RatingFunctions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer := NewPerformanceAnalyzer(logger)

	// Test FCP ratings
	assert.Equal(t, "good", analyzer.rateFCP(1500))
	assert.Equal(t, "needs-improvement", analyzer.rateFCP(2500))
	assert.Equal(t, "poor", analyzer.rateFCP(4000))

	// Test LCP ratings
	assert.Equal(t, "good", analyzer.rateLCP(2000))
	assert.Equal(t, "needs-improvement", analyzer.rateLCP(3000))
	assert.Equal(t, "poor", analyzer.rateLCP(5000))

	// Test CLS ratings
	assert.Equal(t, "good", analyzer.rateCLS(0.05))
	assert.Equal(t, "needs-improvement", analyzer.rateCLS(0.15))
	assert.Equal(t, "poor", analyzer.rateCLS(0.3))

	// Test FID ratings
	assert.Equal(t, "good", analyzer.rateFID(50))
	assert.Equal(t, "needs-improvement", analyzer.rateFID(200))
	assert.Equal(t, "poor", analyzer.rateFID(400))
}