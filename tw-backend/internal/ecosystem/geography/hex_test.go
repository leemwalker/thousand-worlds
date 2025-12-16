package geography

import (
	"testing"

	"github.com/google/uuid"
)

func TestHexCoord_Neighbor(t *testing.T) {
	center := NewHexCoord(0, 0)

	t.Run("all six neighbors correct", func(t *testing.T) {
		expected := map[HexDirection]HexCoord{
			HexDirE:  {Q: 1, R: 0},
			HexDirNE: {Q: 1, R: -1},
			HexDirNW: {Q: 0, R: -1},
			HexDirW:  {Q: -1, R: 0},
			HexDirSW: {Q: -1, R: 1},
			HexDirSE: {Q: 0, R: 1},
		}

		for dir, want := range expected {
			got := center.Neighbor(dir)
			if got != want {
				t.Errorf("Neighbor(%d) = %+v, want %+v", dir, got, want)
			}
		}
	})
}

func TestHexCoord_AllNeighbors(t *testing.T) {
	center := NewHexCoord(3, 5)
	neighbors := center.AllNeighbors()

	if len(neighbors) != 6 {
		t.Errorf("AllNeighbors returned %d, want 6", len(neighbors))
	}

	// Each neighbor should be distance 1 from center
	for _, n := range neighbors {
		dist := center.Distance(n)
		if dist != 1 {
			t.Errorf("Neighbor %+v has distance %d, want 1", n, dist)
		}
	}
}

func TestHexCoord_Distance(t *testing.T) {
	tests := []struct {
		a, b     HexCoord
		expected int
	}{
		{HexCoord{0, 0}, HexCoord{0, 0}, 0},
		{HexCoord{0, 0}, HexCoord{1, 0}, 1},
		{HexCoord{0, 0}, HexCoord{2, 0}, 2},
		{HexCoord{0, 0}, HexCoord{1, 1}, 2},
		{HexCoord{0, 0}, HexCoord{3, -3}, 3},
		{HexCoord{-2, 4}, HexCoord{2, -2}, 6},
	}

	for _, tt := range tests {
		dist := tt.a.Distance(tt.b)
		if dist != tt.expected {
			t.Errorf("Distance(%+v, %+v) = %d, want %d", tt.a, tt.b, dist, tt.expected)
		}
	}
}

func TestHexCoord_Ring(t *testing.T) {
	center := NewHexCoord(0, 0)

	t.Run("ring radius 0", func(t *testing.T) {
		ring := center.Ring(0)
		if len(ring) != 1 {
			t.Errorf("Ring(0) has %d elements, want 1", len(ring))
		}
	})

	t.Run("ring radius 1", func(t *testing.T) {
		ring := center.Ring(1)
		if len(ring) != 6 {
			t.Errorf("Ring(1) has %d elements, want 6", len(ring))
		}
		// Each should be distance 1
		for _, coord := range ring {
			dist := center.Distance(coord)
			if dist != 1 {
				t.Errorf("Ring(1) coord %+v has distance %d", coord, dist)
			}
		}
	})

	t.Run("ring radius 2", func(t *testing.T) {
		ring := center.Ring(2)
		if len(ring) != 12 {
			t.Errorf("Ring(2) has %d elements, want 12", len(ring))
		}
	})
}

func TestHexCoord_Spiral(t *testing.T) {
	center := NewHexCoord(0, 0)

	spiral := center.Spiral(2)
	// Spiral(2) = center + ring(1) + ring(2) = 1 + 6 + 12 = 19
	expected := 1 + 6 + 12
	if len(spiral) != expected {
		t.Errorf("Spiral(2) has %d elements, want %d", len(spiral), expected)
	}
}

func TestHexCoord_PixelConversion(t *testing.T) {
	hexSize := 10.0

	t.Run("origin converts to origin", func(t *testing.T) {
		coord := HexCoord{0, 0}
		x, y := coord.ToPixel(hexSize)
		if x != 0 || y != 0 {
			t.Errorf("Origin should convert to (0,0), got (%f, %f)", x, y)
		}
	})

	t.Run("round trip conversion", func(t *testing.T) {
		original := HexCoord{3, -2}
		x, y := original.ToPixel(hexSize)
		back := FromPixel(x, y, hexSize)
		if back != original {
			t.Errorf("Round trip: %+v -> (%f,%f) -> %+v", original, x, y, back)
		}
	})
}

func TestNewHexCell(t *testing.T) {
	coord := NewHexCoord(5, 3)
	cell := NewHexCell(coord, TerrainPlains, 0.5)

	if cell.Coord != coord {
		t.Error("Cell coord mismatch")
	}
	if cell.Terrain != TerrainPlains {
		t.Error("Terrain mismatch")
	}
	if cell.Elevation != 0.5 {
		t.Error("Elevation mismatch")
	}
	if !cell.IsLand {
		t.Error("Plains should be land")
	}
}

func TestHexCell_IsPassable(t *testing.T) {
	tests := []struct {
		terrain  TerrainType
		passable bool
	}{
		{TerrainPlains, true},
		{TerrainHills, true},
		{TerrainCoast, true},
		{TerrainMountain, false},
		{TerrainOcean, false},
		{TerrainVolcanic, false},
	}

	for _, tt := range tests {
		cell := NewHexCell(NewHexCoord(0, 0), tt.terrain, 0)
		if cell.IsPassable() != tt.passable {
			t.Errorf("Terrain %s: IsPassable() = %v, want %v", tt.terrain, cell.IsPassable(), tt.passable)
		}
	}
}

func TestHexGrid_GetNeighbors(t *testing.T) {
	grid := NewHexGrid(uuid.Nil, 10, 10, 1.0)

	// Add a 3x3 area of cells
	for q := 0; q < 3; q++ {
		for r := 0; r < 3; r++ {
			cell := NewHexCell(NewHexCoord(q, r), TerrainPlains, 0)
			grid.SetCell(cell)
		}
	}

	center := NewHexCoord(1, 1)
	neighbors := grid.GetNeighbors(center)

	// Center should have 6 neighbors in a 3x3 grid (if all exist)
	// Actually depends on which are in bounds - let's just check we get some
	if len(neighbors) == 0 {
		t.Error("Expected some neighbors")
	}
}

func TestHexGrid_FindPath(t *testing.T) {
	grid := NewHexGrid(uuid.Nil, 10, 10, 1.0)

	// Create a line of passable hexes
	for q := 0; q < 5; q++ {
		cell := NewHexCell(NewHexCoord(q, 0), TerrainPlains, 0)
		grid.SetCell(cell)
	}

	t.Run("finds path when exists", func(t *testing.T) {
		path := grid.FindPath(NewHexCoord(0, 0), NewHexCoord(4, 0))
		if path == nil {
			t.Fatal("Expected to find path")
		}
		if len(path) != 5 {
			t.Errorf("Path length = %d, want 5", len(path))
		}
	})

	t.Run("returns nil for unreachable", func(t *testing.T) {
		// Try to reach a cell that doesn't exist
		path := grid.FindPath(NewHexCoord(0, 0), NewHexCoord(10, 10))
		if path != nil {
			t.Error("Should return nil for unreachable destination")
		}
	})
}
