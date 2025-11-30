package effects

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStatCalculation(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	// Base 60
	// +10 Flat
	// +20% Percent

	m.AddModifier(targetID, "Might", 10, false, 60*time.Second, now)
	m.AddModifier(targetID, "Might", 20, true, 60*time.Second, now)

	val := m.CalculateStat("Might", 60, now)

	// (60 + 10) * 1.2 = 84
	if val != 84 {
		t.Errorf("Expected 84, got %d", val)
	}
}

func TestModifierCancellation(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	m.AddModifier(targetID, "Might", 20, false, 60*time.Second, now)
	m.AddModifier(targetID, "Might", -15, false, 60*time.Second, now)

	val := m.CalculateStat("Might", 50, now)

	// 50 + 20 - 15 = 55
	if val != 55 {
		t.Errorf("Expected 55, got %d", val)
	}
}
