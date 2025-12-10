package orchestrator

import (
	"time"

	"mud-platform-backend/internal/worldgen/evolution"
	"mud-platform-backend/internal/worldgen/geography"
	"mud-platform-backend/internal/worldgen/minerals"
	"mud-platform-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// GeneratedWorld holds all procedurally generated content for a world
type GeneratedWorld struct {
	WorldID      uuid.UUID
	Geography    *geography.WorldMap
	Weather      []*weather.WeatherState
	WeatherCells []*weather.GeographyCell
	Minerals     []minerals.MineralDeposit
	Species      []*evolution.Species
	Metadata     GenerationMetadata
}

// GenerationMetadata tracks generation parameters and timing
type GenerationMetadata struct {
	Seed           int64
	GeneratedAt    time.Time
	GenerationTime time.Duration
	DimensionsX    int
	DimensionsY    int
	SeaLevel       float64
	LandRatio      float64
}

// GenerationParams contains all parameters needed for world generation
type GenerationParams struct {
	// Dimensions
	Width  int
	Height int

	// Geography parameters
	PlateCount     int
	LandWaterRatio float64 // 0.0 to 1.0 (0.3 = 30% land)

	// Climate parameters
	TemperatureMin   float64
	TemperatureMax   float64
	PrecipitationMin float64
	PrecipitationMax float64
	ErosionRate      float64 // Multiplier for erosion iterations

	// Resource parameters
	MineralDensity  float64 // 0.0 to 1.0
	ResourceWeights map[string]float64

	// Species parameters
	InitialSpeciesCount int
	BioDiversityRate    float64 // Multiplier for species diversity
	SpeciesTemplates    []evolution.Species

	// Random seed
	Seed int64
}
