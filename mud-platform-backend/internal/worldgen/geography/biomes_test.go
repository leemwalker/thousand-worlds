package geography

import (
	"testing"
)

func TestAssignBiomes(t *testing.T) {
	width, height := 10, 10
	hm := NewHeightmap(width, height)
	seed := int64(12345)

	// Set some elevations
	hm.Set(0, 0, -100) // Ocean
	hm.Set(5, 5, 500)  // Lowland/Highland (Equator)
	hm.Set(9, 9, 5000) // High Mountain (Pole)

	// 5,5 is center (Equator, Lat 0)
	// 9,9 is edge (Pole, Lat 1)

	biomes := AssignBiomes(hm, 0, seed)

	if biomes[0].Type != BiomeOcean {
		t.Errorf("Expected Ocean at 0,0, got %s", biomes[0].Type)
	}

	// 9,9 is High Mountain (Alpine)
	idx := 9*width + 9
	if biomes[idx].Type != BiomeAlpine {
		t.Errorf("Expected Alpine at 9,9, got %s", biomes[idx].Type)
	}

	// 5,5 is Lowland/Highland at Equator.
	// Moisture is random, but let's check it's a valid terrestrial biome
	idxCenter := 5*width + 5
	centerBiome := biomes[idxCenter].Type
	validCenter := centerBiome == BiomeRainforest || centerBiome == BiomeDesert || centerBiome == BiomeGrassland || centerBiome == BiomeDeciduousForest
	if !validCenter {
		t.Errorf("Expected terrestrial biome at 5,5, got %s", centerBiome)
	}
}
