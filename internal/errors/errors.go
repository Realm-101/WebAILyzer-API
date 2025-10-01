package errors

import (
	"fmt"
	"net/http"
	"time"
)

// Error codes as constants
const (
	// Client errors (4xx)
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeValidationFailed   = "VALIDATION_FAILED"
	ErrCodeRateLimited        = "RATE_LIMITED"
	ErrCodeRequestTimeout     = "REQUEST_TIMEOUT"
	ErrCodePayloadTooLarge    = "PAYLOAD_TOO_LARGE"
	ErrCodeUnsupportedMedia   = "UNSUPPORTED_MEDIA_TYPE"

	// Server errors (5xx)
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeBadGateway         = "BAD_GATEWAY"
	ErrCodeGatewayTimeout     = "GATEWAY_TIMEOUT"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeCacheError         = "CACHE_ERROR"
	ErrCodeExternalService    = "EXTERNAL_SERVICE_ERROR"

	// Business logic errors
	ErrCodeWorkspaceNotFound  = "WORKSPACE_NOT_FOUND"
	ErrCodeSessionNotFound    = "SESSION_NOT_FOUND"
	ErrCodeAnalysisNotFound   = "ANALYSIS_NOT_FOUND"
	ErrCodeInsightNotFound    = "INSIGHT_NOT_FOUND"
	ErrCodeInvalidURL         = "INVALID_URL"
	ErrCodeConnectionError    = "CONNECTION_ERROR"
	ErrCodeAnalysisTimeout    = "ANALYSIS_TIMEOUT"
	ErrCodeBatchSizeExceeded  = "BATCH_SIZE_EXCEEDED"
	ErrCodeInsufficientData   = "INSUFFICIENT_DATA"
)

// APIError represents a structured API error
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  string                 `json:"timestamp"`
	StatusCode int                    `json:"-"` // HTTP status code, not included in JSON response
	Cause      error                  `json:"-"` // Original error, not included in JSON response
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.Cause
}

// NewAPIError creates a new API error
func NewAPIError(code, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
}

// NewAPIErrorWithDetails creates a new API error with details
func NewAPIErrorWithDetails(code, message string, statusCode int, details map[string]interface{}) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
}

// NewAPIErrorWithCause creates a new API error with an underlying cause
func NewAPIErrorWithCause(code, message string, statusCode int, cause error) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      cause,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
}

// NewAPIErrorWithDetailsAndCause creates a new API error with details and cause
func NewAPIErrorWithDetailsAndCause(code, message string, statusCode int, details map[string]interface{}, cause error) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
		Cause:      cause,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
}

// Predefined error constructors for common errors

// BadRequest creates a bad request error
func BadRequest(message string) *APIError {
	return NewAPIError(ErrCodeInvalidRequest, message, http.StatusBadRequest)
}

// BadRequestWithDetails creates a bad request error with details
func BadRequestWithDetails(message string, details map[string]interface{}) *APIError {
	return NewAPIErrorWithDetails(ErrCodeInvalidRequest, message, http.StatusBadRequest, details)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *APIError {
	return NewAPIError(ErrCodeForbidden, message, http.StatusForbidden)
}

// NotFound creates a not found error
func NotFound(message string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, http.StatusNotFound)
}

// Conflict creates a conflict error
func Conflict(message string) *APIError {
	return NewAPIError(ErrCodeConflict, message, http.StatusConflict)
}

// ValidationFailed creates a validation failed error
func ValidationFailed(message string, details map[string]interface{}) *APIError {
	return NewAPIErrorWithDetails(ErrCodeValidationFailed, message, http.StatusBadRequest, details)
}

// RateLimited creates a rate limited error
func RateLimited(message string) *APIError {
	return NewAPIError(ErrCodeRateLimited, message, http.StatusTooManyRequests)
}

// RequestTimeout creates a request timeout error
func RequestTimeout(message string) *APIError {
	return NewAPIError(ErrCodeRequestTimeout, message, http.StatusRequestTimeout)
}

// InternalError creates an internal server error
func InternalError(message string) *APIError {
	return NewAPIError(ErrCodeInternalError, message, http.StatusInternalServerError)
}

// InternalErrorWithCause creates an internal server error with cause
func InternalErrorWithCause(message string, cause error) *APIError {
	return NewAPIErrorWithCause(ErrCodeInternalError, message, http.StatusInternalServerError, cause)
}

// ServiceUnavailable creates a service unavailable error
func ServiceUnavailable(message string) *APIError {
	return NewAPIError(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

// BadGateway creates a bad gateway error
func BadGateway(message string) *APIError {
	return NewAPIError(ErrCodeBadGateway, message, http.StatusBadGateway)
}

// GatewayTimeout creates a gateway timeout error
func GatewayTimeout(message string) *APIError {
	return NewAPIError(ErrCodeGatewayTimeout, message, http.StatusGatewayTimeout)
}

// DatabaseError creates a database error
func DatabaseError(message string, cause error) *APIError {
	return NewAPIErrorWithCause(ErrCodeDatabaseError, message, http.StatusInternalServerError, cause)
}

// CacheError creates a cache error
func CacheError(message string, cause error) *APIError {
	return NewAPIErrorWithCause(ErrCodeCacheError, message, http.StatusInternalServerError, cause)
}

// ExternalServiceError creates an external service error
func ExternalServiceError(message string, cause error) *APIError {
	return NewAPIErrorWithCause(ErrCodeExternalService, message, http.StatusBadGateway, cause)
}

// Business logic error constructors

// WorkspaceNotFound creates a workspace not found error
func WorkspaceNotFound(workspaceID string) *APIError {
	return NewAPIErrorWithDetails(ErrCodeWorkspaceNotFound, "Workspace not found", http.StatusNotFound, map[string]interface{}{
		"workspace_id": workspaceID,
	})
}

// SessionNotFound creates a session not found error
func SessionNotFound(sessionID string) *APIError {
	return NewAPIErrorWithDetails(ErrCodeSessionNotFound, "Session not found", http.StatusNotFound, map[string]interface{}{
		"session_id": sessionID,
	})
}

// AnalysisNotFound creates an analysis not found error
func AnalysisNotFound(analysisID string) *APIError {
	return NewAPIErrorWithDetails(ErrCodeAnalysisNotFound, "Analysis not found", http.StatusNotFound, map[string]interface{}{
		"analysis_id": analysisID,
	})
}

// InsightNotFound creates an insight not found error
func InsightNotFound(insightID string) *APIError {
	return NewAPIErrorWithDetails(ErrCodeInsightNotFound, "Insight not found", http.StatusNotFound, map[string]interface{}{
		"insight_id": insightID,
	})
}

// InvalidURL creates an invalid URL error
func InvalidURL(url string, cause error) *APIError {
	return NewAPIErrorWithDetailsAndCause(ErrCodeInvalidURL, "The provided URL is invalid or malformed", http.StatusBadRequest, map[string]interface{}{
		"url": url,
	}, cause)
}

// ConnectionError creates a connection error
func ConnectionError(url string, cause error) *APIError {
	return NewAPIErrorWithDetailsAndCause(ErrCodeConnectionError, "Unable to connect to the target URL", http.StatusBadGateway, map[string]interface{}{
		"url": url,
	}, cause)
}

// AnalysisTimeout creates an analysis timeout error
func AnalysisTimeout(url string) *APIError {
	return NewAPIErrorWithDetails(ErrCodeAnalysisTimeout, "Request timed out while analyzing URL", http.StatusRequestTimeout, map[string]interface{}{
		"url": url,
	})
}

// BatchSizeExceeded creates a batch size exceeded error
func BatchSizeExceeded(size, maxSize int) *APIError {
	return NewAPIErrorWithDetails(ErrCodeBatchSizeExceeded, fmt.Sprintf("Batch size cannot exceed %d URLs", maxSize), http.StatusBadRequest, map[string]interface{}{
		"provided_size": size,
		"max_size":      maxSize,
	})
}

// InsufficientData creates an insufficient data error
func InsufficientData(message string) *APIError {
	return NewAPIError(ErrCodeInsufficientData, message, http.StatusUnprocessableEntity)
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// AsAPIError converts an error to APIError if possible
func AsAPIError(err error) (*APIError, bool) {
	apiErr, ok := err.(*APIError)
	return apiErr, ok
}

// WrapError wraps a generic error into an APIError
func WrapError(err error, code, message string, statusCode int) *APIError {
	if apiErr, ok := AsAPIError(err); ok {
		return apiErr
	}
	return NewAPIErrorWithCause(code, message, statusCode, err)
}