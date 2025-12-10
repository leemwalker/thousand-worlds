package action

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CombatResolver manages combat resolution for multiple combatants
type CombatResolver struct {
	Queue      *CombatQueue
	Combatants map[uuid.UUID]*Combatant
	mu         sync.RWMutex
}

// NewCombatResolver creates a new combat resolver
func NewCombatResolver() *CombatResolver {
	return &CombatResolver{
		Queue:      NewCombatQueue(),
		Combatants: make(map[uuid.UUID]*Combatant),
	}
}

// AddCombatant adds a combatant to the resolver
func (cr *CombatResolver) AddCombatant(combatant *Combatant) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.Combatants[combatant.EntityID] = combatant
}

// GetCombatant retrieves a combatant by ID
func (cr *CombatResolver) GetCombatant(id uuid.UUID) *Combatant {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.Combatants[id]
}

// ProcessTick processes all actions ready to execute at the current time
func (cr *CombatResolver) ProcessTick(now time.Time) []*CombatAction {
	var resolvedActions []*CombatAction

	for {
		// Peek at next action
		action := cr.Queue.Peek()
		if action == nil {
			break // Queue empty
		}

		// Check if action is ready to execute
		if action.ExecuteAt.After(now) {
			break // Not ready yet
		}

		// Dequeue the action
		action = cr.Queue.Dequeue()
		if action == nil {
			break
		}

		// Get combatant
		combatant := cr.GetCombatant(action.ActorID)
		if combatant == nil {
			// Combatant not found, skip
			continue
		}

		// Validate combatant can still act
		if !canExecuteAction(combatant, action, now) {
			// Skip this action
			continue
		}

		// Consume stamina
		staminaCost := GetStaminaCost(action.ActionType, AttackNormal) // TODO: Get actual attack variant
		combatant.CurrentStamina -= staminaCost

		// Execute action (stub for Phase 7.2 - actual damage/effects)
		// For now, just mark as resolved
		action.Resolved = true

		// Update last action time
		combatant.LastActionTime = now

		// Add to resolved list
		resolvedActions = append(resolvedActions, action)
	}

	return resolvedActions
}

// canExecuteAction checks if a combatant can execute their action
func canExecuteAction(combatant *Combatant, action *CombatAction, now time.Time) bool {
	// Check if combatant is alive
	if combatant.CurrentHP <= 0 {
		return false
	}

	// Check if combatant is stunned
	if IsStunned(combatant, now) {
		return false
	}

	// Check stamina
	staminaCost := GetStaminaCost(action.ActionType, AttackNormal)
	if combatant.CurrentStamina < staminaCost {
		return false
	}

	return true
}

// CheckInterruption determines if an action should be interrupted by damage
func CheckInterruption(combatant *Combatant, damagePercent float64) bool {
	if damagePercent <= 0 {
		return false
	}

	// Interrupt chance: damagePercent * 0.5
	interruptChance := damagePercent * 0.5

	// Roll random number 0-100
	roll := rand.Float64() * 100

	return roll < interruptChance
}
