package effects

import (
	"time"

	"github.com/google/uuid"
)

const (
	BleedDamagePerTick = 5
	BleedTickInterval  = 3 * time.Second
	BleedDuration      = 20 * time.Second
)

// ApplyBleed adds bleed
func (m *EffectManager) ApplyBleed(targetID uuid.UUID, now time.Time) {
	// Prompt doesn't specify stacking for bleed, usually just refreshes or separate stacks.
	// "Bleed applies 5 damage per 3 seconds" - implies single instance for simplicity unless specified.
	// "Encourages stationary combat"

	// Let's assume refresh for now to keep it simple, or new instance replaces.
	m.Bleed = &BleedEffect{
		EffectID:        uuid.New(),
		TargetID:        targetID,
		DamagePerTick:   BleedDamagePerTick,
		TickInterval:    BleedTickInterval,
		Duration:        BleedDuration,
		AppliedAt:       now,
		LastTickAt:      now,
		MovementCounter: 0,
	}
}

// OnMovement handles movement reduction logic
func (m *EffectManager) OnMovement() {
	if m.Bleed == nil {
		return
	}

	m.Bleed.MovementCounter++
	if m.Bleed.MovementCounter >= 3 {
		m.Bleed.MovementCounter = 0
		m.Bleed.DamagePerTick--
		if m.Bleed.DamagePerTick < 0 {
			m.Bleed.DamagePerTick = 0
		}
	}
}

// TickBleed processes bleed damage
func (e *BleedEffect) Tick(now time.Time) int {
	if e.DamagePerTick <= 0 {
		return 0
	}
	if now.Sub(e.LastTickAt) >= e.TickInterval {
		e.LastTickAt = now
		return e.DamagePerTick
	}
	return 0
}

// IsExpired checks if bleed has run its course or healed
func (e *BleedEffect) IsExpired(now time.Time) bool {
	if e.DamagePerTick <= 0 {
		return true
	}
	return now.Sub(e.AppliedAt) >= e.Duration
}
