package weather

import (
	"math"
)

// CalculatePrecipitation calculates precipitation for a cell based on upwind conditions
func CalculatePrecipitation(
	cell *GeographyCell,
	upwindCells []*GeographyCell,
	wind Wind,
	currentMoisture float64,
) (precipitation float64, newMoisture float64) {

	// Start with current moisture or accumulate from upwind
	moisture := currentMoisture

	// Accumulate moisture from upwind water bodies
	if len(upwindCells) > 0 {
		for _, upwind := range upwindCells {
			if upwind.IsWater() {
				// 5% moisture per m/s over water
				moisture += wind.Speed * 0.05
			}
		}
	}

	// Cap moisture at 100%
	if moisture > 100 {
		moisture = 100
	}

	precip := 0.0

	// Orographic effect - air forced upward by elevation gain
	if len(upwindCells) > 0 && cell.Elevation > upwindCells[0].Elevation {
		elevationGain := cell.Elevation - upwindCells[0].Elevation

		// Adiabatic cooling: air rises and cools
		// More moisture released with greater elevation gain
		precipMm := moisture * elevationGain * 0.001 // 0.1% per meter gain

		// Cap precipitation per event
		if precipMm > moisture*10 {
			precipMm = moisture * 10
		}

		precip = precipMm

		// Moisture depleted by precipitation (rain shadow effect)
		moisture -= precip * 5 // Significant depletion
		if moisture < 0 {
			moisture = 0
		}
	} else if moisture > 50 {
		// Flat land precipitation when humidity exceeds threshold
		precip = (moisture - 50) * 0.5 // Light rain
		moisture -= precip * 2
		if moisture < 0 {
			moisture = 0
		}
	}

	return precip, moisture
}

// CalculateAnnualPrecipitation estimates annual precipitation for a location
// This is a simplified calculation for testing purposes
func CalculateAnnualPrecipitation(
	latitude float64,
	elevation float64,
	distanceToCoast float64,
	isWindward bool,
) float64 {
	// Base precipitation by latitude
	absLat := math.Abs(latitude)

	var basePrecip float64
	if absLat < 15 {
		// Tropical: high precipitation
		basePrecip = 3000 // mm/year
	} else if absLat < 30 {
		// Subtropics: often dry (descending air)
		basePrecip = 500
	} else if absLat < 60 {
		// Mid-latitudes: moderate
		basePrecip = 1000
	} else {
		// Polar: low precipitation
		basePrecip = 300
	}

	// Coastal effect: more precipitation near coast
	coastalFactor := 1.0
	if distanceToCoast < 100000 { // < 100km
		coastalFactor = 1.5
	} else if distanceToCoast < 500000 { // < 500km
		coastalFactor = 1.2
	}

	// Elevation effect
	elevFactor := 1.0
	if elevation > 1000 && isWindward {
		// Mountains on windward side get more rain
		elevFactor = 2.0
	} else if elevation > 1000 && !isWindward {
		// Rain shadow: dry leeward side
		elevFactor = 0.3
	}

	return basePrecip * coastalFactor * elevFactor
}
