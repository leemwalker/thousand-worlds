// Package ocean implements ocean current simulation and heat transport.
// This breaks the strict "Latitude = Temperature" rule by allowing
// warm water currents to heat high-latitude coastal regions (Gulf Stream effect).
package ocean

import (
	"math"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/weather"
)

// System holds ocean current and temperature data for the planet.
type System struct {
	topology         spatial.Topology
	geo              *geography.SphereHeightmap
	seaLevel         float64
	CurrentMap       map[spatial.Coordinate]spatial.Vector3D
	WaterTemperature map[spatial.Coordinate]float64
}

// NewSystem creates a new ocean system.
func NewSystem(topology spatial.Topology, geo *geography.SphereHeightmap, seaLevel float64) *System {
	return &System{
		topology:         topology,
		geo:              geo,
		seaLevel:         seaLevel,
		CurrentMap:       make(map[spatial.Coordinate]spatial.Vector3D),
		WaterTemperature: make(map[spatial.Coordinate]float64),
	}
}

// IsOcean returns true if the coordinate is below sea level.
// This method is exported to implement weather.OceanTemperatureProvider.
func (s *System) IsOcean(coord spatial.Coordinate) bool {
	return s.geo.Get(coord) <= s.seaLevel
}

// GenerateSurfaceCurrents computes surface current vectors from wind.
// Physics:
//   - Base Stress: wind vector at each ocean cell
//   - Ekman Transport: rotate 45° right (NH) or left (SH)
//   - Boundary Deflection: dampen currents pointing into land
func (s *System) GenerateSurfaceCurrents(windMap map[spatial.Coordinate]spatial.Vector3D) {
	for coord, windVec := range windMap {
		// Skip land cells
		if !s.IsOcean(coord) {
			continue
		}

		// Get hemisphere from latitude
		latitude := weather.GetLatitudeFromCoord(s.topology, coord)

		// Ekman spiral: surface current is ~45° to the right of wind (NH)
		// or ~45° to the left (SH)
		// Clockwise from above (looking down at north pole) = negative rotation around up normal
		// Counter-clockwise from above = positive rotation
		ekmanAngle := -math.Pi / 4 // 45 degrees clockwise (right) for Northern Hemisphere
		if latitude < 0 {
			ekmanAngle = math.Pi / 4 // Counter-clockwise (left) for Southern Hemisphere
		}

		// Get the surface normal at this point (rotation axis)
		px, py, pz := s.topology.ToSphere(coord)
		normal := spatial.Vector3D{X: px, Y: py, Z: pz}

		// Rotate wind vector around the surface normal by Ekman angle
		currentVec := windVec.RotateAround(normal, ekmanAngle)

		// Apply boundary deflection: check if current points into land
		currentVec = s.applyBoundaryDeflection(coord, currentVec)

		// Store the current (even if dampened to zero)
		s.CurrentMap[coord] = currentVec
	}
}

// applyBoundaryDeflection dampens or redirects currents that would flow into land.
func (s *System) applyBoundaryDeflection(coord spatial.Coordinate, currentVec spatial.Vector3D) spatial.Vector3D {
	// Get the direction the current is flowing
	dir := weather.WindToLocalDirection(s.topology, coord, currentVec)

	// Check if neighbor in that direction is land
	neighbor := s.topology.GetNeighbor(coord, dir)
	if !s.IsOcean(neighbor) {
		// Neighbor is land - dampen the current significantly
		// A more sophisticated approach would project onto coastline tangent,
		// but dampening is sufficient for Phase 1
		return currentVec.Scale(0.1)
	}

	return currentVec
}

// CalculateGlobalWindVectors computes wind vectors for all ocean cells.
// This function exposes wind calculation for the ocean system.
func CalculateGlobalWindVectors(topology spatial.Topology, geo *geography.SphereHeightmap, seaLevel float64) map[spatial.Coordinate]spatial.Vector3D {
	windMap := make(map[spatial.Coordinate]spatial.Vector3D)
	resolution := topology.Resolution()

	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}

				// Only compute for ocean cells
				if geo.Get(coord) > seaLevel {
					continue
				}

				latitude := weather.GetLatitudeFromCoord(topology, coord)
				longitude := weather.GetLongitudeFromCoord(topology, coord)

				// Get wind from existing weather system
				wind := weather.CalculateWind(latitude, longitude, weather.SeasonSpring)

				// Convert wind direction/speed to 3D vector on sphere surface
				windVec := windToVector3D(topology, coord, wind)
				windMap[coord] = windVec
			}
		}
	}

	return windMap
}

// windToVector3D converts a Wind (direction + speed) to a 3D vector tangent to the sphere.
func windToVector3D(topology spatial.Topology, coord spatial.Coordinate, wind weather.Wind) spatial.Vector3D {
	// Get position on sphere
	px, py, pz := topology.ToSphere(coord)
	normal := spatial.Vector3D{X: px, Y: py, Z: pz}

	// Convert wind direction (degrees, 0=North) to local tangent vector
	// Direction 0 = North = towards positive Y axis
	// Direction 90 = East = towards positive X axis (on equator)

	// Start with a reference "north" direction (towards +Y, but projected onto tangent plane)
	worldUp := spatial.Vector3D{X: 0, Y: 1, Z: 0}

	// Project world up onto tangent plane
	dot := worldUp.Dot(normal)
	localNorth := worldUp.Sub(normal.Scale(dot)).Normalize()

	// Handle poles where localNorth would be zero
	if localNorth.Length() < 0.01 {
		// At poles, use an arbitrary reference direction
		localNorth = spatial.Vector3D{X: 1, Y: 0, Z: 0}
		localNorth = localNorth.Sub(normal.Scale(localNorth.Dot(normal))).Normalize()
	}

	// Local east is cross product of normal and north
	localEast := normal.Cross(localNorth).Normalize()

	// Convert wind direction to radians
	dirRad := wind.Direction * math.Pi / 180.0

	// Wind direction is "from" direction, we want "to" direction
	// Direction 0 = wind from north = moving south
	// So we add 180 degrees to get the movement direction
	moveRad := dirRad + math.Pi

	// Decompose into north/east components
	northComponent := math.Cos(moveRad) * wind.Speed
	eastComponent := math.Sin(moveRad) * wind.Speed

	// Build 3D vector
	result := localNorth.Scale(northComponent).Add(localEast.Scale(eastComponent))
	return result
}
