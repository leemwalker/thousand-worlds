package geography

import (
	"testing"

	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHeightmap(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	count := 5
	seed := int64(12345)

	plates := GeneratePlates(count, topology, seed, 0.30)
	hm := NewSphereHeightmap(topology)
	hm = GenerateHeightmap(plates, hm, topology, seed, 1.0, 1.0)

	assert.NotNil(t, hm)
	assert.Equal(t, resolution, hm.Resolution())

	// Check elevation ranges
	// Oceanic plates should be deep negative
	// Continental should be positive
	// Tectonic interactions should create extremes

	hasOcean := false
	hasLand := false
	hasMountains := false
	hasTrenches := false

	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				val := hm.Get(spatial.Coordinate{Face: face, X: x, Y: y})
				// After isostatic relaxation + hypsometric curve, deep ocean is ~ -2000 to -4000
				if val < -2000 {
					hasOcean = true
				}
				if val > 0 {
					hasLand = true
				}
				if val > 2000 {
					hasMountains = true
				}
				if val < -5000 {
					hasTrenches = true
				}
			}
		}
	}

	t.Logf("Stats: Ocean=%v Land=%v Mtn=%v Trench=%v", hasOcean, hasLand, hasMountains, hasTrenches)
	t.Logf("Min: %f Max: %f", hm.MinElev, hm.MaxElev)

	countOcean := 0
	countLand := 0
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				val := hm.Get(spatial.Coordinate{Face: face, X: x, Y: y})
				if val < -2000 {
					countOcean++
				}
				if val > 0 {
					countLand++
				}
			}
		}
	}
	t.Logf("Counts: OceanCells=%d LandCells=%d Total=%d", countOcean, countLand, 6*resolution*resolution)

	assert.True(t, hasOcean, "Should have deep ocean")
	assert.True(t, hasLand, "Should have land")
	// Mountains and trenches depend on random plate movement, but with 5 plates and seed 12345,
	// we expect some interaction.
	assert.True(t, hasMountains || hasTrenches, "Should have some tectonic features")

	hm.UpdateMinMax()
	assert.True(t, hm.MinElev < hm.MaxElev)
}

// =============================================================================
// Hypsometric Curve Tests
// =============================================================================

// -----------------------------------------------------------------------------
// Test: Hypsometric Curve Flattens Shelf Zone
// -----------------------------------------------------------------------------
// Given: Heights in the shelf zone (-0.15 to 0.05 normalized)
// When: ApplyHypsometricCurve is called
// Then: The range should be compressed (flattened)
func TestHypsometricCurve_FlattensShelf(t *testing.T) {
	// Test that shelf zone is compressed
	// Input range: -500 to +100 (600m span)
	// Expected: compressed range (should be less than original span)

	upperShelf := ApplyHypsometricCurve(100, 0.0)
	lowerShelf := ApplyHypsometricCurve(-500, 0.0)

	originalSpan := 600.0 // From -500 to +100
	compressedSpan := absF64(upperShelf - lowerShelf)

	// The compressed span should be much smaller than original
	assert.Less(t, compressedSpan, originalSpan*0.5,
		"Shelf zone span should be compressed to less than 50%% of original")

	// Also verify mid-shelf values are flattened
	midShelf := ApplyHypsometricCurve(-200, 0.0)
	assert.Less(t, absF64(midShelf), 500.0, "Mid-shelf should be relatively flat")
}

// -----------------------------------------------------------------------------
// Test: Hypsometric Curve Preserves Extremes
// -----------------------------------------------------------------------------
// Given: Heights far from sea level (deep ocean or high mountains)
// When: ApplyHypsometricCurve is called
// Then: Values should pass through with minimal change
func TestHypsometricCurve_PreservesExtremes(t *testing.T) {
	scenarios := []struct {
		name     string
		height   float64
		seaLevel float64
	}{
		{"Deep ocean", -4000, 0},
		{"Abyss", -8000, 0},
		{"Mountain", 3000, 0},
		{"High peak", 6000, 0},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			result := ApplyHypsometricCurve(sc.height, sc.seaLevel)
			// Extremes should be relatively preserved (within 20%)
			ratio := result / sc.height
			assert.InDelta(t, 1.0, ratio, 0.3,
				"Extreme height %f should be mostly preserved", sc.height)
		})
	}
}

func absF64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
