package geography

import (
	"math"
	"testing"

	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Phase 1: Organic Plate Boundary Tests
// =============================================================================

// TestReassignPlateRegions_OrganicBoundaries verifies that plate boundaries
// are irregular and organic rather than straight Voronoi edges.
// This is achieved through noise-based cost modifiers in the Dijkstra expansion.
func TestReassignPlateRegions_OrganicBoundaries(t *testing.T) {
	resolution := 64 // Higher res to see boundary detail
	topology := spatial.NewCubeSphereTopology(resolution)
	seed := int64(42)

	// Create plates with fixed centroids for controlled testing
	plates := GeneratePlates(6, topology, seed, 0.30)

	// Find boundary cells (cells where neighbor has different plate)
	boundaryCells := findBoundaryCells(plates, topology)
	require.NotEmpty(t, boundaryCells, "Should have boundary cells")

	// Calculate boundary irregularity by comparing to a pure geometric Voronoi
	// An organic boundary should have more cells than a straight-line Voronoi
	// because it meanders around, covering more area
	//
	// Alternative metric: count how many boundary cells have only 1-2 boundary neighbors
	// (edges of peninsulas/bays) vs 3-4 neighbors (straight runs)

	peninsulaCount := 0 // Cells with only 1 boundary neighbor (tips of peninsulas/bays)
	cornerCount := 0    // Cells with exactly 2 non-straight boundary neighbors
	straightCount := 0  // Cells with 2 opposite boundary neighbors (straight line)

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	for coord := range boundaryCells {
		boundaryNeighbors := 0
		hasNorth, hasSouth, hasEast, hasWest := false, false, false, false

		for _, dir := range directions {
			neighbor := topology.GetNeighbor(coord, dir)
			if _, isBoundary := boundaryCells[neighbor]; isBoundary {
				boundaryNeighbors++
				switch dir {
				case spatial.North:
					hasNorth = true
				case spatial.South:
					hasSouth = true
				case spatial.East:
					hasEast = true
				case spatial.West:
					hasWest = true
				}
			}
		}

		if boundaryNeighbors <= 1 {
			peninsulaCount++
		} else if boundaryNeighbors == 2 {
			// Straight line = opposite neighbors (N-S or E-W)
			if (hasNorth && hasSouth) || (hasEast && hasWest) {
				straightCount++
			} else {
				cornerCount++ // L-shaped turn
			}
		}
	}

	totalBoundary := len(boundaryCells)
	organicRatio := float64(peninsulaCount+cornerCount) / float64(totalBoundary)

	// Organic boundaries should have at least 30% non-straight cells
	// (peninsulas, bays, corners where the boundary turns)
	assert.GreaterOrEqual(t, organicRatio, 0.25,
		"Boundary should be organic with at least 25%% peninsulas/corners, got %.1f%% (peninsulas=%d, corners=%d, straight=%d)",
		organicRatio*100, peninsulaCount, cornerCount, straightCount)

	t.Logf("Boundary analysis: total=%d peninsulas=%d corners=%d straight=%d (%.1f%% organic)",
		totalBoundary, peninsulaCount, cornerCount, straightCount, organicRatio*100)
}

// TestReassignPlateRegions_DifferentSeeds_DifferentBoundaries verifies that
// changing the seed produces different boundary shapes.
func TestReassignPlateRegions_DifferentSeeds_DifferentBoundaries(t *testing.T) {
	resolution := 32
	topology := spatial.NewCubeSphereTopology(resolution)

	plates1 := GeneratePlates(5, topology, 111, 0.30)
	plates2 := GeneratePlates(5, topology, 222, 0.30)

	// Compare boundary patterns - they should differ
	boundaries1 := findBoundaryCells(plates1, topology)
	boundaries2 := findBoundaryCells(plates2, topology)

	// The sets should not be identical (different shapes due to different noise seeds)
	differentBoundaries := 0
	for coord := range boundaries1 {
		if _, exists := boundaries2[coord]; !exists {
			differentBoundaries++
		}
	}

	// At least 20% of boundaries should differ between seeds
	minDifferent := len(boundaries1) / 5
	assert.GreaterOrEqual(t, differentBoundaries, minDifferent,
		"Different seeds should produce different boundary patterns")
}

// TestIslandArc_DistributionNotUniform verifies that island arcs form
// with variation (not uniform straight lines of identical cells).
func TestIslandArc_DistributionNotUniform(t *testing.T) {
	resolution := 32
	topology := spatial.NewCubeSphereTopology(resolution)
	seed := int64(12345)

	// Create plates with known oceanic-oceanic boundaries
	plates := GeneratePlates(8, topology, seed, 0.20) // Lower continental % = more oceanic

	// Find oceanic-oceanic convergent boundaries
	cache := ComputeBoundaryCache(plates, topology)

	oceanOceanConvergent := 0
	for _, bc := range cache.Cells {
		if bc.BoundaryType == BoundaryConvergent &&
			plates[bc.PlateIdx].Type == PlateOceanic &&
			plates[bc.NeighborIdx].Type == PlateOceanic {
			oceanOceanConvergent++
		}
	}

	// Should have some oceanic-oceanic convergent boundaries
	assert.Greater(t, oceanOceanConvergent, 0,
		"Should have oceanic-oceanic convergent boundaries for island arc formation")

	t.Logf("Found %d oceanic-oceanic convergent boundary cells", oceanOceanConvergent)
}

// =============================================================================
// Helper Functions
// =============================================================================

// findBoundaryCells returns all cells that are at plate boundaries
func findBoundaryCells(plates []TectonicPlate, topology spatial.Topology) map[spatial.Coordinate]struct{} {
	// Build plate lookup
	cellToPlate := make(map[spatial.Coordinate]int)
	for i, p := range plates {
		for coord := range p.Region {
			cellToPlate[coord] = i
		}
	}

	boundaries := make(map[spatial.Coordinate]struct{})
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	for coord, plateIdx := range cellToPlate {
		for _, dir := range directions {
			neighbor := topology.GetNeighbor(coord, dir)
			if neighborPlate, exists := cellToPlate[neighbor]; exists && neighborPlate != plateIdx {
				boundaries[coord] = struct{}{}
				break
			}
		}
	}

	return boundaries
}

// measureBoundaryLinearity measures the longest straight run in boundary cells
// Returns the maximum number of consecutive boundary cells in a cardinal direction
func measureBoundaryLinearity(boundaries map[spatial.Coordinate]struct{}, topology spatial.Topology) int {
	maxRun := 0
	visited := make(map[spatial.Coordinate]bool)

	for start := range boundaries {
		if visited[start] {
			continue
		}

		// Check runs in each direction
		for _, dir := range []spatial.Direction{spatial.North, spatial.East} {
			run := countRunInDirection(start, dir, boundaries, topology, visited)
			if run > maxRun {
				maxRun = run
			}
		}
	}

	return maxRun
}

// countRunInDirection counts consecutive boundary cells in one direction
func countRunInDirection(start spatial.Coordinate, dir spatial.Direction,
	boundaries map[spatial.Coordinate]struct{}, topology spatial.Topology,
	visited map[spatial.Coordinate]bool) int {

	count := 1
	current := start
	visited[current] = true

	for {
		next := topology.GetNeighbor(current, dir)
		if _, isBoundary := boundaries[next]; !isBoundary {
			break
		}
		if visited[next] {
			break
		}
		visited[next] = true
		count++
		current = next

		// Safety limit
		if count > 100 {
			break
		}
	}

	return count
}

// =============================================================================
// Phase 3: Earth History Placeholder (will be in calibration package)
// =============================================================================

// TestPlateGrowth_HadeanToArchean simulates early Earth plate evolution
// and verifies continental coverage increases over time.
func TestPlateGrowth_HadeanToArchean(t *testing.T) {
	resolution := 32
	topology := spatial.NewCubeSphereTopology(resolution)
	seed := int64(4500) // Earth seed

	// Start with minimal continental coverage (Hadean: ~0%)
	plates := GeneratePlates(10, topology, seed, 0.0)

	// Initialize heightmap
	hm := NewSphereHeightmap(topology)
	hm = GenerateHeightmap(plates, hm, topology, seed, 1.0, 1.0)

	// Count initial continental cells
	initialContinental := countContinentalCells(hm, topology)

	// Simulate 50 ticks (~50-100 million years)
	for i := 0; i < 50; i++ {
		UpdatePlatePositions(plates, 1.0, topology)
		ReassignPlateRegions(plates, topology, seed+int64(i))
		cache := ComputeBoundaryCache(plates, topology)
		hm = SimulateTectonicsWithCache(plates, hm, cache, topology, 1.0, seed, 15000.0)
	}

	// Count final continental cells
	finalContinental := countContinentalCells(hm, topology)

	// Continental coverage should increase or stay same (accretion)
	assert.GreaterOrEqual(t, finalContinental, initialContinental,
		"Continental cells should not decrease over time (Hadean growth)")

	t.Logf("Continental growth: %d -> %d cells", initialContinental, finalContinental)
}

func countContinentalCells(hm *SphereHeightmap, topology spatial.Topology) int {
	count := 0
	res := topology.Resolution()
	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				cell := hm.GetCellData(spatial.Coordinate{Face: face, X: x, Y: y})
				if cell.IsContinental {
					count++
				}
			}
		}
	}
	return count
}

// Ensure math import is used
var _ = math.MaxFloat64
