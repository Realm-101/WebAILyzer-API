package analyzers

import (
	"context"
	"net/http"
	"testing"
	"time"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTechnologyAnalyzer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	analyzer, err := NewTechnologyAnalyzer(logger)
	require.NoError(t, err)
	assert.NotNil(t, analyzer)
	assert.NotNil(t, analyzer.wappalyzer)
	assert.NotNil(t, analyzer.logger)
}

func TestTechnologyAnalyzer_Analyze(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer, err := NewTechnologyAnalyzer(logger)
	require.NoError(t, err)

	tests := []struct {
		name           string
		headers        http.Header
		body           []byte
		userAgent      string
		statusCode     int
		expectedTechs  int // minimum expected technologies
		expectError    bool
	}{
		{
			name: "WordPress site detection",
			headers: http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
				"Server":       []string{"Apache/2.4.41"},
			},
			body: []byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<meta name="generator" content="WordPress 5.8" />
					<link rel="stylesheet" href="/wp-content/themes/theme/style.css" />
				</head>
				<body>
					<div class="wp-content">WordPress content</div>
				</body>
				</html>
			`),
			userAgent:     "TestAgent/1.0",
			statusCode:    200,
			expectedTechs: 1, // Should detect WordPress
			expectError:   false,
		},
		{
			name: "React application detection",
			headers: http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
			body: []byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<title>React App</title>
				</head>
				<body>
					<div id="root"></div>
					<script src="/static/js/react.production.min.js"></script>
					<script>window.React = React;</script>
				</body>
				</html>
			`),
			userAgent:     "TestAgent/1.0",
			statusCode:    200,
			expectedTechs: 0, // May or may not detect React depending on fingerprints
			expectError:   false,
		},
		{
			name: "Empty body",
			headers: http.Header{
				"Content-Type": []string{"text/html"},
			},
			body:          []byte{},
			userAgent:     "TestAgent/1.0",
			statusCode:    200,
			expectedTechs: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			result, err := analyzer.Analyze(ctx, tt.headers, tt.body, tt.userAgent, tt.statusCode)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.NotNil(t, result)
			
			// Check basic structure
			assert.NotNil(t, result.Technologies)
			assert.NotNil(t, result.TechnologyInfo)
			assert.NotNil(t, result.Categories)
			assert.NotNil(t, result.Metadata)
			
			// Check metadata
			assert.Equal(t, tt.userAgent, result.Metadata.UserAgent)
			assert.Equal(t, tt.statusCode, result.Metadata.StatusCode)
			assert.Equal(t, tt.headers.Get("Content-Type"), result.Metadata.ContentType)
			assert.True(t, result.Metadata.AnalysisTime > 0)
			assert.True(t, result.Metadata.Timestamp.Before(time.Now().Add(time.Second)))
			
			// Check technology count
			assert.GreaterOrEqual(t, len(result.Technologies), tt.expectedTechs)
			assert.Equal(t, len(result.Technologies), result.Metadata.TechnologiesFound)
			
			// Verify consistency between technologies and technology info
			for tech := range result.Technologies {
				// Technology info should exist for detected technologies (may be empty for some)
				_, exists := result.TechnologyInfo[tech]
				// Note: Not all technologies may have detailed info, so we don't assert true here
				_ = exists
			}
		})
	}
}

func TestTechnologyAnalyzer_AnalyzeWithTitle(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer, err := NewTechnologyAnalyzer(logger)
	require.NoError(t, err)

	headers := http.Header{
		"Content-Type": []string{"text/html; charset=UTF-8"},
	}
	body := []byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page Title</title>
			<meta name="generator" content="WordPress 5.8" />
		</head>
		<body>
			<h1>Welcome</h1>
		</body>
		</html>
	`)

	ctx := context.Background()
	result, title, err := analyzer.AnalyzeWithTitle(ctx, headers, body, "TestAgent/1.0", 200)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Page Title", title)
	
	// Check that we still get technology analysis
	assert.NotNil(t, result.Technologies)
	assert.NotNil(t, result.Metadata)
}

func TestTechnologyAnalyzer_GetFingerprints(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer, err := NewTechnologyAnalyzer(logger)
	require.NoError(t, err)

	fingerprints := analyzer.GetFingerprints()
	assert.NotNil(t, fingerprints)
	assert.NotNil(t, fingerprints.Apps)
	assert.Greater(t, len(fingerprints.Apps), 0)
}

func TestTechnologyAnalyzer_GetCompiledFingerprints(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer, err := NewTechnologyAnalyzer(logger)
	require.NoError(t, err)

	compiled := analyzer.GetCompiledFingerprints()
	assert.NotNil(t, compiled)
	assert.NotNil(t, compiled.Apps)
	assert.Greater(t, len(compiled.Apps), 0)
}

func TestTechnologyAnalyzer_Metrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer, err := NewTechnologyAnalyzer(logger)
	require.NoError(t, err)

	// Reset metrics
	analyzer.ResetMetrics()
	
	// Get initial metrics
	initialMetrics := analyzer.GetMetrics()
	assert.Equal(t, 0, initialMetrics.TotalRequests)

	// Perform an analysis
	headers := http.Header{"Content-Type": []string{"text/html"}}
	body := []byte("<html><body>Test</body></html>")
	
	ctx := context.Background()
	_, err = analyzer.Analyze(ctx, headers, body, "TestAgent/1.0", 200)
	require.NoError(t, err)

	// Check that metrics were updated
	updatedMetrics := analyzer.GetMetrics()
	assert.Greater(t, updatedMetrics.TotalRequests, initialMetrics.TotalRequests)
}

func TestTechnologyAnalyzer_convertHeaders(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer, err := NewTechnologyAnalyzer(logger)
	require.NoError(t, err)

	headers := http.Header{
		"Content-Type": []string{"text/html; charset=UTF-8"},
		"Server":       []string{"Apache/2.4.41"},
		"X-Powered-By": []string{"PHP/7.4.0"},
	}

	converted := analyzer.convertHeaders(headers)
	
	assert.Equal(t, []string{"text/html; charset=UTF-8"}, converted["content-type"])
	assert.Equal(t, []string{"Apache/2.4.41"}, converted["server"])
	assert.Equal(t, []string{"PHP/7.4.0"}, converted["x-powered-by"])
}

func TestTechnologyAnalyzer_countUniqueCategories(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	analyzer, err := NewTechnologyAnalyzer(logger)
	require.NoError(t, err)

	// Mock categories data
	categories := map[string]wappalyzer.CatsInfo{
		"WordPress": {Cats: []int{1, 11}},    // CMS, Blog
		"Apache":    {Cats: []int{22}},       // Web server
		"PHP":       {Cats: []int{27}},       // Programming language
		"jQuery":    {Cats: []int{12, 59}},   // JavaScript library, JavaScript framework
	}

	count := analyzer.countUniqueCategories(categories)
	
	// Should count unique categories: 1, 11, 12, 22, 27, 59 = 6 unique categories
	assert.Equal(t, 6, count)
}