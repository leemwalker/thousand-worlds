package minerals

import (
	"math/rand"
)

// CalculateDiscoveryChance determines the probability of finding a deposit
func CalculateDiscoveryChance(
	deposit *MineralDeposit,
	miningSkill float64, // 0-100
	perception float64, // 0-100
	timeSpentHours float64,
) float64 {
	// 1. Base chance from visibility
	baseChance := 0.0
	if deposit.SurfaceVisible {
		baseChance = 0.8 // Very high chance if visible
	} else {
		// Subsurface
		// Chance depends on depth and size
		depthFactor := 1.0 - (deposit.Depth / 2000.0) // Harder as it gets deeper
		if depthFactor < 0 {
			depthFactor = 0.01
		}

		sizeFactor := 1.0
		switch deposit.VeinSize {
		case VeinSizeMassive:
			sizeFactor = 1.5
		case VeinSizeLarge:
			sizeFactor = 1.2
		case VeinSizeMedium:
			sizeFactor = 1.0
		case VeinSizeSmall:
			sizeFactor = 0.5
		}

		baseChance = 0.05 * depthFactor * sizeFactor
	}

	// 2. Geological Knowledge
	geoKnowledge := (miningSkill + perception) / 200.0 // 0.0 to 1.0

	// 3. Prospecting Effort
	optimalTime := 4.0 // 4 hours standard
	effort := timeSpentHours / optimalTime
	if effort > 1.5 {
		effort = 1.5 // Cap benefit
	}

	// Final calculation
	chance := baseChance + (geoKnowledge * 0.5)
	chance *= effort

	// Cap at 1.0
	if chance > 1.0 {
		chance = 1.0
	}

	return chance
}

// IsDiscovered checks if the discovery attempt is successful
func IsDiscovered(chance float64) bool {
	return rand.Float64() < chance
}
