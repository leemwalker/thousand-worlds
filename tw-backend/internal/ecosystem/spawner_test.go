package ecosystem

import (
	"testing"
	"tw-backend/internal/ecosystem/state"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

func TestSpawner_SpawnEntitiesForBiome(t *testing.T) {
	// Initialize seeded spawner for deterministic tests
	spawner := NewSpawner(12345)

	// 1. Test Desert
	entities := spawner.SpawnEntitiesForBiome(geography.BiomeDesert, 10)
	assert.Len(t, entities, 10)
	for _, e := range entities {
		isDesertSpecies := e.Species == state.SpeciesCactus ||
			e.Species == state.SpeciesLizard ||
			e.Species == state.SpeciesScorpion ||
			e.Species == state.SpeciesVulture
		assert.True(t, isDesertSpecies, "Entity %s should be a desert species", e.Species)
		assert.NotEmpty(t, e.EntityID)
		assert.Equal(t, 1, e.Generation)
	}

	// 2. Test Ocean
	entities = spawner.SpawnEntitiesForBiome(geography.BiomeOcean, 5)
	assert.Len(t, entities, 5)
	for _, e := range entities {
		assert.Equal(t, state.SpeciesKelp, e.Species)
	}

	// 3. Test Invalid Biome
	entities = spawner.SpawnEntitiesForBiome("Space", 5) // Invalid mock
	if entities != nil {
		// Generic fallback might trigger if we don't handle it, but getSpeciesForBiome has default
		// Default is Grassland species
		assert.Len(t, entities, 5)
	}
}

func TestSpawner_DietAssignment(t *testing.T) {
	spawner := NewSpawner(1)

	wolf := spawner.CreateEntity(state.SpeciesWolf, 1)
	assert.Equal(t, state.DietCarnivore, wolf.Diet)

	cactus := spawner.CreateEntity(state.SpeciesCactus, 1)
	assert.Equal(t, state.DietPhotosynthetic, cactus.Diet)

	rabbit := spawner.CreateEntity(state.SpeciesRabbit, 1)
	assert.Equal(t, state.DietHerbivore, rabbit.Diet)
}
