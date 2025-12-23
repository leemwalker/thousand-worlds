package weather

import (
	"math"

	"tw-backend/internal/spatial"
)

// GetLatitudeFromCoord derives latitude (-90 to 90 degrees) from a spherical coordinate.
// Uses the Y component of the unit sphere vector: sin(latitude) = y
func GetLatitudeFromCoord(topology spatial.Topology, coord spatial.Coordinate) float64 {
	_, y, _ := topology.ToSphere(coord)
	// y component of unit sphere = sin(latitude)
	latRad := math.Asin(y)
	return latRad * 180 / math.Pi
}

// GetLongitudeFromCoord derives longitude (-180 to 180 degrees) from a spherical coordinate.
// Uses atan2(x, z) for the angle in the XZ plane.
func GetLongitudeFromCoord(topology spatial.Topology, coord spatial.Coordinate) float64 {
	x, _, z := topology.ToSphere(coord)
	// Longitude is angle in XZ plane from +Z axis
	lonRad := math.Atan2(x, z)
	return lonRad * 180 / math.Pi
}

// WindToLocalDirection converts a 3D world-space wind vector to a local grid direction.
// This handles the rotation needed when wind crosses face boundaries.
func WindToLocalDirection(topology spatial.Topology, coord spatial.Coordinate, windVec spatial.Vector3D) spatial.Direction {
	// Get the surface normal and tangent vectors at this coordinate
	// For cube mapping, we project the wind onto the face's local axes

	// Get the 3D position on sphere
	px, py, pz := topology.ToSphere(coord)

	// Surface normal at this point (same as position for unit sphere)
	normal := spatial.Vector3D{X: px, Y: py, Z: pz}

	// Project wind onto tangent plane (remove component along normal)
	dot := windVec.Dot(normal)
	tangentWind := windVec.Add(normal.Scale(-dot))

	// Now we need to determine which local direction this corresponds to
	// We'll define local axes based on the face

	// Local "up" direction on the face (towards +Y in world space, projected)
	up := spatial.Vector3D{X: 0, Y: 1, Z: 0}
	upDot := up.Dot(normal)
	localUp := up.Add(normal.Scale(-upDot)).Normalize()

	// Local "right" is cross product of up and normal (order matters!)
	// Using right-hand rule: up × normal = right
	localRight := spatial.Vector3D{
		X: localUp.Y*normal.Z - localUp.Z*normal.Y,
		Y: localUp.Z*normal.X - localUp.X*normal.Z,
		Z: localUp.X*normal.Y - localUp.Y*normal.X,
	}.Normalize()

	// Project tangent wind onto local axes
	rightComponent := tangentWind.Dot(localRight)
	upComponent := tangentWind.Dot(localUp)

	// Determine dominant direction
	if math.Abs(rightComponent) > math.Abs(upComponent) {
		if rightComponent > 0 {
			return spatial.East
		}
		return spatial.West
	}
	if upComponent > 0 {
		return spatial.North
	}
	return spatial.South
}

// SimulateAdvectionSpherical moves moisture in the wind direction using spherical topology.
// Returns the new coordinate, the rotated wind vector (in local space), and the transported moisture.
func SimulateAdvectionSpherical(
	topology spatial.Topology,
	coord spatial.Coordinate,
	windVec spatial.Vector3D,
	moisture float64,
) (newCoord spatial.Coordinate, rotatedWind spatial.Vector3D, transportedMoisture float64) {
	// Get the local direction from the world-space wind
	localDir := WindToLocalDirection(topology, coord, windVec)

	// Move to neighbor in that direction
	newCoord = topology.GetNeighbor(coord, localDir)

	// The wind vector remains the same in world space
	// (it will be re-interpreted on the new face)
	rotatedWind = windVec

	// Moisture is transported unchanged
	transportedMoisture = moisture

	return newCoord, rotatedWind, transportedMoisture
}

// GetUpwindCoords returns the coordinates upwind from the given position.
// Useful for precipitation calculations.
func GetUpwindCoords(
	topology spatial.Topology,
	coord spatial.Coordinate,
	windVec spatial.Vector3D,
	count int,
) []spatial.Coordinate {
	result := make([]spatial.Coordinate, 0, count)

	// Upwind is opposite of wind direction
	upwindVec := windVec.Scale(-1)
	current := coord

	for i := 0; i < count; i++ {
		dir := WindToLocalDirection(topology, current, upwindVec)
		upwindCoord := topology.GetNeighbor(current, dir)
		result = append(result, upwindCoord)
		current = upwindCoord
	}

	return result
}
