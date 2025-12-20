package underground

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrataLayer_Thickness(t *testing.T) {
	stratum := StrataLayer{TopZ: 100, BottomZ: 50}
	assert.Equal(t, 50.0, stratum.Thickness())
}

func TestStrataLayer_ContainsDepth(t *testing.T) {
	stratum := StrataLayer{TopZ: 100, BottomZ: 50}

	tests := []struct {
		depth    float64
		expected bool
	}{
		{100, true},  // At top
		{75, true},   // Middle
		{50, true},   // At bottom
		{101, false}, // Above top
		{49, false},  // Below bottom
	}

	for _, tt := range tests {
		result := stratum.ContainsDepth(tt.depth)
		assert.Equal(t, tt.expected, result, "depth %f", tt.depth)
	}
}

func TestVoidSpace_Height(t *testing.T) {
	void := VoidSpace{MinZ: 10, MaxZ: 30}
	assert.Equal(t, 20.0, void.Height())
}

func TestOrganicSource_Age(t *testing.T) {
	source := OrganicSource{DeathYear: 1000, BurialYear: 1100}

	assert.Equal(t, int64(9000), source.Age(10000))
	assert.Equal(t, int64(8900), source.BurialDuration(10000))
}

func TestMagmaInfo_IsSolidified(t *testing.T) {
	tests := []struct {
		temp     float64
		expected bool
	}{
		{1500, false}, // Still molten
		{1000, false}, // At threshold
		{999, true},   // Just below
		{500, true},   // Well cooled
	}

	for _, tt := range tests {
		m := MagmaInfo{Temperature: tt.temp}
		assert.Equal(t, tt.expected, m.IsSolidified(), "temp %f", tt.temp)
	}
}

func TestNewColumnGrid(t *testing.T) {
	grid := NewColumnGrid(10, 20)

	assert.Equal(t, 10, grid.Width)
	assert.Equal(t, 20, grid.Height)

	// Check all columns exist
	for y := 0; y < 20; y++ {
		for x := 0; x < 10; x++ {
			col := grid.Get(x, y)
			assert.NotNil(t, col)
			assert.Equal(t, x, col.X)
			assert.Equal(t, y, col.Y)
			assert.Equal(t, float64(-10000), col.Bedrock)
		}
	}
}

func TestColumnGrid_Get_OutOfBounds(t *testing.T) {
	grid := NewColumnGrid(5, 5)

	assert.Nil(t, grid.Get(-1, 0))
	assert.Nil(t, grid.Get(0, -1))
	assert.Nil(t, grid.Get(5, 0))
	assert.Nil(t, grid.Get(0, 5))
}

func TestColumnGrid_InitFromSurface(t *testing.T) {
	grid := NewColumnGrid(3, 3)

	elevations := []float64{
		10, 20, 30,
		40, 50, 60,
		70, 80, 90,
	}

	grid.InitFromSurface(elevations)

	assert.Equal(t, 10.0, grid.Get(0, 0).Surface)
	assert.Equal(t, 50.0, grid.Get(1, 1).Surface)
	assert.Equal(t, 90.0, grid.Get(2, 2).Surface)
}

func TestColumnGrid_AllColumns(t *testing.T) {
	grid := NewColumnGrid(3, 3)
	all := grid.AllColumns()

	assert.Equal(t, 9, len(all))
}

func TestColumnGrid_GetStratumAt(t *testing.T) {
	grid := NewColumnGrid(2, 2)

	col := grid.Get(0, 0)
	col.AddStratum("soil", 100, 90, 2, 1000, 0.3)
	col.AddStratum("limestone", 90, 50, 5, 100000, 0.2)
	col.AddStratum("granite", 50, -100, 9, 1000000, 0.05)

	// Find limestone
	stratum := grid.GetStratumAt(0, 0, 70)
	assert.NotNil(t, stratum)
	assert.Equal(t, "limestone", stratum.Material)

	// Find granite
	stratum = grid.GetStratumAt(0, 0, 0)
	assert.NotNil(t, stratum)
	assert.Equal(t, "granite", stratum.Material)

	// Out of range
	stratum = grid.GetStratumAt(0, 0, -200)
	assert.Nil(t, stratum)
}

func TestColumnGrid_GetResourcesAt(t *testing.T) {
	grid := NewColumnGrid(2, 2)

	col := grid.Get(1, 1)
	col.AddResource("iron", -50, 100)
	col.AddResource("gold", -200, 10)
	col.AddResource("coal", -100, 500)

	// Get resources in range -150 to -50
	resources := grid.GetResourcesAt(1, 1, -150, -50)
	assert.Equal(t, 2, len(resources)) // iron and coal

	// Get all resources
	resources = grid.GetResourcesAt(1, 1, -500, 0)
	assert.Equal(t, 3, len(resources))
}

func TestWorldColumn_AddStratum(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0}

	col.AddStratum("limestone", 50, 0, 5, 1000, 0.2)

	assert.Equal(t, 1, len(col.Strata))
	assert.Equal(t, "limestone", col.Strata[0].Material)
	assert.Equal(t, 5.0, col.Strata[0].Hardness)
	assert.Equal(t, 0.2, col.Strata[0].Porosity)
}
