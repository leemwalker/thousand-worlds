package damage

// CriticalResult holds the outcome of a critical check
type CriticalResult struct {
	IsCritical        bool
	IsCriticalFailure bool
	Multiplier        float64
	IgnoreArmor       float64 // Percentage of armor reduction to ignore (0.0 to 1.0)
}

// CalculateCritical determines if a hit is critical or a fumble
// roll: 1-100
// cunning: attribute value
// isHeavyAttack: boolean
func CalculateCritical(roll int, cunning int, isHeavyAttack bool) CriticalResult {
	// Critical Failure: Natural 1-5
	if roll <= 5 {
		return CriticalResult{
			IsCriticalFailure: true,
			Multiplier:        0.0,
			IgnoreArmor:       0.0,
		}
	}

	// Critical Hit Base Threshold: 95+
	threshold := 95

	// Cunning Bonus: +(Cunning / 50)%
	// Example: Cunning 50 -> +1% chance -> threshold 94
	// Example: Cunning 100 -> +2% chance -> threshold 93
	cunningBonus := cunning / 50
	threshold -= cunningBonus

	// Heavy Attack Bonus: +5% chance -> threshold -5
	if isHeavyAttack {
		threshold -= 5
	}

	if roll >= threshold {
		return CriticalResult{
			IsCritical:  true,
			Multiplier:  2.0,
			IgnoreArmor: 0.5, // Ignore 50% of armor
		}
	}

	return CriticalResult{
		Multiplier:  1.0,
		IgnoreArmor: 0.0,
	}
}
