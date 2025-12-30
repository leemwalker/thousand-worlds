package geography

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Test: Isostatic Height Calculator
// =============================================================================
// Verifies Archimedes' Principle: crust floats on mantle based on density.
// Formula:
//   Displacement = thickness × (density / mantleDensity)
//   Freeboard = thickness - Displacement
//   Elevation = Freeboard × 1000 + SeaLevelOffset
//
// With SeaLevelOffset = -5400m, calibrated for Earth-like elevations.

func TestCalculateIsostaticHeight(t *testing.T) {
	scenarios := []struct {
		name        string
		thicknessKm float64
		density     float64
		minHeight   float64 // Minimum expected (meters)
		maxHeight   float64 // Maximum expected (meters)
	}{
		{
			name:        "Standard Oceanic Crust (7km basalt)",
			thicknessKm: 7.0,
			density:     DensityBasalt,
			minHeight:   -4900.0, // Abyssal plain
			maxHeight:   -4600.0,
		},
		{
			name:        "Thin Oceanic Crust (6km basalt)",
			thicknessKm: 6.0,
			density:     DensityBasalt,
			minHeight:   -4900.0, // Deeper
			maxHeight:   -4100.0,
		},
		{
			name:        "Standard Continental Crust (35km granite)",
			thicknessKm: 35.0,
			density:     DensityGranite,
			minHeight:   400.0, // Low-lying plains
			maxHeight:   1200.0,
		},
		{
			name:        "Thin Continental Crust (30km granite)",
			thicknessKm: 30.0,
			density:     DensityGranite,
			minHeight:   0.0, // Near sea level
			maxHeight:   800.0,
		},
		{
			name:        "Thick Mountain Root (60km granite - Himalayas)",
			thicknessKm: 60.0,
			density:     DensityGranite,
			minHeight:   4500.0, // Major mountain range
			maxHeight:   6500.0,
		},
		{
			name:        "Andes-style Coastal Mountains (45km granite)",
			thicknessKm: 45.0,
			density:     DensityGranite,
			minHeight:   2000.0, // High plateau
			maxHeight:   3500.0,
		},
		{
			name:        "Island Arc (12km basalt)",
			thicknessKm: 12.0,
			density:     DensityBasalt,
			minHeight:   -4500.0, // Still below sea level but thickened
			maxHeight:   -4100.0,
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			height := CalculateIsostaticHeight(sc.thicknessKm, sc.density)

			assert.GreaterOrEqual(t, height, sc.minHeight,
				"Height should be >= %v for %s", sc.minHeight, sc.name)
			assert.LessOrEqual(t, height, sc.maxHeight,
				"Height should be <= %v for %s", sc.maxHeight, sc.name)
		})
	}
}

// TestIsostaticHeight_PhysicsConsistency verifies that thicker crust = higher elevation
func TestIsostaticHeight_PhysicsConsistency(t *testing.T) {
	// Thicker crust should always float higher
	thin := CalculateIsostaticHeight(30.0, DensityGranite)
	thick := CalculateIsostaticHeight(40.0, DensityGranite)
	assert.Greater(t, thick, thin, "Thicker crust should float higher")

	// Lighter density should float higher for same thickness
	heavy := CalculateIsostaticHeight(35.0, DensityBasalt)
	light := CalculateIsostaticHeight(35.0, DensityGranite)
	assert.Greater(t, light, heavy, "Lighter density should float higher")

	// Oceanic should be below sea level, continental above
	oceanic := CalculateIsostaticHeight(7.0, DensityBasalt)
	continental := CalculateIsostaticHeight(35.0, DensityGranite)
	assert.Less(t, oceanic, 0.0, "7km oceanic crust should be below sea level")
	assert.Greater(t, continental, 0.0, "35km continental crust should be above sea level")
}
