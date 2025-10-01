package analyzers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SEOAnalyzer handles SEO analysis including meta tags and content structure
type SEOAnalyzer struct {
	logger *logrus.Logger
}

// NewSEOAnalyzer creates a new SEO analyzer instance
func NewSEOAnalyzer(logger *logrus.Logger) *SEOAnalyzer {
	return &SEOAnalyzer{
		logger: logger,
	}
}

// SEOAnalysisResult represents the result of SEO analysis
type SEOAnalysisResult struct {
	MetaTags        MetaTagAnalysis     `json:"meta_tags"`
	ContentStructure ContentStructure   `json:"content_structure"`
	StructuredData  StructuredData     `json:"structured_data"`
	SEOScore        SEOScoring         `json:"seo_score"`
	Metadata        SEOMetadata        `json:"metadata"`
}

// MetaTagAnalysis contains meta tag information
type MetaTagAnalysis struct {
	Title       MetaTag   `json:"title"`
	Description MetaTag   `json:"description"`
	Keywords    MetaTag   `json:"keywords"`
	Viewport    MetaTag   `json:"viewport"`
	Robots      MetaTag   `json:"robots"`
	Canonical   MetaTag   `json:"canonical"`
	OpenGraph   []MetaTag `json:"open_graph"`
	TwitterCard []MetaTag `json:"twitter_card"`
	Other       []MetaTag `json:"other"`
}

// MetaTag represents a single meta tag
type MetaTag struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Present  bool   `json:"present"`
	Length   int    `json:"length"`
	Issues   []string `json:"issues,omitempty"`
}

// ContentStructure contains content structure analysis
type ContentStructure struct {
	Headings    HeadingAnalysis `json:"headings"`
	Links       LinkAnalysis    `json:"links"`
	Images      ImageAnalysis   `json:"images"`
	WordCount   int             `json:"word_count"`
	ReadingTime int             `json:"reading_time_minutes"`
}

// HeadingAnalysis contains heading structure information
type HeadingAnalysis struct {
	H1Count    int      `json:"h1_count"`
	H2Count    int      `json:"h2_count"`
	H3Count    int      `json:"h3_count"`
	H4Count    int      `json:"h4_count"`
	H5Count    int      `json:"h5_count"`
	H6Count    int      `json:"h6_count"`
	Structure  []string `json:"structure"`
	Issues     []string `json:"issues"`
}

// LinkAnalysis contains link analysis information
type LinkAnalysis struct {
	InternalLinks int      `json:"internal_links"`
	ExternalLinks int      `json:"external_links"`
	NoFollowLinks int      `json:"nofollow_links"`
	BrokenLinks   int      `json:"broken_links"`
	Issues        []string `json:"issues"`
}

// ImageAnalysis contains image SEO analysis
type ImageAnalysis struct {
	TotalImages      int      `json:"total_images"`
	ImagesWithAlt    int      `json:"images_with_alt"`
	ImagesWithoutAlt int      `json:"images_without_alt"`
	Issues           []string `json:"issues"`
}

// StructuredData contains structured data analysis
type StructuredData struct {
	JSONLDCount    int                    `json:"jsonld_count"`
	MicrodataCount int                    `json:"microdata_count"`
	RDFaCount      int                    `json:"rdfa_count"`
	Schemas        []StructuredDataSchema `json:"schemas"`
	Issues         []string               `json:"issues"`
}

// StructuredDataSchema represents a detected schema
type StructuredDataSchema struct {
	Type   string `json:"type"`
	Format string `json:"format"` // "json-ld", "microdata", "rdfa"
	Valid  bool   `json:"valid"`
}

// SEOScoring contains SEO scoring information
type SEOScoring struct {
	OverallScore int                `json:"overall_score"` // 0-100
	MetaScore    int                `json:"meta_score"`
	ContentScore int                `json:"content_score"`
	TechnicalScore int              `json:"technical_score"`
	Suggestions  []SEOSuggestion    `json:"suggestions"`
}

// SEOSuggestion represents an SEO improvement suggestion
type SEOSuggestion struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"` // "high", "medium", "low"
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// SEOMetadata contains analysis metadata
type SEOMetadata struct {
	AnalysisTime time.Duration `json:"analysis_time_ms"`
	Timestamp    time.Time     `json:"timestamp"`
	URL          string        `json:"url"`
	UserAgent    string        `json:"user_agent"`
}

// Analyze performs comprehensive SEO analysis
func (sa *SEOAnalyzer) Analyze(ctx context.Context, url string, headers http.Header, body []byte, userAgent string) (*SEOAnalysisResult, error) {
	startTime := time.Now()
	
	sa.logger.WithFields(logrus.Fields{
		"url":            url,
		"content_length": len(body),
		"user_agent":     userAgent,
	}).Debug("Starting SEO analysis")

	htmlContent := string(body)

	// Analyze meta tags
	metaTags := sa.analyzeMetaTags(htmlContent)
	
	// Analyze content structure
	contentStructure := sa.analyzeContentStructure(htmlContent, url)
	
	// Analyze structured data
	structuredData := sa.analyzeStructuredData(htmlContent)
	
	// Calculate SEO scores
	seoScore := sa.calculateSEOScore(metaTags, contentStructure, structuredData)

	analysisTime := time.Since(startTime)

	result := &SEOAnalysisResult{
		MetaTags:        metaTags,
		ContentStructure: contentStructure,
		StructuredData:  structuredData,
		SEOScore:        seoScore,
		Metadata: SEOMetadata{
			AnalysisTime: analysisTime,
			Timestamp:    startTime,
			URL:          url,
			UserAgent:    userAgent,
		},
	}

	sa.logger.WithFields(logrus.Fields{
		"url":              url,
		"overall_score":    seoScore.OverallScore,
		"word_count":       contentStructure.WordCount,
		"h1_count":         contentStructure.Headings.H1Count,
		"analysis_time_ms": analysisTime.Milliseconds(),
	}).Debug("SEO analysis completed")

	return result, nil
}

// analyzeMetaTags analyzes meta tags in the HTML content
func (sa *SEOAnalyzer) analyzeMetaTags(htmlContent string) MetaTagAnalysis {
	analysis := MetaTagAnalysis{
		OpenGraph:   []MetaTag{},
		TwitterCard: []MetaTag{},
		Other:       []MetaTag{},
	}

	// Title tag
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]*)</title>`)
	if matches := titleRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		analysis.Title = MetaTag{
			Name:    "title",
			Content: title,
			Present: true,
			Length:  len(title),
		}
		if len(title) < 30 {
			analysis.Title.Issues = append(analysis.Title.Issues, "Title too short (recommended: 30-60 characters)")
		} else if len(title) > 60 {
			analysis.Title.Issues = append(analysis.Title.Issues, "Title too long (recommended: 30-60 characters)")
		}
	}

	// Meta description
	descRegex := regexp.MustCompile(`<meta[^>]*name=["']description["'][^>]*content=["']([^"']*)["'][^>]*>`)
	if matches := descRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		desc := strings.TrimSpace(matches[1])
		analysis.Description = MetaTag{
			Name:    "description",
			Content: desc,
			Present: true,
			Length:  len(desc),
		}
		if len(desc) < 120 {
			analysis.Description.Issues = append(analysis.Description.Issues, "Description too short (recommended: 120-160 characters)")
		} else if len(desc) > 160 {
			analysis.Description.Issues = append(analysis.Description.Issues, "Description too long (recommended: 120-160 characters)")
		}
	}

	// Meta keywords
	keywordsRegex := regexp.MustCompile(`<meta[^>]*name=["']keywords["'][^>]*content=["']([^"']*)["'][^>]*>`)
	if matches := keywordsRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		keywords := strings.TrimSpace(matches[1])
		analysis.Keywords = MetaTag{
			Name:    "keywords",
			Content: keywords,
			Present: true,
			Length:  len(keywords),
		}
	}

	// Viewport
	viewportRegex := regexp.MustCompile(`<meta[^>]*name=["']viewport["'][^>]*content=["']([^"']*)["'][^>]*>`)
	if matches := viewportRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		viewport := strings.TrimSpace(matches[1])
		analysis.Viewport = MetaTag{
			Name:    "viewport",
			Content: viewport,
			Present: true,
			Length:  len(viewport),
		}
	}

	// Robots
	robotsRegex := regexp.MustCompile(`<meta[^>]*name=["']robots["'][^>]*content=["']([^"']*)["'][^>]*>`)
	if matches := robotsRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		robots := strings.TrimSpace(matches[1])
		analysis.Robots = MetaTag{
			Name:    "robots",
			Content: robots,
			Present: true,
			Length:  len(robots),
		}
	}

	// Canonical
	canonicalRegex := regexp.MustCompile(`<link[^>]*rel=["']canonical["'][^>]*href=["']([^"']*)["'][^>]*>`)
	if matches := canonicalRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		canonical := strings.TrimSpace(matches[1])
		analysis.Canonical = MetaTag{
			Name:    "canonical",
			Content: canonical,
			Present: true,
			Length:  len(canonical),
		}
	}

	return analysis
}

// analyzeContentStructure analyzes the content structure of the HTML
func (sa *SEOAnalyzer) analyzeContentStructure(htmlContent, url string) ContentStructure {
	structure := ContentStructure{
		Headings: HeadingAnalysis{
			Structure: []string{},
			Issues:    []string{},
		},
		Links: LinkAnalysis{
			Issues: []string{},
		},
		Images: ImageAnalysis{
			Issues: []string{},
		},
	}

	// Analyze headings
	headingRegexes := map[string]*regexp.Regexp{
		"h1": regexp.MustCompile(`<h1[^>]*>([^<]*)</h1>`),
		"h2": regexp.MustCompile(`<h2[^>]*>([^<]*)</h2>`),
		"h3": regexp.MustCompile(`<h3[^>]*>([^<]*)</h3>`),
		"h4": regexp.MustCompile(`<h4[^>]*>([^<]*)</h4>`),
		"h5": regexp.MustCompile(`<h5[^>]*>([^<]*)</h5>`),
		"h6": regexp.MustCompile(`<h6[^>]*>([^<]*)</h6>`),
	}

	for tag, regex := range headingRegexes {
		matches := regex.FindAllStringSubmatch(htmlContent, -1)
		count := len(matches)
		
		switch tag {
		case "h1":
			structure.Headings.H1Count = count
			if count == 0 {
				structure.Headings.Issues = append(structure.Headings.Issues, "Missing H1 tag")
			} else if count > 1 {
				structure.Headings.Issues = append(structure.Headings.Issues, "Multiple H1 tags found")
			}
		case "h2":
			structure.Headings.H2Count = count
		case "h3":
			structure.Headings.H3Count = count
		case "h4":
			structure.Headings.H4Count = count
		case "h5":
			structure.Headings.H5Count = count
		case "h6":
			structure.Headings.H6Count = count
		}

		for _, match := range matches {
			if len(match) > 1 {
				structure.Headings.Structure = append(structure.Headings.Structure, fmt.Sprintf("%s: %s", strings.ToUpper(tag), strings.TrimSpace(match[1])))
			}
		}
	}

	// Analyze links
	linkRegex := regexp.MustCompile(`<a[^>]*href=["']([^"']*)["'][^>]*>`)
	linkMatches := linkRegex.FindAllStringSubmatch(htmlContent, -1)
	
	for _, match := range linkMatches {
		if len(match) > 1 {
			href := match[1]
			if strings.HasPrefix(href, "http") {
				structure.Links.ExternalLinks++
			} else if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "#") {
				structure.Links.InternalLinks++
			}
		}
	}

	// Analyze images
	imgRegex := regexp.MustCompile(`<img[^>]*>`)
	imgMatches := imgRegex.FindAllString(htmlContent, -1)
	structure.Images.TotalImages = len(imgMatches)

	altRegex := regexp.MustCompile(`alt=["']([^"']*)["']`)
	for _, img := range imgMatches {
		if altRegex.MatchString(img) {
			structure.Images.ImagesWithAlt++
		} else {
			structure.Images.ImagesWithoutAlt++
		}
	}

	if structure.Images.ImagesWithoutAlt > 0 {
		structure.Images.Issues = append(structure.Images.Issues, fmt.Sprintf("%d images missing alt attributes", structure.Images.ImagesWithoutAlt))
	}

	// Calculate word count
	textRegex := regexp.MustCompile(`<[^>]*>`)
	textContent := textRegex.ReplaceAllString(htmlContent, " ")
	words := strings.Fields(textContent)
	structure.WordCount = len(words)
	structure.ReadingTime = structure.WordCount / 200 // Assuming 200 words per minute

	return structure
}

// analyzeStructuredData analyzes structured data in the HTML
func (sa *SEOAnalyzer) analyzeStructuredData(htmlContent string) StructuredData {
	data := StructuredData{
		Schemas: []StructuredDataSchema{},
		Issues:  []string{},
	}

	// JSON-LD
	jsonldRegex := regexp.MustCompile(`<script[^>]*type=["']application/ld\+json["'][^>]*>([^<]*)</script>`)
	jsonldMatches := jsonldRegex.FindAllStringSubmatch(htmlContent, -1)
	data.JSONLDCount = len(jsonldMatches)

	for _, match := range jsonldMatches {
		if len(match) > 1 {
			var jsonData map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &jsonData); err == nil {
				if schemaType, ok := jsonData["@type"].(string); ok {
					data.Schemas = append(data.Schemas, StructuredDataSchema{
						Type:   schemaType,
						Format: "json-ld",
						Valid:  true,
					})
				}
			} else {
				data.Issues = append(data.Issues, "Invalid JSON-LD found")
			}
		}
	}

	// Microdata
	microdataRegex := regexp.MustCompile(`itemscope[^>]*itemtype=["']([^"']*)["']`)
	microdataMatches := microdataRegex.FindAllStringSubmatch(htmlContent, -1)
	data.MicrodataCount = len(microdataMatches)

	for _, match := range microdataMatches {
		if len(match) > 1 {
			data.Schemas = append(data.Schemas, StructuredDataSchema{
				Type:   match[1],
				Format: "microdata",
				Valid:  true,
			})
		}
	}

	// RDFa
	rdfaRegex := regexp.MustCompile(`typeof=["']([^"']*)["']`)
	rdfaMatches := rdfaRegex.FindAllStringSubmatch(htmlContent, -1)
	data.RDFaCount = len(rdfaMatches)

	return data
}

// calculateSEOScore calculates the overall SEO score
func (sa *SEOAnalyzer) calculateSEOScore(metaTags MetaTagAnalysis, contentStructure ContentStructure, structuredData StructuredData) SEOScoring {
	scoring := SEOScoring{
		Suggestions: []SEOSuggestion{},
	}

	// Meta score (0-40 points)
	metaScore := 0
	if metaTags.Title.Present && len(metaTags.Title.Issues) == 0 {
		metaScore += 15
	} else if metaTags.Title.Present {
		metaScore += 10
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "meta",
			Priority:    "high",
			Description: "Optimize title tag length",
			Impact:      "Title tag issues can impact search rankings",
		})
	} else {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "meta",
			Priority:    "high",
			Description: "Add title tag",
			Impact:      "Missing title tag severely impacts SEO",
		})
	}

	if metaTags.Description.Present && len(metaTags.Description.Issues) == 0 {
		metaScore += 15
	} else if metaTags.Description.Present {
		metaScore += 10
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "meta",
			Priority:    "high",
			Description: "Optimize meta description length",
			Impact:      "Meta description affects click-through rates",
		})
	} else {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "meta",
			Priority:    "high",
			Description: "Add meta description",
			Impact:      "Missing meta description reduces click-through rates",
		})
	}

	if metaTags.Viewport.Present {
		metaScore += 10
	} else {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "meta",
			Priority:    "medium",
			Description: "Add viewport meta tag for mobile optimization",
			Impact:      "Improves mobile user experience",
		})
	}

	scoring.MetaScore = metaScore

	// Content score (0-40 points)
	contentScore := 0
	if contentStructure.Headings.H1Count == 1 {
		contentScore += 15
	} else if contentStructure.Headings.H1Count == 0 {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "content",
			Priority:    "high",
			Description: "Add H1 heading",
			Impact:      "H1 headings help search engines understand page structure",
		})
	} else {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "content",
			Priority:    "medium",
			Description: "Use only one H1 heading per page",
			Impact:      "Multiple H1 tags can confuse search engines",
		})
		contentScore += 10
	}

	if contentStructure.WordCount >= 300 {
		contentScore += 15
	} else {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "content",
			Priority:    "medium",
			Description: "Increase content length (recommended: 300+ words)",
			Impact:      "Longer content tends to rank better",
		})
		contentScore += 5
	}

	if contentStructure.Images.ImagesWithoutAlt == 0 && contentStructure.Images.TotalImages > 0 {
		contentScore += 10
	} else if contentStructure.Images.ImagesWithoutAlt > 0 {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "content",
			Priority:    "medium",
			Description: "Add alt attributes to all images",
			Impact:      "Alt attributes improve accessibility and SEO",
		})
		contentScore += 5
	}

	scoring.ContentScore = contentScore

	// Technical score (0-20 points)
	technicalScore := 0
	if structuredData.JSONLDCount > 0 || structuredData.MicrodataCount > 0 {
		technicalScore += 10
	} else {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "technical",
			Priority:    "low",
			Description: "Add structured data markup",
			Impact:      "Structured data can enhance search result appearance",
		})
	}

	if metaTags.Canonical.Present {
		technicalScore += 10
	} else {
		scoring.Suggestions = append(scoring.Suggestions, SEOSuggestion{
			Type:        "technical",
			Priority:    "low",
			Description: "Add canonical URL",
			Impact:      "Prevents duplicate content issues",
		})
	}

	scoring.TechnicalScore = technicalScore
	scoring.OverallScore = metaScore + contentScore + technicalScore

	return scoring
}