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

// AdvectionLink represents a pre-computed heat transport connection.
type AdvectionLink struct {
	SourceIdx   int     // Index in the [6][] slice
	TargetIdx   int     // Index of neighbor on the same face, or cross-face
	TargetFace  int     // Face of the target neighbor
	SpeedFactor float64 // Current-dependent lerp factor
}

// System holds ocean current and temperature data for the planet.
type System struct {
	topology         spatial.Topology
	geo              *geography.SphereHeightmap
	seaLevel         float64
	CurrentMap       [6][]spatial.Vector3D
	WaterTemperature [6][]float64
	isOcean          [6][]bool
}

// NewSystem creates a new ocean system.
func NewSystem(topology spatial.Topology, geo *geography.SphereHeightmap, seaLevel float64) *System {
	res := topology.Resolution()
	s := &System{
		topology: topology,
		geo:      geo,
		seaLevel: seaLevel,
	}
	for i := 0; i < 6; i++ {
		s.CurrentMap[i] = make([]spatial.Vector3D, res*res)
		s.WaterTemperature[i] = make([]float64, res*res)
		s.isOcean[i] = make([]bool, res*res)
	}
	return s
}

// IsOcean returns true if the coordinate is below sea level.
// After initialization, it uses the static membership to prevent simulation jitter.
func (s *System) IsOcean(coord spatial.Coordinate) bool {
	res := s.topology.Resolution()
	if len(s.isOcean[coord.Face]) > 0 {
		idx := coord.Y*res + coord.X
		return s.isOcean[coord.Face][idx]
	}
	return s.geo.Get(coord) <= s.seaLevel
}

// GenerateSurfaceCurrents computes surface current vectors from wind.
func (s *System) GenerateSurfaceCurrents(windMap map[spatial.Coordinate]spatial.Vector3D) {
	res := s.topology.Resolution()
	// Optimization: Use the windMap's iterations but write into our flat slices
	for coord, windVec := range windMap {
		if !s.IsOcean(coord) {
			continue
		}

		latitude := weather.GetLatitudeFromCoord(s.topology, coord)

		// Physics:
		//   - Base Stress: wind vector at each ocean cell
		//   - Ekman Transport: rotate 45° right (NH) or left (SH)
		//   - Boundary Deflection: dampen currents pointing into land

		// Ekman spiral: surface current is ~45° to the right of wind (NH)
		// or ~45° to the left (SH)
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

		// Store the current
		idx := coord.Y*res + coord.X
		s.CurrentMap[coord.Face][idx] = currentVec
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
		return currentVec.Scale(0.1)
	}

	return currentVec
}

// CalculateGlobalWindVectors computes wind vectors for all ocean cells.
func CalculateGlobalWindVectors(topology spatial.Topology, geo *geography.SphereHeightmap, seaLevel float64) map[spatial.Coordinate]spatial.Vector3D {
	res := topology.Resolution()
	windMap := make(map[spatial.Coordinate]spatial.Vector3D)

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				if geo.Get(coord) > seaLevel {
					continue
				}

				// Standard wind model
				windVec := weather.Get3DWindVector(topology, coord, weather.SeasonSummer)

				windMap[coord] = windVec.Scale(15.0) // 15 m/s base
			}
		}
	}

	return windMap
}

func windToVector3D(topology spatial.Topology, coord spatial.Coordinate, wind weather.Wind) spatial.Vector3D {
	// Standard wind vectors already handled by weather package
	return weather.Get3DWindVector(topology, coord, weather.SeasonSummer).Scale(wind.Speed)
}
