package analyzers

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityAnalyzer(t *testing.T) {
	logger := logrus.New()
	analyzer := NewSecurityAnalyzer(logger)
	
	assert.NotNil(t, analyzer)
	assert.Equal(t, logger, analyzer.logger)
}

func TestSecurityAnalyzer_Analyze(t *testing.T) {
	logger := logrus.New()
	analyzer := NewSecurityAnalyzer(logger)
	ctx := context.Background()

	tests := []struct {
		name        string
		url         string
		headers     http.Header
		body        []byte
		userAgent   string
		expectError bool
	}{
		{
			name:      "basic HTTPS analysis",
			url:       "https://example.com",
			headers:   http.Header{},
			body:      []byte("<html><body>Test</body></html>"),
			userAgent: "test-agent",
		},
		{
			name:      "HTTP analysis",
			url:       "http://example.com",
			headers:   http.Header{},
			body:      []byte("<html><body>Test</body></html>"),
			userAgent: "test-agent",
		},
		{
			name:        "invalid URL",
			url:         "://invalid-url",
			headers:     http.Header{},
			body:        []byte(""),
			userAgent:   "test-agent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.Analyze(ctx, tt.url, tt.headers, tt.body, tt.userAgent)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.url, result.Metadata.URL)
				assert.Equal(t, tt.userAgent, result.Metadata.UserAgent)
				assert.True(t, result.Metadata.AnalysisTime >= 0)
			}
		})
	}
}

func TestSecurityAnalyzer_AnalyzeHTTPSConfiguration(t *testing.T) {
	logger := logrus.New()
	analyzer := NewSecurityAnalyzer(logger)

	tests := []struct {
		name     string
		url      string
		body     []byte
		expected HTTPSConfig
	}{
		{
			name: "HTTPS URL",
			url:  "https://example.com",
			body: []byte("<html><body>Test</body></html>"),
			expected: HTTPSConfig{
				IsHTTPS: true,
				CertificateInfo: CertificateDetails{
					Valid: true,
				},
				HTTPSRedirect: true,
				MixedContent: MixedContentAnalysis{
					HasMixedContent: false,
					HTTPResources:   []string{},
					Count:          0,
				},
			},
		},
		{
			name: "HTTP URL",
			url:  "http://example.com",
			body: []byte("<html><body>Test</body></html>"),
			expected: HTTPSConfig{
				IsHTTPS: false,
				CertificateInfo: CertificateDetails{
					Valid: false,
				},
				HTTPSRedirect: false,
				MixedContent: MixedContentAnalysis{
					HasMixedContent: false,
					HTTPResources:   []string{},
					Count:          0,
				},
			},
		},
		{
			name: "HTTPS with mixed content",
			url:  "https://example.com",
			body: []byte(`<html><body><img src="http://example.com/image.jpg"><script src="http://example.com/script.js"></script></body></html>`),
			expected: HTTPSConfig{
				IsHTTPS: true,
				HTTPSRedirect: true, // This will be true for HTTPS URLs
				MixedContent: MixedContentAnalysis{
					HasMixedContent: true,
					Count:          2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := parseURL(tt.url)
			require.NoError(t, err)
			
			result := analyzer.analyzeHTTPSConfiguration(parsedURL, tt.body)
			
			assert.Equal(t, tt.expected.IsHTTPS, result.IsHTTPS)
			if tt.expected.HTTPSRedirect {
				assert.Equal(t, tt.expected.HTTPSRedirect, result.HTTPSRedirect)
			}
			assert.Equal(t, tt.expected.MixedContent.HasMixedContent, result.MixedContent.HasMixedContent)
			assert.Equal(t, tt.expected.MixedContent.Count, result.MixedContent.Count)
		})
	}
}

func TestSecurityAnalyzer_AnalyzeSecurityHeaders(t *testing.T) {
	logger := logrus.New()
	analyzer := NewSecurityAnalyzer(logger)

	tests := []struct {
		name     string
		headers  http.Header
		expected SecurityHeadersAnalysis
	}{
		{
			name:    "no security headers",
			headers: http.Header{},
			expected: SecurityHeadersAnalysis{
				HSTS:                  HeaderAnalysis{Present: false, Score: 0},
				ContentSecurityPolicy: HeaderAnalysis{Present: false, Score: 0},
				XFrameOptions:        HeaderAnalysis{Present: false, Score: 0},
				XContentTypeOptions:  HeaderAnalysis{Present: false, Score: 0},
				XSSProtection:        HeaderAnalysis{Present: false, Score: 0},
				ReferrerPolicy:       HeaderAnalysis{Present: false, Score: 0},
				PermissionsPolicy:    HeaderAnalysis{Present: false, Score: 0},
				ExpectCT:             HeaderAnalysis{Present: false, Score: 0},
			},
		},
		{
			name: "good security headers",
			headers: http.Header{
				"Strict-Transport-Security": []string{"max-age=31536000; includeSubDomains; preload"},
				"Content-Security-Policy":   []string{"default-src 'self'; script-src 'self'"},
				"X-Frame-Options":          []string{"DENY"},
				"X-Content-Type-Options":   []string{"nosniff"},
				"X-Xss-Protection":         []string{"1; mode=block"},
				"Referrer-Policy":          []string{"strict-origin-when-cross-origin"},
			},
			expected: SecurityHeadersAnalysis{
				HSTS:                  HeaderAnalysis{Present: true, Score: 70},
				ContentSecurityPolicy: HeaderAnalysis{Present: true, Score: 60},
				XFrameOptions:        HeaderAnalysis{Present: true, Score: 100},
				XContentTypeOptions:  HeaderAnalysis{Present: true, Score: 100},
				XSSProtection:        HeaderAnalysis{Present: true, Score: 80},
				ReferrerPolicy:       HeaderAnalysis{Present: true, Score: 90},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeSecurityHeaders(tt.headers)
			
			assert.Equal(t, tt.expected.HSTS.Present, result.HSTS.Present)
			assert.Equal(t, tt.expected.ContentSecurityPolicy.Present, result.ContentSecurityPolicy.Present)
			assert.Equal(t, tt.expected.XFrameOptions.Present, result.XFrameOptions.Present)
			assert.Equal(t, tt.expected.XContentTypeOptions.Present, result.XContentTypeOptions.Present)
			assert.Equal(t, tt.expected.XSSProtection.Present, result.XSSProtection.Present)
			assert.Equal(t, tt.expected.ReferrerPolicy.Present, result.ReferrerPolicy.Present)
			assert.Equal(t, tt.expected.PermissionsPolicy.Present, result.PermissionsPolicy.Present)
			assert.Equal(t, tt.expected.ExpectCT.Present, result.ExpectCT.Present)
		})
	}
}

func TestSecurityAnalyzer_AnalyzeHSTSHeader(t *testing.T) {
	logger := logrus.New()
	analyzer := NewSecurityAnalyzer(logger)

	tests := []struct {
		name     string
		headers  http.Header
		expected HeaderAnalysis
	}{
		{
			name:    "no HSTS header",
			headers: http.Header{},
			expected: HeaderAnalysis{
				Present: false,
				Score:   0,
				Issues:  []string{"HSTS header not present"},
			},
		},
		{
			name: "basic HSTS header",
			headers: http.Header{
				"Strict-Transport-Security": []string{"max-age=31536000"},
			},
			expected: HeaderAnalysis{
				Present: true,
				Value:   "max-age=31536000",
				Score:   45, // 70 - 15 (no includeSubDomains) - 10 (no preload)
			},
		},
		{
			name: "complete HSTS header",
			headers: http.Header{
				"Strict-Transport-Security": []string{"max-age=31536000; includeSubDomains; preload"},
			},
			expected: HeaderAnalysis{
				Present: true,
				Value:   "max-age=31536000; includeSubDomains; preload",
				Score:   70,
			},
		},
		{
			name: "short max-age HSTS header",
			headers: http.Header{
				"Strict-Transport-Security": []string{"max-age=3600; includeSubDomains; preload"},
			},
			expected: HeaderAnalysis{
				Present: true,
				Value:   "max-age=3600; includeSubDomains; preload",
				Score:   65, // 70 - 5 (short max-age)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeHSTSHeader(tt.headers)
			
			assert.Equal(t, tt.expected.Present, result.Present)
			assert.Equal(t, tt.expected.Value, result.Value)
			assert.Equal(t, tt.expected.Score, result.Score)
			if len(tt.expected.Issues) > 0 {
				assert.Contains(t, result.Issues, tt.expected.Issues[0])
			}
		})
	}
}

func TestSecurityAnalyzer_AnalyzeCSPHeader(t *testing.T) {
	logger := logrus.New()
	analyzer := NewSecurityAnalyzer(logger)

	tests := []struct {
		name     string
		headers  http.Header
		expected HeaderAnalysis
	}{
		{
			name:    "no CSP header",
			headers: http.Header{},
			expected: HeaderAnalysis{
				Present: false,
				Score:   0,
				Issues:  []string{"Content Security Policy header not present"},
			},
		},
		{
			name: "basic CSP header",
			headers: http.Header{
				"Content-Security-Policy": []string{"default-src 'self'; script-src 'self'"},
			},
			expected: HeaderAnalysis{
				Present: true,
				Value:   "default-src 'self'; script-src 'self'",
				Score:   60,
			},
		},
		{
			name: "CSP with unsafe-inline",
			headers: http.Header{
				"Content-Security-Policy": []string{"default-src 'self'; script-src 'self' 'unsafe-inline'"},
			},
			expected: HeaderAnalysis{
				Present: true,
				Value:   "default-src 'self'; script-src 'self' 'unsafe-inline'",
				Score:   40, // 60 - 20 (unsafe-inline)
			},
		},
		{
			name: "CSP with unsafe-eval",
			headers: http.Header{
				"Content-Security-Policy": []string{"default-src 'self'; script-src 'self' 'unsafe-eval'"},
			},
			expected: HeaderAnalysis{
				Present: true,
				Value:   "default-src 'self'; script-src 'self' 'unsafe-eval'",
				Score:   45, // 60 - 15 (unsafe-eval)
			},
		},
		{
			name: "CSP report-only",
			headers: http.Header{
				"Content-Security-Policy-Report-Only": []string{"default-src 'self'"},
			},
			expected: HeaderAnalysis{
				Present: true,
				Value:   "default-src 'self'",
				Score:   50, // 60 - 10 (missing script-src)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeCSPHeader(tt.headers)
			
			assert.Equal(t, tt.expected.Present, result.Present)
			assert.Equal(t, tt.expected.Value, result.Value)
			assert.Equal(t, tt.expected.Score, result.Score)
		})
	}
}

func TestSecurityAnalyzer_DetectVulnerabilities(t *testing.T) {
	logger := logrus.New()
	analyzer := NewSecurityAnalyzer(logger)

	tests := []struct {
		name     string
		body     []byte
		headers  http.Header
		url      string
		expected []SecurityVulnerability
	}{
		{
			name: "no vulnerabilities",
			body: []byte("<html><body>Clean content</body></html>"),
			headers: http.Header{
				"X-Frame-Options": []string{"DENY"},
			},
			url:      "https://example.com",
			expected: []SecurityVulnerability{},
		},
		{
			name: "inline JavaScript",
			body: []byte(`<html><body><script>alert('test');</script></body></html>`),
			headers: http.Header{
				"X-Frame-Options": []string{"DENY"},
			},
			url: "https://example.com",
			expected: []SecurityVulnerability{
				{
					Type:     "xss_risk",
					Severity: "medium",
					Title:    "Inline JavaScript Detected",
				},
			},
		},
		{
			name: "form without CSRF protection",
			body: []byte(`<html><body><form method="post"><input type="text" name="data"></form></body></html>`),
			headers: http.Header{
				"X-Frame-Options": []string{"DENY"},
			},
			url: "https://example.com",
			expected: []SecurityVulnerability{
				{
					Type:     "csrf_risk",
					Severity: "high",
					Title:    "Potential CSRF Vulnerability",
				},
			},
		},
		{
			name: "form with CSRF token",
			body: []byte(`<html><body><form method="post"><input type="hidden" name="csrf_token" value="abc123"><input type="text" name="data"></form></body></html>`),
			headers: http.Header{
				"X-Frame-Options": []string{"DENY"},
			},
			url:      "https://example.com",
			expected: []SecurityVulnerability{}, // Should not detect CSRF vulnerability
		},
		{
			name: "sensitive information exposure",
			body: []byte(`<html><body><script>var api_key = "secret123";</script></body></html>`),
			headers: http.Header{
				"X-Frame-Options": []string{"DENY"},
			},
			url: "https://example.com",
			expected: []SecurityVulnerability{
				{
					Type:     "xss_risk",
					Severity: "medium",
					Title:    "Inline JavaScript Detected",
				},
				{
					Type:     "information_disclosure",
					Severity: "critical",
					Title:    "Sensitive Information Exposure",
				},
			},
		},
		{
			name: "mixed content on HTTPS",
			body: []byte(`<html><body><img src="http://example.com/image.jpg"></body></html>`),
			headers: http.Header{
				"X-Frame-Options": []string{"DENY"},
			},
			url: "https://example.com",
			expected: []SecurityVulnerability{
				{
					Type:     "mixed_content",
					Severity: "medium",
					Title:    "Mixed Content Detected",
				},
			},
		},
		{
			name: "clickjacking risk",
			body: []byte(`<html><body>Content</body></html>`),
			headers: http.Header{}, // No X-Frame-Options or CSP
			url:     "https://example.com",
			expected: []SecurityVulnerability{
				{
					Type:     "clickjacking_risk",
					Severity: "medium",
					Title:    "Clickjacking Protection Missing",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := parseURL(tt.url)
			require.NoError(t, err)
			
			result := analyzer.detectVulnerabilities(tt.body, tt.headers, parsedURL)
			
			assert.Len(t, result, len(tt.expected))
			
			for i, expected := range tt.expected {
				if i < len(result) {
					assert.Equal(t, expected.Type, result[i].Type)
					assert.Equal(t, expected.Severity, result[i].Severity)
					assert.Equal(t, expected.Title, result[i].Title)
				}
			}
		})
	}
}

func TestSecurityAnalyzer_CalculateSecurityScore(t *testing.T) {
	logger := logrus.New()
	analyzer := NewSecurityAnalyzer(logger)

	tests := []struct {
		name            string
		httpsConfig     HTTPSConfig
		headers         SecurityHeadersAnalysis
		vulnerabilities []SecurityVulnerability
		expectedMin     int
		expectedMax     int
	}{
		{
			name: "perfect security",
			httpsConfig: HTTPSConfig{
				IsHTTPS: true,
				CertificateInfo: CertificateDetails{Valid: true},
				HTTPSRedirect: true,
				MixedContent: MixedContentAnalysis{HasMixedContent: false},
			},
			headers: SecurityHeadersAnalysis{
				HSTS:                  HeaderAnalysis{Score: 100},
				ContentSecurityPolicy: HeaderAnalysis{Score: 100},
				XFrameOptions:        HeaderAnalysis{Score: 100},
				XContentTypeOptions:  HeaderAnalysis{Score: 100},
				XSSProtection:        HeaderAnalysis{Score: 100},
				ReferrerPolicy:       HeaderAnalysis{Score: 100},
				PermissionsPolicy:    HeaderAnalysis{Score: 100},
				ExpectCT:             HeaderAnalysis{Score: 100},
			},
			vulnerabilities: []SecurityVulnerability{},
			expectedMin:     95,
			expectedMax:     100,
		},
		{
			name: "poor security",
			httpsConfig: HTTPSConfig{
				IsHTTPS: false,
			},
			headers: SecurityHeadersAnalysis{
				HSTS:                  HeaderAnalysis{Score: 0},
				ContentSecurityPolicy: HeaderAnalysis{Score: 0},
				XFrameOptions:        HeaderAnalysis{Score: 0},
				XContentTypeOptions:  HeaderAnalysis{Score: 0},
				XSSProtection:        HeaderAnalysis{Score: 0},
				ReferrerPolicy:       HeaderAnalysis{Score: 0},
				PermissionsPolicy:    HeaderAnalysis{Score: 0},
				ExpectCT:             HeaderAnalysis{Score: 0},
			},
			vulnerabilities: []SecurityVulnerability{
				{Severity: "critical"},
				{Severity: "high"},
			},
			expectedMin:     0,
			expectedMax:     20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateSecurityScore(tt.httpsConfig, tt.headers, tt.vulnerabilities)
			
			assert.GreaterOrEqual(t, result.OverallScore, tt.expectedMin)
			assert.LessOrEqual(t, result.OverallScore, tt.expectedMax)
			assert.NotEmpty(t, result.Recommendations)
		})
	}
}

// Helper function to parse URL for tests
func parseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}