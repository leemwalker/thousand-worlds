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
// GenerateRainfallMap simulates global rainfall based on atmospheric circulation.
// Uses a spherical 3D advection model with pressure-gradient winds.
func GenerateRainfallMap(
	sphereMap *geography.SphereHeightmap,
	topology spatial.Topology,
	config RainfallConfig,
) []float64 {
	res := sphereMap.Resolution()
	totalCells := 6 * res * res
	seaLevel := config.SeaLevel

	// Phase 1.2: Generate pressure map for accurate wind calculations
	// We use Spring as the representative season for annual rainfall
	pressureMap := generateSimplifiedPressureMap(sphereMap, topology, seaLevel, SeasonSpring)

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

	// Pre-calculate upwind neighbors for all cells
	// Wind is constant during advection steps, so we don't need to recalculate it inside the loop
	upwindNeighbors := make([]int, totalCells)
	for idx := 0; idx < totalCells; idx++ {
		coord := indexToCoord(idx, res)

		// Get wind vector from pressure gradients (Phase 1.2 verification)
		// This replaces the simple latitude-based wind model
		windVec := CalculatePressureGradientWind(coord, topology, pressureMap)

		// Apply Coriolis Effect to deflect wind (create zonal flow)
		// Northern Hemisphere: Deflect Right (Clockwise)
		// Southern Hemisphere: Deflect Left (Counter-Clockwise)
		// Magnitude increases with latitude
		lat := GetLatitudeFromCoord(topology, coord)

		// Deflection angle: Approaches Geostrophic (90 deg) away from equator.
		// We use Tanh to ramp up quickly (Trade Winds start near equator).
		// Max deflection 80 degrees leaves a component of meridional flow (Hadley convergence).
		// NH (lat>0): +80 deg -> Deflect Right (cw) -> -80 rad rotation.
		// SH (lat<0): -80 deg -> Deflect Left (ccw) -> +80 rad rotation. (Wait, logic in RotateAround handles sign).
		// Note from before: NH needs NEGATIVE angle for CW rotation.
		// If lat > 0, Tanh > 0. deflectionDeg = 80.
		// We pass -deflectionRad to RotateAround.
		// So -80 rad. Correct.
		deflectionDeg := 80.0 * math.Tanh(lat/5.0)
		deflectionRad := deflectionDeg * math.Pi / 180.0

		// Rotation axis is the surface normal at this point
		px, py, pz := topology.ToSphere(coord)
		normal := spatial.Vector3D{X: px, Y: py, Z: pz}

		// Rotate wind vector
		// NH: Deflect Right (CW). deflectionDeg > 0. Angle = -deflectionRad.
		// SH: Deflect Left (CCW). deflectionDeg < 0. Angle = -deflectionRad (becomes positive).
		windVec = windVec.RotateAround(normal, -deflectionRad)

		// Falls back to simplified wind if gradient is too weak
		if windVec.Length() < 0.1 {
			windVec = Get3DWindVector(topology, coord, SeasonSpring)
		}

		// Upwind is opposite of wind direction
		upwindVec := windVec.Scale(-1)

		// Convert 3D wind to local cardinal direction
		upwindDir := WindToLocalDirection(topology, coord, upwindVec)

		// Find upwind neighbor using topology (handles face transitions)
		// We store the index directly for fast lookup in the loop
		upwindNeighbor := topology.GetNeighbor(coord, upwindDir)
		upwindNeighbors[idx] = coordToIndex(upwindNeighbor, res)
	}

	// Step 2: Moisture advection passes
	for pass := 0; pass < config.AdvectionPasses; pass++ {
		// Copy current moisture for reading
		copy(tempMoisture, moisture)

		for idx := 0; idx < totalCells; idx++ {
			coord := indexToCoord(idx, res)
			elev := sphereMap.Get(coord)

			// Use pre-calculated upwind neighbor
			upwindIdx := upwindNeighbors[idx]

			if upwindIdx >= 0 && upwindIdx < totalCells {
				upwindCoord := indexToCoord(upwindIdx, res)
				upwindElev := sphereMap.Get(upwindCoord)
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

	// Determine day of year for the season
	// We use Spring (Equinox) as the representative season for annual rainfall pattern
	dayOfYear := 80 // Spring
	// For rainfall generation, we use the config season.
	// But generateSimplifiedPressureMap currently duplicates this logic. It's fine for now.

	// Step 4: Apply latitude-based precipitation modifiers
	for idx := 0; idx < totalCells; idx++ {
		coord := indexToCoord(idx, res)
		elev := sphereMap.Get(coord)
		isLand := elev > config.SeaLevel

		if !isLand {
			rainfall[idx] = 0 // No rain over ocean (tracking on land only)
			continue
		}

		lat := GetLatitudeFromCoord(topology, coord)
		absLat := math.Abs(lat)

		// Calculate Thermal Equator (ITCZ position) for this cell
		// Land heats/cools fast (30 day lag), Ocean slow (75 day lag)
		thermalDeclination := CalculateThermalDeclination(dayOfYear, isLand)

		// ITCZ (tropical) bonus - centers on thermal equator
		distFromITCZ := math.Abs(lat - thermalDeclination)
		if distFromITCZ < 15 {
			// Peak bonus at the ITCZ center
			bonus := 1.5 * (1.0 - distFromITCZ/30.0) // decay factor
			rainfall[idx] *= (1.0 + bonus)
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

// normalizeDirection normalizes an angle to 0-360 degrees.
// This helper is kept for potential future use in latitude modifiers.
func normalizeRainfallDirection(direction float64) float64 {
	for direction < 0 {
		direction += 360
	}
	for direction >= 360 {
		direction -= 360
	}
	return direction
}
