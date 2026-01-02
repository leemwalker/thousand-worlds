package geography

import (
	"math/rand"

	"tw-backend/internal/spatial"
)

// DriftEngine handles the movement and geometric updates of tectonic plates.
// Implements Phase 8a of the Deep Time Tectonics plan.

// UpdatePlatePositions moves all plate centroids based on their angular velocity.
// dt is the time delta in ticks.
func UpdatePlatePositions(plates []TectonicPlate, dt float64, topology spatial.Topology) {
	for i := range plates {
		plate := &plates[i]

		if plate.AngularSpeed == 0 {
			continue // Static plate (rare/impossible in real sim)
		}

		// 1. Rotate the Centroid Position Vector
		// NewPosition = Rotate(OldPosition, Axis, Angle)
		angle := plate.AngularSpeed * dt
		newPos := plate.Position.RotateAround(plate.RotationAxis, angle)
		plate.Position = newPos

		// 2. Update the discrete Grid Coordinate (Centroid)
		// We map the 3D vector back to the nearest face coordinate.
		plate.Centroid = topology.FromVector(newPos.X, newPos.Y, newPos.Z)

		// 3. Update the Linear Velocity Vector (for collision calc)
		// Velocity = Tangent vector at this position in direction of rotation
		// V = ω x r (Cross product of rotation vector and position vector)
		// Rotation Vector Ω = Axis * Speed
		omega := plate.RotationAxis.Scale(plate.AngularSpeed)
		// r = Position (unit vector if radius is 1, but we deal with direction)
		// V = Ω x r
		plate.Velocity = omega.Cross(plate.Position).Normalize() // Normalize for direction, or keep magnitude?
		// Existing logic uses Velocity as a direction vector.
		// Let's keep Velocity as the linear direction vector at the centroid.
	}
}

// RecalculateRegions clears existing plate regions and re-assigns every cell
// on the sphere to the nearest plate centroid using Multi-Source BFS (Voronoi).
func RecalculateRegions(plates []TectonicPlate, topology spatial.Topology) {
	// reuse the existing ReassignPlateRegions function in tectonics.go
	// Since it's in the same package 'geography', we can call it directly.
	ReassignPlateRegions(plates, topology)
}

// InitializePlateMotion assigns random rotation axes and speeds to plates.
// Should be called during generation.
func InitializePlateMotion(plates []TectonicPlate, rng *rand.Rand) {
	for i := range plates {
		// Random Axis on sphere
		plates[i].RotationAxis = spatial.RandomPointOnSphere(rng.Int63())

		// Random Speed (Radians per tick)
		// Earth plates move ~1-10cm/year.
		// If 1 tick = 1 Million Years.
		// Circumference = 40,000 km.
		// 1cm/yr = 10km/Myr.
		// 10km / 40,000km = 1/4000 of a rotation per tick.
		// 2π / 4000 ≈ 0.0015 radians/tick.
		// Let's go slightly faster for simulation effect: 0.005 - 0.02 radians/tick.
		speed := 0.005 + rng.Float64()*0.015
		plates[i].AngularSpeed = speed

		// Initialize Velocity vector based on this motion
		omega := plates[i].RotationAxis.Scale(speed)
		plates[i].Velocity = omega.Cross(plates[i].Position).Normalize()
	}
}
