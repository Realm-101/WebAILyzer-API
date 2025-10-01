package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
	"github.com/projectdiscovery/wappalyzergo/internal/config"
	"github.com/sirupsen/logrus"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client *redis.Client
	config *config.RedisConfig
	logger *logrus.Logger
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.RedisConfig, logger *logrus.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: cfg.PoolSize,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	redisClient := &RedisClient{
		client: client,
		config: cfg,
		logger: logger,
	}

	logger.Info("Successfully connected to Redis")
	return redisClient, nil
}

// Set stores a value in Redis with the given key and TTL
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a value from Redis by key
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get value from Redis: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a key from Redis
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	if r.client != nil {
		err := r.client.Close()
		r.logger.Info("Redis connection closed")
		return err
	}
	return nil
}

// HealthCheck performs a health check on the Redis connection
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.client.Ping(ctx).Err()
}

// SetMultiple stores multiple key-value pairs in Redis with the same TTL
func (r *RedisClient) SetMultiple(ctx context.Context, pairs map[string]interface{}, ttl time.Duration) error {
	pipe := r.client.Pipeline()
	
	for key, value := range pairs {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, ttl)
	}
	
	_, err := pipe.Exec(ctx)
	return err
}

// GetMultiple retrieves multiple values from Redis by keys
func (r *RedisClient) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}
	
	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)
	
	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}
	
	results := make(map[string]interface{})
	for key, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				continue // Skip missing keys
			}
			r.logger.WithError(err).WithField("key", key).Warn("Failed to get value in batch operation")
			continue
		}
		
		var value interface{}
		if err := json.Unmarshal([]byte(data), &value); err != nil {
			r.logger.WithError(err).WithField("key", key).Warn("Failed to unmarshal value in batch operation")
			continue
		}
		
		results[key] = value
	}
	
	return results, nil
}

// DeleteMultiple removes multiple keys from Redis
func (r *RedisClient) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	
	return r.client.Del(ctx, keys...).Err()
}

// Keys returns all keys matching a pattern (use with caution in production)
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// TTL returns the time to live for a key
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Expire sets a timeout on a key
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// ErrCacheMiss is returned when a key is not found in the cache
var ErrCacheMiss = fmt.Errorf("cache miss")