package action

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCanQueueAction_ValidAction(t *testing.T) {
	now := time.Now()
	combatant := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second), // Long enough ago
		CurrentAction:  nil,
		CombatState:    StateInCombat,
		StatusEffects:  []StatusEffect{},
	}

	err := CanQueueAction(combatant, ActionAttack, AttackNormal, now)
	assert.NoError(t, err)
}

func TestCanQueueAction_ActionInProgress(t *testing.T) {
	now := time.Now()
	currentAction := &CombatAction{
		ActionID:  uuid.New(),
		ActorID:   uuid.New(),
		TargetID:  uuid.New(),
		Resolved:  false,
		QueuedAt:  now.Add(-500 * time.Millisecond),
		ExecuteAt: now.Add(500 * time.Millisecond),
	}

	combatant := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CurrentAction:  currentAction,
		CombatState:    StateInCombat,
		StatusEffects:  []StatusEffect{},
	}

	err := CanQueueAction(combatant, ActionAttack, AttackNormal, now)
	assert.Error(t, err)
	assert.Equal(t, ErrActionInProgress, err)
}

func TestCanQueueAction_TooSoon(t *testing.T) {
	now := time.Now()
	combatant := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-100 * time.Millisecond), // Only 100ms ago, need 200ms
		CurrentAction:  nil,
		CombatState:    StateInCombat,
		StatusEffects:  []StatusEffect{},
	}

	err := CanQueueAction(combatant, ActionAttack, AttackNormal, now)
	assert.Error(t, err)
	assert.Equal(t, ErrActionTooSoon, err)
}

func TestCanQueueAction_Stunned(t *testing.T) {
	now := time.Now()
	combatant := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CurrentAction:  nil,
		CombatState:    StateInCombat,
		StatusEffects: []StatusEffect{
			{EffectType: EffectStun, ExpiresAt: now.Add(2 * time.Second)},
		},
	}

	err := CanQueueAction(combatant, ActionAttack, AttackNormal, now)
	assert.Error(t, err)
	assert.Equal(t, ErrStunned, err)
}

func TestCanQueueAction_InsufficientStamina(t *testing.T) {
	now := time.Now()
	combatant := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 5, // Not enough for normal attack (15)
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CurrentAction:  nil,
		CombatState:    StateInCombat,
		StatusEffects:  []StatusEffect{},
	}

	err := CanQueueAction(combatant, ActionAttack, AttackNormal, now)
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientStamina, err)
}

func TestGetStaminaCost_AllActionTypes(t *testing.T) {
	tests := []struct {
		name          string
		actionType    ActionType
		attackVariant AttackType
		expectedCost  int
	}{
		{"Quick Attack", ActionAttack, AttackQuick, 10},
		{"Normal Attack", ActionAttack, AttackNormal, 15},
		{"Heavy Attack", ActionAttack, AttackHeavy, 25},
		{"Defend", ActionDefend, "", 5},
		{"Flee", ActionFlee, "", 20},
		{"Use Item", ActionUseItem, "", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := GetStaminaCost(tt.actionType, tt.attackVariant)
			assert.Equal(t, tt.expectedCost, cost)
		})
	}
}

func TestCanQueueAction_EdgeCaseExactly200ms(t *testing.T) {
	now := time.Now()
	combatant := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-200 * time.Millisecond), // Exactly 200ms
		CurrentAction:  nil,
		CombatState:    StateInCombat,
		StatusEffects:  []StatusEffect{},
	}

	err := CanQueueAction(combatant, ActionAttack, AttackNormal, now)
	assert.NoError(t, err, "Exactly 200ms should be allowed")
}

func TestCanQueueAction_ResolvedActionAllowed(t *testing.T) {
	now := time.Now()
	resolvedAction := &CombatAction{
		ActionID:  uuid.New(),
		ActorID:   uuid.New(),
		TargetID:  uuid.New(),
		Resolved:  true, // Already resolved
		QueuedAt:  now.Add(-2 * time.Second),
		ExecuteAt: now.Add(-1 * time.Second),
	}

	combatant := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CurrentAction:  resolvedAction, // Has action but it's resolved
		CombatState:    StateInCombat,
		StatusEffects:  []StatusEffect{},
	}

	err := CanQueueAction(combatant, ActionAttack, AttackNormal, now)
	assert.NoError(t, err, "Resolved actions should not block new actions")
}

func TestIsStunned_ActiveStun(t *testing.T) {
	now := time.Now()
	combatant := &Combatant{
		StatusEffects: []StatusEffect{
			{EffectType: EffectStun, ExpiresAt: now.Add(1 * time.Second)},
		},
	}

	assert.True(t, IsStunned(combatant, now))
}

func TestIsStunned_ExpiredStun(t *testing.T) {
	now := time.Now()
	combatant := &Combatant{
		StatusEffects: []StatusEffect{
			{EffectType: EffectStun, ExpiresAt: now.Add(-1 * time.Second)}, // Expired
		},
	}

	assert.False(t, IsStunned(combatant, now))
}

func TestIsStunned_NoStun(t *testing.T) {
	now := time.Now()
	combatant := &Combatant{
		StatusEffects: []StatusEffect{
			{EffectType: EffectSlow, ExpiresAt: now.Add(1 * time.Second)},
		},
	}

	assert.False(t, IsStunned(combatant, now))
}
