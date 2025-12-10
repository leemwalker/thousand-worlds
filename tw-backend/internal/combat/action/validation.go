package action

import (
	"errors"
	"time"
)

// Validation errors
var (
	ErrActionInProgress    = errors.New("combatant has an unresolved action in progress")
	ErrActionTooSoon       = errors.New("action queued too soon after last action")
	ErrStunned             = errors.New("combatant is stunned and cannot act")
	ErrInsufficientStamina = errors.New("insufficient stamina for this action")
)

// Stamina costs per action type
const (
	StaminaCostQuickAttack  = 10
	StaminaCostNormalAttack = 15
	StaminaCostHeavyAttack  = 25
	StaminaCostDefend       = 5
	StaminaCostFlee         = 20
	StaminaCostUseItem      = 5
)

// CanQueueAction validates if a combatant can queue a new action
func CanQueueAction(combatant *Combatant, actionType ActionType, attackVariant AttackType, now time.Time) error {
	// Check if previous action is still unresolved
	if combatant.CurrentAction != nil && !combatant.CurrentAction.Resolved {
		return ErrActionInProgress
	}

	// Check minimum time since last action (200ms minimum)
	timeSinceLastAction := now.Sub(combatant.LastActionTime)
	if timeSinceLastAction < MinReactionTime {
		return ErrActionTooSoon
	}

	// Check if combatant is stunned
	if IsStunned(combatant, now) {
		return ErrStunned
	}

	// Check stamina
	requiredStamina := GetStaminaCost(actionType, attackVariant)
	if combatant.CurrentStamina < requiredStamina {
		return ErrInsufficientStamina
	}

	return nil
}

// GetStaminaCost returns the stamina cost for a given action type
func GetStaminaCost(actionType ActionType, attackVariant AttackType) int {
	switch actionType {
	case ActionAttack:
		switch attackVariant {
		case AttackQuick:
			return StaminaCostQuickAttack
		case AttackHeavy:
			return StaminaCostHeavyAttack
		default:
			return StaminaCostNormalAttack
		}
	case ActionDefend:
		return StaminaCostDefend
	case ActionFlee:
		return StaminaCostFlee
	case ActionUseItem:
		return StaminaCostUseItem
	default:
		return StaminaCostNormalAttack
	}
}

// IsStunned checks if a combatant is currently stunned
func IsStunned(combatant *Combatant, now time.Time) bool {
	for _, effect := range combatant.StatusEffects {
		if effect.EffectType == EffectStun && now.Before(effect.ExpiresAt) {
			return true
		}
	}
	return false
}
