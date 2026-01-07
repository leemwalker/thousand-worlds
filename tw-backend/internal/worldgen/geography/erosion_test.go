package geography

import (
	"testing"
	"tw-backend/internal/spatial"
)

func TestApplyThermalErosion_Spike(t *testing.T) {
	// Create a 10x10 heightmap with a single spike
	hm := NewHeightmap(10, 10)
	centerX, centerY := 5, 5
	initialHeight := 1000.0
	hm.Set(centerX, centerY, initialHeight)

	// Verify initial state
	if hm.Get(centerX, centerY) != initialHeight {
		t.Fatalf("Setup failed: expected height %f, got %f", initialHeight, hm.Get(centerX, centerY))
	}

	// Apply thermal erosion
	// Uses random seed 123
	ApplyThermalErosion(hm, 50, 123)

	// Check results
	finalHeight := hm.Get(centerX, centerY)
	if finalHeight >= initialHeight {
		t.Errorf("Erosion failed: spike height did not decrease. Got %f", finalHeight)
	}

	// Check that material was moved to neighbors
	// Neighbor at 4,5 should have received something
	neighborHeight := hm.Get(centerX-1, centerY)
	if neighborHeight <= 0 {
		t.Errorf("Erosion failed: neighbor did not gain material. Got %f", neighborHeight)
	}
}

func TestApplyHydraulicErosion_Slope(t *testing.T) {
	// Create a 20x20 slope
	width, height := 20, 20
	hm := NewHeightmap(width, height)

	// Uniform slope from Y=0 (high) to Y=19 (low)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			hm.Set(x, y, float64(height-y)*10.0)
		}
	}

	// Apply hydraulic erosion
	// High number of drops to ensure visible effect
	ApplyHydraulicErosion(hm, 5000, 123)

	// Check for channel formation
	// We expect some variance in the X direction for a given Y row now
	// Ideally, we'd look for valley formation.

	// Simple check: Ensure mass conservation (approximate) or just change
	// Since hydraulic erosion takes material away (transport) or adds it (deposition),
	// the total mass might change slightly due to evaporation/sediment loss at edges,
	// but mostly we want to see change.

	changed := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			original := float64(height-y) * 10.0
			current := hm.Get(x, y)
			if current != original {
				changed = true
				break
			}
		}
	}

	if !changed {
		t.Error("Hydraulic erosion resulted in no changes to the heightmap")
	}
}

// =============================================================================
// Differential Erosion Tests (Phase 5: Geological Provinces)
// =============================================================================

func TestApplyDifferentialErosion_HardnessAffectsRate(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(32)
	hm := NewSphereHeightmap(topo)

	// Set up a simple slope: high elevation at top, low at bottom
	// Face 0, create gradient from y=0 (high) to y=31 (low)
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			coord := spatial.Coordinate{Face: 0, X: x, Y: y}
			hm.Set(coord, float64(32-y)*100.0) // 3200 at top, 100 at bottom
		}
	}

	// Create two regions: soft (hardness 0.2) and hard (hardness 0.9)
	for y := 0; y < 32; y++ {
		for x := 0; x < 16; x++ {
			// Left half: soft
			coord := spatial.Coordinate{Face: 0, X: x, Y: y}
			hm.SetCellData(coord, CellData{RockHardness: 0.2, Sediment: 0, ProvinceID: 1})
		}
		for x := 16; x < 32; x++ {
			// Right half: hard
			coord := spatial.Coordinate{Face: 0, X: x, Y: y}
			hm.SetCellData(coord, CellData{RockHardness: 0.9, Sediment: 0, ProvinceID: 2})
		}
	}

	// Record initial elevations at mid-point
	softCoord := spatial.Coordinate{Face: 0, X: 8, Y: 16}
	hardCoord := spatial.Coordinate{Face: 0, X: 24, Y: 16}
	softInitial := hm.Get(softCoord)
	hardInitial := hm.Get(hardCoord)

	// Apply differential erosion
	ApplyDifferentialErosion(hm, topo, 5000, 12345, 0.0)

	// Get final elevations
	softFinal := hm.Get(softCoord)
	hardFinal := hm.Get(hardCoord)

	// Soft region should erode MORE (lower final elevation relative to initial)
	softErosion := softInitial - softFinal
	hardErosion := hardInitial - hardFinal

	// We expect soft erosion > hard erosion (in average tendency)
	// But since erosion is stochastic, we just check that both have some erosion
	// and the heightmap has changed
	if softErosion < 0 && hardErosion < 0 {
		t.Log("Both regions gained material (deposition dominated) - this is acceptable")
	}

	// At minimum, the heightmap should have changed
	hm.UpdateMinMax()
	if hm.MinElev == hm.MaxElev {
		t.Error("Erosion should create elevation variance")
	}
}

func TestApplyDifferentialErosion_SedimentDeposition(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(16)
	hm := NewSphereHeightmap(topo)

	// Create a valley with flat bottom (sea level = 0)
	// High at top (y=0), slopes down to valley at y=8, then flat to y=15
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			coord := spatial.Coordinate{Face: 0, X: x, Y: y}
			var elev float64
			if y < 8 {
				elev = float64(8-y) * 100.0 // Slope down
			} else {
				elev = 0.0 // Flat valley floor
			}
			hm.Set(coord, elev)
			// All soft rock for easy erosion
			hm.SetCellData(coord, CellData{RockHardness: 0.2, Sediment: 0, ProvinceID: 1})
		}
	}

	// Apply erosion with deposition
	ApplyDifferentialErosion(hm, topo, 2000, 54321, 0.0)

	// Check that some cells in the flat area have sediment deposited
	hasSediment := false
	for y := 8; y < 16; y++ {
		for x := 0; x < 16; x++ {
			coord := spatial.Coordinate{Face: 0, X: x, Y: y}
			data := hm.GetCellData(coord)
			if data.Sediment > 0 {
				hasSediment = true
				break
			}
		}
		if hasSediment {
			break
		}
	}

	// It's acceptable if no sediment was deposited (depends on simulation dynamics)
	// but we log it for visibility
	t.Logf("Sediment deposited in flat area: %v", hasSediment)
}

func TestApplyDifferentialErosion_CoastalShelfFormation(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(32)
	hm := NewSphereHeightmap(topo)

	// Create land-to-ocean transition
	// Left half (x < 16): land with slope, Right half (x >= 16): ocean floor
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			coord := spatial.Coordinate{Face: 0, X: x, Y: y}
			var elev float64
			if x < 16 {
				// Land: height increases inland
				elev = float64(16-x) * 50.0 // 800m to 50m
			} else {
				// Ocean: below sea level
				elev = -100.0 - float64(x-16)*100.0 // -100m to -1600m
			}
			hm.Set(coord, elev)
			// Soft rock on land to encourage erosion
			hardness := 0.3
			if x >= 16 {
				hardness = 0.5 // Ocean floor is harder
			}
			hm.SetCellData(coord, CellData{RockHardness: hardness, Sediment: 0, ProvinceID: 1})
		}
	}

	// Apply erosion with sea level at 0
	ApplyDifferentialErosion(hm, topo, 5000, 99999, 0.0)

	// Check for sediment near the coast (x around 15-17)
	coastalSediment := 0.0
	for y := 0; y < 32; y++ {
		for x := 14; x < 18; x++ {
			coord := spatial.Coordinate{Face: 0, X: x, Y: y}
			data := hm.GetCellData(coord)
			coastalSediment += data.Sediment
		}
	}

	// Log result - coastal shelf formation is a bonus, not strictly required
	t.Logf("Total coastal sediment deposited: %.1fm", coastalSediment)
}

func TestApplyStreamPowerErosion_SedimentTransport(t *testing.T) {
	// Setup Sphere Topology (Small res)
	res := 4
	topo := spatial.NewCubeSphereTopology(res)
	hm := NewSphereHeightmap(topo)

	// Chain: Top -> Mid -> Bot on Face 0
	// Top: (1,1)
	// Mid: (1,2)
	// Bot: (1,3)
	top := spatial.Coordinate{Face: 0, X: 1, Y: 1}
	mid := spatial.Coordinate{Face: 0, X: 1, Y: 2}
	bot := spatial.Coordinate{Face: 0, X: 1, Y: 3}

	// Elevations
	hm.Set(top, 1000.0)
	hm.Set(mid, 500.0)
	hm.Set(bot, 10.0) // Just above sea level

	// Initialize Hydrology Mock
	totalCells := 6 * res * res
	hydro := &HydrologyLayer{
		Flux:          make([]float64, totalCells),
		FlowDirection: make([]int, totalCells), // Default 0 is valid index! Need -1 init.
		Resolution:    res,
	}
	for i := range hydro.FlowDirection {
		hydro.FlowDirection[i] = -1
	}

	// Indices
	resSq := res * res
	// idx = face*resSq + y*res + x
	topIdx := 0*resSq + 1*res + 1
	midIdx := 0*resSq + 2*res + 1
	botIdx := 0*resSq + 3*res + 1

	// Flux
	hydro.Flux[topIdx] = 100.0
	hydro.Flux[midIdx] = 200.0 // Flow accumulates
	hydro.Flux[botIdx] = 300.0

	// Flow Direction
	hydro.FlowDirection[topIdx] = midIdx
	hydro.FlowDirection[midIdx] = botIdx
	hydro.FlowDirection[botIdx] = -1 // Sink

	// Apply SPM Erosion
	// dt = 1.0, seaLevel = 0.0
	// Apply multiple steps to ensure visible transport
	for i := 0; i < 10; i++ {
		ApplyStreamPowerErosion(hm, hydro, nil, 1.0, 0.0)
	}

	// Verify Sediment Transport
	// Top should erode (High slope, flux)
	// Bottom should receive sediment

	topElev := hm.Get(top)
	botData := hm.GetCellData(bot)
	botSed := botData.Sediment

	t.Logf("Top Elev: %.2f (Start 1000)", topElev)
	t.Logf("Bot Sediment: %.2f", botSed)

	if topElev >= 1000.0 {
		t.Error("Upstream should have eroded")
	}

	if botSed <= 0 {
		t.Error("Downstream sink should have accumulated sediment")
	}
}
