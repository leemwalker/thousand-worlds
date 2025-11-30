package effects

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBleedMovement(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	m.ApplyBleed(targetID, now) // 5 damage

	// Move 3 times
	m.OnMovement()
	m.OnMovement()
	m.OnMovement()

	// Should reduce damage by 1 -> 4
	if m.Bleed.DamagePerTick != 4 {
		t.Errorf("Expected 4 damage, got %d", m.Bleed.DamagePerTick)
	}

	// Reduce to 0
	for i := 0; i < 12; i++ {
		m.OnMovement()
	}

	if m.Bleed.DamagePerTick != 0 {
		t.Error("Expected 0 damage")
	}

	if !m.Bleed.IsExpired(now) {
		t.Error("Expected expired when damage is 0")
	}
}
