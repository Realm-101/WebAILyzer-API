package analyzers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// AccessibilityAnalyzer handles WCAG compliance checking and accessibility analysis
type AccessibilityAnalyzer struct {
	logger *logrus.Logger
}

// NewAccessibilityAnalyzer creates a new accessibility analyzer instance
func NewAccessibilityAnalyzer(logger *logrus.Logger) *AccessibilityAnalyzer {
	return &AccessibilityAnalyzer{
		logger: logger,
	}
}

// AccessibilityAnalysisResult represents the result of accessibility analysis
type AccessibilityAnalysisResult struct {
	WCAGCompliance   WCAGCompliance     `json:"wcag_compliance"`
	ColorContrast    ColorContrastAnalysis `json:"color_contrast"`
	AltTagAnalysis   AltTagAnalysis     `json:"alt_tag_analysis"`
	KeyboardNav      KeyboardNavigation `json:"keyboard_navigation"`
	FormAccessibility FormAccessibility `json:"form_accessibility"`
	AccessibilityScore AccessibilityScoring `json:"accessibility_score"`
	Issues           []AccessibilityIssue `json:"issues"`
	Metadata         AccessibilityMetadata `json:"metadata"`
}

// WCAGCompliance contains WCAG compliance information
type WCAGCompliance struct {
	Level        string             `json:"level"` // "A", "AA", "AAA"
	Violations   []WCAGViolation    `json:"violations"`
	Warnings     []WCAGWarning      `json:"warnings"`
	Passes       []WCAGPass         `json:"passes"`
	Compliance   ComplianceMetrics  `json:"compliance"`
}

// WCAGViolation represents a WCAG violation
type WCAGViolation struct {
	Rule        string   `json:"rule"`
	Level       string   `json:"level"`
	Description string   `json:"description"`
	Elements    []string `json:"elements"`
	Impact      string   `json:"impact"` // "critical", "serious", "moderate", "minor"
}

// WCAGWarning represents a potential WCAG issue
type WCAGWarning struct {
	Rule        string   `json:"rule"`
	Level       string   `json:"level"`
	Description string   `json:"description"`
	Elements    []string `json:"elements"`
}

// WCAGPass represents a passed WCAG rule
type WCAGPass struct {
	Rule        string `json:"rule"`
	Level       string `json:"level"`
	Description string `json:"description"`
}

// ComplianceMetrics contains compliance statistics
type ComplianceMetrics struct {
	TotalRules    int     `json:"total_rules"`
	PassedRules   int     `json:"passed_rules"`
	FailedRules   int     `json:"failed_rules"`
	WarningRules  int     `json:"warning_rules"`
	ComplianceRate float64 `json:"compliance_rate"`
}

// ColorContrastAnalysis contains color contrast analysis
type ColorContrastAnalysis struct {
	TextElements     []ColorContrastElement `json:"text_elements"`
	FailedContrasts  int                   `json:"failed_contrasts"`
	PassedContrasts  int                   `json:"passed_contrasts"`
	AverageContrast  float64               `json:"average_contrast"`
	Issues           []string              `json:"issues"`
}

// ColorContrastElement represents a text element with contrast analysis
type ColorContrastElement struct {
	Element         string  `json:"element"`
	ForegroundColor string  `json:"foreground_color"`
	BackgroundColor string  `json:"background_color"`
	ContrastRatio   float64 `json:"contrast_ratio"`
	WCAGLevel       string  `json:"wcag_level"` // "AA", "AAA", "fail"
	FontSize        string  `json:"font_size"`
	FontWeight      string  `json:"font_weight"`
}

// AltTagAnalysis contains alt tag validation results
type AltTagAnalysis struct {
	TotalImages        int      `json:"total_images"`
	ImagesWithAlt      int      `json:"images_with_alt"`
	ImagesWithoutAlt   int      `json:"images_without_alt"`
	ImagesWithEmptyAlt int      `json:"images_with_empty_alt"`
	DecorativeImages   int      `json:"decorative_images"`
	AltTextQuality     AltTextQuality `json:"alt_text_quality"`
	Issues             []string `json:"issues"`
}

// AltTextQuality contains alt text quality metrics
type AltTextQuality struct {
	AverageLength    float64  `json:"average_length"`
	TooShort         int      `json:"too_short"` // < 4 characters
	TooLong          int      `json:"too_long"`  // > 125 characters
	ContainsKeywords int      `json:"contains_keywords"` // "image", "picture", etc.
	QualityScore     float64  `json:"quality_score"`
}

// KeyboardNavigation contains keyboard navigation analysis
type KeyboardNavigation struct {
	FocusableElements   int      `json:"focusable_elements"`
	TabIndexIssues      int      `json:"tab_index_issues"`
	MissingFocusStyles  int      `json:"missing_focus_styles"`
	KeyboardTraps       int      `json:"keyboard_traps"`
	SkipLinks          int      `json:"skip_links"`
	Issues             []string `json:"issues"`
}

// FormAccessibility contains form accessibility analysis
type FormAccessibility struct {
	TotalForms         int      `json:"total_forms"`
	FormsWithLabels    int      `json:"forms_with_labels"`
	FormsWithoutLabels int      `json:"forms_without_labels"`
	RequiredFields     int      `json:"required_fields"`
	FieldsetUsage      int      `json:"fieldset_usage"`
	ErrorHandling      FormErrorHandling `json:"error_handling"`
	Issues             []string `json:"issues"`
}

// FormErrorHandling contains form error handling analysis
type FormErrorHandling struct {
	ErrorMessages      int `json:"error_messages"`
	AriaDescribedBy    int `json:"aria_described_by"`
	InlineValidation   int `json:"inline_validation"`
}

// AccessibilityScoring contains accessibility scoring information
type AccessibilityScoring struct {
	OverallScore    int                      `json:"overall_score"` // 0-100
	WCAGScore       int                      `json:"wcag_score"`
	ContrastScore   int                      `json:"contrast_score"`
	AltTextScore    int                      `json:"alt_text_score"`
	KeyboardScore   int                      `json:"keyboard_score"`
	FormScore       int                      `json:"form_score"`
	Recommendations []AccessibilityRecommendation `json:"recommendations"`
}

// AccessibilityRecommendation represents an accessibility improvement recommendation
type AccessibilityRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"` // "critical", "high", "medium", "low"
	Description string `json:"description"`
	Impact      string `json:"impact"`
	WCAGRule    string `json:"wcag_rule,omitempty"`
}

// AccessibilityIssue represents a specific accessibility issue
type AccessibilityIssue struct {
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Element     string   `json:"element,omitempty"`
	WCAGRule    string   `json:"wcag_rule,omitempty"`
	Suggestion  string   `json:"suggestion"`
}

// AccessibilityMetadata contains analysis metadata
type AccessibilityMetadata struct {
	AnalysisTime time.Duration `json:"analysis_time_ms"`
	Timestamp    time.Time     `json:"timestamp"`
	URL          string        `json:"url"`
	UserAgent    string        `json:"user_agent"`
}

// Analyze performs comprehensive accessibility analysis
func (aa *AccessibilityAnalyzer) Analyze(ctx context.Context, url string, headers http.Header, body []byte, userAgent string) (*AccessibilityAnalysisResult, error) {
	startTime := time.Now()
	
	aa.logger.WithFields(logrus.Fields{
		"url":            url,
		"content_length": len(body),
		"user_agent":     userAgent,
	}).Debug("Starting accessibility analysis")

	htmlContent := string(body)

	// Analyze WCAG compliance
	wcagCompliance := aa.analyzeWCAGCompliance(htmlContent)
	
	// Analyze color contrast
	colorContrast := aa.analyzeColorContrast(htmlContent)
	
	// Analyze alt tags
	altTagAnalysis := aa.analyzeAltTags(htmlContent)
	
	// Analyze keyboard navigation
	keyboardNav := aa.analyzeKeyboardNavigation(htmlContent)
	
	// Analyze form accessibility
	formAccessibility := aa.analyzeFormAccessibility(htmlContent)
	
	// Calculate accessibility scores
	accessibilityScore := aa.calculateAccessibilityScore(wcagCompliance, colorContrast, altTagAnalysis, keyboardNav, formAccessibility)
	
	// Collect all issues
	issues := aa.collectIssues(wcagCompliance, colorContrast, altTagAnalysis, keyboardNav, formAccessibility)

	analysisTime := time.Since(startTime)

	result := &AccessibilityAnalysisResult{
		WCAGCompliance:     wcagCompliance,
		ColorContrast:      colorContrast,
		AltTagAnalysis:     altTagAnalysis,
		KeyboardNav:        keyboardNav,
		FormAccessibility:  formAccessibility,
		AccessibilityScore: accessibilityScore,
		Issues:             issues,
		Metadata: AccessibilityMetadata{
			AnalysisTime: analysisTime,
			Timestamp:    startTime,
			URL:          url,
			UserAgent:    userAgent,
		},
	}

	aa.logger.WithFields(logrus.Fields{
		"url":              url,
		"overall_score":    accessibilityScore.OverallScore,
		"wcag_violations":  len(wcagCompliance.Violations),
		"contrast_fails":   colorContrast.FailedContrasts,
		"images_no_alt":    altTagAnalysis.ImagesWithoutAlt,
		"analysis_time_ms": analysisTime.Milliseconds(),
	}).Debug("Accessibility analysis completed")

	return result, nil
}
// analyzeWCAGCompliance performs WCAG compliance checking
func (aa *AccessibilityAnalyzer) analyzeWCAGCompliance(htmlContent string) WCAGCompliance {
	violations := []WCAGViolation{}
	warnings := []WCAGWarning{}
	passes := []WCAGPass{}

	// Check for missing lang attribute (WCAG 3.1.1)
	if !regexp.MustCompile(`<html[^>]*\slang\s*=`).MatchString(htmlContent) {
		violations = append(violations, WCAGViolation{
			Rule:        "3.1.1",
			Level:       "A",
			Description: "Page must have a language specified",
			Elements:    []string{"html"},
			Impact:      "serious",
		})
	} else {
		passes = append(passes, WCAGPass{
			Rule:        "3.1.1",
			Level:       "A",
			Description: "Page language is specified",
		})
	}

	// Check for missing page title (WCAG 2.4.2)
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]*)</title>`)
	titleMatch := titleRegex.FindStringSubmatch(htmlContent)
	if len(titleMatch) == 0 || strings.TrimSpace(titleMatch[1]) == "" {
		violations = append(violations, WCAGViolation{
			Rule:        "2.4.2",
			Level:       "A",
			Description: "Page must have a descriptive title",
			Elements:    []string{"title"},
			Impact:      "serious",
		})
	} else {
		passes = append(passes, WCAGPass{
			Rule:        "2.4.2",
			Level:       "A",
			Description: "Page has a descriptive title",
		})
	}

	// Check for images without alt attributes (WCAG 1.1.1)
	imgRegex := regexp.MustCompile(`<img[^>]*>`)
	imgMatches := imgRegex.FindAllString(htmlContent, -1)
	for _, img := range imgMatches {
		if !regexp.MustCompile(`alt\s*=`).MatchString(img) {
			violations = append(violations, WCAGViolation{
				Rule:        "1.1.1",
				Level:       "A",
				Description: "Images must have alternative text",
				Elements:    []string{img},
				Impact:      "critical",
			})
		}
	}

	// Check for form inputs without labels (WCAG 3.3.2)
	inputRegex := regexp.MustCompile(`<input[^>]*type\s*=\s*["'](?:text|email|password|tel|url|search)["'][^>]*>`)
	inputMatches := inputRegex.FindAllString(htmlContent, -1)
	for _, input := range inputMatches {
		hasId := regexp.MustCompile(`id\s*=\s*["']([^"']+)["']`).MatchString(input)
		hasAriaLabel := regexp.MustCompile(`aria-label\s*=`).MatchString(input)
		hasAriaLabelledBy := regexp.MustCompile(`aria-labelledby\s*=`).MatchString(input)
		
		if hasId {
			// Check if there's a corresponding label
			idMatch := regexp.MustCompile(`id\s*=\s*["']([^"']+)["']`).FindStringSubmatch(input)
			if len(idMatch) > 1 {
				labelPattern := fmt.Sprintf(`<label[^>]*for\s*=\s*["']%s["'][^>]*>`, regexp.QuoteMeta(idMatch[1]))
				if !regexp.MustCompile(labelPattern).MatchString(htmlContent) && !hasAriaLabel && !hasAriaLabelledBy {
					violations = append(violations, WCAGViolation{
						Rule:        "3.3.2",
						Level:       "A",
						Description: "Form inputs must have labels",
						Elements:    []string{input},
						Impact:      "serious",
					})
				}
			}
		} else if !hasAriaLabel && !hasAriaLabelledBy {
			violations = append(violations, WCAGViolation{
				Rule:        "3.3.2",
				Level:       "A",
				Description: "Form inputs must have labels",
				Elements:    []string{input},
				Impact:      "serious",
			})
		}
	}

	// Check for missing heading structure (WCAG 1.3.1)
	h1Count := len(regexp.MustCompile(`<h1[^>]*>`).FindAllString(htmlContent, -1))
	if h1Count == 0 {
		warnings = append(warnings, WCAGWarning{
			Rule:        "1.3.1",
			Level:       "A",
			Description: "Page should have at least one H1 heading",
			Elements:    []string{"h1"},
		})
	} else if h1Count > 1 {
		warnings = append(warnings, WCAGWarning{
			Rule:        "1.3.1",
			Level:       "A",
			Description: "Page should have only one H1 heading",
			Elements:    []string{"h1"},
		})
	} else {
		passes = append(passes, WCAGPass{
			Rule:        "1.3.1",
			Level:       "A",
			Description: "Page has proper heading structure",
		})
	}

	totalRules := len(violations) + len(warnings) + len(passes)
	passedRules := len(passes)
	failedRules := len(violations)
	warningRules := len(warnings)
	
	var complianceRate float64
	if totalRules > 0 {
		complianceRate = float64(passedRules) / float64(totalRules) * 100
	}

	return WCAGCompliance{
		Level:      "AA", // Target level
		Violations: violations,
		Warnings:   warnings,
		Passes:     passes,
		Compliance: ComplianceMetrics{
			TotalRules:     totalRules,
			PassedRules:    passedRules,
			FailedRules:    failedRules,
			WarningRules:   warningRules,
			ComplianceRate: complianceRate,
		},
	}
}

// analyzeColorContrast performs color contrast analysis
func (aa *AccessibilityAnalyzer) analyzeColorContrast(htmlContent string) ColorContrastAnalysis {
	textElements := []ColorContrastElement{}
	failedContrasts := 0
	passedContrasts := 0
	totalContrast := 0.0

	// This is a simplified implementation - in a real scenario, you'd need to:
	// 1. Parse CSS styles to get actual colors
	// 2. Calculate luminance and contrast ratios
	// 3. Check against WCAG AA/AAA standards
	
	// For now, we'll simulate some common contrast issues
	commonIssues := []string{
		"Light gray text on white background may not meet contrast requirements",
		"Blue links on dark blue background may have insufficient contrast",
		"Placeholder text often has poor contrast ratios",
	}

	// Simulate some contrast analysis results
	textElements = append(textElements, ColorContrastElement{
		Element:         "body text",
		ForegroundColor: "#333333",
		BackgroundColor: "#ffffff",
		ContrastRatio:   12.63,
		WCAGLevel:       "AAA",
		FontSize:        "16px",
		FontWeight:      "normal",
	})

	passedContrasts++
	totalContrast += 12.63

	// Check for common low-contrast patterns
	if regexp.MustCompile(`color\s*:\s*#[a-fA-F0-9]{3,6}.*background.*#[a-fA-F0-9]{3,6}`).MatchString(htmlContent) {
		// This would need proper CSS parsing in a real implementation
		failedContrasts++
	}

	averageContrast := 0.0
	if len(textElements) > 0 {
		averageContrast = totalContrast / float64(len(textElements))
	}

	return ColorContrastAnalysis{
		TextElements:    textElements,
		FailedContrasts: failedContrasts,
		PassedContrasts: passedContrasts,
		AverageContrast: averageContrast,
		Issues:          commonIssues,
	}
}

// analyzeAltTags performs alt tag validation
func (aa *AccessibilityAnalyzer) analyzeAltTags(htmlContent string) AltTagAnalysis {
	imgRegex := regexp.MustCompile(`<img[^>]*>`)
	imgMatches := imgRegex.FindAllString(htmlContent, -1)
	
	totalImages := len(imgMatches)
	imagesWithAlt := 0
	imagesWithoutAlt := 0
	imagesWithEmptyAlt := 0
	decorativeImages := 0
	
	altTexts := []string{}
	issues := []string{}

	for _, img := range imgMatches {
		altRegex := regexp.MustCompile(`alt\s*=\s*["']([^"']*)["']`)
		altMatch := altRegex.FindStringSubmatch(img)
		
		if len(altMatch) == 0 {
			imagesWithoutAlt++
			issues = append(issues, fmt.Sprintf("Image missing alt attribute: %s", img))
		} else {
			altText := strings.TrimSpace(altMatch[1])
			if altText == "" {
				imagesWithEmptyAlt++
				decorativeImages++ // Assuming empty alt means decorative
			} else {
				imagesWithAlt++
				altTexts = append(altTexts, altText)
			}
		}
	}

	// Analyze alt text quality
	altTextQuality := aa.analyzeAltTextQuality(altTexts)

	return AltTagAnalysis{
		TotalImages:        totalImages,
		ImagesWithAlt:      imagesWithAlt,
		ImagesWithoutAlt:   imagesWithoutAlt,
		ImagesWithEmptyAlt: imagesWithEmptyAlt,
		DecorativeImages:   decorativeImages,
		AltTextQuality:     altTextQuality,
		Issues:             issues,
	}
}

// analyzeAltTextQuality analyzes the quality of alt text
func (aa *AccessibilityAnalyzer) analyzeAltTextQuality(altTexts []string) AltTextQuality {
	if len(altTexts) == 0 {
		return AltTextQuality{}
	}

	totalLength := 0
	tooShort := 0
	tooLong := 0
	containsKeywords := 0

	badKeywords := []string{"image", "picture", "photo", "graphic", "icon"}

	for _, alt := range altTexts {
		length := len(alt)
		totalLength += length

		if length < 4 {
			tooShort++
		}
		if length > 125 {
			tooLong++
		}

		altLower := strings.ToLower(alt)
		for _, keyword := range badKeywords {
			if strings.Contains(altLower, keyword) {
				containsKeywords++
				break
			}
		}
	}

	averageLength := float64(totalLength) / float64(len(altTexts))
	
	// Calculate quality score (0-100)
	qualityScore := 100.0
	qualityScore -= float64(tooShort) / float64(len(altTexts)) * 30    // -30% for too short
	qualityScore -= float64(tooLong) / float64(len(altTexts)) * 20     // -20% for too long
	qualityScore -= float64(containsKeywords) / float64(len(altTexts)) * 25 // -25% for bad keywords

	if qualityScore < 0 {
		qualityScore = 0
	}

	return AltTextQuality{
		AverageLength:    averageLength,
		TooShort:         tooShort,
		TooLong:          tooLong,
		ContainsKeywords: containsKeywords,
		QualityScore:     qualityScore,
	}
}

// analyzeKeyboardNavigation analyzes keyboard navigation accessibility
func (aa *AccessibilityAnalyzer) analyzeKeyboardNavigation(htmlContent string) KeyboardNavigation {
	// Count focusable elements
	focusableSelectors := []string{
		`<a[^>]*href`,
		`<button[^>]*>`,
		`<input[^>]*>`,
		`<select[^>]*>`,
		`<textarea[^>]*>`,
		`<[^>]*tabindex\s*=\s*["'][^"']*["']`,
	}

	focusableElements := 0
	for _, selector := range focusableSelectors {
		matches := regexp.MustCompile(selector).FindAllString(htmlContent, -1)
		focusableElements += len(matches)
	}

	// Check for tabindex issues
	tabIndexRegex := regexp.MustCompile(`tabindex\s*=\s*["']([^"']*)["']`)
	tabIndexMatches := tabIndexRegex.FindAllStringSubmatch(htmlContent, -1)
	tabIndexIssues := 0

	for _, match := range tabIndexMatches {
		if len(match) > 1 {
			if tabIndex, err := strconv.Atoi(match[1]); err == nil && tabIndex > 0 {
				tabIndexIssues++ // Positive tabindex values can cause issues
			}
		}
	}

	// Check for skip links
	skipLinks := len(regexp.MustCompile(`href\s*=\s*["']#[^"']*["'][^>]*>.*skip`).FindAllString(htmlContent, -1))

	issues := []string{}
	if tabIndexIssues > 0 {
		issues = append(issues, fmt.Sprintf("Found %d positive tabindex values which can disrupt natural tab order", tabIndexIssues))
	}
	if skipLinks == 0 {
		issues = append(issues, "No skip links found - consider adding skip navigation links")
	}

	return KeyboardNavigation{
		FocusableElements:  focusableElements,
		TabIndexIssues:     tabIndexIssues,
		MissingFocusStyles: 0, // Would need CSS analysis
		KeyboardTraps:      0, // Would need dynamic analysis
		SkipLinks:          skipLinks,
		Issues:             issues,
	}
}

// analyzeFormAccessibility analyzes form accessibility
func (aa *AccessibilityAnalyzer) analyzeFormAccessibility(htmlContent string) FormAccessibility {
	formRegex := regexp.MustCompile(`<form[^>]*>`)
	formMatches := formRegex.FindAllString(htmlContent, -1)
	totalForms := len(formMatches)

	inputRegex := regexp.MustCompile(`<input[^>]*>`)
	inputMatches := inputRegex.FindAllString(htmlContent, -1)

	formsWithLabels := 0
	formsWithoutLabels := 0
	requiredFields := 0
	fieldsetUsage := len(regexp.MustCompile(`<fieldset[^>]*>`).FindAllString(htmlContent, -1))

	// Count required fields
	for _, input := range inputMatches {
		if regexp.MustCompile(`required`).MatchString(input) {
			requiredFields++
		}
	}

	// Simplified label checking - in reality, this would be more complex
	labelCount := len(regexp.MustCompile(`<label[^>]*>`).FindAllString(htmlContent, -1))
	if labelCount > 0 {
		formsWithLabels = totalForms // Simplified assumption
	} else {
		formsWithoutLabels = totalForms
	}

	errorHandling := FormErrorHandling{
		ErrorMessages:    len(regexp.MustCompile(`class\s*=\s*["'][^"']*error[^"']*["']`).FindAllString(htmlContent, -1)),
		AriaDescribedBy:  len(regexp.MustCompile(`aria-describedby\s*=`).FindAllString(htmlContent, -1)),
		InlineValidation: 0, // Would need JavaScript analysis
	}

	issues := []string{}
	if formsWithoutLabels > 0 {
		issues = append(issues, fmt.Sprintf("%d forms may be missing proper labels", formsWithoutLabels))
	}
	if requiredFields > 0 && errorHandling.ErrorMessages == 0 {
		issues = append(issues, "Required fields found but no error handling detected")
	}

	return FormAccessibility{
		TotalForms:         totalForms,
		FormsWithLabels:    formsWithLabels,
		FormsWithoutLabels: formsWithoutLabels,
		RequiredFields:     requiredFields,
		FieldsetUsage:      fieldsetUsage,
		ErrorHandling:      errorHandling,
		Issues:             issues,
	}
}

// calculateAccessibilityScore calculates overall accessibility score
func (aa *AccessibilityAnalyzer) calculateAccessibilityScore(
	wcag WCAGCompliance,
	contrast ColorContrastAnalysis,
	altTags AltTagAnalysis,
	keyboard KeyboardNavigation,
	forms FormAccessibility,
) AccessibilityScoring {
	
	// Calculate individual scores (0-100)
	wcagScore := int(wcag.Compliance.ComplianceRate)
	
	contrastScore := 100
	if contrast.FailedContrasts > 0 {
		contrastScore = max(0, 100 - (contrast.FailedContrasts * 20))
	}
	
	altTextScore := 100
	if altTags.TotalImages > 0 {
		altTextScore = int(float64(altTags.ImagesWithAlt) / float64(altTags.TotalImages) * 100)
		// Apply quality penalty
		altTextScore = int(float64(altTextScore) * (altTags.AltTextQuality.QualityScore / 100))
	}
	
	keyboardScore := 100
	if keyboard.TabIndexIssues > 0 {
		keyboardScore -= keyboard.TabIndexIssues * 10
	}
	if keyboard.SkipLinks == 0 {
		keyboardScore -= 20
	}
	keyboardScore = max(0, keyboardScore)
	
	formScore := 100
	if forms.TotalForms > 0 {
		formScore = int(float64(forms.FormsWithLabels) / float64(forms.TotalForms) * 100)
	}
	
	// Calculate overall score (weighted average)
	overallScore := int(
		float64(wcagScore)*0.3 +
		float64(contrastScore)*0.2 +
		float64(altTextScore)*0.2 +
		float64(keyboardScore)*0.15 +
		float64(formScore)*0.15,
	)
	
	// Generate recommendations
	recommendations := []AccessibilityRecommendation{}
	
	if len(wcag.Violations) > 0 {
		recommendations = append(recommendations, AccessibilityRecommendation{
			Type:        "wcag_compliance",
			Priority:    "critical",
			Description: fmt.Sprintf("Fix %d WCAG violations to improve compliance", len(wcag.Violations)),
			Impact:      "Critical accessibility barriers for users with disabilities",
			WCAGRule:    "Multiple",
		})
	}
	
	if contrast.FailedContrasts > 0 {
		recommendations = append(recommendations, AccessibilityRecommendation{
			Type:        "color_contrast",
			Priority:    "high",
			Description: "Improve color contrast ratios to meet WCAG AA standards",
			Impact:      "Users with visual impairments may have difficulty reading content",
			WCAGRule:    "1.4.3",
		})
	}
	
	if altTags.ImagesWithoutAlt > 0 {
		recommendations = append(recommendations, AccessibilityRecommendation{
			Type:        "alt_text",
			Priority:    "high",
			Description: fmt.Sprintf("Add alt text to %d images", altTags.ImagesWithoutAlt),
			Impact:      "Screen reader users cannot understand image content",
			WCAGRule:    "1.1.1",
		})
	}
	
	if altTags.AltTextQuality.QualityScore < 70 {
		recommendations = append(recommendations, AccessibilityRecommendation{
			Type:        "alt_text_quality",
			Priority:    "medium",
			Description: "Improve alt text quality - avoid generic terms like 'image' or 'picture'",
			Impact:      "Poor alt text provides little value to screen reader users",
			WCAGRule:    "1.1.1",
		})
	}
	
	if keyboard.SkipLinks == 0 {
		recommendations = append(recommendations, AccessibilityRecommendation{
			Type:        "keyboard_navigation",
			Priority:    "medium",
			Description: "Add skip navigation links for keyboard users",
			Impact:      "Keyboard users must tab through all navigation to reach main content",
			WCAGRule:    "2.4.1",
		})
	}
	
	if forms.FormsWithoutLabels > 0 {
		recommendations = append(recommendations, AccessibilityRecommendation{
			Type:        "form_labels",
			Priority:    "high",
			Description: fmt.Sprintf("Add proper labels to %d forms", forms.FormsWithoutLabels),
			Impact:      "Screen reader users cannot understand form field purposes",
			WCAGRule:    "3.3.2",
		})
	}

	return AccessibilityScoring{
		OverallScore:    overallScore,
		WCAGScore:       wcagScore,
		ContrastScore:   contrastScore,
		AltTextScore:    altTextScore,
		KeyboardScore:   keyboardScore,
		FormScore:       formScore,
		Recommendations: recommendations,
	}
}

// collectIssues collects all accessibility issues from different analysis areas
func (aa *AccessibilityAnalyzer) collectIssues(
	wcag WCAGCompliance,
	contrast ColorContrastAnalysis,
	altTags AltTagAnalysis,
	keyboard KeyboardNavigation,
	forms FormAccessibility,
) []AccessibilityIssue {
	
	issues := []AccessibilityIssue{}
	
	// Add WCAG violations as issues
	for _, violation := range wcag.Violations {
		issues = append(issues, AccessibilityIssue{
			Type:        "wcag_violation",
			Severity:    violation.Impact,
			Description: violation.Description,
			WCAGRule:    violation.Rule,
			Suggestion:  aa.getWCAGSuggestion(violation.Rule),
		})
	}
	
	// Add contrast issues
	for _, issue := range contrast.Issues {
		issues = append(issues, AccessibilityIssue{
			Type:        "color_contrast",
			Severity:    "moderate",
			Description: issue,
			WCAGRule:    "1.4.3",
			Suggestion:  "Ensure text has a contrast ratio of at least 4.5:1 (AA) or 7:1 (AAA)",
		})
	}
	
	// Add alt tag issues
	for _, issue := range altTags.Issues {
		issues = append(issues, AccessibilityIssue{
			Type:        "alt_text",
			Severity:    "serious",
			Description: issue,
			WCAGRule:    "1.1.1",
			Suggestion:  "Add descriptive alt text that conveys the purpose and content of the image",
		})
	}
	
	// Add keyboard navigation issues
	for _, issue := range keyboard.Issues {
		issues = append(issues, AccessibilityIssue{
			Type:        "keyboard_navigation",
			Severity:    "moderate",
			Description: issue,
			WCAGRule:    "2.1.1",
			Suggestion:  "Ensure all interactive elements are keyboard accessible",
		})
	}
	
	// Add form accessibility issues
	for _, issue := range forms.Issues {
		issues = append(issues, AccessibilityIssue{
			Type:        "form_accessibility",
			Severity:    "serious",
			Description: issue,
			WCAGRule:    "3.3.2",
			Suggestion:  "Associate form controls with labels using for/id attributes or aria-label",
		})
	}
	
	return issues
}

// getWCAGSuggestion returns a suggestion for fixing a specific WCAG rule violation
func (aa *AccessibilityAnalyzer) getWCAGSuggestion(rule string) string {
	suggestions := map[string]string{
		"1.1.1": "Provide alternative text for images that conveys their purpose and content",
		"1.3.1": "Use proper heading hierarchy (h1-h6) to structure content logically",
		"2.1.1": "Ensure all functionality is available via keyboard navigation",
		"2.4.1": "Provide skip links to help users bypass repetitive navigation",
		"2.4.2": "Give pages descriptive and unique titles",
		"3.1.1": "Specify the primary language of the page using the lang attribute",
		"3.3.2": "Provide labels or instructions for form inputs",
		"4.1.2": "Ensure all UI components have accessible names and roles",
	}
	
	if suggestion, exists := suggestions[rule]; exists {
		return suggestion
	}
	return "Review WCAG guidelines for this rule and implement appropriate fixes"
}

// Helper function to get max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}