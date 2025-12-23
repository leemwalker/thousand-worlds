package geography

import (
	"testing"
	"time"

	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
)

func TestGeographicRealism(t *testing.T) {
	// Integration test for the full pipeline using spherical topology
	resolution := 32
	seed := int64(time.Now().UnixNano())
	topology := spatial.NewCubeSphereTopology(resolution)

	// 1. Tectonics
	plates := GeneratePlates(10, topology, seed)
	assert.Equal(t, 10, len(plates))

	// 2. Heightmap
	sphereHm := NewSphereHeightmap(topology)
	sphereHm = GenerateHeightmap(plates, sphereHm, topology, seed, 1.0, 1.0)
	assert.NotNil(t, sphereHm)

	// Convert to flat for legacy compatibility tests
	width, height := 100, 100
	hm := NewHeightmap(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			face := (x / resolution) % 6
			fx := x % resolution
			fy := y % resolution
			if fx >= resolution {
				fx = resolution - 1
			}
			if fy >= resolution {
				fy = resolution - 1
			}
			coord := spatial.Coordinate{Face: face, X: fx, Y: fy}
			hm.Set(x, y, sphereHm.Get(coord))
		}
	}
	hm.MinElev, hm.MaxElev = sphereHm.MinMax()

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
	assert.InDelta(t, 0.3, ratio, 0.1, "Land ratio should be within 10% of target")

	// 4. Rivers
	rivers := GenerateRivers(hm, seaLevel, seed)
	if len(rivers) == 0 {
		t.Logf("Warning: No rivers generated with seed %d", seed)
	}

	// 5. Biomes
	biomes := AssignBiomes(hm, seaLevel, seed, 0.0)
	assert.Equal(t, width*height, len(biomes))

	// Check biome diversity
	biomeCounts := make(map[BiomeType]int)
	for _, b := range biomes {
		biomeCounts[b.Type]++
	}

	// Should have at least Ocean and some land biomes
	assert.True(t, biomeCounts[BiomeOcean] > 0)
	assert.True(t, len(biomeCounts) > 2, "Should have diverse biomes")
}

func TestVariation(t *testing.T) {
	// Generate two worlds with different seeds
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)

	seed1 := int64(1001)
	plates1 := GeneratePlates(5, topology, seed1)
	hm1 := NewSphereHeightmap(topology)
	hm1 = GenerateHeightmap(plates1, hm1, topology, seed1, 1.0, 1.0)

	seed2 := int64(1002)
	plates2 := GeneratePlates(5, topology, seed2)
	hm2 := NewSphereHeightmap(topology)
	hm2 = GenerateHeightmap(plates2, hm2, topology, seed2, 1.0, 1.0)

	// Compare centroids of first plate
	assert.NotEqual(t, plates1[0].Centroid, plates2[0].Centroid, "Different seeds should produce different plates")

	// Compare heightmaps - sample a few cells
	coord := spatial.Coordinate{Face: 0, X: 0, Y: 0}
	assert.NotEqual(t, hm1.Get(coord), hm2.Get(coord), "Different seeds should produce different heightmaps")
}
