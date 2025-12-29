package geography

import (
	"testing"
	"tw-backend/internal/spatial"
)

func TestFillDepressions_Bowl(t *testing.T) {
	// Create a bowl shape: High edges, low center
	topo := spatial.NewCubeSphereTopology(10)
	hm := NewSphereHeightmap(topo)

	center := spatial.Coordinate{Face: 0, X: 5, Y: 5}
	rimHeight := 100.0
	centerHeight := 10.0

	// Initialize entire face to rim height first to avoid unforeseen outlets
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			hm.Set(spatial.Coordinate{Face: 0, X: x, Y: y}, rimHeight)
		}
	}

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			c := spatial.Coordinate{Face: 0, X: x, Y: y}
			// Distance from center
			dx := float64(x - 5)
			dy := float64(y - 5)
			dist := dx*dx + dy*dy
			elev := centerHeight + dist*5.0 // Steep slope to ensure linear cut is a valley
			if elev > rimHeight {
				elev = rimHeight
			}
			hm.Set(c, elev)
		}
	}

	// Create a "Valley" path to the outlet at (0, 5)
	// Ensuring the path is monotonic or at least < 50
	for x := 0; x <= 5; x++ {
		// Cut a channel
		c := spatial.Coordinate{Face: 0, X: x, Y: 5}
		// Linear slope from Outlet (50) to Center (10)
		elev := 50.0 - float64(5-x)*8.0 // Approx
		if x == 5 {
			elev = 10.0
		}
		if x == 0 {
			elev = 50.0
		} // Outlet

		// Wait, slope should be increasing from center(10) to outlet(50)
		elev = 10.0 + float64(5-x)*((50.0-10.0)/5.0)

		hm.Set(c, elev)
	}

	// Ensure Outlet neighbor (outside) is low
	// Face 0 (0,5) -> West is Face 4
	outside := topo.GetNeighbor(spatial.Coordinate{Face: 0, X: 0, Y: 5}, spatial.West)
	hm.Set(outside, 0.0)

	// Run depression filling
	newLakes := FillDepressions(hm, 0.0)

	// Verify center is a lake
	centerData := hm.GetCellData(center)
	if !centerData.IsLake {
		t.Error("Center of bowl should be a lake")
	}

	// Verify water level matches outlet (50.0)
	// The filling updates the elevation to the water surface
	waterLevel := hm.Get(center)
	if waterLevel != 50.0 {
		t.Errorf("Lake level incorrect. Got %f, want 50.0 (outlet height)", waterLevel)
	}

	// Lake ID should be set
	if centerData.LakeID == 0 {
		t.Error("LakeID should be set")
	}

	if len(newLakes) == 0 {
		t.Error("Should return created lakes list")
	}
}

func TestIdentifySinks(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(5)
	hm := NewSphereHeightmap(topo)

	// Initialize all to high elevation
	for f := 0; f < 6; f++ {
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				hm.Set(spatial.Coordinate{Face: f, X: x, Y: y}, 100.0)
			}
		}
	}

	// Single sink at 2,2
	hm.Set(spatial.Coordinate{Face: 0, X: 2, Y: 2}, 10.0)

	// Surroundings higher (already set to 100)

	sinks := IdentifySinks(hm, 0.0)
	if len(sinks) != 1 {
		t.Errorf("Found %d sinks, want 1", len(sinks))
	} else {
		if sinks[0] != (spatial.Coordinate{Face: 0, X: 2, Y: 2}) {
			t.Errorf("Wrong sink found: %v", sinks[0])
		}
	}
}
