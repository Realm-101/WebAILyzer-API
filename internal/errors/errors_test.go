package errors

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		expected string
	}{
		{
			name: "error without cause",
			apiError: &APIError{
				Code:    ErrCodeInvalidRequest,
				Message: "Invalid request format",
			},
			expected: "INVALID_REQUEST: Invalid request format",
		},
		{
			name: "error with cause",
			apiError: &APIError{
				Code:    ErrCodeDatabaseError,
				Message: "Database operation failed",
				Cause:   fmt.Errorf("connection timeout"),
			},
			expected: "DATABASE_ERROR: Database operation failed (caused by: connection timeout)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.apiError.Error())
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	apiErr := NewAPIErrorWithCause(ErrCodeInternalError, "Internal error", http.StatusInternalServerError, originalErr)

	unwrapped := apiErr.Unwrap()
	assert.Equal(t, originalErr, unwrapped)

	// Test error without cause
	apiErrNoCause := NewAPIError(ErrCodeInvalidRequest, "Invalid request", http.StatusBadRequest)
	assert.Nil(t, apiErrNoCause.Unwrap())
}

func TestNewAPIError(t *testing.T) {
	code := ErrCodeInvalidRequest
	message := "Test error message"
	statusCode := http.StatusBadRequest

	apiErr := NewAPIError(code, message, statusCode)

	assert.Equal(t, code, apiErr.Code)
	assert.Equal(t, message, apiErr.Message)
	assert.Equal(t, statusCode, apiErr.StatusCode)
	assert.Nil(t, apiErr.Details)
	assert.Nil(t, apiErr.Cause)
	assert.NotEmpty(t, apiErr.Timestamp)

	// Verify timestamp format
	_, err := time.Parse(time.RFC3339, apiErr.Timestamp)
	assert.NoError(t, err)
}

func TestNewAPIErrorWithDetails(t *testing.T) {
	code := ErrCodeValidationFailed
	message := "Validation failed"
	statusCode := http.StatusBadRequest
	details := map[string]interface{}{
		"field": "email",
		"value": "invalid-email",
	}

	apiErr := NewAPIErrorWithDetails(code, message, statusCode, details)

	assert.Equal(t, code, apiErr.Code)
	assert.Equal(t, message, apiErr.Message)
	assert.Equal(t, statusCode, apiErr.StatusCode)
	assert.Equal(t, details, apiErr.Details)
	assert.Nil(t, apiErr.Cause)
}

func TestNewAPIErrorWithCause(t *testing.T) {
	code := ErrCodeDatabaseError
	message := "Database error"
	statusCode := http.StatusInternalServerError
	cause := fmt.Errorf("connection failed")

	apiErr := NewAPIErrorWithCause(code, message, statusCode, cause)

	assert.Equal(t, code, apiErr.Code)
	assert.Equal(t, message, apiErr.Message)
	assert.Equal(t, statusCode, apiErr.StatusCode)
	assert.Nil(t, apiErr.Details)
	assert.Equal(t, cause, apiErr.Cause)
}

func TestNewAPIErrorWithDetailsAndCause(t *testing.T) {
	code := ErrCodeExternalService
	message := "External service error"
	statusCode := http.StatusBadGateway
	details := map[string]interface{}{
		"service": "payment-gateway",
		"timeout": "30s",
	}
	cause := fmt.Errorf("service unavailable")

	apiErr := NewAPIErrorWithDetailsAndCause(code, message, statusCode, details, cause)

	assert.Equal(t, code, apiErr.Code)
	assert.Equal(t, message, apiErr.Message)
	assert.Equal(t, statusCode, apiErr.StatusCode)
	assert.Equal(t, details, apiErr.Details)
	assert.Equal(t, cause, apiErr.Cause)
}

func TestPredefinedErrorConstructors(t *testing.T) {
	tests := []struct {
		name           string
		constructor    func() *APIError
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "BadRequest",
			constructor:    func() *APIError { return BadRequest("Bad request") },
			expectedCode:   ErrCodeInvalidRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized",
			constructor:    func() *APIError { return Unauthorized("Unauthorized") },
			expectedCode:   ErrCodeUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Forbidden",
			constructor:    func() *APIError { return Forbidden("Forbidden") },
			expectedCode:   ErrCodeForbidden,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "NotFound",
			constructor:    func() *APIError { return NotFound("Not found") },
			expectedCode:   ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Conflict",
			constructor:    func() *APIError { return Conflict("Conflict") },
			expectedCode:   ErrCodeConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "RateLimited",
			constructor:    func() *APIError { return RateLimited("Rate limited") },
			expectedCode:   ErrCodeRateLimited,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "RequestTimeout",
			constructor:    func() *APIError { return RequestTimeout("Timeout") },
			expectedCode:   ErrCodeRequestTimeout,
			expectedStatus: http.StatusRequestTimeout,
		},
		{
			name:           "InternalError",
			constructor:    func() *APIError { return InternalError("Internal error") },
			expectedCode:   ErrCodeInternalError,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "ServiceUnavailable",
			constructor:    func() *APIError { return ServiceUnavailable("Service unavailable") },
			expectedCode:   ErrCodeServiceUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "BadGateway",
			constructor:    func() *APIError { return BadGateway("Bad gateway") },
			expectedCode:   ErrCodeBadGateway,
			expectedStatus: http.StatusBadGateway,
		},
		{
			name:           "GatewayTimeout",
			constructor:    func() *APIError { return GatewayTimeout("Gateway timeout") },
			expectedCode:   ErrCodeGatewayTimeout,
			expectedStatus: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := tt.constructor()
			assert.Equal(t, tt.expectedCode, apiErr.Code)
			assert.Equal(t, tt.expectedStatus, apiErr.StatusCode)
		})
	}
}

func TestBusinessLogicErrorConstructors(t *testing.T) {
	t.Run("WorkspaceNotFound", func(t *testing.T) {
		workspaceID := "test-workspace-id"
		apiErr := WorkspaceNotFound(workspaceID)

		assert.Equal(t, ErrCodeWorkspaceNotFound, apiErr.Code)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, "Workspace not found", apiErr.Message)
		require.NotNil(t, apiErr.Details)
		assert.Equal(t, workspaceID, apiErr.Details["workspace_id"])
	})

	t.Run("SessionNotFound", func(t *testing.T) {
		sessionID := "test-session-id"
		apiErr := SessionNotFound(sessionID)

		assert.Equal(t, ErrCodeSessionNotFound, apiErr.Code)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, "Session not found", apiErr.Message)
		require.NotNil(t, apiErr.Details)
		assert.Equal(t, sessionID, apiErr.Details["session_id"])
	})

	t.Run("InvalidURL", func(t *testing.T) {
		url := "invalid-url"
		cause := fmt.Errorf("parse error")
		apiErr := InvalidURL(url, cause)

		assert.Equal(t, ErrCodeInvalidURL, apiErr.Code)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, "The provided URL is invalid or malformed", apiErr.Message)
		require.NotNil(t, apiErr.Details)
		assert.Equal(t, url, apiErr.Details["url"])
		assert.Equal(t, cause, apiErr.Cause)
	})

	t.Run("BatchSizeExceeded", func(t *testing.T) {
		size := 150
		maxSize := 100
		apiErr := BatchSizeExceeded(size, maxSize)

		assert.Equal(t, ErrCodeBatchSizeExceeded, apiErr.Code)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Contains(t, apiErr.Message, "cannot exceed 100 URLs")
		require.NotNil(t, apiErr.Details)
		assert.Equal(t, size, apiErr.Details["provided_size"])
		assert.Equal(t, maxSize, apiErr.Details["max_size"])
	})
}

func TestErrorWithDetailsConstructors(t *testing.T) {
	t.Run("BadRequestWithDetails", func(t *testing.T) {
		message := "Validation failed"
		details := map[string]interface{}{
			"field": "email",
			"error": "invalid format",
		}
		apiErr := BadRequestWithDetails(message, details)

		assert.Equal(t, ErrCodeInvalidRequest, apiErr.Code)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, message, apiErr.Message)
		assert.Equal(t, details, apiErr.Details)
	})

	t.Run("ValidationFailed", func(t *testing.T) {
		message := "Field validation failed"
		details := map[string]interface{}{
			"fields": []string{"name", "email"},
		}
		apiErr := ValidationFailed(message, details)

		assert.Equal(t, ErrCodeValidationFailed, apiErr.Code)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, message, apiErr.Message)
		assert.Equal(t, details, apiErr.Details)
	})
}

func TestErrorWithCauseConstructors(t *testing.T) {
	t.Run("InternalErrorWithCause", func(t *testing.T) {
		message := "Database operation failed"
		cause := fmt.Errorf("connection timeout")
		apiErr := InternalErrorWithCause(message, cause)

		assert.Equal(t, ErrCodeInternalError, apiErr.Code)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, message, apiErr.Message)
		assert.Equal(t, cause, apiErr.Cause)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		message := "Query failed"
		cause := fmt.Errorf("syntax error")
		apiErr := DatabaseError(message, cause)

		assert.Equal(t, ErrCodeDatabaseError, apiErr.Code)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, message, apiErr.Message)
		assert.Equal(t, cause, apiErr.Cause)
	})

	t.Run("ExternalServiceError", func(t *testing.T) {
		message := "Payment service failed"
		cause := fmt.Errorf("service unavailable")
		apiErr := ExternalServiceError(message, cause)

		assert.Equal(t, ErrCodeExternalService, apiErr.Code)
		assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
		assert.Equal(t, message, apiErr.Message)
		assert.Equal(t, cause, apiErr.Cause)
	})
}

func TestIsAPIError(t *testing.T) {
	t.Run("is API error", func(t *testing.T) {
		apiErr := NewAPIError(ErrCodeInvalidRequest, "Test error", http.StatusBadRequest)
		assert.True(t, IsAPIError(apiErr))
	})

	t.Run("is not API error", func(t *testing.T) {
		regularErr := fmt.Errorf("regular error")
		assert.False(t, IsAPIError(regularErr))
	})
}

func TestAsAPIError(t *testing.T) {
	t.Run("convert API error", func(t *testing.T) {
		originalErr := NewAPIError(ErrCodeInvalidRequest, "Test error", http.StatusBadRequest)
		apiErr, ok := AsAPIError(originalErr)
		assert.True(t, ok)
		assert.Equal(t, originalErr, apiErr)
	})

	t.Run("convert regular error", func(t *testing.T) {
		regularErr := fmt.Errorf("regular error")
		apiErr, ok := AsAPIError(regularErr)
		assert.False(t, ok)
		assert.Nil(t, apiErr)
	})
}

func TestWrapError(t *testing.T) {
	t.Run("wrap regular error", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		code := ErrCodeInternalError
		message := "Wrapped error"
		statusCode := http.StatusInternalServerError

		apiErr := WrapError(originalErr, code, message, statusCode)

		assert.Equal(t, code, apiErr.Code)
		assert.Equal(t, message, apiErr.Message)
		assert.Equal(t, statusCode, apiErr.StatusCode)
		assert.Equal(t, originalErr, apiErr.Cause)
	})

	t.Run("wrap API error returns original", func(t *testing.T) {
		originalAPIErr := NewAPIError(ErrCodeInvalidRequest, "Original error", http.StatusBadRequest)
		
		wrappedErr := WrapError(originalAPIErr, ErrCodeInternalError, "Wrapped error", http.StatusInternalServerError)

		// Should return the original API error, not wrap it
		assert.Equal(t, originalAPIErr, wrappedErr)
	})
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined as expected
	expectedCodes := map[string]string{
		"INVALID_REQUEST":         ErrCodeInvalidRequest,
		"UNAUTHORIZED":            ErrCodeUnauthorized,
		"FORBIDDEN":               ErrCodeForbidden,
		"NOT_FOUND":               ErrCodeNotFound,
		"CONFLICT":                ErrCodeConflict,
		"VALIDATION_FAILED":       ErrCodeValidationFailed,
		"RATE_LIMITED":            ErrCodeRateLimited,
		"REQUEST_TIMEOUT":         ErrCodeRequestTimeout,
		"PAYLOAD_TOO_LARGE":       ErrCodePayloadTooLarge,
		"UNSUPPORTED_MEDIA_TYPE":  ErrCodeUnsupportedMedia,
		"INTERNAL_ERROR":          ErrCodeInternalError,
		"SERVICE_UNAVAILABLE":     ErrCodeServiceUnavailable,
		"BAD_GATEWAY":             ErrCodeBadGateway,
		"GATEWAY_TIMEOUT":         ErrCodeGatewayTimeout,
		"DATABASE_ERROR":          ErrCodeDatabaseError,
		"CACHE_ERROR":             ErrCodeCacheError,
		"EXTERNAL_SERVICE_ERROR":  ErrCodeExternalService,
		"WORKSPACE_NOT_FOUND":     ErrCodeWorkspaceNotFound,
		"SESSION_NOT_FOUND":       ErrCodeSessionNotFound,
		"ANALYSIS_NOT_FOUND":      ErrCodeAnalysisNotFound,
		"INSIGHT_NOT_FOUND":       ErrCodeInsightNotFound,
		"INVALID_URL":             ErrCodeInvalidURL,
		"CONNECTION_ERROR":        ErrCodeConnectionError,
		"ANALYSIS_TIMEOUT":        ErrCodeAnalysisTimeout,
		"BATCH_SIZE_EXCEEDED":     ErrCodeBatchSizeExceeded,
		"INSUFFICIENT_DATA":       ErrCodeInsufficientData,
	}

	for expected, actual := range expectedCodes {
		assert.Equal(t, expected, actual, "Error code mismatch for %s", expected)
	}
}