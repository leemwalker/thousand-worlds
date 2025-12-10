package action

import (
	"time"
)

// Base reaction times
const (
	BaseTimeQuickAttack  = 800 * time.Millisecond
	BaseTimeNormalAttack = 1000 * time.Millisecond
	BaseTimeHeavyAttack  = 1500 * time.Millisecond
	BaseTimeDefend       = 500 * time.Millisecond
	BaseTimeFlee         = 2000 * time.Millisecond
	BaseTimeUseItem      = 700 * time.Millisecond
	MinReactionTime      = 200 * time.Millisecond
)

// AttackType variants for more granular calculation
type AttackType string

const (
	AttackQuick  AttackType = "quick"
	AttackNormal AttackType = "normal"
	AttackHeavy  AttackType = "heavy"
)

// CalculateReactionTime determines the delay before an action executes
func CalculateReactionTime(actionType ActionType, attackVariant AttackType, agility int, modifiers float64) time.Duration {
	var base time.Duration

	switch actionType {
	case ActionAttack:
		switch attackVariant {
		case AttackQuick:
			base = BaseTimeQuickAttack
		case AttackHeavy:
			base = BaseTimeHeavyAttack
		default:
			base = BaseTimeNormalAttack
		}
	case ActionDefend:
		base = BaseTimeDefend
	case ActionFlee:
		base = BaseTimeFlee
	case ActionUseItem:
		base = BaseTimeUseItem
	default:
		base = BaseTimeNormalAttack
	}

	// Agility modifier: final = base * (1 - (Agility / 100) * 0.3)
	// Max reduction is 30% at 100 Agility
	agilityFactor := (float64(agility) / 100.0) * 0.3
	multiplier := 1.0 - agilityFactor

	// Apply status effect modifiers (passed as float, e.g., 1.5 for Slow, 0.7 for Haste)
	// If modifiers is 0 (no effect), treat as 1.0
	if modifiers == 0 {
		modifiers = 1.0
	}

	finalDuration := time.Duration(float64(base) * multiplier * modifiers)

	if finalDuration < MinReactionTime {
		return MinReactionTime
	}

	return finalDuration
}
