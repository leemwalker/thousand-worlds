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
	res := 4

	// Set everything to land FIRST
	for face := 0; face < 6; face++ {
		for i := 0; i < res*res; i++ {
			geo.Set(spatial.Coordinate{Face: face, X: i % res, Y: i / res}, 100.0)
		}
	}

	sys := NewSystem(topo, geo, 0.0)

	coord := spatial.Coordinate{Face: 0, X: 1, Y: 1}
	avg, found := sys.GetAverageOceanTemp(coord)
	assert.False(t, found) // No ocean neighbors yet

	// Add neighbor temps and make them ocean
	n1 := spatial.Coordinate{Face: 0, X: 1, Y: 0}
	n2 := spatial.Coordinate{Face: 0, X: 1, Y: 2}

	geo.Set(n1, -10.0) // Ocean
	geo.Set(n2, -10.0) // Ocean

	// Initialize to populate isOcean bits
	sys.InitializeTemperature()

	sys.WaterTemperature[0][n1.Y*res+n1.X] = 10.0
	sys.WaterTemperature[0][n2.Y*res+n2.X] = 20.0

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
