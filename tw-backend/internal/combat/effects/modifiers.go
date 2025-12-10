package effects

import (
	"time"

	"github.com/google/uuid"
)

// AddModifier adds a buff or debuff
func (m *EffectManager) AddModifier(targetID uuid.UUID, stat string, val int, isPercent bool, duration time.Duration, now time.Time) {
	mod := &StatModifier{
		EffectID:  uuid.New(),
		TargetID:  targetID,
		Stat:      stat,
		Modifier:  val,
		IsPercent: isPercent,
		Duration:  duration,
		AppliedAt: now,
	}
	m.Modifiers = append(m.Modifiers, mod)
}

// CalculateStat applies all active modifiers to a base stat value
func (m *EffectManager) CalculateStat(stat string, baseValue int, now time.Time) int {
	flatMod := 0
	percentMod := 0.0

	for _, mod := range m.Modifiers {
		// Check expiration
		if now.Sub(mod.AppliedAt) >= mod.Duration {
			continue
		}

		if mod.Stat == stat {
			if mod.IsPercent {
				percentMod += float64(mod.Modifier) / 100.0
			} else {
				flatMod += mod.Modifier
			}
		}
	}

	// Order: base -> flat -> percent
	// Example: 60 + 10 = 70. 70 * (1 + 0.2) = 84.
	val := float64(baseValue + flatMod)
	val = val * (1.0 + percentMod)

	return int(val)
}
