package geography

import (
	"testing"
)

func TestBoundedShape(t *testing.T) {
	shape := GetShape(ShapeBounded, 100, 100)

	p1 := Point{X: 10, Y: 10}
	p2 := Point{X: 20, Y: 10}

	// Distance
	if d := shape.Distance(p1, p2); d != 10 {
		t.Errorf("Expected distance 10, got %f", d)
	}

	// Wrap (Clamp)
	out := Point{X: 150, Y: 150}
	wrapped := shape.WrapCoordinates(out)
	if wrapped.X != 99 || wrapped.Y != 99 {
		t.Errorf("Expected clamped to 99,99, got %f,%f", wrapped.X, wrapped.Y)
	}
}

func TestSphericalShape(t *testing.T) {
	shape := GetShape(ShapeSpherical, 100, 100)

	// Wrapping distance
	p1 := Point{X: 10, Y: 50}
	p2 := Point{X: 90, Y: 50} // 20 units away across the seam

	if d := shape.Distance(p1, p2); d != 20 {
		t.Errorf("Expected wrapping distance 20, got %f", d)
	}

	// Wrap coordinates
	out := Point{X: 110, Y: 50}
	wrapped := shape.WrapCoordinates(out)
	if wrapped.X != 10 {
		t.Errorf("Expected wrapped X to 10, got %f", wrapped.X)
	}

	// Clamp Y
	outY := Point{X: 50, Y: 150}
	wrappedY := shape.WrapCoordinates(outY)
	if wrappedY.Y != 99 {
		t.Errorf("Expected clamped Y to 99, got %f", wrappedY.Y)
	}
}
