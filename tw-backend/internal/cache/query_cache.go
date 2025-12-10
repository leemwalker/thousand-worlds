package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// QueryCache implements cache-aside pattern for database query results
// Reduces database load for read-heavy queries from O(log N) to O(1) cache lookup
type QueryCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewQueryCache creates a new query cache with specified TTL
func NewQueryCache(client *redis.Client, ttl time.Duration) *QueryCache {
	if ttl == 0 {
		ttl = 60 * time.Second // Default 60s TTL
	}
	return &QueryCache{
		client: client,
		ttl:    ttl,
	}
}

// Get retrieves a cached value and unmarshals it into the target
// Returns redis.Nil error if key doesn't exist (cache miss)
func (c *QueryCache) Get(ctx context.Context, key string, target interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err // Returns redis.Nil on cache miss
	}

	return json.Unmarshal(data, target)
}

// Set caches a value with the configured TTL
func (c *QueryCache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// Delete invalidates a cached entry
// Used after write operations to maintain cache consistency
func (c *QueryCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// DeletePattern invalidates all keys matching a pattern
// Example: DeletePattern(ctx, "world:*") clears all world caches
func (c *QueryCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// GetOrSet attempts to get from cache, or executes loader function on miss
// This implements the cache-aside pattern automatically
func (c *QueryCache) GetOrSet(ctx context.Context, key string, target interface{}, loader func() (interface{}, error)) error {
	// Try cache first
	err := c.Get(ctx, key, target)
	if err == nil {
		return nil // Cache hit
	}

	if err != redis.Nil {
		// Unexpected error, but continue to loader
		// This provides resilience if Redis is down
	}

	// Cache miss - execute loader
	value, err := loader()
	if err != nil {
		return err
	}

	// Populate cache (fire-and-forget to avoid blocking)
	// Use background context to avoid cancellation propagation
	go func() {
		_ = c.Set(context.Background(), key, value)
	}()

	// Marshal the loaded value into target
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}
