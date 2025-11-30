package evolution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateClimateFitness(t *testing.T) {
	species := &Species{
		TemperatureTolerance: TemperatureRange{Min: 10, Max: 30, Optimal: 20},
		MoistureTolerance:    MoistureRange{Min: 500, Max: 2000, Optimal: 1000},
	}

	t.Run("Optimal Environment", func(t *testing.T) {
		env := &Environment{Temperature: 20, Moisture: 1000}
		fitness := CalculateClimateFitness(species, env)
		assert.InDelta(t, 1.0, fitness, 0.1)
	})

	t.Run("Suboptimal Temperature", func(t *testing.T) {
		env := &Environment{Temperature: 10, Moisture: 1000}
		fitness := CalculateClimateFitness(species, env)
		assert.True(t, fitness < 1.0)
		assert.True(t, fitness > 0.3)
	})
}

func TestCalculateFoodFitness(t *testing.T) {
	assert.Equal(t, 1.0, CalculateFoodFitness(1.0))
	assert.Equal(t, 0.5, CalculateFoodFitness(0.5))
}

func TestCalculatePredationFitness(t *testing.T) {
	species := &Species{
		Speed:      50,
		Camouflage: 50,
		Armor:      50,
	}

	t.Run("No Predation", func(t *testing.T) {
		fitness := CalculatePredationFitness(species, 0)
		assert.Equal(t, 1.0, fitness)
	})

	t.Run("High Predation with Defenses", func(t *testing.T) {
		fitness := CalculatePredationFitness(species, 0.5)
		assert.True(t, fitness > 0.5)
	})
}

func TestCalculateTotalFitness(t *testing.T) {
	species := &Species{
		SpeciesID:            uuid.New(),
		TemperatureTolerance: TemperatureRange{Min: 10, Max: 30, Optimal: 20},
		MoistureTolerance:    MoistureRange{Min: 500, Max: 2000, Optimal: 1000},
		Speed:                50,
		Camouflage:           50,
		Armor:                50,
		ExtinctionRisk:       0.1,
	}

	env := &Environment{Temperature: 20, Moisture: 1000}

	fitness := CalculateTotalFitness(species, env, 0.8, 0.2, []*Species{})

	assert.True(t, fitness > 0)
	assert.True(t, fitness <= 1.0)
}
