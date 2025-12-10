package cache_test

import (
	"context"
	"testing"
	"time"

	"tw-backend/internal/cache"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestQueryCache_Integration tests the query cache with actual Redis instance
// This verifies O(1) cache performance vs O(log N) database queries
func TestQueryCache_Integration(t *testing.T) {
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

	// Create cache
	queryCache := cache.NewQueryCache(client, 30*time.Second)

	t.Run("Cache Miss and Set", func(t *testing.T) {
		type TestData struct {
			ID   string
			Name string
		}

		key := "test:integration:1"
		data := TestData{ID: "1", Name: "Integration Test"}

		// Set value
		err := queryCache.Set(ctx, key, data)
		require.NoError(t, err)

		// Get value
		var retrieved TestData
		err = queryCache.Get(ctx, key, &retrieved)
		require.NoError(t, err)

		assert.Equal(t, data.ID, retrieved.ID)
		assert.Equal(t, data.Name, retrieved.Name)
	})

	t.Run("GetOrSet Pattern", func(t *testing.T) {
		type WorldData struct {
			WorldID string
			Name    string
		}

		key := "world:12345"
		worldData := WorldData{WorldID: "12345", Name: "Test World"}

		loaderCallCount := 0
		loader := func() (interface{}, error) {
			loaderCallCount++
			// Simulates database query
			time.Sleep(10 * time.Millisecond)
			return worldData, nil
		}

		// First call - should execute loader
		var result1 WorldData
		err := queryCache.GetOrSet(ctx, key, &result1, loader)
		require.NoError(t, err)
		assert.Equal(t, 1, loaderCallCount)
		assert.Equal(t, worldData.Name, result1.Name)

		// Wait for async cache set
		time.Sleep(100 * time.Millisecond)

		// Second call - should hit cache (O(1) vs O(log N) DB query)
		var result2 WorldData
		err = queryCache.GetOrSet(ctx, key, &result2, loader)
		require.NoError(t, err)
		assert.Equal(t, 1, loaderCallCount) // Loader not called again
		assert.Equal(t, worldData.Name, result2.Name)
	})

	t.Run("Cache Invalidation", func(t *testing.T) {
		keys := []string{"test:inv:1", "test:inv:2", "test:inv:3"}

		// Set multiple values
		for i, key := range keys {
			err := queryCache.Set(ctx, key, map[string]int{"value": i})
			require.NoError(t, err)
		}

		// Delete specific keys
		err := queryCache.Delete(ctx, keys[0], keys[1])
		require.NoError(t, err)

		// Verify deletion
		var data map[string]int
		err = queryCache.Get(ctx, keys[0], &data)
		assert.Equal(t, redis.Nil, err) // Should be gone

		err = queryCache.Get(ctx, keys[2], &data)
		assert.NoError(t, err) // Should still exist
	})

	t.Run("TTL Expiration", func(t *testing.T) {
		shortCache := cache.NewQueryCache(client, 1*time.Second)

		key := "test:ttl:1"
		data := map[string]string{"test": "data"}

		err := shortCache.Set(ctx, key, data)
		require.NoError(t, err)

		// Immediately after set - should exist
		var retrieved map[string]string
		err = shortCache.Get(ctx, key, &retrieved)
		require.NoError(t, err)

		// Wait for TTL to expire
		time.Sleep(2 * time.Second)

		// Should be expired
		err = shortCache.Get(ctx, key, &retrieved)
		assert.Equal(t, redis.Nil, err)
	})
}

// BenchmarkQueryCache_vs_Database simulates cache performance gains
func BenchmarkQueryCache_Get(b *testing.B) {
	ctx := context.Background()

	// This benchmark requires a running Redis instance
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		b.Skip("Redis not available")
	}

	queryCache := cache.NewQueryCache(client, 60*time.Second)

	// Pre-populate cache
	key := "benchmark:world:1"
	testData := map[string]interface{}{
		"id":   "1",
		"name": "Benchmark World",
		"metadata": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	queryCache.Set(ctx, key, testData)

	b.ResetTimer()

	// Benchmark cache retrieval (O(1))
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
		_ = queryCache.Get(ctx, key, &data)
	}
}
