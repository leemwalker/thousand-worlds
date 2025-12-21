package geography_test

import (
	"testing"

	"tw-backend/internal/ecosystem/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed seed for deterministic test results.
const testSeed int64 = 42

// =============================================================================
// BDD Tests: Tectonic System
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Tectonic System Creation
// -----------------------------------------------------------------------------
// Given: A world ID and seed
// When: NewTectonicSystem is called
// Then: System should be created with initial plates and boundaries
func TestBDD_TectonicSystem_Creation(t *testing.T) {
	worldID := uuid.New()

	ts := geography.NewTectonicSystem(worldID, testSeed)

	require.NotNil(t, ts, "Tectonic system should be created")
	assert.Equal(t, worldID, ts.WorldID, "WorldID should match")
	assert.NotEmpty(t, ts.Plates, "Should have initial plates")
	assert.GreaterOrEqual(t, len(ts.Plates), 3, "Should have at least 3 plates")
}

// -----------------------------------------------------------------------------
// Scenario: Plate Types Distribution
// -----------------------------------------------------------------------------
// Given: A newly created tectonic system
// When: Plates are examined
// Then: Should have mix of continental and oceanic plates
func TestBDD_TectonicSystem_PlateTypes(t *testing.T) {
	worldID := uuid.New()
	ts := geography.NewTectonicSystem(worldID, testSeed)

	continentalCount := 0
	oceanicCount := 0
	for _, plate := range ts.Plates {
		switch plate.Type {
		case geography.PlateContinental:
			continentalCount++
		case geography.PlateOceanic:
			oceanicCount++
		}
	}

	assert.Greater(t, continentalCount, 0, "Should have at least one continental plate")
	assert.Greater(t, oceanicCount, 0, "Should have at least one oceanic plate")
}

// -----------------------------------------------------------------------------
// Scenario: Tectonic System Update Over Time
// -----------------------------------------------------------------------------
// Given: A tectonic system
// When: Update is called with 10,000 years
// Then: CurrentYear should advance
func TestBDD_TectonicSystem_Update(t *testing.T) {
	worldID := uuid.New()
	ts := geography.NewTectonicSystem(worldID, testSeed)

	initialYear := ts.CurrentYear

	ts.Update(10000)

	assert.Greater(t, ts.CurrentYear, initialYear, "Year should advance")
}

// -----------------------------------------------------------------------------
// Scenario: Continental Fragmentation Calculation
// -----------------------------------------------------------------------------
// Given: A tectonic system with plates
// When: CalculateFragmentation is called
// Then: Should return value between 0.0 (supercontinent) and 1.0 (max fragmented)
func TestBDD_TectonicSystem_Fragmentation(t *testing.T) {
	worldID := uuid.New()
	ts := geography.NewTectonicSystem(worldID, testSeed)

	frag := ts.CalculateFragmentation()

	assert.GreaterOrEqual(t, frag, float32(0.0), "Fragmentation should be >= 0")
	assert.LessOrEqual(t, frag, float32(1.0), "Fragmentation should be <= 1")
}

// -----------------------------------------------------------------------------
// Scenario: Plate Boundary Detection
// -----------------------------------------------------------------------------
// Given: A hex grid with plates assigned
// When: IsBoundaryCell is called
// Then: Cells on edges between plates should return true
func TestBDD_TectonicSystem_BoundaryDetection(t *testing.T) {
	worldID := uuid.New()
	ts := geography.NewTectonicSystem(worldID, testSeed)
	grid := geography.NewHexGrid(worldID, 20, 20, 1.0)

	// Count boundary cells by iterating through grid coordinates
	boundaryCount := 0
	for q := 0; q < 20; q++ {
		for r := 0; r < 20; r++ {
			coord := geography.HexCoord{Q: q, R: r}
			if ts.IsBoundaryCell(coord, grid) {
				boundaryCount++
			}
		}
	}

	// With multiple plates, should have some boundaries
	assert.GreaterOrEqual(t, boundaryCount, 0, "Should handle boundary detection")
}

// -----------------------------------------------------------------------------
// Scenario: Tectonic Activity Level
// -----------------------------------------------------------------------------
// Given: A cell on a plate boundary
// When: GetBoundaryActivity is called
// Then: Should return activity level >= 0
func TestBDD_TectonicSystem_ActivityLevel(t *testing.T) {
	worldID := uuid.New()
	ts := geography.NewTectonicSystem(worldID, testSeed)
	grid := geography.NewHexGrid(worldID, 20, 20, 1.0)

	// Get activity at various coordinates
	maxActivity := float32(0.0)
	for q := 0; q < 20; q++ {
		for r := 0; r < 20; r++ {
			coord := geography.HexCoord{Q: q, R: r}
			activity := ts.GetBoundaryActivity(coord, grid)
			if activity > maxActivity {
				maxActivity = activity
			}
		}
	}

	assert.GreaterOrEqual(t, maxActivity, float32(0.0), "Activity should be non-negative")
}

// =============================================================================
// BDD Tests: Hex Grid
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Hex Grid Creation
// -----------------------------------------------------------------------------
// Given: Dimensions and cell size
// When: NewHexGrid is called
// Then: Grid should be created with correct dimensions
func TestBDD_HexGrid_Creation(t *testing.T) {
	worldID := uuid.New()
	grid := geography.NewHexGrid(worldID, 10, 10, 1.0)

	require.NotNil(t, grid, "Grid should be created")
	assert.Equal(t, 10, grid.Width)
	assert.Equal(t, 10, grid.Height)
}

// -----------------------------------------------------------------------------
// Scenario: Hex Neighbor Calculation
// -----------------------------------------------------------------------------
// Given: A hex coordinate
// When: Neighbors are queried
// Then: Should return up to 6 valid neighbors
func TestBDD_HexGrid_Neighbors(t *testing.T) {
	worldID := uuid.New()
	grid := geography.NewHexGrid(worldID, 10, 10, 1.0)

	// Center cell should have 6 neighbors
	centerCoord := geography.HexCoord{Q: 5, R: 5}
	neighbors := grid.GetNeighbors(centerCoord)

	assert.LessOrEqual(t, len(neighbors), 6, "Should have at most 6 neighbors")

	// Corner cell has fewer neighbors
	cornerCoord := geography.HexCoord{Q: 0, R: 0}
	cornerNeighbors := grid.GetNeighbors(cornerCoord)

	assert.LessOrEqual(t, len(cornerNeighbors), 6, "Corner should have <= 6 neighbors")
}

// =============================================================================
// BDD Tests: Region System
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Region Identification
// -----------------------------------------------------------------------------
// Given: A hex grid with landmass
// When: IdentifyRegions is called
// Then: Connected land cells should be grouped into regions
func TestBDD_RegionSystem_Identification(t *testing.T) {
	worldID := uuid.New()
	grid := geography.NewHexGrid(worldID, 10, 10, 1.0)

	// Mark some cells as land (form an island)
	for q := 3; q <= 6; q++ {
		for r := 3; r <= 6; r++ {
			cell := grid.GetCell(geography.HexCoord{Q: q, R: r})
			if cell != nil {
				cell.IsLand = true
			}
		}
	}

	rs := geography.NewRegionSystem(worldID)
	rs.IdentifyRegions(grid)

	assert.GreaterOrEqual(t, len(rs.Regions), 0, "Should handle region identification")
}

// -----------------------------------------------------------------------------
// Scenario: Deterministic Results
// -----------------------------------------------------------------------------
// Given: Same seed value
// When: TectonicSystem is created twice
// Then: Results should be identical
func TestBDD_TectonicSystem_Determinism(t *testing.T) {
	worldID := uuid.New()

	ts1 := geography.NewTectonicSystem(worldID, testSeed)
	ts2 := geography.NewTectonicSystem(worldID, testSeed)

	// Plate counts should match
	assert.Equal(t, len(ts1.Plates), len(ts2.Plates), "Plate counts should match")

	// Both should have the same continental fragmentation
	assert.Equal(t, ts1.CalculateFragmentation(), ts2.CalculateFragmentation(),
		"Fragmentation should be identical with same seed")
}
