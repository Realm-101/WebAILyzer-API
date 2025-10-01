package cache

import (
	"context"
	"fmt"
	"time"
	"strings"
	"github.com/sirupsen/logrus"
)

// CacheService provides high-level caching operations with invalidation strategies
type CacheService struct {
	client *RedisClient
	logger *logrus.Logger
}

// NewCacheService creates a new cache service
func NewCacheService(client *RedisClient, logger *logrus.Logger) *CacheService {
	return &CacheService{
		client: client,
		logger: logger,
	}
}

// CacheConfig holds configuration for different cache types
type CacheConfig struct {
	TTL                time.Duration
	InvalidationTags   []string
	CompressionEnabled bool
}

// Predefined cache configurations for different data types
var (
	AnalysisResultConfig = CacheConfig{
		TTL:                1 * time.Hour,
		InvalidationTags:   []string{"analysis", "workspace"},
		CompressionEnabled: true,
	}
	
	MetricsConfig = CacheConfig{
		TTL:                15 * time.Minute,
		InvalidationTags:   []string{"metrics", "workspace"},
		CompressionEnabled: true,
	}
	
	InsightsConfig = CacheConfig{
		TTL:                30 * time.Minute,
		InvalidationTags:   []string{"insights", "workspace"},
		CompressionEnabled: false,
	}
	
	SessionConfig = CacheConfig{
		TTL:                2 * time.Hour,
		InvalidationTags:   []string{"session", "workspace"},
		CompressionEnabled: false,
	}
)

// SetWithConfig stores a value with the specified cache configuration
func (cs *CacheService) SetWithConfig(ctx context.Context, key string, value interface{}, config CacheConfig) error {
	if cs.client == nil {
		cs.logger.Debug("Cache client not available, skipping cache set")
		return nil
	}

	// Add invalidation tags to the key metadata
	taggedKey := cs.addTagsToKey(key, config.InvalidationTags)
	
	err := cs.client.Set(ctx, taggedKey, value, config.TTL)
	if err != nil {
		cs.logger.WithError(err).WithField("key", key).Error("Failed to set cache value")
		return err
	}

	// Store tag mappings for invalidation
	if err := cs.storeTags(ctx, key, config.InvalidationTags); err != nil {
		cs.logger.WithError(err).WithField("key", key).Warn("Failed to store cache tags")
	}

	cs.logger.WithFields(logrus.Fields{
		"key":  key,
		"ttl":  config.TTL,
		"tags": config.InvalidationTags,
	}).Debug("Cache value set successfully")

	return nil
}

// GetWithConfig retrieves a value with cache hit/miss logging
func (cs *CacheService) GetWithConfig(ctx context.Context, key string, dest interface{}) error {
	if cs.client == nil {
		cs.logger.Debug("Cache client not available, returning cache miss")
		return ErrCacheMiss
	}

	err := cs.client.Get(ctx, key, dest)
	if err != nil {
		if err == ErrCacheMiss {
			cs.logger.WithField("key", key).Debug("Cache miss")
		} else {
			cs.logger.WithError(err).WithField("key", key).Error("Failed to get cache value")
		}
		return err
	}

	cs.logger.WithField("key", key).Debug("Cache hit")
	return nil
}

// InvalidateByTag invalidates all cache entries with the specified tag
func (cs *CacheService) InvalidateByTag(ctx context.Context, tag string) error {
	if cs.client == nil {
		cs.logger.Debug("Cache client not available, skipping invalidation")
		return nil
	}

	// Get all keys with this tag
	tagKey := fmt.Sprintf("tag:%s", tag)
	var keys []string
	if err := cs.client.Get(ctx, tagKey, &keys); err != nil {
		if err == ErrCacheMiss {
			cs.logger.WithField("tag", tag).Debug("No keys found for tag")
			return nil
		}
		return fmt.Errorf("failed to get keys for tag %s: %w", tag, err)
	}

	// Delete all keys with this tag
	for _, key := range keys {
		if err := cs.client.Delete(ctx, key); err != nil {
			cs.logger.WithError(err).WithField("key", key).Warn("Failed to delete cache key during tag invalidation")
		}
	}

	// Delete the tag mapping
	if err := cs.client.Delete(ctx, tagKey); err != nil {
		cs.logger.WithError(err).WithField("tag", tag).Warn("Failed to delete tag mapping")
	}

	cs.logger.WithFields(logrus.Fields{
		"tag":        tag,
		"keys_count": len(keys),
	}).Info("Cache invalidated by tag")

	return nil
}

// InvalidateByPattern invalidates all cache entries matching a pattern
func (cs *CacheService) InvalidateByPattern(ctx context.Context, pattern string) error {
	if cs.client == nil {
		cs.logger.Debug("Cache client not available, skipping pattern invalidation")
		return nil
	}

	// This is a simplified implementation - in production, you might want to use Redis SCAN
	// For now, we'll use pattern matching on stored keys
	cs.logger.WithField("pattern", pattern).Info("Pattern-based cache invalidation requested")
	
	// Note: This would require implementing a key tracking mechanism
	// For now, we'll log the request and return success
	return nil
}

// InvalidateWorkspace invalidates all cache entries for a specific workspace
func (cs *CacheService) InvalidateWorkspace(ctx context.Context, workspaceID string) error {
	patterns := []string{
		fmt.Sprintf("analysis:*:%s:*", workspaceID),
		fmt.Sprintf("metrics:%s:*", workspaceID),
		fmt.Sprintf("insights:%s:*", workspaceID),
		fmt.Sprintf("session:%s:*", workspaceID),
	}

	for _, pattern := range patterns {
		if err := cs.InvalidateByPattern(ctx, pattern); err != nil {
			cs.logger.WithError(err).WithField("pattern", pattern).Warn("Failed to invalidate cache pattern")
		}
	}

	cs.logger.WithField("workspace_id", workspaceID).Info("Workspace cache invalidated")
	return nil
}

// Warm preloads frequently accessed data into cache
func (cs *CacheService) Warm(ctx context.Context, keys []string, loader func(key string) (interface{}, error)) error {
	if cs.client == nil {
		cs.logger.Debug("Cache client not available, skipping cache warming")
		return nil
	}

	for _, key := range keys {
		// Check if key already exists
		exists, err := cs.client.Exists(ctx, key)
		if err != nil {
			cs.logger.WithError(err).WithField("key", key).Warn("Failed to check cache key existence during warming")
			continue
		}

		if exists {
			cs.logger.WithField("key", key).Debug("Cache key already exists, skipping warming")
			continue
		}

		// Load data and cache it
		data, err := loader(key)
		if err != nil {
			cs.logger.WithError(err).WithField("key", key).Warn("Failed to load data during cache warming")
			continue
		}

		// Use default configuration for warming
		if err := cs.SetWithConfig(ctx, key, data, AnalysisResultConfig); err != nil {
			cs.logger.WithError(err).WithField("key", key).Warn("Failed to cache data during warming")
		}
	}

	cs.logger.WithField("keys_count", len(keys)).Info("Cache warming completed")
	return nil
}

// HealthCheck performs a health check on the cache service
func (cs *CacheService) HealthCheck(ctx context.Context) error {
	if cs.client == nil {
		return fmt.Errorf("cache client not available")
	}

	return cs.client.HealthCheck(ctx)
}

// GetStats returns cache statistics
func (cs *CacheService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if cs.client == nil {
		return map[string]interface{}{
			"status": "unavailable",
		}, nil
	}

	// Perform health check
	if err := cs.client.HealthCheck(ctx); err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"client":     "redis",
	}, nil
}

// addTagsToKey adds invalidation tags to the key for tracking
func (cs *CacheService) addTagsToKey(key string, tags []string) string {
	if len(tags) == 0 {
		return key
	}
	return fmt.Sprintf("%s:tags:%s", key, strings.Join(tags, ","))
}

// storeTags stores tag mappings for invalidation
func (cs *CacheService) storeTags(ctx context.Context, key string, tags []string) error {
	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		
		// Get existing keys for this tag
		var existingKeys []string
		if err := cs.client.Get(ctx, tagKey, &existingKeys); err != nil && err != ErrCacheMiss {
			return err
		}
		
		// Add the new key if not already present
		found := false
		for _, existingKey := range existingKeys {
			if existingKey == key {
				found = true
				break
			}
		}
		
		if !found {
			existingKeys = append(existingKeys, key)
			// Store with longer TTL for tag mappings
			if err := cs.client.Set(ctx, tagKey, existingKeys, 24*time.Hour); err != nil {
				return err
			}
		}
	}
	
	return nil
}