package cache

import (
	"context"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sirupsen/logrus"
	"github.com/webailyzer/webailyzer-lite-api/internal/config"
)

var testConfig = config.RedisConfig{
	Host:     "localhost",
	Port:     6379,
	Password: "",
	Database: 1, // Use database 1 for tests
	PoolSize: 5,
}

func TestRedisClient_SetMultiple(t *testing.T) {
	client, err := NewRedisClient(&testConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	ctx := context.Background()
	
	// Test data
	pairs := map[string]interface{}{
		"test:multi:1": map[string]interface{}{"value": 1, "name": "first"},
		"test:multi:2": map[string]interface{}{"value": 2, "name": "second"},
		"test:multi:3": "simple string value",
	}
	
	// Set multiple values
	err = client.SetMultiple(ctx, pairs, 1*time.Hour)
	assert.NoError(t, err)
	
	// Verify all values were set
	for key, expectedValue := range pairs {
		var retrievedValue interface{}
		err = client.Get(ctx, key, &retrievedValue)
		assert.NoError(t, err)
		
		// For complex objects, compare as maps
		if expectedMap, ok := expectedValue.(map[string]interface{}); ok {
			retrievedMap, ok := retrievedValue.(map[string]interface{})
			require.True(t, ok, "Retrieved value should be a map")
			assert.Equal(t, expectedMap["value"], retrievedMap["value"])
			assert.Equal(t, expectedMap["name"], retrievedMap["name"])
		} else {
			assert.Equal(t, expectedValue, retrievedValue)
		}
		
		// Clean up
		client.Delete(ctx, key)
	}
}

func TestRedisClient_GetMultiple(t *testing.T) {
	client, err := NewRedisClient(&testConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	ctx := context.Background()
	
	// Set up test data
	testData := map[string]interface{}{
		"test:batch:1": "value1",
		"test:batch:2": "value2",
		"test:batch:3": map[string]interface{}{"nested": "value3"},
	}
	
	for key, value := range testData {
		err = client.Set(ctx, key, value, 1*time.Hour)
		require.NoError(t, err)
	}
	
	// Test getting multiple existing keys
	keys := []string{"test:batch:1", "test:batch:2", "test:batch:3"}
	results, err := client.GetMultiple(ctx, keys)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
	
	// Verify values
	assert.Equal(t, "value1", results["test:batch:1"])
	assert.Equal(t, "value2", results["test:batch:2"])
	
	// Test with mix of existing and non-existing keys
	mixedKeys := []string{"test:batch:1", "test:batch:nonexistent", "test:batch:2"}
	results, err = client.GetMultiple(ctx, mixedKeys)
	assert.NoError(t, err)
	assert.Len(t, results, 2) // Only existing keys should be returned
	
	// Test with empty keys slice
	results, err = client.GetMultiple(ctx, []string{})
	assert.NoError(t, err)
	assert.Len(t, results, 0)
	
	// Clean up
	for key := range testData {
		client.Delete(ctx, key)
	}
}

func TestRedisClient_DeleteMultiple(t *testing.T) {
	client, err := NewRedisClient(&testConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	ctx := context.Background()
	
	// Set up test data
	keys := []string{"test:del:1", "test:del:2", "test:del:3"}
	for _, key := range keys {
		err = client.Set(ctx, key, "test value", 1*time.Hour)
		require.NoError(t, err)
	}
	
	// Verify keys exist
	for _, key := range keys {
		exists, err := client.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)
	}
	
	// Delete multiple keys
	err = client.DeleteMultiple(ctx, keys)
	assert.NoError(t, err)
	
	// Verify keys are deleted
	for _, key := range keys {
		exists, err := client.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)
	}
	
	// Test with empty keys slice
	err = client.DeleteMultiple(ctx, []string{})
	assert.NoError(t, err)
}

func TestRedisClient_Keys(t *testing.T) {
	client, err := NewRedisClient(&testConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	ctx := context.Background()
	
	// Set up test data with pattern
	testKeys := []string{"test:pattern:1", "test:pattern:2", "test:other:1"}
	for _, key := range testKeys {
		err = client.Set(ctx, key, "test value", 1*time.Hour)
		require.NoError(t, err)
	}
	
	// Test pattern matching
	keys, err := client.Keys(ctx, "test:pattern:*")
	assert.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "test:pattern:1")
	assert.Contains(t, keys, "test:pattern:2")
	assert.NotContains(t, keys, "test:other:1")
	
	// Clean up
	for _, key := range testKeys {
		client.Delete(ctx, key)
	}
}

func TestRedisClient_TTL(t *testing.T) {
	client, err := NewRedisClient(&testConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	ctx := context.Background()
	testKey := "test:ttl:key"
	
	// Set key with TTL
	err = client.Set(ctx, testKey, "test value", 1*time.Hour)
	require.NoError(t, err)
	
	// Check TTL
	ttl, err := client.TTL(ctx, testKey)
	assert.NoError(t, err)
	assert.True(t, ttl > 0)
	assert.True(t, ttl <= 1*time.Hour)
	
	// Test non-existent key
	ttl, err = client.TTL(ctx, "test:ttl:nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(-2), ttl) // Redis returns -2 for non-existent keys
	
	// Clean up
	client.Delete(ctx, testKey)
}

func TestRedisClient_Expire(t *testing.T) {
	client, err := NewRedisClient(&testConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	ctx := context.Background()
	testKey := "test:expire:key"
	
	// Set key without TTL
	err = client.Set(ctx, testKey, "test value", 0) // 0 means no expiration
	require.NoError(t, err)
	
	// Set expiration
	err = client.Expire(ctx, testKey, 30*time.Minute)
	assert.NoError(t, err)
	
	// Check TTL was set
	ttl, err := client.TTL(ctx, testKey)
	assert.NoError(t, err)
	assert.True(t, ttl > 0)
	assert.True(t, ttl <= 30*time.Minute)
	
	// Clean up
	client.Delete(ctx, testKey)
}

func TestRedisClient_HealthCheck(t *testing.T) {
	client, err := NewRedisClient(&testConfig, logrus.New())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	ctx := context.Background()
	
	// Test health check
	err = client.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestRedisClient_ConnectionFailure(t *testing.T) {
	// Test with invalid configuration
	invalidConfig := config.RedisConfig{
		Host:     "invalid-host",
		Port:     9999,
		Password: "",
		Database: 0,
		PoolSize: 5,
	}
	
	_, err := NewRedisClient(&invalidConfig, logrus.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to Redis")
}

func TestRedisClient_NilClient(t *testing.T) {
	client := &RedisClient{
		client: nil,
		logger: logrus.New(),
	}
	
	ctx := context.Background()
	
	// Health check should fail with nil client
	err := client.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Redis client is not initialized")
}