package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/webailyzer/webailyzer-lite-api/internal/models"
	"github.com/webailyzer/webailyzer-lite-api/internal/repositories"
)

// AuthContextKey is the key used to store auth context in request context
type AuthContextKey string

const (
	AuthContextKeyValue AuthContextKey = "auth_context"
)

// AuthMiddleware provides workspace-based authentication
type AuthMiddleware struct {
	workspaceRepo repositories.WorkspaceRepository
	logger        *logrus.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(workspaceRepo repositories.WorkspaceRepository, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		workspaceRepo: workspaceRepo,
		logger:        logger,
	}
}

// Authenticate validates API key and adds workspace context to request
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from Authorization header
		apiKey := m.extractAPIKey(r)
		if apiKey == "" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "MISSING_API_KEY", "API key is required")
			return
		}

		// Validate API key and get workspace
		workspace, err := m.workspaceRepo.GetByAPIKey(r.Context(), apiKey)
		if err != nil {
			m.logger.WithError(err).WithField("api_key", apiKey).Warn("Failed to validate API key")
			m.writeErrorResponse(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key")
			return
		}

		if workspace == nil {
			m.writeErrorResponse(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key")
			return
		}

		if !workspace.IsActive {
			m.writeErrorResponse(w, http.StatusForbidden, "WORKSPACE_INACTIVE", "Workspace is inactive")
			return
		}

		// Create auth context
		authContext := &models.AuthContext{
			WorkspaceID: workspace.ID,
			APIKey:      apiKey,
			RateLimit:   workspace.RateLimit,
		}

		// Add auth context to request context
		ctx := context.WithValue(r.Context(), AuthContextKeyValue, authContext)
		r = r.WithContext(ctx)

		// Log successful authentication
		m.logger.WithFields(logrus.Fields{
			"workspace_id": workspace.ID,
			"workspace_name": workspace.Name,
		}).Debug("Request authenticated")

		next.ServeHTTP(w, r)
	})
}

// extractAPIKey extracts API key from Authorization header
func (m *AuthMiddleware) extractAPIKey(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Support both "Bearer <key>" and "ApiKey <key>" formats
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	scheme := strings.ToLower(parts[0])
	if scheme == "bearer" || scheme == "apikey" {
		return parts[1]
	}

	return ""
}

// writeErrorResponse writes a standardized error response
func (m *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// GetAuthContext extracts auth context from request context
func GetAuthContext(ctx context.Context) (*models.AuthContext, bool) {
	authContext, ok := ctx.Value(AuthContextKeyValue).(*models.AuthContext)
	return authContext, ok
}

// RequireWorkspace middleware that ensures workspace ID matches authenticated workspace
func (m *AuthMiddleware) RequireWorkspace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authContext, ok := GetAuthContext(r.Context())
		if !ok {
			m.writeErrorResponse(w, http.StatusUnauthorized, "MISSING_AUTH_CONTEXT", "Authentication required")
			return
		}

		// For requests with workspace_id in body, validate it matches authenticated workspace
		if r.Method == "POST" || r.Method == "PUT" {
			var requestBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err == nil {
				if workspaceIDStr, exists := requestBody["workspace_id"]; exists {
					if workspaceIDStr != authContext.WorkspaceID.String() {
						m.writeErrorResponse(w, http.StatusForbidden, "WORKSPACE_MISMATCH", "Workspace ID does not match authenticated workspace")
						return
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}