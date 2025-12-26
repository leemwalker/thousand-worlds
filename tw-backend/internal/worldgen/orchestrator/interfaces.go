package orchestrator

import (
	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/astronomy"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/ocean"
	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// =============================================================================
// Dependency Injection Interfaces
// =============================================================================

// GeographyGenerator generates terrain and biomes
type GeographyGenerator interface {
	GenerateGeography(params *GenerationParams, satellites []astronomy.Satellite) (*geography.WorldMap, float64, error)
}

// DefaultGeographyGenerator is the production implementation
type DefaultGeographyGenerator struct{}

// GenerateGeography creates terrain using geographic subsystem.
// Pipeline order: Tectonics → Topography → Ocean Currents → Weather/Climate → Biomes
// Satellites affect volcanic activity via tidal stress.
func (g *DefaultGeographyGenerator) GenerateGeography(params *GenerationParams, satellites []astronomy.Satellite) (*geography.WorldMap, float64, error) {
	// Create spherical topology for the world
	// Resolution = Height as the face grid size (faces are square)
	topology := spatial.NewCubeSphereTopology(params.Height)

	// Calculate tidal stress from satellites for volcanic activity
	tidalStress := astronomy.CalculateTidalStress(satellites)

	// 1. Generate tectonic plates on the sphere
	plates := geography.GeneratePlates(params.PlateCount, topology, params.Seed)

	// 2. Create sphere heightmap and apply tectonics (with tidal stress and heat for volcanism)
	// For world generation, use modern Earth heat baseline (1.0)
	sphereHeightmap := geography.NewSphereHeightmap(topology)
	sphereHeightmap = geography.GenerateHeightmapWithTidalStress(plates, sphereHeightmap, topology, params.Seed, params.ErosionRate, params.RainfallFactor, tidalStress, 1.0)

	// 3. Convert to flat heightmap for legacy consumers
	heightmap := sphereHeightmap.ToFlatHeightmap(params.Width, params.Height)

	// 4. Assign ocean/land based on desired ratio
	seaLevel := geography.AssignOceanLand(heightmap, params.LandWaterRatio)

	// 5. Generate rivers (before climate, rivers affect local moisture)
	rivers := geography.GenerateRivers(heightmap, seaLevel, params.Seed)

	// 6. NEW: Generate ocean currents and simulate thermodynamics
	oceanSys := ocean.NewSystem(topology, sphereHeightmap, seaLevel)
	windMap := ocean.CalculateGlobalWindVectors(topology, sphereHeightmap, seaLevel)
	oceanSys.GenerateSurfaceCurrents(windMap)
	oceanSys.InitializeTemperature()
	oceanSys.SimulateThermodynamics(50) // 50 iterations for Gulf Stream effect (balanced for performance)

	// 7. Generate climate data using spherical system (with ocean moderation)
	// For world generation, use modern Earth baseline (geothermalOffset = 0)
	// During simulation, this would be driven by ClimateDriver based on planetary age
	geothermalOffset := 0.0
	sphereClimate := weather.GenerateInitialClimateSpherical(
		sphereHeightmap, topology, seaLevel, params.Seed, 0.0, geothermalOffset,
	)
	weather.ApplyOceanModeration(sphereClimate, topology, oceanSys, sphereHeightmap, seaLevel)

	// Convert spherical climate to flat for legacy consumers
	climateData := convertSphereClimateToFlat(sphereClimate, topology, params.Width, params.Height)

	// 8. Assign biomes using climate data
	biomes := assignBiomesFromClimate(heightmap, seaLevel, climateData)

	worldMap := &geography.WorldMap{
		Heightmap: heightmap,
		Plates:    plates,
		Biomes:    biomes,
		Rivers:    rivers,
	}

	return worldMap, seaLevel, nil
}

// convertSphereClimateToFlat converts spherical climate data to flat equirectangular projection.
func convertSphereClimateToFlat(sphereClimate weather.SphereClimateMap, topology spatial.Topology, width, height int) []weather.ClimateData {
	flatClimate := make([]weather.ClimateData, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map flat coordinates to spherical coordinates
			// Same algorithm as SphereHeightmap.ToFlatHeightmap
			lon := (float64(x) / float64(width)) * 2 * 3.141592653589793
			lat := (0.5 - float64(y)/float64(height)) * 3.141592653589793

			// Convert lat/lon to 3D sphere coordinates
			cosLat := cosApprox(lat)
			sinLat := sinApprox(lat)
			cosLon := cosApprox(lon)
			sinLon := sinApprox(lon)

			sphereX := cosLat * cosLon
			sphereY := sinLat
			sphereZ := cosLat * sinLon

			// Map to cube-sphere coordinate
			coord := topology.FromVector(sphereX, sphereY, sphereZ)
			climate := weather.GetClimateAtSpherical(sphereClimate, topology.Resolution(), coord)

			flatClimate[y*width+x] = climate
		}
	}

	return flatClimate
}

// cosApprox provides cosine using Taylor series (matching geography package)
func cosApprox(x float64) float64 {
	const pi = 3.141592653589793
	const twoPi = 2 * pi
	for x > pi {
		x -= twoPi
	}
	for x < -pi {
		x += twoPi
	}
	x2 := x * x
	x4 := x2 * x2
	x6 := x4 * x2
	x8 := x6 * x2
	return 1 - x2/2 + x4/24 - x6/720 + x8/40320
}

// sinApprox provides sine using Taylor series
func sinApprox(x float64) float64 {
	const pi = 3.141592653589793
	const twoPi = 2 * pi
	for x > pi {
		x -= twoPi
	}
	for x < -pi {
		x += twoPi
	}
	x2 := x * x
	x3 := x2 * x
	x5 := x3 * x2
	x7 := x5 * x2
	x9 := x7 * x2
	return x - x3/6 + x5/120 - x7/5040 + x9/362880
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
