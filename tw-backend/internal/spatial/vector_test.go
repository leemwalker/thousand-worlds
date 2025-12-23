package spatial

import (
	"math"
	"testing"
)

func TestVector3D_Cross(t *testing.T) {
	tests := []struct {
		name string
		a, b Vector3D
		want Vector3D
	}{
		{
			name: "X cross Y = Z",
			a:    Vector3D{1, 0, 0},
			b:    Vector3D{0, 1, 0},
			want: Vector3D{0, 0, 1},
		},
		{
			name: "Y cross Z = X",
			a:    Vector3D{0, 1, 0},
			b:    Vector3D{0, 0, 1},
			want: Vector3D{1, 0, 0},
		},
		{
			name: "Z cross X = Y",
			a:    Vector3D{0, 0, 1},
			b:    Vector3D{1, 0, 0},
			want: Vector3D{0, 1, 0},
		},
		{
			name: "parallel vectors = zero",
			a:    Vector3D{1, 0, 0},
			b:    Vector3D{2, 0, 0},
			want: Vector3D{0, 0, 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.Cross(tc.b)
			if !vectorsEqual(got, tc.want, 1e-9) {
				t.Errorf("Cross() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestVector3D_Sub(t *testing.T) {
	a := Vector3D{3, 5, 7}
	b := Vector3D{1, 2, 3}
	got := a.Sub(b)
	want := Vector3D{2, 3, 4}

	if got != want {
		t.Errorf("Sub() = %v, want %v", got, want)
	}
}

func TestVector3D_Length(t *testing.T) {
	tests := []struct {
		name string
		v    Vector3D
		want float64
	}{
		{"unit X", Vector3D{1, 0, 0}, 1.0},
		{"unit Y", Vector3D{0, 1, 0}, 1.0},
		{"3-4-5 triangle", Vector3D{3, 4, 0}, 5.0},
		{"3D pythagorean", Vector3D{1, 2, 2}, 3.0},
		{"zero vector", Vector3D{0, 0, 0}, 0.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.v.Length()
			if math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("Length() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestVector3D_Distance(t *testing.T) {
	a := Vector3D{0, 0, 0}
	b := Vector3D{3, 4, 0}
	got := a.Distance(b)
	want := 5.0

	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Distance() = %v, want %v", got, want)
	}

	// Distance should be symmetric
	if math.Abs(b.Distance(a)-want) > 1e-9 {
		t.Errorf("Distance() not symmetric")
	}
}

func TestRandomPointOnSphere(t *testing.T) {
	// Test that random points are on the unit sphere
	for i := int64(0); i < 100; i++ {
		p := RandomPointOnSphere(i)
		length := p.Length()

		if math.Abs(length-1.0) > 1e-9 {
			t.Errorf("RandomPointOnSphere(%d) has length %v, want 1.0", i, length)
		}
	}
}

func TestRandomPointOnSphere_DifferentSeeds(t *testing.T) {
	// Different seeds should produce different points
	p1 := RandomPointOnSphere(42)
	p2 := RandomPointOnSphere(43)

	if vectorsEqual(p1, p2, 1e-9) {
		t.Error("Different seeds produced identical points")
	}
}

func TestRandomPointOnSphere_SameSeed(t *testing.T) {
	// Same seed should produce same point (deterministic)
	p1 := RandomPointOnSphere(42)
	p2 := RandomPointOnSphere(42)

	if !vectorsEqual(p1, p2, 1e-9) {
		t.Errorf("Same seed produced different points: %v vs %v", p1, p2)
	}
}

// vectorsEqual checks if two vectors are equal within tolerance
func vectorsEqual(a, b Vector3D, tol float64) bool {
	return math.Abs(a.X-b.X) < tol &&
		math.Abs(a.Y-b.Y) < tol &&
		math.Abs(a.Z-b.Z) < tol
}
