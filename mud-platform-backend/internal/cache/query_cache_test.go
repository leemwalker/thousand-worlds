package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestNewQueryCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cache := NewQueryCache(client, 30*time.Second)

	assert.NotNil(t, cache)
	assert.Equal(t, 30*time.Second, cache.ttl)
}

func TestNewQueryCache_DefaultTTL(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cache := NewQueryCache(client, 0)

	assert.Equal(t, 60*time.Second, cache.ttl)
}

func TestQueryCache_GetSet(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	// Skip if Redis not available
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	cache := NewQueryCache(client, 5*time.Second)
	key := "test:data:123"

	// Clean up
	defer client.Del(ctx, key)

	// Test Set
	data := testData{ID: "123", Name: "Test"}
	err := cache.Set(ctx, key, data)
	require.NoError(t, err)

	// Test Get
	var retrieved testData
	err = cache.Get(ctx, key, &retrieved)
	require.NoError(t, err)
	assert.Equal(t, data.ID, retrieved.ID)
	assert.Equal(t, data.Name, retrieved.Name)
}

func TestQueryCache_GetMiss(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	cache := NewQueryCache(client, 5*time.Second)

	var data testData
	err := cache.Get(ctx, "nonexistent:key", &data)
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

func TestQueryCache_Delete(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	cache := NewQueryCache(client, 5*time.Second)
	key := "test:delete:456"

	// Set a value
	data := testData{ID: "456", Name: "Delete Test"}
	err := cache.Set(ctx, key, data)
	require.NoError(t, err)

	// Delete it
	err = cache.Delete(ctx, key)
	require.NoError(t, err)

	// Verify it's gone
	var retrieved testData
	err = cache.Get(ctx, key, &retrieved)
	assert.Equal(t, redis.Nil, err)
}

func TestQueryCache_GetOrSet(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	cache := NewQueryCache(client, 5*time.Second)
	key := "test:getorset:789"

	defer client.Del(ctx, key)

	loaderCalled := false
	loader := func() (interface{}, error) {
		loaderCalled = true
		return testData{ID: "789", Name: "Loaded"}, nil
	}

	// First call should load
	var data testData
	err := cache.GetOrSet(ctx, key, &data, loader)
	require.NoError(t, err)
	assert.True(t, loaderCalled)
	assert.Equal(t, "789", data.ID)

	// Second call should hit cache
	loaderCalled = false
	var data2 testData

	// Wait a bit for async cache set to complete
	time.Sleep(100 * time.Millisecond)

	err = cache.GetOrSet(ctx, key, &data2, loader)
	require.NoError(t, err)
	assert.False(t, loaderCalled) // Loader should not be called
	assert.Equal(t, "789", data2.ID)
}

func TestQueryCache_GetOrSet_LoaderError(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	cache := NewQueryCache(client, 5*time.Second)

	expectedErr := errors.New("loader failed")
	loader := func() (interface{}, error) {
		return nil, expectedErr
	}

	var data testData
	err := cache.GetOrSet(ctx, "test:error", &data, loader)
	assert.Equal(t, expectedErr, err)
}
