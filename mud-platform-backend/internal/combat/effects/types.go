package effects

import (
	"time"

	"github.com/google/uuid"
)

// EffectType identifies the kind of effect
type EffectType string

const (
	EffectPoison EffectType = "poison"
	EffectStun   EffectType = "stun"
	EffectSlow   EffectType = "slow"
	EffectBleed  EffectType = "bleed"
	EffectBuff   EffectType = "buff"
	EffectDebuff EffectType = "debuff"
)

// PoisonEffect represents a damage-over-time poison
type PoisonEffect struct {
	EffectID      uuid.UUID
	TargetID      uuid.UUID
	Stacks        int           // 1-5
	DamagePerTick int           // 3 damage per stack
	TickInterval  time.Duration // 5 seconds
	Duration      time.Duration // 30 seconds
	AppliedAt     time.Time
	LastTickAt    time.Time
}

// StunEffect prevents actions
type StunEffect struct {
	EffectID  uuid.UUID
	TargetID  uuid.UUID
	Duration  time.Duration // Base 2 seconds
	AppliedAt time.Time
	EndsAt    time.Time
}

// SlowEffect increases reaction time
type SlowEffect struct {
	EffectID   uuid.UUID
	TargetID   uuid.UUID
	Multiplier float64       // 1.5 = 50% slower
	Duration   time.Duration // 15 seconds
	AppliedAt  time.Time
}

// BleedEffect deals damage over time, reduced by movement
type BleedEffect struct {
	EffectID        uuid.UUID
	TargetID        uuid.UUID
	DamagePerTick   int           // 5 damage
	TickInterval    time.Duration // 3 seconds
	Duration        time.Duration // 20 seconds
	AppliedAt       time.Time
	LastTickAt      time.Time
	MovementCounter int
}

// StatModifier represents a temporary stat change
type StatModifier struct {
	EffectID  uuid.UUID
	TargetID  uuid.UUID
	Stat      string // e.g., "Might", "Agility"
	Modifier  int    // +10, -15
	IsPercent bool   // True = percentage, false = flat
	Duration  time.Duration
	AppliedAt time.Time
}

// EffectManager manages active effects for an entity
type EffectManager struct {
	Poison    *PoisonEffect
	Stun      *StunEffect
	Slow      *SlowEffect
	Bleed     *BleedEffect
	Modifiers []*StatModifier
}
