package effects

import (
	"time"

	"github.com/google/uuid"
)

const (
	SlowDuration = 15 * time.Second
)

// ApplySlow adds or replaces slow
func (m *EffectManager) ApplySlow(targetID uuid.UUID, multiplier float64, now time.Time) {
	// Newest replaces oldest, or if stronger? Prompt says "newest replaces" in requirements:
	// "Verify only 2.0x active (newest replaces)"

	m.Slow = &SlowEffect{
		EffectID:   uuid.New(),
		TargetID:   targetID,
		Multiplier: multiplier,
		Duration:   SlowDuration,
		AppliedAt:  now,
	}
}

// GetSpeedMultiplier returns the current slow multiplier (default 1.0)
func (m *EffectManager) GetSpeedMultiplier(now time.Time) float64 {
	if m.Slow == nil {
		return 1.0
	}
	if now.Sub(m.Slow.AppliedAt) >= m.Slow.Duration {
		return 1.0 // Expired
	}
	return m.Slow.Multiplier
}
