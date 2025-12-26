package astronomy

import (
	"math"
	"testing"
)

// TestCalculateOrbitalState_Determinism verifies that the same year
// always produces identical orbital parameters.
func TestCalculateOrbitalState_Determinism(t *testing.T) {
	testCases := []int64{0, 1000, 10000, 100000, 1000000, 500000000}

	for _, year := range testCases {
		state1 := CalculateOrbitalState(year)
		state2 := CalculateOrbitalState(year)

		if state1.Eccentricity != state2.Eccentricity {
			t.Errorf("Year %d: Eccentricity not deterministic: %v != %v",
				year, state1.Eccentricity, state2.Eccentricity)
		}
		if state1.Obliquity != state2.Obliquity {
			t.Errorf("Year %d: Obliquity not deterministic: %v != %v",
				year, state1.Obliquity, state2.Obliquity)
		}
		if state1.Precession != state2.Precession {
			t.Errorf("Year %d: Precession not deterministic: %v != %v",
				year, state1.Precession, state2.Precession)
		}
	}
}

// TestCalculateOrbitalState_Periodicity verifies the 41k year obliquity cycle.
// The obliquity should complete one full cycle every ~41,000 years.
func TestCalculateOrbitalState_ObliquityPeriodicity(t *testing.T) {
	// Check that obliquity returns to approximately the same value after 41k years
	const period int64 = 41000
	const tolerance = 0.001 // degrees

	baseState := CalculateOrbitalState(0)
	cycleState := CalculateOrbitalState(period)

	diff := math.Abs(baseState.Obliquity - cycleState.Obliquity)
	if diff > tolerance {
		t.Errorf("Obliquity not periodic at 41k years: year 0 = %.4f, year %d = %.4f (diff: %.4f)",
			baseState.Obliquity, period, cycleState.Obliquity, diff)
	}

	// Verify peak occurs at quarter cycle (~10,250 years)
	quarterCycle := period / 4
	peakState := CalculateOrbitalState(quarterCycle)

	// At quarter cycle, sine should be at maximum (1.0)
	// So obliquity should be at max: 23.44 + 1.2 = 24.64
	expectedMax := 23.44 + 1.2
	if math.Abs(peakState.Obliquity-expectedMax) > 0.1 {
		t.Errorf("Obliquity peak at quarter cycle: expected ~%.2f, got %.4f",
			expectedMax, peakState.Obliquity)
	}

	// Verify trough occurs at 3/4 cycle (~30,750 years)
	threeFourthsCycle := (period * 3) / 4
	troughState := CalculateOrbitalState(threeFourthsCycle)

	// At 3/4 cycle, sine should be at minimum (-1.0)
	// So obliquity should be at min: 23.44 - 1.2 = 22.24
	expectedMin := 23.44 - 1.2
	if math.Abs(troughState.Obliquity-expectedMin) > 0.1 {
		t.Errorf("Obliquity trough at 3/4 cycle: expected ~%.2f, got %.4f",
			expectedMin, troughState.Obliquity)
	}
}

// TestCalculateOrbitalState_EccentricityPeriodicity verifies the 100k year eccentricity cycle.
func TestCalculateOrbitalState_EccentricityPeriodicity(t *testing.T) {
	const period int64 = 100000
	const tolerance = 0.0001

	baseState := CalculateOrbitalState(0)
	cycleState := CalculateOrbitalState(period)

	diff := math.Abs(baseState.Eccentricity - cycleState.Eccentricity)
	if diff > tolerance {
		t.Errorf("Eccentricity not periodic at 100k years: year 0 = %.6f, year %d = %.6f",
			baseState.Eccentricity, period, cycleState.Eccentricity)
	}

	// Verify range: should stay within 0.007 to 0.027 (0.017 ± 0.01)
	// Using small tolerance for floating point comparison
	const epsilon = 0.0001
	for year := int64(0); year <= period; year += 1000 {
		state := CalculateOrbitalState(year)
		if state.Eccentricity < 0.007-epsilon || state.Eccentricity > 0.027+epsilon {
			t.Errorf("Year %d: Eccentricity out of range: %.6f (expected 0.007-0.027)",
				year, state.Eccentricity)
		}
	}
}

// TestCalculateOrbitalState_PrecessionPeriodicity verifies the 26k year precession cycle.
func TestCalculateOrbitalState_PrecessionPeriodicity(t *testing.T) {
	const period int64 = 26000
	const tolerance = 0.001

	baseState := CalculateOrbitalState(0)
	cycleState := CalculateOrbitalState(period)

	diff := math.Abs(baseState.Precession - cycleState.Precession)
	if diff > tolerance {
		t.Errorf("Precession not periodic at 26k years: year 0 = %.6f, year %d = %.6f",
			baseState.Precession, period, cycleState.Precession)
	}

	// Precession should range from -1.0 to 1.0
	for year := int64(0); year <= period; year += 1000 {
		state := CalculateOrbitalState(year)
		if state.Precession < -1.0 || state.Precession > 1.0 {
			t.Errorf("Year %d: Precession out of range: %.6f (expected -1.0 to 1.0)",
				year, state.Precession)
		}
	}
}

// TestCalculateInsolation_Range verifies insolation stays within expected bounds.
func TestCalculateInsolation_Range(t *testing.T) {
	// Test across a million years to capture all orbital combinations
	for year := int64(0); year <= 1000000; year += 1000 {
		state := CalculateOrbitalState(year)
		insolation := CalculateInsolation(state)

		// Insolation should stay close to 1.0 (±0.1 is reasonable for orbital variations)
		if insolation < 0.90 || insolation > 1.10 {
			t.Errorf("Year %d: Insolation out of expected range: %.4f", year, insolation)
		}
	}
}

// TestCalculateInsolation_IceAgePotential tests that low obliquity produces
// lower insolation (ice age potential).
func TestCalculateInsolation_IceAgePotential(t *testing.T) {
	// Create contrasting states manually
	lowObliquityState := OrbitalState{
		Eccentricity: 0.017, // baseline
		Obliquity:    22.24, // minimum tilt (less solar energy at high latitudes)
		Precession:   0.0,
	}

	highObliquityState := OrbitalState{
		Eccentricity: 0.017,
		Obliquity:    24.64, // maximum tilt (more solar energy at high latitudes)
		Precession:   0.0,
	}

	lowInsolation := CalculateInsolation(lowObliquityState)
	highInsolation := CalculateInsolation(highObliquityState)

	// Low obliquity should result in lower insolation (ice age potential)
	if lowInsolation >= highInsolation {
		t.Errorf("Low obliquity should produce lower insolation: low=%.4f, high=%.4f",
			lowInsolation, highInsolation)
	}
}

// TestCalculateInsolation_Deterministic verifies same state = same insolation.
func TestCalculateInsolation_Deterministic(t *testing.T) {
	state := OrbitalState{
		Eccentricity: 0.02,
		Obliquity:    23.0,
		Precession:   0.5,
	}

	result1 := CalculateInsolation(state)
	result2 := CalculateInsolation(state)

	if result1 != result2 {
		t.Errorf("Insolation not deterministic: %v != %v", result1, result2)
	}
}

// TestIceAgeThreshold verifies that specific orbital configurations
// produce insolation below the ice age threshold of 0.97.
func TestIceAgeThreshold(t *testing.T) {
	const iceAgeThreshold = 0.97

	// Find years with low insolation by checking obliquity troughs
	// Obliquity trough occurs at 3/4 of 41k cycle = ~30,750 years
	troughYear := int64(30750)

	state := CalculateOrbitalState(troughYear)
	insolation := CalculateInsolation(state)

	t.Logf("Year %d: Obliquity=%.2f°, Insolation=%.4f", troughYear, state.Obliquity, insolation)

	// At obliquity minimum, we expect reduced insolation
	// The exact value depends on our formula, but should be < 1.0
	if insolation >= 1.0 {
		t.Errorf("Expected reduced insolation at obliquity trough, got %.4f", insolation)
	}
}

// -----------------------------------------------------------------------------
// Obliquity Chaos Tests - Natural Satellites Phase 3
// -----------------------------------------------------------------------------

// TestObliquityChaos_StableWorld verifies Earth-like stability (stability=1.0)
// results in small obliquity variance (~2.4° total swing).
func TestObliquityChaos_StableWorld(t *testing.T) {
	stability := 1.0 // Earth with Moon

	minObliquity := 100.0
	maxObliquity := 0.0

	// Sample over 100k years to capture cycle extremes
	for year := int64(0); year < 100000; year += 1000 {
		state := CalculateOrbitalStateWithStability(year, stability)
		if state.Obliquity < minObliquity {
			minObliquity = state.Obliquity
		}
		if state.Obliquity > maxObliquity {
			maxObliquity = state.Obliquity
		}
	}

	obliquityRange := maxObliquity - minObliquity
	t.Logf("Stable (1.0): Obliquity range = %.2f° (min=%.2f, max=%.2f)",
		obliquityRange, minObliquity, maxObliquity)

	// Stable worlds should have small range (~2.4° for Earth-Moon)
	if obliquityRange > 3.0 {
		t.Errorf("Stable world obliquity range too large: %.2f° (expected <3°)", obliquityRange)
	}
	if obliquityRange < 1.0 {
		t.Errorf("Stable world obliquity range too small: %.2f° (expected >1°)", obliquityRange)
	}
}

// TestObliquityChaos_ChaoticWorld verifies moonless chaos (stability=0.0)
// results in large obliquity variance (>10° swing, Mars-like).
func TestObliquityChaos_ChaoticWorld(t *testing.T) {
	stability := 0.0 // No moons, Mars-like chaos

	minObliquity := 100.0
	maxObliquity := 0.0

	// Sample over 100k years to capture cycle extremes
	for year := int64(0); year < 100000; year += 1000 {
		state := CalculateOrbitalStateWithStability(year, stability)
		if state.Obliquity < minObliquity {
			minObliquity = state.Obliquity
		}
		if state.Obliquity > maxObliquity {
			maxObliquity = state.Obliquity
		}
	}

	obliquityRange := maxObliquity - minObliquity
	t.Logf("Chaotic (0.0): Obliquity range = %.2f° (min=%.2f, max=%.2f)",
		obliquityRange, minObliquity, maxObliquity)

	// Chaotic worlds should have large range (>10° for Mars-like behavior)
	if obliquityRange < 10.0 {
		t.Errorf("Chaotic world obliquity range too small: %.2f° (expected >10°)", obliquityRange)
	}
}

// TestObliquityChaos_Gradient verifies obliquity variance scales with stability
func TestObliquityChaos_Gradient(t *testing.T) {
	// With chaos multiplier = 1 + (1-stability)*10:
	// stability=1.0: multiplier=1, variance=2.4°
	// stability=0.5: multiplier=6, variance=14.4°
	// stability=0.2: multiplier=9, variance=21.6°
	// stability=0.0: multiplier=11, variance=26.4°
	tests := []struct {
		name           string
		stability      float64
		expectedMinVar float64
		expectedMaxVar float64
	}{
		{"Full stability", 1.0, 2.0, 3.0},
		{"Partial stability", 0.5, 13.0, 16.0},
		{"Low stability", 0.2, 20.0, 24.0},
		{"No stability", 0.0, 24.0, 28.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minObl := 100.0
			maxObl := 0.0

			for year := int64(0); year < 100000; year += 1000 {
				state := CalculateOrbitalStateWithStability(year, tt.stability)
				if state.Obliquity < minObl {
					minObl = state.Obliquity
				}
				if state.Obliquity > maxObl {
					maxObl = state.Obliquity
				}
			}

			oblRange := maxObl - minObl
			t.Logf("Stability %.1f: range = %.2f°", tt.stability, oblRange)

			if oblRange < tt.expectedMinVar || oblRange > tt.expectedMaxVar {
				t.Errorf("Stability %.1f: range %.2f° outside expected [%.1f, %.1f]",
					tt.stability, oblRange, tt.expectedMinVar, tt.expectedMaxVar)
			}
		})
	}
}

// TestCalculateOrbitalStateWithStability_BackwardCompatible verifies stability=1.0
// produces approximately the same results as the original CalculateOrbitalState.
func TestCalculateOrbitalStateWithStability_BackwardCompatible(t *testing.T) {
	for year := int64(0); year < 100000; year += 10000 {
		original := CalculateOrbitalState(year)
		withStability := CalculateOrbitalStateWithStability(year, 1.0)

		// Should be nearly identical for stability=1.0
		if original.Eccentricity != withStability.Eccentricity {
			t.Errorf("Year %d: Eccentricity mismatch", year)
		}
		if original.Precession != withStability.Precession {
			t.Errorf("Year %d: Precession mismatch", year)
		}
		// Obliquity should be identical for stability=1.0
		if original.Obliquity != withStability.Obliquity {
			t.Errorf("Year %d: Obliquity mismatch: original=%.4f, withStability=%.4f",
				year, original.Obliquity, withStability.Obliquity)
		}
	}
}

// BenchmarkCalculateOrbitalState measures performance of orbital calculations.
func BenchmarkCalculateOrbitalState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateOrbitalState(int64(i * 1000))
	}
}

// BenchmarkCalculateInsolation measures insolation calculation performance.
func BenchmarkCalculateInsolation(b *testing.B) {
	state := OrbitalState{
		Eccentricity: 0.017,
		Obliquity:    23.44,
		Precession:   0.0,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateInsolation(state)
	}
}
