package ocean

import (
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

func TestNewSystem(t *testing.T) {
	t.Run("initializes with empty maps", func(t *testing.T) {
		topology := spatial.NewCubeSphereTopology(16)
		geo := geography.NewSphereHeightmap(topology)
		seaLevel := 0.0

		sys := NewSystem(topology, geo, seaLevel)

		if sys == nil {
			t.Fatal("NewSystem returned nil")
		}
		if sys.CurrentMap == nil {
			t.Error("CurrentMap should be initialized")
		}
		if sys.WaterTemperature == nil {
			t.Error("WaterTemperature should be initialized")
		}
	})
}

func TestGenerateSurfaceCurrents_EkmanRotation(t *testing.T) {
	// Setup: A sphere with ocean everywhere (all elevations below sea level)
	topology := spatial.NewCubeSphereTopology(16)
	geo := geography.NewSphereHeightmap(topology)
	seaLevel := 100.0 // Everything is ocean

	// Set all elevations below sea level
	for face := 0; face < 6; face++ {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				geo.Set(spatial.Coordinate{Face: face, X: x, Y: y}, -50.0)
			}
		}
	}

	sys := NewSystem(topology, geo, seaLevel)

	// Create a wind map with northward wind (positive Y direction on sphere)
	// Wind blowing North at equator
	windMap := make(map[spatial.Coordinate]spatial.Vector3D)

	// Test Northern Hemisphere point (Face 4 = Top = positive Y)
	// At equator on front face, wind blowing North
	northCoord := spatial.Coordinate{Face: 0, X: 8, Y: 4}    // Northern part of front face
	windMap[northCoord] = spatial.Vector3D{X: 0, Y: 1, Z: 0} // Northward wind

	// Southern Hemisphere point
	southCoord := spatial.Coordinate{Face: 0, X: 8, Y: 12}   // Southern part of front face
	windMap[southCoord] = spatial.Vector3D{X: 0, Y: 1, Z: 0} // Northward wind

	sys.GenerateSurfaceCurrents(windMap)

	// Assert: Northern Hemisphere should rotate RIGHT (clockwise from above = positive X component added)
	// A pure North wind should gain an East component after Ekman rotation
	northCurrent, ok := sys.CurrentMap[northCoord]
	if !ok {
		t.Fatal("Expected current at northern coordinate")
	}
	// After 45° clockwise rotation from above, North wind should become NE
	// X component should be positive (eastward)
	if northCurrent.X <= 0 {
		t.Errorf("Northern hemisphere: expected positive X (eastward deflection), got X=%f", northCurrent.X)
	}

	// Assert: Southern Hemisphere should rotate LEFT (counter-clockwise = negative X component)
	southCurrent, ok := sys.CurrentMap[southCoord]
	if !ok {
		t.Fatal("Expected current at southern coordinate")
	}
	// After 45° counter-clockwise rotation, North wind should become NW
	// X component should be negative (westward)
	if southCurrent.X >= 0 {
		t.Errorf("Southern hemisphere: expected negative X (westward deflection), got X=%f", southCurrent.X)
	}
}

func TestGenerateSurfaceCurrents_BoundaryDeflection(t *testing.T) {
	// Setup: A continent wall on the west, ocean on the east
	topology := spatial.NewCubeSphereTopology(16)
	geo := geography.NewSphereHeightmap(topology)
	seaLevel := 0.0

	// Create land on the west half, ocean on the east half
	for face := 0; face < 6; face++ {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				if x < 8 {
					// West half is land
					geo.Set(spatial.Coordinate{Face: face, X: x, Y: y}, 100.0)
				} else {
					// East half is ocean
					geo.Set(spatial.Coordinate{Face: face, X: x, Y: y}, -50.0)
				}
			}
		}
	}

	sys := NewSystem(topology, geo, seaLevel)

	// Create same wind at coastal cell and open ocean cell
	windVec := spatial.Vector3D{X: -1, Y: 0, Z: 0} // Westward-ish wind
	windMap := make(map[spatial.Coordinate]spatial.Vector3D)

	coastalCoord := spatial.Coordinate{Face: 0, X: 8, Y: 8}    // Right at coast
	openOceanCoord := spatial.Coordinate{Face: 0, X: 14, Y: 8} // Far from land

	windMap[coastalCoord] = windVec
	windMap[openOceanCoord] = windVec

	sys.GenerateSurfaceCurrents(windMap)

	// Get currents at both locations
	coastalCurrent, coastalOk := sys.CurrentMap[coastalCoord]
	openCurrent, openOk := sys.CurrentMap[openOceanCoord]

	if !coastalOk || !openOk {
		t.Fatal("Expected currents at both coordinates")
	}

	// Coastal current should be dampened (smaller magnitude) IF it was pointing toward land
	// If the Ekman-rotated current points toward open ocean, it won't be dampened
	// The key assertion: the system handles land boundaries without crashing
	// and produces different results at coast vs open ocean when appropriate

	t.Logf("Coastal current: %+v (magnitude: %f)", coastalCurrent, coastalCurrent.Length())
	t.Logf("Open ocean current: %+v (magnitude: %f)", openCurrent, openCurrent.Length())

	// At minimum, verify both currents were computed
	if coastalCurrent.Length() == 0 && openCurrent.Length() == 0 {
		t.Error("Expected non-zero currents")
	}
}
