package ocean

import (
	"testing"

	"tw-backend/internal/worldgen/astronomy"

	"github.com/stretchr/testify/assert"
)

// Physical constants for testing
const (
	// Earth's Moon parameters for baseline
	testMoonMass     = 7.342e22 // kg
	testMoonDistance = 384400e3 // meters (384,400 km)
)

// TestCalculateTidalAmplitude_EarthMoon verifies Earth-Moon returns ~1.0m baseline
func TestCalculateTidalAmplitude_EarthMoon(t *testing.T) {
	earthMoon := astronomy.Satellite{
		Mass:     testMoonMass,
		Distance: testMoonDistance,
	}

	amplitude := CalculateTidalAmplitude([]astronomy.Satellite{earthMoon})

	// Earth-Moon baseline should be ~1.0 meters (±0.1 tolerance)
	assert.InDelta(t, 1.0, amplitude, 0.1,
		"Earth-Moon analog should produce ~1.0m tidal amplitude, got %.3f", amplitude)
}

// TestCalculateTidalAmplitude_NoMoons verifies empty slice returns 0.0
func TestCalculateTidalAmplitude_NoMoons(t *testing.T) {
	amplitude := CalculateTidalAmplitude([]astronomy.Satellite{})
	assert.Equal(t, 0.0, amplitude, "No moons should produce zero tidal amplitude")
}

// TestCalculateTidalAmplitude_MultipleMoons verifies additive effect
func TestCalculateTidalAmplitude_MultipleMoons(t *testing.T) {
	moon := astronomy.Satellite{Mass: testMoonMass, Distance: testMoonDistance}

	singleAmp := CalculateTidalAmplitude([]astronomy.Satellite{moon})
	doubleAmp := CalculateTidalAmplitude([]astronomy.Satellite{moon, moon})

	// Two identical moons should produce roughly double the amplitude
	assert.InDelta(t, singleAmp*2, doubleAmp, 0.1,
		"Two identical moons should produce ~2x amplitude")
}

// TestCalculateTidalAmplitude_CloserMoonStronger verifies distance effect
func TestCalculateTidalAmplitude_CloserMoonStronger(t *testing.T) {
	farMoon := astronomy.Satellite{Mass: testMoonMass, Distance: testMoonDistance}
	closeMoon := astronomy.Satellite{Mass: testMoonMass, Distance: testMoonDistance / 2}

	farAmp := CalculateTidalAmplitude([]astronomy.Satellite{farMoon})
	closeAmp := CalculateTidalAmplitude([]astronomy.Satellite{closeMoon})

	// Closer moon (half distance) should have 8x effect (1/d³)
	assert.Greater(t, closeAmp, farAmp*7.0,
		"Moon at half distance should have >7x amplitude")
}

// TestCalculateTidalAmplitude_MassiveMoonStronger verifies mass effect
func TestCalculateTidalAmplitude_MassiveMoonStronger(t *testing.T) {
	normalMoon := astronomy.Satellite{Mass: testMoonMass, Distance: testMoonDistance}
	massiveMoon := astronomy.Satellite{Mass: testMoonMass * 2, Distance: testMoonDistance}

	normalAmp := CalculateTidalAmplitude([]astronomy.Satellite{normalMoon})
	massiveAmp := CalculateTidalAmplitude([]astronomy.Satellite{massiveMoon})

	// Double mass should produce double amplitude
	assert.InDelta(t, normalAmp*2, massiveAmp, 0.1,
		"Double mass moon should produce ~2x amplitude")
}

// TestGetTidalCategory_Classifications verifies category boundaries
func TestGetTidalCategory_Classifications(t *testing.T) {
	tests := []struct {
		name      string
		amplitude float64
		expected  string
	}{
		{"Zero", 0.0, "micro"},
		{"Micro low", 0.3, "micro"},
		{"Micro edge", 0.49, "micro"},
		{"Normal low", 0.5, "normal"},
		{"Normal mid", 1.0, "normal"},
		{"Normal high", 1.9, "normal"},
		{"Strong low", 2.0, "strong"},
		{"Strong mid", 3.5, "strong"},
		{"Strong edge", 4.9, "strong"},
		{"Extreme low", 5.0, "extreme"},
		{"Extreme high", 50.0, "extreme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := GetTidalCategory(tt.amplitude)
			assert.Equal(t, tt.expected, category,
				"Amplitude %.2f should be '%s', got '%s'", tt.amplitude, tt.expected, category)
		})
	}
}

// BenchmarkCalculateTidalAmplitude measures performance
func BenchmarkCalculateTidalAmplitude(b *testing.B) {
	moons := []astronomy.Satellite{
		{Mass: testMoonMass, Distance: testMoonDistance},
		{Mass: testMoonMass / 2, Distance: testMoonDistance * 1.5},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateTidalAmplitude(moons)
	}
}
