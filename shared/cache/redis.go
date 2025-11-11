package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache wraps Redis client for caching operations
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(addr, password string) (*RedisCache, error) {
	if addr == "" {
		addr = os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = "localhost:6379"
		}
	}

	if password == "" {
		password = os.Getenv("REDIS_PASSWORD")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx := context.Background()

	// Ping to check connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Set stores a value in cache with expiration
func (rc *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = rc.client.Set(rc.ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(key string, dest interface{}) error {
	data, err := rc.client.Get(rc.ctx, key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("cache miss: key %s not found", key)
	}
	if err != nil {
		return fmt.Errorf("failed to get cache: %w", err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (rc *RedisCache) Delete(key string) error {
	err := rc.client.Del(rc.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (rc *RedisCache) Exists(key string) (bool, error) {
	count, err := rc.client.Exists(rc.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return count > 0, nil
}

// SetNX sets a value only if the key does not exist (for distributed locks)
func (rc *RedisCache) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	success, err := rc.client.SetNX(rc.ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set NX: %w", err)
	}

	return success, nil
}

// Increment increments a counter
func (rc *RedisCache) Increment(key string) (int64, error) {
	val, err := rc.client.Incr(rc.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}

	return val, nil
}

// IncrementBy increments a counter by a specific amount
func (rc *RedisCache) IncrementBy(key string, value int64) (int64, error) {
	val, err := rc.client.IncrBy(rc.ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment by: %w", err)
	}

	return val, nil
}

// Expire sets a timeout on a key
func (rc *RedisCache) Expire(key string, expiration time.Duration) error {
	err := rc.client.Expire(rc.ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

// TTL returns the remaining time to live of a key
func (rc *RedisCache) TTL(key string) (time.Duration, error) {
	ttl, err := rc.client.TTL(rc.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// FlushAll removes all keys from the cache (use with caution)
func (rc *RedisCache) FlushAll() error {
	err := rc.client.FlushAll(rc.ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// GetOrSet retrieves a value from cache, or computes and stores it if not found
func (rc *RedisCache) GetOrSet(key string, dest interface{}, expiration time.Duration, computeFunc func() (interface{}, error)) error {
	// Try to get from cache first
	err := rc.Get(key, dest)
	if err == nil {
		return nil // Cache hit
	}

	// Cache miss, compute the value
	value, err := computeFunc()
	if err != nil {
		return fmt.Errorf("failed to compute value: %w", err)
	}

	// Store in cache
	if err := rc.Set(key, value, expiration); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache value for key %s: %v\n", key, err)
	}

	// Convert value to dest
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal computed value: %w", err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal computed value: %w", err)
	}

	return nil
}

// InvalidatePattern deletes all keys matching a pattern
func (rc *RedisCache) InvalidatePattern(pattern string) error {
	iter := rc.client.Scan(rc.ctx, 0, pattern, 0).Iterator()
	for iter.Next(rc.ctx) {
		err := rc.client.Del(rc.ctx, iter.Val()).Err()
		if err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}
