package weather

import (
	"math"
)

// CalculateWind calculates wind direction and speed based on atmospheric circulation
func CalculateWind(latitude float64, longitude float64, season Season) Wind {
	absLat := math.Abs(latitude)

	var windDirection float64
	var windSpeed float64

	if absLat < 30 {
		// Hadley cell: Trade winds (easterly)
		windDirection = -90           // Easterly (from east, blowing west)
		windSpeed = 5 + (30-absLat)/6 // 5-10 m/s
	} else if absLat < 60 {
		// Ferrel cell: Westerlies
		windDirection = 90            // Westerly (from west, blowing east)
		windSpeed = 8 + (absLat-30)/6 // 8-13 m/s
	} else {
		// Polar cell: Polar easterlies
		windDirection = -90            // Easterly
		windSpeed = 3 + (90-absLat)/10 // 3-6 m/s
	}

	// Add Coriolis effect deflection
	// Northern hemisphere: deflect right, Southern: deflect left
	coriolisDeflection := 15 * math.Copysign(1, latitude)
	windDirection += coriolisDeflection

	// Normalize direction to 0-360
	windDirection = normalizeDirection(windDirection)

	return Wind{
		Direction: windDirection,
		Speed:     windSpeed,
	}
}

// GetAtmosphericCell returns the type of atmospheric cell for the given latitude
func GetAtmosphericCell(latitude float64) AtmosphericCell {
	absLat := math.Abs(latitude)

	if absLat < 30 {
		return CellHadley
	} else if absLat < 60 {
		return CellFerrel
	}
	return CellPolar
}

// GetPressureAtLatitude determines if a latitude is high or low pressure
func GetPressureAtLatitude(latitude float64) PressureSystem {
	absLat := math.Abs(latitude)

	// Low pressure at: equator (0°), 60°
	// High pressure at: 30°, poles (90°)

	if absLat < 15 {
		return PressureLow // Equatorial low
	} else if absLat < 45 {
		return PressureHigh // Subtropical high (~30°)
	} else if absLat < 75 {
		return PressureLow // Subpolar low (~60°)
	}
	return PressureHigh // Polar high
}

// normalizeDirection normalizes an angle to 0-360 degrees
func normalizeDirection(direction float64) float64 {
	for direction < 0 {
		direction += 360
	}
	for direction >= 360 {
		direction -= 360
	}
	return direction
}
