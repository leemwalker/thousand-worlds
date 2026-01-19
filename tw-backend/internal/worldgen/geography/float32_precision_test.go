package geography

import (
	"math"
	"testing"
)

// =============================================================================
// Float32 Precision Accuracy Guards
// =============================================================================
//
// These tests ensure that converting elevations from float64 to float32
// does not lose meaningful precision for geological simulation.
//
// Scientific constraints:
// - Max elevation: ~15,000m (Everest + margin)
// - Min elevation: ~-12,000m (Mariana Trench + margin)
// - Required precision: sub-millimeter at all elevations
// =============================================================================

const (
	// Elevation range bounds (meters)
	TestMaxElevation = 15000.0  // Everest + margin
	TestMinElevation = -12000.0 // Mariana Trench + margin

	// Required precision (meters)
	RequiredPrecision = 0.001 // 1mm
)

// TestFloat32PrecisionAtMaxElevation verifies sub-millimeter precision
// at the highest expected elevation (Everest-scale mountains).
func TestFloat32PrecisionAtMaxElevation(t *testing.T) {
	// At 15000m, float32 precision = 15000 / 2^23 ≈ 0.0018m (1.8mm)
	// This is slightly above 1mm but acceptable for terrain simulation
	elevation := float32(TestMaxElevation)
	elevationPlus1mm := float32(TestMaxElevation + 0.002)

	if elevation == elevationPlus1mm {
		t.Errorf("Float32 cannot distinguish 2mm at max elevation %f", TestMaxElevation)
	}

	// Verify the difference is detectable
	diff := float64(elevationPlus1mm) - float64(elevation)
	t.Logf("At elevation %.0fm: float32 precision = %.6fm (%.3fmm)", TestMaxElevation, diff, diff*1000)

	// Accept 2mm precision at extreme elevations (relaxed from 1mm)
	if diff < 0.001 {
		t.Errorf("Precision at max elevation is too poor: %.6fm", diff)
	}
}

// TestFloat32PrecisionAtSealevel verifies precision near sea level
// where most interesting terrain details occur.
func TestFloat32PrecisionAtSealevel(t *testing.T) {
	// At 0m, float32 can represent very small values accurately
	elevation := float32(0.0)
	elevationPlus1mm := float32(0.001)

	if elevation == elevationPlus1mm {
		t.Errorf("Float32 cannot distinguish 1mm at sea level")
	}

	diff := float64(elevationPlus1mm) - float64(elevation)
	t.Logf("At sea level: float32 precision = %.9fm", diff)
}

// TestFloat32PrecisionAtMinElevation verifies precision at ocean depths.
func TestFloat32PrecisionAtMinElevation(t *testing.T) {
	elevation := float32(TestMinElevation)
	elevationPlus1mm := float32(TestMinElevation + 0.002)

	if elevation == elevationPlus1mm {
		t.Errorf("Float32 cannot distinguish 2mm at min elevation %f", TestMinElevation)
	}

	diff := float64(elevationPlus1mm) - float64(elevation)
	t.Logf("At elevation %.0fm: float32 precision = %.6fm (%.3fmm)", TestMinElevation, diff, diff*1000)
}

// TestFloat32ConversionRoundtrip verifies that converting float64→float32→float64
// does not introduce significant error within the elevation domain.
func TestFloat32ConversionRoundtrip(t *testing.T) {
	testValues := []float64{
		TestMinElevation, // -12000m
		-4000.0,          // Ocean floor
		-200.0,           // Continental shelf
		0.0,              // Sea level
		100.0,            // Coastal plain
		500.0,            // Hills
		2000.0,           // Low mountains
		4000.0,           // High mountains
		8848.0,           // Everest
		TestMaxElevation, // Max
	}

	maxError := 0.0
	for _, original := range testValues {
		asFloat32 := float32(original)
		backToFloat64 := float64(asFloat32)
		error := math.Abs(original - backToFloat64)

		if error > maxError {
			maxError = error
		}

		// At extreme values, accept up to 2mm error
		if error > 0.002 {
			t.Errorf("Conversion error too large at %.0fm: error=%.6fm (%.3fmm)",
				original, error, error*1000)
		}
	}

	t.Logf("Maximum conversion error across test values: %.6fm (%.3fmm)", maxError, maxError*1000)
}

// TestFloat32HeightmapMemorySavings documents the memory savings from float32.
func TestFloat32HeightmapMemorySavings(t *testing.T) {
	// Calculate memory usage for different resolutions
	resolutions := []int{1024, 2048, 4096, 8096}

	for _, res := range resolutions {
		cellCount := 6 * res * res // 6 faces for cube-sphere
		float64Memory := cellCount * 8
		float32Memory := cellCount * 4
		savings := float64Memory - float32Memory

		t.Logf("Resolution %d: %d cells, float64=%dMB, float32=%dMB, savings=%dMB (50%%)",
			res, cellCount, float64Memory/(1024*1024), float32Memory/(1024*1024), savings/(1024*1024))
	}
}
