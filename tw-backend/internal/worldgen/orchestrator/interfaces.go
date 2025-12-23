package orchestrator

import (
	"tw-backend/internal/spatial"
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
	// Create spherical topology for the world
	// Resolution = Height as the face grid size (faces are square)
	topology := spatial.NewCubeSphereTopology(params.Height)

	// 1. Generate tectonic plates on the sphere
	plates := geography.GeneratePlates(params.PlateCount, topology, params.Seed)

	// 2. Create sphere heightmap and apply tectonics
	sphereHeightmap := geography.NewSphereHeightmap(topology)
	sphereHeightmap = geography.GenerateHeightmap(plates, sphereHeightmap, topology, params.Seed, params.ErosionRate, params.RainfallFactor)

	// 3. Convert to flat heightmap for legacy compatibility
	// TODO: Eventually migrate all consumers to SphereHeightmap
	heightmap := sphereToFlatHeightmap(sphereHeightmap, topology, params.Width, params.Height)

	// 4. Assign ocean/land based on desired ratio
	seaLevel := geography.AssignOceanLand(heightmap, params.LandWaterRatio)

	// 5. Generate rivers (before climate, rivers affect local moisture)
	rivers := geography.GenerateRivers(heightmap, seaLevel, params.Seed)

	// 6. Generate climate data from Weather service
	climateData := weather.GenerateInitialClimate(heightmap, seaLevel, params.Seed, 0.0)

	// 7. Assign biomes using climate data
	biomes := assignBiomesFromClimate(heightmap, seaLevel, climateData)

	worldMap := &geography.WorldMap{
		Heightmap: heightmap,
		Plates:    plates,
		Biomes:    biomes,
		Rivers:    rivers,
	}

	return worldMap, seaLevel, nil
}

// sphereToFlatHeightmap converts a SphereHeightmap to a flat Heightmap for legacy compatibility.
// Uses equirectangular projection from Face 0 (front face).
func sphereToFlatHeightmap(sphere *geography.SphereHeightmap, topology spatial.Topology, width, height int) *geography.Heightmap {
	flat := geography.NewHeightmap(width, height)
	resolution := topology.Resolution()

	// Simple projection: use Face 0 as the main view
	// This is a temporary bridge until all consumers migrate to SphereHeightmap
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map flat coordinates to sphere coordinate
			// Use modulo to wrap around the face grid
			face := (x / resolution) % 6
			fx := x % resolution
			fy := y % resolution

			if fx >= resolution {
				fx = resolution - 1
			}
			if fy >= resolution {
				fy = resolution - 1
			}

			coord := spatial.Coordinate{Face: face, X: fx, Y: fy}
			elev := sphere.Get(coord)
			flat.Set(x, y, elev)
		}
	}

	// Update min/max
	flat.MinElev = sphere.MinElev
	flat.MaxElev = sphere.MaxElev

	return flat
}

// assignBiomesFromClimate creates biomes using pre-computed climate data.
func assignBiomesFromClimate(hm *geography.Heightmap, seaLevel float64, climateData []weather.ClimateData) []geography.Biome {
	biomes := make([]geography.Biome, hm.Width*hm.Height)

	for y := 0; y < hm.Height; y++ {
		for x := 0; x < hm.Width; x++ {
			idx := y*hm.Width + x
			elev := hm.Get(x, y)
			climate := weather.GetClimateAt(climateData, hm.Width, x, y)

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
