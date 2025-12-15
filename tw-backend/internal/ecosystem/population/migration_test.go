package population

import (
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

func TestMigrateSpecies(t *testing.T) {
	// Create source and destination biomes
	sourceBiome := &BiomePopulation{
		BiomeID:   uuid.New(),
		BiomeType: geography.BiomeGrassland,
		Species:   make(map[uuid.UUID]*SpeciesPopulation),
	}
	destBiome := &BiomePopulation{
		BiomeID:   uuid.New(),
		BiomeType: geography.BiomeGrassland,
		Species:   make(map[uuid.UUID]*SpeciesPopulation),
	}

	// Add species to source
	species := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Test Grazer",
		Count:     1000,
		Traits:    DefaultTraitsForDiet(DietHerbivore),
		Diet:      DietHerbivore,
	}
	sourceBiome.AddSpecies(species)

	// Migrate 10% of population
	migrated := MigrateSpecies(sourceBiome, destBiome, species.SpeciesID, 0.1)

	if migrated != 100 {
		t.Errorf("Expected 100 migrants, got %d", migrated)
	}
	if species.Count != 900 {
		t.Errorf("Source population should be 900, got %d", species.Count)
	}
	// Check destination has new species
	if len(destBiome.Species) == 0 {
		t.Error("Destination should have migrated species")
	}
}

func TestMigrateSpecies_IncompatibleBiome(t *testing.T) {
	// Desert species can't migrate to ocean
	sourceBiome := &BiomePopulation{
		BiomeID:   uuid.New(),
		BiomeType: geography.BiomeDesert,
		Species:   make(map[uuid.UUID]*SpeciesPopulation),
	}
	destBiome := &BiomePopulation{
		BiomeID:   uuid.New(),
		BiomeType: geography.BiomeOcean,
		Species:   make(map[uuid.UUID]*SpeciesPopulation),
	}

	species := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Desert Lizard",
		Count:     500,
		Traits:    DefaultTraitsForDiet(DietHerbivore),
		Diet:      DietHerbivore,
	}
	sourceBiome.AddSpecies(species)

	// Should not migrate land species to ocean
	migrated := MigrateSpecies(sourceBiome, destBiome, species.SpeciesID, 0.2)

	if migrated > 0 {
		t.Errorf("Should not migrate land species to ocean, got %d migrants", migrated)
	}
}

func TestBiomeTransition(t *testing.T) {
	biome := &BiomePopulation{
		BiomeID:   uuid.New(),
		BiomeType: geography.BiomeRainforest,
		Species:   make(map[uuid.UUID]*SpeciesPopulation),
	}

	// Add tropical species
	species := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Tropical Bird",
		Count:     500,
		Traits:    EvolvableTraits{HeatResistance: 0.8, ColdResistance: 0.2},
		Diet:      DietOmnivore,
	}
	biome.AddSpecies(species)

	// Transition to temperate (ice age effect)
	TransitionBiome(biome, geography.BiomeDeciduousForest, 0.5)

	if biome.BiomeType != geography.BiomeDeciduousForest {
		t.Errorf("Biome should transition to Deciduous Forest, got %s", biome.BiomeType)
	}

	// Check species population was impacted
	if species.Count >= 500 {
		t.Error("Species should suffer from biome transition")
	}
}

func TestCalculateMigrationChance(t *testing.T) {
	tests := []struct {
		name       string
		population int64
		variance   float64
		minChance  float64
		maxChance  float64
	}{
		{
			name:       "Small stable population",
			population: 100,
			variance:   0.1,
			minChance:  0.0,
			maxChance:  0.1,
		},
		{
			name:       "Large diverse population",
			population: 5000,
			variance:   0.8,
			minChance:  0.05,
			maxChance:  0.4,
		},
		{
			name:       "Overcrowded population",
			population: 10000,
			variance:   0.5,
			minChance:  0.1,
			maxChance:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			species := &SpeciesPopulation{
				Count:         tt.population,
				TraitVariance: tt.variance,
			}
			chance := CalculateMigrationChance(species, 5000) // 5000 carrying capacity

			if chance < tt.minChance || chance > tt.maxChance {
				t.Errorf("Migration chance = %.2f, expected between %.2f and %.2f",
					chance, tt.minChance, tt.maxChance)
			}
		})
	}
}

func TestGetBiomeTransitionTarget(t *testing.T) {
	tests := []struct {
		currentBiome  geography.BiomeType
		event         string
		expectedBiome geography.BiomeType
	}{
		{geography.BiomeRainforest, "ice_age", geography.BiomeDeciduousForest},
		{geography.BiomeGrassland, "ice_age", geography.BiomeTundra},
		{geography.BiomeTundra, "warming", geography.BiomeGrassland},
		{geography.BiomeDesert, "rainfall_increase", geography.BiomeGrassland},
	}

	for _, tt := range tests {
		t.Run(string(tt.currentBiome)+"_"+tt.event, func(t *testing.T) {
			result := GetBiomeTransitionTarget(tt.currentBiome, tt.event)
			if result != tt.expectedBiome {
				t.Errorf("GetBiomeTransitionTarget(%s, %s) = %s, expected %s",
					tt.currentBiome, tt.event, result, tt.expectedBiome)
			}
		})
	}
}
