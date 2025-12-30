package gamemap

import (
	"fmt"
	"sync"
	"time"
)

// MapCacheEntry represents a cached map render.
type MapCacheEntry struct {
	ImageBytes []byte
	Timestamp  time.Time
}

// MapCache handles caching of rendered map images with TTL.
type MapCache struct {
	entries sync.Map // key string -> MapCacheEntry
	ttl     time.Duration
}

// NewMapCache creates a cache with the specified Time-To-Live.
func NewMapCache(ttl time.Duration) *MapCache {
	c := &MapCache{
		ttl: ttl,
	}
	// Start background eviction
	go c.evictionLoop()
	return c
}

// GetCacheKey generates a unique key for the cache.
func GetCacheKey(worldID string, width, height int, version int64) string {
	return fmt.Sprintf("%s:%dx%d:%d", worldID, width, height, version)
}

// Get retrieves an entry from the cache.
func (c *MapCache) Get(key string) ([]byte, bool) {
	val, ok := c.entries.Load(key)
	if !ok {
		return nil, false
	}
	entry := val.(MapCacheEntry)

	// Double check TTL (though eviction loop generally handles this)
	if time.Since(entry.Timestamp) > c.ttl {
		c.entries.Delete(key)
		return nil, false
	}

	return entry.ImageBytes, true
}

// Set adds or updates an entry in the cache.
func (c *MapCache) Set(key string, data []byte) {
	// Store a copy? For read-only bytes like images, we might just store the slice
	// assuming the source buffer isn't reused immediately in a way that races.
	// However, since we use buffer pools, we MUST make a copy here for the cache
	// because the original buffer will be returned to the pool and overwritten.

	cacheCopy := make([]byte, len(data))
	copy(cacheCopy, data)

	c.entries.Store(key, MapCacheEntry{
		ImageBytes: cacheCopy,
		Timestamp:  time.Now(),
	})
}

func (c *MapCache) evictionLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		c.entries.Range(func(key, value interface{}) bool {
			entry := value.(MapCacheEntry)
			if now.Sub(entry.Timestamp) > c.ttl {
				c.entries.Delete(key)
			}
			return true
		})
	}
}
