package geography

import (
	"testing"

	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Phase 2: Mineral-Aware Sediment Tracking Tests
// =============================================================================

// TestCellData_SedimentMinerals verifies that CellData can track mineral content.
func TestCellData_SedimentMinerals(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	hm := NewSphereHeightmap(topology)

	coord := spatial.Coordinate{Face: 0, X: 5, Y: 5}

	// Set cell data with mineral content
	data := CellData{
		RockHardness:     0.5,
		Sediment:         100.0,
		SedimentMinerals: map[string]float64{"Gold": 0.5, "Iron": 10.0},
	}
	hm.SetCellData(coord, data)

	// Verify retrieval
	retrieved := hm.GetCellData(coord)
	require.NotNil(t, retrieved.SedimentMinerals, "SedimentMinerals should not be nil")
	assert.Equal(t, 0.5, retrieved.SedimentMinerals["Gold"], "Gold content should match")
	assert.Equal(t, 10.0, retrieved.SedimentMinerals["Iron"], "Iron content should match")
}

// TestAddSedimentWithMinerals verifies that adding sediment with minerals works.
func TestAddSedimentWithMinerals(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	hm := NewSphereHeightmap(topology)

	coord := spatial.Coordinate{Face: 0, X: 5, Y: 5}
	hm.Set(coord, 1000.0) // Initial elevation

	// Initialize with some minerals
	data := hm.GetCellData(coord)
	data.SedimentMinerals = map[string]float64{"Gold": 0.1}
	hm.SetCellData(coord, data)

	// Add more sediment with minerals
	hm.AddSedimentWithMinerals(coord, 50.0, map[string]float64{"Gold": 0.2, "Iron": 5.0})

	// Verify combined minerals
	result := hm.GetCellData(coord)
	assert.Equal(t, 50.0, result.Sediment, "Sediment depth should increase")
	assert.InDelta(t, 0.3, result.SedimentMinerals["Gold"], 0.001, "Gold should be summed")
	assert.InDelta(t, 5.0, result.SedimentMinerals["Iron"], 0.001, "Iron should be added")
}

// TestErode_ExtractsMinerals verifies that erosion extracts minerals from sediment.
func TestErode_ExtractsMinerals(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	hm := NewSphereHeightmap(topology)

	coord := spatial.Coordinate{Face: 0, X: 5, Y: 5}
	hm.Set(coord, 1000.0)

	// Set up sediment with minerals
	data := CellData{
		Sediment:         100.0,
		SedimentMinerals: map[string]float64{"Gold": 1.0, "Iron": 10.0},
	}
	hm.SetCellData(coord, data)

	// Erode 50% of the sediment
	_, extractedMinerals := hm.ErodeWithMinerals(coord, 50.0)

	// Should extract proportional minerals
	require.NotNil(t, extractedMinerals, "Should return extracted minerals")
	assert.InDelta(t, 0.5, extractedMinerals["Gold"], 0.01, "Should extract 50% of gold")
	assert.InDelta(t, 5.0, extractedMinerals["Iron"], 0.01, "Should extract 50% of iron")

	// Remaining cell should have reduced minerals
	remaining := hm.GetCellData(coord)
	assert.InDelta(t, 0.5, remaining.SedimentMinerals["Gold"], 0.01, "Remaining gold should be 50%")
	assert.InDelta(t, 5.0, remaining.SedimentMinerals["Iron"], 0.01, "Remaining iron should be 50%")
}

// TestDeltaDeposition_CreatesMineralDeposits verifies that deltas create mineral deposits.
func TestDeltaDeposition_CreatesMineralDeposits(t *testing.T) {
	// This test verifies the integration between erosion mineral tracking
	// and delta deposition. Heavy minerals (gold, platinum) should deposit
	// first at river mouths, creating placer deposits.

	t.Skip("Integration test - requires full delta simulation setup")
}

// TestMineralDensitySorting verifies that heavy minerals deposit before light ones.
func TestMineralDensitySorting(t *testing.T) {
	// Mineral densities (g/cm³) for sorting:
	// Gold: 19.3, Platinum: 21.5, Iron: 7.9, Copper: 8.9, Tin: 7.3
	// Lighter minerals travel further downstream

	densities := GetMineralDensities()

	assert.Greater(t, densities["Gold"], densities["Iron"], "Gold should be denser than Iron")
	assert.Greater(t, densities["Platinum"], densities["Gold"], "Platinum should be denser than Gold")
	assert.Greater(t, densities["Iron"], densities["Tin"], "Iron should be denser than Tin")
}
