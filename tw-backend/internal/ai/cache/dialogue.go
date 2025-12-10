package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// DialogueCache stores generated dialogues
type DialogueCache struct {
	store map[string]cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type cacheEntry struct {
	response  string
	expiresAt time.Time
}

// NewDialogueCache creates a new cache
func NewDialogueCache(ttl time.Duration) *DialogueCache {
	return &DialogueCache{
		store: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

// Get retrieves a response if valid
func (c *DialogueCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.store[key]
	if !ok {
		return "", false
	}

	if time.Now().After(entry.expiresAt) {
		return "", false
	}

	return entry.response, true
}

// Set stores a response
func (c *DialogueCache) Set(key, response string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = cacheEntry{
		response:  response,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// GenerateKey creates a cache key from context
func GenerateKey(npcID, speakerID, topic string, contextHash string) string {
	raw := fmt.Sprintf("%s:%s:%s:%s", npcID, speakerID, topic, contextHash)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
