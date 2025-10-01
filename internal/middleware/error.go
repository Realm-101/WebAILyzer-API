package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/webailyzer/webailyzer-lite-api/internal/errors"
	"github.com/sirupsen/logrus"
)

// ErrorHandler middleware handles panics and provides structured error responses
type ErrorHandler struct {
	logger *logrus.Logger
}

// NewErrorHandler creates a new error handler middleware
func NewErrorHandler(logger *logrus.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// Middleware returns the error handling middleware function
func (eh *ErrorHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				eh.logger.WithFields(logrus.Fields{
					"panic":      err,
					"stack":      string(debug.Stack()),
					"method":     r.Method,
					"url":        r.URL.String(),
					"user_agent": r.UserAgent(),
					"remote_ip":  getClientIP(r),
				}).Error("Panic recovered in error handler middleware")

				// Create an internal server error
				apiErr := errors.InternalError("An unexpected error occurred")
				eh.writeErrorResponse(w, r, apiErr)
			}
		}()

		// Create a custom response writer to capture status codes
		crw := &captureResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(crw, r)

		// Log the request if it resulted in an error status
		if crw.statusCode >= 400 {
			eh.logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"url":         r.URL.String(),
				"status_code": crw.statusCode,
				"user_agent":  r.UserAgent(),
				"remote_ip":   getClientIP(r),
			}).Warn("Request completed with error status")
		}
	})
}

// HandleError handles an error and writes the appropriate response
func (eh *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Check if it's already an APIError
	if apiErr, ok := errors.AsAPIError(err); ok {
		eh.writeErrorResponse(w, r, apiErr)
		return
	}

	// Convert common error types to APIErrors
	apiErr := eh.convertToAPIError(err)
	eh.writeErrorResponse(w, r, apiErr)
}

// convertToAPIError converts common error types to APIErrors
func (eh *ErrorHandler) convertToAPIError(err error) *errors.APIError {
	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)

	// Database errors
	if strings.Contains(errMsgLower, "database") || strings.Contains(errMsgLower, "sql") {
		return errors.DatabaseError("Database operation failed", err)
	}

	// Network/connection errors
	if strings.Contains(errMsgLower, "connection refused") || strings.Contains(errMsgLower, "no such host") {
		return errors.NewAPIErrorWithCause(errors.ErrCodeBadGateway, "Unable to connect to external service", http.StatusBadGateway, err)
	}

	// Timeout errors
	if strings.Contains(errMsgLower, "timeout") || strings.Contains(errMsgLower, "deadline exceeded") {
		return errors.NewAPIErrorWithCause(errors.ErrCodeRequestTimeout, "Request timed out", http.StatusRequestTimeout, err)
	}

	// Context cancellation
	if strings.Contains(errMsgLower, "context canceled") {
		return errors.NewAPIErrorWithCause(errors.ErrCodeRequestTimeout, "Request was canceled", http.StatusRequestTimeout, err)
	}

	// URL parsing errors
	if strings.Contains(errMsgLower, "invalid url") || strings.Contains(errMsgLower, "malformed") {
		return errors.NewAPIErrorWithCause(errors.ErrCodeInvalidRequest, "Invalid URL format", http.StatusBadRequest, err)
	}

	// JSON parsing errors
	if strings.Contains(errMsgLower, "json") || strings.Contains(errMsgLower, "unmarshal") {
		return errors.NewAPIErrorWithCause(errors.ErrCodeInvalidRequest, "Invalid JSON format", http.StatusBadRequest, err)
	}

	// Validation errors
	if strings.Contains(errMsgLower, "validation") || strings.Contains(errMsgLower, "invalid") {
		return errors.NewAPIErrorWithCause(errors.ErrCodeInvalidRequest, errMsg, http.StatusBadRequest, err)
	}

	// Default to internal server error
	return errors.InternalErrorWithCause("An unexpected error occurred", err)
}

// writeErrorResponse writes a structured error response
func (eh *ErrorHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, apiErr *errors.APIError) {
	// Log the error with context
	logFields := logrus.Fields{
		"error_code":    apiErr.Code,
		"error_message": apiErr.Message,
		"status_code":   apiErr.StatusCode,
		"method":        r.Method,
		"url":           r.URL.String(),
		"user_agent":    r.UserAgent(),
		"remote_ip":     getClientIP(r),
	}

	// Add correlation ID if present
	if correlationID := r.Context().Value(CorrelationIDKey); correlationID != nil {
		logFields["correlation_id"] = correlationID
	}

	// Add workspace ID if present
	if workspaceID := r.Context().Value(WorkspaceIDKey); workspaceID != nil {
		logFields["workspace_id"] = workspaceID
	}

	// Add details if present
	if apiErr.Details != nil {
		logFields["error_details"] = apiErr.Details
	}

	// Log at appropriate level based on status code
	if apiErr.StatusCode >= 500 {
		if apiErr.Cause != nil {
			logFields["underlying_error"] = apiErr.Cause.Error()
		}
		eh.logger.WithFields(logFields).Error("Server error occurred")
	} else if apiErr.StatusCode >= 400 {
		eh.logger.WithFields(logFields).Warn("Client error occurred")
	}

	// Prepare error response
	errorResponse := map[string]interface{}{
		"error": apiErr,
	}

	// Add correlation ID to response if present
	if correlationID := r.Context().Value(CorrelationIDKey); correlationID != nil {
		errorResponse["correlation_id"] = correlationID
	}

	// Set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		eh.logger.WithError(err).Error("Failed to encode error response")
		// Fallback to plain text response
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Internal Server Error")
	}
}

// captureResponseWriter wraps http.ResponseWriter to capture status codes
type captureResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (crw *captureResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
	crw.ResponseWriter.WriteHeader(statusCode)
}

// Write ensures WriteHeader is called with 200 if not already called
func (crw *captureResponseWriter) Write(data []byte) (int, error) {
	if crw.statusCode == 0 {
		crw.statusCode = http.StatusOK
	}
	return crw.ResponseWriter.Write(data)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// ErrorHandlerFunc is a helper function to create error handling HTTP handlers
func ErrorHandlerFunc(eh *ErrorHandler, handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			eh.HandleError(w, r, err)
		}
	}
}

// WithErrorHandler wraps a handler function with error handling
func WithErrorHandler(logger *logrus.Logger, handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	eh := NewErrorHandler(logger)
	return ErrorHandlerFunc(eh, handler)
}

// ContextKey type for context keys
type ContextKey string

const (
	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey ContextKey = "correlation_id"
	// WorkspaceIDKey is the context key for workspace ID
	WorkspaceIDKey ContextKey = "workspace_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
)

// WithCorrelationID adds a correlation ID to the request context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the request context
func GetCorrelationID(ctx context.Context) (string, bool) {
	correlationID, ok := ctx.Value(CorrelationIDKey).(string)
	return correlationID, ok
}

// WithWorkspaceID adds a workspace ID to the request context
func WithWorkspaceID(ctx context.Context, workspaceID string) context.Context {
	return context.WithValue(ctx, WorkspaceIDKey, workspaceID)
}

// GetWorkspaceID retrieves the workspace ID from the request context
func GetWorkspaceID(ctx context.Context) (string, bool) {
	workspaceID, ok := ctx.Value(WorkspaceIDKey).(string)
	return workspaceID, ok
}