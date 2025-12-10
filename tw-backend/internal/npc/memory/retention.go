package memory

import (
	"time"
)

// CalculateRetentionScore determines if a memory is strong enough to be recalled
// Formula: baseRetention × (1 - decayRate × timeSinceCreation) × (1 + rehearsalBonus × accessCount)
// baseRetention = clarity × emotionalWeight
func CalculateRetentionScore(memory Memory, now time.Time) float64 {
	baseRetention := memory.Clarity * memory.EmotionalWeight

	daysSince := now.Sub(memory.Timestamp).Hours() / 24.0
	if daysSince < 0 {
		daysSince = 0
	}

	// Decay factor (simplified from prompt formula)
	// "1 - decayRate × timeSinceCreation"
	// We use BaseDecayRate here as a general factor
	decayFactor := 1.0 - (BaseDecayRate * daysSince)
	if decayFactor < 0 {
		decayFactor = 0
	}

	// Rehearsal factor
	// "1 + rehearsalBonus × accessCount"
	// Rehearsal bonus is min(access/20, 0.5).
	// Wait, the prompt formula says "rehearsalBonus * accessCount".
	// If bonus is 0.5 and count is 10, factor is 1 + 5 = 6?
	// That seems high.
	// Let's check prompt: "rehearsalBonus = min(accessCount / 20, 0.5)"
	// "retention = ... * (1 + rehearsalBonus × accessCount)"
	// If accessCount is 100 -> Bonus 0.5. Factor 1 + 0.5 * 100 = 51.
	// This makes frequently accessed memories extremely sticky.
	// This seems intended.

	rehearsalBonus := CalculateRehearsalBonus(memory.AccessCount)
	rehearsalFactor := 1.0 + (rehearsalBonus * float64(memory.AccessCount))

	return baseRetention * decayFactor * rehearsalFactor
}

// IsForgotten checks if retention score is below threshold
func IsForgotten(score float64) bool {
	return score < 0.15
}
