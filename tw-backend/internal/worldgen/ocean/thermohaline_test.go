package ocean

import (
	"testing"
	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

func TestInitializeTemperature(t *testing.T) {
	topology := spatial.NewCubeSphereTopology(16)
	geo := geography.NewSphereHeightmap(topology)
	seaLevel := 0.0
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
	sys.InitializeTemperature()

	// Equator should be warm (~28C)
	eqCoord := spatial.Coordinate{Face: 0, X: 8, Y: 8}
	eqIdx := eqCoord.Y*res + eqCoord.X
	equatorTemp := sys.WaterTemperature[0][eqIdx]
	assert.True(t, equatorTemp > 25.0, "Equator should be warm, got %f", equatorTemp)

	// Poles should be cold (~ -2C to 0C)
	polarCoord := spatial.Coordinate{Face: 4, X: 8, Y: 8} // Top face center (North Pole)
	polarIdx := polarCoord.Y*res + polarCoord.X
	polarTemp := sys.WaterTemperature[4][polarIdx]
	assert.True(t, polarTemp < 5.0, "Pole should be cold, got %f", polarTemp)
}

func TestSimulateThermodynamics_HeatAdvection(t *testing.T) {
	res := 16
	topology := spatial.NewCubeSphereTopology(res)
	geo := geography.NewSphereHeightmap(topology)
	seaLevel := 0.0

	// Set all elevations below sea level
	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				geo.Set(spatial.Coordinate{Face: face, X: x, Y: y}, -50.0)
			}
		}
	}

	sys := NewSystem(topology, geo, seaLevel)
	// Initialize isOcean because advection uses it
	sys.InitializeTemperature()

	// Manually set a hot spot and a current
	sourceIdx := 8*res + 8
	targetIdx := 8*res + 9

	sys.WaterTemperature[0][sourceIdx] = 30.0 // Hot
	sys.WaterTemperature[0][targetIdx] = 10.0 // Cold

	// Current flowing from source to target (eastward)
	sys.CurrentMap[0][sourceIdx] = spatial.Vector3D{X: 1.0, Y: 0, Z: 0}

	initialTargetTemp := sys.WaterTemperature[0][targetIdx]

	// Run multiple iterations to see effect
	sys.SimulateThermodynamics(50)

	newTargetTemp := sys.WaterTemperature[0][targetIdx]

	assert.Greater(t, newTargetTemp, initialTargetTemp, "Target temperature should increase due to advection")
	assert.Less(t, newTargetTemp, 30.0, "Target should not exceed source temperature")
}

func TestGulfStreamEffect(t *testing.T) {
	// 32 resolution to allow for a "current" to travel
	res := 32
	topology := spatial.NewCubeSphereTopology(res)
	geo := geography.NewSphereHeightmap(topology)
	seaLevel := 0.0

	// Entire world is ocean
	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				geo.Set(spatial.Coordinate{Face: face, X: x, Y: y}, -50.0)
			}
		}
	}

	sys := NewSystem(topology, geo, seaLevel)
	sys.InitializeTemperature() // Sets baseline (warm equator, cold poles)

	// Create a strong "northward" current from equator to north pole on Face 0
	for y := 0; y < res; y++ {
		for x := 0; x < res; x++ {
			idx := y*res + x
			sys.CurrentMap[0][idx] = spatial.Vector3D{X: 0, Y: 5.0, Z: 0}
		}
	}

	// Record cold cell temperature at high latitude BEFORE advection
	coldIdx := 2*res + 16 // Northern part of Face 0
	baselineTemp := sys.WaterTemperature[0][coldIdx]

	// Run simulation
	sys.SimulateThermodynamics(100)

	newTemp := sys.WaterTemperature[0][coldIdx]

	t.Logf("Baseline cold cell temp: %f", baselineTemp)
	t.Logf("After advection cold cell temp: %f", newTemp)

	assert.Greater(t, newTemp, baselineTemp, "Northward current should bring warm water to cold regions")
}
