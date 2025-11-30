package action

import (
	"time"

	"github.com/google/uuid"
)

// ActionType represents the type of combat action
type ActionType string

const (
	ActionAttack  ActionType = "attack"
	ActionDefend  ActionType = "defend"
	ActionFlee    ActionType = "flee"
	ActionUseItem ActionType = "use_item"
)

// CombatAction represents a queued action in combat
type CombatAction struct {
	ActionID     uuid.UUID
	ActorID      uuid.UUID
	TargetID     uuid.UUID
	ActionType   ActionType
	ReactionTime time.Duration
	QueuedAt     time.Time
	ExecuteAt    time.Time // QueuedAt + ReactionTime
	Resolved     bool
}

// NewCombatAction creates a new action with calculated execution time
func NewCombatAction(actorID, targetID uuid.UUID, actionType ActionType, reactionTime time.Duration) *CombatAction {
	now := time.Now()
	return &CombatAction{
		ActionID:     uuid.New(),
		ActorID:      actorID,
		TargetID:     targetID,
		ActionType:   actionType,
		ReactionTime: reactionTime,
		QueuedAt:     now,
		ExecuteAt:    now.Add(reactionTime),
		Resolved:     false,
	}
}

// CombatState represents the current state of a combatant
type CombatState string

const (
	StateIdle     CombatState = "idle"
	StateInCombat CombatState = "in_combat"
	StateFleeing  CombatState = "fleeing"
	StateDefeated CombatState = "defeated"
)

// EffectType represents different status effect types
type EffectType string

const (
	EffectStun   EffectType = "stun"
	EffectSlow   EffectType = "slow"
	EffectHaste  EffectType = "haste"
	EffectPoison EffectType = "poison"
	EffectBleed  EffectType = "bleed"
)

// StatusEffect represents a temporary effect on a combatant
type StatusEffect struct {
	EffectType EffectType
	ExpiresAt  time.Time
	Magnitude  float64 // For effects like slow (1.5x) or haste (0.7x)
}

// Combatant represents an entity participating in combat
type Combatant struct {
	EntityID       uuid.UUID
	CurrentStamina int
	MaxStamina     int
	CurrentHP      int
	MaxHP          int
	Agility        int
	LastActionTime time.Time
	CurrentAction  *CombatAction
	DefendingUntil time.Time
	StatusEffects  []StatusEffect
	CombatState    CombatState
}
