package analyzers

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccessibilityAnalyzer(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	assert.NotNil(t, analyzer)
	assert.Equal(t, logger, analyzer.logger)
}

func TestAccessibilityAnalyzer_Analyze(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	testHTML := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<title>Test Page</title>
	</head>
	<body>
		<h1>Main Heading</h1>
		<p>Some content</p>
		<img src="test.jpg" alt="Test image">
		<form>
			<label for="email">Email:</label>
			<input type="email" id="email" required>
		</form>
	</body>
	</html>`
	
	ctx := context.Background()
	headers := http.Header{}
	result, err := analyzer.Analyze(ctx, "https://example.com", headers, []byte(testHTML), "test-agent")
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "https://example.com", result.Metadata.URL)
	assert.Equal(t, "test-agent", result.Metadata.UserAgent)
	assert.True(t, result.AccessibilityScore.OverallScore >= 0)
	assert.True(t, result.AccessibilityScore.OverallScore <= 100)
}

func TestAccessibilityAnalyzer_analyzeWCAGCompliance(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	tests := []struct {
		name           string
		html           string
		expectedPasses int
		expectedViolations int
		expectedWarnings int
	}{
		{
			name: "compliant HTML",
			html: `<!DOCTYPE html>
			<html lang="en">
			<head><title>Test Page</title></head>
			<body>
				<h1>Main Heading</h1>
				<img src="test.jpg" alt="Test image">
				<form>
					<label for="email">Email:</label>
					<input type="email" id="email">
				</form>
			</body>
			</html>`,
			expectedPasses: 3, // lang, title, heading structure
			expectedViolations: 0,
			expectedWarnings: 0,
		},
		{
			name: "missing lang attribute",
			html: `<!DOCTYPE html>
			<html>
			<head><title>Test Page</title></head>
			<body><h1>Test</h1></body>
			</html>`,
			expectedPasses: 2, // title, heading structure
			expectedViolations: 1, // missing lang
			expectedWarnings: 0,
		},
		{
			name: "missing title",
			html: `<!DOCTYPE html>
			<html lang="en">
			<head></head>
			<body><h1>Test</h1></body>
			</html>`,
			expectedPasses: 2, // lang, heading structure
			expectedViolations: 1, // missing title
			expectedWarnings: 0,
		},
		{
			name: "image without alt",
			html: `<!DOCTYPE html>
			<html lang="en">
			<head><title>Test</title></head>
			<body>
				<h1>Test</h1>
				<img src="test.jpg">
			</body>
			</html>`,
			expectedPasses: 2, // lang, title, heading structure
			expectedViolations: 1, // image without alt
			expectedWarnings: 0,
		},
		{
			name: "multiple h1 tags",
			html: `<!DOCTYPE html>
			<html lang="en">
			<head><title>Test</title></head>
			<body>
				<h1>First Heading</h1>
				<h1>Second Heading</h1>
			</body>
			</html>`,
			expectedPasses: 2, // lang, title
			expectedViolations: 0,
			expectedWarnings: 1, // multiple h1
		},
		{
			name: "no h1 tag",
			html: `<!DOCTYPE html>
			<html lang="en">
			<head><title>Test</title></head>
			<body>
				<h2>Subheading</h2>
			</body>
			</html>`,
			expectedPasses: 2, // lang, title
			expectedViolations: 0,
			expectedWarnings: 1, // no h1
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeWCAGCompliance(tt.html)
			
			assert.Equal(t, tt.expectedPasses, len(result.Passes), "Unexpected number of passes")
			assert.Equal(t, tt.expectedViolations, len(result.Violations), "Unexpected number of violations")
			assert.Equal(t, tt.expectedWarnings, len(result.Warnings), "Unexpected number of warnings")
			assert.Equal(t, "AA", result.Level)
			
			totalRules := len(result.Passes) + len(result.Violations) + len(result.Warnings)
			assert.Equal(t, totalRules, result.Compliance.TotalRules)
			assert.Equal(t, len(result.Passes), result.Compliance.PassedRules)
			assert.Equal(t, len(result.Violations), result.Compliance.FailedRules)
			assert.Equal(t, len(result.Warnings), result.Compliance.WarningRules)
			
			if totalRules > 0 {
				expectedRate := float64(len(result.Passes)) / float64(totalRules) * 100
				assert.Equal(t, expectedRate, result.Compliance.ComplianceRate)
			}
		})
	}
}

func TestAccessibilityAnalyzer_analyzeAltTags(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	tests := []struct {
		name                   string
		html                   string
		expectedTotal          int
		expectedWithAlt        int
		expectedWithoutAlt     int
		expectedWithEmptyAlt   int
		expectedDecorative     int
	}{
		{
			name: "images with proper alt text",
			html: `<img src="test1.jpg" alt="A beautiful sunset">
			       <img src="test2.jpg" alt="Company logo">`,
			expectedTotal:        2,
			expectedWithAlt:      2,
			expectedWithoutAlt:   0,
			expectedWithEmptyAlt: 0,
			expectedDecorative:   0,
		},
		{
			name: "images without alt attributes",
			html: `<img src="test1.jpg">
			       <img src="test2.jpg">`,
			expectedTotal:        2,
			expectedWithAlt:      0,
			expectedWithoutAlt:   2,
			expectedWithEmptyAlt: 0,
			expectedDecorative:   0,
		},
		{
			name: "decorative images with empty alt",
			html: `<img src="decoration.jpg" alt="">
			       <img src="spacer.gif" alt="">`,
			expectedTotal:        2,
			expectedWithAlt:      0,
			expectedWithoutAlt:   0,
			expectedWithEmptyAlt: 2,
			expectedDecorative:   2,
		},
		{
			name: "mixed image types",
			html: `<img src="content.jpg" alt="Important content">
			       <img src="decoration.jpg" alt="">
			       <img src="missing.jpg">`,
			expectedTotal:        3,
			expectedWithAlt:      1,
			expectedWithoutAlt:   1,
			expectedWithEmptyAlt: 1,
			expectedDecorative:   1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeAltTags(tt.html)
			
			assert.Equal(t, tt.expectedTotal, result.TotalImages)
			assert.Equal(t, tt.expectedWithAlt, result.ImagesWithAlt)
			assert.Equal(t, tt.expectedWithoutAlt, result.ImagesWithoutAlt)
			assert.Equal(t, tt.expectedWithEmptyAlt, result.ImagesWithEmptyAlt)
			assert.Equal(t, tt.expectedDecorative, result.DecorativeImages)
			
			// Check that issues are reported for images without alt
			if tt.expectedWithoutAlt > 0 {
				assert.True(t, len(result.Issues) > 0, "Expected issues for images without alt")
			}
		})
	}
}

func TestAccessibilityAnalyzer_analyzeAltTextQuality(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	tests := []struct {
		name                     string
		altTexts                 []string
		expectedTooShort         int
		expectedTooLong          int
		expectedContainsKeywords int
		minQualityScore          float64
	}{
		{
			name:                     "good quality alt texts",
			altTexts:                 []string{"A beautiful sunset over the mountains", "Company logo with blue background"},
			expectedTooShort:         0,
			expectedTooLong:          0,
			expectedContainsKeywords: 0,
			minQualityScore:          90,
		},
		{
			name:                     "poor quality alt texts",
			altTexts:                 []string{"pic", "image of something", "a very long description that goes on and on and on and exceeds the recommended character limit for alt text which should be concise"},
			expectedTooShort:         1,
			expectedTooLong:          1,
			expectedContainsKeywords: 1,
			minQualityScore:          0,
		},
		{
			name:                     "mixed quality",
			altTexts:                 []string{"Good description", "img", "Another good description"},
			expectedTooShort:         1,
			expectedTooLong:          0,
			expectedContainsKeywords: 0,
			minQualityScore:          50,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeAltTextQuality(tt.altTexts)
			
			assert.Equal(t, tt.expectedTooShort, result.TooShort)
			assert.Equal(t, tt.expectedTooLong, result.TooLong)
			assert.Equal(t, tt.expectedContainsKeywords, result.ContainsKeywords)
			assert.True(t, result.QualityScore >= tt.minQualityScore)
			
			if len(tt.altTexts) > 0 {
				totalLength := 0
				for _, alt := range tt.altTexts {
					totalLength += len(alt)
				}
				expectedAverage := float64(totalLength) / float64(len(tt.altTexts))
				assert.Equal(t, expectedAverage, result.AverageLength)
			}
		})
	}
}

func TestAccessibilityAnalyzer_analyzeKeyboardNavigation(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	tests := []struct {
		name                      string
		html                      string
		expectedFocusableElements int
		expectedTabIndexIssues    int
		expectedSkipLinks         int
	}{
		{
			name: "good keyboard navigation",
			html: `<a href="#main">Skip to main content</a>
			       <a href="/page1">Link 1</a>
			       <button>Button</button>
			       <input type="text">`,
			expectedFocusableElements: 4,
			expectedTabIndexIssues:    0,
			expectedSkipLinks:         1,
		},
		{
			name: "tabindex issues",
			html: `<div tabindex="1">Focusable div</div>
			       <div tabindex="5">Another focusable div</div>
			       <button tabindex="0">Good button</button>`,
			expectedFocusableElements: 3,
			expectedTabIndexIssues:    2, // positive tabindex values
			expectedSkipLinks:         0,
		},
		{
			name: "no focusable elements",
			html: `<div>Just text</div>
			       <span>More text</span>`,
			expectedFocusableElements: 0,
			expectedTabIndexIssues:    0,
			expectedSkipLinks:         0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeKeyboardNavigation(tt.html)
			
			assert.Equal(t, tt.expectedFocusableElements, result.FocusableElements)
			assert.Equal(t, tt.expectedTabIndexIssues, result.TabIndexIssues)
			assert.Equal(t, tt.expectedSkipLinks, result.SkipLinks)
			
			// Check for appropriate issues
			if tt.expectedTabIndexIssues > 0 {
				found := false
				for _, issue := range result.Issues {
					if strings.Contains(issue, "tabindex") {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected tabindex issue to be reported")
			}
			
			if tt.expectedSkipLinks == 0 {
				found := false
				for _, issue := range result.Issues {
					if strings.Contains(issue, "skip") {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected skip link issue to be reported")
			}
		})
	}
}

func TestAccessibilityAnalyzer_analyzeFormAccessibility(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	tests := []struct {
		name                       string
		html                       string
		expectedTotalForms         int
		expectedFormsWithLabels    int
		expectedFormsWithoutLabels int
		expectedRequiredFields     int
		expectedFieldsetUsage      int
	}{
		{
			name: "accessible form",
			html: `<form>
			         <fieldset>
			           <legend>Personal Information</legend>
			           <label for="name">Name:</label>
			           <input type="text" id="name" required>
			           <label for="email">Email:</label>
			           <input type="email" id="email" required>
			         </fieldset>
			       </form>`,
			expectedTotalForms:         1,
			expectedFormsWithLabels:    1,
			expectedFormsWithoutLabels: 0,
			expectedRequiredFields:     2,
			expectedFieldsetUsage:      1,
		},
		{
			name: "inaccessible form",
			html: `<form>
			         <input type="text" placeholder="Enter name">
			         <input type="email" placeholder="Enter email" required>
			       </form>`,
			expectedTotalForms:         1,
			expectedFormsWithLabels:    0,
			expectedFormsWithoutLabels: 1,
			expectedRequiredFields:     1,
			expectedFieldsetUsage:      0,
		},
		{
			name: "no forms",
			html: `<div>No forms here</div>`,
			expectedTotalForms:         0,
			expectedFormsWithLabels:    0,
			expectedFormsWithoutLabels: 0,
			expectedRequiredFields:     0,
			expectedFieldsetUsage:      0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeFormAccessibility(tt.html)
			
			assert.Equal(t, tt.expectedTotalForms, result.TotalForms)
			assert.Equal(t, tt.expectedFormsWithLabels, result.FormsWithLabels)
			assert.Equal(t, tt.expectedFormsWithoutLabels, result.FormsWithoutLabels)
			assert.Equal(t, tt.expectedRequiredFields, result.RequiredFields)
			assert.Equal(t, tt.expectedFieldsetUsage, result.FieldsetUsage)
			
			// Check for appropriate issues
			if tt.expectedFormsWithoutLabels > 0 {
				found := false
				for _, issue := range result.Issues {
					if strings.Contains(issue, "label") {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected label issue to be reported")
			}
		})
	}
}

func TestAccessibilityAnalyzer_calculateAccessibilityScore(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	// Create test data
	wcag := WCAGCompliance{
		Compliance: ComplianceMetrics{
			ComplianceRate: 80.0,
		},
		Violations: []WCAGViolation{
			{Rule: "1.1.1", Level: "A", Impact: "critical"},
		},
	}
	
	contrast := ColorContrastAnalysis{
		FailedContrasts: 1,
	}
	
	altTags := AltTagAnalysis{
		TotalImages:      10,
		ImagesWithAlt:    8,
		ImagesWithoutAlt: 2,
		AltTextQuality: AltTextQuality{
			QualityScore: 75.0,
		},
	}
	
	keyboard := KeyboardNavigation{
		TabIndexIssues: 1,
		SkipLinks:      0,
	}
	
	forms := FormAccessibility{
		TotalForms:      2,
		FormsWithLabels: 1,
	}
	
	result := analyzer.calculateAccessibilityScore(wcag, contrast, altTags, keyboard, forms)
	
	assert.True(t, result.OverallScore >= 0 && result.OverallScore <= 100)
	assert.Equal(t, 80, result.WCAGScore)
	assert.Equal(t, 80, result.ContrastScore) // 100 - (1 * 20)
	assert.Equal(t, 60, result.AltTextScore) // (8/10 * 100) * (75/100)
	assert.Equal(t, 70, result.KeyboardScore) // 100 - 10 - 20
	assert.Equal(t, 50, result.FormScore) // 1/2 * 100
	
	// Check recommendations
	assert.True(t, len(result.Recommendations) > 0)
	
	// Should have recommendations for violations, contrast, alt text, keyboard, and forms
	recommendationTypes := make(map[string]bool)
	for _, rec := range result.Recommendations {
		recommendationTypes[rec.Type] = true
	}
	
	assert.True(t, recommendationTypes["wcag_compliance"])
	assert.True(t, recommendationTypes["color_contrast"])
	assert.True(t, recommendationTypes["alt_text"])
	assert.True(t, recommendationTypes["keyboard_navigation"])
	assert.True(t, recommendationTypes["form_labels"])
}

func TestAccessibilityAnalyzer_collectIssues(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	wcag := WCAGCompliance{
		Violations: []WCAGViolation{
			{Rule: "1.1.1", Description: "Missing alt text", Impact: "critical"},
		},
	}
	
	contrast := ColorContrastAnalysis{
		Issues: []string{"Low contrast detected"},
	}
	
	altTags := AltTagAnalysis{
		Issues: []string{"Image without alt attribute"},
	}
	
	keyboard := KeyboardNavigation{
		Issues: []string{"Positive tabindex found"},
	}
	
	forms := FormAccessibility{
		Issues: []string{"Form missing labels"},
	}
	
	issues := analyzer.collectIssues(wcag, contrast, altTags, keyboard, forms)
	
	assert.Equal(t, 5, len(issues))
	
	// Check that all issue types are present
	issueTypes := make(map[string]bool)
	for _, issue := range issues {
		issueTypes[issue.Type] = true
	}
	
	assert.True(t, issueTypes["wcag_violation"])
	assert.True(t, issueTypes["color_contrast"])
	assert.True(t, issueTypes["alt_text"])
	assert.True(t, issueTypes["keyboard_navigation"])
	assert.True(t, issueTypes["form_accessibility"])
}

func TestAccessibilityAnalyzer_getWCAGSuggestion(t *testing.T) {
	logger := logrus.New()
	analyzer := NewAccessibilityAnalyzer(logger)
	
	tests := []struct {
		rule               string
		expectedContains   string
	}{
		{"1.1.1", "alternative text"},
		{"2.4.2", "descriptive and unique titles"},
		{"3.1.1", "lang attribute"},
		{"unknown", "Review WCAG guidelines"},
	}
	
	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			suggestion := analyzer.getWCAGSuggestion(tt.rule)
			assert.Contains(t, suggestion, tt.expectedContains)
		})
	}
}