package orchestrator

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWorldConfig implements WorldConfig for testing
type mockWorldConfig struct {
	planetSize           string
	landWaterRatio       string
	climateRange         string
	techLevel            string
	magicLevel           string
	sentientSpecies      []string
	resourceDistribution map[string]float64
}

func (m *mockWorldConfig) GetPlanetSize() string        { return m.planetSize }
func (m *mockWorldConfig) GetLandWaterRatio() string    { return m.landWaterRatio }
func (m *mockWorldConfig) GetClimateRange() string      { return m.climateRange }
func (m *mockWorldConfig) GetTechLevel() string         { return m.techLevel }
func (m *mockWorldConfig) GetMagicLevel() string        { return m.magicLevel }
func (m *mockWorldConfig) GetSentientSpecies() []string { return m.sentientSpecies }
func (m *mockWorldConfig) GetResourceDistribution() map[string]float64 {
	return m.resourceDistribution
}

func TestConfigMapper_MapToParams(t *testing.T) {
	mapper := NewConfigMapper()

	tests := []struct {
		name           string
		config         WorldConfig
		expectedWidth  int
		expectedHeight int
		expectedPlates int
		minLandRatio   float64
		maxLandRatio   float64
	}{
		{
			name: "Small planet",
			config: &mockWorldConfig{
				planetSize:     "small",
				landWaterRatio: "30% land, 70% water",
				climateRange:   "temperate",
			},
			expectedWidth:  100,
			expectedHeight: 100,
			expectedPlates: 3,
			minLandRatio:   0.25,
			maxLandRatio:   0.35,
		},
		{
			name: "Large planet",
			config: &mockWorldConfig{
				planetSize:     "large",
				landWaterRatio: "70% land, 30% water",
				climateRange:   "varied",
			},
			expectedWidth:  500,
			expectedHeight: 500,
			expectedPlates: 8,
			minLandRatio:   0.65,
			maxLandRatio:   0.75,
		},
		{
			name: "Medium planet (default)",
			config: &mockWorldConfig{
				planetSize:     "",
				landWaterRatio: "",
				climateRange:   "moderate",
			},
			expectedWidth:  200,
			expectedHeight: 200,
			expectedPlates: 5,
			minLandRatio:   0.25, // Default is 0.3, allow some tolerance
			maxLandRatio:   0.35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := mapper.MapToParams(tt.config)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedWidth, params.Width)
			assert.Equal(t, tt.expectedHeight, params.Height)
			assert.Equal(t, tt.expectedPlates, params.PlateCount)
			assert.GreaterOrEqual(t, params.LandWaterRatio, tt.minLandRatio)
			assert.LessOrEqual(t, params.LandWaterRatio, tt.maxLandRatio)
			assert.NotZero(t, params.Seed)
		})
	}
}

func TestGenerateWorld(t *testing.T) {
	ctx := context.Background()
	service := NewGeneratorService()

	config := &mockWorldConfig{
		planetSize:      "medium",
		landWaterRatio:  "40% land, 60% water",
		climateRange:    "varied",
		techLevel:       "medieval",
		magicLevel:      "common",
		sentientSpecies: []string{"Human", "Elf"},
	}

	worldID := uuid.New()

	generated, err := service.GenerateWorld(ctx, worldID, config)
	require.NoError(t, err)
	require.NotNil(t, generated)

	// Verify world ID
	assert.Equal(t, worldID, generated.WorldID)

	// Verify geography was generated
	assert.NotNil(t, generated.Geography)
	assert.NotNil(t, generated.Geography.Heightmap)
	assert.NotEmpty(t, generated.Geography.Plates)
	assert.NotEmpty(t, generated.Geography.Biomes)

	// Verify heightmap dimensions
	assert.Equal(t, 200, generated.Geography.Heightmap.Width)
	assert.Equal(t, 200, generated.Geography.Heightmap.Height)

	// Verify metadata
	assert.NotZero(t, generated.Metadata.Seed)
	assert.NotZero(t, generated.Metadata.GeneratedAt)
	assert.Greater(t, generated.Metadata.GenerationTime.Milliseconds(), int64(0))
	assert.Equal(t, 200, generated.Metadata.DimensionsX)
	assert.Equal(t, 200, generated.Metadata.DimensionsY)

	// Verify sea level is reasonable (can be deep for oceanic worlds)
	assert.Greater(t, generated.Metadata.SeaLevel, float64(-10000))
	assert.Less(t, generated.Metadata.SeaLevel, float64(10000))
}

func TestGenerateGeography(t *testing.T) {
	service := NewGeneratorService()

	params := &GenerationParams{
		Width:          100,
		Height:         100,
		PlateCount:     5,
		LandWaterRatio: 0.3,
		Seed:           12345,
	}

	geoMap, seaLevel, err := service.generateGeography(params)
	require.NoError(t, err)
	require.NotNil(t, geoMap)

	// Verify all components generated
	assert.NotNil(t, geoMap.Heightmap)
	assert.Len(t, geoMap.Plates, 5)
	assert.NotEmpty(t, geoMap.Biomes)
	// Rivers may be empty depending on terrain
	assert.NotNil(t, geoMap.Rivers)

	// Verify heightmap size
	assert.Equal(t, 100, geoMap.Heightmap.Width)
	assert.Equal(t, 100, geoMap.Heightmap.Height)

	// Verify sea level is set
	assert.NotZero(t, seaLevel)

	// Verify biomes match heightmap size
	expectedBiomeCount := 100 * 100
	assert.Equal(t, expectedBiomeCount, len(geoMap.Biomes))

	// Count land vs water biomes to verify land ratio
	landCount := 0
	for _, biome := range geoMap.Biomes {
		if biome.Type != "Ocean" {
			landCount++
		}
	}

	actualLandRatio := float64(landCount) / float64(expectedBiomeCount)
	// Allow 10% tolerance for land ratio
	assert.InDelta(t, 0.3, actualLandRatio, 0.1, "Land ratio should be approximately 30%")
}

func TestParseLandWaterRatio(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"70% land, 30% water", 0.7},
		{"30% land, 70% water", 0.3},
		{"50% land", 0.5},
		{"mostly land", 0.3}, // Default
		{"", 0.3},            // Default
		{"25% land", 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLandWaterRatio(tt.input)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestParseTemperatureRange(t *testing.T) {
	tests := []struct {
		input   string
		minTemp float64
		maxTemp float64
	}{
		{"frozen", -40.0, 10.0},
		{"cold", -20.0, 15.0},
		{"temperate", -10.0, 30.0},
		{"warm", 10.0, 40.0},
		{"hot", 20.0, 50.0},
		{"varied", -30.0, 45.0},
		{"", -10.0, 30.0}, // Default temperate
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			min, max := parseTemperatureRange(tt.input)
			assert.Equal(t, tt.minTemp, min)
			assert.Equal(t, tt.maxTemp, max)
		})
	}
}

func TestCalculateMineralDensity(t *testing.T) {
	tests := []struct {
		name        string
		techLevel   string
		magicLevel  string
		minDensity  float64
		maxDensity  float64
		description string
	}{
		{
			name:        "Primitive + High Magic",
			techLevel:   "stone age",
			magicLevel:  "dominant",
			minDensity:  0.8,
			maxDensity:  1.0,
			description: "Should have high resource density",
		},
		{
			name:        "Futuristic + No Magic",
			techLevel:   "futuristic",
			magicLevel:  "none",
			minDensity:  0.1,
			maxDensity:  0.4,
			description: "Should have lower resource density",
		},
		{
			name:        "Medieval + Common Magic",
			techLevel:   "medieval",
			magicLevel:  "common",
			minDensity:  0.5,
			maxDensity:  0.7,
			description: "Should have medium resource density",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			density := calculateMineralDensity(tt.techLevel, tt.magicLevel)
			assert.GreaterOrEqual(t, density, tt.minDensity, tt.description)
			assert.LessOrEqual(t, density, tt.maxDensity, tt.description)
		})
	}
}
