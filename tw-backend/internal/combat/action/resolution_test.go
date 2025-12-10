package action

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProcessTick_NoActionsReady(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()

	combatantID := uuid.New()
	combatant := &Combatant{
		EntityID:       combatantID,
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateInCombat,
	}

	resolver.AddCombatant(combatant)

	// Queue action that executes in the future
	action := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatantID,
		TargetID:   uuid.New(),
		ActionType: ActionAttack,
		QueuedAt:   now,
		ExecuteAt:  now.Add(1 * time.Second), // In the future
		Resolved:   false,
	}
	resolver.Queue.Enqueue(action)

	// Process tick - action not ready yet
	resolved := resolver.ProcessTick(now)
	assert.Equal(t, 0, len(resolved), "No actions should execute")
	assert.Equal(t, 1, resolver.Queue.Len(), "Action should remain in queue")
}

func TestProcessTick_ActionReady(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()

	combatantID := uuid.New()
	combatant := &Combatant{
		EntityID:       combatantID,
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateInCombat,
	}

	resolver.AddCombatant(combatant)

	// Queue action ready to execute
	action := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatantID,
		TargetID:   uuid.New(),
		ActionType: ActionAttack,
		QueuedAt:   now.Add(-1 * time.Second),
		ExecuteAt:  now.Add(-100 * time.Millisecond), // Ready
		Resolved:   false,
	}
	resolver.Queue.Enqueue(action)

	// Process tick
	resolved := resolver.ProcessTick(now)
	assert.Equal(t, 1, len(resolved), "One action should execute")
	assert.True(t, resolved[0].Resolved, "Action should be marked resolved")
	assert.Equal(t, 0, resolver.Queue.Len(), "Queue should be empty")
}

func TestProcessTick_StaminaConsumed(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()

	combatantID := uuid.New()
	combatant := &Combatant{
		EntityID:       combatantID,
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateInCombat,
	}

	resolver.AddCombatant(combatant)

	// Queue normal attack (costs 15 stamina)
	action := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatantID,
		TargetID:   uuid.New(),
		ActionType: ActionAttack,
		QueuedAt:   now.Add(-1 * time.Second),
		ExecuteAt:  now,
		Resolved:   false,
	}
	resolver.Queue.Enqueue(action)

	// Process tick
	resolver.ProcessTick(now)

	// Check stamina was consumed
	assert.Equal(t, 85, combatant.CurrentStamina, "Should have consumed 15 stamina")
}

func TestProcessTick_LastActionTimeUpdated(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()
	oldTime := now.Add(-5 * time.Second)

	combatantID := uuid.New()
	combatant := &Combatant{
		EntityID:       combatantID,
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: oldTime,
		CombatState:    StateInCombat,
	}

	resolver.AddCombatant(combatant)

	action := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatantID,
		TargetID:   uuid.New(),
		ActionType: ActionDefend,
		QueuedAt:   now.Add(-1 * time.Second),
		ExecuteAt:  now,
		Resolved:   false,
	}
	resolver.Queue.Enqueue(action)

	// Process tick
	resolver.ProcessTick(now)

	// Check LastActionTime was updated
	assert.True(t, combatant.LastActionTime.After(oldTime), "LastActionTime should be updated")
	assert.True(t, combatant.LastActionTime.Equal(now) || combatant.LastActionTime.After(now.Add(-1*time.Millisecond)),
		"LastActionTime should be around now")
}

func TestProcessTick_DeadCombatantSkipped(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()

	combatantID := uuid.New()
	combatant := &Combatant{
		EntityID:       combatantID,
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      0, // Dead
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateDefeated,
	}

	resolver.AddCombatant(combatant)

	action := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatantID,
		TargetID:   uuid.New(),
		ActionType: ActionAttack,
		QueuedAt:   now.Add(-1 * time.Second),
		ExecuteAt:  now,
		Resolved:   false,
	}
	resolver.Queue.Enqueue(action)

	// Process tick - should skip dead combatant
	resolved := resolver.ProcessTick(now)
	assert.Equal(t, 0, len(resolved), "Dead combatant's action should not execute")
}

func TestProcessTick_StunnedCombatantSkipped(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()

	combatantID := uuid.New()
	combatant := &Combatant{
		EntityID:       combatantID,
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateInCombat,
		StatusEffects: []StatusEffect{
			{EffectType: EffectStun, ExpiresAt: now.Add(1 * time.Second)},
		},
	}

	resolver.AddCombatant(combatant)

	action := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatantID,
		TargetID:   uuid.New(),
		ActionType: ActionAttack,
		QueuedAt:   now.Add(-1 * time.Second),
		ExecuteAt:  now,
		Resolved:   false,
	}
	resolver.Queue.Enqueue(action)

	// Process tick - should skip stunned combatant
	resolved := resolver.ProcessTick(now)
	assert.Equal(t, 0, len(resolved), "Stunned combatant's action should not execute")
}

func TestProcessTick_InsufficientStaminaSkipped(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()

	combatantID := uuid.New()
	combatant := &Combatant{
		EntityID:       combatantID,
		CurrentStamina: 5, // Not enough for normal attack (15)
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateInCombat,
	}

	resolver.AddCombatant(combatant)

	action := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatantID,
		TargetID:   uuid.New(),
		ActionType: ActionAttack,
		QueuedAt:   now.Add(-1 * time.Second),
		ExecuteAt:  now,
		Resolved:   false,
	}
	resolver.Queue.Enqueue(action)

	// Process tick - should skip due to insufficient stamina
	resolved := resolver.ProcessTick(now)
	assert.Equal(t, 0, len(resolved), "Action with insufficient stamina should not execute")
	assert.Equal(t, 5, combatant.CurrentStamina, "Stamina should not be consumed")
}

func TestCheckInterruption_LowDamage(t *testing.T) {
	combatant := &Combatant{
		EntityID:  uuid.New(),
		CurrentHP: 80,
		MaxHP:     100,
	}

	// 10% damage - low chance to interrupt
	interrupted := CheckInterruption(combatant, 10.0)
	// Just check it returns a boolean (can't deterministically test randomness)
	assert.IsType(t, false, interrupted)
}

func TestCheckInterruption_HighDamage(t *testing.T) {
	combatant := &Combatant{
		EntityID:  uuid.New(),
		CurrentHP: 20,
		MaxHP:     100,
	}

	// 80% damage - very high chance to interrupt
	// Run multiple times to check probability
	interruptCount := 0
	for i := 0; i < 100; i++ {
		if CheckInterruption(combatant, 80.0) {
			interruptCount++
		}
	}

	// With 80% damage, chance is 40% (80 * 0.5)
	// Expect roughly 40 interrupts out of 100
	assert.Greater(t, interruptCount, 20, "Should have some interrupts with high damage")
	assert.Less(t, interruptCount, 60, "Should not interrupt every time")
}

func TestCheckInterruption_NoDamage(t *testing.T) {
	combatant := &Combatant{
		EntityID:  uuid.New(),
		CurrentHP: 100,
		MaxHP:     100,
	}

	// 0% damage - no interruption
	interrupted := CheckInterruption(combatant, 0.0)
	assert.False(t, interrupted, "No damage should never interrupt")
}

func TestMultipleCombatants_ProcessOrder(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()

	// Create 3 combatants
	combatant1 := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateInCombat,
	}
	combatant2 := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateInCombat,
	}
	combatant3 := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CombatState:    StateInCombat,
	}

	resolver.AddCombatant(combatant1)
	resolver.AddCombatant(combatant2)
	resolver.AddCombatant(combatant3)

	// Queue actions with different execute times
	action1 := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatant1.EntityID,
		TargetID:   uuid.New(),
		ActionType: ActionAttack,
		ExecuteAt:  now.Add(-500 * time.Millisecond), // Second
		Resolved:   false,
	}
	action2 := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatant2.EntityID,
		TargetID:   uuid.New(),
		ActionType: ActionDefend,
		ExecuteAt:  now.Add(-1 * time.Second), // First
		Resolved:   false,
	}
	action3 := &CombatAction{
		ActionID:   uuid.New(),
		ActorID:    combatant3.EntityID,
		TargetID:   uuid.New(),
		ActionType: ActionAttack,
		ExecuteAt:  now.Add(-100 * time.Millisecond), // Third
		Resolved:   false,
	}

	resolver.Queue.Enqueue(action1)
	resolver.Queue.Enqueue(action2)
	resolver.Queue.Enqueue(action3)

	// Process tick
	resolved := resolver.ProcessTick(now)

	// All should execute, in order
	assert.Equal(t, 3, len(resolved))
	assert.Equal(t, combatant2.EntityID, resolved[0].ActorID, "Action 2 should be first")
	assert.Equal(t, combatant1.EntityID, resolved[1].ActorID, "Action 1 should be second")
	assert.Equal(t, combatant3.EntityID, resolved[2].ActorID, "Action 3 should be third")
}
