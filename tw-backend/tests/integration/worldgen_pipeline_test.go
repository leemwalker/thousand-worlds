package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/internal/worldgen/orchestrator"
)

// simpleWorldConfig implements orchestrator.WorldConfig interface
type simpleWorldConfig struct {
	seed *int64
}

func (c *simpleWorldConfig) GetPlanetSize() string                       { return "medium" }
func (c *simpleWorldConfig) GetLandWaterRatio() string                   { return "30% land" }
func (c *simpleWorldConfig) GetClimateRange() string                     { return "temperate" }
func (c *simpleWorldConfig) GetTechLevel() string                        { return "medieval" }
func (c *simpleWorldConfig) GetMagicLevel() string                       { return "rare" }
func (c *simpleWorldConfig) GetGeologicalAge() string                    { return "mature" }
func (c *simpleWorldConfig) GetSentientSpecies() []string                { return []string{"human"} }
func (c *simpleWorldConfig) GetResourceDistribution() map[string]float64 { return nil }
func (c *simpleWorldConfig) GetSimulationFlags() map[string]bool         { return nil }
func (c *simpleWorldConfig) GetSeaLevel() *float64                       { return nil }
func (c *simpleWorldConfig) GetSeed() *int64                             { return c.seed }
func (c *simpleWorldConfig) GetNaturalSatellites() string                { return "one" }

// TestWorldGenerationPipeline verifies the end-to-end generation process.
func TestWorldGenerationPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Setup Service
	service := orchestrator.NewGeneratorService()

	// 2. Prepare Configuration
	seedVal := int64(123456789)
	config := &simpleWorldConfig{seed: &seedVal}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 3. Execute Generation
	t.Logf("Starting World Generation with seed: %d", seedVal)
	start := time.Now()
	world, err := service.GenerateWorld(ctx, uuid.New(), config)
	duration := time.Since(start)

	// 4. Verification
	require.NoError(t, err, "World generation should succeed")
	require.NotNil(t, world, "Generated world should not be nil")
	t.Logf("Generation completed in %v", duration)

	// A. Metadata
	assert.Equal(t, seedVal, world.Metadata.Seed)
	assert.Equal(t, 200, world.Metadata.DimensionsX) // Medium = 200 (default in Mapper)

	// B. Geography
	require.NotNil(t, world.Geography.Heightmap, "Heightmap should be present")
	hm := world.Geography.Heightmap
	assert.Less(t, hm.MinElev, hm.MaxElev, "Heightmap should have elevation variance")

	// Verify we have both Land and Ocean
	hasLand := false
	hasOcean := false

	// Check a sample of cells
	// Iterate flat 1D array
	for _, elev := range hm.Elevations {
		if elev > 0 {
			hasLand = true
		} else {
			hasOcean = true
		}
	}
	assert.True(t, hasLand, "World should have land")
	assert.True(t, hasOcean, "World should have ocean")

	// C. Weather
	// world.Weather is []*weather.WeatherState
	// world.WeatherCells is []*weather.GeographyCell
	if len(world.Weather) > 0 {
		// Check valid temperatures using direct access to WeatherState struct
		validTemps := true
		for _, wState := range world.Weather {
			if wState.Temperature < -100 || wState.Temperature > 100 {
				validTemps = false
				break
			}
		}
		assert.True(t, validTemps, "Temperatures should be in realistic range (-100 to 100 C)")
	} else {
		t.Log("Warning: No WeatherStates generated (might be optional or flag dependent)")
	}

	// D. Determinism Check
	// Run again with SAME seed
	world2, err2 := service.GenerateWorld(ctx, uuid.New(), config)
	require.NoError(t, err2)

	// Sampling equality
	hm2 := world2.Geography.Heightmap
	assert.Equal(t, hm.MinElev, hm2.MinElev, "Deterministic Min Elevation")
	assert.Equal(t, hm.MaxElev, hm2.MaxElev, "Deterministic Max Elevation")

	// Check the first 10 elevations
	for i := 0; i < 10; i++ {
		assert.Equal(t, hm.Elevations[i], hm2.Elevations[i], "Deterministic elevation at index %d", i)
	}
}

// Helper to allow getting arbitrary coords if needed, but for integration
// checking flat array logic is enough verification of data population.
func getElev(hm *simpleWorldConfig, x, y int) float64 {
	return 0 // Placeholder
}
