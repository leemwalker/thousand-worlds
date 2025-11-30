package area

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AreaCache stores generated area descriptions
type AreaCache struct {
	store map[string]cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type cacheEntry struct {
	description string
	expiresAt   time.Time
}

// NewAreaCache creates a new cache with default TTL of 60 minutes
func NewAreaCache() *AreaCache {
	return &AreaCache{
		store: make(map[string]cacheEntry),
		ttl:   60 * time.Minute,
	}
}

// Get retrieves a description if valid
func (c *AreaCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.store[key]
	if !ok {
		return "", false
	}

	if time.Now().After(entry.expiresAt) {
		return "", false
	}

	return entry.description, true
}

// Set stores a description
func (c *AreaCache) Set(key, description string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = cacheEntry{
		description: description,
		expiresAt:   time.Now().Add(c.ttl),
	}
}

// GenerateKey creates a cache key from context factors
func GenerateKey(worldID uuid.UUID, x, y, z float64, weather, timeOfDay, season string, perception int) string {
	// Bucket perception: 0-25, 26-50, 51-75, 76-100
	perceptionBucket := 0
	if perception > 75 {
		perceptionBucket = 3
	} else if perception > 50 {
		perceptionBucket = 2
	} else if perception > 25 {
		perceptionBucket = 1
	}

	raw := fmt.Sprintf("area:%s:%.2f:%.2f:%.2f:%s:%s:%s:%d",
		worldID, x, y, z, weather, timeOfDay, season, perceptionBucket)

	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// Invalidate clears cache for a specific location (if needed, or we just let TTL expire)
// For major world events, we might want to invalidate.
func (c *AreaCache) Invalidate(worldID uuid.UUID, x, y, z float64) {
	// This is tricky because the key is hashed and includes other factors.
	// A real implementation might use a prefix or tag-based invalidation.
	// For MVP, we'll skip complex invalidation and rely on TTL or manual clear if needed.
}
