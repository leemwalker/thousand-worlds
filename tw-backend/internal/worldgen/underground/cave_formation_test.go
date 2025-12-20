package underground

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimulateCaveFormation_NoLimestone(t *testing.T) {
	grid := NewColumnGrid(5, 5)

	// Add non-limestone strata
	for _, col := range grid.AllColumns() {
		col.AddStratum("granite", 100, 0, 8, 1000, 0.05)
	}

	// No rainfall data - will use default
	caves := SimulateCaveFormation(grid, nil, 1000000, 42, DefaultCaveConfig())

	// No caves should form in granite
	assert.Empty(t, caves)
}

func TestSimulateCaveFormation_WithLimestone(t *testing.T) {
	grid := NewColumnGrid(5, 5)

	// Add thick limestone strata with high porosity
	for _, col := range grid.AllColumns() {
		col.AddStratum("limestone", 100, -100, 4, 100000, 0.4) // 200m thick, porous
	}

	// High rainfall
	rainfall := make([]float64, 25)
	for i := range rainfall {
		rainfall[i] = 0.8
	}

	// Long time period
	config := DefaultCaveConfig()
	config.DissolutionRate = 0.01 // High rate for testing
	caves := SimulateCaveFormation(grid, rainfall, 10000000, 42, config)

	// Should form caves
	assert.NotEmpty(t, caves, "Caves should form in porous limestone over millions of years")
}

func TestCreateCaveInStratum(t *testing.T) {
	col := &WorldColumn{X: 50, Y: 50}
	stratum := &StrataLayer{
		TopZ:     0,
		BottomZ:  -100,
		Material: "limestone",
	}

	rng := rand.New(rand.NewSource(42))
	cave := createCaveInStratum(col, stratum, rng, 1000000)

	assert.NotNil(t, cave)
	assert.Equal(t, "karst", cave.CaveType)
	assert.NotEmpty(t, cave.Nodes)
	assert.Equal(t, int64(1000000), cave.FormationAge)
}

func TestRegisterCaveInColumn(t *testing.T) {
	col := &WorldColumn{X: 10, Y: 10}
	cave := NewCave("karst", 0)

	// Add node centered on column
	cave.AddNode(Vector3{X: 10, Y: 10, Z: -50}, 5, 10)

	RegisterCaveInColumn(col, cave)

	assert.Equal(t, 1, len(col.Voids))
	assert.Equal(t, cave.ID, col.Voids[0].VoidID)
	assert.Equal(t, "karst", col.Voids[0].VoidType)
}

func TestRegisterCaveInColumn_OutOfRange(t *testing.T) {
	col := &WorldColumn{X: 10, Y: 10}
	cave := NewCave("karst", 0)

	// Add node far from column
	cave.AddNode(Vector3{X: 100, Y: 100, Z: -50}, 5, 10)

	RegisterCaveInColumn(col, cave)

	// Should not register - cave is too far
	assert.Empty(t, col.Voids)
}

func TestConnectAdjacentCaves(t *testing.T) {
	caves := []*Cave{}

	// Create two nearby caves
	cave1 := NewCave("karst", 0)
	cave1.AddNode(Vector3{X: 0, Y: 0, Z: -50}, 5, 10)

	cave2 := NewCave("karst", 0)
	cave2.AddNode(Vector3{X: 20, Y: 0, Z: -50}, 5, 10) // 20m away

	caves = append(caves, cave1, cave2)

	// Connect caves within 30m
	ConnectAdjacentCaves(caves, 30.0)

	// Cave1 should now have nodes from cave2
	assert.GreaterOrEqual(t, len(cave1.Nodes), 2, "Caves should be merged")
	assert.NotEmpty(t, cave1.Passages, "Should have passage connecting caves")
}

func TestConnectAdjacentCaves_TooFar(t *testing.T) {
	caves := []*Cave{}

	cave1 := NewCave("karst", 0)
	cave1.AddNode(Vector3{X: 0, Y: 0, Z: -50}, 5, 10)

	cave2 := NewCave("karst", 0)
	cave2.AddNode(Vector3{X: 100, Y: 0, Z: -50}, 5, 10) // 100m away

	caves = append(caves, cave1, cave2)

	ConnectAdjacentCaves(caves, 30.0)

	// Should NOT be merged
	assert.Equal(t, 1, len(cave1.Nodes), "Caves too far apart should not merge")
	assert.Empty(t, cave1.Passages)
}

func TestDistance3D(t *testing.T) {
	a := Vector3{X: 0, Y: 0, Z: 0}
	b := Vector3{X: 3, Y: 4, Z: 0}

	dist := distance3D(a, b)
	assert.InDelta(t, 5.0, dist, 0.001)
}

func TestDefaultCaveConfig(t *testing.T) {
	config := DefaultCaveConfig()

	assert.Greater(t, config.DissolutionRate, 0.0)
	assert.Greater(t, config.MinLimestoneDepth, 0.0)
	assert.Equal(t, 1.0, config.WaterFlowFactor)
	assert.Equal(t, 1.0, config.CO2Factor)
}
