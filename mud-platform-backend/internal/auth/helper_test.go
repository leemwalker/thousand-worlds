package auth

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	return client
}
