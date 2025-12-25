package weather

import (
	"math"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// CalculateSeasonalPrecipitation computes precipitation accounting for ITCZ shift
// and monsoon moisture transport.
//
// Physics:
//   - ITCZ (Intertropical Convergence Zone) follows the sun's overhead position
//   - Heavy rain occurs where air converges and rises (near solar declination)
//   - Monsoon winds bring moisture from ocean to land in summer
//
// Parameters:
//   - cell: The geography cell to calculate precipitation for
//   - upwindCells: Cells upwind of the current cell
//   - monsoonWind: Pressure-gradient driven wind vector (from CalculatePressureGradientWind)
//   - solarDeclination: Current solar declination (from CalculateSolarDeclination)
//   - currentMoisture: Accumulated moisture from upwind transport
//   - topology: Sphere topology for coordinate conversions
//
// Returns:
//   - precipitation: mm of rainfall
//   - newMoisture: Updated moisture after precipitation
func CalculateSeasonalPrecipitation(
	cell *GeographyCell,
	upwindCells []*GeographyCell,
	monsoonWind spatial.Vector3D,
	solarDeclination float64,
	currentMoisture float64,
	topology spatial.Topology,
) (precipitation float64, newMoisture float64) {
	moisture := currentMoisture

	// 1. Get latitude of this cell
	var latitude float64
	if cell.SphereCoord != nil && topology != nil {
		latitude = GetLatitudeFromCoord(topology, *cell.SphereCoord)
	} else {
		// Flat mode: use Y as approximate latitude
		latitude = cell.Location.Y
	}

	// 2. ITCZ Effect: Heavy rain near the solar declination latitude
	// The convergence zone shifts with the sun's position
	distanceFromITCZ := math.Abs(latitude - solarDeclination)

	var itczBonus float64
	if distanceFromITCZ < 10 {
		// Within 10 degrees of ITCZ: heavy convective rainfall
		// Maximum effect at the ITCZ itself
		itczBonus = (10 - distanceFromITCZ) * 5.0 // Up to 50mm bonus
	}

	// 3. Monsoon Moisture Transport
	// Wind magnitude indicates how strongly moisture is being transported
	windMagnitude := math.Sqrt(
		monsoonWind.X*monsoonWind.X +
			monsoonWind.Y*monsoonWind.Y +
			monsoonWind.Z*monsoonWind.Z,
	)

	// Check if upwind cells are ocean (moisture source)
	oceanMoistureGain := 0.0
	for _, upwind := range upwindCells {
		if upwind.IsWater() {
			// Gain moisture from ocean proportional to wind speed
			oceanMoistureGain += windMagnitude * 0.05
		}
	}
	moisture += oceanMoistureGain

	// Cap moisture at 100%
	if moisture > 100 {
		moisture = 100
	}

	// 4. Calculate base precipitation using existing orographic logic
	precip := 0.0

	// Orographic effect - air forced upward by elevation gain
	if len(upwindCells) > 0 && cell.Elevation > upwindCells[0].Elevation {
		elevationGain := cell.Elevation - upwindCells[0].Elevation

		// Adiabatic cooling: air rises and cools, releasing moisture
		precipMm := moisture * elevationGain * 0.001

		// Cap precipitation per event
		if precipMm > moisture*10 {
			precipMm = moisture * 10
		}

		precip = precipMm

		// Moisture depleted by precipitation (rain shadow)
		moisture -= precip * 5
		if moisture < 0 {
			moisture = 0
		}
	} else if moisture > 40 {
		// High humidity precipitation on flat land
		precip = (moisture - 40) * 0.3
		moisture -= precip * 1.5
		if moisture < 0 {
			moisture = 0
		}
	}

	// 5. Add ITCZ convective rainfall
	precip += itczBonus

	return precip, moisture
}

// CalculateSeasonalAnnualPrecipitation estimates annual precipitation with seasonal effects.
// This extends CalculateAnnualPrecipitation to account for ITCZ migration and monsoons.
//
// Parameters:
//   - latitude: Latitude in degrees
//   - elevation: Elevation in meters
//   - distanceToCoast: Distance to nearest coast in meters
//   - isWindward: Whether location is on windward side of terrain
//   - solarDeclination: Current solar declination for seasonal effects
//
// Returns: Estimated annual precipitation in mm/year
func CalculateSeasonalAnnualPrecipitation(
	latitude float64,
	elevation float64,
	distanceToCoast float64,
	isWindward bool,
	solarDeclination float64,
) float64 {
	// Base precipitation by latitude (from existing function)
	absLat := math.Abs(latitude)

	var basePrecip float64
	if absLat < 15 {
		basePrecip = 3000 // Tropical: high precipitation
	} else if absLat < 30 {
		basePrecip = 500 // Subtropics: often dry
	} else if absLat < 60 {
		basePrecip = 1000 // Mid-latitudes: moderate
	} else {
		basePrecip = 300 // Polar: low precipitation
	}

	// ITCZ seasonal shift effect
	// When ITCZ is near this latitude, increase precipitation
	distanceFromITCZ := math.Abs(latitude - solarDeclination)
	if distanceFromITCZ < 15 {
		// ITCZ proximity bonus
		itczFactor := 1.0 + (15-distanceFromITCZ)*0.05
		basePrecip *= itczFactor
	}

	// Coastal monsoon effect
	// Coastal areas in tropics get extra rainfall when onshore winds
	coastalFactor := 1.0
	if distanceToCoast < 100000 { // < 100km
		// Check if this latitude is in monsoon belt (10-25° from equator)
		if absLat >= 10 && absLat <= 25 {
			// Strong monsoon effect
			coastalFactor = 1.8
		} else {
			coastalFactor = 1.5
		}
	} else if distanceToCoast < 500000 { // < 500km
		coastalFactor = 1.2
	}

	// Elevation effect
	elevFactor := 1.0
	if elevation > 1000 && isWindward {
		elevFactor = 2.0 // Mountains on windward side
	} else if elevation > 1000 && !isWindward {
		elevFactor = 0.3 // Rain shadow
	}

	return basePrecip * coastalFactor * elevFactor
}

// GeneratePressureMap creates a pressure map for seasonal weather simulation.
// Called once per day to establish global pressure patterns.
func GeneratePressureMap(
	sphereMap *geography.SphereHeightmap,
	topology spatial.Topology,
	climateData SphereClimateMap,
	dayOfYear int,
	seaLevel float64,
) map[spatial.Coordinate]float64 {
	pressureMap := make(map[spatial.Coordinate]float64)

	faceSize := sphereMap.Resolution()
	declination := CalculateSolarDeclination(dayOfYear)

	for face := 0; face < 6; face++ {
		for y := 0; y < faceSize; y++ {
			for x := 0; x < faceSize; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				idx := y*faceSize + x

				elevation := sphereMap.Get(coord)
				isLand := elevation > seaLevel

				// Get base temperature from climate data
				baseTemp := climateData[face][idx].Temperature

				// Apply seasonal temperature modifier
				lat := GetLatitudeFromCoord(topology, coord)
				tempMod := GetSeasonalTemperatureModifier(lat, declination)
				temp := baseTemp + tempMod

				// Calculate pressure
				pressure := CalculateSurfacePressure(isLand, temp)
				pressureMap[coord] = pressure
			}
		}
	}

	return pressureMap
}
