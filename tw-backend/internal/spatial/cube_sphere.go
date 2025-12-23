package spatial

import (
	"math"
)

// Face indices for the cube sphere
const (
	FaceFront  = 0 // +Z
	FaceBack   = 1 // -Z
	FaceLeft   = 2 // -X
	FaceRight  = 3 // +X
	FaceTop    = 4 // +Y
	FaceBottom = 5 // -Y
)

// Edge identifiers
const (
	EdgeNorth = iota // y = 0
	EdgeSouth        // y = resolution - 1
	EdgeEast         // x = resolution - 1
	EdgeWest         // x = 0
)

// faceTransition describes how to transition from one face to another
type faceTransition struct {
	targetFace int
	// transform maps (x, y) from source edge to (newX, newY) on target edge
	// rotationType: 0=none, 1=90°CW, 2=180°, 3=90°CCW
	rotationType int
}

// CubeSphereTopology implements Topology for a cube-sphere projection
type CubeSphereTopology struct {
	resolution int
	// faceConnections[sourceFace][edge] = transition
	faceConnections [6][4]faceTransition
}

// NewCubeSphereTopology creates a new cube sphere topology with the given resolution per face
func NewCubeSphereTopology(resolution int) *CubeSphereTopology {
	if resolution <= 0 {
		resolution = 64
	}

	t := &CubeSphereTopology{
		resolution: resolution,
	}
	t.initFaceConnections()
	return t
}

// initFaceConnections sets up the face adjacency lookup table
// Face layout (unwrapped net):
//
//	         [4: Top]
//	[2: Left][0: Front][3: Right][1: Back]
//	         [5: Bottom]
func (t *CubeSphereTopology) initFaceConnections() {
	// Front face (0) connections
	t.faceConnections[FaceFront][EdgeNorth] = faceTransition{FaceTop, 0}
	t.faceConnections[FaceFront][EdgeSouth] = faceTransition{FaceBottom, 0}
	t.faceConnections[FaceFront][EdgeEast] = faceTransition{FaceRight, 0}
	t.faceConnections[FaceFront][EdgeWest] = faceTransition{FaceLeft, 0}

	// Back face (1) connections
	t.faceConnections[FaceBack][EdgeNorth] = faceTransition{FaceTop, 2}
	t.faceConnections[FaceBack][EdgeSouth] = faceTransition{FaceBottom, 2}
	t.faceConnections[FaceBack][EdgeEast] = faceTransition{FaceLeft, 0}
	t.faceConnections[FaceBack][EdgeWest] = faceTransition{FaceRight, 0}

	// Left face (2) connections
	t.faceConnections[FaceLeft][EdgeNorth] = faceTransition{FaceTop, 3}
	t.faceConnections[FaceLeft][EdgeSouth] = faceTransition{FaceBottom, 1}
	t.faceConnections[FaceLeft][EdgeEast] = faceTransition{FaceFront, 0}
	t.faceConnections[FaceLeft][EdgeWest] = faceTransition{FaceBack, 0}

	// Right face (3) connections
	t.faceConnections[FaceRight][EdgeNorth] = faceTransition{FaceTop, 1}
	t.faceConnections[FaceRight][EdgeSouth] = faceTransition{FaceBottom, 3}
	t.faceConnections[FaceRight][EdgeEast] = faceTransition{FaceBack, 0}
	t.faceConnections[FaceRight][EdgeWest] = faceTransition{FaceFront, 0}

	// Top face (4) connections
	t.faceConnections[FaceTop][EdgeNorth] = faceTransition{FaceBack, 2}
	t.faceConnections[FaceTop][EdgeSouth] = faceTransition{FaceFront, 0}
	t.faceConnections[FaceTop][EdgeEast] = faceTransition{FaceRight, 3}
	t.faceConnections[FaceTop][EdgeWest] = faceTransition{FaceLeft, 1}

	// Bottom face (5) connections
	t.faceConnections[FaceBottom][EdgeNorth] = faceTransition{FaceFront, 0}
	t.faceConnections[FaceBottom][EdgeSouth] = faceTransition{FaceBack, 2}
	t.faceConnections[FaceBottom][EdgeEast] = faceTransition{FaceRight, 1}
	t.faceConnections[FaceBottom][EdgeWest] = faceTransition{FaceLeft, 3}
}

// Resolution returns the grid size of each face
func (t *CubeSphereTopology) Resolution() int {
	return t.resolution
}

// GetNeighbor returns the coordinate one step in the given direction
func (t *CubeSphereTopology) GetNeighbor(coord Coordinate, direction Direction) Coordinate {
	dx, dy := DirectionDelta(direction)
	newX := coord.X + dx
	newY := coord.Y + dy

	// Check if we stay within the same face
	if newX >= 0 && newX < t.resolution && newY >= 0 && newY < t.resolution {
		return Coordinate{Face: coord.Face, X: newX, Y: newY}
	}

	// Determine which edge we're crossing
	var edge int
	var edgePos int

	if newY < 0 {
		edge = EdgeNorth
		edgePos = newX
		newY = 0
	} else if newY >= t.resolution {
		edge = EdgeSouth
		edgePos = newX
		newY = t.resolution - 1
	} else if newX >= t.resolution {
		edge = EdgeEast
		edgePos = newY
		newX = t.resolution - 1
	} else { // newX < 0
		edge = EdgeWest
		edgePos = newY
		newX = 0
	}

	trans := t.faceConnections[coord.Face][edge]
	targetX, targetY := t.transformCoordinate(edge, edgePos, trans.rotationType)

	return Coordinate{Face: trans.targetFace, X: targetX, Y: targetY}
}

// transformCoordinate applies rotation and maps edge position to target face coordinates
func (t *CubeSphereTopology) transformCoordinate(sourceEdge, edgePos, rotationType int) (int, int) {
	max := t.resolution - 1

	// First, determine where on the target face we enter based on source edge and rotation
	// This is complex because each rotation affects how we enter the target face

	switch rotationType {
	case 0: // No rotation
		switch sourceEdge {
		case EdgeNorth:
			return edgePos, max // Enter from south edge of target
		case EdgeSouth:
			return edgePos, 0 // Enter from north edge of target
		case EdgeEast:
			return 0, edgePos // Enter from west edge of target
		case EdgeWest:
			return max, edgePos // Enter from east edge of target
		}
	case 1: // 90° clockwise
		switch sourceEdge {
		case EdgeNorth:
			return max, edgePos // Enter from east edge
		case EdgeSouth:
			return 0, max - edgePos // Enter from west edge
		case EdgeEast:
			return edgePos, 0 // Enter from north edge
		case EdgeWest:
			return max - edgePos, max // Enter from south edge
		}
	case 2: // 180° rotation
		switch sourceEdge {
		case EdgeNorth:
			return max - edgePos, 0 // Enter from north edge, reversed
		case EdgeSouth:
			return max - edgePos, max // Enter from south edge, reversed
		case EdgeEast:
			return max, max - edgePos // Enter from east edge, reversed
		case EdgeWest:
			return 0, max - edgePos // Enter from west edge, reversed
		}
	case 3: // 90° counter-clockwise
		switch sourceEdge {
		case EdgeNorth:
			return 0, max - edgePos // Enter from west edge
		case EdgeSouth:
			return max, edgePos // Enter from east edge
		case EdgeEast:
			return max - edgePos, max // Enter from south edge
		case EdgeWest:
			return edgePos, 0 // Enter from north edge
		}
	}

	// Fallback (shouldn't reach here)
	return edgePos, 0
}

// ToSphere converts a face coordinate to a unit sphere vector (x, y, z)
// Uses normalized cube mapping to reduce distortion
func (t *CubeSphereTopology) ToSphere(coord Coordinate) (x, y, z float64) {
	// Convert grid coordinates to [-1, 1] range
	u := (float64(coord.X)+0.5)/float64(t.resolution)*2 - 1
	v := (float64(coord.Y)+0.5)/float64(t.resolution)*2 - 1

	// Map to cube face
	switch coord.Face {
	case FaceFront: // +Z
		x, y, z = u, -v, 1
	case FaceBack: // -Z
		x, y, z = -u, -v, -1
	case FaceLeft: // -X
		x, y, z = -1, -v, u
	case FaceRight: // +X
		x, y, z = 1, -v, -u
	case FaceTop: // +Y
		x, y, z = u, 1, v
	case FaceBottom: // -Y
		x, y, z = u, -1, -v
	}

	// Normalize to unit sphere
	mag := math.Sqrt(x*x + y*y + z*z)
	return x / mag, y / mag, z / mag
}

// FromVector converts a unit sphere vector to a face coordinate
func (t *CubeSphereTopology) FromVector(x, y, z float64) Coordinate {
	// Normalize input vector
	mag := math.Sqrt(x*x + y*y + z*z)
	if mag > 0 {
		x, y, z = x/mag, y/mag, z/mag
	}

	// Determine which face based on dominant axis
	absX, absY, absZ := math.Abs(x), math.Abs(y), math.Abs(z)

	var face int
	var u, v float64

	if absZ >= absX && absZ >= absY {
		if z > 0 {
			face = FaceFront
			u, v = x/z, -y/z
		} else {
			face = FaceBack
			u, v = -x/(-z), -y/(-z)
		}
	} else if absX >= absY {
		if x > 0 {
			face = FaceRight
			u, v = -z/x, -y/x
		} else {
			face = FaceLeft
			u, v = z/(-x), -y/(-x)
		}
	} else {
		if y > 0 {
			face = FaceTop
			u, v = x/y, z/y
		} else {
			face = FaceBottom
			u, v = x/(-y), -z/(-y)
		}
	}

	// Convert from [-1, 1] to grid coordinates
	gridX := int((u + 1) / 2 * float64(t.resolution))
	gridY := int((v + 1) / 2 * float64(t.resolution))

	// Clamp to valid range
	if gridX < 0 {
		gridX = 0
	} else if gridX >= t.resolution {
		gridX = t.resolution - 1
	}
	if gridY < 0 {
		gridY = 0
	} else if gridY >= t.resolution {
		gridY = t.resolution - 1
	}

	return Coordinate{Face: face, X: gridX, Y: gridY}
}

// Distance returns the physical distance (great circle) between two coordinates
// Returns distance as a fraction of the sphere's circumference (0 to 0.5)
func (t *CubeSphereTopology) Distance(a, b Coordinate) float64 {
	if a == b {
		return 0
	}

	// Convert to sphere vectors
	ax, ay, az := t.ToSphere(a)
	bx, by, bz := t.ToSphere(b)

	// Great circle distance using dot product
	// cos(angle) = a · b (for unit vectors)
	dot := ax*bx + ay*by + az*bz

	// Clamp to [-1, 1] to avoid floating point errors in Acos
	if dot > 1 {
		dot = 1
	} else if dot < -1 {
		dot = -1
	}

	// Return angle in radians (0 to π)
	return math.Acos(dot)
}
