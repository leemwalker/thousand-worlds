package effects

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSlowReplacement(t *testing.T) {
	m := NewManager()
	targetID := uuid.New()
	now := time.Now()

	// Apply 1.5x
	m.ApplySlow(targetID, 1.5, now)
	if m.GetSpeedMultiplier(now) != 1.5 {
		t.Error("Expected 1.5x")
	}

	// Apply 2.0x (should replace)
	m.ApplySlow(targetID, 2.0, now)
	if m.GetSpeedMultiplier(now) != 2.0 {
		t.Error("Expected 2.0x")
	}
}
