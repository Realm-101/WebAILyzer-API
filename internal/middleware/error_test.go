package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/webailyzer/webailyzer-lite-api/internal/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandler_Middleware_Panic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	eh := NewErrorHandler(logger)

	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap with error handling middleware
	handler := eh.Middleware(panicHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check error structure
	require.Contains(t, response, "error")
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, errors.ErrCodeInternalError, errorObj["code"])
	assert.Equal(t, "An unexpected error occurred", errorObj["message"])
	assert.Contains(t, errorObj, "timestamp")
}

func TestErrorHandler_Middleware_NormalRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	eh := NewErrorHandler(logger)

	// Create a normal handler
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with error handling middleware
	handler := eh.Middleware(normalHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestErrorHandler_HandleError(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	eh := NewErrorHandler(logger)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "API error",
			err:            errors.BadRequest("Invalid request"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   errors.ErrCodeInvalidRequest,
		},
		{
			name:           "nil error",
			err:            nil,
			expectedStatus: 0, // No response written
		},
		{
			name:           "generic error",
			err:            fmt.Errorf("generic error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   errors.ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			eh.HandleError(w, req, tt.err)

			if tt.expectedStatus == 0 {
				// No response should be written for nil error
				assert.Equal(t, 200, w.Code) // Default status
				assert.Empty(t, w.Body.String())
				return
			}

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				require.Contains(t, response, "error")
				errorObj := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedCode, errorObj["code"])
			}
		})
	}
}

func TestErrorHandler_convertToAPIError(t *testing.T) {
	logger := logrus.New()
	eh := NewErrorHandler(logger)

	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "database error",
			err:          fmt.Errorf("database connection failed"),
			expectedCode: errors.ErrCodeDatabaseError,
		},
		{
			name:         "SQL error",
			err:          fmt.Errorf("SQL syntax error"),
			expectedCode: errors.ErrCodeDatabaseError,
		},
		{
			name:         "connection refused",
			err:          fmt.Errorf("connection refused"),
			expectedCode: errors.ErrCodeBadGateway,
		},
		{
			name:         "no such host",
			err:          fmt.Errorf("no such host"),
			expectedCode: errors.ErrCodeBadGateway,
		},
		{
			name:         "timeout error",
			err:          fmt.Errorf("request timeout"),
			expectedCode: errors.ErrCodeRequestTimeout,
		},
		{
			name:         "deadline exceeded",
			err:          fmt.Errorf("context deadline exceeded"),
			expectedCode: errors.ErrCodeRequestTimeout,
		},
		{
			name:         "context canceled",
			err:          fmt.Errorf("context canceled"),
			expectedCode: errors.ErrCodeRequestTimeout,
		},
		{
			name:         "invalid URL",
			err:          fmt.Errorf("invalid URL format"),
			expectedCode: errors.ErrCodeInvalidRequest,
		},
		{
			name:         "malformed URL",
			err:          fmt.Errorf("malformed URL"),
			expectedCode: errors.ErrCodeInvalidRequest,
		},
		{
			name:         "JSON error",
			err:          fmt.Errorf("invalid JSON format"),
			expectedCode: errors.ErrCodeInvalidRequest,
		},
		{
			name:         "unmarshal error",
			err:          fmt.Errorf("json: cannot unmarshal"),
			expectedCode: errors.ErrCodeInvalidRequest,
		},
		{
			name:         "validation error",
			err:          fmt.Errorf("validation failed"),
			expectedCode: errors.ErrCodeInvalidRequest,
		},
		{
			name:         "invalid field",
			err:          fmt.Errorf("invalid field value"),
			expectedCode: errors.ErrCodeInvalidRequest,
		},
		{
			name:         "generic error",
			err:          fmt.Errorf("some generic error"),
			expectedCode: errors.ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := eh.convertToAPIError(tt.err)
			assert.Equal(t, tt.expectedCode, apiErr.Code)
			assert.Equal(t, tt.err, apiErr.Cause)
		})
	}
}

func TestErrorHandler_writeErrorResponse(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	eh := NewErrorHandler(logger)

	t.Run("basic error response", func(t *testing.T) {
		apiErr := errors.BadRequest("Invalid request")
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		eh.writeErrorResponse(w, req, apiErr)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Contains(t, response, "error")
		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, errors.ErrCodeInvalidRequest, errorObj["code"])
		assert.Equal(t, "Invalid request", errorObj["message"])
		assert.Contains(t, errorObj, "timestamp")
	})

	t.Run("error response with details", func(t *testing.T) {
		details := map[string]interface{}{
			"field": "email",
			"value": "invalid",
		}
		apiErr := errors.BadRequestWithDetails("Validation failed", details)
		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()

		eh.writeErrorResponse(w, req, apiErr)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, details, errorObj["details"])
	})

	t.Run("error response with correlation ID", func(t *testing.T) {
		correlationID := "test-correlation-id"
		ctx := WithCorrelationID(context.Background(), correlationID)
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		apiErr := errors.InternalError("Internal error")
		eh.writeErrorResponse(w, req, apiErr)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, correlationID, response["correlation_id"])
	})
}

func TestCaptureResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		crw := &captureResponseWriter{ResponseWriter: w}

		crw.WriteHeader(http.StatusBadRequest)
		assert.Equal(t, http.StatusBadRequest, crw.statusCode)
	})

	t.Run("defaults to 200 on write", func(t *testing.T) {
		w := httptest.NewRecorder()
		crw := &captureResponseWriter{ResponseWriter: w}

		crw.Write([]byte("test"))
		assert.Equal(t, http.StatusOK, crw.statusCode)
	})
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "X-Forwarded-For single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1, 172.16.0.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "RemoteAddr with port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1",
			expected:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			ip := getClientIP(req)
			assert.Equal(t, tt.expected, ip)
		})
	}
}

func TestErrorHandlerFunc(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	eh := NewErrorHandler(logger)

	t.Run("handler returns error", func(t *testing.T) {
		handler := ErrorHandlerFunc(eh, func(w http.ResponseWriter, r *http.Request) error {
			return errors.BadRequest("Test error")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("handler returns nil", func(t *testing.T) {
		handler := ErrorHandlerFunc(eh, func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
			return nil
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})
}

func TestWithErrorHandler(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	handler := WithErrorHandler(logger, func(w http.ResponseWriter, r *http.Request) error {
		return errors.NotFound("Resource not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, errors.ErrCodeNotFound, errorObj["code"])
}

func TestContextHelpers(t *testing.T) {
	t.Run("WithCorrelationID and GetCorrelationID", func(t *testing.T) {
		ctx := context.Background()
		correlationID := "test-correlation-id"

		// Add correlation ID to context
		ctx = WithCorrelationID(ctx, correlationID)

		// Retrieve correlation ID from context
		retrievedID, ok := GetCorrelationID(ctx)
		assert.True(t, ok)
		assert.Equal(t, correlationID, retrievedID)

		// Test with context without correlation ID
		emptyCtx := context.Background()
		_, ok = GetCorrelationID(emptyCtx)
		assert.False(t, ok)
	})

	t.Run("WithWorkspaceID and GetWorkspaceID", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "test-workspace-id"

		// Add workspace ID to context
		ctx = WithWorkspaceID(ctx, workspaceID)

		// Retrieve workspace ID from context
		retrievedID, ok := GetWorkspaceID(ctx)
		assert.True(t, ok)
		assert.Equal(t, workspaceID, retrievedID)

		// Test with context without workspace ID
		emptyCtx := context.Background()
		_, ok = GetWorkspaceID(emptyCtx)
		assert.False(t, ok)
	})
}

func TestErrorHandler_Middleware_StatusCodeLogging(t *testing.T) {
	// Create a custom logger to capture log output
	logger := logrus.New()
	var logOutput strings.Builder
	logger.SetOutput(&logOutput)
	logger.SetLevel(logrus.WarnLevel)

	eh := NewErrorHandler(logger)

	// Create a handler that returns 404
	notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	// Wrap with error handling middleware
	handler := eh.Middleware(notFoundHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "Not Found", w.Body.String())

	// Verify that error status was logged
	logContent := logOutput.String()
	assert.Contains(t, logContent, "Request completed with error status")
	assert.Contains(t, logContent, "status_code=404")
}