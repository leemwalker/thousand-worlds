package geography

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClassifyBiome_PureClassification verifies that ClassifyBiome
// uses only temperature and moisture inputs, with no latitude math.
// This is the new decoupled API for the Weather→Biome causal chain.
func TestClassifyBiome_PureClassification(t *testing.T) {
	tests := []struct {
		name        string
		tempC       float64
		rainfallMM  float64
		drainage    float64
		elevation   float64
		seaLevel    float64
		wantBiome   BiomeType
		description string
	}{
		// Tropical biomes (hot + varying moisture)
		{
			name:        "tropical_rainforest",
			tempC:       28.0,
			rainfallMM:  2500.0,
			drainage:    0.5,
			elevation:   100.0,
			seaLevel:    0.0,
			wantBiome:   BiomeRainforest,
			description: "Hot + very wet → Rainforest",
		},
		{
			name:        "tropical_desert",
			tempC:       35.0,
			rainfallMM:  100.0,
			drainage:    0.8,
			elevation:   200.0,
			seaLevel:    0.0,
			wantBiome:   BiomeDesert,
			description: "Hot + very dry → Desert",
		},
		{
			name:        "tropical_grassland",
			tempC:       25.0,
			rainfallMM:  600.0,
			drainage:    0.5,
			elevation:   300.0,
			seaLevel:    0.0,
			wantBiome:   BiomeGrassland,
			description: "Hot + moderate moisture → Savanna/Grassland",
		},

		// Temperate biomes (moderate temp)
		{
			name:        "temperate_deciduous",
			tempC:       15.0,
			rainfallMM:  1400.0,
			drainage:    0.5,
			elevation:   200.0,
			seaLevel:    0.0,
			wantBiome:   BiomeDeciduousForest,
			description: "Moderate temp + wet → Deciduous Forest",
		},
		{
			name:        "temperate_grassland",
			tempC:       12.0,
			rainfallMM:  800.0, // 800/2000 = 0.4 moisture, within 0.3-0.6 range
			drainage:    0.5,
			elevation:   400.0,
			seaLevel:    0.0,
			wantBiome:   BiomeGrassland,
			description: "Moderate temp + moderate moisture → Grassland",
		},

		// Cold biomes
		{
			name:        "boreal_taiga",
			tempC:       -2.0,
			rainfallMM:  1400.0, // 1400/2000 = 0.7 moisture, above 0.6 threshold
			drainage:    0.4,
			elevation:   300.0,
			seaLevel:    0.0,
			wantBiome:   BiomeTaiga,
			description: "Cold + wet → Taiga/Boreal Forest",
		},
		{
			name:        "arctic_tundra",
			tempC:       -15.0,
			rainfallMM:  200.0,
			drainage:    0.3,
			elevation:   100.0,
			seaLevel:    0.0,
			wantBiome:   BiomeTundra,
			description: "Very cold + dry → Tundra",
		},

		// Elevation-based biomes
		{
			name:        "high_mountain_alpine",
			tempC:       -5.0,
			rainfallMM:  800.0,
			drainage:    0.7,
			elevation:   4000.0,
			seaLevel:    0.0,
			wantBiome:   BiomeAlpine,
			description: "Very high elevation → Alpine regardless of temp",
		},
		{
			name:        "ocean",
			tempC:       15.0,
			rainfallMM:  0.0,
			drainage:    0.0,
			elevation:   -500.0,
			seaLevel:    0.0,
			wantBiome:   BiomeOcean,
			description: "Below sea level → Ocean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyBiome(tt.tempC, tt.rainfallMM, tt.drainage, tt.elevation, tt.seaLevel)
			assert.Equal(t, tt.wantBiome, got, tt.description)
		})
	}
}

// TestClassifyBiome_NoLatitudeDependency verifies the function has no
// hidden dependency on coordinates or latitude.
func TestClassifyBiome_NoLatitudeDependency(t *testing.T) {
	// Same climate inputs should produce same biome regardless of any
	// imaginary "position" - the function only takes physical parameters.
	tempC := 25.0
	rainfallMM := 2000.0
	drainage := 0.5
	elevation := 100.0
	seaLevel := 0.0

	result1 := ClassifyBiome(tempC, rainfallMM, drainage, elevation, seaLevel)
	result2 := ClassifyBiome(tempC, rainfallMM, drainage, elevation, seaLevel)

	assert.Equal(t, result1, result2, "Same inputs must produce same output")
	assert.Equal(t, BiomeRainforest, result1, "Hot + wet should be rainforest")
}

// TestClassifyBiome_TemperatureShiftChangesResult proves the Weather→Biome
// causal link: changing temperature (from Weather) changes the biome.
func TestClassifyBiome_TemperatureShiftChangesResult(t *testing.T) {
	rainfallMM := 1000.0
	drainage := 0.5
	elevation := 200.0
	seaLevel := 0.0

	// Warm climate
	warmBiome := ClassifyBiome(25.0, rainfallMM, drainage, elevation, seaLevel)

	// Cold climate (same moisture)
	coldBiome := ClassifyBiome(-10.0, rainfallMM, drainage, elevation, seaLevel)

	assert.NotEqual(t, warmBiome, coldBiome,
		"Temperature change must shift biome classification")
	assert.Contains(t, []BiomeType{BiomeTaiga, BiomeTundra}, coldBiome,
		"Cold temperature should produce cold biome")
}
