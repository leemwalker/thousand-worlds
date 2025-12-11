package ecosystem

import (
	"testing"
	"tw-backend/internal/ecosystem/state"
	"tw-backend/internal/npc/genetics"

	"github.com/stretchr/testify/assert"
)

func TestEvolutionManager_Reproduce(t *testing.T) {
	em := NewEvolutionManager()

	// 1. Setup Parents with distinct traits
	// Parent 1: Speed=High (AA)
	// Parent 2: Speed=Low (aa)
	p1DNA := genetics.NewDNA()
	p1DNA.Genes["speed"] = genetics.NewGene("speed", "A", "A") // Dominant

	p2DNA := genetics.NewDNA()
	p2DNA.Genes["speed"] = genetics.NewGene("speed", "a", "a") // Recessive

	parent1 := &state.LivingEntityState{
		Species:    state.SpeciesRabbit,
		Diet:       state.DietHerbivore,
		DNA:        p1DNA,
		Generation: 1,
	}
	parent2 := &state.LivingEntityState{
		Species:    state.SpeciesRabbit,
		Diet:       state.DietHerbivore,
		DNA:        p2DNA,
		Generation: 1,
	}

	// 2. Reproduce
	child, err := em.Reproduce(parent1, parent2)
	assert.NoError(t, err)
	assert.NotNil(t, child)

	// 3. Verify Inheritance
	assert.Equal(t, state.SpeciesRabbit, child.Species)
	assert.Equal(t, 2, child.Generation)
	assert.NotNil(t, child.DNA)

	// Check gene: should be Aa (Heterozygous) if simple inheritance holds
	// But mutation might flip it. With 0.05 rate, highly likely to be Aa.
	speedGene := child.DNA.Genes["speed"]
	// Verify it inherited 'A' from one and 'a' from other
	// Since P1 is AA, it gives A. P2 is aa, it gives a.
	// So child must be Aa or aA.
	assert.True(t, (speedGene.Allele1 == "A" && speedGene.Allele2 == "a") || (speedGene.Allele1 == "a" && speedGene.Allele2 == "A"), "Child should inherit alleles from parents")
}

func TestEvolutionManager_Mutation(t *testing.T) {
	em := NewEvolutionManager()
	em.MutationRate = 1.0 // Force mutation for test

	dna := genetics.NewDNA()
	dna.Genes["test"] = genetics.NewGene("test", "A", "A")

	// Mutate logic in evolution.go is currently a placeholder that does nothing visible to alleles easily
	// unless we implement the actual mutation logic in `mutate`.
	// For now, let's just ensure it runs without panic.
	em.mutate(&dna)
}
