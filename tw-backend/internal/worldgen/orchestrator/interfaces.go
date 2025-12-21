package orchestrator

import "tw-backend/internal/worldgen/geography"

// =============================================================================
// Dependency Injection Interfaces
// =============================================================================

// GeographyGenerator generates terrain and biomes
type GeographyGenerator interface {
	GenerateGeography(params *GenerationParams) (*geography.WorldMap, float64, error)
}

// DefaultGeographyGenerator is the production implementation
type DefaultGeographyGenerator struct{}

// GenerateGeography creates terrain using geographic subsystem
func (g *DefaultGeographyGenerator) GenerateGeography(params *GenerationParams) (*geography.WorldMap, float64, error) {
	// Generate tectonic plates
	plates := geography.GeneratePlates(params.PlateCount, params.Width, params.Height, params.Seed)

	// Generate heightmap from tectonic activity
	heightmap := geography.GenerateHeightmap(params.Width, params.Height, plates, params.Seed, params.ErosionRate, params.RainfallFactor)

	// Assign ocean/land based on desired ratio
	seaLevel := geography.AssignOceanLand(heightmap, params.LandWaterRatio)

	// Generate rivers
	rivers := geography.GenerateRivers(heightmap, seaLevel, params.Seed)

	// Assign biomes
	biomes := geography.AssignBiomes(heightmap, seaLevel, params.Seed, 0.0)

	worldMap := &geography.WorldMap{
		Heightmap: heightmap,
		Plates:    plates,
		Biomes:    biomes,
		Rivers:    rivers,
	}

	return worldMap, seaLevel, nil
}
