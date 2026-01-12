package geography

import (
	"testing"
	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
)

// Hydrology Tests
func TestHydrology_LayerMethods(t *testing.T) {
	// Setup
	topo := spatial.NewCubeSphereTopology(4)
	hm := NewSphereHeightmap(topo)

	// Create HydrologyLayer via CalculateFlowField
	hydro := CalculateFlowField(hm, nil)
	assert.NotNil(t, hydro)

	// Test IndexToCoord / CoordToIndex
	idx := 5
	coord := hydro.IndexToCoord(idx)
	newIdx := hydro.CoordToIndex(coord)
	assert.Equal(t, idx, newIdx)

	// Test GetFlux (should be default 1.0 or similar)
	flux := hydro.GetFlux(coord)
	assert.GreaterOrEqual(t, flux, 0.0)

	// Test IsSink
	// We didn't set elevations, so defaults might create sinks or flat areas
	_ = hydro.IsSink(coord)
}

func TestCalculateGlobalFlux(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(4)
	hm := NewSphereHeightmap(topo)

	// Create a bowl shape
	// Center is lowest
	center := spatial.Coordinate{Face: 0, X: 2, Y: 2}
	hm.Set(center, -100.0)

	// Everything else is 0.0 (default)
	// Actually defaults to heightmap min/max? No, slice defaults to 0.

	// Runs in place
	CalculateGlobalFlux(hm)

	// Check results in CellData
	centerData := hm.GetCellData(center)

	// Center should have accumulated flux from at least one neighbor
	// Since it is the lowest point, all neighbors (elev 0) should flow into it (elev -100)
	// Flux at center should be 1.0 (its own) + 4.0 (neighbors) = 5.0
	assert.Greater(t, centerData.Flux, 1.0, "Sink should accumulate flux from neighbors")
}

func TestHydrology_FlowField_IsSink(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(4)
	hm := NewSphereHeightmap(topo)
	coord := spatial.Coordinate{Face: 0, X: 1, Y: 1}

	// Sink Setup: Surrounded by higher elevation
	hm.Set(coord, 0.0)
	hm.Set(spatial.Coordinate{Face: 0, X: 0, Y: 1}, 10.0)
	hm.Set(spatial.Coordinate{Face: 0, X: 2, Y: 1}, 10.0)
	hm.Set(spatial.Coordinate{Face: 0, X: 1, Y: 0}, 10.0)
	hm.Set(spatial.Coordinate{Face: 0, X: 1, Y: 2}, 10.0)

	hydro := CalculateFlowField(hm, nil)

	isSink := hydro.IsSink(coord)
	assert.True(t, isSink)

	// Open one side
	hm.Set(spatial.Coordinate{Face: 0, X: 0, Y: 1}, -10.0) // Lower neighbor
	hydro = CalculateFlowField(hm, nil)                    // Re-calc
	isSink = hydro.IsSink(coord)
	assert.False(t, isSink)
}

// POI Tests
func TestGeneratePOIs_Peak(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(4)
	hm := NewSphereHeightmap(topo)

	// Add a peak
	peak := spatial.Coordinate{Face: 0, X: 2, Y: 2}
	hm.Set(peak, 5000.0) // High peak
	hm.MinElev = -1000.0 // Set range
	hm.MaxElev = 6000.0

	pois := GeneratePOIs(hm, 0.0, 10)
	assert.NotEmpty(t, pois)

	foundPeak := false
	for _, p := range pois {
		if p.Type == POITypeMountainPeak {
			foundPeak = true
			assert.Equal(t, 5000.0, p.Elevation)
		}
	}
	assert.True(t, foundPeak, "Should find high peak")

	// Test FindHighestPeak helper
	highest := FindHighestPeak(pois)
	assert.NotNil(t, highest)
	assert.Equal(t, 5000.0, highest.Elevation)
}

func TestGeneratePOIs_DeepOcean(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(4)
	hm := NewSphereHeightmap(topo)

	// Add a deep trench
	trench := spatial.Coordinate{Face: 0, X: 2, Y: 2}
	hm.Set(trench, -8000.0)
	hm.MinElev = -10000.0
	hm.MaxElev = 1000.0

	pois := GeneratePOIs(hm, 0.0, 10)

	foundTrench := false
	for _, p := range pois {
		if p.Type == POITypeDeepOcean {
			foundTrench = true
		}
	}
	assert.True(t, foundTrench, "Should find deep ocean trench")
}
