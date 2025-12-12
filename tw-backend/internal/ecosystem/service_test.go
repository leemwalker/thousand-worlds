package ecosystem

import (
	"testing"
	"tw-backend/internal/ecosystem/state"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	sim := NewService(12345)

	assert.NotNil(t, sim.Entities)
	assert.NotNil(t, sim.Spawner)
	assert.NotNil(t, sim.Needs)
	assert.NotNil(t, sim.Planner)
	assert.NotNil(t, sim.EvolutionManager)
	assert.NotNil(t, sim.Behaviors)
	assert.Equal(t, int64(12345), sim.Spawner.Seed)
}

func TestService_SpawnBiomes(t *testing.T) {
	sim := NewService(999)

	biomes := []geography.Biome{
		{Type: geography.BiomeGrassland},
	}
	sim.SpawnBiomes(biomes)

	// Grassland should spawn 5 entities (default count)
	assert.Equal(t, 5, len(sim.Entities), "Should spawn 5 entities for grassland")
	assert.Equal(t, 5, len(sim.Behaviors), "Each entity should have a behavior")
}

func TestService_SpawnBiomes_DesertSparse(t *testing.T) {
	sim := NewService(999)

	biomes := []geography.Biome{
		{Type: geography.BiomeDesert},
	}
	sim.SpawnBiomes(biomes)

	// Desert should spawn only 2 entities (sparse)
	assert.Equal(t, 2, len(sim.Entities), "Should spawn 2 entities for desert")
}

func TestService_SpawnBiomes_RainforestDense(t *testing.T) {
	sim := NewService(999)

	biomes := []geography.Biome{
		{Type: geography.BiomeRainforest},
	}
	sim.SpawnBiomes(biomes)

	// Rainforest should spawn 10 entities (dense)
	assert.Equal(t, 10, len(sim.Entities), "Should spawn 10 entities for rainforest")
}

func TestService_Tick_NeedsDecay(t *testing.T) {
	sim := NewService(999)

	// Manually create an entity to test tick behavior
	entity := &state.LivingEntityState{
		EntityID: uuid.New(),
		Species:  state.SpeciesRabbit,
		Diet:     state.DietHerbivore,
		Needs: state.NeedState{
			Energy: 100,
			Hunger: 0,
			Thirst: 0,
			Safety: 100,
		},
	}
	sim.Entities[entity.EntityID] = entity

	initialEnergy := entity.Needs.Energy
	sim.Tick()

	// Energy should decay after tick
	assert.Less(t, entity.Needs.Energy, initialEnergy, "Energy should decay after tick")
}

func TestService_Tick_DeathFromStarvation(t *testing.T) {
	sim := NewService(999)

	// Create a starving entity
	entity := &state.LivingEntityState{
		EntityID: uuid.New(),
		Species:  state.SpeciesRabbit,
		Diet:     state.DietHerbivore,
		Needs: state.NeedState{
			Hunger: 100, // Starving
			Energy: 10,
		},
	}
	sim.Entities[entity.EntityID] = entity

	sim.Tick()

	// Entity should be dead
	_, exists := sim.Entities[entity.EntityID]
	assert.False(t, exists, "Entity should die from starvation")
}

func TestService_Tick_FloraDoesNotDie(t *testing.T) {
	sim := NewService(999)

	// Create flora with "critical" needs (should not die)
	flora := &state.LivingEntityState{
		EntityID: uuid.New(),
		Species:  state.SpeciesGrass,
		Diet:     state.DietPhotosynthetic,
		Needs: state.NeedState{
			Hunger: 100, // Flora ignores hunger
		},
	}
	sim.Entities[flora.EntityID] = flora

	sim.Tick()

	// Flora should still exist
	_, exists := sim.Entities[flora.EntityID]
	assert.True(t, exists, "Flora should not die from hunger")
}

func TestService_GetEntity(t *testing.T) {
	sim := NewService(999)

	entity := &state.LivingEntityState{
		EntityID: uuid.New(),
		Species:  state.SpeciesRabbit,
	}
	sim.Entities[entity.EntityID] = entity

	found := sim.GetEntity(entity.EntityID)
	require.NotNil(t, found)
	assert.Equal(t, entity.EntityID, found.EntityID)

	// Non-existent entity
	notFound := sim.GetEntity(uuid.New())
	assert.Nil(t, notFound)
}

func TestService_GetEntitiesAt(t *testing.T) {
	sim := NewService(999)
	worldID := uuid.New()

	// Create entities at different positions
	e1 := &state.LivingEntityState{
		EntityID:  uuid.New(),
		WorldID:   worldID,
		PositionX: 100,
		PositionY: 100,
	}
	e2 := &state.LivingEntityState{
		EntityID:  uuid.New(),
		WorldID:   worldID,
		PositionX: 105,
		PositionY: 105,
	}
	e3 := &state.LivingEntityState{
		EntityID:  uuid.New(),
		WorldID:   worldID,
		PositionX: 500,
		PositionY: 500,
	}
	sim.Entities[e1.EntityID] = e1
	sim.Entities[e2.EntityID] = e2
	sim.Entities[e3.EntityID] = e3

	// Query near (100, 100) with radius 20
	found := sim.GetEntitiesAt(worldID, 100, 100, 20)
	assert.Len(t, found, 2, "Should find 2 entities within radius")

	// Query with different world ID
	otherWorld := uuid.New()
	notFound := sim.GetEntitiesAt(otherWorld, 100, 100, 1000)
	assert.Len(t, notFound, 0, "Should find no entities in other world")
}

func TestService_GetEvolutionManager(t *testing.T) {
	sim := NewService(999)
	em := sim.GetEvolutionManager()
	assert.NotNil(t, em)
	assert.Equal(t, sim.EvolutionManager, em)
}
