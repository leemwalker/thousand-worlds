package weather

import (
	"math"
	"time"
)

// UpdateWeather updates weather for all cells (6-hour cycle)
func UpdateWeather(cells []*GeographyCell, currentTime time.Time, season Season) []*WeatherState {
	states := make([]*WeatherState, len(cells))

	for i, cell := range cells {
		// Calculate temperature with diurnal and seasonal variation
		temp := CalculateTemperature(cell, currentTime, season)

		// Calculate evaporation
		waterProximity := CalculateWaterProximity(cell, 0) // Simplified
		evap := CalculateEvaporation(temp, waterProximity, cell.Location.Y, season)

		// Calculate wind (assuming cell location Y is latitude)
		wind := CalculateWind(cell.Location.Y, cell.Location.X, season)

		// Calculate precipitation (simplified - would need upwind cells in full implementation)
		precip, moisture := CalculatePrecipitation(cell, []*GeographyCell{}, wind, evap*10) // Simplified

		// Determine weather state
		weatherType := DetermineWeatherState(temp, precip, moisture, wind.Speed)
		visibility := CalculateVisibility(weatherType)

		states[i] = &WeatherState{
			CellID:        cell.CellID,
			Timestamp:     currentTime,
			State:         weatherType,
			Temperature:   temp,
			Precipitation: precip,
			Wind:          wind,
			Humidity:      moisture,
			Visibility:    visibility,
		}
	}

	return states
}

// CalculateTemperature calculates temperature with diurnal and seasonal variation
func CalculateTemperature(cell *GeographyCell, currentTime time.Time, season Season) float64 {
	baseTemp := cell.Temperature

	// Seasonal variation
	seasonalDelta := GetSeasonalTemperatureDelta(season, cell.Location.Y)
	baseTemp += seasonalDelta

	// Diurnal variation (daily cycle)
	diurnalDelta := GetDiurnalTemperatureDelta(currentTime)
	baseTemp += diurnalDelta

	// Elevation lapse rate: -6.5°C per 1000m (Standard Atmosphere)
	elevationDelta := -(cell.Elevation / 1000.0) * 6.5
	baseTemp += elevationDelta

	return baseTemp
}

// GetSeasonalTemperatureDelta returns temperature change due to season
func GetSeasonalTemperatureDelta(season Season, latitude float64) float64 {
	// Mid-latitudes have largest seasonal swings (20-30°C)
	// Tropics have minimal variation (<5°C)
	// Polar regions have moderate variation (15-20°C)

	absLat := math.Abs(latitude)

	var maxSwing float64
	if absLat < 15 {
		maxSwing = 2.5 // Tropics: ±2.5°C
	} else if absLat < 60 {
		maxSwing = 15.0 // Mid-latitudes: ±15°C (30°C total swing)
	} else {
		maxSwing = 10.0 // Polar: ±10°C
	}

	switch season {
	case SeasonSummer:
		return maxSwing
	case SeasonWinter:
		return -maxSwing
	case SeasonSpring, SeasonFall:
		return 0
	default:
		return 0
	}
}

// GetDiurnalTemperatureDelta returns temperature change due to time of day
func GetDiurnalTemperatureDelta(currentTime time.Time) float64 {
	// Typical diurnal swing: ±5-10°C
	hour := currentTime.Hour()

	// Warmest at ~14:00, coolest at ~4:00
	// Use sine wave approximation
	hourRad := float64(hour-4) * (2 * math.Pi / 24)
	delta := math.Sin(hourRad) * 7.5 // ±7.5°C swing

	return delta
}
