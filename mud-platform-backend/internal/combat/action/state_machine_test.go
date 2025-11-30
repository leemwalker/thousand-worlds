package action

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransitionState_IdleToInCombat(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateIdle,
	}

	err := TransitionState(combatant, StateInCombat)
	assert.NoError(t, err)
	assert.Equal(t, StateInCombat, combatant.CombatState)
}

func TestTransitionState_InCombatToFleeing(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateInCombat,
	}

	err := TransitionState(combatant, StateFleeing)
	assert.NoError(t, err)
	assert.Equal(t, StateFleeing, combatant.CombatState)
}

func TestTransitionState_InCombatToDefeated(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateInCombat,
		CurrentHP:   0,
	}

	err := TransitionState(combatant, StateDefeated)
	assert.NoError(t, err)
	assert.Equal(t, StateDefeated, combatant.CombatState)
}

func TestTransitionState_FleeingToInCombat(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateFleeing,
	}

	err := TransitionState(combatant, StateInCombat)
	assert.NoError(t, err)
	assert.Equal(t, StateInCombat, combatant.CombatState)
}

func TestTransitionState_InCombatToIdle(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateInCombat,
	}

	err := TransitionState(combatant, StateIdle)
	assert.NoError(t, err)
	assert.Equal(t, StateIdle, combatant.CombatState)
}

func TestTransitionState_InvalidTransition_IdleToFleeing(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateIdle,
	}

	err := TransitionState(combatant, StateFleeing)
	assert.Error(t, err)
	assert.Equal(t, StateIdle, combatant.CombatState, "State should not change on invalid transition")
}

func TestTransitionState_InvalidTransition_IdleToDefeated(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateIdle,
	}

	err := TransitionState(combatant, StateDefeated)
	assert.Error(t, err)
	assert.Equal(t, StateIdle, combatant.CombatState, "State should not change on invalid transition")
}

func TestTransitionState_InvalidTransition_DefeatedToInCombat(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateDefeated,
	}

	err := TransitionState(combatant, StateInCombat)
	assert.Error(t, err)
	assert.Equal(t, StateDefeated, combatant.CombatState, "Cannot transition from defeated state")
}

func TestTransitionState_InvalidTransition_FleeingToIdle(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateFleeing,
	}

	err := TransitionState(combatant, StateIdle)
	assert.Error(t, err)
	assert.Equal(t, StateFleeing, combatant.CombatState, "Cannot go directly from fleeing to idle")
}

func TestTransitionState_SameState(t *testing.T) {
	combatant := &Combatant{
		EntityID:    uuid.New(),
		CombatState: StateInCombat,
	}

	err := TransitionState(combatant, StateInCombat)
	assert.NoError(t, err, "Transitioning to same state should be allowed")
	assert.Equal(t, StateInCombat, combatant.CombatState)
}

func TestIsValidTransition_AllValidCases(t *testing.T) {
	validTransitions := []struct {
		from CombatState
		to   CombatState
	}{
		{StateIdle, StateInCombat},
		{StateInCombat, StateFleeing},
		{StateInCombat, StateDefeated},
		{StateInCombat, StateIdle},
		{StateFleeing, StateInCombat},
		// Same state transitions
		{StateIdle, StateIdle},
		{StateInCombat, StateInCombat},
		{StateFleeing, StateFleeing},
		{StateDefeated, StateDefeated},
	}

	for _, tt := range validTransitions {
		assert.True(t, IsValidTransition(tt.from, tt.to),
			"Expected %s -> %s to be valid", tt.from, tt.to)
	}
}

func TestIsValidTransition_AllInvalidCases(t *testing.T) {
	invalidTransitions := []struct {
		from CombatState
		to   CombatState
	}{
		{StateIdle, StateFleeing},
		{StateIdle, StateDefeated},
		{StateFleeing, StateIdle},
		{StateFleeing, StateDefeated},
		{StateDefeated, StateIdle},
		{StateDefeated, StateInCombat},
		{StateDefeated, StateFleeing},
	}

	for _, tt := range invalidTransitions {
		assert.False(t, IsValidTransition(tt.from, tt.to),
			"Expected %s -> %s to be invalid", tt.from, tt.to)
	}
}
