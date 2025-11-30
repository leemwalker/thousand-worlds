package evolution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePredationSuccess(t *testing.T) {
	t.Run("Fast Predator vs Slow Prey", func(t *testing.T) {
		success := CalculatePredationSuccess(20, 10, 0, 0.8)
		assert.True(t, success > 0.5)
	})

	t.Run("Slow Predator vs Fast Prey", func(t *testing.T) {
		success := CalculatePredationSuccess(10, 20, 0, 0.8)
		assert.True(t, success < 0.5)
	})

	t.Run("High Camouflage Reduces Success", func(t *testing.T) {
		noCamo := CalculatePredationSuccess(15, 15, 0, 0.8)
		highCamo := CalculatePredationSuccess(15, 15, 80, 0.8)
		assert.True(t, noCamo > highCamo)
	})
}

func TestConsumePreyPopulation(t *testing.T) {
	predator := &Species{
		SpeciesID:  uuid.New(),
		Population: 100,
		Speed:      15,
	}

	prey := &Species{
		SpeciesID:  uuid.New(),
		Population: 1000,
		Speed:      10,
		Camouflage: 20,
	}

	preyKilled := ConsumePreyPopulation(predator, prey, 0.5)
	assert.True(t, preyKilled > 0)
	assert.True(t, preyKilled <= prey.Population)
}

func TestCalculateIntraspecificCompetition(t *testing.T) {
	t.Run("Below Capacity", func(t *testing.T) {
		survival := CalculateIntraspecificCompetition(100, 200)
		assert.Equal(t, 1.0, survival)
	})

	t.Run("Above Capacity", func(t *testing.T) {
		survival := CalculateIntraspecificCompetition(200, 100)
		assert.Equal(t, 0.5, survival)
	})
}

func TestCalculateNicheOverlap(t *testing.T) {
	species1 := &Species{
		SpeciesID:            uuid.New(),
		Diet:                 DietHerbivore,
		PreferredBiomes:      []string{"tropical"},
		TemperatureTolerance: TemperatureRange{Min: 15, Max: 30, Optimal: 22},
	}

	species2 := &Species{
		SpeciesID:            uuid.New(),
		Diet:                 DietHerbivore,
		PreferredBiomes:      []string{"tropical"},
		TemperatureTolerance: TemperatureRange{Min: 18, Max: 32, Optimal: 25},
	}

	overlap := CalculateNicheOverlap(species1, species2)
	assert.True(t, overlap > 0.5) // High overlap

	// Different diet
	species2.Diet = DietCarnivore
	overlap = CalculateNicheOverlap(species1, species2)
	assert.True(t, overlap < 0.8) // Lower overlap
}

func TestCalculateInterspecificCompetition(t *testing.T) {
	species := &Species{
		SpeciesID:            uuid.New(),
		Diet:                 DietHerbivore,
		PreferredBiomes:      []string{"temperate"},
		TemperatureTolerance: TemperatureRange{Min: 10, Max: 25, Optimal: 18},
	}

	competitor := &Species{
		SpeciesID:            uuid.New(),
		Diet:                 DietHerbivore,
		PreferredBiomes:      []string{"temperate"},
		TemperatureTolerance: TemperatureRange{Min: 12, Max: 27, Optimal: 20},
		PopulationDensity:    50,
	}

	fitness := CalculateInterspecificCompetition(species, []*Species{competitor})
	assert.True(t, fitness < 1.0)
	assert.True(t, fitness > 0)
}
