package world

import "time"

// TimeOfDay represents the descriptive time of day
type TimeOfDay string

const (
	TimeOfDayNight     TimeOfDay = "Night"
	TimeOfDayDawn      TimeOfDay = "Dawn"
	TimeOfDayMorning   TimeOfDay = "Morning"
	TimeOfDayNoon      TimeOfDay = "Noon"
	TimeOfDayAfternoon TimeOfDay = "Afternoon"
	TimeOfDayDusk      TimeOfDay = "Dusk"
	TimeOfDayEvening   TimeOfDay = "Evening"
)

// DefaultDayLength is 24 game-hours
const DefaultDayLength = 24 * time.Hour

// CalculateSunPosition calculates the sun position (0.0-1.0) based on game time
// 0.0 = Midnight, 0.5 = Noon, 1.0 = Midnight
func CalculateSunPosition(gameTime time.Duration, dayLength time.Duration) float64 {
	if dayLength <= 0 {
		return 0.0
	}
	// Calculate progress through the current day
	dayProgress := float64(gameTime%dayLength) / float64(dayLength)
	return dayProgress
}

// GetTimeOfDay returns the descriptive time of day based on sun position
func GetTimeOfDay(sunPosition float64) TimeOfDay {
	switch {
	case sunPosition < 0.25:
		return TimeOfDayNight
	case sunPosition < 0.30:
		return TimeOfDayDawn
	case sunPosition < 0.45:
		return TimeOfDayMorning
	case sunPosition < 0.55:
		return TimeOfDayNoon
	case sunPosition < 0.70:
		return TimeOfDayAfternoon
	case sunPosition < 0.75:
		return TimeOfDayDusk
	case sunPosition < 0.90:
		return TimeOfDayEvening
	default:
		return TimeOfDayNight
	}
}
