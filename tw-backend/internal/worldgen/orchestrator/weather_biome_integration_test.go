package orchestrator

import (
	"testing"

	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/weather"

	"github.com/stretchr/testify/assert"
)

// TestBiomesDependOnWeather proves the Weather→Biome causal link:
// changing global temperature (e.g., volcanic winter) shifts biome distribution.
func TestBiomesDependOnWeather(t *testing.T) {
	// Create a flat heightmap (all land at same elevation)
	hm := geography.NewHeightmap(20, 20)
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			hm.Set(x, y, 500.0) // All land at 500m
		}
	}
	seaLevel := 0.0
	seed := int64(12345)

	// Generate climate with normal temperature
	normalClimate := weather.GenerateInitialClimate(hm, seaLevel, seed, 0.0)
	normalBiomes := assignBiomesFromClimate(hm, seaLevel, normalClimate)

	// Generate climate with volcanic winter (-20°C global shift)
	coldClimate := weather.GenerateInitialClimate(hm, seaLevel, seed, -20.0)
	coldBiomes := assignBiomesFromClimate(hm, seaLevel, coldClimate)

	// Count biomes in each scenario
	normalCounts := countBiomeTypes(normalBiomes)
	coldCounts := countBiomeTypes(coldBiomes)

	// Assert that distributions are different
	assert.NotEqual(t, normalCounts, coldCounts,
		"Volcanic winter should change biome distribution")

	// Specifically: cold climate should have more tundra/taiga, less rainforest/desert
	coldTundra := coldCounts[geography.BiomeTundra] + coldCounts[geography.BiomeTaiga]
	normalTundra := normalCounts[geography.BiomeTundra] + normalCounts[geography.BiomeTaiga]

	assert.Greater(t, coldTundra, normalTundra,
		"Volcanic winter should produce more cold biomes (tundra/taiga)")

	// Rainforest should decrease in cold climate
	assert.GreaterOrEqual(t, normalCounts[geography.BiomeRainforest], coldCounts[geography.BiomeRainforest],
		"Volcanic winter should reduce rainforest")

	t.Logf("Normal biomes: %v", normalCounts)
	t.Logf("Cold biomes:   %v", coldCounts)
}

// TestBiomesUseWeatherTemperature verifies biome temperature comes from weather, not latitude calc.
func TestBiomesUseWeatherTemperature(t *testing.T) {
	hm := geography.NewHeightmap(10, 10)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			hm.Set(x, y, 100.0)
		}
	}
	seaLevel := 0.0
	seed := int64(12345)

	// Generate with extremely hot modifier (+30°C)
	hotClimate := weather.GenerateInitialClimate(hm, seaLevel, seed, 30.0)
	hotBiomes := assignBiomesFromClimate(hm, seaLevel, hotClimate)

	// Even at "poles", temperature should be warm due to modifier
	// Pole is at y=0 (top edge)
	poleClimate := weather.GetClimateAt(hotClimate, 10, 5, 0)

	// Pole base temp = 30 - (1.0 * 50) = -20°C
	// With +30 modifier = 10°C (temperate, not frozen)
	assert.Greater(t, poleClimate.Temperature, 0.0,
		"Pole should be above freezing with +30°C modifier")

	// All biomes should reflect the modified temperature
	for i, biome := range hotBiomes {
		// With hot modifier, even poles should not be tundra
		// (unless we get dry desert at poles)
		if biome.Temperature > 10.0 {
			assert.NotEqual(t, geography.BiomeTundra, biome.Type,
				"Warm areas should not be tundra (at index %d)", i)
		}
	}
}

func countBiomeTypes(biomes []geography.Biome) map[geography.BiomeType]int {
	counts := make(map[geography.BiomeType]int)
	for _, b := range biomes {
		counts[b.Type]++
	}
	return counts
}
