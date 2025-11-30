package evolution

import (
	"math"
)

// CalculateIntraspecificCompetition calculates competition within a species
// When population exceeds carrying capacity, survival rate decreases
func CalculateIntraspecificCompetition(
	population int,
	carryingCapacity int,
) float64 {
	if carryingCapacity == 0 || population == 0 {
		return 1.0
	}

	if population <= carryingCapacity {
		return 1.0 // No competition, plenty of resources
	}

	// Survival rate decreases as population exceeds capacity
	survivalRate := float64(carryingCapacity) / float64(population)

	return survivalRate
}

// CalculateNicheOverlap calculates how much two species' niches overlap
// Returns 0-1, higher = more overlap = more competition
func CalculateNicheOverlap(species1, species2 *Species) float64 {
	if species1.SpeciesID == species2.SpeciesID {
		return 0 // Same species, handled by intraspecific competition
	}

	overlap := 0.0
	factors := 0

	// Diet overlap
	if species1.Diet == species2.Diet {
		overlap += 1.0
		factors++
	} else {
		factors++
	}

	// Biome overlap
	biomeOverlap := 0.0
	for _, b1 := range species1.PreferredBiomes {
		for _, b2 := range species2.PreferredBiomes {
			if b1 == b2 {
				biomeOverlap += 1.0
			}
		}
	}
	if len(species1.PreferredBiomes) > 0 && len(species2.PreferredBiomes) > 0 {
		biomeOverlap /= float64(math.Max(float64(len(species1.PreferredBiomes)), float64(len(species2.PreferredBiomes))))
		overlap += biomeOverlap
		factors++
	}

	// Temperature tolerance overlap
	tempOverlap := CalculateRangeOverlap(
		species1.TemperatureTolerance.Min,
		species1.TemperatureTolerance.Max,
		species2.TemperatureTolerance.Min,
		species2.TemperatureTolerance.Max,
	)
	overlap += tempOverlap
	factors++

	if factors == 0 {
		return 0
	}

	return overlap / float64(factors)
}

// CalculateRangeOverlap calculates overlap between two numeric ranges
func CalculateRangeOverlap(min1, max1, min2, max2 float64) float64 {
	overlapStart := math.Max(min1, min2)
	overlapEnd := math.Min(max1, max2)

	if overlapStart >= overlapEnd {
		return 0 // No overlap
	}

	overlapSize := overlapEnd - overlapStart
	range1Size := max1 - min1
	range2Size := max2 - min2
	avgRangeSize := (range1Size + range2Size) / 2

	if avgRangeSize == 0 {
		return 0
	}

	overlapRatio := overlapSize / avgRangeSize
	if overlapRatio > 1.0 {
		overlapRatio = 1.0
	}

	return overlapRatio
}

// CalculateInterspecificCompetition calculates competition between different species
func CalculateInterspecificCompetition(
	species *Species,
	competitors []*Species,
) float64 {
	if len(competitors) == 0 {
		return 1.0 // No competition
	}

	totalCompetitionPressure := 0.0

	for _, competitor := range competitors {
		// Calculate niche overlap
		overlap := CalculateNicheOverlap(species, competitor)

		if overlap == 0 {
			continue
		}

		// Competition pressure is proportional to overlap and competitor population density
		competitionPressure := overlap * competitor.PopulationDensity / 100.0 // Normalize density
		totalCompetitionPressure += competitionPressure
	}

	// Fitness reduction from competition
	fitnessReduction := totalCompetitionPressure * 0.1

	competitionFitness := 1.0 - fitnessReduction
	if competitionFitness < 0 {
		competitionFitness = 0
	}

	return competitionFitness
}
