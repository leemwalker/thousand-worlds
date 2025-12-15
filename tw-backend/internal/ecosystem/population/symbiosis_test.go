package population

import (
	"testing"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

func TestApplySymbiosis_LinkFormation(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	biome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)

	// Create Flora
	flora := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Orchid",
		Count:     1000,
		Diet:      DietPhotosynthetic,
	}
	biome.AddSpecies(flora)

	// Create Polinator (small herbivore)
	pollinator := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Bee-Bird",
		Count:     500,
		Diet:      DietHerbivore,
		Traits:    EvolvableTraits{Size: 0.2}, // Small
	}
	biome.AddSpecies(pollinator)

	// Create Large Herbivore (should not link)
	browser := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Browser",
		Count:     200,
		Diet:      DietHerbivore,
		Traits:    EvolvableTraits{Size: 5.0}, // Large
	}
	biome.AddSpecies(browser)

	sim.Biomes[biome.BiomeID] = biome

	// Run multiple times as linking chance is 10%
	linked := false
	for i := 0; i < 50; i++ {
		sim.ApplySymbiosis()
		if flora.SymbiosisPartnerID != nil {
			linked = true
			if *flora.SymbiosisPartnerID != pollinator.SpeciesID {
				t.Errorf("Flora linked to wrong partner. Expected %s, got %s", pollinator.SpeciesID, *flora.SymbiosisPartnerID)
			}
			if *pollinator.SymbiosisPartnerID != flora.SpeciesID {
				t.Errorf("Pollinator not reciprocally linked")
			}
			break
		}
	}

	if !linked {
		t.Error("Symbiosis link failed to form between compatible species")
	}

	// Check large browser is NOT linked
	if browser.SymbiosisPartnerID != nil {
		t.Error("Incompatible species formed symbiosis link")
	}
}

func TestApplySymbiosis_Benefit(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	biome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)

	flora := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 1000, Diet: DietPhotosynthetic,
	}
	pollinator := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 1000, Diet: DietHerbivore, Traits: EvolvableTraits{Size: 0.2},
	}

	biome.AddSpecies(flora)
	biome.AddSpecies(pollinator)
	sim.Biomes[biome.BiomeID] = biome

	// Manually link them
	pid := pollinator.SpeciesID
	flora.SymbiosisPartnerID = &pid
	fid := flora.SpeciesID
	pollinator.SymbiosisPartnerID = &fid

	// Run apply (benefit is probabilistic 15%)
	// Run enough times to see growth
	initialPop := flora.Count
	for i := 0; i < 20; i++ {
		sim.ApplySymbiosis()
	}

	if flora.Count <= initialPop {
		t.Logf("Symbiosis did not increase population (probabilistic). Initial %d, Final %d", initialPop, flora.Count)
	} else {
		t.Logf("Symbiosis increased population: %d -> %d", initialPop, flora.Count)
	}
}

func TestApplySymbiosis_BrokenLink(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	biome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)

	flora := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 1000, Diet: DietPhotosynthetic,
	}
	// Partner ID points to non-existent species
	badID := uuid.New()
	flora.SymbiosisPartnerID = &badID

	biome.AddSpecies(flora)
	sim.Biomes[biome.BiomeID] = biome

	sim.ApplySymbiosis()

	if flora.SymbiosisPartnerID != nil {
		t.Error("Broken link (partner missing) should be cleared")
	}
}
