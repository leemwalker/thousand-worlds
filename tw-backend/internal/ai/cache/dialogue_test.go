package cache

import (
	"testing"
	"time"
)

func TestDialogueCache(t *testing.T) {
	c := NewDialogueCache(100 * time.Millisecond)

	key := GenerateKey("npc1", "speaker1", "topic", "hash")
	c.Set(key, "Response 1")

	// Hit
	val, ok := c.Get(key)
	if !ok || val != "Response 1" {
		t.Error("Cache miss or wrong value")
	}

	// Expiry
	time.Sleep(150 * time.Millisecond)
	_, ok = c.Get(key)
	if ok {
		t.Error("Cache should have expired")
	}
}

func TestGenerateContextHash(t *testing.T) {
	h1 := GenerateContextHash("happy", "food", 50, 50, 0, "None")
	h2 := GenerateContextHash("happy", "food", 50, 50, 0, "None")
	h3 := GenerateContextHash("sad", "food", 50, 50, 0, "None")

	if h1 != h2 {
		t.Error("Hash should be deterministic")
	}
	if h1 == h3 {
		t.Error("Different context should produce different hash")
	}
}
