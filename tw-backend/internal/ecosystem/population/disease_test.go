package population

import (
	"testing"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

func TestApplyDisease_OutbreakChance(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	biome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)
	biome.CarryingCapacity = 1000

	// Dense population (2000 > 1000)
	denseSpecies := &SpeciesPopulation{
		SpeciesID: uuid.New(), Name: "Crowded Rat", Count: 2000,
		Diet: DietHerbivore, Traits: EvolvableTraits{DiseaseResistance: 0.0},
	}
	biome.AddSpecies(denseSpecies)
	sim.Biomes[biome.BiomeID] = biome

	// Dense population means density = 2.0
	// Chance = 0.01 + 2*2*5 = 20.01 -> capped at 0.2 (20%)

	outbreaks := 0
	iterations := 100
	for i := 0; i < iterations; i++ {
		// Reset count to maintain density
		denseSpecies.Count = 2000
		if sim.ApplyDisease() > 0 {
			outbreaks++
		}
	}

	if outbreaks == 0 {
		t.Error("Zero outbreaks in dense population, expected ~20%")
	}
	if outbreaks > 40 {
		t.Logf("Warning: High outbreak rate %d/%d (expected ~20)", outbreaks, iterations)
	}
}

func TestApplyDisease_Resistance(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	biome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)
	biome.CarryingCapacity = 1000

	// Two identical populations, one resistant
	vulnerable := &SpeciesPopulation{
		SpeciesID: uuid.New(), Name: "Sickly", Count: 2000,
		Traits: EvolvableTraits{DiseaseResistance: 0.0},
	}
	resistant := &SpeciesPopulation{
		SpeciesID: uuid.New(), Name: "Healthy", Count: 2000,
		Traits: EvolvableTraits{DiseaseResistance: 1.0},
	}

	biome.AddSpecies(vulnerable)
	biome.AddSpecies(resistant)
	sim.Biomes[biome.BiomeID] = biome

	// Force outbreak logic by mocking randomness?
	// Or just run until outbreak happens.

	// Given 20% chance, running 20 times ensures high probability of outbreak
	outbreakOccurred := false
	for i := 0; i < 20; i++ {
		origVul := vulnerable.Count
		origRes := resistant.Count

		if sim.ApplyDisease() > 0 {
			// Check if reduction happened
			if vulnerable.Count < origVul {
				outbreakOccurred = true
				if resistant.Count < origRes {
					t.Error("Resistant species (100% immunity) should not lose population")
				}
				break
			}
		}
		// Reset counts if no outbreak yet
		vulnerable.Count = 2000
		resistant.Count = 2000
	}

	if !outbreakOccurred {
		t.Log("Warning: No outbreak occurred in random test, verify RNG logic")
	}
}

func TestApplyDisease_Evolution(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	biome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)
	biome.CarryingCapacity = 1000

	species := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 2000,
		Traits: EvolvableTraits{DiseaseResistance: 0.0},
	}
	biome.AddSpecies(species)
	sim.Biomes[biome.BiomeID] = biome

	// Run until resistance increases
	initialRes := species.Traits.DiseaseResistance
	evolved := false

	for i := 0; i < 50; i++ {
		species.Count = 2000 // Reset pop to keep density high
		sim.ApplyDisease()
		if species.Traits.DiseaseResistance > initialRes {
			evolved = true
			break
		}
	}

	if !evolved {
		t.Error("Species failed to evolve resistance after outbreaks")
	}
}
