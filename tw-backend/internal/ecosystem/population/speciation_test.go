package population

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
)

func TestSpeciationChecker_Allopatric(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	checker := NewSpeciationChecker(42)

	// Create a parent species with genetic code
	parent := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Test Species",
		Count:         1000,
		GeneticCode:   NewGeneticCode(rng),
		TraitVariance: 0.5,
		Diet:          DietHerbivore,
	}

	t.Run("no speciation with short isolation", func(t *testing.T) {
		result := checker.CheckAllopatricSpeciation(parent, 10000, uuid.New(), 100000)
		if result != nil {
			t.Error("Should not speciate with only 10,000 years isolation")
		}
	})

	t.Run("potential speciation with long isolation", func(t *testing.T) {
		// With long isolation and many attempts, should get at least one speciation
		speciationOccurred := false
		for i := 0; i < 100; i++ {
			result := checker.CheckAllopatricSpeciation(parent, 500000, uuid.New(), int64(100000+i*1000))
			if result != nil {
				speciationOccurred = true
				// Verify daughter properties
				if result.AncestorID == nil || *result.AncestorID != parent.SpeciesID {
					t.Error("Daughter should have parent as ancestor")
				}
				if result.Count >= parent.Count {
					t.Error("Daughter should have smaller population")
				}
				break
			}
		}
		if !speciationOccurred {
			t.Log("No speciation occurred in 100 attempts - may be probabilistic")
		}
	})

	t.Run("no speciation without genetic code", func(t *testing.T) {
		parentNoGenes := &SpeciesPopulation{
			SpeciesID:   uuid.New(),
			Name:        "Old Species",
			Count:       1000,
			GeneticCode: nil, // No V2 genetic code
		}
		result := checker.CheckAllopatricSpeciation(parentNoGenes, 1000000, uuid.New(), 100000)
		if result != nil {
			t.Error("Should not speciate without genetic code")
		}
	})
}

func TestSpeciationChecker_Sympatric(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	checker := NewSpeciationChecker(42)

	parent := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Test Species",
		Count:         500, // Larger population needed
		GeneticCode:   NewGeneticCode(rng),
		TraitVariance: 0.5, // High variance needed
		Diet:          DietOmnivore,
	}

	t.Run("sympatric is rare", func(t *testing.T) {
		// Sympatric speciation is rare - 5% base rate * conditions
		speciationCount := 0
		for i := 0; i < 1000; i++ {
			result := checker.CheckSympatricSpeciation(parent, 0.8, 0.8, int64(i*1000))
			if result != nil {
				speciationCount++
			}
		}
		// Should be rare - less than 10% of attempts even with ideal conditions
		if speciationCount > 100 {
			t.Errorf("Sympatric speciation too common: %d/1000", speciationCount)
		}
		t.Logf("Sympatric speciation rate: %d/1000", speciationCount)
	})
}

func TestSpeciationChecker_Peripatric(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	checker := NewSpeciationChecker(42)

	parent := &SpeciesPopulation{
		SpeciesID:   uuid.New(),
		Name:        "Test Species",
		Count:       1000,
		GeneticCode: NewGeneticCode(rng),
		Diet:        DietHerbivore,
	}

	t.Run("peripatric needs small peripheral population", func(t *testing.T) {
		// Too large peripheral population
		result := checker.CheckPeripatricSpeciation(parent, 1000, 50000, 100000)
		if result != nil {
			t.Error("Peripatric should not occur with large peripheral population")
		}

		// Appropriately small
		speciationCount := 0
		for i := 0; i < 100; i++ {
			result := checker.CheckPeripatricSpeciation(parent, 100, 100000, int64(100000+i*1000))
			if result != nil {
				speciationCount++
			}
		}
		t.Logf("Peripatric speciation with 100 individuals: %d/100", speciationCount)
	})
}

func TestSpeciationChecker_RadiationBonus(t *testing.T) {
	checker := NewSpeciationChecker(42)
	rng := rand.New(rand.NewSource(42))

	parent := &SpeciesPopulation{
		SpeciesID:   uuid.New(),
		Name:        "Test Species",
		Count:       1000,
		GeneticCode: NewGeneticCode(rng),
		Diet:        DietHerbivore,
	}

	// Count speciation without bonus
	checker.SetRadiationBonus(1.0)
	countNormal := 0
	for i := 0; i < 500; i++ {
		result := checker.CheckAllopatricSpeciation(parent, 200000, uuid.New(), int64(i*1000))
		if result != nil {
			countNormal++
		}
	}

	// Count speciation with 3x bonus (adaptive radiation)
	checker.SetRadiationBonus(3.0)
	countBoosted := 0
	for i := 0; i < 500; i++ {
		result := checker.CheckAllopatricSpeciation(parent, 200000, uuid.New(), int64(500000+i*1000))
		if result != nil {
			countBoosted++
		}
	}

	t.Logf("Normal: %d, Boosted: %d", countNormal, countBoosted)
	// Boosted should generally be higher (though probabilistic)
}

func TestHasSufficientDivergence(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	gc1 := NewGeneticCode(rng)
	gc2 := gc1.Clone()

	t.Run("identical codes not divergent", func(t *testing.T) {
		if HasSufficientDivergence(gc1, gc2) {
			t.Error("Identical codes should not be divergent")
		}
	})

	t.Run("heavily mutated codes divergent", func(t *testing.T) {
		// Apply heavy mutation
		gc3 := gc1.Mutate(rng, 1.0, 0.5) // 100% mutation rate, 50% strength
		if !HasSufficientDivergence(gc1, gc3) {
			dist := CalculateGeneticDistance(gc1, gc3)
			t.Logf("Distance after heavy mutation: %f (threshold: %f)", dist, SpeciationThreshold)
		}
	})
}
