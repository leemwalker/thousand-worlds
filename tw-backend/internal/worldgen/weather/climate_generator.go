package weather

import (
	"math"

	"tw-backend/internal/worldgen/geography"
)

// ClimateData represents computed climate for a grid cell.
// This is the interface between Weather and Biome systems.
// Biome classification should use only this data, not lat/lon.
type ClimateData struct {
	Temperature    float64 // Annual average °C
	AnnualRainfall float64 // mm/year
	Seasonality    float64 // 0-1, temperature variance (0=stable, 1=extreme)
	SoilDrainage   float64 // 0-1 (0=waterlogged, 1=well-drained)
}

// GenerateInitialClimate creates a climate map from geography data.
// This moves the "latitude → temperature" physics from the old biomes.go
// into the Weather service where it belongs.
//
// Parameters:
//   - heightmap: Terrain elevation data
//   - seaLevel: Current sea level in meters
//   - seed: Random seed for moisture noise
//   - globalTempMod: Global temperature modifier (e.g., volcanic winter = -10)
//
// Returns a slice of ClimateData, one per grid cell (row-major order).
func GenerateInitialClimate(heightmap *geography.Heightmap, seaLevel float64, seed int64, globalTempMod float64) []ClimateData {
	width := heightmap.Width
	height := heightmap.Height
	climateData := make([]ClimateData, width*height)

	// Use Perlin noise for moisture patterns (same as old biomes.go)
	noise := geography.NewPerlinGenerator(seed)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			elevation := heightmap.Get(x, y)

			// Calculate latitude factor: 0 at equator, 1 at poles
			// normalizedY: 0.0 (top) to 1.0 (bottom)
			// latitude: 0.0 (center/equator) to 1.0 (edges/poles)
			normalizedY := float64(y) / float64(height)
			latitude := math.Abs(normalizedY-0.5) * 2.0

			// Temperature from latitude and elevation
			temp := calculateTemperatureFromLatitude(latitude, elevation, seaLevel, globalTempMod)

			// Moisture from Perlin noise (same algorithm as old biomes.go)
			n := noise.Noise2D(float64(x)*0.05, float64(y)*0.05)
			moisture := (n + 1.0) / 2.0 // Normalize to 0-1

			// Convert moisture factor to rainfall (mm/year)
			// Using 2000mm as "very wet" baseline (matches ClassifyBiome)
			rainfall := moisture * 2000.0

			// Seasonality: higher at poles and in continental interiors
			// Simplified model based on latitude
			seasonality := latitude * 0.8 // 0.0 at equator, 0.8 at poles

			// Soil drainage: simplified model
			// Higher elevation = better drainage, ocean = 0
			drainage := 0.5 // Default moderate
			if elevation <= seaLevel {
				drainage = 0.0 // Ocean/flooded
			} else {
				altitudeAboveSea := elevation - seaLevel
				drainage = math.Min(1.0, 0.3+altitudeAboveSea/5000.0)
			}

			climateData[idx] = ClimateData{
				Temperature:    temp,
				AnnualRainfall: rainfall,
				Seasonality:    seasonality,
				SoilDrainage:   drainage,
			}
		}
	}

	return climateData
}

// calculateTemperatureFromLatitude computes temperature from geographic position.
// This is the physics model extracted from the old biomes.go.
func calculateTemperatureFromLatitude(latitude, elevation, seaLevel, globalTempMod float64) float64 {
	// Base temp at equator: 30°C
	// Base temp at poles: -20°C
	// Range = 50°C across 0-1 latitude
	baseTemp := 30.0 - (latitude * 50.0)

	// Lapse rate: -6.5°C per 1000m altitude
	altitudeAboveSea := elevation - seaLevel
	if altitudeAboveSea < 0 {
		altitudeAboveSea = 0
	}
	temp := baseTemp - (altitudeAboveSea/1000.0)*6.5

	// Apply global modifier (volcanic winter, greenhouse, etc.)
	temp += globalTempMod

	return temp
}

// GetClimateAt returns climate data for a specific cell from a pre-computed climate map.
func GetClimateAt(climateMap []ClimateData, width, x, y int) ClimateData {
	idx := y*width + x
	if idx < 0 || idx >= len(climateMap) {
		return ClimateData{} // Default empty climate
	}
	return climateMap[idx]
}
