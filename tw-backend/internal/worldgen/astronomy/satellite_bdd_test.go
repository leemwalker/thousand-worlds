package astronomy_test

import (
	"math"
	"testing"

	"tw-backend/internal/worldgen/astronomy"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// BDD Tests: Satellite Physics for FPV Mode
// =============================================================================
// These tests document user-visible behavior for moon physics in First-Person View.
// When a watcher enters FPV on a moon, they experience that moon's unique gravity.

// -----------------------------------------------------------------------------
// Scenario: Surface Gravity Calculation (FPV Movement Physics)
// -----------------------------------------------------------------------------
// Given: The watcher enters FPV mode on a moon
// When: They move and jump
// Then: The gravity should be calculated from the moon's mass and radius
//   AND: Smaller moons should have weaker gravity (float longer)
//   AND: Larger moons should have stronger gravity (faster falls)
//   AND: Earth's Moon should be ~1.62 m/s² (for reference)

func TestBDD_Satellite_SurfaceGravityForFPV(t *testing.T) {
	scenarios := []struct {
		name        string
		massRatio   float64 // relative to Moon
		radiusRatio float64 // relative to Moon
		gravityMin  float64 // m/s² lower bound
		gravityMax  float64 // m/s² upper bound
		playerFeel  string  // User experience description
	}{
		{
			name:        "Earth's Moon (reference)",
			massRatio:   1.0,
			radiusRatio: 1.0,
			gravityMin:  1.5,
			gravityMax:  1.7,
			playerFeel:  "Slow, floaty jumps - about 1/6 Earth gravity",
		},
		{
			name:        "Small asteroid moon (Phobos-like)",
			massRatio:   1.0659e16 / astronomy.MoonMassKg,    // Actual Phobos mass
			radiusRatio: 11.1e3 / astronomy.MoonRadiusMeters, // Actual Phobos radius
			gravityMin:  0.003,
			gravityMax:  0.01,
			playerFeel:  "Nearly weightless, can leap enormous distances",
		},
		{
			name:        "Large moon (Ganymede-like)",
			massRatio:   2.0,
			radiusRatio: 1.5,
			gravityMin:  1.3,
			gravityMax:  1.6,
			playerFeel:  "Similar to our Moon but slightly lower due to larger radius",
		},
		{
			name:        "Dense small moon (high iron content)",
			massRatio:   0.5,
			radiusRatio: 0.5,
			gravityMin:  2.5,
			gravityMax:  3.5,
			playerFeel:  "Heavier than expected for size - noticeable pull",
		},
		{
			name:        "Gas giant moon (Titan-like)",
			massRatio:   1.8,
			radiusRatio: 1.48,
			gravityMin:  1.2,
			gravityMax:  1.5,
			playerFeel:  "Slightly lighter than Moon due to lower density",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			// Given: A moon with specified properties
			moon := astronomy.Satellite{
				ID:     uuid.New(),
				Name:   "Test Moon",
				Mass:   astronomy.MoonMassKg * sc.massRatio,
				Radius: astronomy.MoonRadiusMeters * sc.radiusRatio,
			}

			// When: Surface gravity is calculated for FPV physics
			gravity := moon.SurfaceGravity()

			// Then: Gravity should be within expected range
			assert.GreaterOrEqual(t, gravity, sc.gravityMin,
				"Gravity too low for %s: %s", sc.name, sc.playerFeel)
			assert.LessOrEqual(t, gravity, sc.gravityMax,
				"Gravity too high for %s: %s", sc.name, sc.playerFeel)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Horizon Distance Calculation (FPV Visibility)
// -----------------------------------------------------------------------------
// Given: The watcher is in FPV mode on a moon surface
// When: They look toward the horizon
// Then: Smaller bodies should have closer horizons (curved surface visible)
//   AND: The horizon distance should be calculable for rendering optimization
//   AND: Standard eye height (1.7m) should be used

func TestBDD_Satellite_HorizonDistanceForFPV(t *testing.T) {
	const standardEyeHeight = 1.7 // meters

	scenarios := []struct {
		name             string
		radiusRatio      float64 // relative to Moon
		horizonMin       float64 // meters lower bound
		horizonMax       float64 // meters upper bound
		playerExperience string
	}{
		{
			name:             "Earth's Moon",
			radiusRatio:      1.0,
			horizonMin:       2000,
			horizonMax:       3000,
			playerExperience: "Horizon about 2.4km away - noticeably closer than Earth",
		},
		{
			name:             "Small asteroid (1km radius)",
			radiusRatio:      1000.0 / astronomy.MoonRadiusMeters,
			horizonMin:       40,
			horizonMax:       80,
			playerExperience: "Horizon very close - can see curve of surface clearly",
		},
		{
			name:             "Mars-sized moon",
			radiusRatio:      3389.0 / (astronomy.MoonRadiusMeters / 1000),
			horizonMin:       3000,
			horizonMax:       5000,
			playerExperience: "Horizon similar to Earth, feels more familiar",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			// Given: A moon with specified radius
			moon := astronomy.Satellite{
				ID:     uuid.New(),
				Name:   "Test Moon",
				Radius: astronomy.MoonRadiusMeters * sc.radiusRatio,
				Mass:   astronomy.MoonMassKg, // Mass doesn't affect horizon
			}

			// When: Horizon distance is calculated for FPV rendering
			horizon := moon.HorizonDistance(standardEyeHeight)

			// Then: Horizon should be within expected range
			if sc.horizonMin > 0 {
				assert.GreaterOrEqual(t, horizon, sc.horizonMin,
					"Horizon too close for %s: %s", sc.name, sc.playerExperience)
			}
			assert.LessOrEqual(t, horizon, sc.horizonMax,
				"Horizon too far for %s: %s", sc.name, sc.playerExperience)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Moon Destruction State Tracking (Watcher Information)
// -----------------------------------------------------------------------------
// Given: A moon that has been destroyed by asteroid impacts
// When: The watcher views the solar system
// Then: The moon should be marked as destroyed
//   AND: The destruction year should be recorded
//   AND: Whether it formed a ring should be tracked

func TestBDD_Satellite_DestructionTracking(t *testing.T) {
	// Given: A moon that exists initially
	moon := astronomy.Satellite{
		ID:        uuid.New(),
		Name:      "Doomed Moon",
		Mass:      astronomy.MoonMassKg,
		Radius:    astronomy.MoonRadiusMeters,
		Distance:  astronomy.MoonDistanceMeters,
		Destroyed: false,
	}

	// Then: Initially should not be destroyed
	assert.False(t, moon.Destroyed, "Moon should start not destroyed")
	assert.Zero(t, moon.DestroyedAt, "Destruction time should be zero initially")
	assert.False(t, moon.RingFormed, "Ring formed should be false initially")

	// When: Moon is destroyed
	moon.Destroyed = true
	moon.DestroyedAt = 4500000000 // 4.5 billion years
	moon.RingFormed = true

	// Then: Destruction state should be tracked
	assert.True(t, moon.Destroyed, "Moon should be marked destroyed")
	assert.Equal(t, int64(4500000000), moon.DestroyedAt,
		"Destruction year should be recorded for history display")
	assert.True(t, moon.RingFormed,
		"Should track if destruction created a ring for watcher tooltip")
}

// -----------------------------------------------------------------------------
// Scenario: Gravity Formula Accuracy (Physics Validation)
// -----------------------------------------------------------------------------
// Given: Known astronomical body properties
// When: Surface gravity is calculated
// Then: It should match real-world values within tolerance
//   This validates the g = G × M / r² formula implementation

func TestBDD_Satellite_GravityFormulaAccuracy(t *testing.T) {
	scenarios := []struct {
		name            string
		mass            float64
		radius          float64
		expectedGravity float64
		tolerance       float64
		source          string
	}{
		{
			name:            "Earth's Moon (real values)",
			mass:            7.342e22, // kg
			radius:          1.7374e6, // meters
			expectedGravity: 1.62,     // m/s²
			tolerance:       0.05,
			source:          "NASA Moon Fact Sheet",
		},
		{
			name:            "Phobos (Mars moon)",
			mass:            1.0659e16,
			radius:          11.1e3, // ~11km mean radius
			expectedGravity: 0.0057,
			tolerance:       0.002,
			source:          "JPL Small Bodies Database",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			moon := astronomy.Satellite{
				ID:     uuid.New(),
				Mass:   sc.mass,
				Radius: sc.radius,
			}

			gravity := moon.SurfaceGravity()

			assert.InDelta(t, sc.expectedGravity, gravity, sc.tolerance,
				"Gravity calculation should match %s (source: %s)", sc.name, sc.source)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Edge Cases for FPV Physics
// -----------------------------------------------------------------------------
// Given: Invalid or edge-case moon properties
// When: Physics calculations are requested
// Then: Reasonable defaults should be returned (no crashes, no infinities)

func TestBDD_Satellite_EdgeCasesForFPV(t *testing.T) {
	// Zero radius moon (shouldn't crash)
	t.Run("Zero radius moon", func(t *testing.T) {
		moon := astronomy.Satellite{ID: uuid.New(), Mass: 1e20, Radius: 0}
		gravity := moon.SurfaceGravity()
		assert.Equal(t, 0.0, gravity, "Zero radius should return 0 gravity, not infinity")

		horizon := moon.HorizonDistance(1.7)
		assert.Equal(t, 0.0, horizon, "Zero radius should return 0 horizon distance")
	})

	// Zero mass moon
	t.Run("Zero mass moon", func(t *testing.T) {
		moon := astronomy.Satellite{ID: uuid.New(), Mass: 0, Radius: 1e6}
		gravity := moon.SurfaceGravity()
		assert.Equal(t, 0.0, gravity, "Zero mass should return 0 gravity")
		assert.False(t, math.IsNaN(gravity), "Gravity should not be NaN")
	})

	// Negative eye height for horizon
	t.Run("Negative eye height", func(t *testing.T) {
		moon := astronomy.Satellite{ID: uuid.New(), Mass: 1e22, Radius: 1e6}
		horizon := moon.HorizonDistance(-1.0)
		assert.Equal(t, 0.0, horizon, "Negative eye height should return 0")
		assert.False(t, math.IsNaN(horizon), "Horizon should not be NaN")
	})
}
