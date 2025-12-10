package evolution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateRangeOverlap(t *testing.T) {
	t.Run("Complete Overlap", func(t *testing.T) {
		overlap := CalculateRangeOverlap(10, 20, 10, 20)
		assert.InDelta(t, 1.0, overlap, 0.1)
	})

	t.Run("Partial Overlap", func(t *testing.T) {
		overlap := CalculateRangeOverlap(10, 20, 15, 25)
		assert.True(t, overlap > 0 && overlap < 1.0)
	})

	t.Run("No Overlap", func(t *testing.T) {
		overlap := CalculateRangeOverlap(10, 20, 30, 40)
		assert.Equal(t, 0.0, overlap)
	})
}

func TestDetermineFavoredTraits(t *testing.T) {
	t.Run("High Predation Favors Defenses", func(t *testing.T) {
		favored := DetermineFavoredTraits(0.5, 0.8, 0.3)
		assert.True(t, favored["speed"])
		assert.True(t, favored["camouflage"])
		assert.True(t, favored["armor"])
	})

	t.Run("Low Food Favors Efficiency", func(t *testing.T) {
		favored := DetermineFavoredTraits(0.1, 0.3, 0.2)
		assert.True(t, favored["efficiency"])
	})

	t.Run("High Competition Favors Specialization", func(t *testing.T) {
		favored := DetermineFavoredTraits(0.1, 0.8, 0.8)
		assert.True(t, favored["specialization"])
	})
}

func TestCalculateFoodAvailability(t *testing.T) {
	herbivore := &Species{
		SpeciesID:       uuid.New(),
		Type:            SpeciesHerbivore,
		Population:      100,
		CaloriesPerDay:  2000,
		PreferredPlants: []uuid.UUID{uuid.New()},
	}

	floraID := herbivore.PreferredPlants[0]
	floraPopulations := map[uuid.UUID]int{
		floraID: 1000,
	}

	availability := CalculateFoodAvailability(herbivore, floraPopulations, map[uuid.UUID]int{})
	assert.True(t, availability >= 0 && availability <= 1.0)
}

func TestCalculateHerbivoreAndCarnivoreEnergy(t *testing.T) {
	biomass := 100.0

	herbEnergy := CalculateHerbivoreEnergyGain(biomass)
	assert.Equal(t, 10.0, herbEnergy)

	carnEnergy := CalculateCarnivoreEnergyGain(biomass)
	assert.Equal(t, 10.0, carnEnergy)
}

func TestCalculatePredationPressure(t *testing.T) {
	prey := &Species{
		SpeciesID:  uuid.New(),
		Population: 1000,
		Speed:      10,
		Camouflage: 20,
	}

	predator := &Species{
		SpeciesID:     uuid.New(),
		Population:    50,
		Speed:         15,
		PreferredPrey: []uuid.UUID{prey.SpeciesID},
	}

	pressure := CalculatePredationPressure(prey, []*Species{predator})
	assert.True(t, pressure >= 0 && pressure <= 1.0)
}

func TestSpeciesIsBottlenecked(t *testing.T) {
	t.Run("Not Bottlenecked", func(t *testing.T) {
		s := &Species{Population: 900, PeakPopulation: 1000}
		assert.False(t, s.IsBottlenecked())
	})

	t.Run("Bottlenecked", func(t *testing.T) {
		s := &Species{Population: 150, PeakPopulation: 1000}
		assert.True(t, s.IsBottlenecked())
	})
}
