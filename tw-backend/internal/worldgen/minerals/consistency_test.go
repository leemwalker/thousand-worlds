package minerals

import (
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

func TestGeologicalConsistency(t *testing.T) {
	// 1. Gold near volcanoes
	volcanicCtx := &TectonicContext{IsVolcanic: true}
	goldDeposit := GenerateMineralVein(volcanicCtx, MineralGold, geography.Point{})
	assert.True(t, goldDeposit.Quantity > 0)
	// Gold should be favored/larger in volcanic (checked in DetermineVeinSize logic)

	// 2. Coal in sedimentary basins
	sedCtx := &TectonicContext{IsSedimentaryBasin: true}
	coalDeposit := GenerateMineralVein(sedCtx, MineralCoal, geography.Point{})
	assert.Equal(t, FormationSedimentary, coalDeposit.FormationType)

	// 3. Iron in metamorphic/convergent
	metaCtx := &TectonicContext{PlateBoundaryType: geography.BoundaryConvergent}
	ironDeposit := GenerateMineralVein(metaCtx, MineralIron, geography.Point{})
	assert.Equal(t, FormationMetamorphic, ironDeposit.FormationType)
}

func TestQuantityDistribution(t *testing.T) {
	// Generate a large number of deposits and check stats
	// This simulates a "world" generation

	ctx := &TectonicContext{
		IsVolcanic:         true,
		IsSedimentaryBasin: true, // Mixed context
	}

	ironCount := 0
	goldCount := 0
	diamondCount := 0

	for i := 0; i < 1000; i++ {
		// Randomly pick a mineral to generate
		// In real worldgen, we'd pick based on biome/tectonics
		// Here we just test the generation function itself

		d1 := GenerateMineralVein(ctx, MineralIron, geography.Point{})
		if d1.Quantity > 10000 {
			ironCount++
		}

		d2 := GenerateMineralVein(ctx, MineralGold, geography.Point{})
		if d2.Quantity > 500 {
			goldCount++
		}

		d3 := GenerateMineralVein(ctx, MineralDiamond, geography.Point{})
		if d3.Quantity > 50 {
			diamondCount++
		}
	}

	// Just verify we can generate them
	assert.True(t, ironCount > 0)
	assert.True(t, goldCount > 0)
	assert.True(t, diamondCount > 0)
}
