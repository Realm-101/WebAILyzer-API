package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/projectdiscovery/wappalyzergo/internal/models"
)

// MockWorkspaceRepository is a mock implementation of WorkspaceRepository
type MockWorkspaceRepository struct {
	mock.Mock
}

func (m *MockWorkspaceRepository) Create(ctx context.Context, workspace *models.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Workspace, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, workspace *models.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) List(ctx context.Context, limit, offset int) ([]*models.Workspace, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Workspace), args.Error(1)
}

func TestAuthMiddleware_Authenticate(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(*MockWorkspaceRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing API key",
			authHeader:     "",
			setupMock:      func(m *MockWorkspaceRepository) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "MISSING_API_KEY",
		},
		{
			name:       "Valid API key with Bearer format",
			authHeader: "Bearer valid-api-key",
			setupMock: func(m *MockWorkspaceRepository) {
				workspace := &models.Workspace{
					ID:        uuid.New(),
					Name:      "Test Workspace",
					APIKey:    "valid-api-key",
					IsActive:  true,
					RateLimit: 1000,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("GetByAPIKey", mock.Anything, "valid-api-key").Return(workspace, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Valid API key with ApiKey format",
			authHeader: "ApiKey valid-api-key",
			setupMock: func(m *MockWorkspaceRepository) {
				workspace := &models.Workspace{
					ID:        uuid.New(),
					Name:      "Test Workspace",
					APIKey:    "valid-api-key",
					IsActive:  true,
					RateLimit: 1000,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("GetByAPIKey", mock.Anything, "valid-api-key").Return(workspace, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Invalid API key",
			authHeader: "Bearer invalid-api-key",
			setupMock: func(m *MockWorkspaceRepository) {
				m.On("GetByAPIKey", mock.Anything, "invalid-api-key").Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_API_KEY",
		},
		{
			name:       "Inactive workspace",
			authHeader: "Bearer inactive-api-key",
			setupMock: func(m *MockWorkspaceRepository) {
				workspace := &models.Workspace{
					ID:        uuid.New(),
					Name:      "Inactive Workspace",
					APIKey:    "inactive-api-key",
					IsActive:  false,
					RateLimit: 1000,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("GetByAPIKey", mock.Anything, "inactive-api-key").Return(workspace, nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "WORKSPACE_INACTIVE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockWorkspaceRepository{}
			tt.setupMock(mockRepo)

			middleware := NewAuthMiddleware(mockRepo, logger)

			// Create a test handler that checks if auth context is set
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authContext, ok := GetAuthContext(r.Context())
				if ok {
					assert.NotNil(t, authContext)
					assert.NotEqual(t, uuid.Nil, authContext.WorkspaceID)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute middleware
			middleware.Authenticate(testHandler).ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check error response if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				errorObj, ok := response["error"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware_ExtractAPIKey(t *testing.T) {
	logger := logrus.New()
	mockRepo := &MockWorkspaceRepository{}
	middleware := NewAuthMiddleware(mockRepo, logger)

	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{
			name:       "Bearer format",
			authHeader: "Bearer test-api-key",
			expected:   "test-api-key",
		},
		{
			name:       "ApiKey format",
			authHeader: "ApiKey test-api-key",
			expected:   "test-api-key",
		},
		{
			name:       "Invalid format",
			authHeader: "Invalid test-api-key",
			expected:   "",
		},
		{
			name:       "Missing scheme",
			authHeader: "test-api-key",
			expected:   "",
		},
		{
			name:       "Empty header",
			authHeader: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			result := middleware.extractAPIKey(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthMiddleware_RequireWorkspace(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	mockRepo := &MockWorkspaceRepository{}
	middleware := NewAuthMiddleware(mockRepo, logger)

	workspaceID := uuid.New()
	authContext := &models.AuthContext{
		WorkspaceID: workspaceID,
		APIKey:      "test-api-key",
		RateLimit:   1000,
	}

	tests := []struct {
		name           string
		setupContext   func(context.Context) context.Context
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Missing auth context",
			setupContext: func(ctx context.Context) context.Context {
				return ctx // No auth context
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "MISSING_AUTH_CONTEXT",
		},
		{
			name: "Valid workspace ID match",
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, AuthContextKeyValue, authContext)
			},
			requestBody:    `{"workspace_id": "` + workspaceID.String() + `", "url": "https://example.com"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Workspace ID mismatch",
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, AuthContextKeyValue, authContext)
			},
			requestBody:    `{"workspace_id": "` + uuid.New().String() + `", "url": "https://example.com"}`,
			expectedStatus: http.StatusForbidden,
			expectedError:  "WORKSPACE_MISMATCH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			var req *http.Request
			if tt.requestBody != "" {
				req = httptest.NewRequest("POST", "/test", strings.NewReader(tt.requestBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest("GET", "/test", nil)
			}

			// Setup context
			ctx := tt.setupContext(req.Context())
			req = req.WithContext(ctx)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute middleware
			middleware.RequireWorkspace(testHandler).ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check error response if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				errorObj, ok := response["error"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}
		})
	}
}

func TestGetAuthContext(t *testing.T) {
	workspaceID := uuid.New()
	authContext := &models.AuthContext{
		WorkspaceID: workspaceID,
		APIKey:      "test-api-key",
		RateLimit:   1000,
	}

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectFound bool
	}{
		{
			name: "Auth context present",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), AuthContextKeyValue, authContext)
			},
			expectFound: true,
		},
		{
			name: "Auth context missing",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result, found := GetAuthContext(ctx)

			assert.Equal(t, tt.expectFound, found)
			if tt.expectFound {
				assert.Equal(t, authContext, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}