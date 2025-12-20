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
	cfg := GetConfig()

	// Critical Failure: Natural 1-5 (configurable)
	if roll <= cfg.GetCriticalFailureThreshold() {
		return CriticalResult{
			IsCriticalFailure: true,
			Multiplier:        0.0,
			IgnoreArmor:       0.0,
		}
	}

	// Critical Hit Base Threshold (configurable)
	threshold := cfg.GetCriticalHitBaseThreshold()

	// Cunning Bonus: +(Cunning / divisor)%
	// Example: Cunning 50, Divisor 50 -> +1% chance -> threshold -1
	cunningBonus := cunning / cfg.GetCunningBonusDivisor()
	threshold -= cunningBonus

	// Heavy Attack Bonus (configurable)
	if isHeavyAttack {
		threshold -= cfg.GetHeavyAttackBonus()
	}

	if roll >= threshold {
		return CriticalResult{
			IsCritical:  true,
			Multiplier:  cfg.GetCriticalMultiplier(),
			IgnoreArmor: cfg.GetCriticalIgnoreArmor(),
		}
	}

	return CriticalResult{
		Multiplier:  1.0,
		IgnoreArmor: 0.0,
	}
}
