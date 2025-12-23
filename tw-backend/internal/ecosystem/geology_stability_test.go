package ecosystem

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBillionYearStability verifies that geological simulation produces
// realistic values after a billion years of simulated time.
// This tests the equilibrium-based tectonics and clamping mechanisms.
func TestBillionYearStability(t *testing.T) {
	// Create a test world with standard parameters
	worldID := uuid.New()
	seed := int64(42)
	circumference := 40_000_000.0 // Earth-like circumference in meters

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology()

	require.NotNil(t, geology.Heightmap, "Heightmap should be initialized")
	require.NotNil(t, geology.SphereHeightmap, "SphereHeightmap should be initialized")

	// Simulate geological events for 1 billion years
	// Using 1M year steps for test performance (simulation logic handles dt)
	totalYears := int64(1_000_000_000)
	stepSize := int64(1_000_000)

	geoManager := NewGeologicalEventManager()

	for year := int64(0); year < totalYears; year += stepSize {
		// Check for geological events
		tick := year * 365
		geoManager.CheckForNewEvents(tick, stepSize*365)
		geoManager.UpdateActiveEvents(tick)

		// Get climate modifiers
		tempMod, _, _ := geoManager.GetEnvironmentModifiers()

		// Apply events to geology
		for _, event := range geoManager.ActiveEvents {
			eventAge := tick - event.StartTick
			isNewEvent := eventAge < stepSize*365
			if isNewEvent {
				geology.ApplyEvent(event)
			}
		}

		// Run geological simulation
		geology.SimulateGeology(stepSize, tempMod)
	}

	// Get final statistics
	stats := geology.GetStats()

	// Assert realistic values after 1 billion years
	t.Logf("=== Billion Year Stability Test Results ===")
	t.Logf("Years Simulated: %d", stats.YearsSimulated)
	t.Logf("Avg Elevation: %.1fm", stats.AverageElevation)
	t.Logf("Max Elevation: %.1fm", stats.MaxElevation)
	t.Logf("Min Elevation: %.1fm", stats.MinElevation)
	t.Logf("Sea Level: %.1fm", stats.SeaLevel)
	t.Logf("Land Coverage: %.1f%%", stats.LandPercent)
	t.Logf("Avg Temperature: %.1f°C", stats.AverageTemperature)
	t.Logf("Plates: %d", stats.PlateCount)

	// Maximum elevation should be capped at ~15km
	assert.LessOrEqual(t, stats.MaxElevation, 20000.0,
		"Max elevation should not exceed 20km (with some tolerance)")

	// Minimum elevation should be capped at ~-11km
	assert.GreaterOrEqual(t, stats.MinElevation, -15000.0,
		"Min elevation should not go below -15km (with some tolerance)")

	// Average elevation should be reasonable
	assert.Greater(t, stats.AverageElevation, -10000.0,
		"Average elevation should be above -10km")
	assert.Less(t, stats.AverageElevation, 10000.0,
		"Average elevation should be below 10km")

	// Sea level should have recovered toward 0
	assert.Greater(t, stats.SeaLevel, -1000.0,
		"Sea level should not drop below -1km")
	assert.Less(t, stats.SeaLevel, 500.0,
		"Sea level should not rise above 500m")

	// Land coverage should be reasonable (not 94%+)
	assert.Less(t, stats.LandPercent, 80.0,
		"Land coverage should be less than 80%")
	assert.Greater(t, stats.LandPercent, 5.0,
		"Land coverage should be greater than 5%")

	// Temperature should be realistic
	assert.Greater(t, stats.AverageTemperature, -60.0,
		"Average temperature should be above -60°C")
	assert.Less(t, stats.AverageTemperature, 60.0,
		"Average temperature should be below 60°C")
}

// TestEquilibriumTectonicsConvergence verifies that running tectonics
// multiple times converges rather than accumulates.
func TestEquilibriumTectonicsConvergence(t *testing.T) {
	worldID := uuid.New()
	seed := int64(123)
	circumference := 40_000_000.0

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology()

	initialMaxElev := geology.SphereHeightmap.MaxElev

	// Run 100 continental drift events
	for i := 0; i < 100; i++ {
		geology.applyContinentalDrift(0.8) // High severity
	}

	finalMaxElev := geology.SphereHeightmap.MaxElev

	t.Logf("Initial Max Elevation: %.1fm", initialMaxElev)
	t.Logf("Final Max Elevation after 100 drift events: %.1fm", finalMaxElev)

	// With equilibrium model, elevation should NOT grow unboundedly
	// Max should be capped regardless of how many events occur
	assert.LessOrEqual(t, finalMaxElev, 20000.0,
		"Max elevation should remain capped after 100 drift events")
}

// TestSeaLevelHomeostasis verifies sea level recovery mechanism.
func TestSeaLevelHomeostasis(t *testing.T) {
	worldID := uuid.New()
	seed := int64(456)
	circumference := 40_000_000.0

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology()

	// Artificially drop sea level (simulate many ice ages)
	geology.SeaLevel = -5000.0

	// Run simulation for extended period
	// 100 steps of 100k years = 10M years
	for i := 0; i < 100; i++ {
		geology.SimulateGeology(100000, 0.0) // 100k years per step
	}

	t.Logf("Initial Sea Level: -5000m")
	t.Logf("Final Sea Level after 10M years: %.1fm", geology.SeaLevel)

	// Sea level should have recovered significantly toward 0
	assert.Greater(t, geology.SeaLevel, -1000.0,
		"Sea level should recover toward 0 over geological time")
}
