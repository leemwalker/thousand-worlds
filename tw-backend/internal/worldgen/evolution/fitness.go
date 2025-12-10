package evolution

import (
	"math"
)

// CalculateClimateFitness calculates fitness based on environmental tolerance
func CalculateClimateFitness(species *Species, env *Environment) float64 {
	// Temperature fitness
	tempDiff := math.Abs(env.Temperature - species.TemperatureTolerance.Optimal)
	tempRange := species.TemperatureTolerance.Max - species.TemperatureTolerance.Min

	tempFitness := 1.0
	if tempRange > 0 {
		tempFitness = 1.0 - (tempDiff / tempRange)
		if tempFitness < 0 {
			tempFitness = 0
		}
	}

	// Moisture fitness
	moistDiff := math.Abs(env.Moisture - species.MoistureTolerance.Optimal)
	moistRange := species.MoistureTolerance.Max - species.MoistureTolerance.Min

	moistFitness := 1.0
	if moistRange > 0 {
		moistFitness = 1.0 - (moistDiff / moistRange)
		if moistFitness < 0 {
			moistFitness = 0
		}
	}

	// Combined climate fitness (geometric mean for multiplicative effect)
	climateFitness := math.Sqrt(tempFitness * moistFitness)

	return climateFitness
}

// CalculateFoodFitness calculates fitness based on food availability
func CalculateFoodFitness(foodAvailability float64) float64 {
	// Direct relationship: more food = higher fitness
	return foodAvailability
}

// CalculatePredationFitness calculates fitness based on predation pressure
// High predation reduces fitness unless species has defensive traits
func CalculatePredationFitness(
	species *Species,
	predationRate float64,
) float64 {
	if predationRate == 0 {
		return 1.0
	}

	// Defensive traits help survive predation
	defenseFactor := 0.0
	defenseFactor += species.Speed / 100.0 // Normalize speed
	defenseFactor += species.Camouflage / 100.0
	defenseFactor += species.Armor / 100.0
	defenseFactor /= 3.0 // Average of defensive traits

	// Fitness loss from predation, mitigated by defenses
	fitnessLoss := predationRate * (1.0 - defenseFactor)

	fitness := 1.0 - fitnessLoss
	if fitness < 0 {
		fitness = 0
	}

	return fitness
}

// CalculateCompetitionFitness wraps interspecific competition calculation
func CalculateCompetitionFitness(
	species *Species,
	competitors []*Species,
) float64 {
	return CalculateInterspecificCompetition(species, competitors)
}

// CalculateTotalFitness combines all fitness components
// Uses multiplicative combination to ensure all factors matter
func CalculateTotalFitness(
	species *Species,
	env *Environment,
	foodAvailability float64,
	predationRate float64,
	competitors []*Species,
) float64 {
	climateFitness := CalculateClimateFitness(species, env)
	foodFitness := CalculateFoodFitness(foodAvailability)
	predationFitness := CalculatePredationFitness(species, predationRate)
	competitionFitness := CalculateCompetitionFitness(species, competitors)

	// Multiplicative combination
	totalFitness := climateFitness * foodFitness * predationFitness * competitionFitness

	// Apply extinction risk
	survivalProbability := totalFitness * (1.0 - species.ExtinctionRisk)

	return survivalProbability
}

// DetermineFavoredTraits analyzes selection pressures and returns traits to favor
func DetermineFavoredTraits(
	predationRate float64,
	foodAvailability float64,
	competitionPressure float64,
) map[string]bool {
	favored := make(map[string]bool)

	// High predation favors speed, camouflage, armor
	if predationRate > 0.3 {
		favored["speed"] = true
		favored["camouflage"] = true
		favored["armor"] = true
	}

	// Food scarcity favors efficiency
	if foodAvailability < 0.5 {
		favored["efficiency"] = true // Lower calories per day
		favored["diet_flexibility"] = true
	}

	// High competition favors specialization
	if competitionPressure > 0.7 {
		favored["specialization"] = true
	}

	return favored
}
