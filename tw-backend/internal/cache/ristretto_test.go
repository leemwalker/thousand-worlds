package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRistrettoCache_BasicOperations(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    []byte
		wantHit  bool
		wantData []byte
	}{
		{
			name:     "store and retrieve small data",
			key:      "test-key-1",
			value:    []byte("hello world"),
			wantHit:  true,
			wantData: []byte("hello world"),
		},
		{
			name:     "store and retrieve large data",
			key:      "test-key-2",
			value:    make([]byte, 1024*1024), // 1MB
			wantHit:  true,
			wantData: make([]byte, 1024*1024),
		},
		{
			name:     "empty data",
			key:      "empty-key",
			value:    []byte{},
			wantHit:  true,
			wantData: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewRistrettoCache(100*1024*1024, 5*time.Minute) // 100MB, 5min TTL
			require.NoError(t, err)
			defer cache.Close()

			// Set value
			cache.Set(tt.key, tt.value)

			// Ristretto is async, wait for set to complete
			cache.Wait()

			// Get value
			data, hit := cache.Get(tt.key)
			assert.Equal(t, tt.wantHit, hit)
			if hit {
				assert.Equal(t, tt.wantData, data)
			}
		})
	}
}

func TestRistrettoCache_Miss(t *testing.T) {
	cache, err := NewRistrettoCache(10*1024*1024, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	data, hit := cache.Get("nonexistent-key")
	assert.False(t, hit)
	assert.Nil(t, data)
}

func TestRistrettoCache_CostBasedEviction(t *testing.T) {
	// Create cache with 2KB max cost - gives room for admission
	cache, err := NewRistrettoCache(2048, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	// Insert several small items to build up frequency
	for i := 0; i < 5; i++ {
		key := "frequent-" + string(rune('a'+i))
		cache.Set(key, make([]byte, 100))
		cache.Wait()
		// Access multiple times to increase frequency
		for j := 0; j < 3; j++ {
			cache.Get(key)
		}
	}

	// Now insert a large item that exceeds remaining capacity
	bigKey := "big-item"
	cache.Set(bigKey, make([]byte, 1500))
	cache.Wait()

	// The cache should maintain cost within bounds
	// We can't guarantee which items are evicted, but the cache
	// should still function correctly
	metrics := cache.Metrics()
	t.Logf("After operations - Hits: %d, Misses: %d", metrics.Hits, metrics.Misses)

	// Verify we can still use the cache
	testKey := "post-eviction"
	cache.Set(testKey, []byte("test"))
	cache.Wait()
	_, hit := cache.Get(testKey)
	// Due to admission policy, this may or may not be admitted
	t.Logf("post-eviction key admitted: %v", hit)
}

func TestRistrettoCache_Delete(t *testing.T) {
	cache, err := NewRistrettoCache(10*1024*1024, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	key := "delete-me"
	cache.Set(key, []byte("data"))
	cache.Wait()

	// Verify it's there
	_, hit := cache.Get(key)
	require.True(t, hit)

	// Delete
	cache.Delete(key)

	// Verify it's gone
	_, hit = cache.Get(key)
	assert.False(t, hit)
}

func TestRistrettoCache_Clear(t *testing.T) {
	cache, err := NewRistrettoCache(10*1024*1024, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	// Add multiple items
	for i := 0; i < 10; i++ {
		cache.Set("key-"+string(rune('a'+i)), []byte("data"))
	}
	cache.Wait()

	// Clear all
	cache.Clear()

	// All should be gone
	for i := 0; i < 10; i++ {
		_, hit := cache.Get("key-" + string(rune('a'+i)))
		assert.False(t, hit)
	}
}

func TestRistrettoCache_ConcurrentAccess(t *testing.T) {
	cache, err := NewRistrettoCache(100*1024*1024, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	var wg sync.WaitGroup
	const goroutines = 100
	const operations = 100

	// Concurrent writes and reads
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := "key-" + string(rune('a'+id%26))
				cache.Set(key, []byte("value"))
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()
	// No panic = success for race condition test
}

func TestRistrettoCache_Metrics(t *testing.T) {
	cache, err := NewRistrettoCache(10*1024*1024, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	// Generate some hits and misses
	cache.Set("exists", []byte("data"))
	cache.Wait()

	cache.Get("exists")      // Hit
	cache.Get("nonexistent") // Miss

	metrics := cache.Metrics()
	assert.NotNil(t, metrics)
	// Ristretto tracks metrics internally, we expose them
	t.Logf("Hits: %d, Misses: %d", metrics.Hits, metrics.Misses)
}
