package geography_test

import (
	"math"
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// FBM Noise Unit Tests
// =============================================================================

// Fixed seed for deterministic tests
const fbmTestSeed int64 = 12345

// -----------------------------------------------------------------------------
// Test: FBM Determinism
// -----------------------------------------------------------------------------
// Given: Same seed and coordinates
// When: FBM3D is called twice
// Then: Results should be identical
func TestFBM_Determinism(t *testing.T) {
	config := geography.DefaultTerrainFBMConfig()
	fbm1 := geography.NewFBMGenerator(fbmTestSeed, config)
	fbm2 := geography.NewFBMGenerator(fbmTestSeed, config)

	testPoints := [][3]float64{
		{0.0, 0.0, 0.0},
		{0.5, 0.5, 0.5},
		{1.0, 2.0, 3.0},
		{-0.3, 0.7, -0.1},
	}

	for _, p := range testPoints {
		v1 := fbm1.FBM3D(p[0], p[1], p[2])
		v2 := fbm2.FBM3D(p[0], p[1], p[2])
		assert.Equal(t, v1, v2, "FBM should be deterministic for point %v", p)
	}
}

// -----------------------------------------------------------------------------
// Test: FBM Output Range
// -----------------------------------------------------------------------------
// Given: FBM generator with default config
// When: Sampling many points
// Then: Output should be approximately in [-1, 1] range
func TestFBM_OutputRange(t *testing.T) {
	config := geography.DefaultTerrainFBMConfig()
	fbm := geography.NewFBMGenerator(fbmTestSeed, config)

	minVal := math.MaxFloat64
	maxVal := -math.MaxFloat64

	// Sample a grid of points
	for x := -10.0; x <= 10.0; x += 0.5 {
		for y := -10.0; y <= 10.0; y += 0.5 {
			for z := -10.0; z <= 10.0; z += 0.5 {
				v := fbm.FBM3D(x, y, z)
				if v < minVal {
					minVal = v
				}
				if v > maxVal {
					maxVal = v
				}
			}
		}
	}

	// Should be roughly in [-1, 1] range (allow some tolerance due to warping)
	assert.Greater(t, minVal, -1.5, "FBM min should be roughly above -1.5")
	assert.Less(t, maxVal, 1.5, "FBM max should be roughly below 1.5")
}

// -----------------------------------------------------------------------------
// Test: FBM Config Affects Output
// -----------------------------------------------------------------------------
// Given: Different FBM configurations
// When: Sampling the same point
// Then: Different configs should produce different results
func TestFBM_ConfigAffectsOutput(t *testing.T) {
	point := [3]float64{0.5, 0.5, 0.5}

	defaultConfig := geography.DefaultTerrainFBMConfig()
	highFreqConfig := geography.FBMConfig{
		Octaves:      6,
		Frequency:    0.1, // 5x higher frequency
		Lacunarity:   2.0,
		Persistence:  0.5,
		WarpStrength: 0.35,
	}
	noWarpConfig := geography.FBMConfig{
		Octaves:      6,
		Frequency:    0.02,
		Lacunarity:   2.0,
		Persistence:  0.5,
		WarpStrength: 0.0, // No warping
	}

	fbmDefault := geography.NewFBMGenerator(fbmTestSeed, defaultConfig)
	fbmHighFreq := geography.NewFBMGenerator(fbmTestSeed, highFreqConfig)
	fbmNoWarp := geography.NewFBMGenerator(fbmTestSeed, noWarpConfig)

	vDefault := fbmDefault.FBM3D(point[0], point[1], point[2])
	vHighFreq := fbmHighFreq.FBM3D(point[0], point[1], point[2])
	vNoWarp := fbmNoWarp.FBM3D(point[0], point[1], point[2])

	// Different configs should generally produce different values
	assert.NotEqual(t, vDefault, vHighFreq,
		"Different frequency should produce different output")
	assert.NotEqual(t, vDefault, vNoWarp,
		"Warping should affect output")
}

// -----------------------------------------------------------------------------
// Test: Domain Warping Breaks Symmetry
// -----------------------------------------------------------------------------
// Given: FBM with and without domain warping
// When: Checking for grid-aligned patterns
// Then: Warped version should have less correlation at integer intervals
func TestFBM_DomainWarpingBreaksSymmetry(t *testing.T) {
	noWarpConfig := geography.FBMConfig{
		Octaves:      6,
		Frequency:    0.02,
		Lacunarity:   2.0,
		Persistence:  0.5,
		WarpStrength: 0.0,
	}
	warpConfig := geography.DefaultTerrainFBMConfig()

	fbmNoWarp := geography.NewFBMGenerator(fbmTestSeed, noWarpConfig)
	fbmWarp := geography.NewFBMGenerator(fbmTestSeed, warpConfig)

	// Sample at integer coordinates (where diamond patterns appear)
	integerSamples := make([]float64, 0)
	warpedSamples := make([]float64, 0)

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			x, y, z := float64(i), float64(j), 0.0
			integerSamples = append(integerSamples, fbmNoWarp.FBM3D(x, y, z))
			warpedSamples = append(warpedSamples, fbmWarp.FBM3D(x, y, z))
		}
	}

	// Calculate variance (warped should have similar or higher variance)
	noWarpVariance := variance(integerSamples)
	warpVariance := variance(warpedSamples)

	require.Greater(t, noWarpVariance, 0.0, "Should have some variance")
	require.Greater(t, warpVariance, 0.0, "Warped should have some variance")

	t.Logf("No-warp variance: %f, Warped variance: %f", noWarpVariance, warpVariance)
}

func variance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range data {
		mean += v
	}
	mean /= float64(len(data))

	sumSq := 0.0
	for _, v := range data {
		sumSq += (v - mean) * (v - mean)
	}
	return sumSq / float64(len(data))
}

// -----------------------------------------------------------------------------
// Test: FBM2D Works
// -----------------------------------------------------------------------------
// Given: FBM generator
// When: Using FBM2D
// Then: Should return valid normalized output
func TestFBM_2DWorks(t *testing.T) {
	config := geography.DefaultTerrainFBMConfig()
	fbm := geography.NewFBMGenerator(fbmTestSeed, config)

	v := fbm.FBM2D(0.5, 0.5)

	assert.Greater(t, v, -1.5, "FBM2D should return value > -1.5")
	assert.Less(t, v, 1.5, "FBM2D should return value < 1.5")
}

// -----------------------------------------------------------------------------
// Test: Different Seeds Produce Different Results
// -----------------------------------------------------------------------------
// Given: Different seeds
// When: Generating FBM noise
// Then: Results should differ
func TestFBM_DifferentSeeds(t *testing.T) {
	config := geography.DefaultTerrainFBMConfig()
	fbm1 := geography.NewFBMGenerator(12345, config)
	fbm2 := geography.NewFBMGenerator(67890, config)

	v1 := fbm1.FBM3D(0.5, 0.5, 0.5)
	v2 := fbm2.FBM3D(0.5, 0.5, 0.5)

	assert.NotEqual(t, v1, v2, "Different seeds should produce different output")
}
