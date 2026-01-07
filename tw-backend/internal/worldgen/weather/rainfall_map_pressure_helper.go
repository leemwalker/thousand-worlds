package weather

import (
	"math"
	"sync"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// generateSimplifiedPressureMap creates a transient pressure map for rainfall generation.
// It estimates temperature based on latitude and elevation without full climate simulation.
// Parallelized by face for performance.
func generateSimplifiedPressureMap(
	sphereMap *geography.SphereHeightmap,
	topology spatial.Topology,
	seaLevel float64,
	season Season,
) map[spatial.Coordinate]float64 {
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

	// Process each face in parallel, collect results per-face
	faceResults := make([]map[spatial.Coordinate]float64, 6)
	var wg sync.WaitGroup
	wg.Add(6)

	for face := 0; face < 6; face++ {
		go func(f int) {
			defer wg.Done()
			local := make(map[spatial.Coordinate]float64, res*res)

			for y := 0; y < res; y++ {
				for x := 0; x < res; x++ {
					coord := spatial.Coordinate{Face: f, X: x, Y: y}
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
					declination := declination_Ocean
					if isLand {
						declination = declination_Land
					}
					tempMod := GetSeasonalTemperatureModifier(lat, declination)
					temp := baseTemp + tempMod

					// Calculate thermal pressure (Land vs Ocean, Hot vs Cold)
					pressure := CalculateSurfacePressure(isLand, temp)

					// Add Planetary Pressure Bands (Global Circulation)
					absLat := math.Abs(lat)
					var latPressureMod float64

					if absLat < 20 {
						// Equatorial Low (ITCZ)
						latPressureMod = -15.0 * (1.0 - absLat/20.0)
					} else if absLat >= 20 && absLat < 45 {
						// Subtropical High (Horse Latitudes)
						dist := math.Abs(absLat - 30.0)
						if dist < 15 {
							latPressureMod = 20.0 * (1.0 - dist/15.0)
						}
					} else if absLat >= 45 && absLat < 75 {
						// Subpolar Low
						dist := math.Abs(absLat - 60.0)
						if dist < 15 {
							latPressureMod = -15.0 * (1.0 - dist/15.0)
						}
					} else {
						// Polar High
						dist := 90.0 - absLat
						latPressureMod = 10.0 * (1.0 - dist/15.0)
					}

					pressure += latPressureMod
					local[coord] = pressure
				}
			}
			faceResults[f] = local
		}(face)
	}
	wg.Wait()

	// Merge results from all faces
	pressureMap := make(map[spatial.Coordinate]float64, 6*res*res)
	for _, fm := range faceResults {
		for coord, pressure := range fm {
			pressureMap[coord] = pressure
		}
	}

	return pressureMap
}
