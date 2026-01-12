package ocean

import (
	"testing"
	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/astronomy"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/weather"

	"github.com/stretchr/testify/assert"
)

func TestCalculateGlobalWindVectors(t *testing.T) {
	// Setup
	topology := spatial.NewCubeSphereTopology(8)
	geo := geography.NewSphereHeightmap(topology)
	seaLevel := 0.0

	// Create some land to verify filtering
	landCoord := spatial.Coordinate{Face: 0, X: 4, Y: 4}
	geo.Set(landCoord, 100.0)

	// Run calculation
	windMap := CalculateGlobalWindVectors(topology, geo, seaLevel)

	// Verify
	if len(windMap) == 0 {
		t.Fatal("Expected wind map to be populated")
	}

	if _, exists := windMap[landCoord]; exists {
		t.Error("Land coordinate should not have wind vector calculated in ocean system")
	}

	oceanCoord := spatial.Coordinate{Face: 0, X: 0, Y: 0}
	if _, exists := windMap[oceanCoord]; !exists {
		t.Error("Ocean coordinate should have wind vector")
	}
}

func TestWindToVector3D_PoleHandling(t *testing.T) {
	topology := spatial.NewCubeSphereTopology(4)
	coord := spatial.Coordinate{Face: 4, X: 2, Y: 2} // Roughly top center

	wind := weather.Wind{
		Speed:     10.0,
		Direction: 0.0, // North
	}

	// This should not panic/NaN
	vec := windToVector3D(topology, coord, wind)

	if vec.Length() == 0 {
		t.Error("Expected non-zero vector at pole")
	}
}

func TestGetAverageOceanTemp(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(4)
	geo := geography.NewSphereHeightmap(topo)
	sys := NewSystem(topo, geo, 0.0)

	// Default is 0, false (no ocean neighbors if empty? or just returns 0?)
	// GetAverageOceanTemp checks neighbors.
	// If all neighbors are empty (not in map), count is 0, returns false.
	coord := spatial.Coordinate{Face: 0, X: 1, Y: 1}
	avg, found := sys.GetAverageOceanTemp(coord)
	assert.False(t, found)
	assert.Equal(t, 0.0, avg)

	// Add neighbor temps
	// Neighbors for (0,1,1) in 4x4 topology: (0,1,0), (0,1,2), (0,2,1), (0,0,1)
	n1 := spatial.Coordinate{Face: 0, X: 1, Y: 0}
	n2 := spatial.Coordinate{Face: 0, X: 1, Y: 2}

	sys.WaterTemperature[n1] = 10.0
	sys.WaterTemperature[n2] = 20.0

	// (10+20) / 2 = 15
	avg, found = sys.GetAverageOceanTemp(coord)
	assert.True(t, found)
	assert.Equal(t, 15.0, avg)
}

func TestCalculateSpringNeapRatio(t *testing.T) {
	// 1 moon -> 1.0
	sats1 := []astronomy.Satellite{{Name: "Moon1"}}
	ratio1 := CalculateSpringNeapRatio(sats1)
	assert.Equal(t, 1.0, ratio1)

	// 2 moons -> 1.5
	sats2 := []astronomy.Satellite{{Name: "Moon1"}, {Name: "Moon2"}}
	ratio2 := CalculateSpringNeapRatio(sats2)
	assert.Equal(t, 1.5, ratio2)
}
