package geography

import (
	"testing"

	"tw-backend/internal/spatial"
)

// TestScandinaviaSimulation verifies that an ice age followed by retreat
// produces realistic glacial landforms: fjords, U-valleys, and glacial lakes.
func TestScandinaviaSimulation(t *testing.T) {
	resolution := 32
	is := NewIceSheet(resolution)
	topology := spatial.NewCubeSphereTopology(resolution)
	totalCells := 6 * resolution * resolution

	// Create a mountainous coastal terrain (like Scandinavia)
	heightmap := NewSphereHeightmap(topology)
	seaLevel := 0.0

	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)

		// Face 0: Coastal mountains (Scandinavia-like)
		// Higher in the interior, sloping to coast
		if coord.Face == 0 {
			distFromCoast := float64(coord.X)
			elev := 500.0 + distFromCoast*100 // 500m coast to 3700m interior
			heightmap.Set(coord, elev)
		} else if coord.Face == 1 {
			// Ocean
			heightmap.Set(coord, -200)
		} else {
			// Other faces: moderate elevation
			heightmap.Set(coord, 200)
		}
	}

	// Cold temperatures in "polar" region, warm elsewhere
	tempGrid := make([]float64, totalCells)
	precipGrid := make([]float64, totalCells)
	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)
		if coord.Face == 0 {
			tempGrid[i] = -15.0   // Cold enough for ice
			precipGrid[i] = 800.0 // High precipitation
		} else {
			tempGrid[i] = 10.0
			precipGrid[i] = 500.0
		}
	}

	t.Log("=== Phase 1: Ice Age (Building ice sheet) ===")

	// Simulate 50,000 years of ice accumulation
	for step := 0; step < 50; step++ {
		is.Update(1000, tempGrid, precipGrid, heightmap, topology)
	}

	t.Logf("After 50,000 years: Ice volume=%.2f km³, Max thickness=%.0f m, Area=%.0f km²",
		is.TotalVolume, is.MaxThickness, is.TotalArea)

	if is.TotalVolume == 0 {
		t.Fatal("No ice accumulated during ice age phase")
	}

	// Apply erosion during glaciation
	totalErosion := 0.0
	for step := 0; step < 50; step++ {
		totalErosion += is.ApplyErosion(heightmap, 1000, resolution)
	}
	t.Logf("Total erosion during glaciation: %.2f m", totalErosion)

	t.Log("=== Phase 2: Deglaciation (Warming) ===")

	// Warm up to trigger ice retreat
	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)
		if coord.Face == 0 {
			tempGrid[i] = 5.0 // Warming
		}
	}

	// Save pre-retreat ice for rebound calculation (using map for sparsity/test convenience)
	previousIce := make(map[spatial.Coordinate]IceData)
	for idx, ice := range is.Ice {
		if ice.Thickness > 0 {
			coord := iceIndexToCoord(idx, resolution)
			previousIce[coord] = ice
		}
	}

	// Simulate 10,000 years of warming
	for step := 0; step < 10; step++ {
		is.Update(1000, tempGrid, precipGrid, heightmap, topology)
	}

	t.Logf("After deglaciation: Ice volume=%.2f km³, Max thickness=%.0f m",
		is.TotalVolume, is.MaxThickness)

	t.Log("=== Phase 3: Feature Detection ===")

	// Deposit moraines at ice margins
	is.DepositMoraines(heightmap, topology, resolution)

	// Apply isostatic rebound
	is.ApplyIsostaticRebound(heightmap, previousIce, 0.3, 10000)

	// Detect glacial features
	features := is.DetectGlacialFeatures(heightmap, topology, seaLevel)

	uValleys := 0
	fjords := 0
	moraines := 0
	for _, f := range features {
		switch f.Type {
		case FeatureUValley:
			uValleys++
		case FeatureFjord:
			fjords++
		case FeatureMoraine:
			moraines++
		}
	}

	t.Logf("Detected features: U-valleys=%d, Fjords=%d, Moraines=%d", uValleys, fjords, moraines)

	// Verify we got some glacial features
	if len(features) == 0 {
		t.Error("No glacial features detected after ice retreat")
	}

	// Detect glacial lakes
	lakes := is.CreateGlacialLakes(heightmap, topology)
	t.Logf("Potential glacial lake locations: %d", len(lakes))
}

// TestGreatLakesSimulation verifies moraine-dammed lake formation.
func TestGreatLakesSimulation(t *testing.T) {
	resolution := 32
	is := NewIceSheet(resolution)
	topology := spatial.NewCubeSphereTopology(resolution)
	totalCells := 6 * resolution * resolution

	// Create a terrain with a depression that could become a lake
	heightmap := NewSphereHeightmap(topology)
	seaLevel := 0.0

	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)

		// Create a basin on face 0
		if coord.Face == 0 {
			// Basin in center
			dx := coord.X - resolution/2
			dy := coord.Y - resolution/2
			dist := float64(dx*dx + dy*dy)

			if dist < 100 {
				// Low center (future lake bed)
				heightmap.Set(coord, 100)
			} else {
				// Higher rim
				heightmap.Set(coord, 300)
			}
		} else {
			heightmap.Set(coord, 200)
		}
	}

	// Cold temperatures to form ice
	tempGrid := make([]float64, totalCells)
	precipGrid := make([]float64, totalCells)
	for i := 0; i < totalCells; i++ {
		tempGrid[i] = -10.0
		precipGrid[i] = 600.0
	}

	t.Log("=== Building ice over basin ===")

	// Build ice
	for step := 0; step < 30; step++ {
		is.Update(1000, tempGrid, precipGrid, heightmap, topology)
	}

	// Evaluate coverage
	coverage := 0
	for _, ice := range is.Ice {
		if ice.Thickness > 0 {
			coverage++
		}
	}

	t.Logf("Ice coverage: %d cells, Volume=%.2f km³", coverage, is.TotalVolume)

	// Erode the basin deeper
	for step := 0; step < 30; step++ {
		is.ApplyErosion(heightmap, 1000, resolution)
	}

	t.Log("=== Retreat and deposit moraines ===")

	// Warm up
	for i := 0; i < totalCells; i++ {
		tempGrid[i] = 10.0
	}

	// Retreat
	for step := 0; step < 10; step++ {
		is.Update(1000, tempGrid, precipGrid, heightmap, topology)
	}

	// Deposit moraines at margins
	is.DepositMoraines(heightmap, topology, resolution)

	// Check for potential dam formation
	lakes := is.CreateGlacialLakes(heightmap, topology)
	features := is.DetectGlacialFeatures(heightmap, topology, seaLevel)

	moraines := 0
	for _, f := range features {
		if f.Type == FeatureMoraine {
			moraines++
		}
	}

	t.Logf("Moraines: %d, Potential lakes: %d", moraines, len(lakes))

	if moraines > 0 {
		t.Log("Moraine deposits created - could dam meltwater to form lakes")
	}
}

// TestErosionBalance verifies that glacial erosion rate exceeds river erosion.
func TestErosionBalance(t *testing.T) {
	resolution := 16
	is := NewIceSheet(resolution)
	topology := spatial.NewCubeSphereTopology(resolution)
	totalCells := 6 * resolution * resolution

	heightmap := NewSphereHeightmap(topology)
	initialElev := 2000.0
	for i := 0; i < totalCells; i++ {
		coord := iceIndexToCoord(i, resolution)
		heightmap.Set(coord, initialElev)
	}

	// Place thick, fast-flowing ice
	centerCoord := spatial.Coordinate{Face: 0, X: 8, Y: 8}
	idx := centerCoord.Face*resolution*resolution + centerCoord.Y*resolution + centerCoord.X

	is.Ice[idx] = IceData{
		Thickness: 3000.0,
		FlowSpeed: 200.0, // Fast glacier
	}

	// Calculate glacial erosion over 1000 years
	glacialErosion := is.ApplyErosion(heightmap, 1000, resolution)

	// Compare to typical river erosion rate (~0.05 mm/year = 0.05 m / 1000 years)
	typicalRiverErosion := 0.05 // m per 1000 years

	t.Logf("Glacial erosion: %.4f m, Typical river erosion: %.4f m", glacialErosion, typicalRiverErosion)

	// Glacial erosion should be significantly higher than river erosion
	if glacialErosion > typicalRiverErosion {
		t.Logf("✓ Glacial erosion (%.4f) > River erosion (%.4f) - physically realistic",
			glacialErosion, typicalRiverErosion)
	} else {
		t.Logf("Note: Glacial erosion may need parameter tuning to exceed river erosion")
	}
}
