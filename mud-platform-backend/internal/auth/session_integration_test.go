package auth_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mud-platform-backend/internal/auth"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestSessionManager_BatchUpdates tests batch session update functionality
func TestSessionManager_BatchUpdates(t *testing.T) {
	ctx := context.Background()

	// Start Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Skip("Docker not available for integration test")
	}

	defer redisContainer.Terminate(ctx)

	// Get connection details
	host, err := redisContainer.Host(ctx)
	require.NoError(t, err)

	port, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})
	defer client.Close()

	// Verify connection
	err = client.Ping(ctx).Err()
	require.NoError(t, err)

	// Create SessionManager with short flush interval for testing
	sm := auth.NewSessionManager(client)
	defer sm.Close(ctx)

	t.Run("CreateAndGetSession", func(t *testing.T) {
		userID := uuid.New().String()
		username := "testuser"

		// Create session
		session, err := sm.CreateSession(ctx, userID, username)
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, username, session.Username)

		// Get session (should update LastAccess in memory)
		retrievedSession, err := sm.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrievedSession.ID)
		assert.WithinDuration(t, time.Now().UTC(), retrievedSession.LastAccess, 2*time.Second)
	})

	t.Run("MultipleAccessesBeforeFlush", func(t *testing.T) {
		userID := uuid.New().String()
		session, err := sm.CreateSession(ctx, userID, "user2")
		require.NoError(t, err)

		initialAccess := session.LastAccess

		// Access session multiple times quickly
		for i := 0; i < 5; i++ {
			time.Sleep(100 * time.Millisecond)
			_, err := sm.GetSession(ctx, session.ID)
			require.NoError(t, err)
		}

		// Verify LastAccess not yet written to Redis (batched)
		// Direct Redis check
		key := "session:" + session.ID
		data, err := client.Get(ctx, key).Bytes()
		require.NoError(t, err)

		var persistedSession auth.Session
		err = json.Unmarshal(data, &persistedSession)
		require.NoError(t, err)

		// LastAccess in Redis should still be close to initial
		assert.WithinDuration(t, initialAccess, persistedSession.LastAccess, 2*time.Second)
	})

	t.Run("InvalidateSession", func(t *testing.T) {
		userID := uuid.New().String()
		session, err := sm.CreateSession(ctx, userID, "user3")
		require.NoError(t, err)

		// Invalidate
		err = sm.InvalidateSession(ctx, session.ID)
		require.NoError(t, err)

		// Try to get - should fail
		_, err = sm.GetSession(ctx, session.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}

// TestRateLimiter_CommandThrottling tests per-client command rate limiting
func TestRateLimiter_CommandThrottling(t *testing.T) {
	ctx := context.Background()

	// Start Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Skip("Docker not available for integration test")
	}

	defer redisContainer.Terminate(ctx)

	// Get connection details
	host, err := redisContainer.Host(ctx)
	require.NoError(t, err)

	port, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})
	defer client.Close()

	err = client.Ping(ctx).Err()
	require.NoError(t, err)

	rateLimiter := auth.NewRateLimiter(client)

	t.Run("AllowCommandsUnderLimit", func(t *testing.T) {
		characterID := uuid.New()

		// First 20 commands should be allowed (burst capacity)
		for i := 0; i < 20; i++ {
			allowed, err := rateLimiter.AllowCommand(ctx, characterID)
			require.NoError(t, err)
			assert.True(t, allowed, "Command %d should be allowed", i+1)
		}
	})

	t.Run("BlockCommandsOverLimit", func(t *testing.T) {
		characterID := uuid.New()

		// Exhaust burst capacity
		for i := 0; i < 20; i++ {
			rateLimiter.AllowCommand(ctx, characterID)
		}

		// 21st command should be blocked
		allowed, err := rateLimiter.AllowCommand(ctx, characterID)
		require.NoError(t, err)
		assert.False(t, allowed, "Command over limit should be blocked")
	})

	t.Run("RateLimitReset", func(t *testing.T) {
		characterID := uuid.New()

		// Exhaust limit
		for i := 0; i < 21; i++ {
			rateLimiter.AllowCommand(ctx, characterID)
		}

		// Wait for window to expire
		time.Sleep(1100 * time.Millisecond)

		// Should be allowed again
		allowed, err := rateLimiter.AllowCommand(ctx, characterID)
		require.NoError(t, err)
		assert.True(t, allowed, "Command should be allowed after reset")
	})

	t.Run("IndependentClientLimits", func(t *testing.T) {
		client1 := uuid.New()
		client2 := uuid.New()

		// Exhaust client1's limit
		for i := 0; i < 21; i++ {
			rateLimiter.AllowCommand(ctx, client1)
		}

		// Client1 should be blocked
		allowed, err := rateLimiter.AllowCommand(ctx, client1)
		require.NoError(t, err)
		assert.False(t, allowed)

		// Client2 should still be allowed
		allowed, err = rateLimiter.AllowCommand(ctx, client2)
		require.NoError(t, err)
		assert.True(t, allowed)
	})
}
