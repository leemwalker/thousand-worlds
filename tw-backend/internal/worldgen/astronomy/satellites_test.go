package astronomy

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Physical constants for testing
const (
	// Earth's Moon parameters for baseline comparisons
	EarthMass    = 5.972e24  // kg
	MoonMass     = 7.342e22  // kg
	MoonDistance = 384400e3  // meters (384,400 km)
	EarthRadius  = 6.371e6   // meters (6,371 km)
	GravConstant = 6.674e-11 // m³/(kg·s²)
)

// TestGenerateMoons_RandomDistribution verifies the statistical distribution of moon counts
// when no override is set. Expected: ~10% 0, ~60% 1, ~20% 2, ~10% 3+
func TestGenerateMoons_RandomDistribution(t *testing.T) {
	iterations := 1000
	counts := make(map[int]int)

	for i := 0; i < iterations; i++ {
		config := SatelliteConfig{Override: false}
		moons := GenerateMoons(int64(i), EarthMass, config)
		count := len(moons)
		if count >= 3 {
			counts[3]++ // Group 3+ together
		} else {
			counts[count]++
		}
	}

	// Allow 5% tolerance for statistical variance
	tolerance := float64(iterations) * 0.05

	// ~10% should have 0 moons
	assert.InDelta(t, float64(iterations)*0.10, float64(counts[0]), tolerance*2,
		"Expected ~10%% to have 0 moons, got %.1f%%", float64(counts[0])/float64(iterations)*100)

	// ~60% should have 1 moon
	assert.InDelta(t, float64(iterations)*0.60, float64(counts[1]), tolerance*2,
		"Expected ~60%% to have 1 moon, got %.1f%%", float64(counts[1])/float64(iterations)*100)

	// ~20% should have 2 moons
	assert.InDelta(t, float64(iterations)*0.20, float64(counts[2]), tolerance*2,
		"Expected ~20%% to have 2 moons, got %.1f%%", float64(counts[2])/float64(iterations)*100)

	// ~10% should have 3+ moons
	assert.InDelta(t, float64(iterations)*0.10, float64(counts[3]), tolerance*2,
		"Expected ~10%% to have 3+ moons, got %.1f%%", float64(counts[3])/float64(iterations)*100)
}

// TestGenerateMoons_ConfiguredCount verifies that Override=true generates exact count
func TestGenerateMoons_ConfiguredCount(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"Zero moons", 0},
		{"One moon", 1},
		{"Two moons", 2},
		{"Three moons", 3},
		{"Five moons", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SatelliteConfig{Override: true, Count: tt.count}
			moons := GenerateMoons(12345, EarthMass, config)
			assert.Len(t, moons, tt.count)
		})
	}
}

// TestGenerateMoons_OrbitalConstraints verifies all generated moons obey physical constraints
func TestGenerateMoons_OrbitalConstraints(t *testing.T) {
	rocheLimit := 2.5 * EarthRadius // ~15,927 km
	hillSphere := 1.5e9             // 1.5 billion meters

	// Test across many seeds to ensure consistent constraint satisfaction
	for seed := int64(0); seed < 100; seed++ {
		config := SatelliteConfig{Override: true, Count: 3}
		moons := GenerateMoons(seed, EarthMass, config)

		for i, moon := range moons {
			// Roche Limit: orbit must be > 2.5 × PlanetRadius
			assert.Greater(t, moon.Distance, rocheLimit,
				"Seed %d, Moon %d: Distance %.0f < Roche limit %.0f",
				seed, i, moon.Distance, rocheLimit)

			// Hill Sphere: orbit must be < 1.5 × 10^9 m
			assert.Less(t, moon.Distance, hillSphere,
				"Seed %d, Moon %d: Distance %.0f > Hill sphere %.0f",
				seed, i, moon.Distance, hillSphere)

			// Basic sanity checks
			assert.Greater(t, moon.Mass, 0.0, "Moon mass must be positive")
			assert.Greater(t, moon.Radius, 0.0, "Moon radius must be positive")
			assert.Greater(t, moon.Period, 0.0, "Moon period must be positive")
		}
	}
}

// TestGenerateMoons_KeplerPeriod verifies orbital period follows Kepler's 3rd Law
func TestGenerateMoons_KeplerPeriod(t *testing.T) {
	config := SatelliteConfig{Override: true, Count: 1}
	moons := GenerateMoons(42, EarthMass, config)
	require.Len(t, moons, 1)

	moon := moons[0]

	// Calculate expected period using Kepler's 3rd Law: T = 2π√(a³/GM)
	expectedPeriod := 2 * math.Pi * math.Sqrt(math.Pow(moon.Distance, 3)/(GravConstant*EarthMass))

	// Allow 1% tolerance for floating point
	assert.InDelta(t, expectedPeriod, moon.Period, expectedPeriod*0.01,
		"Orbital period should follow Kepler's 3rd Law")
}

// TestGenerateMoons_Deterministic verifies same seed produces same results
func TestGenerateMoons_Deterministic(t *testing.T) {
	config := SatelliteConfig{Override: true, Count: 2}

	moons1 := GenerateMoons(12345, EarthMass, config)
	moons2 := GenerateMoons(12345, EarthMass, config)

	require.Len(t, moons1, 2)
	require.Len(t, moons2, 2)

	for i := range moons1 {
		assert.Equal(t, moons1[i].Mass, moons2[i].Mass, "Moon %d mass should be deterministic", i)
		assert.Equal(t, moons1[i].Distance, moons2[i].Distance, "Moon %d distance should be deterministic", i)
		assert.Equal(t, moons1[i].Period, moons2[i].Period, "Moon %d period should be deterministic", i)
	}
}

// TestCalculateTidalStress_EarthMoonAnalog verifies baseline returns ~1.0
func TestCalculateTidalStress_EarthMoonAnalog(t *testing.T) {
	earthMoon := Satellite{
		Mass:     MoonMass,
		Distance: MoonDistance,
	}

	stress := CalculateTidalStress([]Satellite{earthMoon})

	// Should return ~1.0 for Earth-Moon baseline (±0.05 tolerance)
	assert.InDelta(t, 1.0, stress, 0.05,
		"Earth-Moon analog should produce tidal stress ~1.0, got %.3f", stress)
}

// TestCalculateTidalStress_NoMoons verifies empty slice returns 0.0
func TestCalculateTidalStress_NoMoons(t *testing.T) {
	stress := CalculateTidalStress([]Satellite{})
	assert.Equal(t, 0.0, stress, "No moons should produce zero tidal stress")
}

// TestCalculateTidalStress_MultipleMoons verifies additive effect
func TestCalculateTidalStress_MultipleMoons(t *testing.T) {
	moon := Satellite{Mass: MoonMass, Distance: MoonDistance}

	singleStress := CalculateTidalStress([]Satellite{moon})
	doubleStress := CalculateTidalStress([]Satellite{moon, moon})

	// Two identical moons should produce roughly double the stress
	assert.InDelta(t, singleStress*2, doubleStress, 0.1,
		"Two identical moons should produce ~2x tidal stress")
}

// TestCalculateObliquityStability_LargeMoon verifies large moon provides stability
func TestCalculateObliquityStability_LargeMoon(t *testing.T) {
	// Earth's Moon is ~1.2% of Earth's mass (>0.01 threshold)
	largeMoon := Satellite{Mass: EarthMass * 0.015} // 1.5% of planet mass

	stability := CalculateObliquityStability([]Satellite{largeMoon}, EarthMass)

	assert.Equal(t, 1.0, stability,
		"Large moon (>1%% planet mass) should provide maximum stability")
}

// TestCalculateObliquityStability_SmallMoons verifies small moons provide minimal stability
func TestCalculateObliquityStability_SmallMoons(t *testing.T) {
	// Small moon at 0.1% of planet mass
	smallMoon := Satellite{Mass: EarthMass * 0.001}

	stability := CalculateObliquityStability([]Satellite{smallMoon}, EarthMass)

	assert.Equal(t, 0.1, stability,
		"Small moon (<1%% planet mass) should provide minimal stability")
}

// TestCalculateObliquityStability_NoMoons verifies no moons means chaotic
func TestCalculateObliquityStability_NoMoons(t *testing.T) {
	stability := CalculateObliquityStability([]Satellite{}, EarthMass)
	assert.Equal(t, 0.1, stability, "No moons should result in chaotic obliquity (0.1)")
}

// TestCalculateImpactShielding_Range verifies shielding is bounded correctly
func TestCalculateImpactShielding_Range(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected float64
	}{
		{"No moons", 0, 0.0},
		{"One moon", 1, 0.05},
		{"Two moons", 2, 0.10},
		{"Three moons", 3, 0.15},
		{"Four moons", 4, 0.20},
		{"Five moons (capped)", 5, 0.20}, // Max is 0.20
		{"Ten moons (capped)", 10, 0.20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moons := make([]Satellite, tt.count)
			shielding := CalculateImpactShielding(moons)
			assert.InDelta(t, tt.expected, shielding, 0.0001,
				"Expected %f, got %f", tt.expected, shielding)
		})
	}
}

// BenchmarkGenerateMoons measures moon generation performance
func BenchmarkGenerateMoons(b *testing.B) {
	config := SatelliteConfig{Override: true, Count: 3}
	for i := 0; i < b.N; i++ {
		GenerateMoons(int64(i), EarthMass, config)
	}
}

// BenchmarkCalculateTidalStress measures tidal calculation performance
func BenchmarkCalculateTidalStress(b *testing.B) {
	moons := []Satellite{
		{Mass: MoonMass, Distance: MoonDistance},
		{Mass: MoonMass / 2, Distance: MoonDistance * 1.5},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateTidalStress(moons)
	}
}
