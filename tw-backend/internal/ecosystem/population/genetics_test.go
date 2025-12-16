package population

import (
	"math/rand"
	"testing"
)

func TestNewGeneticCode(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	t.Run("creates valid genetic code", func(t *testing.T) {
		gc := NewGeneticCode(rng)

		// Check all defined genes are in range [0, 1]
		for i, val := range gc.DefinedGenes {
			if val < 0 || val > 1 {
				t.Errorf("DefinedGenes[%d] = %f, want in [0, 1]", i, val)
			}
		}

		// Check blank genes start at 0
		for i, val := range gc.BlankGenes {
			if val != 0 {
				t.Errorf("BlankGenes[%d] = %f, want 0", i, val)
			}
		}

		// Check active blanks is empty
		if len(gc.ActiveBlanks) != 0 {
			t.Errorf("ActiveBlanks length = %d, want 0", len(gc.ActiveBlanks))
		}
	})
}

func TestGeneticCode_Clone(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	original := NewGeneticCode(rng)
	original.ActivateBlankGene(5, 0.5)

	clone := original.Clone()

	// Modify original
	original.DefinedGenes[0] = 0.999

	// Clone should be independent
	if clone.DefinedGenes[0] == 0.999 {
		t.Error("Clone should be independent of original")
	}

	// Clone should have same active blanks
	if len(clone.ActiveBlanks) != 1 {
		t.Errorf("Clone should have 1 active blank, got %d", len(clone.ActiveBlanks))
	}
}

func TestGeneticCode_Mutate(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	original := NewGeneticCode(rng)
	original.ActivateBlankGene(3, 0.5)

	t.Run("mutation creates different genes", func(t *testing.T) {
		mutated := original.Mutate(rng, 1.0, 0.5) // 100% mutation rate

		differences := 0
		for i := 0; i < DefinedGeneCount; i++ {
			if original.DefinedGenes[i] != mutated.DefinedGenes[i] {
				differences++
			}
		}

		if differences == 0 {
			t.Error("Expected some genes to be different after mutation")
		}
	})

	t.Run("mutation respects bounds", func(t *testing.T) {
		mutated := original.Mutate(rng, 1.0, 1.0) // Max mutation

		for i, val := range mutated.DefinedGenes {
			if val < 0 || val > 1 {
				t.Errorf("Mutated gene[%d] = %f, out of bounds", i, val)
			}
		}
	})

	t.Run("zero mutation rate preserves genes", func(t *testing.T) {
		mutated := original.Mutate(rng, 0.0, 0.5)

		for i := 0; i < DefinedGeneCount; i++ {
			if original.DefinedGenes[i] != mutated.DefinedGenes[i] {
				t.Errorf("Gene[%d] changed with 0%% mutation rate", i)
			}
		}
	})
}

func TestGeneticCode_ActivateBlankGene(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	t.Run("activates blank gene", func(t *testing.T) {
		gc := NewGeneticCode(rng)

		result := gc.ActivateBlankGene(10, 0.7)
		if !result {
			t.Error("ActivateBlankGene should return true")
		}

		if !gc.IsBlankActive(10) {
			t.Error("Blank gene 10 should be active")
		}

		if gc.BlankGenes[10] != 0.7 {
			t.Errorf("BlankGenes[10] = %f, want 0.7", gc.BlankGenes[10])
		}
	})

	t.Run("rejects duplicate activation", func(t *testing.T) {
		gc := NewGeneticCode(rng)
		gc.ActivateBlankGene(5, 0.5)

		result := gc.ActivateBlankGene(5, 0.9)
		if result {
			t.Error("Should reject duplicate activation")
		}
	})

	t.Run("rejects invalid index", func(t *testing.T) {
		gc := NewGeneticCode(rng)

		if gc.ActivateBlankGene(-1, 0.5) {
			t.Error("Should reject negative index")
		}

		if gc.ActivateBlankGene(BlankGeneCount, 0.5) {
			t.Error("Should reject index >= BlankGeneCount")
		}
	})
}

func TestCalculateGeneticDistance(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	t.Run("identical codes have zero distance", func(t *testing.T) {
		gc1 := NewGeneticCode(rng)
		gc2 := gc1.Clone()

		dist := CalculateGeneticDistance(gc1, gc2)
		if dist != 0 {
			t.Errorf("Distance = %f, want 0 for identical codes", dist)
		}
	})

	t.Run("different codes have positive distance", func(t *testing.T) {
		gc1 := NewGeneticCode(rng)
		gc2 := NewGeneticCode(rng)

		dist := CalculateGeneticDistance(gc1, gc2)
		if dist <= 0 {
			t.Error("Distance should be positive for different codes")
		}
	})

	t.Run("body plan genes contribute more", func(t *testing.T) {
		// Create two codes identical except for one gene
		baseGc := NewGeneticCode(rng)

		// Modify a body plan gene (index 0)
		gc2BodyDiff := baseGc.Clone()
		gc2BodyDiff.DefinedGenes[0] = 1.0 - baseGc.DefinedGenes[0]
		bodyDist := CalculateGeneticDistance(baseGc, gc2BodyDiff)

		// Modify a minor gene (index 99)
		gc2MinorDiff := baseGc.Clone()
		gc2MinorDiff.DefinedGenes[99] = 1.0 - baseGc.DefinedGenes[99]
		minorDist := CalculateGeneticDistance(baseGc, gc2MinorDiff)

		// Body plan change should have more impact
		if bodyDist <= minorDist {
			t.Errorf("Body plan distance (%f) should be > minor distance (%f)", bodyDist, minorDist)
		}
	})
}

func TestCrossover(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	parent1 := NewGeneticCode(rng)
	parent2 := NewGeneticCode(rng)

	offspring := Crossover(parent1, parent2, rng)

	t.Run("offspring has valid genes", func(t *testing.T) {
		for i, val := range offspring.DefinedGenes {
			if val < 0 || val > 1 {
				t.Errorf("Offspring gene[%d] = %f, out of bounds", i, val)
			}
		}
	})

	t.Run("offspring inherits from both parents", func(t *testing.T) {
		// At least some genes should match each parent
		matchesP1 := 0
		matchesP2 := 0

		for i := 0; i < DefinedGeneCount; i++ {
			if offspring.DefinedGenes[i] == parent1.DefinedGenes[i] {
				matchesP1++
			}
			if offspring.DefinedGenes[i] == parent2.DefinedGenes[i] {
				matchesP2++
			}
		}

		if matchesP1 == 0 {
			t.Error("Offspring should have some genes from parent1")
		}
		if matchesP2 == 0 {
			t.Error("Offspring should have some genes from parent2")
		}
	})
}

func TestDefaultExpressionMatrix(t *testing.T) {
	em := DefaultExpressionMatrix()

	t.Run("has weights for all genes and traits", func(t *testing.T) {
		// Check that there are some non-zero weights
		nonZero := 0
		for g := 0; g < DefinedGeneCount; g++ {
			for tr := 0; tr < PhenotypeCount; tr++ {
				if em.Weights[g][tr] != 0 {
					nonZero++
				}
			}
		}

		if nonZero == 0 {
			t.Error("Expression matrix should have some non-zero weights")
		}
	})
}

func TestGeneticCode_ToPhenotype(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	gc := NewGeneticCode(rng)
	em := DefaultExpressionMatrix()

	phenotype := gc.ToPhenotype(em)

	t.Run("produces correct number of traits", func(t *testing.T) {
		if len(phenotype) != PhenotypeCount {
			t.Errorf("Phenotype length = %d, want %d", len(phenotype), PhenotypeCount)
		}
	})

	t.Run("all traits in valid range", func(t *testing.T) {
		for i, val := range phenotype {
			if val < 0 || val > 1 {
				t.Errorf("Phenotype[%d] = %f, out of bounds [0, 1]", i, val)
			}
		}
	})
}

func TestGetGeneCategory(t *testing.T) {
	tests := []struct {
		gene     int
		expected GeneCategory
	}{
		{0, GeneBodyPlan},
		{5, GeneBodyPlan},
		{6, GeneMorphology},
		{20, GeneMorphology},
		{21, GeneBehavior},
		{50, GeneBehavior},
		{51, GeneMinor},
		{99, GeneMinor},
	}

	for _, tt := range tests {
		result := GetGeneCategory(tt.gene)
		if result != tt.expected {
			t.Errorf("GetGeneCategory(%d) = %d, want %d", tt.gene, result, tt.expected)
		}
	}
}

func TestIsFantasticalTrait(t *testing.T) {
	if IsFantasticalTrait(0) {
		t.Error("Index 0 should not be fantastical")
	}
	if IsFantasticalTrait(49) {
		t.Error("Index 49 should not be fantastical")
	}
	if !IsFantasticalTrait(50) {
		t.Error("Index 50 should be fantastical")
	}
	if !IsFantasticalTrait(99) {
		t.Error("Index 99 should be fantastical")
	}
}
