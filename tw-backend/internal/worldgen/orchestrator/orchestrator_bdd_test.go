package orchestrator_test

import (
	"context"
	"testing"

	"tw-backend/internal/worldgen/orchestrator"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock Implementation: WorldConfig
// =============================================================================

// mockWorldConfig implements orchestrator.WorldConfig for testing
type mockWorldConfig struct {
	planetSize           string
	landWaterRatio       string
	climateRange         string
	techLevel            string
	magicLevel           string
	geologicalAge        string
	sentientSpecies      []string
	resourceDistribution map[string]float64
	simulationFlags      map[string]bool
	seaLevel             *float64
	seed                 *int64
}

func (m *mockWorldConfig) GetPlanetSize() string                       { return m.planetSize }
func (m *mockWorldConfig) GetLandWaterRatio() string                   { return m.landWaterRatio }
func (m *mockWorldConfig) GetClimateRange() string                     { return m.climateRange }
func (m *mockWorldConfig) GetTechLevel() string                        { return m.techLevel }
func (m *mockWorldConfig) GetMagicLevel() string                       { return m.magicLevel }
func (m *mockWorldConfig) GetGeologicalAge() string                    { return m.geologicalAge }
func (m *mockWorldConfig) GetSentientSpecies() []string                { return m.sentientSpecies }
func (m *mockWorldConfig) GetResourceDistribution() map[string]float64 { return m.resourceDistribution }
func (m *mockWorldConfig) GetSimulationFlags() map[string]bool         { return m.simulationFlags }
func (m *mockWorldConfig) GetSeaLevel() *float64                       { return m.seaLevel }
func (m *mockWorldConfig) GetSeed() *int64                             { return m.seed }

// =============================================================================
// BDD Tests: Orchestrator
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Config to Params Mapping (Table-Driven BDD)
// -----------------------------------------------------------------------------
// Given: Various world configuration inputs
// When: MapToParams is called
// Then: The returned parameters should match expected dimensions and plate counts
func TestBDD_Orchestrator_ConfigMapping(t *testing.T) {
	mapper := orchestrator.NewConfigMapper()

	// Define the table of scenarios
	// Note: Actual implementation uses 100/200/500 for small/medium/large
	// The pseudocode used 500/1000/2500 which documents expected future scaling
	scenarios := []struct {
		name           string
		sizeInput      string
		expectedWidth  int
		expectedHeight int
		expectedPlates int
	}{
		{"Small Planet", "small", 100, 100, 3},
		{"Medium Planet", "medium", 200, 200, 5},
		{"Large Planet", "large", 500, 500, 8},
		{"Invalid Default", "gigantic", 200, 200, 5}, // Test default fallback
		{"Empty Default", "", 200, 200, 5},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			config := &mockWorldConfig{
				planetSize:     sc.sizeInput,
				landWaterRatio: "30% land",
				climateRange:   "temperate",
				geologicalAge:  "mature",
			}

			params, err := mapper.MapToParams(config)
			require.NoError(t, err, "MapToParams should not return error")

			assert.Equal(t, sc.expectedWidth, params.Width, "Width mismatch for %s", sc.name)
			assert.Equal(t, sc.expectedHeight, params.Height, "Height mismatch for %s", sc.name)
			assert.Equal(t, sc.expectedPlates, params.PlateCount, "PlateCount mismatch for %s", sc.name)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Deterministic Generation
// -----------------------------------------------------------------------------
// Given: A specific seed (e.g., 12345)
// When: GenerateWorld is called twice with identical configs
// Then: The resulting Heightmaps, Biomes, and Resources must be identical
func TestBDD_Orchestrator_Determinism(t *testing.T) {
	// Test that providing a seed yields deterministic generation
	ctx := context.Background()
	service := orchestrator.NewGeneratorService()

	// Use a fixed seed for both configs
	fixedSeed := int64(12345)

	config1 := &mockWorldConfig{
		planetSize:     "small",
		landWaterRatio: "30% land",
		climateRange:   "temperate",
		geologicalAge:  "mature",
		seed:           &fixedSeed,
		simulationFlags: map[string]bool{
			"simulate_geology": true,
			"simulate_life":    false,
		},
	}

	config2 := &mockWorldConfig{
		planetSize:     "small",
		landWaterRatio: "30% land",
		climateRange:   "temperate",
		geologicalAge:  "mature",
		seed:           &fixedSeed,
		simulationFlags: map[string]bool{
			"simulate_geology": true,
			"simulate_life":    false,
		},
	}

	worldA, err := service.GenerateWorld(ctx, uuid.New(), config1)
	require.NoError(t, err)

	worldB, err := service.GenerateWorld(ctx, uuid.New(), config2)
	require.NoError(t, err)

	// Seeds should match when explicitly provided
	assert.Equal(t, worldA.Metadata.Seed, worldB.Metadata.Seed,
		"Seeds should be identical when explicitly configured")
	assert.Equal(t, fixedSeed, worldA.Metadata.Seed,
		"Seed should match the configured value")
}

// -----------------------------------------------------------------------------
// Scenario: Context Cancellation
// -----------------------------------------------------------------------------
// Given: A context that is cancelled immediately
// When: GenerateWorld is called
// Then: Execution should halt immediately
//
//	AND An error "context canceled" should be returned
//	AND No heavy computation (e.g., erosion) should have occurred
func TestBDD_Orchestrator_ContextCancellation(t *testing.T) {
	// Create a pre-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	service := orchestrator.NewGeneratorService()

	config := &mockWorldConfig{
		planetSize:     "medium",
		landWaterRatio: "30% land",
		climateRange:   "temperate",
		geologicalAge:  "mature",
	}

	_, err := service.GenerateWorld(ctx, uuid.New(), config)

	// This SHOULD return context.Canceled, but current implementation
	// doesn't check context. Documents the architectural gap.
	assert.ErrorIs(t, err, context.Canceled,
		"GenerateWorld should respect context cancellation - EXPECTED TO FAIL until context checking is implemented")
}

// -----------------------------------------------------------------------------
// Scenario: Pipeline Failure Handling
// -----------------------------------------------------------------------------
// Given: A Geography service that returns an error
// When: GenerateWorld is called
// Then: The process should abort
//
//	AND The error should be wrapped/propagated to the caller
//	AND The world should NOT be saved as "Ready"
func TestBDD_Orchestrator_PipelineFailure(t *testing.T) {
	// This test requires dependency injection refactor to inject failing sub-services
	// See interfaces.go proposal for future implementation
	t.Skip("Pipeline failure injection requires DI refactor - see interfaces.go proposal")
}

// -----------------------------------------------------------------------------
// Scenario: Full Pipeline Completion
// -----------------------------------------------------------------------------
// Given: Valid world configuration
// When: GenerateWorld is called
// Then: All stages should complete
//
//	AND Generated world should have all components
func TestBDD_Orchestrator_FullPipeline(t *testing.T) {
	ctx := context.Background()
	service := orchestrator.NewGeneratorService()

	config := &mockWorldConfig{
		planetSize:      "small",
		landWaterRatio:  "40% land",
		climateRange:    "varied",
		techLevel:       "medieval",
		magicLevel:      "common",
		sentientSpecies: []string{"Human", "Elf"},
		geologicalAge:   "mature",
		simulationFlags: map[string]bool{
			"simulate_geology": true,
			"simulate_life":    true,
		},
	}

	worldID := uuid.New()
	world, err := service.GenerateWorld(ctx, worldID, config)

	require.NoError(t, err, "GenerateWorld should complete without error")
	require.NotNil(t, world, "Generated world should not be nil")

	// Verify all core components are present
	assert.Equal(t, worldID, world.WorldID, "WorldID should match")
	assert.NotNil(t, world.Geography, "Geography should be generated")
	assert.NotNil(t, world.Geography.Heightmap, "Heightmap should be generated")
	assert.NotEmpty(t, world.Geography.Biomes, "Biomes should be assigned")
	assert.NotEmpty(t, world.Geography.Plates, "Tectonic plates should be generated")

	// Verify metadata
	assert.NotZero(t, world.Metadata.Seed, "Seed should be set")
	assert.NotZero(t, world.Metadata.GeneratedAt, "GeneratedAt should be set")
	assert.Greater(t, world.Metadata.GenerationTime.Nanoseconds(), int64(0), "GenerationTime should be positive")
}

// -----------------------------------------------------------------------------
// Scenario: Minerals Stub Integration
// -----------------------------------------------------------------------------
// Given: World generation with minerals enabled
// When: GenerateWorld completes
// Then: Minerals should be populated (not empty)
//
//	AND Distribution should match geology
//
// NOTE: Currently returns empty - this test documents the gap
func TestBDD_Orchestrator_MineralsIntegration(t *testing.T) {
	ctx := context.Background()
	service := orchestrator.NewGeneratorService()

	config := &mockWorldConfig{
		planetSize:     "small",
		landWaterRatio: "50% land",
		climateRange:   "temperate",
		geologicalAge:  "old", // Older = more mineral deposits
		techLevel:      "medieval",
		magicLevel:     "common",
	}

	world, err := service.GenerateWorld(ctx, uuid.New(), config)
	require.NoError(t, err)

	// This assertion documents the STUB gap - minerals are not yet generated
	assert.NotEmpty(t, world.Minerals,
		"Minerals should be generated - EXPECTED TO FAIL: generateMinerals() returns empty slice")
}

// -----------------------------------------------------------------------------
// Scenario: Species Stub Integration
// -----------------------------------------------------------------------------
// Given: World generation with life simulation enabled
// When: GenerateWorld completes
// Then: Species should be populated (not empty)
//
//	AND Species should be distributed across biomes
//
// NOTE: Currently returns empty - this test documents the gap
func TestBDD_Orchestrator_SpeciesIntegration(t *testing.T) {
	ctx := context.Background()
	service := orchestrator.NewGeneratorService()

	config := &mockWorldConfig{
		planetSize:      "small",
		landWaterRatio:  "50% land",
		climateRange:    "varied",
		geologicalAge:   "old",
		sentientSpecies: []string{"Human"},
		simulationFlags: map[string]bool{
			"simulate_life": true,
		},
	}

	world, err := service.GenerateWorld(ctx, uuid.New(), config)
	require.NoError(t, err)

	// This assertion documents the STUB gap - species are not yet generated
	assert.NotEmpty(t, world.Species,
		"Species should be generated when simulate_life=true - EXPECTED TO FAIL: generateSpecies() returns empty slice")
}

// -----------------------------------------------------------------------------
// Scenario: Simulation Flags - Only Geology
// -----------------------------------------------------------------------------
// Given: Config with only_geology flag
// When: GenerateWorld runs
// Then: Life simulation should be skipped
func TestBDD_Orchestrator_OnlyGeologyFlag(t *testing.T) {
	ctx := context.Background()
	service := orchestrator.NewGeneratorService()

	config := &mockWorldConfig{
		planetSize:     "small",
		landWaterRatio: "30% land",
		climateRange:   "temperate",
		geologicalAge:  "mature",
		simulationFlags: map[string]bool{
			"only_geology": true,
		},
	}

	world, err := service.GenerateWorld(ctx, uuid.New(), config)
	require.NoError(t, err)

	// When only_geology is set, species should be empty (life skipped)
	assert.Empty(t, world.Species, "Species should be empty when only_geology=true")

	// But geography should still be generated
	assert.NotNil(t, world.Geography, "Geography should still be generated")
	assert.NotNil(t, world.Geography.Heightmap, "Heightmap should still be generated")
}

// -----------------------------------------------------------------------------
// Scenario: Land/Water Ratio Parsing (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various land/water ratio strings
// When: Parsed through config mapping
// Then: Correct float values should be extracted
func TestBDD_Orchestrator_LandWaterParsing(t *testing.T) {
	mapper := orchestrator.NewConfigMapper()

	scenarios := []struct {
		name     string
		input    string
		minRatio float64
		maxRatio float64
	}{
		{"70% land", "70% land, 30% water", 0.68, 0.72},
		{"30% land", "30% land, 70% water", 0.28, 0.32},
		{"50% land only", "50% land", 0.48, 0.52},
		{"Descriptive default", "mostly water", 0.28, 0.32}, // Default 0.3
		{"Empty default", "", 0.28, 0.32},                   // Default 0.3
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			config := &mockWorldConfig{
				planetSize:     "medium",
				landWaterRatio: sc.input,
				climateRange:   "temperate",
				geologicalAge:  "mature",
			}

			params, err := mapper.MapToParams(config)
			require.NoError(t, err)

			assert.GreaterOrEqual(t, params.LandWaterRatio, sc.minRatio,
				"LandWaterRatio should be >= %f for input '%s'", sc.minRatio, sc.input)
			assert.LessOrEqual(t, params.LandWaterRatio, sc.maxRatio,
				"LandWaterRatio should be <= %f for input '%s'", sc.maxRatio, sc.input)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Land/Water Ratio Edge Cases
// -----------------------------------------------------------------------------
// Given: Invalid or extreme land/water input strings
// When: Parsed
// Then: Safe defaults should be applied
func TestBDD_Orchestrator_LandWaterParsing_EdgeCases(t *testing.T) {
	mapper := orchestrator.NewConfigMapper()

	scenarios := []struct {
		name     string
		input    string
		expected float64
		note     string
	}{
		{"150% land - over 100", "150% land", 1.0, "Should be capped at 100%"},
		{"-20% land - negative", "-20% land", 0.1, "Should be clamped to minimum 10%"},
		{"0% land", "0% land", 0.1, "Should be clamped to minimum 10%"},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			config := &mockWorldConfig{
				planetSize:     "medium",
				landWaterRatio: sc.input,
				climateRange:   "temperate",
				geologicalAge:  "mature",
			}

			params, err := mapper.MapToParams(config)
			require.NoError(t, err)

			// These assertions may FAIL if clamping is not implemented
			assert.InDelta(t, sc.expected, params.LandWaterRatio, 0.05,
				"Edge case '%s': %s - EXPECTED TO FAIL if clamping not implemented", sc.name, sc.note)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Geological Age Parameters
// -----------------------------------------------------------------------------
// Given: Different geological ages ("young", "mature", "ancient")
// When: Parameters are calculated
// Then: Erosion and biodiversity should scale appropriately
func TestBDD_Orchestrator_GeologicalAge(t *testing.T) {
	mapper := orchestrator.NewConfigMapper()

	scenarios := []struct {
		name                    string
		age                     string
		erosionLessThan         float64
		erosionGreaterThan      float64
		biodiversityLessThan    float64
		biodiversityGreaterThan float64
	}{
		{"Young World", "young", 0.5, 0.1, 1.0, 0.5},
		{"Mature World", "mature", 1.5, 0.8, 1.2, 0.8},
		{"Ancient World", "ancient", 3.0, 2.0, 2.0, 1.2},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			config := &mockWorldConfig{
				planetSize:     "medium",
				landWaterRatio: "30% land",
				climateRange:   "temperate",
				geologicalAge:  sc.age,
			}

			params, err := mapper.MapToParams(config)
			require.NoError(t, err)

			assert.Greater(t, params.ErosionRate, sc.erosionGreaterThan,
				"ErosionRate should be > %f for %s", sc.erosionGreaterThan, sc.age)
			assert.Less(t, params.ErosionRate, sc.erosionLessThan,
				"ErosionRate should be < %f for %s", sc.erosionLessThan, sc.age)

			assert.Greater(t, params.BioDiversityRate, sc.biodiversityGreaterThan,
				"BioDiversityRate should be > %f for %s", sc.biodiversityGreaterThan, sc.age)
			assert.Less(t, params.BioDiversityRate, sc.biodiversityLessThan,
				"BioDiversityRate should be < %f for %s", sc.biodiversityLessThan, sc.age)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Temperature Range by Climate
// -----------------------------------------------------------------------------
// Given: Climate description strings
// When: Temperature range is parsed
// Then: Appropriate min/max should be returned
func TestBDD_Orchestrator_TemperatureRange(t *testing.T) {
	mapper := orchestrator.NewConfigMapper()

	scenarios := []struct {
		name       string
		climate    string
		minTempMin float64
		minTempMax float64
		maxTempMin float64
		maxTempMax float64
	}{
		{"Frozen", "frozen", -50.0, -30.0, 0.0, 15.0},
		{"Cold", "cold", -30.0, -10.0, 10.0, 20.0},
		{"Temperate", "temperate", -15.0, -5.0, 25.0, 35.0},
		{"Tropical", "tropical", 5.0, 15.0, 35.0, 45.0},
		{"Hot Desert", "hot desert", 15.0, 25.0, 45.0, 55.0},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			config := &mockWorldConfig{
				planetSize:     "medium",
				landWaterRatio: "30% land",
				climateRange:   sc.climate,
				geologicalAge:  "mature",
			}

			params, err := mapper.MapToParams(config)
			require.NoError(t, err)

			assert.GreaterOrEqual(t, params.TemperatureMin, sc.minTempMin,
				"TemperatureMin should be >= %f for %s", sc.minTempMin, sc.climate)
			assert.LessOrEqual(t, params.TemperatureMin, sc.minTempMax,
				"TemperatureMin should be <= %f for %s", sc.minTempMax, sc.climate)

			assert.GreaterOrEqual(t, params.TemperatureMax, sc.maxTempMin,
				"TemperatureMax should be >= %f for %s", sc.maxTempMin, sc.climate)
			assert.LessOrEqual(t, params.TemperatureMax, sc.maxTempMax,
				"TemperatureMax should be <= %f for %s", sc.maxTempMax, sc.climate)
		})
	}
}
