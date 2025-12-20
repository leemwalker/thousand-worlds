package evolution

import (
	"math/rand"
)

// AlleleFrequency represents the frequency of an allele in a population
type AlleleFrequency struct {
	Allele    string
	Frequency float64
}

// SimulateGeneticDrift simulates random allele frequency changes in small populations.
// In small populations, allele frequencies fluctuate by chance (genetic drift).
// Returns the modified allele frequencies after the specified generations.
func SimulateGeneticDrift(alleles []AlleleFrequency, populationSize int, generations int, seed int64) []AlleleFrequency {
	if len(alleles) == 0 || populationSize <= 0 || generations <= 0 {
		return alleles
	}

	rng := rand.New(rand.NewSource(seed))

	// Copy alleles to avoid modifying input
	result := make([]AlleleFrequency, len(alleles))
	copy(result, alleles)

	// Larger populations have less drift
	driftMagnitude := 1.0 / float64(populationSize)
	if driftMagnitude > 0.5 {
		driftMagnitude = 0.5
	}

	// Simulate drift over generations
	for gen := 0; gen < generations; gen++ {
		// Apply random frequency changes
		for i := range result {
			// Random walk: add or subtract based on population size
			change := (rng.Float64()*2 - 1) * driftMagnitude
			result[i].Frequency += change

			// Clamp to valid range
			if result[i].Frequency < 0 {
				result[i].Frequency = 0
			}
			if result[i].Frequency > 1 {
				result[i].Frequency = 1
			}
		}

		// Normalize frequencies to sum to 1
		totalFreq := 0.0
		for _, af := range result {
			totalFreq += af.Frequency
		}
		if totalFreq > 0 {
			for i := range result {
				result[i].Frequency /= totalFreq
			}
		}
	}

	return result
}

// SimulateAdaptiveRadiation simulates rapid speciation when species colonize new niches.
// Like Darwin's finches, a single ancestor diversifies into multiple species.
// Returns the number of new species created.
func SimulateAdaptiveRadiation(ancestorSpecies *Species, availableNiches []string, pressureLevel float64, seed int64) int {
	if ancestorSpecies == nil || len(availableNiches) == 0 {
		return 0
	}

	rng := rand.New(rand.NewSource(seed))

	// Number of new species depends on niches and pressure
	baseSpeciation := len(availableNiches)
	speciationModifier := pressureLevel * 2.0 // Higher pressure increases radiation

	// Random variation
	variation := rng.Float64() * float64(baseSpeciation) * 0.5

	newSpeciesCount := int(float64(baseSpeciation)*speciationModifier + variation)
	if newSpeciesCount < 1 {
		newSpeciesCount = 1
	}
	if newSpeciesCount > len(availableNiches)*2 {
		newSpeciesCount = len(availableNiches) * 2
	}

	return newSpeciesCount
}

// ConvergentTrait represents a trait that evolved independently in multiple lineages
type ConvergentTrait struct {
	TraitName   string
	SpeciesIDs  []string
	Environment string
}

// DetectConvergentEvolution identifies similar traits in unrelated species.
// Returns traits that evolved independently due to similar environmental pressures.
func DetectConvergentEvolution(species []*Species, environmentType string) []ConvergentTrait {
	if len(species) < 2 {
		return nil
	}

	traits := make([]ConvergentTrait, 0)

	// Define trait thresholds for detection
	speedySpecies := make([]*Species, 0)
	armoredSpecies := make([]*Species, 0)
	camoSpecies := make([]*Species, 0)

	for _, s := range species {
		if s.Speed > 10.0 {
			speedySpecies = append(speedySpecies, s)
		}
		if s.Armor > 50.0 {
			armoredSpecies = append(armoredSpecies, s)
		}
		if s.Camouflage > 50.0 {
			camoSpecies = append(camoSpecies, s)
		}
	}

	// If multiple unrelated species have the same trait, it's convergent
	if len(speedySpecies) >= 2 {
		ids := make([]string, len(speedySpecies))
		for i, s := range speedySpecies {
			ids[i] = s.SpeciesID.String()
		}
		traits = append(traits, ConvergentTrait{
			TraitName:   "High Speed",
			SpeciesIDs:  ids,
			Environment: environmentType,
		})
	}

	if len(armoredSpecies) >= 2 {
		ids := make([]string, len(armoredSpecies))
		for i, s := range armoredSpecies {
			ids[i] = s.SpeciesID.String()
		}
		traits = append(traits, ConvergentTrait{
			TraitName:   "Heavy Armor",
			SpeciesIDs:  ids,
			Environment: environmentType,
		})
	}

	if len(camoSpecies) >= 2 {
		ids := make([]string, len(camoSpecies))
		for i, s := range camoSpecies {
			ids[i] = s.SpeciesID.String()
		}
		traits = append(traits, ConvergentTrait{
			TraitName:   "Camouflage",
			SpeciesIDs:  ids,
			Environment: environmentType,
		})
	}

	return traits
}
