package gamemap

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheMetrics_HitMiss verifies that cache operations increment the correct metrics.
func TestCacheMetrics_HitMiss(t *testing.T) {
	// Reset metrics for clean test (metrics are global singletons)
	// We'll read metrics before and after to measure delta

	// Create cache with short TTL for testing
	cache := NewMapCache(5 * time.Second)

	// Record initial metric values
	initialHits := getCounterValue(t, metricCacheHits)
	initialMisses := getCounterValue(t, metricCacheMisses)

	// Test 1: Cache miss on empty cache
	_, found := cache.Get("nonexistent_key")
	assert.False(t, found, "Should not find key in empty cache")

	// Verify cache miss metric incremented
	afterMiss := getCounterValue(t, metricCacheMisses)
	assert.Equal(t, initialMisses+1, afterMiss, "Cache miss counter should increment")

	// Test 2: Set a value
	testData := []byte("test image data")
	cache.Set("test_key", testData)
	cache.Wait() // Wait for async set to complete

	// Test 3: Cache hit
	data, found := cache.Get("test_key")
	assert.True(t, found, "Should find key after set")
	assert.Equal(t, testData, data, "Data should match")

	// Verify cache hit metric incremented
	afterHit := getCounterValue(t, metricCacheHits)
	assert.Equal(t, initialHits+1, afterHit, "Cache hit counter should increment")

	// Test 4: Second cache hit
	_, found = cache.Get("test_key")
	assert.True(t, found, "Should find key on second access")
	afterHit2 := getCounterValue(t, metricCacheHits)
	assert.Equal(t, initialHits+2, afterHit2, "Cache hit counter should increment again")
}

// TestCacheMetrics_TTLExpiration verifies that expired entries count as misses.
func TestCacheMetrics_TTLExpiration(t *testing.T) {
	// Create cache with very short TTL
	cache := NewMapCache(50 * time.Millisecond)

	initialMisses := getCounterValue(t, metricCacheMisses)

	// Set a value
	cache.Set("expiring_key", []byte("temp data"))
	cache.Wait() // Wait for async set

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Access should count as miss (TTL expired)
	_, found := cache.Get("expiring_key")
	assert.False(t, found, "Should not find expired key")

	afterExpiredMiss := getCounterValue(t, metricCacheMisses)
	assert.Equal(t, initialMisses+1, afterExpiredMiss, "Expired key access should count as miss")
}

// getCounterValue extracts the current value from a Prometheus Counter.
func getCounterValue(t *testing.T, counter prometheus.Counter) float64 {
	t.Helper()

	var m dto.Metric
	err := counter.Write(&m)
	require.NoError(t, err, "Failed to read metric")

	if m.Counter != nil {
		return *m.Counter.Value
	}
	return 0
}
