package evolution

import (
	"math"

	"github.com/google/uuid"
)

const (
	// EnergyTransferEfficiency is the percentage of energy transferred up trophic levels
	EnergyTransferEfficiency = 0.1 // 10%

	// Base biomass production rates
	BaseBiomassProductionRate = 1000.0 // kg/km²/year
)

// CalculateBiomassProduction calculates flora biomass production
// Formula: biomassProduction = baseRate × sunlight × moisture × temperature
func CalculateBiomassProduction(
	sunlight float64,
	moisture float64, // mm/year
	temperature float64, // °C
) float64 {
	// Normalize factors to 0-1 scale
	sunlightFactor := sunlight // Already 0-1

	// Moisture: optimal around 1000-2000mm/year
	moistureFactor := 0.0
	if moisture < 1000 {
		moistureFactor = moisture / 1000.0
	} else if moisture <= 2000 {
		moistureFactor = 1.0
	} else {
		// Too much moisture reduces efficiency
		moistureFactor = 2000.0 / moisture
		if moistureFactor > 1.0 {
			moistureFactor = 1.0
		}
	}

	// Temperature: optimal 15-25°C
	tempFactor := 0.0
	if temperature >= 15 && temperature <= 25 {
		tempFactor = 1.0
	} else if temperature < 15 && temperature > 0 {
		tempFactor = temperature / 15.0
	} else if temperature > 25 && temperature < 40 {
		tempFactor = 1.0 - ((temperature - 25) / 15.0)
	}

	if tempFactor < 0 {
		tempFactor = 0
	}

	biomass := BaseBiomassProductionRate * sunlightFactor * moistureFactor * tempFactor
	return math.Max(0, biomass)
}

// CalculateEnergyTransfer calculates energy gained from consuming prey/plants
// Uses 10% energy transfer efficiency
func CalculateEnergyTransfer(biomassConsumed float64) float64 {
	return biomassConsumed * EnergyTransferEfficiency
}

// CalculateHerbivoreEnergyGain calculates energy herbivores get from flora
func CalculateHerbivoreEnergyGain(floraConsumed float64) float64 {
	return CalculateEnergyTransfer(floraConsumed)
}

// CalculateCarnivoreEnergyGain calculates energy carnivores get from prey
func CalculateCarnivoreEnergyGain(preyConsumed float64) float64 {
	return CalculateEnergyTransfer(preyConsumed)
}

// CalculateCarryingCapacity estimates max population based on available food
func CalculateCarryingCapacity(
	availableBiomass float64, // kg/km²
	caloriesPerIndividual int,
	area float64, // km²
) int {
	// Convert biomass to calories (roughly 1kg = 4000 calories)
	totalCalories := availableBiomass * area * 4000

	// Annual calories needed per individual
	annualCalories := float64(caloriesPerIndividual) * 365

	capacity := int(totalCalories / annualCalories)
	return capacity
}

// CalculateFoodAvailability determines how much food is available (0-1 scale)
func CalculateFoodAvailability(
	species *Species,
	floraPopulations map[uuid.UUID]int,
	herbivorePopulations map[uuid.UUID]int,
) float64 {
	if species.IsFlora() {
		// Flora don't need food from other species
		return 1.0
	}

	totalAvailable := 0.0
	totalNeeded := float64(species.Population * species.CaloriesPerDay * 365)

	if species.IsHerbivore() {
		// Sum up available plant biomass
		for _, plantID := range species.PreferredPlants {
			if pop, ok := floraPopulations[plantID]; ok {
				// Estimate biomass from population (simplified)
				totalAvailable += float64(pop * 100) // 100 kg per plant
			}
		}
	} else if species.IsCarnivore() {
		// Sum up available prey biomass
		for _, preyID := range species.PreferredPrey {
			if pop, ok := herbivorePopulations[preyID]; ok {
				// Estimate biomass from population (simplified)
				totalAvailable += float64(pop * 50) // 50 kg per prey
			}
		}
	}

	if totalNeeded == 0 {
		return 1.0
	}

	availability := totalAvailable / totalNeeded
	if availability > 1.0 {
		availability = 1.0
	}

	return availability
}
