package geography

import (
	"testing"
	"tw-backend/internal/spatial"

	"github.com/google/uuid"
)

func TestHierarchicalPlateGrid(t *testing.T) {
	// Setup mock plates
	res := 128
	blockSize := 32

	plates := make([]TectonicPlate, 2)
	plates[0] = TectonicPlate{
		ID:     uuid.New(),
		Region: make(map[spatial.Coordinate]struct{}),
	}
	plates[1] = TectonicPlate{
		ID:     uuid.New(),
		Region: make(map[spatial.Coordinate]struct{}),
	}

	// Plate 0 covers top-left 32x32 block (Face 0)
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			c := spatial.Coordinate{Face: 0, X: x, Y: y}
			plates[0].Region[c] = struct{}{}
		}
	}

	// Plate 1 covers next block (Face 0, 32-63)
	for y := 0; y < 32; y++ {
		for x := 32; x < 64; x++ {
			c := spatial.Coordinate{Face: 0, X: x, Y: y}
			plates[1].Region[c] = struct{}{}
		}
	}

	// Create mixed block at (Face 0, 0, 32)
	// Half Plate 0, Half Plate 1
	for y := 32; y < 64; y++ {
		for x := 0; x < 16; x++ {
			c := spatial.Coordinate{Face: 0, X: x, Y: y}
			plates[0].Region[c] = struct{}{}
		}
		for x := 16; x < 32; x++ {
			c := spatial.Coordinate{Face: 0, X: x, Y: y}
			plates[1].Region[c] = struct{}{}
		}
	}

	// Initialize Grid
	grid := NewHierarchicalPlateGrid(plates, res, blockSize)

	// Case 1: Pure Block (Plate 0)
	c1 := spatial.Coordinate{Face: 0, X: 10, Y: 10}
	pid1 := grid.GetPlateID(c1)
	if pid1 != 0 {
		t.Errorf("Expected Plate 0 for pure block, got %d", pid1)
	}

	// Case 2: Pure Block (Plate 1)
	c2 := spatial.Coordinate{Face: 0, X: 40, Y: 10}
	pid2 := grid.GetPlateID(c2)
	if pid2 != 1 {
		t.Errorf("Expected Plate 1 for pure block, got %d", pid2)
	}

	// Case 3: Mixed Block (Plate 0 side)
	c3 := spatial.Coordinate{Face: 0, X: 5, Y: 40}
	pid3 := grid.GetPlateID(c3)
	if pid3 != 0 {
		t.Errorf("Expected Plate 0 for mixed block, got %d", pid3)
	}

	// Case 4: Mixed Block (Plate 1 side)
	c4 := spatial.Coordinate{Face: 0, X: 20, Y: 40}
	pid4 := grid.GetPlateID(c4)
	if pid4 != 1 {
		t.Errorf("Expected Plate 1 for mixed block, got %d", pid4)
	}

	// Case 5: Empty/Unassigned
	// Should return -1
	c5 := spatial.Coordinate{Face: 5, X: 10, Y: 10} // Far away
	pid5 := grid.GetPlateID(c5)
	if pid5 != -1 {
		t.Errorf("Expected -1 for unassigned block, got %d", pid5)
	}
}
