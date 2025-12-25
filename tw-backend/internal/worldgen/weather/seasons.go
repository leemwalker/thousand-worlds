package weather

import (
	"math"

	"tw-backend/internal/spatial"
)

// DefaultAxialTilt is Earth's axial tilt in degrees.
// This determines the maximum latitude where the sun can be directly overhead.
// For dynamic orbital mechanics (Milankovitch cycles), use CalculateSolarDeclinationWithTilt.
const DefaultAxialTilt = 23.5

// AxialTilt is a backward-compatible alias for DefaultAxialTilt.
// Deprecated: Use DefaultAxialTilt or CalculateSolarDeclinationWithTilt for new code.
const AxialTilt = DefaultAxialTilt

// Baseline surface pressure in millibars (standard atmosphere at sea level)
const baselinePressure = 1013.25

// Average reference temperature for pressure calculations (°C)
const referenceTemp = 15.0

// CalculateSolarDeclination returns the latitude where the sun is directly overhead
// for a given day of year (0-365).
//
// Formula: δ ≈ 23.45 * sin((360/365) * (dayOfYear + 284))
// This is the simplified equation of declination.
//
// Returns:
//   - Positive values: Northern Hemisphere (sun overhead north of equator)
//   - Negative values: Southern Hemisphere (sun overhead south of equator)
//   - Range: [-23.5, +23.5] degrees
func CalculateSolarDeclination(dayOfYear int) float64 {
	return CalculateSolarDeclinationWithTilt(dayOfYear, DefaultAxialTilt)
}

// CalculateSolarDeclinationWithTilt returns the latitude where the sun is directly overhead
// for a given day of year and axial tilt.
//
// This function supports dynamic axial tilt from Milankovitch orbital cycles.
// For the default Earth-like tilt (23.5°), use CalculateSolarDeclination instead.
//
// Parameters:
//   - dayOfYear: Day of year (0-365)
//   - axialTilt: Axial tilt in degrees (typically 22.1° to 24.5° for Earth-like)
//
// Returns:
//   - Positive values: Northern Hemisphere (sun overhead north of equator)
//   - Negative values: Southern Hemisphere (sun overhead south of equator)
//   - Range: [-axialTilt, +axialTilt] degrees
func CalculateSolarDeclinationWithTilt(dayOfYear int, axialTilt float64) float64 {
	// Convert the angular component to radians
	// The +284 offset shifts the sine wave so that:
	// - Day 172 (June 21) → maximum (Northern summer solstice)
	// - Day 355 (Dec 21) → minimum (Southern summer solstice)
	// - Day 80 (Mar 21) and Day 266 (Sep 23) → 0° (equinoxes)
	angularPosition := (360.0 / 365.0) * float64(dayOfYear+284)
	angularPositionRad := angularPosition * math.Pi / 180.0

	declination := axialTilt * math.Sin(angularPositionRad)
	return declination
}

// GetSeasonalTemperatureModifier returns a temperature adjustment based on
// how close a latitude is to the current solar declination.
//
// Physics:
//   - When the sun is directly overhead (lat ≈ declination), it's "summer"
//   - When the sun is on the opposite hemisphere, it's "winter"
//   - The effect is modulated by latitude (tropics have less variation)
//
// Parameters:
//   - lat: Latitude in degrees (-90 to +90)
//   - declination: Current solar declination (from CalculateSolarDeclination)
//
// Returns:
//   - Positive values: Summer bonus (warmer)
//   - Negative values: Winter penalty (colder)
//   - Range: approximately [-15, +15] degrees C
func GetSeasonalTemperatureModifier(lat float64, declination float64) float64 {
	// Calculate how "aligned" the latitude is with the sun's position
	// Positive alignment = summer, negative = winter
	//
	// For Northern Hemisphere (lat > 0):
	//   - When declination > 0 (summer), we get a positive modifier
	//   - When declination < 0 (winter), we get a negative modifier
	//
	// For Southern Hemisphere (lat < 0):
	//   - The opposite applies

	// Seasonal effect strength based on latitude
	// Tropics (|lat| < 15°): minimal seasonal variation
	// Mid-latitudes (15-60°): maximum seasonal variation
	// Polar (|lat| > 60°): moderate variation (always cold/dark in winter)
	absLat := math.Abs(lat)

	var seasonalAmplitude float64
	if absLat < 15 {
		// Tropics: very small seasonal swing (max ±2.5°C)
		seasonalAmplitude = 2.5
	} else if absLat < 60 {
		// Mid-latitudes: large seasonal swing (max ±15°C)
		// Scale from 2.5 at 15° to 15 at 45°, then back to 12 at 60°
		if absLat < 45 {
			seasonalAmplitude = 2.5 + (absLat-15.0)*(15.0-2.5)/(45.0-15.0)
		} else {
			seasonalAmplitude = 15.0 - (absLat-45.0)*(15.0-12.0)/(60.0-45.0)
		}
	} else {
		// Polar: moderate amplitude but extreme conditions
		seasonalAmplitude = 12.0 - (absLat-60.0)*(12.0-8.0)/(90.0-60.0)
	}

	// The modifier depends on:
	// 1. Whether the lat and declination have the same sign (same hemisphere as sun)
	// 2. How far the sun has moved toward or away from this latitude
	//
	// Simple model: modifier = amplitude * (declination / AxialTilt) * sign(lat)
	// When lat > 0 and declination > 0: positive (summer)
	// When lat > 0 and declination < 0: negative (winter)
	// Vice versa for southern hemisphere

	if lat == 0 {
		// Equator: minimal effect, just slight warming when sun is overhead
		return (1.0 - math.Abs(declination)/AxialTilt) * 2.0
	}

	// Normalize declination to [-1, 1] range
	normalizedDeclination := declination / AxialTilt

	// Apply hemisphere-aware modifier
	if lat > 0 {
		// Northern hemisphere
		return seasonalAmplitude * normalizedDeclination
	}
	// Southern hemisphere: opposite season
	return -seasonalAmplitude * normalizedDeclination
}

// CalculateSurfacePressure computes surface pressure based on land/ocean and temperature.
// Models the differential heating between land and ocean that drives monsoons.
//
// Physics:
//   - Hot land: Air rises, pressure drops (thermal low)
//   - Cold land: Air sinks, pressure rises (thermal high)
//   - Ocean: More thermally stable, smaller pressure deviations due to high heat capacity
//
// Parameters:
//   - isLand: true for land cells, false for ocean
//   - temp: Surface temperature in °C
//
// Returns: Surface pressure in millibars (mb), baseline ~1013 mb
func CalculateSurfacePressure(isLand bool, temp float64) float64 {
	// Temperature deviation from reference
	tempDeviation := temp - referenceTemp

	// Pressure change factor (mb per °C deviation)
	// Warm air rises → lower surface pressure
	// Cold air sinks → higher surface pressure
	//
	// Typical thermal low: 10-20 mb below normal
	// Typical thermal high: 10-20 mb above normal
	var pressureFactor float64

	if isLand {
		// Land heats and cools quickly → strong pressure response
		// ~0.5 mb change per °C deviation
		pressureFactor = 0.5
	} else {
		// Ocean has high heat capacity → damped pressure response
		// ~0.2 mb change per °C deviation
		pressureFactor = 0.2
	}

	// Calculate pressure deviation (negative for hot, positive for cold)
	pressureDeviation := -tempDeviation * pressureFactor

	return baselinePressure + pressureDeviation
}

// CalculatePressureGradientWind computes the wind vector from pressure gradients.
// Wind flows from high pressure to low pressure (down the pressure gradient).
//
// Physics:
//   - Wind speed proportional to pressure gradient magnitude
//   - Direction from high to low pressure
//   - Coriolis effect would deflect this, but we apply that separately
//
// Parameters:
//   - coord: The coordinate to calculate wind for
//   - topology: Sphere topology for neighbor lookups
//   - pressureMap: Map of coordinates to pressure values
//
// Returns: 3D wind vector pointing from high to low pressure
func CalculatePressureGradientWind(
	coord spatial.Coordinate,
	topology spatial.Topology,
	pressureMap map[spatial.Coordinate]float64,
) spatial.Vector3D {
	// Get pressure at current location
	centerPressure, hasCenterPressure := pressureMap[coord]
	if !hasCenterPressure {
		centerPressure = baselinePressure
	}

	// Calculate gradient by checking neighbors in all 4 cardinal directions
	directions := []spatial.Direction{
		spatial.North, spatial.South, spatial.East, spatial.West,
	}

	// Accumulate gradient vector
	var gradientX, gradientY, gradientZ float64
	neighborCount := 0

	for _, dir := range directions {
		neighbor := topology.GetNeighbor(coord, dir)
		neighborPressure, hasNeighbor := pressureMap[neighbor]
		if !hasNeighbor {
			continue
		}

		// Get 3D positions on the sphere
		cx, cy, cz := topology.ToSphere(coord)
		nx, ny, nz := topology.ToSphere(neighbor)

		// Direction vector from center to neighbor
		dx := nx - cx
		dy := ny - cy
		dz := nz - cz

		// Normalize
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dist == 0 {
			continue
		}
		dx /= dist
		dy /= dist
		dz /= dist

		// Pressure gradient component in this direction
		// Positive means neighbor has higher pressure
		pressureDiff := neighborPressure - centerPressure

		// Add contribution to gradient
		// Wind blows FROM high TO low, so we use -pressureDiff
		gradientX += -pressureDiff * dx
		gradientY += -pressureDiff * dy
		gradientZ += -pressureDiff * dz
		neighborCount++
	}

	// Average the gradient
	if neighborCount > 0 {
		gradientX /= float64(neighborCount)
		gradientY /= float64(neighborCount)
		gradientZ /= float64(neighborCount)
	}

	// Scale factor: convert pressure gradient to wind speed
	// Roughly 1 m/s per 0.1 mb gradient
	windScale := 10.0

	return spatial.Vector3D{
		X: gradientX * windScale,
		Y: gradientY * windScale,
		Z: gradientZ * windScale,
	}
}
