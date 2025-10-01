package cache

import (
	"context"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sirupsen/logrus"
	"github.com/projectdiscovery/wappalyzergo/internal/config"
)

var testRedisConfig = config.RedisConfig{
	Host:     "localhost",
	Port:     6379,
	Password: "",
	Database: 1, // Use database 1 for tests
	PoolSize: 5,
}

func TestCacheService_SetWithConfig(t *testing.T) {
	// Skip if Redis is not available
	client, err := NewRedisClient(&testRedisConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	service := NewCacheService(client, logrus.New())
	ctx := context.Background()

	tests := []struct {
		name     string
		key      string
		value    interface{}
		config   CacheConfig
		wantErr  bool
	}{
		{
			name:  "set analysis result",
			key:   "test:analysis:123",
			value: map[string]interface{}{"url": "https://example.com", "score": 85},
			config: AnalysisResultConfig,
			wantErr: false,
		},
		{
			name:  "set metrics data",
			key:   "test:metrics:456",
			value: map[string]interface{}{"conversion_rate": 3.2, "bounce_rate": 45.6},
			config: MetricsConfig,
			wantErr: false,
		},
		{
			name:  "set with custom config",
			key:   "test:custom:789",
			value: "test value",
			config: CacheConfig{
				TTL:              5 * time.Minute,
				InvalidationTags: []string{"custom", "test"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SetWithConfig(ctx, tt.key, tt.value, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify the value was stored
				var retrieved interface{}
				err = service.GetWithConfig(ctx, tt.key, &retrieved)
				assert.NoError(t, err)
				
				// Clean up
				client.Delete(ctx, tt.key)
			}
		})
	}
}

func TestCacheService_GetWithConfig(t *testing.T) {
	client, err := NewRedisClient(&testRedisConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	service := NewCacheService(client, logrus.New())
	ctx := context.Background()

	// Test cache hit
	testKey := "test:get:hit"
	testValue := map[string]interface{}{"test": "value"}
	
	err = service.SetWithConfig(ctx, testKey, testValue, AnalysisResultConfig)
	require.NoError(t, err)

	var retrieved map[string]interface{}
	err = service.GetWithConfig(ctx, testKey, &retrieved)
	assert.NoError(t, err)
	assert.Equal(t, testValue["test"], retrieved["test"])

	// Test cache miss
	var missResult interface{}
	err = service.GetWithConfig(ctx, "test:get:miss", &missResult)
	assert.Equal(t, ErrCacheMiss, err)

	// Clean up
	client.Delete(ctx, testKey)
}

func TestCacheService_InvalidateByTag(t *testing.T) {
	client, err := NewRedisClient(&testRedisConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	service := NewCacheService(client, logrus.New())
	ctx := context.Background()

	// Set up test data with tags
	testKeys := []string{"test:tag:1", "test:tag:2", "test:tag:3"}
	testTag := "test_tag"
	config := CacheConfig{
		TTL:              1 * time.Hour,
		InvalidationTags: []string{testTag},
	}

	for _, key := range testKeys {
		err = service.SetWithConfig(ctx, key, "test value", config)
		require.NoError(t, err)
	}

	// Verify all keys exist
	for _, key := range testKeys {
		var value string
		err = service.GetWithConfig(ctx, key, &value)
		assert.NoError(t, err)
	}

	// Invalidate by tag
	err = service.InvalidateByTag(ctx, testTag)
	assert.NoError(t, err)

	// Verify all keys are invalidated
	for _, key := range testKeys {
		var value string
		err = service.GetWithConfig(ctx, key, &value)
		assert.Equal(t, ErrCacheMiss, err)
	}
}

func TestCacheService_InvalidateWorkspace(t *testing.T) {
	client, err := NewRedisClient(&testRedisConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	service := NewCacheService(client, logrus.New())
	ctx := context.Background()

	workspaceID := "test-workspace-123"

	// This test verifies the method runs without error
	// Full pattern-based invalidation would require more complex Redis operations
	err = service.InvalidateWorkspace(ctx, workspaceID)
	assert.NoError(t, err)
}

func TestCacheService_Warm(t *testing.T) {
	client, err := NewRedisClient(&testRedisConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	service := NewCacheService(client, logrus.New())
	ctx := context.Background()

	// Define test keys and loader function
	testKeys := []string{"warm:1", "warm:2", "warm:3"}
	loader := func(key string) (interface{}, error) {
		return map[string]interface{}{"key": key, "loaded": true}, nil
	}

	// Warm the cache
	err = service.Warm(ctx, testKeys, loader)
	assert.NoError(t, err)

	// Verify all keys were loaded
	for _, key := range testKeys {
		var value map[string]interface{}
		err = service.GetWithConfig(ctx, key, &value)
		assert.NoError(t, err)
		assert.Equal(t, key, value["key"])
		assert.Equal(t, true, value["loaded"])
		
		// Clean up
		client.Delete(ctx, key)
	}
}

func TestCacheService_GetStats(t *testing.T) {
	// Test with nil client
	service := NewCacheService(nil, logrus.New())
	ctx := context.Background()

	stats, err := service.GetStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "unavailable", stats["status"])

	// Test with real client
	client, err := NewRedisClient(&testRedisConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	service = NewCacheService(client, logrus.New())
	stats, err = service.GetStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", stats["status"])
	assert.Contains(t, stats, "timestamp")
	assert.Equal(t, "redis", stats["client"])
}

func TestCacheService_NilClient(t *testing.T) {
	// Test that service gracefully handles nil client
	service := NewCacheService(nil, logrus.New())
	ctx := context.Background()

	// All operations should succeed without error when client is nil
	err := service.SetWithConfig(ctx, "test", "value", AnalysisResultConfig)
	assert.NoError(t, err)

	var value string
	err = service.GetWithConfig(ctx, "test", &value)
	assert.Equal(t, ErrCacheMiss, err)

	err = service.InvalidateByTag(ctx, "test")
	assert.NoError(t, err)

	err = service.InvalidateWorkspace(ctx, "test")
	assert.NoError(t, err)

	err = service.Warm(ctx, []string{"test"}, func(key string) (interface{}, error) {
		return "value", nil
	})
	assert.NoError(t, err)
}

// Test configuration constants
func TestCacheConfigurations(t *testing.T) {
	// Test that all predefined configurations are valid
	configs := []CacheConfig{
		AnalysisResultConfig,
		MetricsConfig,
		InsightsConfig,
		SessionConfig,
	}

	for _, config := range configs {
		assert.True(t, config.TTL > 0, "TTL should be positive")
		assert.NotEmpty(t, config.InvalidationTags, "Should have invalidation tags")
	}

	// Test specific configuration values
	assert.Equal(t, 1*time.Hour, AnalysisResultConfig.TTL)
	assert.Contains(t, AnalysisResultConfig.InvalidationTags, "analysis")
	assert.True(t, AnalysisResultConfig.CompressionEnabled)

	assert.Equal(t, 15*time.Minute, MetricsConfig.TTL)
	assert.Contains(t, MetricsConfig.InvalidationTags, "metrics")

	assert.Equal(t, 30*time.Minute, InsightsConfig.TTL)
	assert.Contains(t, InsightsConfig.InvalidationTags, "insights")

	assert.Equal(t, 2*time.Hour, SessionConfig.TTL)
	assert.Contains(t, SessionConfig.InvalidationTags, "session")
}