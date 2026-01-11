package astronomy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormRingFromMoonDebris(t *testing.T) {
	tests := []struct {
		name                string
		destructionSeverity float64
		expectRing          bool
		expectedStage       RingFormationStage
	}{
		{
			name:                "complete destruction forms ring",
			destructionSeverity: 1.0,
			expectRing:          true,
			expectedStage:       RingStageChunks,
		},
		{
			name:                "moderate destruction forms ring",
			destructionSeverity: 0.5,
			expectRing:          true,
			expectedStage:       RingStageChunks,
		},
		{
			name:                "minimal destruction threshold",
			destructionSeverity: 0.3,
			expectRing:          true,
			expectedStage:       RingStageChunks,
		},
		{
			name:                "insufficient destruction no ring",
			destructionSeverity: 0.2,
			expectRing:          false,
		},
		{
			name:                "zero destruction no ring",
			destructionSeverity: 0.0,
			expectRing:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moon := Satellite{
				Mass:     MoonMassKg,
				Distance: MoonDistanceMeters,
				Radius:   MoonRadiusMeters,
			}

			ring := FormRingFromMoonDebris(moon, tt.destructionSeverity, 1000)

			if tt.expectRing {
				require.NotNil(t, ring, "expected ring to form")
				assert.Equal(t, tt.expectedStage, ring.Stage)
				assert.Greater(t, ring.Mass, 0.0)
				assert.Greater(t, ring.InnerRadius, 0.0)
				assert.Greater(t, ring.OuterRadius, ring.InnerRadius)
			} else {
				assert.Nil(t, ring, "expected no ring")
			}
		})
	}
}

func TestRingSystemStageProgression(t *testing.T) {
	rs := NewRingSystem()

	moon := Satellite{
		Mass:     MoonMassKg,
		Distance: MoonDistanceMeters,
		Radius:   MoonRadiusMeters,
	}

	ring := FormRingFromMoonDebris(moon, 1.0, 0)
	require.NotNil(t, ring)
	rs.AddRing(ring)

	// Initially at chunks stage
	assert.Equal(t, RingStageChunks, rs.Rings[0].Stage)
	assert.Len(t, rs.GetVisibleRings(), 0, "chunks not visible yet")

	// After 1 year: debris stage
	rs.UpdateRingStages(ChunksToDebrisYears)
	assert.Equal(t, RingStageDebris, rs.Rings[0].Stage)
	assert.Len(t, rs.GetVisibleRings(), 1, "debris is visible")

	// After 10 years: spreading stage
	rs.UpdateRingStages(DebrisToSpreadingYears)
	assert.Equal(t, RingStageSpreading, rs.Rings[0].Stage)

	// After 100 years: stable stage
	rs.UpdateRingStages(SpreadingToStableYears)
	assert.Equal(t, RingStageStable, rs.Rings[0].Stage)
}

func TestRingMassCalculation(t *testing.T) {
	rs := NewRingSystem()

	moon := Satellite{
		Mass:     MoonMassKg,
		Distance: MoonDistanceMeters,
		Radius:   MoonRadiusMeters,
	}

	ring := FormRingFromMoonDebris(moon, 1.0, 0)
	require.NotNil(t, ring)
	rs.AddRing(ring)

	// Ring mass should be ~10% of moon mass for complete destruction
	expectedMaxMass := MoonMassKg * 0.1
	assert.LessOrEqual(t, ring.Mass, expectedMaxMass)
	assert.Greater(t, ring.Mass, 0.0)

	// Total mass method
	assert.Equal(t, ring.Mass, rs.TotalRingMass())
}

func TestRingRadiusConstraints(t *testing.T) {
	moon := Satellite{
		Mass:     MoonMassKg,
		Distance: MoonDistanceMeters,
		Radius:   MoonRadiusMeters,
	}

	ring := FormRingFromMoonDebris(moon, 1.0, 0)
	require.NotNil(t, ring)

	// Inner radius should be at or above Roche limit
	rocheLimit := RocheLimitFactor * EarthRadiusMeters
	assert.GreaterOrEqual(t, ring.InnerRadius, rocheLimit*0.8, "inner radius near Roche limit")

	// Outer radius should be reasonable relative to moon distance
	assert.LessOrEqual(t, ring.OuterRadius, moon.Distance*1.5)
}
