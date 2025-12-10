package evolution

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBiomassProduction(t *testing.T) {
	t.Run("Optimal Conditions", func(t *testing.T) {
		biomass := CalculateBiomassProduction(1.0, 1500, 20)
		assert.True(t, biomass > 0)
		assert.True(t, biomass <= BaseBiomassProductionRate)
	})

	t.Run("Low Moisture", func(t *testing.T) {
		biomass := CalculateBiomassProduction(1.0, 100, 20)
		optimalBiomass := CalculateBiomassProduction(1.0, 1500, 20)
		assert.True(t, biomass < optimalBiomass)
	})

	t.Run("Cold Temperature", func(t *testing.T) {
		biomass := CalculateBiomassProduction(1.0, 1500, 5)
		optimalBiomass := CalculateBiomassProduction(1.0, 1500, 20)
		assert.True(t, biomass < optimalBiomass)
	})
}

func TestCalculateEnergyTransfer(t *testing.T) {
	t.Run("10% Efficiency", func(t *testing.T) {
		consumed := 100.0
		energy := CalculateEnergyTransfer(consumed)
		assert.Equal(t, 10.0, energy)
	})
}

func TestCalculateCarryingCapacity(t *testing.T) {
	capacity := CalculateCarryingCapacity(1000, 2000, 100)
	assert.True(t, capacity > 0)
}
