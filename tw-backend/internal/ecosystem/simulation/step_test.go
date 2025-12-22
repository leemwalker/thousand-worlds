package simulation

import (
	"testing"

	"tw-backend/internal/ecosystem/population"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestStep_BasicExecution verifies the unified step function works
func TestStep_BasicExecution(t *testing.T) {
	worldID := uuid.New()
	seed := int64(12345)

	popSim := population.NewPopulationSimulator(worldID, seed)

	// Add a test biome with species
	biome := population.NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	flora := &population.SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Test Flora",
		Count:     1000,
		Traits:    population.DefaultTraitsForDiet(population.DietPhotosynthetic),
		Diet:      population.DietPhotosynthetic,
	}
	biome.AddSpecies(flora)
	popSim.Biomes[biome.BiomeID] = biome

	config := DefaultStepConfig()
	subsystems := Subsystems{}

	// Run 100 years
	result := Step(popSim, subsystems, 100, config)

	assert.Equal(t, int64(100), result.YearsAdvanced)
	assert.Equal(t, int64(100), popSim.CurrentYear)
}

// TestStep_DisabledLife verifies simulation skips when life is disabled
func TestStep_DisabledLife(t *testing.T) {
	worldID := uuid.New()
	seed := int64(12345)

	popSim := population.NewPopulationSimulator(worldID, seed)
	initialYear := popSim.CurrentYear

	config := StepConfig{
		SimulateLife:     false, // Disabled
		SimulateGeology:  true,
		SimulateDiseases: false,
	}
	subsystems := Subsystems{}

	// Run 100 years with life disabled
	result := Step(popSim, subsystems, 100, config)

	assert.Equal(t, int64(100), result.YearsAdvanced)
	// PopSim year doesn't advance when SimulateLife is false because SimulateYear is skipped
	assert.Equal(t, initialYear, popSim.CurrentYear)
}

// TestApplyPeriodicEvolution verifies evolution only runs at 1000 year intervals
func TestApplyPeriodicEvolution(t *testing.T) {
	worldID := uuid.New()
	seed := int64(12345)

	popSim := population.NewPopulationSimulator(worldID, seed)
	config := DefaultStepConfig()

	// At year 0, should return false
	assert.False(t, ApplyPeriodicEvolution(popSim, config))

	// At year 999, should return false
	popSim.CurrentYear = 999
	assert.False(t, ApplyPeriodicEvolution(popSim, config))

	// At year 1000, should return true
	popSim.CurrentYear = 1000
	assert.True(t, ApplyPeriodicEvolution(popSim, config))
}

// TestApplyPeriodicSpeciation verifies speciation only runs at 10000 year intervals
func TestApplyPeriodicSpeciation(t *testing.T) {
	worldID := uuid.New()
	seed := int64(12345)

	popSim := population.NewPopulationSimulator(worldID, seed)
	config := DefaultStepConfig()

	// At year 0, should return 0
	newSpecies, migrants := ApplyPeriodicSpeciation(popSim, config)
	assert.Equal(t, 0, newSpecies)
	assert.Equal(t, int64(0), migrants)

	// At year 10000, function is called but result depends on population state
	popSim.CurrentYear = 10000
	// Add biomes for speciation to work with
	biome := population.NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	flora := &population.SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Test Flora",
		Count:     1000,
		Traits:    population.DefaultTraitsForDiet(population.DietPhotosynthetic),
		Diet:      population.DietPhotosynthetic,
	}
	biome.AddSpecies(flora)
	popSim.Biomes[biome.BiomeID] = biome

	// Should run without error at 10000
	newSpecies, migrants = ApplyPeriodicSpeciation(popSim, config)
	// Values depend on internal state, we just verify no panic
	assert.GreaterOrEqual(t, newSpecies, 0)
	assert.GreaterOrEqual(t, migrants, int64(0))
}

// TestShouldUpdateGeology verifies geology update timing
func TestShouldUpdateGeology(t *testing.T) {
	config := DefaultStepConfig()

	// At year 0, should return false
	assert.False(t, ShouldUpdateGeology(0, config))

	// At year 9999, should return false
	assert.False(t, ShouldUpdateGeology(9999, config))

	// At year 10000, should return true
	assert.True(t, ShouldUpdateGeology(10000, config))

	// At year 100000, should return true
	assert.True(t, ShouldUpdateGeology(100000, config))

	// With geology disabled
	config.SimulateGeology = false
	assert.False(t, ShouldUpdateGeology(10000, config))
}

// TestStep_EventsGenerated verifies events are generated correctly
func TestStep_EventsGenerated(t *testing.T) {
	worldID := uuid.New()
	seed := int64(12345)

	popSim := population.NewPopulationSimulator(worldID, seed)

	// Set up biomes that might produce speciation
	biome := population.NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	for i := 0; i < 3; i++ {
		species := &population.SpeciesPopulation{
			SpeciesID:     uuid.New(),
			Name:          "Test Species",
			Count:         1000,
			Traits:        population.DefaultTraitsForDiet(population.DietHerbivore),
			TraitVariance: 0.5,
			Diet:          population.DietHerbivore,
		}
		biome.AddSpecies(species)
	}
	popSim.Biomes[biome.BiomeID] = biome

	var receivedEvents []SimulationEvent
	config := StepConfig{
		SimulateLife:     true,
		SimulateGeology:  false,
		SimulateDiseases: false,
		EventHandler: func(event SimulationEvent) {
			receivedEvents = append(receivedEvents, event)
		},
	}
	subsystems := Subsystems{}

	// Run to year 10000 to trigger speciation check
	result := Step(popSim, subsystems, 10000, config)

	assert.Equal(t, int64(10000), result.YearsAdvanced)
	// Events in result should match events sent to handler
	assert.Equal(t, len(result.Events), len(receivedEvents))
}
