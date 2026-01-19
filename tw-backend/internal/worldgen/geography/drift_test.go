package geography

import (
	"math"
	"testing"

	"tw-backend/internal/spatial"

	"github.com/google/uuid"
)

// TestHadeanDrift verifies that plates actually move and change relative positions
// when the drift engine is engaged.
func TestHadeanDrift(t *testing.T) {
	// Initialize strict topology
	topology := spatial.NewCubeSphereTopology(32) // Low res for speed

	// 1. Initialize Plates
	// Create 2 plates at known positions
	plates := make([]TectonicPlate, 2)

	// Plate 0: Centered at Face 3 (Equator, +X)
	p0Center := spatial.Coordinate{Face: 3, X: 16, Y: 16}
	p0Pos := VectorFromCoord(t, topology, p0Center)

	// Plate 1: Centered at Face 1 (0,0) -> 90 degrees away
	p1Center := spatial.Coordinate{Face: 1, X: 16, Y: 16}
	p1Pos := VectorFromCoord(t, topology, p1Center)

	plates[0] = TectonicPlate{
		ID:           uuid.New(),
		Type:         PlateOceanic,
		Centroid:     p0Center,
		Position:     p0Pos,
		RotationAxis: spatial.Vector3D{X: 0, Y: 0, Z: 1}, // Rotate around Z-axis (North Pole)
		AngularSpeed: 0.5,                                // Very Fast rotation for testing
		Region:       make(map[spatial.Coordinate]struct{}),
	}

	plates[1] = TectonicPlate{
		ID:           uuid.New(),
		Type:         PlateOceanic,
		Centroid:     p1Center,
		Position:     p1Pos,
		RotationAxis: spatial.Vector3D{X: 0, Y: 1, Z: 0}, // Rotate around Y-axis
		AngularSpeed: 0.0,                                // Static
		Region:       make(map[spatial.Coordinate]struct{}),
	}

	// Calculate initial distance
	initialDist := plates[0].Position.Distance(plates[1].Position)
	t.Logf("Initial Distance: %.4f", initialDist)

	// 2. Run Drift Loop
	steps := 10
	dt := 2.0

	for i := 0; i < steps; i++ {
		UpdatePlatePositions(plates, dt, topology)
		// RecalculateRegions is expensive, usually run less often, but needed for visual check
		// For this test we only check Centroid movement
	}

	// 3. Verify Movement
	finalDist := plates[0].Position.Distance(plates[1].Position)
	t.Logf("Final Distance: %.4f", finalDist)

	// Check that Plate 0 has moved
	movedDist := plates[0].Position.Distance(p0Pos)
	if movedDist < 0.1 {
		t.Errorf("Plate 0 should have moved, but dist is %.4f", movedDist)
	}

	// Check that relative distance changed (unless they rotate perfectly parallel, which they don't here)
	if math.Abs(finalDist-initialDist) < 0.01 {
		t.Errorf("Relative distance should change (was %.4f, now %.4f)", initialDist, finalDist)
	}

	// Double check Plate 1 remained static
	p1Moved := plates[1].Position.Distance(p1Pos)
	if p1Moved > 0.0001 {
		t.Errorf("Plate 1 should stay static, moved %.4f", p1Moved)
	}
}

// Helper to get vector from coord
func VectorFromCoord(t *testing.T, topo spatial.Topology, c spatial.Coordinate) spatial.Vector3D {
	x, y, z := topo.ToSphere(c)
	return spatial.Vector3D{X: x, Y: y, Z: z}
}

// TestAccretion verifies that Ocean-Ocean convergence leads to mass accretion
// and (eventually) continental crust formation.
func TestAccretion(t *testing.T) {
	// 1. Setup Topology and Heightmap
	res := 16
	topology := spatial.NewCubeSphereTopology(res)
	shm := NewSphereHeightmap(topology)

	// Initialize Heightmap to Deep Ocean
	for i := 0; i < 6; i++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				shm.Set(spatial.Coordinate{Face: i, X: x, Y: y}, -4000)
			}
		}
	}

	// 2. Setup Plates (converging)
	plates := make([]TectonicPlate, 2)

	// Plate A (Left, Face 3) moving Right (+X)
	// Plate B (Right, Face 3) moving Left (-X)
	// They will collide in the middle of Face 3

	centerA := spatial.Coordinate{Face: 3, X: 4, Y: 8}
	centerB := spatial.Coordinate{Face: 3, X: 12, Y: 8}

	posA := VectorFromCoord(t, topology, centerA)
	posB := VectorFromCoord(t, topology, centerB)

	// Calculate direction from A to B
	dir := posB.Sub(posA).Normalize()

	plates[0] = TectonicPlate{
		ID:        uuid.New(),
		Type:      PlateOceanic,
		Centroid:  centerA,
		Position:  posA,
		Velocity:  dir, // Moving towards B
		Region:    make(map[spatial.Coordinate]struct{}),
		Thickness: 6.0,
		Age:       100.0, // Older -> Denser -> Subducts
	}

	plates[1] = TectonicPlate{
		ID:        uuid.New(),
		Type:      PlateOceanic,
		Centroid:  centerB,
		Position:  posB,
		Velocity:  dir.Scale(-1.0), // Moving towards A
		Region:    make(map[spatial.Coordinate]struct{}),
		Thickness: 6.0,
		Age:       10.0, // Younger -> Lighter -> Overrides
	}

	// Manually assign regions (half and half)
	for x := 0; x < 8; x++ {
		for y := 0; y < res; y++ {
			plates[0].Region[spatial.Coordinate{Face: 3, X: x, Y: y}] = struct{}{}
		}
	}
	for x := 8; x < res; x++ {
		for y := 0; y < res; y++ {
			plates[1].Region[spatial.Coordinate{Face: 3, X: x, Y: y}] = struct{}{}
		}
	}

	// 3. Compute Cache
	cache := ComputeBoundaryCache(plates, topology)
	t.Logf("Computed %d boundary cells", len(cache.Cells))
	if len(cache.Cells) == 0 {
		t.Fatal("Expected boundary cells, got 0")
	}

	// 4. Run Simulation
	// Use high scale factor to guarantee accretion event
	shm = SimulateTectonicsWithCache(plates, shm, cache, nil, topology, 1.0, 12345, 10000.0)

	// 5. Verify Accretion matches Subduction Roles
	// Plate 0 (Old) should subduct under Plate 1 (Young)
	// So Plate 1 (Overriding) should gain AccretedMass

	if plates[1].AccretedMass <= 0 {
		t.Errorf("Overriding Plate 1 should have gained mass, got %.4f", plates[1].AccretedMass)
	}
	t.Logf("Plate 1 Accreted Mass: %.4f", plates[1].AccretedMass)

	if plates[0].AccretedMass > 0 {
		t.Errorf("Subducting Plate 0 should NOT gain mass, got %.4f", plates[0].AccretedMass)
	}

	// 6. Verify Continental Crust Formation
	// Check boundary cells of Plate 1
	continentalCount := 0
	for coord := range plates[1].Region {
		cellData := shm.GetCellData(coord)
		if cellData.IsContinental {
			continentalCount++
		}
	}

	t.Logf("Formed %d Continental Cells on Plate 1", continentalCount)
	if continentalCount == 0 {
		t.Error("Expected at least one cell to transform to Continental crust")
	}
}
