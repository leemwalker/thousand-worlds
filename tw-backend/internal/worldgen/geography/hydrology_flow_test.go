package geography

import (
	"testing"

	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test: HydrologyLayer Flow Field
// =============================================================================

func TestCalculateFlowField_SteepestDescent(t *testing.T) {
	// Create a simple slope: high at Y=0, low at Y=9
	topo := spatial.NewCubeSphereTopology(10)
	hm := NewSphereHeightmap(topo)

	// Set elevation: slope down from Y=0 to Y=9
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			hm.Set(spatial.Coordinate{Face: 0, X: x, Y: y}, float64(10-y)*100.0)
		}
	}

	hydro := CalculateFlowField(hm, nil)
	require.NotNil(t, hydro)

	// Check that flow directions point downhill (toward increasing Y)
	// Cell at (5, 5) should flow to (5, 6) which is South in our coordinate system
	idx := coordToIndex(spatial.Coordinate{Face: 0, X: 5, Y: 5}, 10)
	downhillIdx := hydro.FlowDirection[idx]

	// Verify it's not a sink
	assert.NotEqual(t, -1, downhillIdx, "Mid-slope cell should have a flow direction")

	// Verify downhill neighbor has lower elevation
	downhillCoord := indexToCoord(downhillIdx, 10)
	assert.Less(t, hm.Get(downhillCoord), hm.Get(spatial.Coordinate{Face: 0, X: 5, Y: 5}),
		"Flow should go to lower elevation")
}

func TestCalculateFlowField_FluxAccumulation(t *testing.T) {
	// Create slope with V-shaped valley
	topo := spatial.NewCubeSphereTopology(10)
	hm := NewSphereHeightmap(topo)

	// V-valley sloping down Y with center at X=5
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			// Y slope + X V-shape
			elev := float64(10-y)*100.0 + float64(absInt(x-5))*50.0
			hm.Set(spatial.Coordinate{Face: 0, X: x, Y: y}, elev)
		}
	}

	hydro := CalculateFlowField(hm, nil)

	// Center of valley at bottom should have high flux
	centerIdx := coordToIndex(spatial.Coordinate{Face: 0, X: 5, Y: 9}, 10)
	sideIdx := coordToIndex(spatial.Coordinate{Face: 0, X: 0, Y: 9}, 10)

	assert.Greater(t, hydro.Flux[centerIdx], hydro.Flux[sideIdx],
		"Valley center should accumulate more flux than sides")

	// Center flux should be significantly higher (collecting from both valley walls)
	assert.Greater(t, hydro.Flux[centerIdx], 5.0,
		"Valley bottom should collect water from multiple cells")
}

func TestCalculateFlowField_IdentifiesSinks(t *testing.T) {
	// Create a depression in the middle
	topo := spatial.NewCubeSphereTopology(10)
	hm := NewSphereHeightmap(topo)

	// Set uniform elevation
	for face := 0; face < 6; face++ {
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				hm.Set(spatial.Coordinate{Face: face, X: x, Y: y}, 100.0)
			}
		}
	}

	// Create a pit at (5, 5) on face 0
	hm.Set(spatial.Coordinate{Face: 0, X: 5, Y: 5}, 50.0)

	hydro := CalculateFlowField(hm, nil)

	// The pit should be a sink (FlowDirection == -1)
	pitIdx := coordToIndex(spatial.Coordinate{Face: 0, X: 5, Y: 5}, 10)
	assert.Equal(t, -1, hydro.FlowDirection[pitIdx],
		"Depression should be identified as a sink")
}

// =============================================================================
// Test: River Continuity (Peak to Ocean)
// =============================================================================

func TestRiverContinuity_PeakToOcean(t *testing.T) {
	// Create a mountain with ocean surrounding
	topo := spatial.NewCubeSphereTopology(16)
	hm := NewSphereHeightmap(topo)
	seaLevel := 0.0

	// Set ocean everywhere
	for face := 0; face < 6; face++ {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				hm.Set(spatial.Coordinate{Face: face, X: x, Y: y}, -1000.0)
			}
		}
	}

	// Create a mountain peak on face 0, center
	peak := spatial.Coordinate{Face: 0, X: 8, Y: 8}
	hm.Set(peak, 5000.0)

	// Create radial slope from peak
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if x == 8 && y == 8 {
				continue // Skip peak
			}
			dist := absInt(x-8) + absInt(y-8)
			elev := 5000.0 - float64(dist)*600.0
			if elev > hm.Get(spatial.Coordinate{Face: 0, X: x, Y: y}) {
				hm.Set(spatial.Coordinate{Face: 0, X: x, Y: y}, elev)
			}
		}
	}

	hydro := CalculateFlowField(hm, nil)

	// Trace from peak to see if we reach ocean or a sink
	currentIdx := coordToIndex(peak, 16)
	maxSteps := 100
	reachedOcean := false

	for step := 0; step < maxSteps; step++ {
		nextIdx := hydro.FlowDirection[currentIdx]
		if nextIdx == -1 {
			// Reached a sink - check if it's ocean
			currentCoord := indexToCoord(currentIdx, 16)
			if hm.Get(currentCoord) <= seaLevel {
				reachedOcean = true
			}
			break
		}
		currentIdx = nextIdx
		// Check if we're in ocean
		currentCoord := indexToCoord(currentIdx, 16)
		if hm.Get(currentCoord) <= seaLevel {
			reachedOcean = true
			break
		}
	}

	assert.True(t, reachedOcean, "Water from peak should reach ocean level")
}

// =============================================================================
// Test: Stream Power Erosion with Isostatic Rebound
// =============================================================================

func TestErosionRebound_LessThanEroded(t *testing.T) {
	// Eroding 10m of rock should result in <10m elevation drop due to buoyancy
	topo := spatial.NewCubeSphereTopology(10)
	hm := NewSphereHeightmap(topo)

	// Set initial conditions: continental plateau at 35km thickness
	initialThickness := 35.0 // km
	initialElev := CalculateIsostaticHeight(initialThickness, DensityGranite)

	coord := spatial.Coordinate{Face: 0, X: 5, Y: 5}
	hm.Set(coord, initialElev)

	// Store initial state
	data := hm.GetCellData(coord)
	data.RockHardness = 0.5
	data.Flux = 100.0 // High flux for erosion
	hm.SetCellData(coord, data)

	// Calculate expected: erode 10m (0.01km) of rock
	erodedKm := 0.01
	newThickness := initialThickness - erodedKm
	newElev := CalculateIsostaticHeight(newThickness, DensityGranite)

	// The elevation drop should be less than 10m due to isostatic rebound
	elevDrop := initialElev - newElev

	assert.Less(t, elevDrop, 10.0,
		"Isostatic rebound should reduce elevation drop below eroded amount")
	assert.Greater(t, elevDrop, 0.0,
		"Erosion should still cause some elevation drop")

	// Expected: ~1.82m drop for 10m erosion (1 - 2700/3300 = 0.182)
	expectedDrop := 10.0 * (1.0 - DensityGranite/DensityMantle)
	assert.InDelta(t, expectedDrop, elevDrop, 0.5,
		"Elevation drop should match isostatic physics")
}

// Helper functions for coordinate <-> index conversion
func coordToIndex(coord spatial.Coordinate, res int) int {
	return coord.Face*res*res + coord.Y*res + coord.X
}

func indexToCoord(idx, res int) spatial.Coordinate {
	resSq := res * res
	face := idx / resSq
	rem := idx % resSq
	y := rem / res
	x := rem % res
	return spatial.Coordinate{Face: face, X: x, Y: y}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
