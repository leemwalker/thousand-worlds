package player

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateMovementCost(t *testing.T) {
	tests := []struct {
		name     string
		distance float64
		mode     MovementType
		expected int
	}{
		{"Walk 10m", 10.0, MoveWalk, 10},     // 10 * 1.0
		{"Run 10m", 10.0, MoveRun, 20},       // 10 * 2.0
		{"Sneak 10m", 10.0, MoveSneak, 15},   // 10 * 1.5
		{"Sprint 10m", 10.0, MoveSprint, 40}, // 10 * 4.0
		{"Walk 1.5m (Round Up)", 1.5, MoveWalk, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := CalculateMovementCost(tt.distance, tt.mode)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

func TestStaminaManager_Consume(t *testing.T) {
	sm := NewStaminaManager(100)

	// Valid consumption
	err := sm.Consume(50)
	assert.NoError(t, err)
	assert.Equal(t, 50, sm.Current())

	// Invalid consumption (insufficient)
	err = sm.Consume(60)
	assert.Error(t, err)
	assert.Equal(t, 50, sm.Current()) // Should not change
}

func TestCalculateRegenRate(t *testing.T) {
	// Endurance 50 -> 5.0 per second
	rate := CalculateRegenRate(50)
	assert.Equal(t, 5.0, rate)

	// Endurance 100 -> 10.0 per second
	rate = CalculateRegenRate(100)
	assert.Equal(t, 10.0, rate)
}

func TestStaminaManager_Regenerate(t *testing.T) {
	sm := NewStaminaManager(100)
	sm.Consume(50) // Current = 50

	// Regen 1 second at rate 5.0
	added := sm.Regenerate(1.0, 5.0)
	assert.Equal(t, 5, added)
	assert.Equal(t, 55, sm.Current())

	// Regen to max
	added = sm.Regenerate(10.0, 5.0) // 50 potential, but capped at 45 needed
	assert.Equal(t, 45, added)
	assert.Equal(t, 100, sm.Current())
}

func TestValidateMovement(t *testing.T) {
	err := ValidateMovement(10, 5)
	assert.NoError(t, err)

	err = ValidateMovement(5, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stamina")
}

func TestMove(t *testing.T) {
	sm := NewStaminaManager(100)
	charID := uuid.New()

	// Move 10 meters walking (cost 10)
	movedEvent, staminaEvent, err := Move(sm, charID, 0, 0, 0, 10, 0, 0, MoveWalk)
	assert.NoError(t, err)
	assert.NotNil(t, movedEvent)
	assert.NotNil(t, staminaEvent)

	assert.Equal(t, 90, sm.Current())
	assert.Equal(t, 10, movedEvent.StaminaCost)
	assert.Equal(t, 100, staminaEvent.OldValue)
	assert.Equal(t, 90, staminaEvent.NewValue)

	// Try to sprint 100 meters (cost 400) - should fail
	_, _, err = Move(sm, charID, 0, 0, 0, 100, 0, 0, MoveSprint)
	assert.Error(t, err)
	assert.Equal(t, 90, sm.Current()) // Should not change
}
