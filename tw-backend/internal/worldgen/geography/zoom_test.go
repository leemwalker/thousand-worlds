package geography_test

import (
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Zoom System Unit Tests
// =============================================================================

const zoomTestSeed int64 = 54321

// -----------------------------------------------------------------------------
// Test: Lat/Lon to Sphere Conversion
// -----------------------------------------------------------------------------
func TestZoom_LatLonToSphere(t *testing.T) {
	testCases := []struct {
		name        string
		lat, lon    float64
		expectedY   float64 // Y is up, so north pole = 1
		description string
	}{
		{"North Pole", 90, 0, 1.0, "Y should be 1"},
		{"South Pole", -90, 0, -1.0, "Y should be -1"},
		{"Equator at 0°", 0, 0, 0.0, "Y should be 0"},
		{"Equator at 90°E", 0, 90, 0.0, "Y should be 0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			x, y, z := geography.LatLonToSphere(tc.lat, tc.lon)

			// Check unit vector
			mag := x*x + y*y + z*z
			assert.InDelta(t, 1.0, mag, 0.0001, "Should be unit vector")

			// Check expected Y
			assert.InDelta(t, tc.expectedY, y, 0.0001, tc.description)
		})
	}
}

// -----------------------------------------------------------------------------
// Test: Sphere to Lat/Lon Roundtrip
// -----------------------------------------------------------------------------
func TestZoom_LatLonRoundtrip(t *testing.T) {
	testCases := []struct {
		lat, lon float64
	}{
		{45, 120},
		{-30, -60},
		{0, 0},
		{89, 0},   // Near pole
		{0, 180},  // Date line
		{-89, 45}, // Near south pole
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			x, y, z := geography.LatLonToSphere(tc.lat, tc.lon)
			latBack, lonBack := geography.SphereToLatLon(x, y, z)

			assert.InDelta(t, tc.lat, latBack, 0.0001, "Latitude roundtrip")
			// Longitude has discontinuity at ±180, so use modular comparison
			lonDiff := tc.lon - lonBack
			if lonDiff > 180 {
				lonDiff -= 360
			} else if lonDiff < -180 {
				lonDiff += 360
			}
			assert.InDelta(t, 0, lonDiff, 0.0001, "Longitude roundtrip")
		})
	}
}

// -----------------------------------------------------------------------------
// Test: GlobalToMacro Returns Valid Coordinates
// -----------------------------------------------------------------------------
func TestZoom_GlobalToMacro(t *testing.T) {
	topology := spatial.NewCubeSphereTopology(64)

	testCases := []struct {
		name     string
		lat, lon float64
	}{
		{"Equator 0°", 0, 0},
		{"Equator 90°E", 0, 90},
		{"North Pole Region", 80, 0},
		{"South Pole Region", -80, 45},
		{"Mid-latitude", 45, -120},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loc := geography.GlobalToMacro(tc.lat, tc.lon, topology)

			// Face should be valid
			assert.GreaterOrEqual(t, loc.Face, 0)
			assert.Less(t, loc.Face, 6)

			// Cell coordinates should be within resolution
			assert.GreaterOrEqual(t, loc.X, 0)
			assert.Less(t, loc.X, 64)
			assert.GreaterOrEqual(t, loc.Y, 0)
			assert.Less(t, loc.Y, 64)

			// Offsets should be in [0, 1)
			assert.GreaterOrEqual(t, loc.U, 0.0)
			assert.Less(t, loc.U, 1.0)
			assert.GreaterOrEqual(t, loc.V, 0.0)
			assert.Less(t, loc.V, 1.0)
		})
	}
}

// -----------------------------------------------------------------------------
// Test: GenerateLocalSeed Determinism
// -----------------------------------------------------------------------------
func TestZoom_SeedDeterminism(t *testing.T) {
	seed1 := geography.GenerateLocalSeed(zoomTestSeed, 0, 10, 20, geography.LODLocal)
	seed2 := geography.GenerateLocalSeed(zoomTestSeed, 0, 10, 20, geography.LODLocal)

	assert.Equal(t, seed1, seed2, "Same inputs should produce same seed")
}

// -----------------------------------------------------------------------------
// Test: GenerateLocalSeed Uniqueness
// -----------------------------------------------------------------------------
func TestZoom_SeedUniqueness(t *testing.T) {
	seeds := make(map[int64]bool)

	// Generate seeds for different locations and LODs
	for face := 0; face < 6; face++ {
		for x := 0; x < 10; x++ {
			for y := 0; y < 10; y++ {
				for lod := geography.LODGlobal; lod <= geography.LODNPC; lod++ {
					seed := geography.GenerateLocalSeed(zoomTestSeed, face, x, y, lod)
					seeds[seed] = true
				}
			}
		}
	}

	// 6 faces * 10 * 10 * 6 LODs = 3600 combinations
	expectedCount := 6 * 10 * 10 * 6
	assert.Equal(t, expectedCount, len(seeds), "All seeds should be unique")
}

// -----------------------------------------------------------------------------
// Test: Different Global Seeds Produce Different Local Seeds
// -----------------------------------------------------------------------------
func TestZoom_DifferentGlobalSeeds(t *testing.T) {
	seed1 := geography.GenerateLocalSeed(12345, 0, 10, 20, geography.LODLocal)
	seed2 := geography.GenerateLocalSeed(67890, 0, 10, 20, geography.LODLocal)

	assert.NotEqual(t, seed1, seed2, "Different global seeds should produce different local seeds")
}

// -----------------------------------------------------------------------------
// Test: GetMicroNoiseConfig Returns Valid Configs
// -----------------------------------------------------------------------------
func TestZoom_MicroNoiseConfigs(t *testing.T) {
	biomes := []geography.BiomeType{
		geography.BiomeOcean,
		geography.BiomeMountain,
		geography.BiomeGrassland,
		geography.BiomeDesert,
		geography.BiomeRainforest,
		geography.BiomeTundra,
	}

	for _, biome := range biomes {
		t.Run(string(biome), func(t *testing.T) {
			config := geography.GetMicroNoiseConfig(biome)

			assert.Greater(t, config.Octaves, 0, "Octaves should be positive")
			assert.Greater(t, config.Frequency, 0.0, "Frequency should be positive")
			assert.Greater(t, config.Lacunarity, 1.0, "Lacunarity should be > 1")
			assert.Greater(t, config.Persistence, 0.0, "Persistence should be positive")
			assert.Less(t, config.Persistence, 1.0, "Persistence should be < 1")
			assert.GreaterOrEqual(t, config.WarpStrength, 0.0, "WarpStrength should be >= 0")
		})
	}
}

// -----------------------------------------------------------------------------
// Test: Mountain Biome Has Higher Roughness Than Plains
// -----------------------------------------------------------------------------
func TestZoom_BiomeRoughnessComparison(t *testing.T) {
	mountainConfig := geography.GetMicroNoiseConfig(geography.BiomeMountain)
	grasslandConfig := geography.GetMicroNoiseConfig(geography.BiomeGrassland)

	assert.Greater(t, mountainConfig.Persistence, grasslandConfig.Persistence,
		"Mountains should have higher roughness (persistence) than grassland")
	assert.Greater(t, mountainConfig.Octaves, grasslandConfig.Octaves,
		"Mountains should have more detail (octaves) than grassland")
}

// -----------------------------------------------------------------------------
// Test: GenerateMicroTerrain Produces Bounded Output
// -----------------------------------------------------------------------------
func TestZoom_MicroTerrainBounded(t *testing.T) {
	topology := spatial.NewCubeSphereTopology(64)
	loc := geography.GlobalToMacro(45, -120, topology)

	params := geography.MicroTerrainParams{
		GlobalSeed:    zoomTestSeed,
		Location:      loc,
		BaseElevation: 1000,
		Biome:         geography.BiomeMountain,
		LOD:           geography.LODLocal,
	}

	// Sample multiple times
	minVal := 1000.0
	maxVal := -1000.0

	for u := 0.0; u < 1.0; u += 0.1 {
		for v := 0.0; v < 1.0; v += 0.1 {
			params.Location.U = u
			params.Location.V = v
			offset := geography.GenerateMicroTerrain(params)

			if offset < minVal {
				minVal = offset
			}
			if offset > maxVal {
				maxVal = offset
			}
		}
	}

	// Should be bounded reasonably
	require.Less(t, minVal, 0.0, "Should have some negative offsets")
	require.Greater(t, maxVal, 0.0, "Should have some positive offsets")
	assert.Greater(t, minVal, -200.0, "Offsets should be bounded")
	assert.Less(t, maxVal, 200.0, "Offsets should be bounded")
}

// -----------------------------------------------------------------------------
// Test: MacroToGlobal Roundtrip
// -----------------------------------------------------------------------------
func TestZoom_MacroLocationRoundtrip(t *testing.T) {
	topology := spatial.NewCubeSphereTopology(64)

	testCases := []struct {
		lat, lon float64
	}{
		{45, 120},
		{-30, -60},
		{30, -90},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			// Convert to macro
			loc := geography.GlobalToMacro(tc.lat, tc.lon, topology)

			// Convert back
			latBack, lonBack := geography.MacroToGlobal(loc, topology)

			// Should be close (within cell size tolerance: ~5 degrees for 64 resolution)
			assert.InDelta(t, tc.lat, latBack, 5.0, "Latitude roundtrip")

			lonDiff := tc.lon - lonBack
			if lonDiff > 180 {
				lonDiff -= 360
			} else if lonDiff < -180 {
				lonDiff += 360
			}
			assert.InDelta(t, 0, lonDiff, 5.0, "Longitude roundtrip")
		})
	}
}
