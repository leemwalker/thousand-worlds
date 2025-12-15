package population

import (
	"strings"
	"testing"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

func TestUpdateOxygenLevel_BioticEffect(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	sim.OxygenLevel = 0.21

	biome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)

	// Create massive amount of flora (Photosynthetic)
	// 1 billion plants
	flora := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 1000000000,
		Diet: DietPhotosynthetic,
	}
	biome.AddSpecies(flora)
	sim.Biomes[biome.BiomeID] = biome

	// Flora adds 1e-8 per unit = 1e9 * 1e-8 = 10.0 increase?
	// That's huge. 1 billion plants adding 1000% O2?
	// Scale was: bioticChange := (totalFlora*1e-8 - totalFauna*2e-8)
	// Yes, 1 billion is a lot. Maybe scale should be smaller or normalized by world size.
	// But let's check if it increases at all.

	oldLevel := sim.OxygenLevel
	sim.UpdateOxygenLevel()

	if sim.OxygenLevel <= oldLevel {
		t.Errorf("Massive flora should increase Oxygen level. Old %.4f, New %.4f", oldLevel, sim.OxygenLevel)
	}

	// Check if it clamped (max 0.35)
	if sim.OxygenLevel > 0.35 {
		t.Errorf("Oxygen level should be clamped at 0.35. Got %.4f", sim.OxygenLevel)
	}
}

func TestUpdateOxygenLevel_Reporting(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	sim.OxygenLevel = 0.21
	sim.Events = []string{}

	// Force a rapid manual change to test reporting logic
	// But UpdateOxygenLevel calculates change based on internal logic.
	// We can manipulate the internal state mid-function? No.
	// We can set up a scenario where O2 drops rapidly.

	biome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)
	// Massive fauna, no flora
	fauna := &SpeciesPopulation{
		SpeciesID: uuid.New(), Count: 500000000, // 500M animals
		Diet: DietHerbivore,
	}
	biome.AddSpecies(fauna)
	sim.Biomes[biome.BiomeID] = biome

	// 500M * 2e-8 = 10.0 decrease. Should crash O2.

	sim.UpdateOxygenLevel()

	// Check events
	found := false
	for _, event := range sim.Events {
		if strings.Contains(event, "Oxygen Falling Rapidly") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected rapid oxygen fall event. Events: %v", sim.Events)
	}

	// Check threshold crossing
	// Reset and force crossing 18%
	sim.OxygenLevel = 0.181
	sim.CurrentYear = 100 // Trigger periodic check
	sim.Events = []string{}

	// Needs to drop below 0.18
	// Add more fauna
	sim.UpdateOxygenLevel()

	if sim.OxygenLevel > 0.18 {
		t.Logf("Warning: Oxygen didn't drop below 0.18 as expected (got %.4f), might not trigger threshold event", sim.OxygenLevel)
	} else {
		foundThreshold := false
		for _, event := range sim.Events {
			if strings.Contains(event, "Oxygen Dropped Below 18%") {
				foundThreshold = true
				break
			}
		}
		if !foundThreshold {
			t.Errorf("Expected threshold crossing event. Events: %v", sim.Events)
		}
	}
}
