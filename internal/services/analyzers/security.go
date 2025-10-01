package analyzers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SecurityAnalyzer handles security analysis including HTTP headers and HTTPS configuration
type SecurityAnalyzer struct {
	logger *logrus.Logger
}

// NewSecurityAnalyzer creates a new security analyzer instance
func NewSecurityAnalyzer(logger *logrus.Logger) *SecurityAnalyzer {
	return &SecurityAnalyzer{
		logger: logger,
	}
}

// SecurityAnalysisResult represents the result of security analysis
type SecurityAnalysisResult struct {
	HTTPSConfiguration HTTPSConfig             `json:"https_configuration"`
	SecurityHeaders    SecurityHeadersAnalysis `json:"security_headers"`
	SecurityScore      SecurityScoring         `json:"security_score"`
	Vulnerabilities    []SecurityVulnerability `json:"vulnerabilities"`
	Metadata          SecurityMetadata        `json:"metadata"`
}

// HTTPSConfig contains HTTPS configuration analysis
type HTTPSConfig struct {
	IsHTTPS           bool                `json:"is_https"`
	CertificateInfo   CertificateDetails  `json:"certificate_info"`
	TLSVersion        string              `json:"tls_version"`
	CipherSuite       string              `json:"cipher_suite"`
	HTTPSRedirect     bool                `json:"https_redirect"`
	MixedContent      MixedContentAnalysis `json:"mixed_content"`
}

// CertificateDetails contains SSL certificate information
type CertificateDetails struct {
	Valid         bool      `json:"valid"`
	Issuer        string    `json:"issuer"`
	Subject       string    `json:"subject"`
	ExpiryDate    time.Time `json:"expiry_date"`
	DaysUntilExpiry int     `json:"days_until_expiry"`
	SelfSigned    bool      `json:"self_signed"`
	Wildcard      bool      `json:"wildcard"`
}

// MixedContentAnalysis analyzes mixed content issues
type MixedContentAnalysis struct {
	HasMixedContent bool     `json:"has_mixed_content"`
	HTTPResources   []string `json:"http_resources"`
	Count          int      `json:"count"`
}

// SecurityHeadersAnalysis contains analysis of security headers
type SecurityHeadersAnalysis struct {
	HSTS                    HeaderAnalysis `json:"hsts"`
	ContentSecurityPolicy   HeaderAnalysis `json:"content_security_policy"`
	XFrameOptions          HeaderAnalysis `json:"x_frame_options"`
	XContentTypeOptions    HeaderAnalysis `json:"x_content_type_options"`
	XSSProtection          HeaderAnalysis `json:"x_xss_protection"`
	ReferrerPolicy         HeaderAnalysis `json:"referrer_policy"`
	PermissionsPolicy      HeaderAnalysis `json:"permissions_policy"`
	ExpectCT               HeaderAnalysis `json:"expect_ct"`
}

// HeaderAnalysis represents analysis of a specific security header
type HeaderAnalysis struct {
	Present    bool     `json:"present"`
	Value      string   `json:"value"`
	Score      int      `json:"score"` // 0-100
	Issues     []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

// SecurityScoring contains overall security scoring
type SecurityScoring struct {
	OverallScore    int                    `json:"overall_score"` // 0-100
	HTTPSScore      int                    `json:"https_score"`
	HeadersScore    int                    `json:"headers_score"`
	VulnerabilityScore int                 `json:"vulnerability_score"`
	Recommendations []SecurityRecommendation `json:"recommendations"`
}

// SecurityRecommendation represents a security improvement recommendation
type SecurityRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"` // "critical", "high", "medium", "low"
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// SecurityVulnerability represents a detected security vulnerability
type SecurityVulnerability struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"` // "critical", "high", "medium", "low"
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Remediation string `json:"remediation"`
}

// SecurityMetadata contains analysis metadata
type SecurityMetadata struct {
	AnalysisTime time.Duration `json:"analysis_time_ms"`
	Timestamp    time.Time     `json:"timestamp"`
	UserAgent    string        `json:"user_agent"`
	URL          string        `json:"url"`
}

// Analyze performs comprehensive security analysis
func (sa *SecurityAnalyzer) Analyze(ctx context.Context, targetURL string, headers http.Header, body []byte, userAgent string) (*SecurityAnalysisResult, error) {
	startTime := time.Now()
	
	sa.logger.WithFields(logrus.Fields{
		"url":            targetURL,
		"content_length": len(body),
		"user_agent":     userAgent,
	}).Debug("Starting security analysis")

	// Parse URL for HTTPS analysis
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Analyze HTTPS configuration
	httpsConfig := sa.analyzeHTTPSConfiguration(parsedURL, body)
	
	// Analyze security headers
	securityHeaders := sa.analyzeSecurityHeaders(headers)
	
	// Detect vulnerabilities
	vulnerabilities := sa.detectVulnerabilities(body, headers, parsedURL)
	
	// Calculate security scores
	securityScore := sa.calculateSecurityScore(httpsConfig, securityHeaders, vulnerabilities)

	analysisTime := time.Since(startTime)

	result := &SecurityAnalysisResult{
		HTTPSConfiguration: httpsConfig,
		SecurityHeaders:    securityHeaders,
		SecurityScore:      securityScore,
		Vulnerabilities:    vulnerabilities,
		Metadata: SecurityMetadata{
			AnalysisTime: analysisTime,
			Timestamp:    startTime,
			UserAgent:    userAgent,
			URL:          targetURL,
		},
	}

	sa.logger.WithFields(logrus.Fields{
		"url":                targetURL,
		"overall_score":      securityScore.OverallScore,
		"https_score":        securityScore.HTTPSScore,
		"headers_score":      securityScore.HeadersScore,
		"vulnerabilities":    len(vulnerabilities),
		"analysis_time_ms":   analysisTime.Milliseconds(),
	}).Debug("Security analysis completed")

	return result, nil
}

// analyzeHTTPSConfiguration analyzes HTTPS configuration and certificate details
func (sa *SecurityAnalyzer) analyzeHTTPSConfiguration(parsedURL *url.URL, body []byte) HTTPSConfig {
	isHTTPS := parsedURL.Scheme == "https"
	
	config := HTTPSConfig{
		IsHTTPS: isHTTPS,
		CertificateInfo: CertificateDetails{
			Valid: isHTTPS, // Simplified - in real implementation would check actual cert
		},
		HTTPSRedirect: sa.checkHTTPSRedirect(parsedURL),
		MixedContent:  sa.analyzeMixedContent(body, isHTTPS),
	}

	// If HTTPS, analyze certificate (simplified implementation)
	if isHTTPS {
		config.CertificateInfo = sa.analyzeCertificate(parsedURL.Host)
		config.TLSVersion = "TLS 1.2+" // Simplified
		config.CipherSuite = "Strong" // Simplified
	}

	return config
}

// analyzeCertificate analyzes SSL certificate details (simplified implementation)
func (sa *SecurityAnalyzer) analyzeCertificate(host string) CertificateDetails {
	// In a real implementation, this would make an actual TLS connection
	// and inspect the certificate chain
	return CertificateDetails{
		Valid:           true,
		Issuer:          "Unknown", // Would be extracted from actual cert
		Subject:         host,
		ExpiryDate:      time.Now().AddDate(0, 3, 0), // Placeholder
		DaysUntilExpiry: 90,                          // Placeholder
		SelfSigned:      false,
		Wildcard:        strings.HasPrefix(host, "*."),
	}
}

// checkHTTPSRedirect checks if HTTP redirects to HTTPS
func (sa *SecurityAnalyzer) checkHTTPSRedirect(parsedURL *url.URL) bool {
	// Simplified implementation - would need to make actual HTTP request
	// to check for redirect
	return parsedURL.Scheme == "https"
}

// analyzeMixedContent analyzes mixed content issues
func (sa *SecurityAnalyzer) analyzeMixedContent(body []byte, isHTTPS bool) MixedContentAnalysis {
	if !isHTTPS {
		return MixedContentAnalysis{
			HasMixedContent: false,
			HTTPResources:   []string{},
			Count:          0,
		}
	}

	htmlContent := string(body)
	httpRegex := regexp.MustCompile(`(?i)src=["']http://[^"']*["']|href=["']http://[^"']*["']`)
	matches := httpRegex.FindAllString(htmlContent, -1)

	var httpResources []string
	for _, match := range matches {
		httpResources = append(httpResources, match)
	}

	return MixedContentAnalysis{
		HasMixedContent: len(matches) > 0,
		HTTPResources:   httpResources,
		Count:          len(matches),
	}
}

// analyzeSecurityHeaders analyzes all security-related HTTP headers
func (sa *SecurityAnalyzer) analyzeSecurityHeaders(headers http.Header) SecurityHeadersAnalysis {
	return SecurityHeadersAnalysis{
		HSTS:                  sa.analyzeHSTSHeader(headers),
		ContentSecurityPolicy: sa.analyzeCSPHeader(headers),
		XFrameOptions:        sa.analyzeXFrameOptionsHeader(headers),
		XContentTypeOptions:  sa.analyzeXContentTypeOptionsHeader(headers),
		XSSProtection:        sa.analyzeXSSProtectionHeader(headers),
		ReferrerPolicy:       sa.analyzeReferrerPolicyHeader(headers),
		PermissionsPolicy:    sa.analyzePermissionsPolicyHeader(headers),
		ExpectCT:             sa.analyzeExpectCTHeader(headers),
	}
}

// analyzeHSTSHeader analyzes HTTP Strict Transport Security header
func (sa *SecurityAnalyzer) analyzeHSTSHeader(headers http.Header) HeaderAnalysis {
	hsts := headers.Get("Strict-Transport-Security")
	if hsts == "" {
		return HeaderAnalysis{
			Present: false,
			Score:   0,
			Issues:  []string{"HSTS header not present"},
			Recommendations: []string{
				"Add Strict-Transport-Security header to enforce HTTPS",
				"Include 'includeSubDomains' directive for comprehensive protection",
			},
		}
	}

	score := 70 // Base score for presence
	var issues []string
	var recommendations []string

	// Check for includeSubDomains
	if !strings.Contains(strings.ToLower(hsts), "includesubdomains") {
		score -= 15
		issues = append(issues, "HSTS header missing 'includeSubDomains' directive")
		recommendations = append(recommendations, "Add 'includeSubDomains' to protect all subdomains")
	}

	// Check for preload
	if !strings.Contains(strings.ToLower(hsts), "preload") {
		score -= 10
		recommendations = append(recommendations, "Consider adding 'preload' directive for browser preload lists")
	}

	// Check max-age value
	maxAgeRegex := regexp.MustCompile(`max-age=(\d+)`)
	matches := maxAgeRegex.FindStringSubmatch(hsts)
	if len(matches) > 1 {
		maxAge, err := strconv.Atoi(matches[1])
		if err == nil {
			if maxAge < 31536000 { // Less than 1 year
				score -= 5
				issues = append(issues, "HSTS max-age is less than recommended 1 year")
				recommendations = append(recommendations, "Set max-age to at least 31536000 (1 year)")
			}
		}
	}

	return HeaderAnalysis{
		Present:         true,
		Value:          hsts,
		Score:          score,
		Issues:         issues,
		Recommendations: recommendations,
	}
}

// analyzeCSPHeader analyzes Content Security Policy header
func (sa *SecurityAnalyzer) analyzeCSPHeader(headers http.Header) HeaderAnalysis {
	csp := headers.Get("Content-Security-Policy")
	if csp == "" {
		// Check for report-only version
		csp = headers.Get("Content-Security-Policy-Report-Only")
		if csp == "" {
			return HeaderAnalysis{
				Present: false,
				Score:   0,
				Issues:  []string{"Content Security Policy header not present"},
				Recommendations: []string{
					"Implement Content Security Policy to prevent XSS attacks",
					"Start with report-only mode to test policy",
				},
			}
		}
	}

	score := 60 // Base score for presence
	var issues []string
	var recommendations []string

	// Check for unsafe directives
	if strings.Contains(csp, "'unsafe-inline'") {
		score -= 20
		issues = append(issues, "CSP allows 'unsafe-inline' which reduces security")
		recommendations = append(recommendations, "Remove 'unsafe-inline' and use nonces or hashes")
	}

	if strings.Contains(csp, "'unsafe-eval'") {
		score -= 15
		issues = append(issues, "CSP allows 'unsafe-eval' which can enable code injection")
		recommendations = append(recommendations, "Remove 'unsafe-eval' directive")
	}

	// Check for important directives
	if !strings.Contains(csp, "default-src") {
		score -= 10
		issues = append(issues, "CSP missing 'default-src' directive")
		recommendations = append(recommendations, "Add 'default-src' as fallback policy")
	}

	if !strings.Contains(csp, "script-src") {
		score -= 10
		issues = append(issues, "CSP missing 'script-src' directive")
		recommendations = append(recommendations, "Add 'script-src' to control script execution")
	}

	return HeaderAnalysis{
		Present:         true,
		Value:          csp,
		Score:          score,
		Issues:         issues,
		Recommendations: recommendations,
	}
}

// analyzeXFrameOptionsHeader analyzes X-Frame-Options header
func (sa *SecurityAnalyzer) analyzeXFrameOptionsHeader(headers http.Header) HeaderAnalysis {
	xfo := headers.Get("X-Frame-Options")
	if xfo == "" {
		return HeaderAnalysis{
			Present: false,
			Score:   0,
			Issues:  []string{"X-Frame-Options header not present"},
			Recommendations: []string{
				"Add X-Frame-Options header to prevent clickjacking",
				"Use 'DENY' or 'SAMEORIGIN' value",
			},
		}
	}

	score := 100
	var issues []string
	var recommendations []string

	xfoLower := strings.ToLower(strings.TrimSpace(xfo))
	if xfoLower != "deny" && xfoLower != "sameorigin" && !strings.HasPrefix(xfoLower, "allow-from") {
		score = 50
		issues = append(issues, "X-Frame-Options has invalid value")
		recommendations = append(recommendations, "Use 'DENY', 'SAMEORIGIN', or 'ALLOW-FROM uri'")
	}

	return HeaderAnalysis{
		Present:         true,
		Value:          xfo,
		Score:          score,
		Issues:         issues,
		Recommendations: recommendations,
	}
}

// analyzeXContentTypeOptionsHeader analyzes X-Content-Type-Options header
func (sa *SecurityAnalyzer) analyzeXContentTypeOptionsHeader(headers http.Header) HeaderAnalysis {
	xcto := headers.Get("X-Content-Type-Options")
	if xcto == "" {
		return HeaderAnalysis{
			Present: false,
			Score:   0,
			Issues:  []string{"X-Content-Type-Options header not present"},
			Recommendations: []string{
				"Add X-Content-Type-Options: nosniff to prevent MIME type sniffing",
			},
		}
	}

	score := 100
	var issues []string
	var recommendations []string

	if strings.ToLower(strings.TrimSpace(xcto)) != "nosniff" {
		score = 50
		issues = append(issues, "X-Content-Type-Options should be set to 'nosniff'")
		recommendations = append(recommendations, "Set X-Content-Type-Options to 'nosniff'")
	}

	return HeaderAnalysis{
		Present:         true,
		Value:          xcto,
		Score:          score,
		Issues:         issues,
		Recommendations: recommendations,
	}
}

// analyzeXSSProtectionHeader analyzes X-XSS-Protection header
func (sa *SecurityAnalyzer) analyzeXSSProtectionHeader(headers http.Header) HeaderAnalysis {
	xxp := headers.Get("X-XSS-Protection")
	if xxp == "" {
		return HeaderAnalysis{
			Present: false,
			Score:   0,
			Issues:  []string{"X-XSS-Protection header not present"},
			Recommendations: []string{
				"Add X-XSS-Protection: 1; mode=block for legacy browser protection",
				"Note: Modern browsers rely on CSP instead",
			},
		}
	}

	score := 80
	var issues []string
	var recommendations []string

	xxpLower := strings.ToLower(strings.TrimSpace(xxp))
	if xxpLower == "0" {
		score = 20
		issues = append(issues, "X-XSS-Protection is disabled")
		recommendations = append(recommendations, "Enable X-XSS-Protection or rely on strong CSP")
	} else if !strings.Contains(xxpLower, "mode=block") {
		score = 60
		recommendations = append(recommendations, "Consider adding 'mode=block' for better protection")
	}

	return HeaderAnalysis{
		Present:         true,
		Value:          xxp,
		Score:          score,
		Issues:         issues,
		Recommendations: recommendations,
	}
}

// analyzeReferrerPolicyHeader analyzes Referrer-Policy header
func (sa *SecurityAnalyzer) analyzeReferrerPolicyHeader(headers http.Header) HeaderAnalysis {
	rp := headers.Get("Referrer-Policy")
	if rp == "" {
		return HeaderAnalysis{
			Present: false,
			Score:   0,
			Issues:  []string{"Referrer-Policy header not present"},
			Recommendations: []string{
				"Add Referrer-Policy header to control referrer information",
				"Consider 'strict-origin-when-cross-origin' for balanced privacy",
			},
		}
	}

	score := 70
	var recommendations []string

	rpLower := strings.ToLower(strings.TrimSpace(rp))
	switch rpLower {
	case "no-referrer":
		score = 100
	case "strict-origin-when-cross-origin":
		score = 90
	case "same-origin":
		score = 85
	case "strict-origin":
		score = 80
	default:
		score = 50
		recommendations = append(recommendations, "Consider using a more privacy-focused referrer policy")
	}

	return HeaderAnalysis{
		Present:         true,
		Value:          rp,
		Score:          score,
		Issues:         []string{},
		Recommendations: recommendations,
	}
}

// analyzePermissionsPolicyHeader analyzes Permissions-Policy header
func (sa *SecurityAnalyzer) analyzePermissionsPolicyHeader(headers http.Header) HeaderAnalysis {
	pp := headers.Get("Permissions-Policy")
	if pp == "" {
		// Check for legacy Feature-Policy header
		pp = headers.Get("Feature-Policy")
		if pp == "" {
			return HeaderAnalysis{
				Present: false,
				Score:   0,
				Issues:  []string{"Permissions-Policy header not present"},
				Recommendations: []string{
					"Add Permissions-Policy header to control browser features",
					"Disable unused features like camera, microphone, geolocation",
				},
			}
		}
	}

	score := 80 // Good score for having any permissions policy
	var recommendations []string

	// Basic analysis - in real implementation would parse the policy
	if !strings.Contains(pp, "camera") {
		recommendations = append(recommendations, "Consider explicitly controlling camera access")
	}
	if !strings.Contains(pp, "microphone") {
		recommendations = append(recommendations, "Consider explicitly controlling microphone access")
	}

	return HeaderAnalysis{
		Present:         true,
		Value:          pp,
		Score:          score,
		Issues:         []string{},
		Recommendations: recommendations,
	}
}

// analyzeExpectCTHeader analyzes Expect-CT header
func (sa *SecurityAnalyzer) analyzeExpectCTHeader(headers http.Header) HeaderAnalysis {
	ect := headers.Get("Expect-CT")
	if ect == "" {
		return HeaderAnalysis{
			Present: false,
			Score:   0,
			Issues:  []string{"Expect-CT header not present"},
			Recommendations: []string{
				"Consider adding Expect-CT header for certificate transparency",
				"Note: This header is being deprecated in favor of Certificate Transparency logs",
			},
		}
	}

	score := 70
	var issues []string
	var recommendations []string

	if !strings.Contains(ect, "enforce") {
		score = 50
		recommendations = append(recommendations, "Consider adding 'enforce' directive for stronger protection")
	}

	return HeaderAnalysis{
		Present:         true,
		Value:          ect,
		Score:          score,
		Issues:         issues,
		Recommendations: recommendations,
	}
}

// detectVulnerabilities detects common security vulnerabilities
func (sa *SecurityAnalyzer) detectVulnerabilities(body []byte, headers http.Header, parsedURL *url.URL) []SecurityVulnerability {
	var vulnerabilities []SecurityVulnerability
	htmlContent := string(body)

	// Check for inline JavaScript (potential XSS risk)
	inlineJSRegex := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	if inlineJSRegex.MatchString(htmlContent) {
		vulnerabilities = append(vulnerabilities, SecurityVulnerability{
			Type:        "xss_risk",
			Severity:    "medium",
			Title:       "Inline JavaScript Detected",
			Description: "Inline JavaScript can increase XSS attack surface",
			Location:    "HTML content",
			Remediation: "Move JavaScript to external files and implement CSP",
		})
	}

	// Check for forms without CSRF protection
	formRegex := regexp.MustCompile(`<form[^>]*method=["']post["'][^>]*>`)
	csrfRegex := regexp.MustCompile(`<input[^>]*name=["'][^"']*csrf[^"']*["'][^>]*>|<input[^>]*name=["'][^"']*token[^"']*["'][^>]*>`)
	if formRegex.MatchString(htmlContent) && !csrfRegex.MatchString(htmlContent) {
		vulnerabilities = append(vulnerabilities, SecurityVulnerability{
			Type:        "csrf_risk",
			Severity:    "high",
			Title:       "Potential CSRF Vulnerability",
			Description: "POST forms detected without apparent CSRF protection",
			Location:    "HTML forms",
			Remediation: "Implement CSRF tokens in all state-changing forms",
		})
	}

	// Check for password fields without proper attributes
	passwordRegex := regexp.MustCompile(`<input[^>]*type=["']password["'][^>]*>`)
	if passwordRegex.MatchString(htmlContent) {
		// Check for autocomplete="off" or autocomplete="new-password"
		if !regexp.MustCompile(`autocomplete=["'](?:off|new-password)["']`).MatchString(htmlContent) {
			vulnerabilities = append(vulnerabilities, SecurityVulnerability{
				Type:        "password_security",
				Severity:    "low",
				Title:       "Password Field Security",
				Description: "Password fields should have appropriate autocomplete attributes",
				Location:    "Password input fields",
				Remediation: "Add autocomplete='new-password' or 'current-password' as appropriate",
			})
		}
	}

	// Check for sensitive information exposure
	sensitiveRegex := regexp.MustCompile(`(?i)(api[_-]?key|secret|password|token)\s*[:=]\s*["'][^"']+["']`)
	if sensitiveRegex.MatchString(htmlContent) {
		vulnerabilities = append(vulnerabilities, SecurityVulnerability{
			Type:        "information_disclosure",
			Severity:    "critical",
			Title:       "Sensitive Information Exposure",
			Description: "Potential API keys, secrets, or passwords found in HTML",
			Location:    "HTML content",
			Remediation: "Remove sensitive information from client-side code",
		})
	}

	// Check for HTTP links on HTTPS pages (mixed content)
	if parsedURL.Scheme == "https" {
		httpLinkRegex := regexp.MustCompile(`(?i)(?:src|href)=["']http://[^"']*["']`)
		if httpLinkRegex.MatchString(htmlContent) {
			vulnerabilities = append(vulnerabilities, SecurityVulnerability{
				Type:        "mixed_content",
				Severity:    "medium",
				Title:       "Mixed Content Detected",
				Description: "HTTP resources loaded on HTTPS page",
				Location:    "Resource links",
				Remediation: "Update all resource URLs to use HTTPS",
			})
		}
	}

	// Check for missing security headers (converted to vulnerabilities)
	if headers.Get("X-Frame-Options") == "" && headers.Get("Content-Security-Policy") == "" {
		vulnerabilities = append(vulnerabilities, SecurityVulnerability{
			Type:        "clickjacking_risk",
			Severity:    "medium",
			Title:       "Clickjacking Protection Missing",
			Description: "No X-Frame-Options or CSP frame-ancestors directive found",
			Location:    "HTTP headers",
			Remediation: "Add X-Frame-Options or CSP frame-ancestors directive",
		})
	}

	return vulnerabilities
}

// calculateSecurityScore calculates overall security scoring
func (sa *SecurityAnalyzer) calculateSecurityScore(httpsConfig HTTPSConfig, headers SecurityHeadersAnalysis, vulnerabilities []SecurityVulnerability) SecurityScoring {
	// Calculate HTTPS score
	httpsScore := sa.calculateHTTPSScore(httpsConfig)
	
	// Calculate headers score
	headersScore := sa.calculateHeadersScore(headers)
	
	// Calculate vulnerability score
	vulnerabilityScore := sa.calculateVulnerabilityScore(vulnerabilities)
	
	// Overall score is weighted average
	overallScore := int(float64(httpsScore)*0.4 + float64(headersScore)*0.4 + float64(vulnerabilityScore)*0.2)
	
	// Generate recommendations
	recommendations := sa.generateSecurityRecommendations(httpsConfig, headers, vulnerabilities, overallScore)

	return SecurityScoring{
		OverallScore:       overallScore,
		HTTPSScore:         httpsScore,
		HeadersScore:       headersScore,
		VulnerabilityScore: vulnerabilityScore,
		Recommendations:    recommendations,
	}
}

// calculateHTTPSScore calculates HTTPS configuration score
func (sa *SecurityAnalyzer) calculateHTTPSScore(config HTTPSConfig) int {
	if !config.IsHTTPS {
		return 0
	}

	score := 70 // Base score for HTTPS

	if config.CertificateInfo.Valid {
		score += 20
	}

	if config.HTTPSRedirect {
		score += 5
	}

	if !config.MixedContent.HasMixedContent {
		score += 5
	} else {
		score -= config.MixedContent.Count * 2 // Penalty for mixed content
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// calculateHeadersScore calculates security headers score
func (sa *SecurityAnalyzer) calculateHeadersScore(headers SecurityHeadersAnalysis) int {
	totalScore := 0
	headerCount := 8 // Number of headers we analyze

	totalScore += headers.HSTS.Score
	totalScore += headers.ContentSecurityPolicy.Score
	totalScore += headers.XFrameOptions.Score
	totalScore += headers.XContentTypeOptions.Score
	totalScore += headers.XSSProtection.Score
	totalScore += headers.ReferrerPolicy.Score
	totalScore += headers.PermissionsPolicy.Score
	totalScore += headers.ExpectCT.Score

	return totalScore / headerCount
}

// calculateVulnerabilityScore calculates vulnerability score
func (sa *SecurityAnalyzer) calculateVulnerabilityScore(vulnerabilities []SecurityVulnerability) int {
	if len(vulnerabilities) == 0 {
		return 100
	}

	score := 100
	for _, vuln := range vulnerabilities {
		switch vuln.Severity {
		case "critical":
			score -= 30
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// generateSecurityRecommendations generates prioritized security recommendations
func (sa *SecurityAnalyzer) generateSecurityRecommendations(httpsConfig HTTPSConfig, headers SecurityHeadersAnalysis, vulnerabilities []SecurityVulnerability, overallScore int) []SecurityRecommendation {
	var recommendations []SecurityRecommendation

	// Critical recommendations
	if !httpsConfig.IsHTTPS {
		recommendations = append(recommendations, SecurityRecommendation{
			Type:        "https",
			Priority:    "critical",
			Title:       "Enable HTTPS",
			Description: "Website is not using HTTPS encryption",
			Impact:      "Data transmitted between users and server is not encrypted",
		})
	}

	// High priority recommendations
	if !headers.ContentSecurityPolicy.Present {
		recommendations = append(recommendations, SecurityRecommendation{
			Type:        "csp",
			Priority:    "high",
			Title:       "Implement Content Security Policy",
			Description: "Add CSP header to prevent XSS attacks",
			Impact:      "Reduces risk of cross-site scripting attacks",
		})
	}

	if !headers.HSTS.Present && httpsConfig.IsHTTPS {
		recommendations = append(recommendations, SecurityRecommendation{
			Type:        "hsts",
			Priority:    "high",
			Title:       "Enable HTTP Strict Transport Security",
			Description: "Add HSTS header to enforce HTTPS connections",
			Impact:      "Prevents protocol downgrade attacks",
		})
	}

	// Medium priority recommendations
	if !headers.XFrameOptions.Present {
		recommendations = append(recommendations, SecurityRecommendation{
			Type:        "clickjacking",
			Priority:    "medium",
			Title:       "Add Clickjacking Protection",
			Description: "Implement X-Frame-Options or CSP frame-ancestors",
			Impact:      "Prevents clickjacking attacks",
		})
	}

	if httpsConfig.MixedContent.HasMixedContent {
		recommendations = append(recommendations, SecurityRecommendation{
			Type:        "mixed_content",
			Priority:    "medium",
			Title:       "Fix Mixed Content Issues",
			Description: fmt.Sprintf("Found %d HTTP resources on HTTPS page", httpsConfig.MixedContent.Count),
			Impact:      "Eliminates security warnings and potential vulnerabilities",
		})
	}

	// Low priority recommendations
	if !headers.XContentTypeOptions.Present {
		recommendations = append(recommendations, SecurityRecommendation{
			Type:        "mime_sniffing",
			Priority:    "low",
			Title:       "Prevent MIME Type Sniffing",
			Description: "Add X-Content-Type-Options: nosniff header",
			Impact:      "Prevents MIME type confusion attacks",
		})
	}

	// Add vulnerability-based recommendations
	for _, vuln := range vulnerabilities {
		priority := "medium"
		if vuln.Severity == "critical" || vuln.Severity == "high" {
			priority = "high"
		} else if vuln.Severity == "low" {
			priority = "low"
		}

		recommendations = append(recommendations, SecurityRecommendation{
			Type:        vuln.Type,
			Priority:    priority,
			Title:       fmt.Sprintf("Fix %s", vuln.Title),
			Description: vuln.Description,
			Impact:      vuln.Remediation,
		})
	}

	return recommendations
}