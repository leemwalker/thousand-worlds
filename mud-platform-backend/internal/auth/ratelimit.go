package auth

import (
	"context"
	"time"

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
