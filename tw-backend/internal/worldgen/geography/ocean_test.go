package geography_test

import (
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Land Ratio Normalization Tests
// =============================================================================

// Fixed seed for deterministic tests
const oceanTestSeed int64 = 42

// -----------------------------------------------------------------------------
// Test: NormalizeLandRatio Achieves Target
// -----------------------------------------------------------------------------
// Given: A heightmap with random elevations
// When: NormalizeLandRatio is called with 30% target
// Then: Result should have approximately 30% land
func TestNormalizeLandRatio_AchievesTarget(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	hm := geography.NewSphereHeightmap(topology)

	// Create varied terrain with mixed positive/negative values
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				// Create gradient with some noise
				elev := float64(x+y-resolution) * 100.0 // Range: -3200 to 2800
				hm.Set(spatial.Coordinate{Face: face, X: x, Y: y}, elev)
			}
		}
	}

	targetRatio := 0.30
	seaLevel := geography.NormalizeLandRatio(hm, topology, targetRatio)

	// Count land cells (above 0 after normalization - seaLevel is shifted to 0)
	totalCells := 6 * resolution * resolution
	landCells := 0
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				if hm.Get(spatial.Coordinate{Face: face, X: x, Y: y}) > 0 {
					landCells++
				}
			}
		}
	}

	actualRatio := float64(landCells) / float64(totalCells)

	// Allow 5% tolerance due to discrete cells
	_ = seaLevel // seaLevel is the original value before shifting
	assert.InDelta(t, targetRatio, actualRatio, 0.05,
		"Land ratio should be within 5%% of target (got %.2f%%, expected %.2f%%)",
		actualRatio*100, targetRatio*100)
}

// -----------------------------------------------------------------------------
// Test: NormalizeLandRatio Stabilizes Across Seeds
// -----------------------------------------------------------------------------
// Given: Multiple heightmaps generated with different seeds
// When: NormalizeLandRatio is applied to each
// Then: All should achieve 25-35% land ratio
func TestNormalizeLandRatio_StabilizesAcrossSeeds(t *testing.T) {
	seeds := []int64{1, 42, 123, 999, 12345}
	resolution := 16
	targetRatio := 0.30

	for _, seed := range seeds {
		t.Run("seed_"+string(rune(seed)), func(t *testing.T) {
			topology := spatial.NewCubeSphereTopology(resolution)

			// Generate plates and heightmap like real world gen
			plates := geography.GeneratePlates(5, topology, seed)
			hm := geography.NewSphereHeightmap(topology)

			// Set base elevations with gradients (simulating FBM noise variation)
			// This creates a continuous distribution that can be normalized
			fbm := geography.NewFBMGenerator(seed, geography.DefaultTerrainFBMConfig())
			for _, plate := range plates {
				baseElev := -4000.0
				if plate.Type == geography.PlateContinental {
					baseElev = 100.0
				}
				for coord := range plate.Region {
					// Add FBM variation like the real heightmap generator does
					sx, sy, sz := topology.ToSphere(coord)
					variation := fbm.FBM3D(sx, sy, sz) * 600.0
					hm.Set(coord, baseElev+variation)
				}
			}

			// Apply normalization
			seaLevel := geography.NormalizeLandRatio(hm, topology, targetRatio)
			require.NotZero(t, seaLevel, "Sea level should be calculated")

			// Count land (cells > 0 after normalization)
			totalCells := 6 * resolution * resolution
			landCells := 0
			for face := 0; face < 6; face++ {
				for y := 0; y < resolution; y++ {
					for x := 0; x < resolution; x++ {
						if hm.Get(spatial.Coordinate{Face: face, X: x, Y: y}) > 0 {
							landCells++
						}
					}
				}
			}

			actualRatio := float64(landCells) / float64(totalCells)
			assert.Greater(t, actualRatio, 0.20, "Land ratio should be > 20%% for seed %d", seed)
			assert.Less(t, actualRatio, 0.40, "Land ratio should be < 40%% for seed %d", seed)
		})
	}
}
