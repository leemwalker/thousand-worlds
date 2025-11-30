package evolution

import (
	"math"
)

// CalculatePredationSuccess calculates success rate for predator catching prey
// Formula: successRate = (predatorSpeed / preySpeed) × (1 - preyCamouflage/100) × predatorHunger
func CalculatePredationSuccess(
	predatorSpeed float64,
	preySpeed float64,
	preyCamouflage float64,
	predatorHunger float64, // 0-1, higher = more desperate
) float64 {
	// Speed advantage
	speedRatio := 1.0
	if preySpeed > 0 {
		speedRatio = predatorSpeed / preySpeed
	}

	// Camouflage effect (reduces detection)
	detectionRate := 1.0 - (preyCamouflage / 100.0)

	// Combine factors
	successRate := speedRatio * detectionRate * predatorHunger

	// Cap at reasonable max (90%)
	if successRate > 0.9 {
		successRate = 0.9
	}

	// Minimum success rate (even slow predators occasionally succeed)
	if successRate < 0.05 {
		successRate = 0.05
	}

	return successRate
}

// ConsumePreyPopulation reduces prey population based on predation
// Returns number of prey killed
func ConsumePreyPopulation(
	predator *Species,
	prey *Species,
	huntsPerDay float64,
) int {
	if predator.Population == 0 || prey.Population == 0 {
		return 0
	}

	// Calculate predator hunger (based on food availability)
	hunger := 0.7 // Default moderate hunger

	// Calculate success rate
	successRate := CalculatePredationSuccess(
		predator.Speed,
		prey.Speed,
		prey.Camouflage,
		hunger,
	)

	// Calculate prey killed
	// predatorPopulation × successRate × huntsPerDay × daysPerYear
	preyKilledPerYear := float64(predator.Population) * successRate * huntsPerDay * 365

	preyKilled := int(math.Min(preyKilledPerYear, float64(prey.Population)))

	return preyKilled
}

// CalculatePredationPressure determines how much predation a species faces
// Returns 0-1 scale, higher = more predation
func CalculatePredationPressure(
	prey *Species,
	predators []*Species,
) float64 {
	if prey.Population == 0 {
		return 0
	}

	totalPredationRate := 0.0

	for _, predator := range predators {
		// Check if this predator hunts this prey
		hunts := false
		for _, preyID := range predator.PreferredPrey {
			if preyID == prey.SpeciesID {
				hunts = true
				break
			}
		}

		if !hunts {
			continue
		}

		// Estimate predation impact
		preyKilled := ConsumePreyPopulation(predator, prey, 0.5) // 0.5 hunts/day average
		predationRate := float64(preyKilled) / float64(prey.Population)
		totalPredationRate += predationRate
	}

	// Cap at 1.0
	if totalPredationRate > 1.0 {
		totalPredationRate = 1.0
	}

	return totalPredationRate
}
