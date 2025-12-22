package weather

import (
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

// TestGenerateInitialClimate verifies the climate generator produces valid output.
func TestGenerateInitialClimate(t *testing.T) {
	// Create a simple heightmap
	hm := geography.NewHeightmap(10, 10)
	// Set some elevations: center higher, edges lower
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			// Simple elevation pattern: higher in center
			distFromCenter := float64((x-5)*(x-5) + (y-5)*(y-5))
			elev := 1000.0 - distFromCenter*20
			hm.Set(x, y, elev)
		}
	}

	seaLevel := 0.0
	seed := int64(12345)
	globalTempMod := 0.0

	climateData := GenerateInitialClimate(hm, seaLevel, seed, globalTempMod)

	assert.Len(t, climateData, 100, "Should have 10x10=100 climate cells")

	// Verify all values are within reasonable bounds
	for i, cd := range climateData {
		assert.GreaterOrEqual(t, cd.Temperature, -50.0, "Temp too low at %d", i)
		assert.LessOrEqual(t, cd.Temperature, 50.0, "Temp too high at %d", i)
		assert.GreaterOrEqual(t, cd.AnnualRainfall, 0.0, "Rainfall negative at %d", i)
		assert.LessOrEqual(t, cd.AnnualRainfall, 2000.0, "Rainfall too high at %d", i)
		assert.GreaterOrEqual(t, cd.Seasonality, 0.0, "Seasonality negative at %d", i)
		assert.LessOrEqual(t, cd.Seasonality, 1.0, "Seasonality too high at %d", i)
		assert.GreaterOrEqual(t, cd.SoilDrainage, 0.0, "Drainage negative at %d", i)
		assert.LessOrEqual(t, cd.SoilDrainage, 1.0, "Drainage too high at %d", i)
	}
}

// TestGenerateInitialClimate_LatitudeAffectsTemperature proves that latitude
// (y position) affects temperature - the physics model we moved from biomes.go.
func TestGenerateInitialClimate_LatitudeAffectsTemperature(t *testing.T) {
	hm := geography.NewHeightmap(10, 10)
	// Flat terrain at sea level
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			hm.Set(x, y, 100.0) // All land at 100m
		}
	}

	seaLevel := 0.0
	seed := int64(12345)
	globalTempMod := 0.0

	climateData := GenerateInitialClimate(hm, seaLevel, seed, globalTempMod)

	// Equator is at y=5 (center), poles at y=0 and y=9
	equatorClimate := GetClimateAt(climateData, 10, 5, 5)
	poleClimate := GetClimateAt(climateData, 10, 5, 0)

	assert.Greater(t, equatorClimate.Temperature, poleClimate.Temperature,
		"Equator should be warmer than poles")
}

// TestGenerateInitialClimate_GlobalTempModAffectsAll proves that volcanic winter
// or greenhouse effects shift all temperatures.
func TestGenerateInitialClimate_GlobalTempModAffectsAll(t *testing.T) {
	hm := geography.NewHeightmap(5, 5)
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			hm.Set(x, y, 100.0)
		}
	}

	seaLevel := 0.0
	seed := int64(12345)

	// Normal climate
	normalClimate := GenerateInitialClimate(hm, seaLevel, seed, 0.0)

	// Volcanic winter (-10°C)
	coldClimate := GenerateInitialClimate(hm, seaLevel, seed, -10.0)

	// All temperatures should be 10°C lower in volcanic winter
	for i := range normalClimate {
		expectedCold := normalClimate[i].Temperature - 10.0
		assert.InDelta(t, expectedCold, coldClimate[i].Temperature, 0.01,
			"Volcanic winter should reduce temp by exactly 10°C at cell %d", i)
	}
}

// TestCalculateTemperatureFromLatitude verifies the physics model.
func TestCalculateTemperatureFromLatitude(t *testing.T) {
	seaLevel := 0.0
	elevation := 0.0 // At sea level
	globalTempMod := 0.0

	// Equator (latitude = 0)
	equatorTemp := calculateTemperatureFromLatitude(0.0, elevation, seaLevel, globalTempMod)
	assert.InDelta(t, 30.0, equatorTemp, 0.1, "Equator should be ~30°C")

	// Pole (latitude = 1)
	poleTemp := calculateTemperatureFromLatitude(1.0, elevation, seaLevel, globalTempMod)
	assert.InDelta(t, -20.0, poleTemp, 0.1, "Pole should be ~-20°C")

	// Mid-latitude (latitude = 0.5)
	midTemp := calculateTemperatureFromLatitude(0.5, elevation, seaLevel, globalTempMod)
	assert.InDelta(t, 5.0, midTemp, 0.1, "Mid-latitude should be ~5°C")

	// High altitude at equator (3000m)
	highAltTemp := calculateTemperatureFromLatitude(0.0, 3000.0, seaLevel, globalTempMod)
	// 30 - (3000/1000)*6.5 = 30 - 19.5 = 10.5
	assert.InDelta(t, 10.5, highAltTemp, 0.1, "High altitude should reduce temp by lapse rate")
}
