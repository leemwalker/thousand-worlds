package relationship

import (
	"math"
)

// CalculateDecay reduces affection and trust if inactive
// Formula: 0.5 per 30 days (approx 0.016 per day)
// Strong relationships (>75 affection) decay 50% slower
// Negative relationships (<-50 affection) decay toward neutral faster (2x)
func CalculateDecay(rel *Relationship, daysInactive float64) {
	if daysInactive <= 0 {
		return
	}

	baseRate := 0.5 / 30.0 // ~0.0167 per day

	// Affection Decay
	affectionDecay := baseRate * daysInactive
	if rel.CurrentAffinity.Affection > 75 {
		affectionDecay *= 0.5 // Slower decay for strong bonds
	} else if rel.CurrentAffinity.Affection < -50 {
		affectionDecay *= 2.0 // Faster return to neutral for enemies?
		// Wait, "decay toward neutral faster".
		// If affection is -80, "decay" usually means magnitude decreases (goes to 0).
		// So -80 -> -70.
		// If affection is +80, "decay" means +80 -> +70.
		// So we should move TOWARDS 0.
	}

	rel.CurrentAffinity.Affection = applyDecayTowardZero(rel.CurrentAffinity.Affection, affectionDecay)

	// Trust Decay (same logic assumed unless specified otherwise)
	trustDecay := baseRate * daysInactive
	rel.CurrentAffinity.Trust = applyDecayTowardZero(rel.CurrentAffinity.Trust, trustDecay)

	// Fear does not decay according to prompt ("Only affects affection and trust")
}

func applyDecayTowardZero(current int, amount float64) int {
	if current == 0 {
		return 0
	}

	val := float64(current)
	if val > 0 {
		val -= amount
		if val < 0 {
			val = 0
		}
	} else {
		val += amount
		if val > 0 {
			val = 0
		}
	}
	return int(math.Round(val))
}
