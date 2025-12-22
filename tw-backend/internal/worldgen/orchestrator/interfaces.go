package orchestrator

import (
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// =============================================================================
// Dependency Injection Interfaces
// =============================================================================

// GeographyGenerator generates terrain and biomes
type GeographyGenerator interface {
	GenerateGeography(params *GenerationParams) (*geography.WorldMap, float64, error)
}

// DefaultGeographyGenerator is the production implementation
type DefaultGeographyGenerator struct{}

// GenerateGeography creates terrain using geographic subsystem.
// Pipeline order: Tectonics → Topography → Weather/Climate → Biomes
func (g *DefaultGeographyGenerator) GenerateGeography(params *GenerationParams) (*geography.WorldMap, float64, error) {
	// 1. Generate tectonic plates
	plates := geography.GeneratePlates(params.PlateCount, params.Width, params.Height, params.Seed)

	// 2. Generate heightmap from tectonic activity
	heightmap := geography.GenerateHeightmap(params.Width, params.Height, plates, params.Seed, params.ErosionRate, params.RainfallFactor)

	// 3. Assign ocean/land based on desired ratio
	seaLevel := geography.AssignOceanLand(heightmap, params.LandWaterRatio)

	// 4. Generate rivers (before climate, rivers affect local moisture)
	rivers := geography.GenerateRivers(heightmap, seaLevel, params.Seed)

	// 5. Generate climate data from Weather service (NEW!)
	// This moves latitude→temperature physics from biomes.go to weather package
	climateData := weather.GenerateInitialClimate(heightmap, seaLevel, params.Seed, 0.0)

	// 6. Assign biomes using climate data (NEW!)
	// Biomes now depend on Weather data, not latitude directly
	biomes := assignBiomesFromClimate(heightmap, seaLevel, climateData)

	worldMap := &geography.WorldMap{
		Heightmap: heightmap,
		Plates:    plates,
		Biomes:    biomes,
		Rivers:    rivers,
	}

	return worldMap, seaLevel, nil
}

// assignBiomesFromClimate creates biomes using pre-computed climate data.
// This is the new Weather→Biome causal chain.
func assignBiomesFromClimate(hm *geography.Heightmap, seaLevel float64, climateData []weather.ClimateData) []geography.Biome {
	biomes := make([]geography.Biome, hm.Width*hm.Height)

	for y := 0; y < hm.Height; y++ {
		for x := 0; x < hm.Width; x++ {
			idx := y*hm.Width + x
			elev := hm.Get(x, y)
			climate := weather.GetClimateAt(climateData, hm.Width, x, y)

			// Use the new pure classification function
			biomeType := geography.ClassifyBiome(
				climate.Temperature,
				climate.AnnualRainfall,
				climate.SoilDrainage,
				elev,
				seaLevel,
			)

			biomes[idx] = geography.Biome{
				BiomeID:       uuid.New(),
				Name:          string(biomeType),
				Type:          biomeType,
				Temperature:   climate.Temperature,
				Precipitation: climate.AnnualRainfall,
			}
		}
	}

	return biomes
}
