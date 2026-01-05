// Package cache provides caching abstractions for the application.
// L1 cache uses Ristretto for high-performance in-process caching with
// cost-based eviction and frequency-based admission policy.
package cache

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

// CacheMetrics contains cache performance statistics.
type CacheMetrics struct {
	Hits   uint64
	Misses uint64
}

// RistrettoCache wraps Ristretto cache with a simplified byte-slice interface.
// It provides cost-based eviction where cost equals the byte length of stored data.
type RistrettoCache struct {
	cache *ristretto.Cache
	ttl   time.Duration
}

// NewRistrettoCache creates a new Ristretto-backed cache.
// maxCost is the maximum total cost (sum of all entry sizes in bytes).
// ttl is the time-to-live for cache entries.
func NewRistrettoCache(maxCost int64, ttl time.Duration) (*RistrettoCache, error) {
	// NumCounters should be ~10x the expected number of items for good frequency tracking
	// BufferItems determines the size of the get buffer
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // 10M counters for frequency tracking
		MaxCost:     maxCost, // Maximum total size in bytes
		BufferItems: 64,      // Number of keys per Get buffer
	})
	if err != nil {
		return nil, err
	}

	return &RistrettoCache{
		cache: cache,
		ttl:   ttl,
	}, nil
}

// Get retrieves an entry from the cache.
// Returns the data and true if found, nil and false otherwise.
func (c *RistrettoCache) Get(key string) ([]byte, bool) {
	val, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	data, ok := val.([]byte)
	if !ok {
		return nil, false
	}

	return data, true
}

// Set stores data in the cache with cost equal to the data length.
// The operation is asynchronous - the entry may not be immediately available.
func (c *RistrettoCache) Set(key string, data []byte) {
	// Make a copy to prevent issues if caller reuses the buffer
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Cost is the size in bytes
	cost := int64(len(dataCopy))
	if cost == 0 {
		cost = 1 // Minimum cost to track empty entries
	}

	// SetWithTTL handles expiration
	c.cache.SetWithTTL(key, dataCopy, cost, c.ttl)
}

// Delete removes an entry from the cache.
func (c *RistrettoCache) Delete(key string) {
	c.cache.Del(key)
}

// Clear removes all entries from the cache.
func (c *RistrettoCache) Clear() {
	c.cache.Clear()
}

// Close releases cache resources.
func (c *RistrettoCache) Close() {
	c.cache.Close()
}

// Metrics returns current cache performance statistics.
func (c *RistrettoCache) Metrics() *CacheMetrics {
	m := c.cache.Metrics
	return &CacheMetrics{
		Hits:   m.Hits(),
		Misses: m.Misses(),
	}
}

// Wait blocks until all pending sets are processed.
// Useful in tests to ensure consistency.
func (c *RistrettoCache) Wait() {
	c.cache.Wait()
}
