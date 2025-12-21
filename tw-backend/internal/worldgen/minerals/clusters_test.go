package minerals

import (
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCluster(t *testing.T) {
	ctx := &TectonicContext{
		IsVolcanic:         true,
		MagmaFlowDirection: geography.Vector{X: 1, Y: 0},
	}
	primary := GenerateMineralVein(ctx, MineralGold, geography.Point{X: 0, Y: 0})
	primary.VeinSize = VeinSizeLarge // Force large for testing

	cluster := GenerateCluster(primary, ctx)

	assert.NotEmpty(t, cluster)
	assert.Equal(t, primary, cluster[0])

	// Check consistency
	for _, deposit := range cluster {
		assert.Equal(t, primary.MineralType.Name, deposit.MineralType.Name)
		assert.True(t, deposit.Quantity > 0)
	}
}

func TestGenerateCluster_MultipleIterations(t *testing.T) {
	// Run multiple iterations to ensure satellite vein generation is exercised
	ctx := &TectonicContext{
		IsVolcanic:         true,
		MagmaFlowDirection: geography.Vector{X: 1, Y: 0},
	}

	totalDeposits := 0
	for i := 0; i < 50; i++ {
		primary := GenerateMineralVein(ctx, MineralCopper, geography.Point{X: 0, Y: 0})
		primary.VeinSize = VeinSizeMassive
		cluster := GenerateCluster(primary, ctx)
		totalDeposits += len(cluster)
	}

	// Should have generated some secondary/tertiary veins over 50 iterations
	assert.Greater(t, totalDeposits, 50, "Should generate secondary/tertiary veins")
}

func TestDowngradeSize(t *testing.T) {
	tests := []struct {
		input    VeinSize
		expected VeinSize
	}{
		{VeinSizeMassive, VeinSizeLarge},
		{VeinSizeLarge, VeinSizeMedium},
		{VeinSizeMedium, VeinSizeSmall},
		{VeinSizeSmall, VeinSizeSmall},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := downgradeSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSatelliteVein(t *testing.T) {
	ctx := &TectonicContext{
		IsVolcanic:         true,
		MagmaFlowDirection: geography.Vector{X: 0, Y: 1},
	}
	primary := GenerateMineralVein(ctx, MineralIron, geography.Point{X: 100, Y: 100})
	primary.VeinSize = VeinSizeLarge
	primary.Quantity = 1000

	// Test with high scale (keeps similar size)
	satellite := generateSatelliteVein(primary, ctx, 0.9, 1.0, 10, 50)
	assert.Equal(t, primary.MineralType.Name, satellite.MineralType.Name)
	assert.NotEqual(t, primary.Location, satellite.Location, "Should have different location")

	// Test with medium scale (downgrades once)
	satellite2 := generateSatelliteVein(primary, ctx, 0.5, 0.7, 100, 500)
	assert.NotNil(t, satellite2)

	// Test with low scale (downgrades twice)
	satellite3 := generateSatelliteVein(primary, ctx, 0.2, 0.4, 500, 1000)
	assert.NotNil(t, satellite3)
}
