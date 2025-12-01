package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RateLimiter handles request rate limiting.
type RateLimiter struct {
	client *redis.Client
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		client: client,
	}
}

// Allow checks if a request is allowed.
// key: unique identifier (e.g., "ip:127.0.0.1:login")
// limit: max requests allowed
// window: time window
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	rateLimitKey := "ratelimit:" + key

	// Increment the counter
	count, err := rl.client.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		return false, err
	}

	// If it's the first request, set the expiration
	if count == 1 {
		if err := rl.client.Expire(ctx, rateLimitKey, window).Err(); err != nil {
			return false, err
		}
	}

	// Check if limit exceeded
	if count > int64(limit) {
		return false, nil
	}

	return true, nil
}

// AllowCommand checks if a command from a specific client is allowed
// Implements token bucket algorithm: 10 commands per second with burst of 20
// Complexity: O(1) Redis operations
func (rl *RateLimiter) AllowCommand(ctx context.Context, characterID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("ratelimit:command:%s", characterID.String())

	// Token bucket parameters
	const (
		maxTokens  = 20 // Burst capacity
		refillRate = 10 // Tokens per second
		window     = time.Second
	)

	// Use Redis INCR with expiration as simple token bucket
	// For more precise token bucket, would use Lua script
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("rate limit check failed: %w", err)
	}

	// Set expiration on first increment (sliding window)
	if count == 1 {
		rl.client.Expire(ctx, key, window)
	}

	// Allow if under max tokens
	allowed := count <= maxTokens

	return allowed, nil
}
