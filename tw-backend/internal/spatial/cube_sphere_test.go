package spatial

import (
	"math"
	"testing"
)

func TestNewCubeSphereTopology(t *testing.T) {
	topo := NewCubeSphereTopology(256)
	if topo.Resolution() != 256 {
		t.Errorf("Resolution() = %d, want 256", topo.Resolution())
	}
}

func TestCubeSphereTopology_GetNeighbor_WithinFace(t *testing.T) {
	topo := NewCubeSphereTopology(64)

	tests := []struct {
		name  string
		start Coordinate
		dir   Direction
		want  Coordinate
	}{
		{"center north", Coordinate{Face: 0, X: 32, Y: 32}, North, Coordinate{Face: 0, X: 32, Y: 31}},
		{"center south", Coordinate{Face: 0, X: 32, Y: 32}, South, Coordinate{Face: 0, X: 32, Y: 33}},
		{"center east", Coordinate{Face: 0, X: 32, Y: 32}, East, Coordinate{Face: 0, X: 33, Y: 32}},
		{"center west", Coordinate{Face: 0, X: 32, Y: 32}, West, Coordinate{Face: 0, X: 31, Y: 32}},
		{"diagonal NE", Coordinate{Face: 0, X: 32, Y: 32}, NorthEast, Coordinate{Face: 0, X: 33, Y: 31}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := topo.GetNeighbor(tc.start, tc.dir)
			if got != tc.want {
				t.Errorf("GetNeighbor(%v, %s) = %v, want %v", tc.start, tc.dir, got, tc.want)
			}
		})
	}
}

func TestCubeSphereTopology_GetNeighbor_FaceTransitions(t *testing.T) {
	// Face layout:
	//         [4: Top]
	// [2: Left][0: Front][3: Right][1: Back]
	//         [5: Bottom]

	topo := NewCubeSphereTopology(64)
	res := 64

	tests := []struct {
		name     string
		start    Coordinate
		dir      Direction
		wantFace int
	}{
		// Front face (0) transitions
		{"front east to right", Coordinate{Face: 0, X: res - 1, Y: 32}, East, 3},
		{"front west to left", Coordinate{Face: 0, X: 0, Y: 32}, West, 2},
		{"front north to top", Coordinate{Face: 0, X: 32, Y: 0}, North, 4},
		{"front south to bottom", Coordinate{Face: 0, X: 32, Y: res - 1}, South, 5},

		// Right face (3) transitions
		{"right east to back", Coordinate{Face: 3, X: res - 1, Y: 32}, East, 1},

		// Top face (4) transitions
		{"top south to front", Coordinate{Face: 4, X: 32, Y: res - 1}, South, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := topo.GetNeighbor(tc.start, tc.dir)
			if got.Face != tc.wantFace {
				t.Errorf("GetNeighbor(%v, %s).Face = %d, want %d", tc.start, tc.dir, got.Face, tc.wantFace)
			}
			// Verify coordinate is within bounds
			if got.X < 0 || got.X >= res || got.Y < 0 || got.Y >= res {
				t.Errorf("GetNeighbor(%v, %s) = %v, coordinate out of bounds [0, %d)", tc.start, tc.dir, got, res)
			}
		})
	}
}

func TestCubeSphereTopology_ToSphere_Normalized(t *testing.T) {
	topo := NewCubeSphereTopology(64)

	// Test various coordinates on all faces
	coords := []Coordinate{
		{Face: 0, X: 0, Y: 0},
		{Face: 0, X: 32, Y: 32},
		{Face: 0, X: 63, Y: 63},
		{Face: 1, X: 32, Y: 32},
		{Face: 2, X: 32, Y: 32},
		{Face: 3, X: 32, Y: 32},
		{Face: 4, X: 32, Y: 32},
		{Face: 5, X: 32, Y: 32},
	}

	for _, c := range coords {
		t.Run("face"+string(rune('0'+c.Face)), func(t *testing.T) {
			x, y, z := topo.ToSphere(c)
			mag := math.Sqrt(x*x + y*y + z*z)
			if math.Abs(mag-1.0) > 1e-9 {
				t.Errorf("ToSphere(%v) magnitude = %f, want 1.0", c, mag)
			}
		})
	}
}

func TestCubeSphereTopology_ToSphere_FaceOrientations(t *testing.T) {
	topo := NewCubeSphereTopology(64)
	center := 32

	tests := []struct {
		name         string
		coord        Coordinate
		dominantAxis string
		expectedSign float64
	}{
		{"front face center", Coordinate{Face: 0, X: center, Y: center}, "z", 1},
		{"back face center", Coordinate{Face: 1, X: center, Y: center}, "z", -1},
		{"left face center", Coordinate{Face: 2, X: center, Y: center}, "x", -1},
		{"right face center", Coordinate{Face: 3, X: center, Y: center}, "x", 1},
		{"top face center", Coordinate{Face: 4, X: center, Y: center}, "y", 1},
		{"bottom face center", Coordinate{Face: 5, X: center, Y: center}, "y", -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x, y, z := topo.ToSphere(tc.coord)

			var dominant float64
			switch tc.dominantAxis {
			case "x":
				dominant = x
			case "y":
				dominant = y
			case "z":
				dominant = z
			}

			if math.Copysign(1, dominant) != tc.expectedSign {
				t.Errorf("ToSphere(%v) %s = %f, expected sign %f", tc.coord, tc.dominantAxis, dominant, tc.expectedSign)
			}
		})
	}
}

func TestCubeSphereTopology_FromVector_FaceDetection(t *testing.T) {
	topo := NewCubeSphereTopology(64)

	tests := []struct {
		name     string
		x, y, z  float64
		wantFace int
	}{
		{"front (+Z)", 0, 0, 1, 0},
		{"back (-Z)", 0, 0, -1, 1},
		{"left (-X)", -1, 0, 0, 2},
		{"right (+X)", 1, 0, 0, 3},
		{"top (+Y)", 0, 1, 0, 4},
		{"bottom (-Y)", 0, -1, 0, 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := topo.FromVector(tc.x, tc.y, tc.z)
			if got.Face != tc.wantFace {
				t.Errorf("FromVector(%f, %f, %f).Face = %d, want %d", tc.x, tc.y, tc.z, got.Face, tc.wantFace)
			}
		})
	}
}

func TestCubeSphereTopology_FromVector_RoundTrip(t *testing.T) {
	topo := NewCubeSphereTopology(64)

	// Test round-trip for center of each face
	for face := 0; face < 6; face++ {
		original := Coordinate{Face: face, X: 32, Y: 32}
		x, y, z := topo.ToSphere(original)
		recovered := topo.FromVector(x, y, z)

		if recovered.Face != original.Face {
			t.Errorf("Face %d: round-trip face mismatch, got %d", face, recovered.Face)
		}
		// Allow tolerance of 1 for floating point
		if abs(recovered.X-original.X) > 1 || abs(recovered.Y-original.Y) > 1 {
			t.Errorf("Face %d: round-trip coord mismatch, original=%v, recovered=%v", face, original, recovered)
		}
	}
}

func TestCubeSphereTopology_Distance_SamePoint(t *testing.T) {
	topo := NewCubeSphereTopology(64)
	coord := Coordinate{Face: 0, X: 32, Y: 32}

	dist := topo.Distance(coord, coord)
	if dist != 0 {
		t.Errorf("Distance to self = %f, want 0", dist)
	}
}

func TestCubeSphereTopology_Distance_Symmetry(t *testing.T) {
	topo := NewCubeSphereTopology(64)
	a := Coordinate{Face: 0, X: 10, Y: 10}
	b := Coordinate{Face: 3, X: 50, Y: 50}

	distAB := topo.Distance(a, b)
	distBA := topo.Distance(b, a)

	if math.Abs(distAB-distBA) > 1e-9 {
		t.Errorf("Distance asymmetry: A→B=%f, B→A=%f", distAB, distBA)
	}
}

func TestCubeSphereTopology_Distance_CrossFace(t *testing.T) {
	topo := NewCubeSphereTopology(64)

	// Adjacent cells should have small distance
	a := Coordinate{Face: 0, X: 32, Y: 32}
	b := Coordinate{Face: 0, X: 33, Y: 32}

	distNear := topo.Distance(a, b)

	// Opposite side of globe should have larger distance
	c := Coordinate{Face: 1, X: 32, Y: 32} // Back face

	distFar := topo.Distance(a, c)

	if distNear >= distFar {
		t.Errorf("Expected near distance (%f) < far distance (%f)", distNear, distFar)
	}
}

// Run contract tests on CubeSphereTopology
func TestCubeSphereTopology_ContractTests(t *testing.T) {
	topo := NewCubeSphereTopology(64)
	ct := NewTopologyContractTest(t, topo)

	t.Run("DistanceNonNegative", func(t *testing.T) {
		ct.TestDistanceNonNegative()
	})

	t.Run("DistanceSymmetry", func(t *testing.T) {
		ct.TestDistanceSymmetry()
	})

	t.Run("DistanceTriangleInequality", func(t *testing.T) {
		ct.TestDistanceTriangleInequality()
	})

	t.Run("ToSphereNormalized", func(t *testing.T) {
		ct.TestToSphereNormalized()
	})

	t.Run("FromVectorRoundTrip", func(t *testing.T) {
		ct.TestFromVectorRoundTrip()
	})

	t.Run("NeighborReversibility", func(t *testing.T) {
		ct.TestNeighborReversibility()
	})
}
