package action

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransition = errors.New("invalid state transition")
)

// TransitionState attempts to transition a combatant to a new state
func TransitionState(combatant *Combatant, newState CombatState) error {
	if !IsValidTransition(combatant.CombatState, newState) {
		return fmt.Errorf("%w: cannot transition from %s to %s",
			ErrInvalidTransition, combatant.CombatState, newState)
	}

	combatant.CombatState = newState
	return nil
}

// IsValidTransition checks if a state transition is allowed
func IsValidTransition(from, to CombatState) bool {
	// Same state transitions are always allowed
	if from == to {
		return true
	}

	// Define valid transitions
	switch from {
	case StateIdle:
		// Idle can only go to InCombat
		return to == StateInCombat

	case StateInCombat:
		// InCombat can go to Fleeing, Defeated, or Idle
		return to == StateFleeing || to == StateDefeated || to == StateIdle

	case StateFleeing:
		// Fleeing can only go back to InCombat (flee failed)
		return to == StateInCombat

	case StateDefeated:
		// Defeated is terminal - no transitions allowed except to itself
		return false

	default:
		return false
	}
}
