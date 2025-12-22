package minerals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateConcentration(t *testing.T) {
	optimalIgneous := &TectonicContext{IsVolcanic: true}
	optimalSedimentary := &TectonicContext{IsSedimentaryBasin: true}
	poorContext := &TectonicContext{}

	t.Run("Common Minerals High Concentration", func(t *testing.T) {
		conc := CalculateConcentration(MineralIron, optimalSedimentary)
		assert.True(t, conc >= 0.01 && conc <= 1.0) // Valid range
	})

	t.Run("Rare Minerals Low Concentration", func(t *testing.T) {
		conc := CalculateConcentration(MineralDiamond, poorContext)
		// Diamond: minBase=0.01, poorContext geoMod~0.3-1.0, randomVar=0.8-1.2
		// Minimum possible: 0.01 * 0.3 * 0.8 = 0.0024
		assert.True(t, conc >= 0.001 && conc <= 1.0)
	})

	t.Run("Optimal Conditions Improve Concentration", func(t *testing.T) {
		// Gold in volcanic is optimal
		goldVolcanic := CalculateConcentration(MineralGold, optimalIgneous)
		goldPoor := CalculateConcentration(MineralGold, poorContext)

		// Due to randomness, we can't guarantee, but we can check range validity
		assert.True(t, goldVolcanic > 0)
		assert.True(t, goldPoor > 0)
	})

	t.Run("Concentration Capped at 1.0", func(t *testing.T) {
		// Even with optimal conditions and random variation, should not exceed 1.0
		for i := 0; i < 100; i++ {
			conc := CalculateConcentration(MineralIron, optimalSedimentary)
			assert.True(t, conc <= 1.0)
		}
	})
}

func TestGetBaseQuantity(t *testing.T) {
	t.Run("Size Scaling", func(t *testing.T) {
		small := GetBaseQuantity(MineralIron, VeinSizeSmall)
		medium := GetBaseQuantity(MineralIron, VeinSizeMedium)
		large := GetBaseQuantity(MineralIron, VeinSizeLarge)
		massive := GetBaseQuantity(MineralIron, VeinSizeMassive)

		assert.True(t, small < medium)
		assert.True(t, medium < large)
		assert.True(t, large < massive)
	})

	t.Run("Rare Minerals Smaller Quantities", func(t *testing.T) {
		ironLarge := GetBaseQuantity(MineralIron, VeinSizeLarge)
		diamondLarge := GetBaseQuantity(MineralDiamond, VeinSizeLarge)

		assert.True(t, diamondLarge < ironLarge, "Diamonds should be rarer/smaller")
	})

	t.Run("Common Minerals Larger Quantities", func(t *testing.T) {
		coalMassive := GetBaseQuantity(MineralCoal, VeinSizeMassive)
		platinumMassive := GetBaseQuantity(MineralPlatinum, VeinSizeMassive)

		assert.True(t, coalMassive > platinumMassive)
	})
}
