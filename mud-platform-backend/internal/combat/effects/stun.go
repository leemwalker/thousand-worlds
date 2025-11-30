package effects

import (
	"time"

	"github.com/google/uuid"
)

const (
	StunBaseDuration = 2 * time.Second
	StunMaxDuration  = 10 * time.Second
)

// ApplyStun adds or extends stun
func (m *EffectManager) ApplyStun(targetID uuid.UUID, duration time.Duration, now time.Time) {
	if m.Stun == nil {
		if duration > StunMaxDuration {
			duration = StunMaxDuration
		}
		m.Stun = &StunEffect{
			EffectID:  uuid.New(),
			TargetID:  targetID,
			Duration:  duration,
			AppliedAt: now,
			EndsAt:    now.Add(duration),
		}
	} else {
		// Extend duration
		// newEndsAt = max(existingEndsAt, now) + newDuration
		baseEnd := m.Stun.EndsAt
		if now.After(baseEnd) {
			baseEnd = now
		}
		newEndsAt := baseEnd.Add(duration)

		// Cap total duration from now
		if newEndsAt.Sub(now) > StunMaxDuration {
			newEndsAt = now.Add(StunMaxDuration)
		}

		m.Stun.EndsAt = newEndsAt
		m.Stun.Duration = newEndsAt.Sub(m.Stun.AppliedAt)
	}
}

// IsStunned checks if stun is active
func (m *EffectManager) IsStunned(now time.Time) bool {
	if m.Stun == nil {
		return false
	}
	return now.Before(m.Stun.EndsAt)
}
