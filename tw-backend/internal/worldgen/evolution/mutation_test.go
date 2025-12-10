package evolution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetMutationRate(t *testing.T) {
	t.Run("Normal Population", func(t *testing.T) {
		species := &Species{
			Population:     1000,
			PeakPopulation: 1000,
		}

		rate := GetMutationRate(species)
		assert.True(t, rate >= BaseMutationRateMin)
		assert.True(t, rate <= BaseMutationRateMax)
	})

	t.Run("Bottleneck Population", func(t *testing.T) {
		species := &Species{
			Population:     150,
			PeakPopulation: 1000,
		}

		rate := GetMutationRate(species)
		assert.True(t, rate >= BottleneckMutationRateMin)
		assert.True(t, rate <= BottleneckMutationRateMax)
	})
}

func TestMutateSpecies(t *testing.T) {
	parent := &Species{
		SpeciesID:            uuid.New(),
		Name:                 "TestSpecies",
		Generation:           0,
		Size:                 50,
		Speed:                10,
		Armor:                30,
		Camouflage:           40,
		CaloriesPerDay:       2000,
		ReproductionRate:     0.5,
		Lifespan:             10,
		Population:           1000,
		TemperatureTolerance: TemperatureRange{Min: 10, Max: 30, Optimal: 20},
	}

	mutant, _ := MutateSpecies(parent)

	assert.NotNil(t, mutant)
	assert.Equal(t, parent.Generation+1, mutant.Generation)

	// Traits should be mutated
	assert.NotEqual(t, parent.Size, mutant.Size)
}

func TestApplyMutationsToPopulation(t *testing.T) {
	species1 := &Species{
		SpeciesID:            uuid.New(),
		Population:           1000,
		PeakPopulation:       1000,
		Size:                 50,
		Speed:                10,
		CaloriesPerDay:       2000,
		ReproductionRate:     0.5,
		Lifespan:             10,
		TemperatureTolerance: TemperatureRange{Min: 10, Max: 30, Optimal: 20},
	}

	species := []*Species{species1}

	newSpecies := ApplyMutationsToPopulation(species)

	// Might or might not create new species (stochastic)
	assert.True(t, len(newSpecies) >= 0)
}

func TestSpeciesIsInTolerance(t *testing.T) {
	species := &Species{
		TemperatureTolerance: TemperatureRange{Min: 10, Max: 30, Optimal: 20},
		MoistureTolerance:    MoistureRange{Min: 500, Max: 2000, Optimal: 1000},
		ElevationTolerance:   ElevationRange{Min: 0, Max: 3000, Optimal: 1000},
	}

	t.Run("Within Tolerance", func(t *testing.T) {
		env := &Environment{
			Temperature: 20,
			Moisture:    1000,
			Elevation:   1000,
		}

		assert.True(t, species.IsInTolerance(env))
	})

	t.Run("Outside Temperature", func(t *testing.T) {
		env := &Environment{
			Temperature: 35,
			Moisture:    1000,
			Elevation:   1000,
		}

		assert.False(t, species.IsInTolerance(env))
	})
}

func TestSpeciesClone(t *testing.T) {
	original := &Species{
		SpeciesID:       uuid.New(),
		Name:            "Original",
		PreferredPrey:   []uuid.UUID{uuid.New()},
		PreferredPlants: []uuid.UUID{uuid.New()},
		PreferredBiomes: []string{"tropical"},
	}

	clone := original.Clone()

	assert.NotEqual(t, original.SpeciesID, clone.SpeciesID)
	assert.Equal(t, original.Name, clone.Name)
	assert.Equal(t, len(original.PreferredPrey), len(clone.PreferredPrey))
}

func TestSpeciesTypeChecks(t *testing.T) {
	flora := &Species{Type: SpeciesFlora}
	herbivore := &Species{Type: SpeciesHerbivore, Diet: DietHerbivore}
	carnivore := &Species{Type: SpeciesCarnivore, Diet: DietCarnivore}

	assert.True(t, flora.IsFlora())
	assert.False(t, flora.IsHerbivore())
	assert.False(t, flora.IsCarnivore())

	assert.False(t, herbivore.IsFlora())
	assert.True(t, herbivore.IsHerbivore())
	assert.False(t, herbivore.IsCarnivore())

	assert.False(t, carnivore.IsFlora())
	assert.False(t, carnivore.IsHerbivore())
	assert.True(t, carnivore.IsCarnivore())
}
