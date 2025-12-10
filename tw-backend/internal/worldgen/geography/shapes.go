package geography

import (
	"math"
)

// WorldShapeType defines the geometry of the world
type WorldShapeType string

const (
	ShapeSpherical WorldShapeType = "spherical"
	ShapeBounded   WorldShapeType = "bounded"
	ShapeInfinite  WorldShapeType = "infinite"
)

// WorldShape handles coordinate wrapping and distance calculations
type WorldShape interface {
	// Distance returns the distance between two points
	Distance(p1, p2 Point) float64
	// WrapCoordinates returns the valid coordinates for a point, handling wrapping
	WrapCoordinates(p Point) Point
	// IsValid checks if a point is within the world boundaries
	IsValid(p Point) bool
}

// BoundedShape represents a flat world with hard edges
type BoundedShape struct {
	Width, Height float64
}

func (s *BoundedShape) Distance(p1, p2 Point) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2))
}

func (s *BoundedShape) WrapCoordinates(p Point) Point {
	// Clamp to boundaries
	x := math.Max(0, math.Min(s.Width-1, p.X))
	y := math.Max(0, math.Min(s.Height-1, p.Y))
	return Point{X: x, Y: y}
}

func (s *BoundedShape) IsValid(p Point) bool {
	return p.X >= 0 && p.X < s.Width && p.Y >= 0 && p.Y < s.Height
}

// SphericalShape represents a world that wraps east-west (and potentially north-south or clamps)
// For simplicity, we'll implement cylindrical wrapping (East-West wraps, North-South clamps)
// A true sphere is complex on a 2D grid, but this is standard for games (e.g. Civ)
type SphericalShape struct {
	Width, Height float64
}

func (s *SphericalShape) Distance(p1, p2 Point) float64 {
	// Calculate dx handling wrapping
	dx := math.Abs(p1.X - p2.X)
	if dx > s.Width/2 {
		dx = s.Width - dx
	}
	dy := math.Abs(p1.Y - p2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

func (s *SphericalShape) WrapCoordinates(p Point) Point {
	// Wrap X
	x := math.Mod(p.X, s.Width)
	if x < 0 {
		x += s.Width
	}

	// Clamp Y (Poles)
	y := math.Max(0, math.Min(s.Height-1, p.Y))

	return Point{X: x, Y: y}
}

func (s *SphericalShape) IsValid(p Point) bool {
	// Y must be in bounds, X is always valid due to wrapping (conceptually)
	// But for array access, we need to wrap it first.
	// This function checks if the *raw* coordinate is "on the map" in a way that makes sense?
	// Actually, for spherical, any X is valid, but Y is bounded.
	return p.Y >= 0 && p.Y < s.Height
}

// InfiniteShape represents a world generated on demand (chunked)
// For this interface, we assume local coordinates are always valid
type InfiniteShape struct {
}

func (s *InfiniteShape) Distance(p1, p2 Point) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2))
}

func (s *InfiniteShape) WrapCoordinates(p Point) Point {
	return p // No wrapping
}

func (s *InfiniteShape) IsValid(p Point) bool {
	return true
}

// GetShape returns the appropriate WorldShape implementation
func GetShape(shapeType WorldShapeType, width, height int) WorldShape {
	w, h := float64(width), float64(height)
	switch shapeType {
	case ShapeSpherical:
		return &SphericalShape{Width: w, Height: h}
	case ShapeBounded:
		return &BoundedShape{Width: w, Height: h}
	case ShapeInfinite:
		return &InfiniteShape{}
	default:
		return &BoundedShape{Width: w, Height: h}
	}
}
