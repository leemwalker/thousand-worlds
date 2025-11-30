package weather

import (
	"time"

	"mud-platform-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// WeatherType represents the current weather condition
type WeatherType string

const (
	WeatherClear  WeatherType = "clear"
	WeatherCloudy WeatherType = "cloudy"
	WeatherRain   WeatherType = "rain"
	WeatherSnow   WeatherType = "snow"
	WeatherStorm  WeatherType = "storm"
)

// Season represents the current season
type Season string

const (
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonFall   Season = "fall"
	SeasonWinter Season = "winter"
)

// SeasonalModifier returns the sunlight modifier for the given season
func (s Season) Modifier() float64 {
	switch s {
	case SeasonSummer:
		return 1.3
	case SeasonSpring, SeasonFall:
		return 1.0
	case SeasonWinter:
		return 0.7
	default:
		return 1.0
	}
}

// Wind represents wind direction and speed
type Wind struct {
	Direction float64 // Degrees (0 = North, 90 = East, etc.)
	Speed     float64 // m/s
}

// ToVector converts wind to a 2D vector
func (w Wind) ToVector() geography.Vector {
	// Convert to Cartesian (X = East, Y = North)
	// Direction 0 = North, so we need to adjust
	radians := (90 - w.Direction) * (3.14159265359 / 180)
	return geography.Vector{
		X: w.Speed * 3.14159265359 * radians, // Simplified for now
		Y: w.Speed * 3.14159265359 * radians,
	}
}

// WeatherState represents the weather at a specific location and time
type WeatherState struct {
	StateID       uuid.UUID
	CellID        uuid.UUID
	Timestamp     time.Time
	State         WeatherType
	Temperature   float64 // °C
	Precipitation float64 // mm/day
	Wind          Wind
	Humidity      float64 // 0-100%
	Visibility    float64 // km
}

// AtmosphericCell represents the type of atmospheric circulation cell
type AtmosphericCell string

const (
	CellHadley AtmosphericCell = "hadley"
	CellFerrel AtmosphericCell = "ferrel"
	CellPolar  AtmosphericCell = "polar"
)

// PressureSystem represents high or low pressure
type PressureSystem string

const (
	PressureHigh PressureSystem = "high"
	PressureLow  PressureSystem = "low"
)

// GeographyCell represents a cell in the geography grid
// This will be populated from the geography package
type GeographyCell struct {
	CellID      uuid.UUID
	Location    geography.Point
	Elevation   float64 // meters
	IsOcean     bool
	RiverWidth  float64 // meters (0 if no river)
	Temperature float64 // °C (base temperature)
}

// IsWater returns true if the cell is ocean or has a river
func (c *GeographyCell) IsWater() bool {
	return c.IsOcean || c.RiverWidth > 0
}
