package orchestrator

import (
	"time"

	"tw-backend/internal/worldgen/astronomy"
	"tw-backend/internal/worldgen/evolution"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/minerals"
	"tw-backend/internal/worldgen/weather"

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
	Satellites   []astronomy.Satellite // Natural satellites (moons)
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
	RainfallFactor   float64 // Multiplier for hydraulic erosion (0.0 to 2.0)

	// Resource parameters
	MineralDensity  float64 // 0.0 to 1.0
	ResourceWeights map[string]float64

	// Species parameters
	InitialSpeciesCount int
	BioDiversityRate    float64 // Multiplier for species diversity

	SpeciesTemplates []evolution.Species

	// Simulation flags
	SimulateGeology  bool     // If true, catastrophes and geological shifts occur
	SimulateLife     bool     // If true, species are generated
	DisableDiseases  bool     // If true, no diseases are generated
	SeaLevelOverride *float64 // If non-nil, overrides the land/water ratio calc

	// Satellite configuration
	SatelliteConfig astronomy.SatelliteConfig

	// Random seed
	Seed int64
}
