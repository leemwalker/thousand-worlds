package player

import (
	"math"
)

// CalculateRegenRate returns the stamina regenerated per second based on Endurance
func CalculateRegenRate(endurance int) float64 {
	// Formula: Endurance / 10 per second
	return float64(endurance) / 10.0
}

// Regenerate adds stamina based on elapsed time and rate
// Returns the amount actually added
func (sm *StaminaManager) Regenerate(elapsedSeconds float64, ratePerSecond float64) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentStamina >= sm.maxStamina {
		return 0
	}

	amountToAdd := int(math.Floor(elapsedSeconds * ratePerSecond))
	if amountToAdd <= 0 {
		return 0
	}

	oldStamina := sm.currentStamina
	sm.currentStamina += amountToAdd
	if sm.currentStamina > sm.maxStamina {
		sm.currentStamina = sm.maxStamina
	}

	return sm.currentStamina - oldStamina
}
