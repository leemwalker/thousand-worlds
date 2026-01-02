package geography

import (
	"tw-backend/internal/spatial"
)

// CrustDynamics handles the geological consequences of plate movement.
// Implements boundary updates and basic crust tracking.

// TectonicMap (conceptually part of WorldGeology, but helper structs here)

// UpdateBoundaries identifies the type of boundary between every pair of neighboring cells
// belonging to different plates.
//
// Returns a map of cell -> BoundaryType (Convergent, Divergent, Transform)
// UpdateBoundaries identifies the type of boundary between every pair of neighboring cells
// belonging to different plates.
//
// Returns a map of cell -> BoundaryType (Convergent, Divergent, Transform)
func UpdateBoundaries(plates []TectonicPlate, topology spatial.Topology) map[spatial.Coordinate]BoundaryType {
	boundaryMap := make(map[spatial.Coordinate]BoundaryType)

	// Pre-compute plate velocities maps or access directly?
	// Access directly via plates slice index.

	// Helper to get plate by ID or Index.
	// Our regions map cells to plate Index in ReassignPlateRegions usually,
	// but TectonicPlate.Region is a set of coords.
	// We need a quick lookup: Coord -> PlateIndex.
	ownerMap := make(map[spatial.Coordinate]int)
	for idx, p := range plates {
		for cell := range p.Region {
			ownerMap[cell] = idx
		}
	}

	// Iterate all cells
	// Optimization: Only iterate cells at the edge of regions?
	// For now, iterate all plate regions (which is all cells).

	directions := []spatial.Direction{
		spatial.North, spatial.South, spatial.East, spatial.West,
	}

	for coord, plateIdx := range ownerMap {
		myPlate := &plates[plateIdx]

		isBoundary := false

		for _, dir := range directions {
			neighbor := topology.GetNeighbor(coord, dir)
			neighborIdx, ok := ownerMap[neighbor]

			if ok && neighborIdx != plateIdx {
				// Found a boundary!
				isBoundary = true
				otherPlate := &plates[neighborIdx]

				// Determine Boundary Type
				bType := CalculateBoundaryVector(myPlate.Velocity, otherPlate.Velocity, myPlate.Position, otherPlate.Position)

				// Store biggest impact? Or just overwrite?
				// Usually Convergent > Divergent > Transform priority
				boundaryMap[coord] = bType
			}
		}

		if !isBoundary {
			// Internal cell
		}
	}

	return boundaryMap
}

// CalculateBoundaryVector determines if two plates are converging, diverging, or sliding
// based on their velocity vectors at the boundary.
// (Replaces/Enhances logic in tectonics.go)
func CalculateBoundaryVector(v1, v2, pos1, pos2 spatial.Vector3D) BoundaryType {
	// Relative velocity
	relV := v1.Sub(v2)

	// Vector pointing from plate 1 to plate 2
	direction := pos2.Sub(pos1).Normalize()

	// Dot product:
	// If positive: moving towards each other? No, relV = v1 - v2.
	// if v1 is towards v2, and v2 is towards v1 (opposing), relV is large towards v2.
	// Let's use simple divergence check.
	// Divergence = Dot(v1, dir) - Dot(v2, dir) ??

	// Simpler: Dot product of Relative Velocity and Direction Vector.
	// If v1 moves towards pos2 and v2 moves towards pos1...
	// v1 dot dir > 0. v2 dot dir < 0.
	// (v1 - v2) dot dir = positive - negative = positive.

	magnitude := relV.Dot(direction)

	if magnitude > 0.1 {
		return BoundaryConvergent
	} else if magnitude < -0.1 {
		return BoundaryDivergent
	} else {
		return BoundaryTransform
	}
}
