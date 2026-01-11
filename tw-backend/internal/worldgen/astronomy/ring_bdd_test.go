package astronomy_test

import (
	"testing"

	"tw-backend/internal/worldgen/astronomy"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BDD Tests: Ring System
// =============================================================================
// These tests document user-visible behavior for the ring formation system.
// Rings form from destroyed moons and progress through visible stages.

// -----------------------------------------------------------------------------
// Scenario: Ring Formation from Moon Destruction (User Observation)
// -----------------------------------------------------------------------------
// Given: A moon is destroyed by asteroid impacts (destruction severity 50%+)
// When: The watcher observes the solar system over time
// Then: Debris should initially appear as chunks near moon's orbit
//   AND: After ~1 year, debris should spread along the orbital path
//   AND: After ~10 years, debris should flatten into a visible ring plane
//   AND: After ~100 years, the ring should appear stable and Saturn-like

func TestBDD_Ring_FormationFromMoonDestruction(t *testing.T) {
	scenarios := []struct {
		name                string
		destructionSeverity float64
		expectRingFormed    bool
		reason              string
	}{
		{
			name:                "Complete moon destruction creates ring",
			destructionSeverity: 1.0,
			expectRingFormed:    true,
			reason:              "Total destruction should create visible debris field",
		},
		{
			name:                "Severe asteroid damage (70%) creates ring",
			destructionSeverity: 0.7,
			expectRingFormed:    true,
			reason:              "Major damage should eject enough material for ring",
		},
		{
			name:                "Moderate damage (50%) creates ring",
			destructionSeverity: 0.5,
			expectRingFormed:    true,
			reason:              "50% destruction is threshold for ring formation",
		},
		{
			name:                "Threshold damage (30%) creates minimal ring",
			destructionSeverity: 0.3,
			expectRingFormed:    true,
			reason:              "30% is minimum for ring formation",
		},
		{
			name:                "Surface scarring (20%) does not create ring",
			destructionSeverity: 0.2,
			expectRingFormed:    false,
			reason:              "Surface damage only, not enough ejecta",
		},
		{
			name:                "Crater formation (10%) does not create ring",
			destructionSeverity: 0.1,
			expectRingFormed:    false,
			reason:              "Normal impact craters don't form rings",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			// Given: A moon with Earth-Moon-like properties
			moon := astronomy.Satellite{
				ID:       uuid.New(),
				Name:     "Test Moon",
				Mass:     astronomy.MoonMassKg,
				Distance: astronomy.MoonDistanceMeters,
				Radius:   astronomy.MoonRadiusMeters,
			}

			// When: The moon is destroyed with the given severity
			ring := astronomy.FormRingFromMoonDebris(moon, sc.destructionSeverity, 1000)

			// Then: Ring formation should match expectations
			if sc.expectRingFormed {
				require.NotNil(t, ring, "Ring should form: %s", sc.reason)
				assert.Equal(t, astronomy.RingStageChunks, ring.Stage,
					"New ring should start in chunks stage")
				assert.Equal(t, moon.ID, ring.OriginMoonID,
					"Ring should track its origin moon")
			} else {
				assert.Nil(t, ring, "Ring should NOT form: %s", sc.reason)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Ring Stage Progression Over Time (User Observation Timeline)
// -----------------------------------------------------------------------------
// Given: A ring has formed from moon debris
// When: The watcher observes the system over simulation years
// Then: The ring should progress through visually distinct stages:
//   - Year 0-1: Chunks (large fragments, sparse, chaotic)
//   - Year 1-10: Debris (spreading, some structure visible)
//   - Year 10-100: Spreading (flattening, ring shape emerging)
//   - Year 100+: Stable (smooth, Saturn-like appearance)

func TestBDD_Ring_StageProgressionTimeline(t *testing.T) {
	// Given: A ring system with a freshly formed ring
	rs := astronomy.NewRingSystem()
	moon := astronomy.Satellite{
		ID:       uuid.New(),
		Mass:     astronomy.MoonMassKg,
		Distance: astronomy.MoonDistanceMeters,
		Radius:   astronomy.MoonRadiusMeters,
	}
	ring := astronomy.FormRingFromMoonDebris(moon, 1.0, 0)
	require.NotNil(t, ring)
	rs.AddRing(ring)

	// Timeline scenarios
	observations := []struct {
		description   string
		yearsElapsed  int64
		expectedStage astronomy.RingFormationStage
		isVisible     bool
		visualNotes   string
	}{
		{
			description:   "Day 1: Fresh destruction",
			yearsElapsed:  0,
			expectedStage: astronomy.RingStageChunks,
			isVisible:     false,
			visualNotes:   "Large irregular fragments orbiting chaotically",
		},
		{
			description:   "Year 1: Debris spreading",
			yearsElapsed:  1,
			expectedStage: astronomy.RingStageDebris,
			isVisible:     true,
			visualNotes:   "Debris spreading along orbital path, some structure visible",
		},
		{
			description:   "Year 10: Ring plane forming",
			yearsElapsed:  10,
			expectedStage: astronomy.RingStageSpreading,
			isVisible:     true,
			visualNotes:   "Debris flattening into ring plane, disc shape emerging",
		},
		{
			description:   "Year 100: Stable ring",
			yearsElapsed:  100,
			expectedStage: astronomy.RingStageStable,
			isVisible:     true,
			visualNotes:   "Smooth Saturn-like ring, stable orbital mechanics",
		},
	}

	for _, obs := range observations {
		t.Run(obs.description, func(t *testing.T) {
			// When: Time passes in the simulation
			testRings := astronomy.NewRingSystem()
			testRing := astronomy.FormRingFromMoonDebris(moon, 1.0, 0)
			testRings.AddRing(testRing)
			testRings.UpdateRingStages(obs.yearsElapsed)

			// Then: Ring should be in expected stage
			assert.Equal(t, obs.expectedStage, testRings.Rings[0].Stage,
				"Ring stage at year %d: %s", obs.yearsElapsed, obs.visualNotes)

			// And: Visibility should match (chunks not visible from space)
			visibleRings := testRings.GetVisibleRings()
			if obs.isVisible {
				assert.Len(t, visibleRings, 1,
					"Ring should be visible to watcher at year %d", obs.yearsElapsed)
			} else {
				assert.Len(t, visibleRings, 0,
					"Ring chunks not yet visible from space at year %d", obs.yearsElapsed)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Ring Physical Properties (Watcher Information Display)
// -----------------------------------------------------------------------------
// Given: A ring formed from a destroyed moon
// When: The watcher clicks on the ring to see information
// Then: The following properties should be displayed:
//   - Mass (smaller than original moon)
//   - Inner and outer radius
//   - Composition (ice, rock, dust, mixed)
//   - Formation date (simulation year)

func TestBDD_Ring_PhysicalPropertiesForDisplay(t *testing.T) {
	// Given: A moon with known properties
	moon := astronomy.Satellite{
		ID:       uuid.New(),
		Name:     "Europa",
		Mass:     astronomy.MoonMassKg * 0.65, // Europa-sized
		Distance: astronomy.MoonDistanceMeters * 1.7,
		Radius:   astronomy.MoonRadiusMeters * 0.9,
	}

	// When: Ring forms from destruction
	currentYear := int64(4500000000) // 4.5 billion years
	ring := astronomy.FormRingFromMoonDebris(moon, 1.0, currentYear)
	require.NotNil(t, ring)

	// Then: All display properties should be set
	assert.NotEqual(t, uuid.Nil, ring.ID, "Ring should have unique ID for tracking")
	assert.Greater(t, ring.Mass, 0.0, "Ring mass should be positive")
	assert.Less(t, ring.Mass, moon.Mass, "Ring mass should be less than moon (most debris escapes)")
	assert.Greater(t, ring.InnerRadius, 0.0, "Inner radius should be set")
	assert.Greater(t, ring.OuterRadius, ring.InnerRadius, "Outer radius > inner radius")
	assert.NotEmpty(t, ring.Composition, "Composition should be specified for display")
	assert.Equal(t, currentYear, ring.FormedAtYear, "Formation year should be recorded")
	assert.NotEmpty(t, ring.Color, "Ring should have color for visualization")
}

// -----------------------------------------------------------------------------
// Scenario: Multiple Rings from Multiple Moon Destructions
// -----------------------------------------------------------------------------
// Given: A planet with multiple moons
// When: Multiple moons are destroyed over time
// Then: The ring system should display multiple distinct rings
//   AND: Each ring should track its origin moon

func TestBDD_Ring_MultipleMoonDestructions(t *testing.T) {
	rs := astronomy.NewRingSystem()

	// Given: Three moons at different orbital distances
	moons := []astronomy.Satellite{
		{ID: uuid.New(), Name: "Inner Moon", Mass: astronomy.MoonMassKg * 0.1, Distance: astronomy.MoonDistanceMeters * 0.5, Radius: astronomy.MoonRadiusMeters * 0.3},
		{ID: uuid.New(), Name: "Middle Moon", Mass: astronomy.MoonMassKg * 0.5, Distance: astronomy.MoonDistanceMeters, Radius: astronomy.MoonRadiusMeters * 0.5},
		{ID: uuid.New(), Name: "Outer Moon", Mass: astronomy.MoonMassKg * 2.0, Distance: astronomy.MoonDistanceMeters * 2.0, Radius: astronomy.MoonRadiusMeters * 1.2},
	}

	// When: Each moon is destroyed in different years
	for i, moon := range moons {
		ring := astronomy.FormRingFromMoonDebris(moon, 0.8, int64(i*10))
		rs.AddRing(ring)
	}

	// Then: System should have three rings
	assert.Len(t, rs.Rings, 3, "Should have 3 rings from 3 moon destructions")

	// And: Each ring should track its origin
	for i, ring := range rs.Rings {
		assert.Equal(t, moons[i].ID, ring.OriginMoonID,
			"Ring %d should track origin moon", i)
	}

	// And: Total mass should be calculable for display
	totalMass := rs.TotalRingMass()
	assert.Greater(t, totalMass, 0.0, "Total ring mass should be positive")
}
