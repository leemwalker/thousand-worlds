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

// ExtinctionEvent represents a catastrophic event that affects species survival
type ExtinctionEvent struct {
	Type     string  // "asteroid", "volcano", "ice_age", etc.
	Severity float64 // 0.0 to 1.0
}

// CalculateO2Effects returns the effects of atmospheric oxygen level on arthropod size limits.
// Higher O2 allows larger arthropods (Carboniferous period).
// RED STATE: Returns zero values - not yet implemented.
func CalculateO2Effects(o2Level, co2Level float64) (maxArthropodSize float64, sizeMultiplier float64) {
	// TODO: Implement O2-size relationship
	// High O2 (35%+) should allow giant arthropods (size >= 3.0)
	// Current O2 (21%) should limit to ~0.5
	return 0, 0
}

// ApplyExtinctionEvent applies a mass extinction event to a set of species.
// Returns the number of species that went extinct.
// RED STATE: Returns 0 - not yet implemented.
func ApplyExtinctionEvent(species []*Species, event ExtinctionEvent) int {
	// TODO: Implement extinction logic
	// Severity 0.9 should kill 75%+ of species
	// Large animals (size > 5) suffer more
	return 0
}

// SimulateCambrianExplosion simulates rapid diversification when O2 > 10%.
// Returns the number of new species created.
// RED STATE: Returns 0 - not yet implemented.
func SimulateCambrianExplosion(species []*Species, o2Level float64, years int64) int {
	// TODO: Implement Cambrian explosion
	// Should increase species count 10x or more over 50M years
	return 0
}
