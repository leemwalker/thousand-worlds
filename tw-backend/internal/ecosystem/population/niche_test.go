package population

import (
	"math"
	"testing"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

func TestNicheOverlapCalculation(t *testing.T) {
	// Case 1: Identical species
	s1 := &SpeciesPopulation{
		Diet:   DietHerbivore,
		Traits: EvolvableTraits{Size: 5.0, NightVision: 0.5},
	}
	s2 := &SpeciesPopulation{
		Diet:   DietHerbivore,
		Traits: EvolvableTraits{Size: 5.0, NightVision: 0.5},
	}

	overlap := calculateNicheOverlap(s1, s2)
	if math.Abs(overlap-1.0) > 0.001 {
		t.Errorf("Identical species should have 1.0 overlap, got %.2f", overlap)
	}

	// Case 2: Different Diets
	s2.Diet = DietCarnivore
	overlap = calculateNicheOverlap(s1, s2)
	if overlap != 0.0 {
		t.Errorf("Different diets should have 0.0 overlap, got %.2f", overlap)
	}

	// Case 3: Different Sizes (Size 5 vs Size 10)
	// Diff = 5, Max = 10, Overlap = 1 - 0.5 = 0.5
	// Weighted = 0.5*0.7 + 1.0*0.3 = 0.35 + 0.3 = 0.65
	s2.Diet = DietHerbivore
	s2.Traits.Size = 10.0
	s1.Traits.Size = 5.0
	overlap = calculateNicheOverlap(s1, s2)
	if math.Abs(overlap-0.65) > 0.001 {
		t.Errorf("Expected 0.65 overlap, got %.2f", overlap)
	}
}

func TestApplyNichePartitioning_Competition(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)

	// Create two identical species
	s1 := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 1000,
		Diet:   DietHerbivore,
		Traits: EvolvableTraits{Size: 5.0, NightVision: 0.5},
	}
	s2 := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 1000,
		Diet:   DietHerbivore,
		Traits: EvolvableTraits{Size: 5.0, NightVision: 0.5},
	}

	biome.AddSpecies(s1)
	biome.AddSpecies(s2)
	sim.Biomes[biome.BiomeID] = biome

	// Run niche partitioning
	// With 1.0 overlap and high density, they should suffer penalties
	initialCount := s1.Count

	// Force deterministic behavior or run enough times
	for i := 0; i < 20; i++ {
		sim.ApplyNichePartitioning()
	}

	if s1.Count >= initialCount {
		t.Errorf("Species 1 population should decrease due to competition. Got %d", s1.Count)
	}
	if s2.Count >= initialCount {
		t.Errorf("Species 2 population should decrease due to competition. Got %d", s2.Count)
	}
}

func TestApplyNichePartitioning_Divergence(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)

	// Two overlapping species with slight difference
	s1 := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 1000,
		Diet:   DietHerbivore,
		Traits: EvolvableTraits{Size: 5.0, NightVision: 0.5},
	}
	s2 := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 1000,
		Diet:   DietHerbivore,
		Traits: EvolvableTraits{Size: 5.1, NightVision: 0.5}, // Slightly larger
	}

	biome.AddSpecies(s1)
	biome.AddSpecies(s2)
	sim.Biomes[biome.BiomeID] = biome

	// Run many iterations to see divergence
	for i := 0; i < 50; i++ {
		sim.ApplyNichePartitioning()
	}

	// s2 started larger, so it should get larger. s1 should get smaller.
	if s2.Traits.Size <= 5.1 {
		t.Logf("Warning: s2 size did not increase significantly (got %.2f), depends on RNG", s2.Traits.Size)
	}
	if s1.Traits.Size >= 5.0 {
		t.Logf("Warning: s1 size did not decrease significantly (got %.2f)", s1.Traits.Size)
	}

	// Check that the difference increased
	finalDiff := s2.Traits.Size - s1.Traits.Size
	if finalDiff <= 0.1 {
		t.Errorf("Traits did not diverge. Initial diff 0.1, final diff %.2f", finalDiff)
	}
}
