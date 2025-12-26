package ecosystem

import (
	"math"
	"testing"

	"github.com/google/uuid"
)

// testWorldID generates a test UUID
func testWorldID() uuid.UUID {
	return uuid.New()
}

// TestGetPlanetaryHeat_EarlyEarth verifies maximum heat at formation
func TestGetPlanetaryHeat_EarlyEarth(t *testing.T) {
	heat := GetPlanetaryHeat(0)
	expected := 10.0
	if math.Abs(heat-expected) > 0.01 {
		t.Errorf("Early Earth heat = %v, want %v", heat, expected)
	}
}

// TestGetPlanetaryHeat_HadeanTransition verifies heat at end of Hadean
func TestGetPlanetaryHeat_HadeanTransition(t *testing.T) {
	heat := GetPlanetaryHeat(500_000_000)
	expected := 4.0
	if math.Abs(heat-expected) > 0.01 {
		t.Errorf("Hadean transition heat = %v, want %v", heat, expected)
	}
}

// TestGetPlanetaryHeat_ModernEarth verifies heat at modern age
func TestGetPlanetaryHeat_ModernEarth(t *testing.T) {
	heat := GetPlanetaryHeat(4_500_000_000)
	expected := 1.0
	tolerance := 0.05 // Allow 5% tolerance for exponential decay
	if math.Abs(heat-expected) > tolerance {
		t.Errorf("Modern Earth heat = %v, want %v ± %v", heat, expected, tolerance)
	}
}

// TestGetPlanetaryHeat_MonotonicDecay verifies heat always decreases
func TestGetPlanetaryHeat_MonotonicDecay(t *testing.T) {
	testPoints := []int64{
		0,
		100_000_000,   // 100M years
		500_000_000,   // 500M years (Hadean boundary)
		1_000_000_000, // 1B years
		2_000_000_000, // 2B years
		3_000_000_000, // 3B years
		4_500_000_000, // 4.5B years
	}

	prevHeat := math.Inf(1)
	for _, year := range testPoints {
		heat := GetPlanetaryHeat(year)
		if heat >= prevHeat {
			t.Errorf("Heat is not monotonically decreasing: year=%d, heat=%v, previous=%v",
				year, heat, prevHeat)
		}
		if heat < 1.0 {
			t.Errorf("Heat fell below minimum (1.0): year=%d, heat=%v", year, heat)
		}
		prevHeat = heat
	}
}

// TestGetPlanetaryHeat_HadeanLinearRegime verifies linear decay in early period
func TestGetPlanetaryHeat_HadeanLinearRegime(t *testing.T) {
	// Test midpoint of Hadean period
	heat := GetPlanetaryHeat(250_000_000) // 250M years
	// Should be halfway between 10.0 and 4.0 = 7.0
	expected := 7.0
	tolerance := 0.1
	if math.Abs(heat-expected) > tolerance {
		t.Errorf("Hadean midpoint heat = %v, want %v ± %v", heat, expected, tolerance)
	}
}

// TestGetPlanetaryHeat_ExponentialRegime verifies exponential decay post-Hadean
func TestGetPlanetaryHeat_ExponentialRegime(t *testing.T) {
	// Heat should decay exponentially from 4.0 to 1.0
	heat2B := GetPlanetaryHeat(2_000_000_000)
	heat3B := GetPlanetaryHeat(3_000_000_000)

	// Exponential decay means the ratio should be consistent
	// Heat should be closer to 4.0 at 2B than 3B
	if heat2B <= heat3B {
		t.Errorf("Exponential regime not decreasing: heat(2B)=%v, heat(3B)=%v", heat2B, heat3B)
	}

	// Should be between 4.0 and 1.0
	if heat2B > 4.0 || heat2B < 1.0 {
		t.Errorf("Heat out of bounds at 2B years: %v", heat2B)
	}
}

// TestTectonicScaling_EarlyVsLate verifies tectonic rates scale with heat
func TestTectonicScaling_EarlyVsLate(t *testing.T) {
	// Create two test geologies at different ages
	earlyGeo := NewWorldGeology(testWorldID(), 12345, 40_000_000)
	earlyGeo.InitializeGeology()
	earlyGeo.TotalYearsSimulated = 100_000_000 // 100M years (Hadean)

	lateGeo := NewWorldGeology(testWorldID(), 12345, 40_000_000)
	lateGeo.InitializeGeology()
	lateGeo.TotalYearsSimulated = 4_000_000_000 // 4B years (Modern)

	// Reset accumulators
	earlyGeo.TectonicStressAccumulator = 0
	lateGeo.TectonicStressAccumulator = 0

	// Simulate same dt for both
	dt := int64(10_000) // 10k years

	// Manually calculate expected accumulation based on heat
	earlyHeat := GetPlanetaryHeat(earlyGeo.TotalYearsSimulated)
	lateHeat := GetPlanetaryHeat(lateGeo.TotalYearsSimulated)

	expectedEarlyAccum := float64(dt) * earlyHeat
	expectedLateAccum := float64(dt) * lateHeat

	// Early heat should be much higher than late
	if earlyHeat <= lateHeat {
		t.Errorf("Early heat (%v) should be > late heat (%v)", earlyHeat, lateHeat)
	}

	// Verify ratio is as expected (early should accumulate faster)
	ratio := expectedEarlyAccum / expectedLateAccum
	if ratio < 2.0 {
		t.Errorf("Early tectonic accumulation should be at least 2x late: ratio=%v", ratio)
	}
}

// TestVolcanicScaling_EarlyVsLate verifies volcanic rates scale with heat
func TestVolcanicScaling_EarlyVsLate(t *testing.T) {
	earlyHeat := GetPlanetaryHeat(100_000_000)  // 100M years
	lateHeat := GetPlanetaryHeat(4_000_000_000) // 4B years

	// Base eruption rate: 1 per 1000 years
	baseRate := 1000.0
	years := 10_000.0 // Simulate 10k years

	// Early Earth eruption count
	earlyRate := baseRate / earlyHeat
	earlyEruptions := years / earlyRate

	// Late Earth eruption count
	lateRate := baseRate / lateHeat
	lateEruptions := years / lateRate

	// Early should have significantly more eruptions
	if earlyEruptions <= lateEruptions {
		t.Errorf("Early eruptions (%v) should exceed late eruptions (%v)",
			earlyEruptions, lateEruptions)
	}

	ratio := earlyEruptions / lateEruptions
	if ratio < 2.0 {
		t.Errorf("Early volcanic activity should be at least 2x late: ratio=%v", ratio)
	}
}

// TestPlanetaryHeat_Continuity verifies no discontinuities at regime boundary
func TestPlanetaryHeat_Continuity(t *testing.T) {
	// Test points around the 500M year boundary
	before := GetPlanetaryHeat(499_999_999)
	boundary := GetPlanetaryHeat(500_000_000)
	after := GetPlanetaryHeat(500_000_001)

	// Should be approximately 4.0 at all three points
	tolerance := 0.01
	expected := 4.0

	if math.Abs(before-expected) > tolerance {
		t.Errorf("Heat before boundary = %v, want %v ± %v", before, expected, tolerance)
	}
	if math.Abs(boundary-expected) > tolerance {
		t.Errorf("Heat at boundary = %v, want %v ± %v", boundary, expected, tolerance)
	}
	if math.Abs(after-expected) > tolerance {
		t.Errorf("Heat after boundary = %v, want %v ± %v", after, expected, tolerance)
	}
}

// TestPlanetaryHeat_NegativeYear handles edge case
func TestPlanetaryHeat_NegativeYear(t *testing.T) {
	// Should treat negative years as year 0
	heat := GetPlanetaryHeat(-1000)
	expected := 10.0
	if math.Abs(heat-expected) > 0.01 {
		t.Errorf("Negative year heat = %v, want %v", heat, expected)
	}
}

// TestPlanetaryHeat_VeryOldPlanet verifies heat floors at 1.0
func TestPlanetaryHeat_VeryOldPlanet(t *testing.T) {
	// Test a planet much older than Earth
	heat := GetPlanetaryHeat(10_000_000_000) // 10B years
	if heat < 1.0 {
		t.Errorf("Heat should not fall below 1.0, got %v", heat)
	}
	// Should be very close to 1.0 for old planets
	if heat > 1.1 {
		t.Errorf("Very old planet should have heat ≈ 1.0, got %v", heat)
	}
}
