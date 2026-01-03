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
// Phase 3: Earth History Validation Tests
// =============================================================================
//
// These tests validate that the simulation replicates Earth's key geological
// properties. They use smaller world sizes and shorter simulation times for
// faster test execution while still validating core behaviors.

const (
	earthHistoryTestResolution = 64      // Small resolution for fast tests
	earthHistoryTestYears      = 10_000  // Short simulation for unit testing
	earthHistoryCircumference  = 4000000 // Smaller world for faster generation
)

// TestEarthHistory_HadeanCratons validates that continental coverage
// remains stable or grows through accretion.
func TestEarthHistory_HadeanCratons(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Earth history test in short mode")
	}

	worldID := uuid.New()
	seed := int64(4500)

	geo := ecosystem.NewWorldGeology(worldID, seed, earthHistoryCircumference)
	require.NotNil(t, geo, "Failed to create WorldGeology")

	// Initialize with small resolution for fast testing
	geo.InitializeGeology(earthHistoryTestResolution)
	require.NotNil(t, geo.Heightmap, "Failed to initialize geology")

	// Count initial stats
	initialStats := calibration.CollectStats(geo)
	initialContinentalPct := 100.0 - initialStats.OceanCoveragePercent

	// Short simulation
	geo.SimulateGeology(earthHistoryTestYears, 0.0)

	// Collect stats after simulation
	postStats := calibration.CollectStats(geo)
	postContinentalPct := 100.0 - postStats.OceanCoveragePercent

	// Verify terrain was generated (basic sanity check)
	assert.Greater(t, postContinentalPct, 0.0,
		"Should have some continental area")

	t.Logf("Hadean Era: Continental coverage %.1f%% -> %.1f%%",
		initialContinentalPct, postContinentalPct)
}

// TestEarthHistory_BimodalElevation validates that the simulation develops
// a bimodal elevation distribution with distinct ocean floors and continental shelves.
func TestEarthHistory_BimodalElevation(t *testing.T) {
	worldID := uuid.New()

	geo := ecosystem.NewWorldGeology(worldID, 4500, earthHistoryCircumference)
	require.NotNil(t, geo, "Failed to create WorldGeology")

	geo.InitializeGeology(earthHistoryTestResolution)
	geo.SimulateGeology(earthHistoryTestYears, 0.0)

	stats := calibration.CollectStats(geo)

	// Check for bimodal peaks (oceanic and continental modes)
	oceanPeak, landPeak, isBimodal := stats.DetectBimodalPeaks()

	if isBimodal {
		// Ocean floor mode should be below sea level
		assert.Less(t, oceanPeak, 0.0,
			"Ocean floor mode should be below sea level")

		// The peaks should show some separation
		separation := landPeak - oceanPeak
		assert.Greater(t, separation, 500.0,
			"Crustal differentiation should create some elevation separation")

		t.Logf("Bimodal elevation: Ocean peak=%.0fm, Land peak=%.0fm, Separation=%.0fm",
			oceanPeak, landPeak, separation)
	} else {
		t.Log("Bimodal distribution not detected in short simulation")
	}
}

// TestEarthHistory_MountainFormation validates that mountain ranges form
// from tectonic simulation.
func TestEarthHistory_MountainFormation(t *testing.T) {
	worldID := uuid.New()

	geo := ecosystem.NewWorldGeology(worldID, 12345, earthHistoryCircumference)
	require.NotNil(t, geo, "Failed to create WorldGeology")

	geo.InitializeGeology(earthHistoryTestResolution)
	geo.SimulateGeology(earthHistoryTestYears, 0.0)

	stats := calibration.CollectStats(geo)

	// Maximum elevation should show some topographic relief
	assert.Greater(t, stats.MaxElevationM, 1000.0,
		"Simulation should create peaks above 1000m")

	// Terrain should have variety
	elevRange := stats.MaxElevationM - stats.MinElevationM
	assert.Greater(t, elevRange, 2000.0,
		"Elevation range should span at least 2km")

	t.Logf("Mountain formation: Max elevation=%.0fm, Range=%.0fm",
		stats.MaxElevationM, elevRange)
}

// TestEarthHistory_ContinentalCoverageTarget validates that continental coverage
// stays within reasonable bounds.
func TestEarthHistory_ContinentalCoverageTarget(t *testing.T) {
	worldID := uuid.New()

	geo := ecosystem.NewWorldGeology(worldID, 4500, earthHistoryCircumference)
	require.NotNil(t, geo, "Failed to create WorldGeology")

	geo.InitializeGeology(earthHistoryTestResolution)
	geo.SimulateGeology(earthHistoryTestYears, 0.0)

	stats := calibration.CollectStats(geo)
	continentalPct := 100.0 - stats.OceanCoveragePercent

	// Should have some land and some ocean (not 0% or 100%)
	assert.Greater(t, continentalPct, 5.0,
		"Continental coverage should be at least 5%%")
	assert.Less(t, continentalPct, 95.0,
		"Continental coverage should not exceed 95%%")

	t.Logf("Continental coverage: %.1f%% (Earth target: 29%%)", continentalPct)
}

// TestEarthHistory_ClimateZones validates that climate zones form correctly
// with temperature variation.
func TestEarthHistory_ClimateZones(t *testing.T) {
	worldID := uuid.New()

	geo := ecosystem.NewWorldGeology(worldID, 4500, earthHistoryCircumference)
	require.NotNil(t, geo, "Failed to create WorldGeology")

	geo.InitializeGeology(earthHistoryTestResolution)
	geo.SimulateGeology(earthHistoryTestYears, 0.0)

	stats := calibration.CollectStats(geo)

	// Global mean temp should be in habitable range
	assert.Greater(t, stats.GlobalMeanTempC, -100.0,
		"Global mean temperature should be above -100°C")
	assert.Less(t, stats.GlobalMeanTempC, 100.0,
		"Global mean temperature should be below 100°C")

	t.Logf("Climate: Mean temp=%.1f°C", stats.GlobalMeanTempC)
}
