package effects

import (
	"time"

	"github.com/google/uuid"
)

const (
	MaxPoisonStacks    = 5
	PoisonBaseDamage   = 3
	PoisonTickInterval = 5 * time.Second
	PoisonDuration     = 30 * time.Second
)

// ApplyPoison adds or updates poison on the manager
func (m *EffectManager) ApplyPoison(targetID uuid.UUID, now time.Time) {
	if m.Poison == nil {
		m.Poison = &PoisonEffect{
			EffectID:      uuid.New(),
			TargetID:      targetID,
			Stacks:        1,
			DamagePerTick: PoisonBaseDamage,
			TickInterval:  PoisonTickInterval,
			Duration:      PoisonDuration,
			AppliedAt:     now,
			LastTickAt:    now,
		}
	} else {
		// Stack up to max
		if m.Poison.Stacks < MaxPoisonStacks {
			m.Poison.Stacks++
			m.Poison.DamagePerTick = PoisonBaseDamage * m.Poison.Stacks
		}
		// Refresh duration
		m.Poison.Duration = PoisonDuration
		m.Poison.AppliedAt = now
	}
}

// TickPoison processes poison damage
// Returns damage amount if tick occurred, 0 otherwise
func (e *PoisonEffect) Tick(now time.Time) int {
	if now.Sub(e.LastTickAt) >= e.TickInterval {
		e.LastTickAt = now
		return e.DamagePerTick
	}
	return 0
}

// IsExpired checks if poison has run its course
func (e *PoisonEffect) IsExpired(now time.Time) bool {
	return now.Sub(e.AppliedAt) >= e.Duration
}
