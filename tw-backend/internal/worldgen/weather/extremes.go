package weather

import (
	"time"
)

// ExtremeWeatherType represents types of extreme weather
type ExtremeWeatherType string

const (
	ExtremeHurricane ExtremeWeatherType = "hurricane"
	ExtremeBlizzard  ExtremeWeatherType = "blizzard"
	ExtremeDrought   ExtremeWeatherType = "drought"
	ExtremeHeatWave  ExtremeWeatherType = "heatwave"
)

// ExtremeWeatherEvent represents an ongoing extreme weather event
type ExtremeWeatherEvent struct {
	EventType     ExtremeWeatherType
	StartTime     time.Time
	Duration      time.Duration
	AffectedCells []string // Cell IDs
	Intensity     float64  // 0-1 scale
}

// CheckForHurricane determines if conditions are right for hurricane formation
func CheckForHurricane(cell *GeographyCell, temperature float64, latitude float64) bool {
	// Hurricanes form over warm ocean (>26°C) in tropics (5-20° latitude)
	absLat := latitude
	if absLat < 0 {
		absLat = -absLat
	}

	return cell.IsOcean &&
		temperature > 26.0 &&
		absLat >= 5 && absLat <= 20
}

// CheckForBlizzard determines if conditions are right for blizzard
func CheckForBlizzard(temperature float64, precipitation float64, windSpeed float64) bool {
	// Blizzards: cold (<-5°C), high precipitation, high winds
	return temperature < -5.0 &&
		precipitation > 10.0 &&
		windSpeed > 10.0
}

// CheckForDrought checks if a region is experiencing drought
// This would need historical precipitation data in a real implementation
func CheckForDrought(recentPrecipitation float64, normalPrecipitation float64, daysSinceLast int) bool {
	// Drought: <50% normal precipitation for 90+ days
	return recentPrecipitation < normalPrecipitation*0.5 && daysSinceLast > 90
}

// CheckForHeatWave checks for heat wave conditions
func CheckForHeatWave(temperature float64, normalTemperature float64, daysAboveNormal int) bool {
	// Heat wave: >10°C above normal for 7+ days
	return temperature > normalTemperature+10 && daysAboveNormal >= 7
}

// GenerateExtremeWeather evaluates conditions and generates extreme weather events
func GenerateExtremeWeather(
	cell *GeographyCell,
	state *WeatherState,
	latitude float64,
	recentHistory []WeatherState, // Last 90+ days
) *ExtremeWeatherEvent {

	// Check hurricane
	if CheckForHurricane(cell, state.Temperature, latitude) {
		return &ExtremeWeatherEvent{
			EventType:     ExtremeHurricane,
			StartTime:     state.Timestamp,
			Duration:      time.Hour * 24 * 7, // Typical duration: 1 week
			AffectedCells: []string{cell.CellID.String()},
			Intensity:     0.8,
		}
	}

	// Check blizzard
	if CheckForBlizzard(state.Temperature, state.Precipitation, state.Wind.Speed) {
		return &ExtremeWeatherEvent{
			EventType:     ExtremeBlizzard,
			StartTime:     state.Timestamp,
			Duration:      time.Hour * 48, // Typical duration: 2 days
			AffectedCells: []string{cell.CellID.String()},
			Intensity:     0.7,
		}
	}

	// For drought and heat wave, we'd need historical analysis
	// Simplified check here
	if len(recentHistory) >= 90 {
		avgPrecip := averagePrecipitation(recentHistory)
		if CheckForDrought(avgPrecip, cell.Temperature, 90) {
			return &ExtremeWeatherEvent{
				EventType:     ExtremeDrought,
				StartTime:     state.Timestamp,
				Duration:      time.Hour * 24 * 120, // Long duration
				AffectedCells: []string{cell.CellID.String()},
				Intensity:     0.6,
			}
		}
	}

	return nil
}

func averagePrecipitation(history []WeatherState) float64 {
	if len(history) == 0 {
		return 0
	}

	sum := 0.0
	for _, state := range history {
		sum += state.Precipitation
	}
	return sum / float64(len(history))
}
