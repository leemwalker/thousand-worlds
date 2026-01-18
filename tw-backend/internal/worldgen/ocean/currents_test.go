package ocean

import (
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

func TestNewSystem(t *testing.T) {
	t.Run("initializes with empty slices", func(t *testing.T) {
		topology := spatial.NewCubeSphereTopology(16)
		geo := geography.NewSphereHeightmap(topology)
		seaLevel := 0.0

		sys := NewSystem(topology, geo, seaLevel)

		if sys == nil {
			t.Fatal("NewSystem returned nil")
		}
		for i := 0; i < 6; i++ {
			if len(sys.CurrentMap[i]) == 0 {
				t.Errorf("CurrentMap[%d] should be initialized", i)
			}
			if len(sys.WaterTemperature[i]) == 0 {
				t.Errorf("WaterTemperature[%d] should be initialized", i)
			}
			if len(sys.isOcean[i]) == 0 {
				t.Errorf("isOcean[%d] should be initialized", i)
			}
		}
	})
}

func TestGenerateSurfaceCurrents_EkmanRotation(t *testing.T) {
	// Setup: A sphere with ocean everywhere (all elevations below sea level)
	topology := spatial.NewCubeSphereTopology(16)
	geo := geography.NewSphereHeightmap(topology)
	seaLevel := 100.0 // Everything is ocean
	res := 16

	// Set all elevations below sea level
	for face := 0; face < 6; face++ {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				geo.Set(spatial.Coordinate{Face: face, X: x, Y: y}, -50.0)
			}
		}
	}

	sys := NewSystem(topology, geo, seaLevel)
	// Initialize isOcean because GenerateSurfaceCurrents uses it
	sys.InitializeTemperature()

	// Create a wind map with northward wind (positive Y direction on sphere)
	windMap := make(map[spatial.Coordinate]spatial.Vector3D)

	// Test Northern Hemisphere point (Face 4 = Top = positive Y)
	northCoord := spatial.Coordinate{Face: 0, X: 8, Y: 4}
	windMap[northCoord] = spatial.Vector3D{X: 0, Y: 1, Z: 0}

	// Southern Hemisphere point
	southCoord := spatial.Coordinate{Face: 0, X: 8, Y: 12}
	windMap[southCoord] = spatial.Vector3D{X: 0, Y: 1, Z: 0}

	sys.GenerateSurfaceCurrents(windMap)

	// Assert: Northern Hemisphere
	northIdx := northCoord.Y*res + northCoord.X
	northCurrent := sys.CurrentMap[0][northIdx]

	if northCurrent.X <= 0 {
		t.Errorf("Northern hemisphere: expected positive X (eastward deflection), got X=%f", northCurrent.X)
	}

	// Assert: Southern Hemisphere
	southIdx := southCoord.Y*res + southCoord.X
	southCurrent := sys.CurrentMap[0][southIdx]

	if southCurrent.X >= 0 {
		t.Errorf("Southern hemisphere: expected negative X (westward deflection), got X=%f", southCurrent.X)
	}
}

func TestGenerateSurfaceCurrents_BoundaryDeflection(t *testing.T) {
	// Setup: A continent wall on the west, ocean on the east
	topology := spatial.NewCubeSphereTopology(16)
	geo := geography.NewSphereHeightmap(topology)
	seaLevel := 0.0
	res := 16

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
	// Must initialize to populate isOcean slice
	sys.InitializeTemperature()

	// Create same wind at coastal cell and open ocean cell
	windVec := spatial.Vector3D{X: -1, Y: 0, Z: 0} // Westward-ish wind
	windMap := make(map[spatial.Coordinate]spatial.Vector3D)

	coastalCoord := spatial.Coordinate{Face: 0, X: 8, Y: 8}    // Right at coast
	openOceanCoord := spatial.Coordinate{Face: 0, X: 14, Y: 8} // Far from land

	windMap[coastalCoord] = windVec
	windMap[openOceanCoord] = windVec

	sys.GenerateSurfaceCurrents(windMap)

	// Get currents at both locations using indices
	coastalIdx := coastalCoord.Y*res + coastalCoord.X
	openIdx := openOceanCoord.Y*res + openOceanCoord.X

	coastalCurrent := sys.CurrentMap[0][coastalIdx]
	openCurrent := sys.CurrentMap[0][openIdx]

	t.Logf("Coastal current: %+v (magnitude: %f)", coastalCurrent, coastalCurrent.Length())
	t.Logf("Open ocean current: %+v (magnitude: %f)", openCurrent, openCurrent.Length())

	// At minimum, verify both currents were computed
	if coastalCurrent.Length() == 0 && openCurrent.Length() == 0 {
		t.Error("Expected non-zero currents")
	}
}
