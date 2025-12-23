package spatial

import (
	"math"
	"testing"
)

// TestDirectionDelta verifies direction to delta conversion
func TestDirectionDelta(t *testing.T) {
	tests := []struct {
		dir    Direction
		dx, dy int
	}{
		{North, 0, -1},
		{South, 0, 1},
		{East, 1, 0},
		{West, -1, 0},
		{NorthEast, 1, -1},
		{SouthEast, 1, 1},
		{SouthWest, -1, 1},
		{NorthWest, -1, -1},
	}

	for _, tc := range tests {
		t.Run(string(tc.dir), func(t *testing.T) {
			dx, dy := DirectionDelta(tc.dir)
			if dx != tc.dx || dy != tc.dy {
				t.Errorf("DirectionDelta(%q): got (%d, %d), want (%d, %d)", tc.dir, dx, dy, tc.dx, tc.dy)
			}
		})
	}
}

// TestDirectionDelta_Unknown verifies unknown direction returns zero delta
func TestDirectionDelta_Unknown(t *testing.T) {
	dx, dy := DirectionDelta("UNKNOWN")
	if dx != 0 || dy != 0 {
		t.Errorf("DirectionDelta(unknown): got (%d, %d), want (0, 0)", dx, dy)
	}
}

// TestCoordinate_Equality verifies coordinate comparison
func TestCoordinate_Equality(t *testing.T) {
	tests := []struct {
		name  string
		a, b  Coordinate
		equal bool
	}{
		{
			name:  "same coordinates are equal",
			a:     Coordinate{Face: 0, X: 5, Y: 10},
			b:     Coordinate{Face: 0, X: 5, Y: 10},
			equal: true,
		},
		{
			name:  "different face not equal",
			a:     Coordinate{Face: 0, X: 5, Y: 10},
			b:     Coordinate{Face: 1, X: 5, Y: 10},
			equal: false,
		},
		{
			name:  "different X not equal",
			a:     Coordinate{Face: 0, X: 5, Y: 10},
			b:     Coordinate{Face: 0, X: 6, Y: 10},
			equal: false,
		},
		{
			name:  "different Y not equal",
			a:     Coordinate{Face: 0, X: 5, Y: 10},
			b:     Coordinate{Face: 0, X: 5, Y: 11},
			equal: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a == tc.b
			if got != tc.equal {
				t.Errorf("Coordinate equality: got %v, want %v", got, tc.equal)
			}
		})
	}
}

// TopologyContractTest provides reusable contract tests for any Topology implementation
type TopologyContractTest struct {
	t        *testing.T
	topology Topology
}

// NewTopologyContractTest creates a contract test suite
func NewTopologyContractTest(t *testing.T, topology Topology) *TopologyContractTest {
	return &TopologyContractTest{t: t, topology: topology}
}

// TestDistanceNonNegative verifies Distance is always >= 0
func (ct *TopologyContractTest) TestDistanceNonNegative() {
	ct.t.Helper()
	res := ct.topology.Resolution()
	coords := []Coordinate{
		{Face: 0, X: 0, Y: 0},
		{Face: 1, X: res / 4, Y: res / 4},
		{Face: 4, X: res / 2, Y: res / 2},
	}

	for _, a := range coords {
		for _, b := range coords {
			dist := ct.topology.Distance(a, b)
			if dist < 0 {
				ct.t.Errorf("Distance(%v, %v) = %f, want >= 0", a, b, dist)
			}
		}
	}
}

// TestDistanceSymmetry verifies Distance(a, b) == Distance(b, a)
func (ct *TopologyContractTest) TestDistanceSymmetry() {
	ct.t.Helper()
	res := ct.topology.Resolution()
	coords := []Coordinate{
		{Face: 0, X: 0, Y: 0},
		{Face: 2, X: res / 4, Y: res / 4},
		{Face: 4, X: res / 2, Y: res / 2},
	}

	for _, a := range coords {
		for _, b := range coords {
			distAB := ct.topology.Distance(a, b)
			distBA := ct.topology.Distance(b, a)
			if math.Abs(distAB-distBA) > 1e-9 {
				ct.t.Errorf("Distance asymmetry: Distance(%v, %v)=%f != Distance(%v, %v)=%f",
					a, b, distAB, b, a, distBA)
			}
		}
	}
}

// TestDistanceTriangleInequality verifies Distance(a, c) <= Distance(a, b) + Distance(b, c)
func (ct *TopologyContractTest) TestDistanceTriangleInequality() {
	ct.t.Helper()
	res := ct.topology.Resolution()
	a := Coordinate{Face: 0, X: 0, Y: 0}
	b := Coordinate{Face: 0, X: res / 4, Y: res / 4}
	c := Coordinate{Face: 0, X: res / 2, Y: 0}

	distAC := ct.topology.Distance(a, c)
	distAB := ct.topology.Distance(a, b)
	distBC := ct.topology.Distance(b, c)

	if distAC > distAB+distBC+1e-9 {
		ct.t.Errorf("Triangle inequality violated: Distance(a,c)=%f > Distance(a,b)+Distance(b,c)=%f",
			distAC, distAB+distBC)
	}
}

// TestToSphereNormalized verifies ToSphere produces unit vectors
func (ct *TopologyContractTest) TestToSphereNormalized() {
	ct.t.Helper()
	res := ct.topology.Resolution()
	coords := []Coordinate{
		{Face: 0, X: 0, Y: 0},
		{Face: 0, X: res / 2, Y: res / 2},
		{Face: 3, X: res - 1, Y: res - 1},
		{Face: 4, X: res / 2, Y: res / 2}, // Top face center
	}

	for _, c := range coords {
		x, y, z := ct.topology.ToSphere(c)
		mag := math.Sqrt(x*x + y*y + z*z)
		if math.Abs(mag-1.0) > 1e-9 {
			ct.t.Errorf("ToSphere(%v) magnitude = %f, want 1.0", c, mag)
		}
	}
}

// TestFromVectorRoundTrip verifies ToSphere -> FromVector round-trip
func (ct *TopologyContractTest) TestFromVectorRoundTrip() {
	ct.t.Helper()
	res := ct.topology.Resolution()
	// Test center of each face (should round-trip exactly)
	for face := 0; face < 6; face++ {
		original := Coordinate{Face: face, X: res / 2, Y: res / 2}
		x, y, z := ct.topology.ToSphere(original)
		recovered := ct.topology.FromVector(x, y, z)

		// Face should match exactly
		if recovered.Face != original.Face {
			ct.t.Errorf("FromVector round-trip face mismatch: original=%v, recovered=%v", original, recovered)
		}
		// X/Y allow small tolerance for floating point
		tolerance := 1
		if abs(recovered.X-original.X) > tolerance || abs(recovered.Y-original.Y) > tolerance {
			ct.t.Errorf("FromVector round-trip coordinate mismatch: original=%v, recovered=%v", original, recovered)
		}
	}
}

// TestNeighborReversibility verifies GetNeighbor in opposite directions returns to start
func (ct *TopologyContractTest) TestNeighborReversibility() {
	ct.t.Helper()
	res := ct.topology.Resolution()
	// Test from center of face (no edge transitions)
	start := Coordinate{Face: 0, X: res / 2, Y: res / 2}

	opposites := map[Direction]Direction{
		North: South,
		South: North,
		East:  West,
		West:  East,
	}

	for dir, opposite := range opposites {
		moved := ct.topology.GetNeighbor(start, dir)
		returned := ct.topology.GetNeighbor(moved, opposite)
		if returned != start {
			ct.t.Errorf("GetNeighbor reversibility failed: start=%v, dir=%v, returned=%v",
				start, dir, returned)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
