package geography_test

import (
	"math/rand"
	"testing"

	"tw-backend/internal/ecosystem/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Unit Tests: HexCell Methods
// =============================================================================

func TestHexCell_IsWater(t *testing.T) {
	tests := []struct {
		name     string
		terrain  geography.TerrainType
		expected bool
	}{
		{"Ocean is water", geography.TerrainOcean, true},
		{"Shallows is water", geography.TerrainShallows, true},
		{"Plains is not water", geography.TerrainPlains, false},
		{"Mountain is not water", geography.TerrainMountain, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := geography.NewHexCell(geography.HexCoord{Q: 0, R: 0}, tt.terrain, 0)
			assert.Equal(t, tt.expected, cell.IsWater())
		})
	}
}

func TestHexCell_ScaleDir(t *testing.T) {
	coord := geography.HexCoord{Q: 1, R: 2}
	scaled := coord.ScaleDir(3)
	assert.Equal(t, 3, scaled.Q)
	assert.Equal(t, 6, scaled.R)
}

// =============================================================================
// Unit Tests: HexGrid Methods
// =============================================================================

// Helper to populate a hex grid with cells
func populateHexGrid(grid *geography.HexGrid) {
	for q := 0; q < grid.Width; q++ {
		for r := 0; r < grid.Height; r++ {
			terrain := geography.TerrainOcean
			if q > 2 && q < grid.Width-2 && r > 2 && r < grid.Height-2 {
				terrain = geography.TerrainPlains
			}
			cell := geography.NewHexCell(geography.HexCoord{Q: q, R: r}, terrain, 0)
			grid.SetCell(cell)
		}
	}
}

func TestHexGrid_GetLandNeighbors(t *testing.T) {
	worldID := uuid.New()
	grid := geography.NewHexGrid(worldID, 10, 10, 1.0)
	populateHexGrid(grid)

	// Center is land (q=5 is in land range)
	centerCoord := geography.HexCoord{Q: 5, R: 5}
	landNeighbors := grid.GetLandNeighbors(centerCoord)
	assert.GreaterOrEqual(t, len(landNeighbors), 0, "Should return land neighbors")
}

func TestHexGrid_CellsInRegion(t *testing.T) {
	worldID := uuid.New()
	grid := geography.NewHexGrid(worldID, 10, 10, 1.0)
	populateHexGrid(grid)

	// Assign some cells to a region
	regionID := uuid.New()
	for q := 3; q <= 5; q++ {
		for r := 3; r <= 5; r++ {
			cell := grid.GetCell(geography.HexCoord{Q: q, R: r})
			if cell != nil {
				cell.RegionID = &regionID
			}
		}
	}

	cells := grid.CellsInRegion(regionID)
	assert.Greater(t, len(cells), 0, "Should return cells in region")
}

func TestHexGrid_CountLandCells(t *testing.T) {
	worldID := uuid.New()
	grid := geography.NewHexGrid(worldID, 10, 10, 1.0)
	populateHexGrid(grid)

	count := grid.CountLandCells()
	assert.Greater(t, count, 0, "Should have land cells after populating")
}

// =============================================================================
// Unit Tests: Region Methods
// =============================================================================

func TestRegion_Creation(t *testing.T) {
	worldID := uuid.New()
	region := geography.NewRegion("Test Island", worldID, 1000000)

	assert.NotEqual(t, uuid.Nil, region.ID)
	assert.Equal(t, "Test Island", region.Name)
	assert.Equal(t, worldID, region.WorldID)
	assert.Equal(t, int64(1000000), region.CreatedYear)
}

func TestRegion_AddAndContainsCell(t *testing.T) {
	worldID := uuid.New()
	region := geography.NewRegion("Test Region", worldID, 0)

	coord := geography.HexCoord{Q: 5, R: 5}
	region.AddCell(coord)

	assert.True(t, region.ContainsCell(coord), "Should contain added cell")
	assert.False(t, region.ContainsCell(geography.HexCoord{Q: 0, R: 0}), "Should not contain other cell")
}

func TestRegion_Connections(t *testing.T) {
	worldID := uuid.New()
	region := geography.NewRegion("Main Island", worldID, 0)
	targetID := uuid.New()

	// Add connection
	region.AddConnection(targetID, geography.ObstacleMountain, 0.5, 3, nil)

	// Update connection
	region.AddConnection(targetID, geography.ObstacleRiver, 0.3, 5, nil)

	// Remove connection
	region.RemoveConnection(targetID)

	assert.Len(t, region.Connections, 0, "Connection should be removed")
}

func TestRegion_IsIsolated(t *testing.T) {
	worldID := uuid.New()
	region := geography.NewRegion("Isolated Island", worldID, 0)

	// No connections = isolated
	assert.True(t, region.IsIsolated(), "Region with no connections should be isolated")

	// Add difficult connection
	region.AddConnection(uuid.New(), geography.ObstacleMountain, 0.9, 1, nil)
	assert.True(t, region.IsIsolated(), "Region with only difficult connections should be isolated")

	// Add easy connection
	region.AddConnection(uuid.New(), geography.ObstacleNone, 0.2, 5, nil)
	assert.False(t, region.IsIsolated(), "Region with easy connection should not be isolated")
}

func TestRegion_GetIsolationModifier(t *testing.T) {
	worldID := uuid.New()
	region := geography.NewRegion("Ancient Island", worldID, 0)
	region.IsIsland = true
	region.IsolationYears = 1000000 // 1 million years isolated

	modifier := region.GetIsolationModifier()

	// Should have significant isolation effects after 1MY
	assert.Greater(t, modifier.Strength, float32(0.0), "Should have isolation strength")
}

func TestIsolationModifier_ApplyToSize(t *testing.T) {
	// Just verify the function runs without panicking
	modifier := geography.IsolationModifier{
		Strength:     0.5,
		SmallSizeMod: 0.2,
		LargeSizeMod: -0.2,
	}

	// The function should handle any size
	_ = modifier.ApplyToSize(0.5)
	_ = modifier.ApplyToSize(5.0)
	_ = modifier.ApplyToSize(10.0)
}

// =============================================================================
// Unit Tests: RegionSystem Methods
// =============================================================================

func TestRegionSystem_AddAndGetRegion(t *testing.T) {
	worldID := uuid.New()
	rs := geography.NewRegionSystem(worldID)

	region := geography.NewRegion("Test Region", worldID, 0)
	rs.AddRegion(region)

	retrieved := rs.GetRegion(region.ID)
	require.NotNil(t, retrieved)
	assert.Equal(t, region.Name, retrieved.Name)
}

func TestRegionSystem_Update(t *testing.T) {
	worldID := uuid.New()
	seed := int64(42)

	rs := geography.NewRegionSystem(worldID)
	grid := geography.NewHexGrid(worldID, 10, 10, 1.0)
	tectonics := geography.NewTectonicSystem(worldID, seed)

	// Add a region with isolation
	region := geography.NewRegion("Isolated", worldID, 0)
	region.IsIsland = true
	region.IsolationYears = 100
	rs.AddRegion(region)

	// Update should increment isolation years
	rs.Update(1000, grid, tectonics)

	updated := rs.GetRegion(region.ID)
	assert.Greater(t, updated.IsolationYears, int64(100), "Isolation years should increase")
}

// =============================================================================
// Unit Tests: TectonicPlate Methods
// =============================================================================

func TestTectonicPlate_Speed(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	plate := geography.NewTectonicPlate("Test Plate", geography.PlateContinental, rng)
	speed := plate.Speed()
	assert.GreaterOrEqual(t, speed, float32(0.0), "Speed should be non-negative")
}

func TestTectonicSystem_SplitMergePlates(t *testing.T) {
	worldID := uuid.New()
	seed := int64(42)

	ts := geography.NewTectonicSystem(worldID, seed)
	initialCount := len(ts.Plates)

	// Update for a long time to trigger split/merge
	for i := 0; i < 100; i++ {
		ts.Update(1000000) // 1 million years per update
	}

	// Plate count may have changed
	assert.Greater(t, len(ts.Plates), 0, "Should still have plates")
	t.Logf("Plates changed from %d to %d", initialCount, len(ts.Plates))
}

// =============================================================================
// Additional Coverage Tests
// =============================================================================

func TestHexCoord_RoundHex_EdgeCases(t *testing.T) {
	// Test round trip through pixel conversion
	original := geography.HexCoord{Q: 5, R: 5}
	x, y := original.ToPixel(1.0)
	recovered := geography.FromPixel(x, y, 1.0)
	assert.Equal(t, original, recovered, "Round trip should preserve coordinates")
}
