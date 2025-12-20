package weather

// SpawnDisaster creates extreme weather events based on conditions
func SpawnDisaster(temp float64, windSpeed float64, overOcean bool, humidity float64) string {
	// Periodic check or threshold check
	if overOcean && temp > 27.0 && windSpeed > 90.0 {
		return "hurricane"
	}
	if !overOcean && temp < -10.0 && windSpeed > 50.0 {
		return "blizzard"
	}
	if !overOcean && temp > 35.0 && windSpeed > 40.0 && humidity < 0.2 {
		return "sandstorm"
	}
	return ""
}
