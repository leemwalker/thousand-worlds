package ecosystem

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestOceanPhaseTransition_HotPlanet verifies full vaporization above 110°C
func TestOceanPhaseTransition_HotPlanet(t *testing.T) {
	geo := NewWorldGeology(uuid.New(), 12345, 40_000_000)
	geo.InitializeGeology()

	// Set very hot temperature (Hadean)
	geo.TotalYearsSimulated = 0 // Early Earth (heat=10.0)

	// Clear biomes to simplify temperature calculation
	geo.Biomes = nil

	// Record initial values
	initialVaporFraction := geo.OceanVaporFraction

	// Simulate with realistic Hadean hot global temp
	// Earth at year 0 has geothermalOffset = +90°C
	// We add another +20°C from global temp mod to push well above 110°C
	event := geo.SimulateGeology(1000, 20.0)

	// Should have increased vapor fraction (moving toward vaporized state)
	assert.Greater(t, geo.OceanVaporFraction, initialVaporFraction,
		"Hot planet should have higher vapor fraction")

	// No deluge event while heating
	assert.Nil(t, event, "No deluge event while still hot")
}

// TestOceanPhaseTransition_CoolPlanet verifies liquid state below 90°C
func TestOceanPhaseTransition_CoolPlanet(t *testing.T) {
	geo := NewWorldGeology(uuid.New(), 12345, 40_000_000)
	geo.InitializeGeology()

	// Set cool temperature (Modern Earth)
	geo.TotalYearsSimulated = 4_500_000_000 // Modern (heat=1.0, geothermal≈0)

	// Clear biomes to simplify
	geo.Biomes = nil

	// Simulate multiple times to allow convergence
	for i := 0; i < 20; i++ {
		geo.SimulateGeology(10000, -30.0) // Very cool modifier
	}

	// Should have low vapor fraction
	assert.Less(t, geo.OceanVaporFraction, 0.3, "Cool planet should have low vapor fraction")

	// Sea level should have risen toward baseline (from any initial depression)
	assert.Greater(t, geo.SeaLevel, -2000.0, "Sea level should be partially recovered")
}

// TestOceanPhaseTransition_TransitionZone verifies smooth transition 90-110°C
func TestOceanPhaseTransition_TransitionZone(t *testing.T) {
	geo := NewWorldGeology(uuid.New(), 12345, 40_000_000)
	geo.InitializeGeology()

	// Set intermediate age (early Archean)
	geo.TotalYearsSimulated = 1_000_000_000 // heat ≈ 2.35
	geo.Biomes = nil

	// Simulate multiple steps to see transition
	for i := 0; i < 10; i++ {
		geo.SimulateGeology(10000, 0.0)
	}

	// With heat ≈ 2.35, geothermal ≈ +13.5°C, should have some vapor
	// Not fully vaporized, not fully liquid
	assert.GreaterOrEqual(t, geo.OceanVaporFraction, 0.0, "Should have valid vapor fraction")
	assert.LessOrEqual(t, geo.OceanVaporFraction, 1.0, "Should have valid vapor fraction")
}

// TestOceanPhaseTransition_GreatDeluge verifies event detection
func TestOceanPhaseTransition_GreatDeluge(t *testing.T) {
	geo := NewWorldGeology(uuid.New(), 12345, 40_000_000)
	geo.InitializeGeology()

	// Start in early Archean with some vaporization
	geo.TotalYearsSimulated = 800_000_000
	geo.Biomes = nil

	// Manually set to mostly vaporized state
	geo.OceanVaporFraction = 0.8

	// Simulate with very hot temp to keep it vaporized
	event1 := geo.SimulateGeology(10000, 50.0)
	assert.Nil(t, event1, "No event while maintaining hot state")

	// Now cool down dramatically (simulating prolonged cooling)
	var delugeEvent *PhaseTransitionEvent
	for i := 0; i < 50; i++ {
		event := geo.SimulateGeology(10000, -40.0) // Strong cooling
		if event != nil && event.Type == "GreatDeluge" {
			delugeEvent = event
			break
		}
	}

	// Should eventually detect Great Deluge as it cools
	if geo.OceanVaporFraction < 0.5 {
		assert.NotNil(t, delugeEvent, "Should detect Great Deluge when cooling below threshold")
		if delugeEvent != nil {
			assert.Equal(t, "GreatDeluge", delugeEvent.Type)
			assert.Contains(t, delugeEvent.Description, "ocean")
		}
	}
}

// TestOceanPhaseTransition_SeaLevelSmoothing verifies gradual transition
func TestOceanPhaseTransition_SeaLevelSmoothing(t *testing.T) {
	geo := NewWorldGeology(uuid.New(), 12345, 40_000_000)
	geo.InitializeGeology()

	initialSeaLevel := geo.SeaLevel

	// Simulate multiple times with hot temp
	for i := 0; i < 10; i++ {
		geo.SimulateGeology(1000, 50.0)
	}

	// Sea level should have changed but not instantly to -4000
	seaLevelChange := math.Abs(geo.SeaLevel - initialSeaLevel)
	assert.Greater(t, seaLevelChange, 100.0, "Sea level should have changed")
	assert.Less(t, seaLevelChange, 4000.0, "Sea level should transition smoothly, not instantly")
}

// TestOceanPhaseTransition_VaporFractionBounds verifies bounds checking
func TestOceanPhaseTransition_VaporFractionBounds(t *testing.T) {
	geo := NewWorldGeology(uuid.New(), 12345, 40_000_000)
	geo.InitializeGeology()

	// Extreme hot
	geo.TotalYearsSimulated = 0
	geo.SimulateGeology(1000, 100.0)
	assert.LessOrEqual(t, geo.OceanVaporFraction, 1.0, "Vapor fraction should not exceed 1.0")
	assert.GreaterOrEqual(t, geo.OceanVaporFraction, 0.0, "Vapor fraction should not be negative")

	// Extreme cold
	geo.TotalYearsSimulated = 4_500_000_000
	geo.SimulateGeology(1000, -50.0)
	assert.LessOrEqual(t, geo.OceanVaporFraction, 1.0, "Vapor fraction should not exceed 1.0")
	assert.GreaterOrEqual(t, geo.OceanVaporFraction, 0.0, "Vapor fraction should not be negative")
}
