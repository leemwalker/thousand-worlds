package auth_test

import (
	"context"
	"testing"
	"time"

	"tw-backend/internal/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_Allow(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	rl := auth.NewRateLimiter(client)
	ctx := context.Background()

	t.Run("allows requests within limit", func(t *testing.T) {
		key := "test:allow:" + uuid.New().String()
		limit := 5
		window := 1 * time.Minute

		for i := 0; i < limit; i++ {
			allowed, err := rl.Allow(ctx, key, limit, window)
			require.NoError(t, err)
			assert.True(t, allowed, "request %d should be allowed", i+1)
		}
	})

	t.Run("blocks requests exceeding limit", func(t *testing.T) {
		key := "test:block:" + uuid.New().String()
		limit := 3
		window := 1 * time.Minute

		// Consume limit
		for i := 0; i < limit; i++ {
			allowed, err := rl.Allow(ctx, key, limit, window)
			require.NoError(t, err)
			require.True(t, allowed)
		}

		// Exceed limit
		allowed, err := rl.Allow(ctx, key, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed, "request exceeding limit should be blocked")
	})

	t.Run("resets after window", func(t *testing.T) {
		key := "test:reset:" + uuid.New().String()
		limit := 2
		window := 1 * time.Second // Short window for test

		// Consume limit
		rl.Allow(ctx, key, limit, window)
		rl.Allow(ctx, key, limit, window)

		// Verify blocked
		allowed, err := rl.Allow(ctx, key, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed)

		// Wait for window to expire
		time.Sleep(1100 * time.Millisecond)

		// Should be allowed again
		allowed, err = rl.Allow(ctx, key, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed, "request should be allowed after window reset")
	})
}
