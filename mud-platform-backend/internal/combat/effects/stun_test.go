package effects

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStunExtension(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	// Apply 2s stun
	m.ApplyStun(targetID, 2*time.Second, now)
	if !m.IsStunned(now) {
		t.Error("Expected stunned")
	}

	// Advance 1s
	future := now.Add(1 * time.Second)

	// Apply another 2s stun
	// Should extend from max(endsAt, now) -> endsAt was now+2s.
	// New endsAt = (now+2s) + 2s = now+4s.
	// Remaining duration from 'future' (now+1s) should be 3s.
	m.ApplyStun(targetID, 2*time.Second, future)

	expectedEnd := now.Add(4 * time.Second)
	if !m.Stun.EndsAt.Equal(expectedEnd) {
		t.Errorf("Expected end at %v, got %v", expectedEnd, m.Stun.EndsAt)
	}
}

func TestStunCap(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	// Apply huge stun
	m.ApplyStun(targetID, 20*time.Second, now)

	// Should be capped at 10s
	expectedEnd := now.Add(10 * time.Second)
	if m.Stun.EndsAt.After(expectedEnd) {
		t.Error("Stun exceeded max duration cap")
	}
}
