package underground

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimulateMagmaChambers_Cooling(t *testing.T) {
	grid := NewColumnGrid(10, 10)

	// Set up column with strata
	col := grid.Get(5, 5)
	col.Surface = 100
	col.AddStratum("granite", 100, -5000, 8, 1000, 0.05)

	chamber := &MagmaChamber{
		Center:      Vector3{X: 5, Y: 5, Z: -2000},
		Volume:      1000000,
		Temperature: 1100, // Just above solidification
		Pressure:    30,
	}

	chambers := []*MagmaChamber{chamber}
	config := DefaultMagmaConfig()
	config.CoolingRatePerYear = 10 // Fast cooling for test

	// Simulate 100 years
	_, newTubes, _ := SimulateMagmaChambers(grid, chambers, nil, 100, 42, config)

	// Chamber should have cooled and solidified
	assert.Less(t, chamber.Temperature, 1100.0, "Chamber should cool")

	// Since rock is hard (granite=8), should form cave
	assert.NotEmpty(t, newTubes, "Should form cave in hard rock")
}

func TestSimulateMagmaChambers_Eruption(t *testing.T) {
	grid := NewColumnGrid(10, 10)

	col := grid.Get(5, 5)
	col.Surface = 100
	col.AddStratum("basalt", 100, -5000, 6, 1000, 0.1)

	chamber := &MagmaChamber{
		Center:      Vector3{X: 5, Y: 5, Z: -2000},
		Volume:      1000000,
		Temperature: 1500,
		Pressure:    85, // Above eruption threshold
	}

	chambers := []*MagmaChamber{chamber}
	config := DefaultMagmaConfig()

	erupted, _, _ := SimulateMagmaChambers(grid, chambers, nil, 1000, 42, config)

	assert.NotEmpty(t, erupted, "High pressure should cause eruption")
	assert.Less(t, chamber.Pressure, 85.0, "Pressure should decrease after eruption")
}

func TestCreateLavaTube_WithRng(t *testing.T) {
	grid := NewColumnGrid(10, 10)

	col := grid.Get(5, 5)
	col.Surface = 100
	col.AddStratum("basalt", 100, -5000, 6, 1000, 0.1)

	chamber := &MagmaChamber{
		Center: Vector3{X: 5, Y: 5, Z: -2000},
		Age:    100000,
	}

	rng := rand.New(rand.NewSource(42))
	tube := createLavaTube(grid, chamber, rng)

	assert.NotNil(t, tube)
	assert.Equal(t, "lava_tube", tube.CaveType)
	assert.NotEmpty(t, tube.Nodes)
	assert.NotEmpty(t, tube.Passages)
}

func TestGetTectonicBoundaries(t *testing.T) {
	// Two plates with opposing movement
	centroids := []Vector3{
		{X: 2, Y: 5, Z: 0},
		{X: 8, Y: 5, Z: 0},
	}
	movements := []Vector3{
		{X: 0.5, Y: 0, Z: 0},  // Moving right
		{X: -0.5, Y: 0, Z: 0}, // Moving left
	}

	boundaries := GetTectonicBoundaries(10, 10, centroids, movements)

	assert.NotEmpty(t, boundaries, "Should detect boundaries between plates")

	// Check that we have convergent boundaries (plates moving toward each other)
	hasConvergent := false
	for _, b := range boundaries {
		if b.BoundaryType == "convergent" {
			hasConvergent = true
			break
		}
	}
	assert.True(t, hasConvergent, "Should have convergent boundaries")
}

func TestProcessSolidifiedChamber_HardRock(t *testing.T) {
	grid := NewColumnGrid(10, 10)

	col := grid.Get(5, 5)
	col.Surface = 100
	col.AddStratum("granite", 100, -5000, 9, 1000, 0.02) // Very hard

	chamber := &MagmaChamber{
		Center:      Vector3{X: 5, Y: 5, Z: -2000},
		Temperature: 500, // Solidified
		Age:         1000000,
	}

	rng := rand.New(rand.NewSource(42))
	config := DefaultMagmaConfig()

	result := processSolidifiedChamber(grid, chamber, rng, config)

	assert.NotNil(t, result)
	assert.Equal(t, "magma_chamber", result.CaveType, "Hard rock preserves chamber as cave")
}

func TestProcessSolidifiedChamber_SoftRock(t *testing.T) {
	grid := NewColumnGrid(10, 10)

	col := grid.Get(5, 5)
	col.Surface = 100
	col.AddStratum("soil", 100, -5000, 2, 1000, 0.5) // Very soft

	chamber := &MagmaChamber{
		Center:      Vector3{X: 5, Y: 5, Z: -2000},
		Temperature: 500, // Solidified
		Age:         1000000,
	}

	rng := rand.New(rand.NewSource(42))
	config := DefaultMagmaConfig()

	result := processSolidifiedChamber(grid, chamber, rng, config)

	assert.NotNil(t, result)
	assert.Equal(t, "collapsed", result.CaveType, "Soft rock causes chamber to collapse")
}

func TestDefaultMagmaConfig(t *testing.T) {
	config := DefaultMagmaConfig()

	assert.Greater(t, config.CoolingRatePerYear, 0.0)
	assert.Greater(t, config.EruptionThreshold, 0.0)
	assert.Greater(t, config.MagmaChamberRadius, 0.0)
	assert.Greater(t, config.LavaTubeFormationProb, 0.0)
	assert.LessOrEqual(t, config.LavaTubeFormationProb, 1.0)
}
