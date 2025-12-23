package spatial

import (
	"math"
	"math/rand"
)

// Coordinate represents a position on a cube-sphere topology.
type Coordinate struct {
	Face int // 0-5 representing the 6 cube faces
	X, Y int // Position within the face grid
}

// Vector3D represents a 3D vector for wind direction, movement, etc.
type Vector3D struct {
	X, Y, Z float64
}

// Normalize returns a unit vector in the same direction
func (v Vector3D) Normalize() Vector3D {
	mag := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if mag == 0 {
		return Vector3D{}
	}
	return Vector3D{X: v.X / mag, Y: v.Y / mag, Z: v.Z / mag}
}

// Length returns the magnitude of the vector
func (v Vector3D) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Dot returns the dot product of two vectors
func (v Vector3D) Dot(other Vector3D) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross returns the cross product of two vectors
func (v Vector3D) Cross(other Vector3D) Vector3D {
	return Vector3D{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

// Scale returns the vector multiplied by a scalar
func (v Vector3D) Scale(s float64) Vector3D {
	return Vector3D{X: v.X * s, Y: v.Y * s, Z: v.Z * s}
}

// Add returns the sum of two vectors
func (v Vector3D) Add(other Vector3D) Vector3D {
	return Vector3D{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

// Sub returns the difference of two vectors (v - other)
func (v Vector3D) Sub(other Vector3D) Vector3D {
	return Vector3D{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

// Distance returns the Euclidean distance between two points
func (v Vector3D) Distance(other Vector3D) float64 {
	return v.Sub(other).Length()
}

// RandomPointOnSphere generates a uniformly distributed random point on a unit sphere.
// Uses the Gaussian/Marsaglia method for uniform sphere coverage.
func RandomPointOnSphere(seed int64) Vector3D {
	r := rand.New(rand.NewSource(seed))

	// Gaussian distribution method: three independent normal samples
	// normalized to unit length gives uniform distribution on sphere
	x := r.NormFloat64()
	y := r.NormFloat64()
	z := r.NormFloat64()

	return Vector3D{X: x, Y: y, Z: z}.Normalize()
}

// DirectionDelta returns the (dx, dy) movement delta for a direction.
// Uses delta-based movement for extensibility to diagonals.
func DirectionDelta(d Direction) (dx, dy int) {
	switch d {
	case North:
		return 0, -1
	case South:
		return 0, 1
	case East:
		return 1, 0
	case West:
		return -1, 0
	case NorthEast:
		return 1, -1
	case SouthEast:
		return 1, 1
	case SouthWest:
		return -1, 1
	case NorthWest:
		return -1, -1
	default:
		return 0, 0
	}
}

// Topology abstracts world surface adjacency, allowing simulations
// to work independently of the underlying grid shape.
type Topology interface {
	// GetNeighbor returns the coordinate one step in the given direction.
	// Handles cross-face transitions and coordinate rotation.
	GetNeighbor(coord Coordinate, direction Direction) Coordinate

	// Distance returns the physical distance (great circle approximation)
	// between two coordinates on the sphere.
	Distance(a, b Coordinate) float64

	// ToSphere converts a face coordinate to a unit sphere vector (x, y, z).
	// Uses normalized cube mapping: v / ||v|| where v = (u, v, 1).
	ToSphere(coord Coordinate) (x, y, z float64)

	// FromVector converts a unit sphere vector to a face coordinate.
	// Essential for wind vector "landing" and user interaction.
	FromVector(x, y, z float64) Coordinate

	// Resolution returns the grid size of each face.
	Resolution() int
}
