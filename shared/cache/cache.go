package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache provides a unified caching interface
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	FlushAll(ctx context.Context) error
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int, prefix string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: prefix,
	}, nil
}

// prefixKey adds prefix to cache key
func (rc *RedisCache) prefixKey(key string) string {
	if rc.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", rc.prefix, key)
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.client.Get(ctx, rc.prefixKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rc.client.Set(ctx, rc.prefixKey(key), data, ttl).Err()
}

// Delete removes keys from cache
func (rc *RedisCache) Delete(ctx context.Context, keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = rc.prefixKey(key)
	}

	return rc.client.Del(ctx, prefixedKeys...).Err()
}

// Exists checks if key exists in cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, rc.prefixKey(key)).Result()
	return count > 0, err
}

// Increment increments a counter
func (rc *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return rc.client.Incr(ctx, rc.prefixKey(key)).Result()
}

// Decrement decrements a counter
func (rc *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	return rc.client.Decr(ctx, rc.prefixKey(key)).Result()
}

// Expire sets expiration on a key
func (rc *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return rc.client.Expire(ctx, rc.prefixKey(key), ttl).Err()
}

// FlushAll removes all keys from cache
func (rc *RedisCache) FlushAll(ctx context.Context) error {
	return rc.client.FlushDB(ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// ErrCacheMiss is returned when a cache key doesn't exist
var ErrCacheMiss = fmt.Errorf("cache miss")

// CacheStrategy defines different caching strategies
type CacheStrategy string

const (
	// CacheAside (Lazy Loading) - read from cache, if miss read from DB and populate cache
	CacheAside CacheStrategy = "cache_aside"

	// WriteThrough - write to cache and DB simultaneously
	WriteThrough CacheStrategy = "write_through"

	// WriteBack - write to cache immediately, write to DB asynchronously
	WriteBack CacheStrategy = "write_back"

	// RefreshAhead - asynchronously refresh cache before expiration
	RefreshAhead CacheStrategy = "refresh_ahead"
)

// CacheManager manages caching with different strategies
type CacheManager struct {
	cache    Cache
	strategy CacheStrategy
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache Cache, strategy CacheStrategy) *CacheManager {
	return &CacheManager{
		cache:    cache,
		strategy: strategy,
	}
}

// GetOrSet implements cache-aside pattern
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, loader func() (interface{}, error)) error {
	// Try to get from cache
	err := cm.cache.Get(ctx, key, dest)
	if err == nil {
		return nil // Cache hit
	}

	if err != ErrCacheMiss {
		return err // Actual error
	}

	// Cache miss - load from source
	value, err := loader()
	if err != nil {
		return err
	}

	// Store in cache
	if err := cm.cache.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to set cache: %v\n", err)
	}

	// Copy value to dest
	data, _ := json.Marshal(value)
	return json.Unmarshal(data, dest)
}

// CacheKeyBuilder helps build consistent cache keys
type CacheKeyBuilder struct {
	namespace string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(namespace string) *CacheKeyBuilder {
	return &CacheKeyBuilder{namespace: namespace}
}

// Build creates a cache key from parts
func (ckb *CacheKeyBuilder) Build(parts ...interface{}) string {
	key := ckb.namespace
	for _, part := range parts {
		key = fmt.Sprintf("%s:%v", key, part)
	}
	return key
}

// UserKey builds a cache key for user data
func (ckb *CacheKeyBuilder) UserKey(userID string) string {
	return ckb.Build("user", userID)
}

// TenantKey builds a cache key for tenant data
func (ckb *CacheKeyBuilder) TenantKey(tenantID string) string {
	return ckb.Build("tenant", tenantID)
}

// QueryKey builds a cache key for query results
func (ckb *CacheKeyBuilder) QueryKey(query string, params ...interface{}) string {
	return ckb.Build("query", query, params)
}

// SessionKey builds a cache key for session data
func (ckb *CacheKeyBuilder) SessionKey(sessionID string) string {
	return ckb.Build("session", sessionID)
}

// TTLStrategy defines TTL strategies for different data types
type TTLStrategy struct {
	ShortTTL  time.Duration // For frequently changing data (1-5 minutes)
	MediumTTL time.Duration // For moderately stable data (5-60 minutes)
	LongTTL   time.Duration // For stable data (1-24 hours)
}

// DefaultTTLStrategy returns default TTL settings
func DefaultTTLStrategy() *TTLStrategy {
	return &TTLStrategy{
		ShortTTL:  2 * time.Minute,
		MediumTTL: 15 * time.Minute,
		LongTTL:   1 * time.Hour,
	}
}

// CacheTags allows cache invalidation by tags
type CacheTags struct {
	cache Cache
}

// NewCacheTags creates a new cache tags manager
func NewCacheTags(cache Cache) *CacheTags {
	return &CacheTags{cache: cache}
}

// Set stores a value with tags
func (ct *CacheTags) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags ...string) error {
	// Store the main value
	if err := ct.cache.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Store tag associations
	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		// Add key to tag set (simplified - in production use Redis SADD)
		// This is a simplified version
		ct.cache.Set(ctx, fmt.Sprintf("%s:%s", tagKey, key), true, ttl)
	}

	return nil
}

// InvalidateByTag removes all cache entries with a specific tag
func (ct *CacheTags) InvalidateByTag(ctx context.Context, tag string) error {
	// In production, use Redis SMEMBERS to get all keys with this tag
	// Then delete them all
	// This is a simplified placeholder
	tagKey := fmt.Sprintf("tag:%s", tag)
	return ct.cache.Delete(ctx, tagKey)
}

// RateLimiter implements rate limiting using cache
type RateLimiter struct {
	cache Cache
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cache Cache) *RateLimiter {
	return &RateLimiter{cache: cache}
}

// Allow checks if request is allowed based on rate limit
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	current, err := rl.cache.Increment(ctx, key)
	if err != nil {
		return false, err
	}

	if current == 1 {
		// First request in this window, set expiration
		if err := rl.cache.Expire(ctx, key, window); err != nil {
			return false, err
		}
	}

	return current <= limit, nil
}

// CacheStats provides cache statistics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	HitRate     float64
	Size        int64
	KeyCount    int64
	MemoryUsage int64
}

// GetStats returns cache statistics (Redis specific)
func (rc *RedisCache) GetStats(ctx context.Context) (*CacheStats, error) {
	info, err := rc.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	// Parse Redis INFO output
	// This is simplified - in production, parse the actual INFO output
	stats := &CacheStats{
		Hits:   0,
		Misses: 0,
	}

	// Calculate hit rate
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	// Get key count
	dbSize, err := rc.client.DBSize(ctx).Result()
	if err == nil {
		stats.KeyCount = dbSize
	}

	fmt.Println(info) // For debugging

	return stats, nil
}

// WarmUp preloads cache with frequently accessed data
func (cm *CacheManager) WarmUp(ctx context.Context, loaders map[string]func() (interface{}, error)) error {
	for key, loader := range loaders {
		value, err := loader()
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", key, err)
		}

		if err := cm.cache.Set(ctx, key, value, 1*time.Hour); err != nil {
			return fmt.Errorf("failed to cache %s: %w", key, err)
		}
	}

	return nil
}
