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
