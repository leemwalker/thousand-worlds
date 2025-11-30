package effects

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPoisonStacking(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	// Apply 1st stack
	m.ApplyPoison(targetID, now)
	if m.Poison.Stacks != 1 {
		t.Errorf("Expected 1 stack, got %d", m.Poison.Stacks)
	}
	if m.Poison.DamagePerTick != 3 {
		t.Errorf("Expected 3 damage, got %d", m.Poison.DamagePerTick)
	}

	// Apply 5 more stacks (should cap at 5)
	for i := 0; i < 5; i++ {
		m.ApplyPoison(targetID, now)
	}
	if m.Poison.Stacks != 5 {
		t.Errorf("Expected 5 stacks, got %d", m.Poison.Stacks)
	}
	if m.Poison.DamagePerTick != 15 {
		t.Errorf("Expected 15 damage, got %d", m.Poison.DamagePerTick)
	}
}

func TestPoisonTick(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	m.ApplyPoison(targetID, now) // 3 damage

	// Tick immediately (should be 0, interval 5s)
	dmg := m.Poison.Tick(now)
	if dmg != 0 {
		t.Error("Expected 0 damage on immediate tick")
	}

	// Tick after 5s
	future := now.Add(5 * time.Second)
	dmg = m.Poison.Tick(future)
	if dmg != 3 {
		t.Errorf("Expected 3 damage, got %d", dmg)
	}
}
