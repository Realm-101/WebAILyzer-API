package wappalyzer

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", ve.Field, ve.Message)
}

// Validator provides validation utilities
type Validator struct {
	urlRegex   *regexp.Regexp
	emailRegex *regexp.Regexp
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		urlRegex:   regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`),
		emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	}
}

// ValidateURL validates a URL string
func (v *Validator) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return ValidationError{Field: "url", Message: "URL cannot be empty"}
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ValidationError{Field: "url", Message: fmt.Sprintf("invalid URL format: %v", err)}
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ValidationError{Field: "url", Message: "URL must use http or https scheme"}
	}

	// Check host
	if parsedURL.Host == "" {
		return ValidationError{Field: "url", Message: "URL must have a valid host"}
	}

	return nil
}

// ValidateHeaders validates HTTP headers
func (v *Validator) ValidateHeaders(headers map[string][]string) error {
	if headers == nil {
		return ValidationError{Field: "headers", Message: "headers cannot be nil"}
	}

	for key, values := range headers {
		if key == "" {
			return ValidationError{Field: "headers", Message: "header key cannot be empty"}
		}

		if len(values) == 0 {
			return ValidationError{Field: "headers", Message: fmt.Sprintf("header '%s' has no values", key)}
		}

		// Check for suspicious header values that might indicate injection attempts
		for _, value := range values {
			if strings.Contains(value, "\n") || strings.Contains(value, "\r") {
				return ValidationError{Field: "headers", Message: fmt.Sprintf("header '%s' contains invalid characters", key)}
			}
		}
	}

	return nil
}

// ValidateBody validates the response body
func (v *Validator) ValidateBody(body []byte) error {
	if body == nil {
		return ValidationError{Field: "body", Message: "body cannot be nil"}
	}

	// Check for reasonable size limits (10MB)
	const maxBodySize = 10 * 1024 * 1024
	if len(body) > maxBodySize {
		return ValidationError{Field: "body", Message: fmt.Sprintf("body size (%d bytes) exceeds maximum allowed size (%d bytes)", len(body), maxBodySize)}
	}

	return nil
}

// ValidateFingerprint validates a fingerprint structure
func (v *Validator) ValidateFingerprint(name string, fp *Fingerprint) []ValidationError {
	var errors []ValidationError

	if name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "fingerprint name cannot be empty"})
	}

	if fp == nil {
		errors = append(errors, ValidationError{Field: "fingerprint", Message: "fingerprint cannot be nil"})
		return errors
	}

	// Validate categories
	if len(fp.Cats) == 0 {
		errors = append(errors, ValidationError{Field: "cats", Message: "fingerprint must have at least one category"})
	}

	// Validate patterns
	for header, pattern := range fp.Headers {
		if header == "" {
			errors = append(errors, ValidationError{Field: "headers", Message: "header key cannot be empty"})
		}
		if _, err := ParsePattern(pattern); err != nil {
			errors = append(errors, ValidationError{Field: "headers", Message: fmt.Sprintf("invalid pattern for header '%s': %v", header, err)})
		}
	}

	for i, pattern := range fp.HTML {
		if _, err := ParsePattern(pattern); err != nil {
			errors = append(errors, ValidationError{Field: "html", Message: fmt.Sprintf("invalid HTML pattern at index %d: %v", i, err)})
		}
	}

	for cookie, pattern := range fp.Cookies {
		if cookie == "" {
			errors = append(errors, ValidationError{Field: "cookies", Message: "cookie key cannot be empty"})
		}
		if _, err := ParsePattern(pattern); err != nil {
			errors = append(errors, ValidationError{Field: "cookies", Message: fmt.Sprintf("invalid pattern for cookie '%s': %v", cookie, err)})
		}
	}

	// Validate URLs
	if fp.Website != "" {
		if err := v.ValidateURL(fp.Website); err != nil {
			errors = append(errors, ValidationError{Field: "website", Message: fmt.Sprintf("invalid website URL: %v", err)})
		}
	}

	return errors
}

// SanitizeInput sanitizes user input to prevent injection attacks
func (v *Validator) SanitizeInput(input string) string {
	// Remove control characters
	input = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(input, "")
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// Limit length
	const maxInputLength = 1000
	if len(input) > maxInputLength {
		input = input[:maxInputLength]
	}
	
	return input
}