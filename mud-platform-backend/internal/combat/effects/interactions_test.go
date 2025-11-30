package effects

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInteractions(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	// Apply Poison, Bleed, Slow
	m.ApplyPoison(targetID, now)
	m.ApplyBleed(targetID, now)
	m.ApplySlow(targetID, 1.5, now)

	// Verify all active
	if m.Poison == nil || m.Bleed == nil || m.Slow == nil {
		t.Error("Expected all effects active")
	}

	// Advance time to trigger ticks
	future := now.Add(5 * time.Second)
	res := m.Tick(future)

	// Poison tick (5s interval) -> 3 dmg
	if res.PoisonDamage != 3 {
		t.Errorf("Expected 3 poison damage, got %d", res.PoisonDamage)
	}
	// Bleed tick (3s interval) -> 5 dmg (triggered at 3s, checked at 5s)
	// Wait, Tick logic checks interval from LastTickAt.
	// If we jump 5s, Bleed (3s interval) should tick once.
	if res.BleedDamage != 5 {
		t.Errorf("Expected 5 bleed damage, got %d", res.BleedDamage)
	}
}
