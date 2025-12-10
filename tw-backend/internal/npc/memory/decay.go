package memory

import (
	"time"
)

const (
	BaseDecayRate = 0.001 // 0.1% per day
	MinClarity    = 0.1
)

// CalculateCurrentClarity computes the decayed clarity of a memory
func CalculateCurrentClarity(memory Memory, now time.Time) float64 {
	daysSince := now.Sub(memory.Timestamp).Hours() / 24.0
	if daysSince < 0 {
		daysSince = 0
	}

	// Phase 3.4 Formula: effectiveDecayRate = baseDecayRate * (1 - emotionalWeight * 0.5)
	// Peak emotion (1.0) -> 50% decay resistance
	// Neutral (0.0) -> 0% resistance (Base Rate)
	effectiveRate := BaseDecayRate * (1.0 - (memory.EmotionalWeight * 0.5))

	// Rehearsal Bonus
	rehearsalBonus := float64(memory.AccessCount) / 20.0
	if rehearsalBonus > 0.5 {
		rehearsalBonus = 0.5
	}

	// Final decay amount
	totalDecay := effectiveRate * daysSince * (1.0 - rehearsalBonus)

	currentClarity := memory.Clarity * (1.0 - totalDecay)

	if currentClarity < MinClarity {
		currentClarity = MinClarity
	}

	return currentClarity
}
