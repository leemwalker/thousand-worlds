package orchestrator

import (
	"context"
	"fmt"
	"time"

	"tw-backend/internal/worldgen/evolution"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/minerals"
	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// GeneratorService orchestrates procedural world generation
type GeneratorService struct {
	mapper *ConfigMapper
	geoGen GeographyGenerator
}

// Option configures the GeneratorService
type Option func(*GeneratorService)

// WithGeographyGenerator sets a custom geography generator (for testing)
func WithGeographyGenerator(g GeographyGenerator) Option {
	return func(s *GeneratorService) {
		s.geoGen = g
	}
}

// NewGeneratorService creates a new generator service
func NewGeneratorService(opts ...Option) *GeneratorService {
	s := &GeneratorService{
		mapper: NewConfigMapper(),
		geoGen: &DefaultGeographyGenerator{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// GenerateWorld creates a complete procedurally generated world
func (s *GeneratorService) GenerateWorld(
	ctx context.Context,
	worldID uuid.UUID,
	config WorldConfig,
) (*GeneratedWorld, error) {
	// Check context before starting
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	startTime := time.Now()

	// 1. Map configuration to generation parameters
	params, err := s.mapper.MapToParams(config)
	if err != nil {
		return nil, fmt.Errorf("failed to map config to params: %w", err)
	}

	generated := &GeneratedWorld{
		WorldID: worldID,
		Metadata: GenerationMetadata{
			Seed:        params.Seed,
			GeneratedAt: startTime,
			DimensionsX: params.Width,
			DimensionsY: params.Height,
		},
	}

	// Check context before geography generation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 2. Generate geography
	if params.SeaLevelOverride != nil {
		params.LandWaterRatio = 1.0 - clamp(*params.SeaLevelOverride, 0.0, 1.0)
	}

	geoMap, seaLevel, err := s.geoGen.GenerateGeography(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate geography: %w", err)
	}
	generated.Geography = geoMap
	generated.Metadata.SeaLevel = seaLevel
	generated.Metadata.LandRatio = params.LandWaterRatio

	// Check context before weather generation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 3. Generate weather patterns
	// Only generate full weather simulation if geology simulation is enabled
	// If disabled, we might want a static simple weather or skip it
	if params.SimulateGeology {
		weatherStates, weatherCells, err := s.generateWeather(params, geoMap)
		if err != nil {
			return nil, fmt.Errorf("failed to generate weather: %w", err)
		}
		generated.Weather = weatherStates
		generated.WeatherCells = weatherCells
	} else {
		// Initialize empty weather for basic functionality
		generated.Weather = []*weather.WeatherState{}
		generated.WeatherCells = []*weather.GeographyCell{}
	}

	// Check context before mineral generation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 4. Generate mineral deposits
	mineralDeposits, err := s.generateMinerals(params, geoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to generate minerals: %w", err)
	}
	generated.Minerals = mineralDeposits

	// Check context before species generation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 5. Generate species
	if params.SimulateLife {
		species, err := s.generateSpecies(params, geoMap)
		if err != nil {
			return nil, fmt.Errorf("failed to generate species: %w", err)
		}
		generated.Species = species
	} else {
		generated.Species = []*evolution.Species{}
	}

	// Record generation time
	generated.Metadata.GenerationTime = time.Since(startTime)

	return generated, nil
}

// generateWeather creates weather patterns
func (s *GeneratorService) generateWeather(params *GenerationParams, geoMap *geography.WorldMap) ([]*weather.WeatherState, []*weather.GeographyCell, error) {
	cells := s.mapToGeographyCells(geoMap, params)

	// Initial weather update
	// Use Summer as start season? Or based on time?
	// For generation, let's just use Spring/Summer default
	states := weather.UpdateWeather(cells, time.Now(), weather.SeasonSpring)

	return states, cells, nil
}

// mapToGeographyCells converts world map to weather cells
func (s *GeneratorService) mapToGeographyCells(geoMap *geography.WorldMap, params *GenerationParams) []*weather.GeographyCell {
	width := geoMap.Heightmap.Width
	height := geoMap.Heightmap.Height
	cells := make([]*weather.GeographyCell, width*height)

	// Map rivers to grid for quick lookup
	riverMap := make(map[int]float64)
	for _, river := range geoMap.Rivers {
		for _, point := range river {
			idx := int(point.Y)*width + int(point.X)
			if idx >= 0 && idx < len(cells) {
				riverMap[idx] = 10.0 // Default river width 10m
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			elev := geoMap.Heightmap.Get(x, y)
			biome := geoMap.Biomes[idx]

			cells[idx] = &weather.GeographyCell{
				CellID:      uuid.New(),
				Location:    geography.Point{X: float64(x), Y: float64(y)},
				Elevation:   elev,
				IsOcean:     biome.Type == geography.BiomeOcean,
				RiverWidth:  riverMap[idx],
				Temperature: biome.Temperature,
			}
		}
	}

	return cells
}

// generateMinerals distributes mineral deposits based on geology
func (s *GeneratorService) generateMinerals(params *GenerationParams, geoMap *geography.WorldMap) ([]minerals.MineralDeposit, error) {
	deposits := []minerals.MineralDeposit{}

	// Generate hydrothermal deposits at plate boundaries
	// Extract boundary points from plate centroids (simplified approach)
	ridgePoints := []minerals.Point{}
	for _, plate := range geoMap.Plates {
		// Use the plate centroid as a representative boundary point
		// In a full spherical implementation, we'd trace actual boundaries
		sx := float64(geoMap.Heightmap.Width / 2)
		sy := float64(geoMap.Heightmap.Height / 2)
		if plate.Position.X != 0 || plate.Position.Y != 0 || plate.Position.Z != 0 {
			// Map 3D sphere position to 2D for minerals (simplified)
			sx = float64(plate.Centroid.Face*geoMap.Heightmap.Width/6 + plate.Centroid.X)
			sy = float64(plate.Centroid.Y)
		}
		ridgePoints = append(ridgePoints, minerals.Point{X: sx, Y: sy})
	}
	hydro := minerals.GenerateHydrothermalDeposits(ridgePoints)
	for _, d := range hydro {
		deposits = append(deposits, *d)
	}

	// Generate tool stone deposits
	tools := minerals.GenerateToolStoneDeposits(true, true)
	for _, d := range tools {
		deposits = append(deposits, *d)
	}

	return deposits, nil
}

// generateSpecies creates flora and fauna based on biomes
func (s *GeneratorService) generateSpecies(params *GenerationParams, geoMap *geography.WorldMap) ([]*evolution.Species, error) {
	// Collect unique biome names
	biomeSet := make(map[string]bool)
	for _, biome := range geoMap.Biomes {
		biomeSet[string(biome.Type)] = true
	}

	biomes := []string{}
	for b := range biomeSet {
		biomes = append(biomes, b)
	}

	// Generate species based on biome diversity
	species := evolution.GenerateInitialSpecies(biomes)

	return species, nil
}
