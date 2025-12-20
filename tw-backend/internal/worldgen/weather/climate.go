package weather

// SimulateRainShadow reduces precipitation downwind of mountains
func SimulateRainShadow(elevation float64, isDownwindOfMountain bool) float64 {
	if isDownwindOfMountain && elevation > 2000 {
		return 0.2 // 80% reduction (Rain Shadow)
	}
	return 1.0
}

// SimulateENSO handles El Nino Southern Oscillation effects
func SimulateENSO(isElNino bool) (tempChange float64, precipMultiplier float64) {
	if isElNino {
		return 1.0, 2.0 // Warmer, wetter (e.g., eastern Pacific)
	}
	return -0.5, 0.5 // Cooler, drier (La Nina)
}

// CalculateAxialTiltEffect determines season based on hemisphere and month
func CalculateAxialTiltEffect(latitude float64, month int) Season {
	// Simplified: Months 1-6 are Northern Summer-building?
	// Usually June-August is Northern Summer.
	// Let's match the user requirement: "If Month < 6, North=Summer" (User said "Month < 6, North=Summer"??)
	// Actually user said: "If Month < 6, North=Summer; else North=Winter."
	// (Note: This is reversed from Earth, where Jan-Jun is Winter/Spring, but I will follow user spec if rigid,
	// OR follow standard Earth logic if user just wants "Seasons work".
	// User Stub: "If Month < 6, North=Summer". Okay will follow constraint.)

	isNorth := latitude > 0
	if month < 6 {
		if isNorth {
			return SeasonSummer
		}
		return SeasonWinter
	}
	// Month >= 6
	if isNorth {
		return SeasonWinter
	}
	return SeasonSummer
}

// SimulateWaterCycle returns total mass for conservation check
func SimulateWaterCycle(ocean, land, atmosphere float64) float64 {
	return ocean + land + atmosphere
}
