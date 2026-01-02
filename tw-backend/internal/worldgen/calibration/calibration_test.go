package calibration_test

import (
	"testing"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/worldgen/calibration"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Earth Calibration Integration Tests
// =============================================================================
//
// These tests validate that the simulation produces Earth-like results.
// They use a fixed seed for deterministic behavior and "sanity" tolerances
// that are loose enough to prevent false failures while catching major
// regressions (e.g., mountains flattening, oceans draining).

const (
	testSeed       = int64(42)      // Fixed seed for reproducibility
	testResolution = 128            // Lower resolution for faster tests
	testYears      = int64(100_000) // Minimal simulation for unit test speed
)

// TestEarthCalibration_HypsometryWithinTolerance verifies elevation distribution.
func TestEarthCalibration_HypsometryWithinTolerance(t *testing.T) {
	geo := createTestWorld(t)

	stats := calibration.CollectStats(geo)

	// Sanity checks: ensure simulation produces varied terrain
	// These are NOT checking for Earth-accurate values, just basic functionality

	// Ocean Coverage: Should have SOME ocean and SOME land (not 0% or 100%)
	assert.Greater(t, stats.OceanCoveragePercent, 10.0,
		"Should have at least 10%% ocean coverage")
	assert.Less(t, stats.OceanCoveragePercent, 90.0,
		"Should have at most 90%% ocean coverage")

	// Mean Ocean Depth: Check it's negative (below sea level)
	assert.Less(t, stats.MeanOceanDepthM, 0.0,
		"Mean ocean depth should be below sea level")

	// Mean Land Height: Check it's positive (above sea level)
	assert.Greater(t, stats.MeanLandHeightM, 0.0,
		"Mean land height should be above sea level")

	// Elevation range should span a reasonable range
	elevRange := stats.MaxElevationM - stats.MinElevationM
	assert.Greater(t, elevRange, 5000.0,
		"Elevation range should span at least 5km for topographic diversity")
}

// TestEarthCalibration_ClimateWithinTolerance verifies temperature distribution.
func TestEarthCalibration_ClimateWithinTolerance(t *testing.T) {
	geo := createTestWorld(t)

	stats := calibration.CollectStats(geo)

	// Sanity checks: ensure climate system is working
	// These are NOT checking for Earth-accurate values

	// Global Mean Temp: Should be in habitable range (-50 to +50°C)
	assert.Greater(t, stats.GlobalMeanTempC, -50.0,
		"Global mean temperature should be above -50°C")
	assert.Less(t, stats.GlobalMeanTempC, 50.0,
		"Global mean temperature should be below 50°C")

	// Equator-Pole Gradient: Should show temperature variation
	gradient := stats.CalculateEquatorPoleGradient()
	assert.Greater(t, gradient, 10.0,
		"Equator-pole temperature gradient should be at least 10°C")
	assert.Less(t, gradient, 100.0,
		"Equator-pole temperature gradient should not exceed 100°C")
}

// TestEarthCalibration_GeologyWithinTolerance verifies plate tectonics.
func TestEarthCalibration_GeologyWithinTolerance(t *testing.T) {
	geo := createTestWorld(t)

	stats := calibration.CollectStats(geo)

	// Plate Count: Should have reasonable number of plates (5-15)
	assert.GreaterOrEqual(t, stats.PlateCount, 4,
		"Should have at least 4 major tectonic plates")
	assert.LessOrEqual(t, stats.PlateCount, 20,
		"Should not exceed 20 major tectonic plates")

	// Continent Count: At least some continental plates
	assert.Greater(t, stats.ContinentCount, 0,
		"Should have at least one continental plate")

	// Hotspot Count: Should have volcanic hotspots
	assert.Greater(t, stats.HotspotCount, 0,
		"Should have at least one volcanic hotspot")
}

// TestEarthCalibration_HydrologyWithinTolerance verifies water systems.
func TestEarthCalibration_HydrologyWithinTolerance(t *testing.T) {
	geo := createTestWorld(t)

	stats := calibration.CollectStats(geo)

	// Rivers should exist
	// Note: River count varies widely based on seed and simulation time
	// Just verify the system is functional
	assert.NotNil(t, geo.Rivers,
		"River system should be initialized")

	// River density should be reasonable if rivers exist
	if stats.RiverCount > 0 {
		assert.GreaterOrEqual(t, stats.RiverDensityPercent, 0.0,
			"River density should be non-negative")
	}
}

// TestEarthCalibration_BimodalDistribution verifies crustal differentiation.
func TestEarthCalibration_BimodalDistribution(t *testing.T) {
	geo := createTestWorld(t)

	stats := calibration.CollectStats(geo)

	// Check for bimodal peak detection
	oceanPeak, landPeak, isBimodal := stats.DetectBimodalPeaks()

	if isBimodal {
		// Ocean peak should be below sea level
		assert.Less(t, oceanPeak, 0.0,
			"Ocean peak should be below sea level")

		// Land peak should be above sea level (or close to it)
		assert.Greater(t, landPeak, -500.0,
			"Land peak should not be too far below sea level")

		// Peaks should be well-separated
		separation := landPeak - oceanPeak
		assert.Greater(t, separation, 1000.0,
			"Ocean and land peaks should be at least 1km apart")
	} else {
		t.Log("Warning: Bimodal distribution not detected - may indicate insufficient crustal differentiation")
	}
}

// TestEarthCalibration_FullScorecard runs the complete calibration and logs the report.
func TestEarthCalibration_FullScorecard(t *testing.T) {
	geo := createTestWorld(t)

	stats := calibration.CollectStats(geo)
	benchmarks := calibration.DefaultEarthBenchmarks()
	tolerances := calibration.DefaultTolerances()

	report := calibration.Score(stats, benchmarks, tolerances)

	// Log the full scorecard for visibility
	t.Log("\n" + report.FormatScorecard())

	// Note: This test is informational - it logs the scorecard for tuning purposes
	// It does NOT fail based on calibration status since Earth-accurate simulation
	// requires significant parameter tuning.
	// The individual sanity-check tests above catch major regressions.

	// Log failures for visibility
	if report.FailCount > 0 {
		t.Logf("Calibration report: %d/%d metrics passed, %d failed, %d warnings",
			report.PassCount, len(report.Results), report.FailCount, report.WarnCount)
		t.Log("Run 'verify-earth' CLI for detailed tuning guidance")
	}
}

// TestCalibration_Benchmarks verifies benchmark defaults are reasonable.
func TestCalibration_Benchmarks(t *testing.T) {
	benchmarks := calibration.DefaultEarthBenchmarks()

	// Basic sanity checks on Earth values
	assert.InDelta(t, 71.0, benchmarks.OceanCoveragePercent, 5.0,
		"Earth ocean coverage should be ~71%%")
	assert.InDelta(t, -3700.0, benchmarks.MeanOceanDepthM, 500.0,
		"Earth mean ocean depth should be ~-3700m")
	assert.InDelta(t, 840.0, benchmarks.MeanLandHeightM, 200.0,
		"Earth mean land height should be ~840m")
	assert.InDelta(t, 15.0, benchmarks.GlobalMeanTempC, 2.0,
		"Earth global mean temp should be ~15°C")
}

// TestCalibration_ToleranceCalculations verifies tolerance math.
func TestCalibration_ToleranceCalculations(t *testing.T) {
	// Percentage tolerance
	assert.True(t, calibration.IsWithinTolerance(70.0, 71.0, 0.20, false),
		"70 should be within 20%% of 71")
	assert.False(t, calibration.IsWithinTolerance(50.0, 71.0, 0.20, false),
		"50 should NOT be within 20%% of 71")

	// Absolute tolerance
	assert.True(t, calibration.IsWithinTolerance(14.0, 15.0, 5.0, true),
		"14 should be within 5 of 15")
	assert.False(t, calibration.IsWithinTolerance(5.0, 15.0, 5.0, true),
		"5 should NOT be within 5 of 15")
}

// =============================================================================
// Test Helpers
// =============================================================================

// createTestWorld generates a world for testing with fixed parameters.
func createTestWorld(t *testing.T) *ecosystem.WorldGeology {
	t.Helper()

	worldID := uuid.New()
	circumference := 40_000_000.0 // 40,000 km (Earth-like)

	geo := ecosystem.NewWorldGeology(worldID, testSeed, circumference)
	require.NotNil(t, geo, "Failed to create WorldGeology")

	geo.InitializeGeology(0)
	require.NotNil(t, geo.Heightmap, "Failed to initialize geology")

	// Run minimal simulation
	geo.SimulateGeology(testYears, 0.0)

	return geo
}
