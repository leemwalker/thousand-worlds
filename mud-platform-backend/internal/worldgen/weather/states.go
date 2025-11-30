package weather

// DetermineWeatherState assigns a weather state based on conditions
func DetermineWeatherState(
	temperature float64,
	precipitation float64,
	humidity float64,
	windSpeed float64,
) WeatherType {

	// Storm conditions: heavy precipitation OR high winds
	if precipitation > 20 || windSpeed > 15 {
		return WeatherStorm
	}

	// Snow: cold temperature with precipitation
	if temperature <= 0 && precipitation > 2 {
		return WeatherSnow
	}

	// Rain: precipitation above threshold
	if precipitation > 2 {
		return WeatherRain
	}

	// Cloudy: high humidity but no significant precipitation
	if humidity >= 30 && humidity < 60 {
		return WeatherCloudy
	}

	// Clear: low humidity, no precipitation
	return WeatherClear
}

// CalculateVisibility determines visibility based on weather conditions
func CalculateVisibility(weatherType WeatherType) float64 {
	switch weatherType {
	case WeatherClear:
		return 50.0 // km
	case WeatherCloudy:
		return 30.0
	case WeatherRain:
		return 10.0
	case WeatherSnow:
		return 5.0
	case WeatherStorm:
		return 2.0
	default:
		return 50.0
	}
}
