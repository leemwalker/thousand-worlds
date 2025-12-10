package geography

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeographicRealism(t *testing.T) {
	// Integration test for the full pipeline
	width, height := 100, 100
	seed := int64(time.Now().UnixNano())

	// 1. Tectonics
	plates := GeneratePlates(10, width, height, seed)
	assert.Equal(t, 10, len(plates))

	// 2. Heightmap
	hm := GenerateHeightmap(width, height, plates, seed, 1.0)
	assert.NotNil(t, hm)

	// 3. Ocean/Land
	seaLevel := AssignOceanLand(hm, 0.3) // 30% land

	// Check actual ratio
	landCount := 0
	for _, elev := range hm.Elevations {
		if elev > seaLevel {
			landCount++
		}
	}
	ratio := float64(landCount) / float64(width*height)
	assert.InDelta(t, 0.3, ratio, 0.05, "Land ratio should be within 5% of target")

	// 4. Rivers
	rivers := GenerateRivers(hm, seaLevel, seed)
	// We expect some rivers
	if len(rivers) == 0 {
		t.Logf("Warning: No rivers generated with seed %d", seed)
	}

	// 5. Biomes
	biomes := AssignBiomes(hm, seaLevel, seed)
	assert.Equal(t, width*height, len(biomes))

	// Check biome diversity
	biomeCounts := make(map[BiomeType]int)
	for _, b := range biomes {
		biomeCounts[b.Type]++
	}

	// Should have at least Ocean and some land biomes
	assert.True(t, biomeCounts[BiomeOcean] > 0)
	assert.True(t, len(biomeCounts) > 3, "Should have diverse biomes")
}

func TestVariation(t *testing.T) {
	// Generate two worlds with different seeds
	width, height := 50, 50

	seed1 := int64(1001)
	plates1 := GeneratePlates(5, width, height, seed1)
	hm1 := GenerateHeightmap(width, height, plates1, seed1, 1.0)

	seed2 := int64(1002)
	plates2 := GeneratePlates(5, width, height, seed2)
	hm2 := GenerateHeightmap(width, height, plates2, seed2, 1.0)

	// Compare centroids of first plate
	assert.NotEqual(t, plates1[0].Centroid, plates2[0].Centroid, "Different seeds should produce different plates")

	// Compare heightmaps
	assert.NotEqual(t, hm1.Elevations[0], hm2.Elevations[0], "Different seeds should produce different heightmaps")
}
