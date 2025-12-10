package player

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// MovementType defines the mode of movement
type MovementType string

const (
	MoveWalk   MovementType = "Walk"
	MoveRun    MovementType = "Run"
	MoveSneak  MovementType = "Sneak"
	MoveSprint MovementType = "Sprint"
)

const (
	CostWalk   = 1.0
	CostRun    = 2.0
	CostSneak  = 1.5
	CostSprint = 4.0
)

// CalculateMovementCost calculates the stamina cost for a given distance and mode
func CalculateMovementCost(distance float64, mode MovementType) int {
	var multiplier float64
	switch mode {
	case MoveWalk:
		multiplier = CostWalk
	case MoveRun:
		multiplier = CostRun
	case MoveSneak:
		multiplier = CostSneak
	case MoveSprint:
		multiplier = CostSprint
	default:
		multiplier = CostWalk
	}

	// Round up to nearest integer
	return int(math.Ceil(distance * multiplier))
}

// ValidateMovement checks if the character has enough stamina
func ValidateMovement(currentStamina int, cost int) error {
	if currentStamina < cost {
		return fmt.Errorf("insufficient stamina: have %d, need %d", currentStamina, cost)
	}
	return nil
}

// Move attempts to move the character, consuming stamina and returning events
func Move(sm *StaminaManager, charID uuid.UUID, fromX, fromY, fromZ, toX, toY, toZ float64, mode MovementType) (*PlayerMovedEvent, *StaminaChangedEvent, error) {
	distance := math.Sqrt(math.Pow(toX-fromX, 2) + math.Pow(toY-fromY, 2) + math.Pow(toZ-fromZ, 2))
	cost := CalculateMovementCost(distance, mode)

	oldStamina := sm.Current()
	if err := sm.Consume(cost); err != nil {
		return nil, nil, err
	}
	newStamina := sm.Current()

	movedEvent := &PlayerMovedEvent{
		CharacterID:  charID,
		FromX:        fromX,
		FromY:        fromY,
		FromZ:        fromZ,
		ToX:          toX,
		ToY:          toY,
		ToZ:          toZ,
		MovementType: mode,
		StaminaCost:  cost,
		Timestamp:    time.Now(),
	}

	staminaEvent := &StaminaChangedEvent{
		CharacterID: charID,
		OldValue:    oldStamina,
		NewValue:    newStamina,
		Reason:      fmt.Sprintf("Moved %s", mode),
		Timestamp:   time.Now(),
	}

	return movedEvent, staminaEvent, nil
}
