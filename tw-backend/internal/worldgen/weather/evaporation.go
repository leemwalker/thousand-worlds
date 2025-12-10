package weather

import (
	"math"
)

// CalculateEvaporation calculates evaporation rate in mm/day
// Formula: evaporation = baseRate × temperatureFactor × waterProximity × sunlight
func CalculateEvaporation(
	temperature float64,
	waterProximity float64,
	latitude float64,
	season Season,
) float64 {
	baseRate := 5.0 // mm/day in optimal conditions

	// Temperature factor: max(0, (temperature - 5°C) / 30°C)
	tempFactor := CalculateTemperatureFactor(temperature)

	// Sunlight factor: cos(latitude) × seasonalModifier
	sunlight := CalculateSunlight(latitude, season)

	evaporation := baseRate * tempFactor * waterProximity * sunlight

	return math.Max(0, evaporation)
}

// CalculateTemperatureFactor returns the temperature contribution to evaporation
func CalculateTemperatureFactor(temperature float64) float64 {
	if temperature < 5.0 {
		return 0.0 // Below 5°C: no evaporation (ice)
	}

	factor := (temperature - 5.0) / 30.0
	if factor > 1.0 {
		return 1.0 // Cap at 1.0 for temperatures above 35°C
	}

	return factor
}

// CalculateWaterProximity determines water proximity factor
func CalculateWaterProximity(cell *GeographyCell, nearestWaterDistance float64) float64 {
	if cell.IsOcean {
		return 1.0
	}

	if cell.RiverWidth > 0 {
		// River cells: riverWidth / 100m
		proximity := cell.RiverWidth / 100.0
		if proximity > 1.0 {
			return 1.0
		}
		return proximity
	}

	// Adjacent to water: check distance
	if nearestWaterDistance < 1000 { // < 1km
		return 0.1
	}

	// Distant from water: > 10km
	if nearestWaterDistance > 10000 {
		return 0.01
	}

	// Linear interpolation between 1km and 10km
	return 0.1 - ((nearestWaterDistance-1000)/9000)*0.09
}

// CalculateSunlight calculates sunlight factor based on latitude and season
func CalculateSunlight(latitude float64, season Season) float64 {
	// Convert latitude to radians
	latRad := latitude * (math.Pi / 180.0)

	// Base sunlight from latitude
	baseSunlight := math.Cos(latRad)

	// Apply seasonal modifier
	seasonalMod := season.Modifier()

	sunlight := baseSunlight * seasonalMod

	// Ensure non-negative
	if sunlight < 0 {
		sunlight = 0
	}

	return sunlight
}
