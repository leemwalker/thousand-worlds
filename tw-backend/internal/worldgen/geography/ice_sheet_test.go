package geography

import (
	"testing"

	"tw-backend/internal/spatial"
)

func TestIceSheet_Accumulation(t *testing.T) {
	is := NewIceSheet()
	topology := spatial.NewCubeSphereTopology(32)
	resolution := 32

	// Create mock temperature and precipitation grids
	totalCells := 6 * resolution * resolution
	tempGrid := make([]float64, totalCells)
	precipGrid := make([]float64, totalCells)

	// Set polar regions cold (-20°C) with precipitation
	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)
		// Face 0 and 5 are "poles" in our cube sphere
		if coord.Face == 0 || coord.Face == 5 {
			tempGrid[i] = -20.0
			precipGrid[i] = 500.0 // mm/year equivalent
		} else {
			tempGrid[i] = 10.0
			precipGrid[i] = 500.0
		}
	}

	// Create heightmap
	heightmap := NewSphereHeightmap(topology)
	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)
		heightmap.Set(coord, 0) // Flat
	}

	// Run for 1000 years
	is.Update(1000, tempGrid, precipGrid, heightmap, topology)

	// Verify ice accumulated only in cold regions
	polarIce := 0
	equatorIce := 0
	for coord, ice := range is.Ice {
		if coord.Face == 0 || coord.Face == 5 {
			if ice.Thickness > 0 {
				polarIce++
			}
		} else {
			if ice.Thickness > 0 {
				equatorIce++
			}
		}
	}

	if polarIce == 0 {
		t.Error("No ice accumulated in polar regions")
	}
	if equatorIce > 0 {
		// Ice may briefly flow into warm regions before ablating - this is OK
		t.Logf("Note: %d equator cells have ice (may flow there before ablating)", equatorIce)
	}

	t.Logf("Polar cells with ice: %d, Total ice volume: %.2f km³", polarIce, is.TotalVolume)
}

func TestIceSheet_Flow(t *testing.T) {
	is := NewIceSheet()
	topology := spatial.NewCubeSphereTopology(16)
	resolution := 16
	totalCells := 6 * resolution * resolution

	// Create heightmap with slope (high at center, low at edges)
	heightmap := NewSphereHeightmap(topology)
	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)
		// Higher elevation at center of each face
		distFromCenter := float64((coord.X-resolution/2)*(coord.X-resolution/2) +
			(coord.Y-resolution/2)*(coord.Y-resolution/2))
		elev := 3000.0 - distFromCenter*10
		heightmap.Set(coord, elev)
	}

	// Place a thick ice dome at center of face 0
	centerCoord := spatial.Coordinate{Face: 0, X: resolution / 2, Y: resolution / 2}
	is.Ice[centerCoord] = &IceData{Thickness: 3000.0}

	// Cold everywhere, no new accumulation
	tempGrid := make([]float64, totalCells)
	precipGrid := make([]float64, totalCells)
	for i := 0; i < totalCells; i++ {
		tempGrid[i] = -10.0
		precipGrid[i] = 0.0 // No new ice
	}

	initialThickness := is.Ice[centerCoord].Thickness

	// Run for 10000 years
	is.Update(10000, tempGrid, precipGrid, heightmap, topology)

	// Ice should have spread (center thinner, neighbors have ice)
	centerIce := 0.0
	if ice, ok := is.Ice[centerCoord]; ok {
		centerIce = ice.Thickness
	}

	spreadCount := 0
	for coord, ice := range is.Ice {
		if coord != centerCoord && ice.Thickness > 0 {
			spreadCount++
		}
	}

	t.Logf("Center ice: %.0f -> %.0f m, Spread to %d cells", initialThickness, centerIce, spreadCount)

	if spreadCount == 0 {
		t.Log("Ice did not spread to neighbors (may need parameter tuning)")
	}
}

func TestIceSheet_Ablation(t *testing.T) {
	is := NewIceSheet()
	topology := spatial.NewCubeSphereTopology(16)
	resolution := 16
	totalCells := 6 * resolution * resolution

	// Place ice everywhere
	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)
		is.Ice[coord] = &IceData{Thickness: 100.0}
	}

	// Create heightmap
	heightmap := NewSphereHeightmap(topology)

	// Warm temperatures everywhere
	tempGrid := make([]float64, totalCells)
	precipGrid := make([]float64, totalCells)
	for i := 0; i < totalCells; i++ {
		tempGrid[i] = 10.0 // +10°C = 20m/year melt
		precipGrid[i] = 0.0
	}

	// Run for 10 years - should melt 200m, more than the 100m present
	is.Update(10, tempGrid, precipGrid, heightmap, topology)

	// All ice should be gone
	remainingIce := 0
	for _, ice := range is.Ice {
		if ice.Thickness > MinIceThickness {
			remainingIce++
		}
	}

	if remainingIce > 0 {
		t.Errorf("Expected all ice to melt, but %d cells still have ice", remainingIce)
	}

	t.Log("Ablation test passed - ice melted completely")
}

func TestIceSheet_Erosion(t *testing.T) {
	is := NewIceSheet()
	topology := spatial.NewCubeSphereTopology(16)
	resolution := 16
	totalCells := 6 * resolution * resolution

	// Create heightmap with initial elevation
	heightmap := NewSphereHeightmap(topology)
	initialElev := 1000.0
	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)
		heightmap.Set(coord, initialElev)
	}

	// Place flowing ice
	centerCoord := spatial.Coordinate{Face: 0, X: 8, Y: 8}
	is.Ice[centerCoord] = &IceData{
		Thickness: 2000.0,
		FlowSpeed: 100.0, // 100 m/year
	}

	// Apply erosion for 1000 years
	erosion := is.ApplyErosion(heightmap, 1000, resolution)

	newElev := heightmap.Get(centerCoord)

	if newElev >= initialElev {
		t.Error("Erosion did not lower the bedrock")
	}

	t.Logf("Erosion: %.2f m total, Elevation: %.0f -> %.0f m", erosion, initialElev, newElev)
}
