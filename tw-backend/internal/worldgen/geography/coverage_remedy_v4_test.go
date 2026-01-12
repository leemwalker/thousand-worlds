package geography

import (
	"testing"
	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
)

func TestRouteFluxThroughLakes(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(4)
	hm := NewSphereHeightmap(topo)

	// Setup Lake
	lakeCells := []spatial.Coordinate{
		{Face: 0, X: 1, Y: 1},
		{Face: 0, X: 1, Y: 2},
	}
	outlet := spatial.Coordinate{Face: 0, X: 2, Y: 1}

	lake := &Lake{
		ID:            1,
		SurfaceHeight: 100.0,
		Cells:         lakeCells,
		Outlet:        outlet,
	}

	// Determine flow to ensure cells are sinks?
	// The implementation checks 'isSink = true' if neighbors >= elev.
	// But it sets isSink=true, then iterates neighbors.
	// We need to set elevations such that lakeCells are local minima/flat.
	hm.Set(lakeCells[0], 90.0)
	hm.Set(lakeCells[1], 90.0)
	// Surrounding neighbors need to be > 90.0
	// Lake outlet is usually a spillover, so it might be lower?
	// But RouteFluxThroughLakes checks if cell is a SINK.
	// A sink means NO neighbor is strictly lower.
	// So neighbors must be >= 90.0.

	// Initialize Flux
	data0 := hm.GetCellData(lakeCells[0])
	data0.Flux = 10.0
	hm.SetCellData(lakeCells[0], data0)

	data1 := hm.GetCellData(lakeCells[1])
	data1.Flux = 20.0
	hm.SetCellData(lakeCells[1], data1)

	// Neighbors (including outlet) need to be >= 90.0 so logic sees lakeCells as sinks.
	// Default is 0.0, so they would flow OUT.
	// We must raise surrounding terrain.
	// Since we are checking IsSink manually in the loop.
	// Let's brute force all neighbors to 100.0.

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}
	for _, c := range lakeCells {
		for _, dir := range directions {
			n := topo.GetNeighbor(c, dir)
			hm.Set(n, 100.0)
		}
	}
	// Re-set lake cells (GetNeighbor might have overwritten if neighbor was a lake cell)
	hm.Set(lakeCells[0], 90.0)
	hm.Set(lakeCells[1], 90.0)

	// Run
	RouteFluxThroughLakes(hm, []*Lake{lake})

	// Check Outlet Flux
	// Should encompass 10+20 = 30.
	outletData := hm.GetCellData(outlet)
	assert.Equal(t, 30.0, outletData.Flux, "Outlet should receive accumulated flux from lake sinks")
}

func TestWorldShapes(t *testing.T) {
	// 1. Bounded
	minW, minH := 10, 10
	bounded := GetShape(ShapeBounded, minW, minH)
	p1 := Point{X: 1, Y: 1}
	p2 := Point{X: 4, Y: 5}
	// Dist = sqrt(3^2 + 4^2) = 5
	assert.InDelta(t, 5.0, bounded.Distance(p1, p2), 0.001)

	assert.True(t, bounded.IsValid(p1))
	assert.False(t, bounded.IsValid(Point{X: -1, Y: 5}))
	assert.False(t, bounded.IsValid(Point{X: 11, Y: 5}))

	wrapped := bounded.WrapCoordinates(Point{X: -5, Y: -5})
	assert.Equal(t, 0.0, wrapped.X) // Clamped
	assert.Equal(t, 0.0, wrapped.Y)

	// 2. Spherical
	spherical := GetShape(ShapeSpherical, 100, 50)
	// Wrapping X
	pA := Point{X: 10, Y: 10}
	pB := Point{X: 90, Y: 10}
	// Naive dist = 80. Wrapped dist = 20.
	assert.InDelta(t, 20.0, spherical.Distance(pA, pB), 0.001)

	// Wrap coords
	wP := spherical.WrapCoordinates(Point{X: -10, Y: 10})
	assert.Equal(t, 90.0, wP.X)
	assert.Equal(t, 10.0, wP.Y)

	wP2 := spherical.WrapCoordinates(Point{X: 110, Y: 10})
	assert.Equal(t, 10.0, wP2.X)

	// IsValid (Y bound)
	assert.True(t, spherical.IsValid(Point{X: 500, Y: 25})) // X doesn't fail valid check usually (implied wrap) but standard check might be looser?
	// wait, IsValid implemention: p.Y >= 0 && p.Y < s.Height. X is ignored.
	assert.True(t, spherical.IsValid(Point{X: 500, Y: 25}))
	assert.False(t, spherical.IsValid(Point{X: 10, Y: 55}))

	// 3. Infinite
	infinite := GetShape(ShapeInfinite, 0, 0)
	assert.True(t, infinite.IsValid(Point{X: -999, Y: 999}))
	assert.Equal(t, Point{X: -10, Y: 10}, infinite.WrapCoordinates(Point{X: -10, Y: 10}))
}
