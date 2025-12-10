package minerals

import (
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

func TestGenerateVeinGeometry(t *testing.T) {
	ctx := &TectonicContext{
		MagmaFlowDirection: geography.Vector{X: 1, Y: 0},
		FaultLineDirection: geography.Vector{X: 0, Y: 1},
	}

	t.Run("Igneous Linear", func(t *testing.T) {
		shape, orient, length, width := GenerateVeinGeometry(MineralGold, ctx, VeinSizeMedium)
		assert.Equal(t, VeinShapeLinear, shape)
		assert.Equal(t, ctx.MagmaFlowDirection, orient)
		assert.True(t, length > 0)
		assert.True(t, width > 0)
	})

	t.Run("Sedimentary Planar", func(t *testing.T) {
		shape, _, length, width := GenerateVeinGeometry(MineralCoal, ctx, VeinSizeLarge)
		assert.Equal(t, VeinShapePlanar, shape)
		assert.True(t, length > 0)
		assert.True(t, width > 0)
		// Sedimentary usually wider/longer
		assert.True(t, length >= 1000)
	})

	t.Run("Size Scaling", func(t *testing.T) {
		_, _, lenSmall, widthSmall := GenerateVeinGeometry(MineralGold, ctx, VeinSizeSmall)
		_, _, lenMassive, widthMassive := GenerateVeinGeometry(MineralGold, ctx, VeinSizeMassive)

		assert.True(t, lenMassive > lenSmall)
		assert.True(t, widthMassive > widthSmall)
	})
}
