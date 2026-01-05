package gamemap

import (
	"fmt"
	"time"

	"tw-backend/internal/cache"
)

// DefaultMapCacheMaxCost is the default maximum cache size (1GB)
const DefaultMapCacheMaxCost = 1 << 30 // 1GB

// MapCache handles caching of rendered map images with cost-based eviction.
// Uses Ristretto for high-performance concurrent caching with LFU eviction.
type MapCache struct {
	cache *cache.RistrettoCache
}

// NewMapCache creates a cache with the specified Time-To-Live and default max size (1GB).
func NewMapCache(ttl time.Duration) *MapCache {
	return NewMapCacheWithMaxCost(DefaultMapCacheMaxCost, ttl)
}

// NewMapCacheWithMaxCost creates a cache with specified max cost and TTL.
// maxCostBytes is the maximum total size of cached data in bytes.
func NewMapCacheWithMaxCost(maxCostBytes int64, ttl time.Duration) *MapCache {
	rc, err := cache.NewRistrettoCache(maxCostBytes, ttl)
	if err != nil {
		// Fall back to smaller cache if creation fails
		rc, _ = cache.NewRistrettoCache(100*1024*1024, ttl) // 100MB fallback
	}

	return &MapCache{
		cache: rc,
	}
}

// GetCacheKey generates a unique key for the cache.
func GetCacheKey(worldID string, width, height int, version int64) string {
	return fmt.Sprintf("%s:%dx%d:%d", worldID, width, height, version)
}

// Get retrieves an entry from the cache.
func (c *MapCache) Get(key string) ([]byte, bool) {
	data, hit := c.cache.Get(key)
	if hit {
		metricCacheHits.Inc()
	} else {
		metricCacheMisses.Inc()
	}
	return data, hit
}

// Set adds or updates an entry in the cache.
// The data is copied internally, so the caller can safely reuse the buffer.
func (c *MapCache) Set(key string, data []byte) {
	c.cache.Set(key, data)
}

// Delete removes an entry from the cache.
func (c *MapCache) Delete(key string) {
	c.cache.Delete(key)
}

// Clear removes all entries from the cache.
func (c *MapCache) Clear() {
	c.cache.Clear()
}

// Metrics returns cache performance statistics.
func (c *MapCache) Metrics() *cache.CacheMetrics {
	return c.cache.Metrics()
}

// Close releases cache resources.
func (c *MapCache) Close() {
	if c.cache != nil {
		c.cache.Close()
	}
}

// Wait blocks until all pending sets are processed.
// Useful in tests to ensure consistency.
func (c *MapCache) Wait() {
	c.cache.Wait()
}
