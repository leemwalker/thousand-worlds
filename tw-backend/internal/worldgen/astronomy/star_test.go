package astronomy

import (
	"math"
	"testing"
)

// TestGetSolarLuminosity verifies the Gough (1981) solar evolution model
func TestGetSolarLuminosity(t *testing.T) {
	tests := []struct {
		name      string
		year      int64
		expected  float64
		tolerance float64
	}{
		{
			name:      "early Hadean (year 0)",
			year:      0,
			expected:  0.714, // ~71.4% modern brightness
			tolerance: 0.01,
		},
		{
			name:      "mid Archean (1 billion years)",
			year:      1_000_000_000,
			expected:  0.758,
			tolerance: 0.01,
		},
		{
			name:      "mid history (2.25 billion years)",
			year:      2_250_000_000,
			expected:  0.833,
			tolerance: 0.01,
		},
		{
			name:      "late history (4 billion years)",
			year:      4_000_000_000,
			expected:  0.956,
			tolerance: 0.01,
		},
		{
			name:      "modern Earth (4.5 billion years)",
			year:      4_500_000_000,
			expected:  1.0,
			tolerance: 0.001,
		},
		{
			name:      "negative year clamped to zero",
			year:      -1_000_000_000,
			expected:  0.714,
			tolerance: 0.01,
		},
		{
			name:      "future year clamped to maximum",
			year:      5_000_000_000,
			expected:  1.0,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSolarLuminosity(tt.year)

			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("GetSolarLuminosity(%d) = %f, want %f (±%f)",
					tt.year, result, tt.expected, tt.tolerance)
			}
		})
	}
}

// TestSolarLuminosity_Monotonic verifies brightness increases over time
func TestSolarLuminosity_Monotonic(t *testing.T) {
	// Sample points across Earth's history
	years := []int64{
		0,
		500_000_000,
		1_000_000_000,
		2_000_000_000,
		3_000_000_000,
		4_000_000_000,
		4_500_000_000,
	}

	var prevLuminosity float64 = 0.0
	for _, year := range years {
		luminosity := GetSolarLuminosity(year)

		if luminosity <= prevLuminosity {
			t.Errorf("Solar luminosity not monotonically increasing at year %d: %f <= %f",
				year, luminosity, prevLuminosity)
		}

		prevLuminosity = luminosity
	}
}

// TestSolarLuminosity_Range verifies values stay within physical bounds
func TestSolarLuminosity_Range(t *testing.T) {
	// Test a broad range of years
	for year := int64(0); year <= 4_500_000_000; year += 100_000_000 {
		luminosity := GetSolarLuminosity(year)

		if luminosity < 0.7 || luminosity > 1.0 {
			t.Errorf("Solar luminosity out of physical range at year %d: %f",
				year, luminosity)
		}
	}
}

// TestSolarLuminosity_GoughFormula verifies the mathematical formula
func TestSolarLuminosity_GoughFormula(t *testing.T) {
	// Manually calculate using Gough (1981) formula for verification
	// L(t) = 1.0 / (1.0 + 0.4 * (1.0 - t/t_now))

	year := int64(2_000_000_000)
	t_now := 4_500_000_000.0
	t_norm := float64(year) / t_now

	expected := 1.0 / (1.0 + 0.4*(1.0-t_norm))
	result := GetSolarLuminosity(year)

	if math.Abs(result-expected) > 0.001 {
		t.Errorf("Formula mismatch: got %f, want %f", result, expected)
	}
}
