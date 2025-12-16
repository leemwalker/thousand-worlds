package population

import (
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// MigrateSpecies moves a percentage of a species population to another biome
// Returns the number of individuals that successfully migrated
func MigrateSpecies(source, dest *BiomePopulation, speciesID uuid.UUID, percentage float64) int64 {
	species, exists := source.Species[speciesID]
	if !exists || species.Count == 0 {
		return 0
	}

	// Check biome compatibility (can't migrate land species to ocean or vice versa)
	if !AreBiomesCompatible(source.BiomeType, dest.BiomeType) {
		return 0
	}

	// Calculate migrants
	migrants := int64(float64(species.Count) * percentage)
	if migrants <= 0 {
		return 0
	}

	// Reduce source population
	species.Count -= migrants

	// Check if species already exists in destination
	var destSpecies *SpeciesPopulation
	for _, sp := range dest.Species {
		if sp.Name == species.Name && sp.Diet == species.Diet {
			destSpecies = sp
			break
		}
	}

	if destSpecies != nil {
		// Add to existing population
		destSpecies.Count += migrants
	} else {
		// Create new population with slightly mutated traits (founder effect)
		newSpecies := &SpeciesPopulation{
			SpeciesID:     uuid.New(),
			AncestorID:    &species.SpeciesID,
			Name:          species.Name,
			Count:         migrants,
			Traits:        species.Traits,
			TraitVariance: species.TraitVariance * 1.2, // Increased variance from founder effect
			Diet:          species.Diet,
			Generation:    species.Generation,
			CreatedYear:   species.CreatedYear,
		}
		dest.AddSpecies(newSpecies)
	}

	return migrants
}

// AreBiomesCompatible checks if species can migrate between two biomes
func AreBiomesCompatible(source, dest geography.BiomeType) bool {
	// Ocean is incompatible with land biomes
	isSourceOcean := source == geography.BiomeOcean
	isDestOcean := dest == geography.BiomeOcean

	if isSourceOcean != isDestOcean {
		return false
	}

	// All land biomes are compatible (with varying difficulty)
	// All ocean biomes are compatible with each other
	return true
}

// TransitionBiome changes a biome's type due to climate change
// severity is 0.0-1.0, affects how much species are impacted
func TransitionBiome(biome *BiomePopulation, newType geography.BiomeType, severity float64) {
	oldType := biome.BiomeType
	biome.BiomeType = newType

	// Apply stress to species based on trait/biome mismatch
	for _, species := range biome.Species {
		oldFitness := CalculateBiomeFitness(species.Traits, oldType)
		newFitness := CalculateBiomeFitness(species.Traits, newType)

		// If fitness decreases, apply population loss
		if newFitness < oldFitness {
			fitnessLoss := (oldFitness - newFitness) * severity
			deathRate := fitnessLoss * 0.5 // Up to 50% die from biome change
			deaths := int64(float64(species.Count) * deathRate)
			species.Count -= deaths
			if species.Count < 0 {
				species.Count = 0
			}
		}
	}
}

// CalculateMigrationChance determines likelihood of migration for a species
// Based on population pressure, genetic diversity, and biome capacity
func CalculateMigrationChance(species *SpeciesPopulation, carryingCapacity int64) float64 {
	if species.Count == 0 || carryingCapacity == 0 {
		return 0.0
	}

	// Base chance from population pressure (higher pop = more migration)
	populationPressure := float64(species.Count) / float64(carryingCapacity)

	// Diversity factor (more diverse = more likely to explore)
	diversityFactor := species.TraitVariance

	// Combined chance
	chance := populationPressure * 0.2 * (1 + diversityFactor)

	// Clamp to reasonable range
	if chance < 0 {
		chance = 0
	}
	if chance > 0.5 {
		chance = 0.5
	}

	return chance
}

// GetBiomeTransitionTarget determines what biome type a biome transitions to
func GetBiomeTransitionTarget(current geography.BiomeType, event string) geography.BiomeType {
	transitions := map[geography.BiomeType]map[string]geography.BiomeType{
		geography.BiomeRainforest: {
			"ice_age": geography.BiomeDeciduousForest,
			"drought": geography.BiomeGrassland,
			"warming": geography.BiomeRainforest, // No change
		},
		geography.BiomeDeciduousForest: {
			"ice_age": geography.BiomeTaiga,
			"warming": geography.BiomeRainforest,
			"drought": geography.BiomeGrassland,
		},
		geography.BiomeGrassland: {
			"ice_age":           geography.BiomeTundra,
			"warming":           geography.BiomeDesert,
			"rainfall_increase": geography.BiomeDeciduousForest,
		},
		geography.BiomeTaiga: {
			"ice_age": geography.BiomeTundra,
			"warming": geography.BiomeGrassland,
		},
		geography.BiomeTundra: {
			"warming": geography.BiomeGrassland,
			"ice_age": geography.BiomeTundra, // No change
		},
		geography.BiomeDesert: {
			"rainfall_increase": geography.BiomeGrassland,
			"warming":           geography.BiomeDesert, // No change
			"ice_age":           geography.BiomeTundra,
		},
		geography.BiomeAlpine: {
			"warming": geography.BiomeTaiga,
			"ice_age": geography.BiomeTundra,
		},
	}

	if biomeTransitions, ok := transitions[current]; ok {
		if target, ok := biomeTransitions[event]; ok {
			return target
		}
	}

	return current // No change if not defined
}

// ApplyMigrationCycle checks all species in all biomes for potential migration
func (ps *PopulationSimulator) ApplyMigrationCycle() int64 {
	var totalMigrants int64

	// Get list of biome pairs for potential migration
	biomes := make([]*BiomePopulation, 0, len(ps.Biomes))
	for _, biome := range ps.Biomes {
		biomes = append(biomes, biome)
	}

	// Check each biome for migration opportunities
	for _, sourceBiome := range biomes {
		for speciesID, species := range sourceBiome.Species {
			migrationChance := CalculateMigrationChance(species, sourceBiome.CarryingCapacity)

			if ps.rng.Float64() < migrationChance {
				// Find a compatible destination
				for _, destBiome := range biomes {
					if destBiome.BiomeID == sourceBiome.BiomeID {
						continue
					}
					if AreBiomesCompatible(sourceBiome.BiomeType, destBiome.BiomeType) {
						migrants := MigrateSpecies(sourceBiome, destBiome, speciesID, 0.05)
						totalMigrants += migrants
						break
					}
				}
			}
		}
	}

	return totalMigrants
}

// ApplyBiomeTransitions checks for and applies biome type changes
// Returns the number of biomes that transitioned
func (ps *PopulationSimulator) ApplyBiomeTransitions(eventType ExtinctionEventType, severity float64) int {
	var event string
	minSeverity := 0.3 // Lowered to trigger more transitions

	switch eventType {
	case EventIceAge:
		event = "ice_age"
		minSeverity = 0.4
	case EventAsteroidImpact:
		if severity > 0.6 {
			event = "ice_age" // Nuclear winter from impact debris
			minSeverity = 0.6
		} else {
			return 0
		}
	case EventFloodBasalt:
		// Short-term cooling effect - biomes don't transition much
		// The warming comes later from EventGreenhouseSpike
		return 0
	case EventVolcanicWinter:
		if severity > 0.5 {
			event = "ice_age" // Volcanic winter
			minSeverity = 0.5
		} else {
			return 0
		}
	case EventContinentalDrift:
		event = "drought"
		minSeverity = 0.5

	// V2: Warming events that restore biome diversity
	case EventWarming:
		event = "warming"
		minSeverity = 0.3
	case EventGreenhouseSpike:
		event = "warming"
		minSeverity = 0.4
	case EventOceanAnoxia:
		// Ocean anoxia causes warming (CO2 from dead marine life)
		event = "warming"
		minSeverity = 0.5

	default:
		return 0
	}

	if severity < minSeverity {
		return 0
	}

	transitioned := 0
	for _, biome := range ps.Biomes {
		newType := GetBiomeTransitionTarget(biome.BiomeType, event)
		if newType != biome.BiomeType {
			TransitionBiome(biome, newType, severity)
			transitioned++
		}
	}
	return transitioned
}
