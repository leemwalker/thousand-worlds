package evolution

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateInitialSpecies(t *testing.T) {
	biomes := []string{"tropical", "temperate", "desert", "arctic"}

	species := GenerateInitialSpecies(biomes)

	// Should generate 5-10 species per biome
	assert.True(t, len(species) >= 20) // 4 biomes × 5 min
	assert.True(t, len(species) <= 40) // 4 biomes × 10 max

	// All species should have valid data
	for _, s := range species {
		assert.NotEmpty(t, s.Name)
		assert.True(t, s.Population > 0)
		assert.True(t, s.Size > 0)
		assert.True(t, len(s.PreferredBiomes) > 0)
	}
}

func TestSpeciesTypesDistribution(t *testing.T) {
	biomes := []string{"tropical"}
	species := GenerateInitialSpecies(biomes)

	floraCount := 0
	herbivoreCount := 0
	carnivoreCount := 0

	for _, s := range species {
		if s.IsFlora() {
			floraCount++
		} else if s.IsHerbivore() {
			herbivoreCount++
		} else if s.IsCarnivore() {
			carnivoreCount++
		}
	}

	// Should have a mix of all types
	assert.True(t, floraCount > 0)
	assert.True(t, herbivoreCount > 0)
	assert.True(t, carnivoreCount > 0)
}
