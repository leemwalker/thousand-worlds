package ecosystem

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVariableTimeStepEquivalence(t *testing.T) {
	worldID := uuid.New()
	seed := int64(999)
	circumference := 1_000_000.0 // 1,000 km (Small world for fast test)

	// Helper to get total elevation
	getTotalElevation := func(g *WorldGeology) float64 {
		sum := 0.0
		for _, e := range g.Heightmap.Elevations {
			sum += e
		}
		return sum
	}

	// Case A: 100,000 years, step = 1 year
	geoA := NewWorldGeology(worldID, seed, circumference)
	geoA.InitializeGeology()
	initialElev := getTotalElevation(geoA)

	// Run 1000 steps of 1 year
	for i := 0; i < 1000; i++ {
		geoA.SimulateGeology(1, 0.0)
	}
	elevA := getTotalElevation(geoA)

	// Case B: 1000 years, step = 1000 years (single step)
	geoB := NewWorldGeology(worldID, seed, circumference)
	geoB.InitializeGeology()
	// Run 1 step of 1000 years
	geoB.SimulateGeology(1000, 0.0)
	elevB := getTotalElevation(geoB)

	// Comparison
	// They should be very close. Due to floating point accumulation order, might be slightly different.
	// Also Erosion uses iterations.
	// Logic:
	// A: 100k calls.
	//    Tectonics: Every 100k calls, threshold reached -> 1 tectonic update.
	//    Erosion: Every 10k calls, threshold reached -> 1 erosion iteration.
	//             Total 10 erosion iterations.
	// B: 1 call (100k).
	//    Tectonics: dt=100k. intervals = 1. -> 1 tectonic update.
	//    Erosion: dt=100k. intervals = 10. -> 10 erosion iterations.
	// They should be functionally identical.

	// Check that we actually did something
	assert.NotEqual(t, initialElev, elevA, "Simulation A should modify world")
	assert.NotEqual(t, initialElev, elevB, "Simulation B should modify world")

	// Compare global metrics
	// Allow small tolerance
	tolerance := math.Abs(initialElev) * 0.01 // 1% tolerance
	diff := math.Abs(elevA - elevB)

	assert.Less(t, diff, tolerance, "Variable time steps should produce equivalent results. Diff: %v", diff)

	// Check Sea Level
	assert.InDelta(t, geoA.SeaLevel, geoB.SeaLevel, 1.0, "Sea Levels should be consistent")
}
