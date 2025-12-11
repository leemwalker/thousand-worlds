package ecosystem

import (
	"testing"
	"tw-backend/internal/ecosystem/state"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

func TestService_Integration_Lifecycle(t *testing.T) {
	// 1. Setup Service
	sim := NewService(999)

	// 2. Spawn Biome
	biomes := []geography.Biome{
		{Type: geography.BiomeGrassland},
	}
	sim.SpawnBiomes(biomes)

	// Verify initialization
	assert.NotEmpty(t, sim.Entities)
	var rabbit *state.LivingEntityState

	// Find a rabbit
	for _, e := range sim.Entities {
		if e.Species == state.SpeciesRabbit {
			rabbit = e
			break
		}
	}
	assert.NotNil(t, rabbit, "Should have spawned a rabbit")

	// 3. Sim Loop - Needs Decay
	initialEnergy := rabbit.Needs.Energy
	sim.Tick()
	assert.Less(t, rabbit.Needs.Energy, initialEnergy, "Energy should decay after tick")

	// 4. Test Death
	// Force critical state
	rabbit.Needs.Hunger = 100
	// Remove AI so it doesn't eat
	delete(sim.Behaviors, rabbit.EntityID)
	sim.Tick()

	// Check value
	// Note: Since rabbit is a pointer, if it was deleted from map, the struct it points to still exists in memory.
	// We check the map for existence.
	_, exists := sim.Entities[rabbit.EntityID]
	if exists {
		t.Logf("Rabbit still exists. Hunger: %f", rabbit.Needs.Hunger)
	}
	assert.False(t, exists, "Rabbit should die from starvation")
}
