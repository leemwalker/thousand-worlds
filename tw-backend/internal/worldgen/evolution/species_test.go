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
	// Use multiple biomes to ensure sufficient sample size for distribution test
	// With 4 biomes × 5-10 species = 20-40 species total
	// At 40% flora / 30% herbivore / 30% carnivore distribution,
	// getting zero of any type is extremely unlikely (<0.01%)
	biomes := []string{"tropical", "temperate", "desert", "arctic"}
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
	assert.True(t, floraCount > 0, "Should have at least one flora species")
	assert.True(t, herbivoreCount > 0, "Should have at least one herbivore species")
	assert.True(t, carnivoreCount > 0, "Should have at least one carnivore species")
}
