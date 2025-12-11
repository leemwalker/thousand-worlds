package ecosystem

import (
	"math/rand"
	"tw-backend/internal/ecosystem/state"
	"tw-backend/internal/npc/genetics"

	"github.com/google/uuid"
)

// EvolutionManager handles reproduction and mutation
type EvolutionManager struct {
	MutationRate float64
}

func NewEvolutionManager() *EvolutionManager {
	return &EvolutionManager{
		MutationRate: 0.05, // 5% chance per gene
	}
}

// Reproduce creates a new entity from two parents
func (em *EvolutionManager) Reproduce(parent1, parent2 *state.LivingEntityState) (*state.LivingEntityState, error) {
	// Crossover
	childDNA, err := genetics.Inherit(parent1.DNA, parent2.DNA)
	if err != nil {
		return nil, err
	}

	// Mutation
	em.mutate(&childDNA)

	// Create Offspring
	species := parent1.Species // Assume same species mating for now
	generation := parent1.Generation
	if parent2.Generation > generation {
		generation = parent2.Generation
	}
	generation++

	child := &state.LivingEntityState{
		EntityID:   uuid.New(),
		Species:    species,
		Diet:       parent1.Diet, // Inherit diet
		Age:        0,
		Generation: generation,
		Needs: state.NeedState{
			Hunger:           0,
			Thirst:           0,
			Energy:           100,
			ReproductionUrge: 0,
			Safety:           100,
		},
		DNA:       childDNA,
		Parent1ID: &parent1.EntityID,
		Parent2ID: &parent2.EntityID,
	}

	return child, nil
}

func (em *EvolutionManager) mutate(dna *genetics.DNA) {
	for key, gene := range dna.Genes {
		if rand.Float64() < em.MutationRate {
			// Mutation Event
			// 1. Point Mutation: Randomly selecting a new allele
			// For simplicity, we just randomise the allele to A/a if it was something else,
			// or flip if binary.

			// Let's assume binary traits "A" (dominant) and "a" (recessive) for now
			// A -> a, a -> A
			mutateAllele := func(a string) string {
				if a == "A" {
					return "a"
				}
				if a == "a" {
					return "A"
				}
				return a // No change for unknown
			}

			if rand.Float64() < 0.5 {
				gene.Allele1 = mutateAllele(gene.Allele1)
			} else {
				gene.Allele2 = mutateAllele(gene.Allele2)
			}

			// Update the gene in the map
			dna.Genes[key] = gene
		}
	}
}
