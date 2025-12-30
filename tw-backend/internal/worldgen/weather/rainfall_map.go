package weather

import (
	"math"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// =============================================================================
// Rainfall Map Generation with Moisture Advection
// =============================================================================

// RainfallConfig holds parameters for rainfall simulation
type RainfallConfig struct {
	SeaLevel        float64 // Sea level elevation in meters
	OceanEvapRate   float64 // Base evaporation rate from ocean (moisture/cell)
	AdvectionPasses int     // Number of advection iterations
}

// DefaultRainfallConfig returns standard rainfall simulation parameters
func DefaultRainfallConfig(seaLevel float64) RainfallConfig {
	return RainfallConfig{
		SeaLevel:        seaLevel,
		OceanEvapRate:   10.0, // Base moisture units from ocean
		AdvectionPasses: 5,    // Iterate moisture transport
	}
}

// GenerateRainfallMap computes precipitation for all cells using global circulation.
// Algorithm:
//  1. Ocean cells emit moisture based on location (tropical > polar)
//  2. Moisture advects downwind using 3-cell circulation model
//  3. Orographic lift causes precipitation on windward slopes
//  4. Rain shadow forms on leeward sides
//
// Returns a flat array of rainfall values indexed by face*res*res + y*res + x
func GenerateRainfallMap(
	sphereMap *geography.SphereHeightmap,
	topology spatial.Topology,
	config RainfallConfig,
) []float64 {
	res := sphereMap.Resolution()
	totalCells := 6 * res * res
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Initialize rainfall and moisture grids
	rainfall := make([]float64, totalCells)
	moisture := make([]float64, totalCells)
	tempMoisture := make([]float64, totalCells)

	// Step 1: Initial moisture from ocean evaporation
	for idx := 0; idx < totalCells; idx++ {
		coord := indexToCoord(idx, res)
		elev := sphereMap.Get(coord)

		if elev <= config.SeaLevel {
			// Ocean cell: evaporate moisture
			// Tropical oceans evaporate more
			lat := GetLatitudeFromCoord(topology, coord)
			absLat := math.Abs(lat)

			// Evaporation decreases toward poles
			latFactor := 1.0 - (absLat / 90.0 * 0.5) // 1.0 at equator, 0.5 at poles
			moisture[idx] = config.OceanEvapRate * latFactor
		}
	}

	// Step 2: Moisture advection passes
	for pass := 0; pass < config.AdvectionPasses; pass++ {
		// Copy current moisture for reading
		copy(tempMoisture, moisture)

		for idx := 0; idx < totalCells; idx++ {
			coord := indexToCoord(idx, res)
			elev := sphereMap.Get(coord)

			// Get wind at this location
			wind := CalculateWindSpherical(topology, coord, SeasonSpring)

			// Calculate upwind direction (opposite of wind)
			upwindDir := normalizeDirection(wind.Direction + 180)

			// Find upwind neighbor
			upwindNeighbor := getNeighborInDirection(topology, coord, upwindDir, directions)
			upwindIdx := coordToIndex(upwindNeighbor, res)

			if upwindIdx >= 0 && upwindIdx < totalCells {
				upwindElev := sphereMap.Get(upwindNeighbor)
				upwindMoisture := tempMoisture[upwindIdx]

				// Transport moisture from upwind
				transportRate := 0.8 // 80% of upwind moisture transported
				transportedMoisture := upwindMoisture * transportRate

				// Step 3: Orographic lift - check elevation change
				elevDiff := elev - upwindElev

				if elevDiff > 0 {
					// WINDWARD: Air rises, cools, dumps rain
					// Precipitation rate proportional to elevation gain and moisture
					liftFactor := math.Min(elevDiff/1000.0, 1.0) // Max effect at 1000m gain
					precipAmount := transportedMoisture * liftFactor * 0.5

					rainfall[idx] += precipAmount

					// Reduce moisture (rain shadow effect)
					moisture[idx] += transportedMoisture * (1.0 - liftFactor*0.8)
				} else if elevDiff < 0 {
					// LEEWARD: Air descends, warms, holds moisture (dry)
					// No precipitation, but moisture continues
					moisture[idx] += transportedMoisture
				} else {
					// Flat terrain: some precipitation, some transport
					flatPrecip := transportedMoisture * 0.1
					rainfall[idx] += flatPrecip
					moisture[idx] += transportedMoisture - flatPrecip
				}
			}

			// Ocean cells continuously evaporate
			if elev <= config.SeaLevel {
				lat := GetLatitudeFromCoord(topology, coord)
				absLat := math.Abs(lat)
				latFactor := 1.0 - (absLat / 90.0 * 0.5)
				moisture[idx] = math.Max(moisture[idx], config.OceanEvapRate*latFactor)
			}
		}
	}

	// Step 4: Apply latitude-based precipitation modifiers
	for idx := 0; idx < totalCells; idx++ {
		coord := indexToCoord(idx, res)
		elev := sphereMap.Get(coord)

		if elev <= config.SeaLevel {
			rainfall[idx] = 0 // No rain over ocean (tracking on land only)
			continue
		}

		lat := GetLatitudeFromCoord(topology, coord)
		absLat := math.Abs(lat)

		// ITCZ (tropical) bonus
		if absLat < 15 {
			rainfall[idx] *= 1.5
		}

		// Subtropical high (desert latitudes) penalty
		if absLat > 20 && absLat < 35 {
			rainfall[idx] *= 0.5
		}

		// Cap rainfall at reasonable maximum (mm/year equivalent)
		if rainfall[idx] > 3000 {
			rainfall[idx] = 3000
		}
	}

	return rainfall
}

// Helper: convert index to coordinate
func indexToCoord(idx, res int) spatial.Coordinate {
	resSq := res * res
	face := idx / resSq
	rem := idx % resSq
	y := rem / res
	x := rem % res
	return spatial.Coordinate{Face: face, X: x, Y: y}
}

// Helper: convert coordinate to index
func coordToIndex(coord spatial.Coordinate, res int) int {
	resSq := res * res
	return coord.Face*resSq + coord.Y*res + coord.X
}

// getNeighborInDirection finds the neighbor closest to the given wind direction
func getNeighborInDirection(topology spatial.Topology, coord spatial.Coordinate, windDir float64, directions []spatial.Direction) spatial.Coordinate {
	// Map wind direction (degrees) to cardinal direction
	// 0° = North, 90° = East, 180° = South, 270° = West

	// Normalize to 0-360
	dir := normalizeDirection(windDir)

	// Find closest cardinal direction
	var bestDir spatial.Direction
	if dir >= 315 || dir < 45 {
		bestDir = spatial.North
	} else if dir >= 45 && dir < 135 {
		bestDir = spatial.East
	} else if dir >= 135 && dir < 225 {
		bestDir = spatial.South
	} else {
		bestDir = spatial.West
	}

	return topology.GetNeighbor(coord, bestDir)
}
