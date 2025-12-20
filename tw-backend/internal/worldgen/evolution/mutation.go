package evolution

import (
	"math/rand"
)

const (
	// BaseMutationRate is the normal mutation rate
	BaseMutationRateMin = 0.01 // 1%
	BaseMutationRateMax = 0.05 // 5%

	// BottleneckMutationRate is elevated during genetic bottlenecks
	BottleneckMutationRateMin = 0.10 // 10%
	BottleneckMutationRateMax = 0.15 // 15%

	// SpeciationThreshold determines when mutations create new species
	SpeciationThresholdMultiplier = 1.5 // 50% change
)

// GetMutationRate returns the appropriate mutation rate based on population status
func GetMutationRate(species *Species) float64 {
	if species.IsBottlenecked() {
		// Elevated mutation rate during bottleneck
		return BottleneckMutationRateMin + rand.Float64()*(BottleneckMutationRateMax-BottleneckMutationRateMin)
	}

	// Normal mutation rate
	return BaseMutationRateMin + rand.Float64()*(BaseMutationRateMax-BaseMutationRateMin)
}

// MutateSpecies creates a mutated variant of a species
// Returns the mutated species and whether it's a new species (speciation)
func MutateSpecies(parent *Species) (*Species, bool) {
	mutant := parent.Clone()
	mutant.Generation = parent.Generation + 1
	parentID := parent.SpeciesID
	mutant.ParentSpeciesID = &parentID

	// Track if mutation is significant enough for speciation
	speciationTriggered := false

	// Mutate physical traits
	originalSize := mutant.Size
	mutant.Size *= 0.9 + rand.Float64()*0.2 // ±10%
	if mutant.Size > originalSize*SpeciationThresholdMultiplier ||
		mutant.Size < originalSize/SpeciationThresholdMultiplier {
		speciationTriggered = true
	}

	if !mutant.IsFlora() {
		originalSpeed := mutant.Speed
		mutant.Speed *= 0.9 + rand.Float64()*0.2 // ±10%
		if mutant.Speed < 0 {
			mutant.Speed = 0
		}
		if mutant.Speed > originalSpeed*SpeciationThresholdMultiplier ||
			mutant.Speed < originalSpeed/SpeciationThresholdMultiplier {
			speciationTriggered = true
		}

		mutant.Armor += (rand.Float64()*10 - 5) // ±5
		if mutant.Armor < 0 {
			mutant.Armor = 0
		}
		if mutant.Armor > 100 {
			mutant.Armor = 100
		}

		mutant.Camouflage += (rand.Float64()*10 - 5) // ±5
		if mutant.Camouflage < 0 {
			mutant.Camouflage = 0
		}
		if mutant.Camouflage > 100 {
			mutant.Camouflage = 100
		}
	}

	// Mutate dietary needs
	originalCalories := mutant.CaloriesPerDay
	mutant.CaloriesPerDay = int(float64(mutant.CaloriesPerDay) * (0.9 + rand.Float64()*0.2))
	calorieChange := float64(mutant.CaloriesPerDay) / float64(max(1, originalCalories))
	if calorieChange > SpeciationThresholdMultiplier ||
		calorieChange < 1.0/SpeciationThresholdMultiplier {
		speciationTriggered = true
	}

	// Mutate temperature tolerance
	mutant.TemperatureTolerance.Optimal += (rand.Float64()*4 - 2) // ±2°C
	mutant.TemperatureTolerance.Min += (rand.Float64()*2 - 1)     // ±1°C
	mutant.TemperatureTolerance.Max += (rand.Float64()*2 - 1)     // ±1°C

	// Mutate reproduction rate
	mutant.ReproductionRate *= 0.9 + rand.Float64()*0.2 // ±10%
	if mutant.ReproductionRate < 0.1 {
		mutant.ReproductionRate = 0.1
	}

	// Mutate lifespan
	mutant.Lifespan = int(float64(mutant.Lifespan) * (0.9 + rand.Float64()*0.2))
	if mutant.Lifespan < 1 {
		mutant.Lifespan = 1
	}

	// Start with small population if creating new species
	if speciationTriggered {
		mutant.Population = int(float64(parent.Population) * 0.1) // 10% of parent pop
		mutant.Name = parent.Name + " (gen " + string(rune(mutant.Generation)) + ")"
		return mutant, true
	}

	// Not significant enough for new species, modify parent in place
	mutant.SpeciesID = parent.SpeciesID
	return mutant, false
}

// ApplyMutationsToPopulation applies mutations across a population
// Returns list of new species created through speciation
func ApplyMutationsToPopulation(species []*Species) []*Species {
	newSpecies := []*Species{}

	for _, s := range species {
		if s.Population == 0 {
			continue // Skip extinct species
		}

		mutationRate := GetMutationRate(s)

		if rand.Float64() < mutationRate {
			mutant, isNewSpecies := MutateSpecies(s)

			if isNewSpecies {
				// Speciation occurred
				newSpecies = append(newSpecies, mutant)
				// Reduce parent population
				s.Population -= mutant.Population
			} else {
				// Apply mutations to existing species
				s.Size = mutant.Size
				s.Speed = mutant.Speed
				s.Armor = mutant.Armor
				s.Camouflage = mutant.Camouflage
				s.CaloriesPerDay = mutant.CaloriesPerDay
				s.TemperatureTolerance = mutant.TemperatureTolerance
				s.ReproductionRate = mutant.ReproductionRate
				s.Lifespan = mutant.Lifespan
			}
		}
	}

	return newSpecies
}

// MutationOperator defines the type of genetic mutation
type MutationOperator string

const (
	MutationPoint       MutationOperator = "point"
	MutationInsertion   MutationOperator = "insertion"
	MutationDeletion    MutationOperator = "deletion"
	MutationDuplication MutationOperator = "duplication"
)

// ApplyMutationOperator applies a specific mutation type to a genome string.
// Returns the mutated genome.
// - Point: substitutes a single nucleotide (same length)
// - Insertion: adds a nucleotide (length + 1)
// - Deletion: removes a nucleotide (length - 1)
// - Duplication: duplicates a gene segment (length increases)
func ApplyMutationOperator(genome string, operator MutationOperator, position int, seed int64) string {
	if len(genome) == 0 {
		return genome
	}

	rng := rand.New(rand.NewSource(seed))
	nucleotides := []byte("ATCG")

	// Ensure position is within bounds
	if position < 0 || position >= len(genome) {
		position = rng.Intn(len(genome))
	}

	switch operator {
	case MutationPoint:
		// Substitute one nucleotide with a different one
		genomeBytes := []byte(genome)
		current := genomeBytes[position]
		var newNuc byte
		for {
			newNuc = nucleotides[rng.Intn(4)]
			if newNuc != current {
				break
			}
		}
		genomeBytes[position] = newNuc
		return string(genomeBytes)

	case MutationInsertion:
		// Insert a random nucleotide at position
		newNuc := nucleotides[rng.Intn(4)]
		return genome[:position] + string(newNuc) + genome[position:]

	case MutationDeletion:
		// Remove the nucleotide at position
		if len(genome) <= 1 {
			return "" // Can't delete from single nucleotide
		}
		return genome[:position] + genome[position+1:]

	case MutationDuplication:
		// Duplicate a segment (3-5 nucleotides) at position
		segLen := 3 + rng.Intn(3)
		endPos := position + segLen
		if endPos > len(genome) {
			endPos = len(genome)
		}
		segment := genome[position:endPos]
		return genome[:endPos] + segment + genome[endPos:]

	default:
		return genome
	}
}
