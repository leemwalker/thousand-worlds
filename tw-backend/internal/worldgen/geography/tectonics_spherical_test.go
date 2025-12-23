package geography

import (
	"testing"

	"tw-backend/internal/spatial"
)

func TestPlateGeneration_Spherical(t *testing.T) {
	// Test with small grid for fast execution
	resolution := 10
	topology := spatial.NewCubeSphereTopology(resolution)
	seed := int64(42)
	plateCount := 5

	// Generate plates
	plates := GeneratePlates(plateCount, topology, seed)

	// Verify correct number of plates
	if len(plates) != plateCount {
		t.Errorf("Expected %d plates, got %d", plateCount, len(plates))
	}

	// Count total cells assigned to plates
	totalAssigned := 0
	for _, plate := range plates {
		totalAssigned += len(plate.Region)
	}

	// Total cells on cube sphere = 6 * resolution^2
	expectedTotal := 6 * resolution * resolution
	if totalAssigned != expectedTotal {
		t.Errorf("Expected %d cells assigned, got %d", expectedTotal, totalAssigned)
	}

	// Verify every cell is assigned to exactly one plate
	cellCounts := make(map[spatial.Coordinate]int)
	for _, plate := range plates {
		for coord := range plate.Region {
			cellCounts[coord]++
		}
	}

	for coord, count := range cellCounts {
		if count != 1 {
			t.Errorf("Cell %v assigned to %d plates (expected 1)", coord, count)
		}
	}
}

func TestPlateGeneration_CrossFaceBoundary(t *testing.T) {
	// Test that plates cross face boundaries
	resolution := 10
	topology := spatial.NewCubeSphereTopology(resolution)
	seed := int64(42)
	plateCount := 3 // Few plates = larger regions = higher chance of crossing

	plates := GeneratePlates(plateCount, topology, seed)

	// Check if any plate exists on multiple faces
	crossFacePlateFound := false
	for _, plate := range plates {
		facesPresent := make(map[int]bool)
		for coord := range plate.Region {
			facesPresent[coord.Face] = true
		}
		if len(facesPresent) > 1 {
			crossFacePlateFound = true
			break
		}
	}

	if !crossFacePlateFound {
		t.Error("No plates found crossing face boundaries - BFS should expand across faces")
	}
}

func TestCalculateBoundaryType(t *testing.T) {
	tests := []struct {
		name     string
		plateA   TectonicPlate
		plateB   TectonicPlate
		expected BoundaryType
	}{
		{
			name: "Convergent - plates moving toward each other",
			plateA: TectonicPlate{
				Position: spatial.Vector3D{X: 1, Y: 0, Z: 0},
				Velocity: spatial.Vector3D{X: -0.5, Y: 0, Z: 0}, // Moving toward B (toward -X)
			},
			plateB: TectonicPlate{
				Position: spatial.Vector3D{X: -1, Y: 0, Z: 0},
				Velocity: spatial.Vector3D{X: 0.5, Y: 0, Z: 0}, // Moving toward A (toward +X)
			},
			expected: BoundaryConvergent,
		},
		{
			name: "Divergent - plates moving apart",
			plateA: TectonicPlate{
				Position: spatial.Vector3D{X: 1, Y: 0, Z: 0},
				Velocity: spatial.Vector3D{X: 0.5, Y: 0, Z: 0}, // Moving away from B (toward +X)
			},
			plateB: TectonicPlate{
				Position: spatial.Vector3D{X: -1, Y: 0, Z: 0},
				Velocity: spatial.Vector3D{X: -0.5, Y: 0, Z: 0}, // Moving away from A (toward -X)
			},
			expected: BoundaryDivergent,
		},
		{
			name: "Transform - plates sliding past",
			plateA: TectonicPlate{
				Position: spatial.Vector3D{X: 1, Y: 0, Z: 0},
				Velocity: spatial.Vector3D{X: 0, Y: 0.1, Z: 0}, // Moving perpendicular
			},
			plateB: TectonicPlate{
				Position: spatial.Vector3D{X: -1, Y: 0, Z: 0},
				Velocity: spatial.Vector3D{X: 0, Y: 0.1, Z: 0}, // Same perpendicular movement
			},
			expected: BoundaryTransform,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateBoundaryType(tc.plateA, tc.plateB)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestSphericalVoronoi_AllCellsCovered(t *testing.T) {
	// Ensure BFS covers all cells
	resolution := 8
	topology := spatial.NewCubeSphereTopology(resolution)
	seed := int64(123)
	plateCount := 4

	plates := GeneratePlates(plateCount, topology, seed)

	// Build map of all assigned cells
	assigned := make(map[spatial.Coordinate]bool)
	for _, plate := range plates {
		for coord := range plate.Region {
			assigned[coord] = true
		}
	}

	// Verify all cells on all faces are assigned
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				if !assigned[coord] {
					t.Errorf("Cell %v not assigned to any plate", coord)
				}
			}
		}
	}
}

func TestSimulateTectonics_Spherical(t *testing.T) {
	resolution := 8
	topology := spatial.NewCubeSphereTopology(resolution)
	seed := int64(42)
	plateCount := 3

	// Generate plates
	plates := GeneratePlates(plateCount, topology, seed)

	// Create heightmap with base elevations
	hm := NewSphereHeightmap(topology)
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				hm.Set(coord, 0) // Start at 0
			}
		}
	}

	// Simulate tectonics
	SimulateTectonics(plates, hm, topology)

	// Check that some elevations changed (boundaries should have effects)
	changedCells := 0
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				if hm.Get(coord) != 0 {
					changedCells++
				}
			}
		}
	}

	if changedCells == 0 {
		t.Error("No elevation changes detected - tectonics should affect some cells")
	}
}
