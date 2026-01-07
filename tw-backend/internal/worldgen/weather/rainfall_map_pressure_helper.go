package weather

import (
	"math"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// generateSimplifiedPressureMap creates a transient pressure map for rainfall generation.
// It estimates temperature based on latitude and elevation without full climate simulation.
func generateSimplifiedPressureMap(
	sphereMap *geography.SphereHeightmap,
	topology spatial.Topology,
	seaLevel float64,
	season Season,
) map[spatial.Coordinate]float64 {
	pressureMap := make(map[spatial.Coordinate]float64)
	res := sphereMap.Resolution()

	// Get day of year for the season (approximate)
	dayOfYear := 80 // Spring
	if season == SeasonSummer {
		dayOfYear = 172
	} else if season == SeasonFall {
		dayOfYear = 264
	} else if season == SeasonWinter {
		dayOfYear = 355
	}
	declination_Land := CalculateThermalDeclination(dayOfYear, true)
	declination_Ocean := CalculateThermalDeclination(dayOfYear, false)

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				elevation := sphereMap.Get(coord)
				isLand := elevation > seaLevel

				// Calculate base temperature
				lat := GetLatitudeFromCoord(topology, coord)
				latNormalized := math.Abs(lat) / 90.0

				// Estimate temp (simplified from calculateTemperatureFromLatitude)
				// Equator: 30C, Poles: -30C
				baseTemp := 30.0 - (latNormalized * 60.0)

				// Lapse rate: -6.5C per km
				if elevation > seaLevel {
					heightAboveSea := (elevation - seaLevel) / 1000.0
					baseTemp -= heightAboveSea * 6.5
				}

				// Apply seasonal modifier using Thermal Declination
				// This simulates phase lag (Land warms/cools faster than Ocean)
				declination := declination_Ocean
				if isLand {
					declination = declination_Land
				}
				tempMod := GetSeasonalTemperatureModifier(lat, declination)
				temp := baseTemp + tempMod

				// Calculate thermal pressure (Land vs Ocean, Hot vs Cold)
				pressure := CalculateSurfacePressure(isLand, temp)

				// Add Planetary Pressure Bands (Global Circulation)
				// Equator (0): Low
				// Subtropics (30): High
				// Subpolar (60): Low
				// Poles (90): High
				absLat := math.Abs(lat)
				var latPressureMod float64

				if absLat < 20 {
					// Equatorial Low (ITCZ)
					// peak at 0, taper to 20
					latPressureMod = -15.0 * (1.0 - absLat/20.0)
				} else if absLat >= 20 && absLat < 45 {
					// Subtropical High (Horse Latitudes)
					// peak at 30
					dist := math.Abs(absLat - 30.0)
					if dist < 15 {
						latPressureMod = 20.0 * (1.0 - dist/15.0)
					}
				} else if absLat >= 45 && absLat < 75 {
					// Subpolar Low
					// peak at 60
					dist := math.Abs(absLat - 60.0)
					if dist < 15 {
						latPressureMod = -15.0 * (1.0 - dist/15.0)
					}
				} else {
					// Polar High
					// peak at 90
					dist := 90.0 - absLat
					latPressureMod = 10.0 * (1.0 - dist/15.0)
				}

				pressure += latPressureMod
				pressureMap[coord] = pressure
			}
		}
	}

	return pressureMap
}
