package geography

import (
	"log"
	"math/rand"
	"time"

	"tw-backend/internal/debug"
	"tw-backend/internal/spatial"

	"github.com/google/uuid"
)

// Elevation physical limits (in meters)
const (
	// MaxElevation is the upper bound for terrain (above Olympus Mons scale)
	MaxElevation = 15000.0
	// MinElevation is the lower bound for terrain (below Mariana Trench scale)
	MinElevation = -11000.0
	// TectonicConvergenceRate controls how quickly boundaries approach target elevation
	// Value of 0.1 means 10% of remaining difference per tectonic step
	TectonicConvergenceRate = 0.1
)

// GeneratePlates creates tectonic plates using spherical topology.
// Uses Multi-Source BFS to assign regions efficiently in O(N) time.
func GeneratePlates(count int, topology spatial.Topology, seed int64) []TectonicPlate {
	r := rand.New(rand.NewSource(seed))
	resolution := topology.Resolution()
	plates := make([]TectonicPlate, count)

	// 1. Initialize plates with random centroids distributed across all faces
	for i := 0; i < count; i++ {
		face := r.Intn(6)
		x := r.Intn(resolution)
		y := r.Intn(resolution)
		centroid := spatial.Coordinate{Face: face, X: x, Y: y}

		// Get 3D position on sphere from coordinate
		sx, sy, sz := topology.ToSphere(centroid)
		position := spatial.Vector3D{X: sx, Y: sy, Z: sz}

		// Generate random tangent velocity (perpendicular to position)
		velocity := randomTangentVector(position, r)

		// Assign type (30% continental, 70% oceanic)
		plateType := PlateOceanic
		thickness := 5 + r.Float64()*5 // 5-10km
		if i < int(float64(count)*0.3) {
			plateType = PlateContinental
			thickness = 30 + r.Float64()*20 // 30-50km
		}

		plates[i] = TectonicPlate{
			ID:        uuid.New(),
			Type:      plateType,
			Centroid:  centroid,
			Position:  position,
			Velocity:  velocity,
			Region:    make(map[spatial.Coordinate]struct{}),
			Thickness: thickness,
			Age:       r.Float64() * 100, // 0-100 million years
		}
	}

	// 2. Multi-Source BFS to assign all cells to nearest plate
	ReassignPlateRegions(plates, topology)

	return plates
}

// randomTangentVector generates a random unit vector tangent to the sphere at position.
func randomTangentVector(position spatial.Vector3D, r *rand.Rand) spatial.Vector3D {
	// Generate random vector
	arbitrary := spatial.Vector3D{
		X: r.NormFloat64(),
		Y: r.NormFloat64(),
		Z: r.NormFloat64(),
	}.Normalize()

	// Project out the radial component to get tangent
	// tangent = arbitrary - (arbitrary · position) * position
	dot := arbitrary.Dot(position)
	tangent := spatial.Vector3D{
		X: arbitrary.X - dot*position.X,
		Y: arbitrary.Y - dot*position.Y,
		Z: arbitrary.Z - dot*position.Z,
	}

	return tangent.Normalize()
}

// bfsItem represents a work item in the BFS queue
type bfsItem struct {
	coord    spatial.Coordinate
	plateIdx int
}

// ReassignPlateRegions uses Multi-Source BFS to assign every cell to the nearest plate.
// This naturally handles wrap-around and creates perfect Voronoi regions.
// Can be called after plate movement to update regions.
func ReassignPlateRegions(plates []TectonicPlate, topology spatial.Topology) {
	resolution := topology.Resolution()
	totalCells := 6 * resolution * resolution

	// Track which cells are assigned
	assigned := make(map[spatial.Coordinate]int, totalCells)

	// Initialize queue with all plate centroids
	queue := make([]bfsItem, 0, len(plates))
	for i, p := range plates {
		queue = append(queue, bfsItem{coord: p.Centroid, plateIdx: i})
		assigned[p.Centroid] = i
		plates[i].Region[p.Centroid] = struct{}{}
	}

	// Cardinal directions for neighbor traversal
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// BFS expansion
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check all 4 neighbors
		for _, dir := range directions {
			neighbor := topology.GetNeighbor(current.coord, dir)

			// If not already assigned, claim it for this plate
			if _, exists := assigned[neighbor]; !exists {
				assigned[neighbor] = current.plateIdx
				plates[current.plateIdx].Region[neighbor] = struct{}{}
				queue = append(queue, bfsItem{coord: neighbor, plateIdx: current.plateIdx})
			}
		}
	}
}

// ComputeBoundaryCache pre-computes all cells that are at plate boundaries.
// This is expensive but only needs to run when plates are reassigned.
// After this, SimulateTectonicsWithCache can process only boundary cells.
func ComputeBoundaryCache(plates []TectonicPlate, topology spatial.Topology) *BoundaryCache {
	resolution := topology.Resolution()
	totalCells := 6 * resolution * resolution

	cache := &BoundaryCache{
		Cells:      make([]BoundaryCell, 0, totalCells/10), // Expect ~10% boundaries
		PlateGrid:  make([]int, totalCells),
		Resolution: resolution,
		Valid:      true,
	}

	// Initialize with -1 (no plate)
	for i := range cache.PlateGrid {
		cache.PlateGrid[i] = -1
	}

	// Populate grid
	for i, p := range plates {
		for coord := range p.Region {
			idx := (coord.Face * resolution * resolution) + (coord.Y * resolution) + coord.X
			if idx >= 0 && idx < totalCells {
				cache.PlateGrid[idx] = i
			}
		}
	}

	// Find all boundary cells
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}
	resSq := resolution * resolution

	for idx := 0; idx < totalCells; idx++ {
		currentPlateIdx := cache.PlateGrid[idx]
		if currentPlateIdx == -1 {
			continue
		}

		// Reconstruct coordinate from index
		face := idx / resSq
		rem := idx % resSq
		y := rem / resolution
		x := rem % resolution
		coord := spatial.Coordinate{Face: face, X: x, Y: y}

		currentPlate := plates[currentPlateIdx]

		// Check neighbors for boundary
		for _, dir := range directions {
			neighbor := topology.GetNeighbor(coord, dir)
			nIdx := (neighbor.Face * resSq) + (neighbor.Y * resolution) + neighbor.X

			var neighborPlateIdx int
			if nIdx >= 0 && nIdx < totalCells {
				neighborPlateIdx = cache.PlateGrid[nIdx]
			} else {
				neighborPlateIdx = -1
			}

			if neighborPlateIdx == -1 || neighborPlateIdx == currentPlateIdx {
				continue
			}

			// Found a boundary - add to cache
			neighborPlate := plates[neighborPlateIdx]
			boundaryType := CalculateBoundaryType(currentPlate, neighborPlate)

			cache.Cells = append(cache.Cells, BoundaryCell{
				Coord:        coord,
				PlateIdx:     currentPlateIdx,
				NeighborIdx:  neighborPlateIdx,
				BoundaryType: boundaryType,
			})
		}
	}

	if debug.Is(debug.Perf | debug.Geology) {
		log.Printf("[BOUNDARY CACHE] Built cache with %d boundary cells out of %d total (%.1f%%)",
			len(cache.Cells), totalCells, float64(len(cache.Cells))/float64(totalCells)*100)
	}

	return cache
}

// SimulateTectonicsWithCache uses a pre-computed boundary cache for fast processing.
// Only iterates over boundary cells instead of all cells - typically 90% faster.
func SimulateTectonicsWithCache(plates []TectonicPlate, heightmap *SphereHeightmap, cache *BoundaryCache, topology spatial.Topology, scaleFactor float64) *SphereHeightmap {
	if debug.Is(debug.Perf) {
		defer debug.Time(debug.Perf, "SimulateTectonicsWithCache")()
	}

	// Process only cached boundary cells
	for _, bc := range cache.Cells {
		currentPlate := plates[bc.PlateIdx]
		neighborPlate := plates[bc.NeighborIdx]

		// Apply equilibrium-based elevation change
		currentElev := heightmap.Get(bc.Coord)
		elevationDelta := calculateEquilibriumElevationChange(currentPlate, neighborPlate, bc.BoundaryType, currentElev)

		// Apply scale factor for variable time steps
		elevationDelta *= scaleFactor

		applyBoundaryEffectSpherical(heightmap, bc.Coord, elevationDelta, topology)
	}

	return heightmap
}

// SimulateTectonics calculates elevation based on plate interactions on a sphere.
// Uses equilibrium-based approach where elevation approaches target asymptotically.
// Returns a SphereHeightmap with elevation modifiers.
// scaleFactor allows adjusting the intensity based on time step (1.0 = standard 100k year interval)
func SimulateTectonics(plates []TectonicPlate, heightmap *SphereHeightmap, topology spatial.Topology, scaleFactor float64) *SphereHeightmap {
	// Debug timing for overall function
	if debug.Is(debug.Perf) {
		defer debug.Time(debug.Perf, "SimulateTectonics(Total)")()
	}

	resolution := topology.Resolution()
	totalCells := 6 * resolution * resolution

	// === Phase 1: Grid Population ===
	gridStart := time.Now()

	// OPTIMIZATION: Use flat slice for O(1) plate lookup instead of map
	// Initialize with -1 (no plate)
	plateGrid := make([]int, totalCells)
	for i := range plateGrid {
		plateGrid[i] = -1
	}

	// Populate grid
	// O(N) where N is total cells (sum of all plate regions)
	for i, p := range plates {
		for coord := range p.Region {
			idx := (coord.Face * resolution * resolution) + (coord.Y * resolution) + coord.X
			if idx >= 0 && idx < totalCells {
				plateGrid[idx] = i
			}
		}
	}

	gridTime := time.Since(gridStart)

	// === Phase 2: Boundary Processing ===
	boundaryStart := time.Now()

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}
	resSq := resolution * resolution

	// Process all cells
	// O(N) linear scan with O(1) lookups is much faster than map iteration
	for idx := 0; idx < totalCells; idx++ {
		currentPlateIdx := plateGrid[idx]
		if currentPlateIdx == -1 {
			continue
		}

		// Reconstruct coordinate from index
		// idx = (face * res * res) + (y * res) + x
		face := idx / resSq
		rem := idx % resSq
		y := rem / resolution
		x := rem % resolution
		coord := spatial.Coordinate{Face: face, X: x, Y: y}

		currentPlate := plates[currentPlateIdx]

		// Check neighbors for boundary
		for _, dir := range directions {
			neighbor := topology.GetNeighbor(coord, dir)
			// Calculate neighbor index
			nIdx := (neighbor.Face * resSq) + (neighbor.Y * resolution) + neighbor.X

			var neighborPlateIdx int
			if nIdx >= 0 && nIdx < totalCells {
				neighborPlateIdx = plateGrid[nIdx]
			} else {
				neighborPlateIdx = -1 // Should not happen with valid topology
			}

			if neighborPlateIdx == -1 || neighborPlateIdx == currentPlateIdx {
				continue
			}

			// Found a boundary between two plates
			neighborPlate := plates[neighborPlateIdx]
			boundaryType := CalculateBoundaryType(currentPlate, neighborPlate)

			// Apply equilibrium-based elevation change
			currentElev := heightmap.Get(coord)
			elevationDelta := calculateEquilibriumElevationChange(currentPlate, neighborPlate, boundaryType, currentElev)

			// Apply scale factor for variable time steps
			elevationDelta *= scaleFactor

			applyBoundaryEffectSpherical(heightmap, coord, elevationDelta, topology)
		}
	}

	boundaryTime := time.Since(boundaryStart)

	// Log phase breakdown (only when --debug-perf is set)
	if debug.Is(debug.Perf | debug.Geology) {
		log.Printf("[TECTONICS PERF] Grid: %v (%.0f%%) | Boundary: %v (%.0f%%)",
			gridTime, float64(gridTime)/float64(gridTime+boundaryTime)*100,
			boundaryTime, float64(boundaryTime)/float64(gridTime+boundaryTime)*100)
	}

	return heightmap
}

// CalculateBoundaryType determines the type of interaction between two plates.
// Uses 3D vector math on the sphere surface.
func CalculateBoundaryType(plateA, plateB TectonicPlate) BoundaryType {
	// Normal vector from A to B (direction of boundary)
	normal := plateB.Position.Sub(plateA.Position).Normalize()

	// Relative velocity: how plates move relative to each other
	relativeVelocity := plateA.Velocity.Sub(plateB.Velocity)

	// Convergence score: positive = convergent, negative = divergent
	score := relativeVelocity.Dot(normal)

	if score > 0.2 {
		return BoundaryConvergent
	} else if score < -0.2 {
		return BoundaryDivergent
	}
	return BoundaryTransform
}

// GetTargetElevation returns the target elevation for a given boundary type.
// This is the equilibrium elevation that boundaries approach asymptotically.
func GetTargetElevation(p1, p2 TectonicPlate, boundaryType BoundaryType) float64 {
	switch boundaryType {
	case BoundaryDivergent:
		if p1.Type == PlateOceanic && p2.Type == PlateOceanic {
			return -2000 // Mid-ocean ridge (relative to ocean floor at -4000)
		} else if p1.Type == PlateContinental && p2.Type == PlateContinental {
			return -200 // Rift valley
		}
		return 100 // Mixed

	case BoundaryConvergent:
		if p1.Type == PlateOceanic && p2.Type == PlateOceanic {
			return -8000 // Oceanic trench (Mariana-scale)
		} else if p1.Type == PlateContinental && p2.Type == PlateContinental {
			return 6000 // Himalaya-scale mountains
		}
		return 4000 // Oceanic-Continental (Andes-scale coastal mountains)

	case BoundaryTransform:
		return 0 // No significant elevation change
	}
	return 0
}

// calculateEquilibriumElevationChange returns the delta to apply using an asymptotic approach.
// Instead of adding fixed amounts, we move toward a target elevation at a convergence rate.
// This prevents runaway elevation accumulation over geological time.
func calculateEquilibriumElevationChange(p1, p2 TectonicPlate, boundaryType BoundaryType, currentElev float64) float64 {
	target := GetTargetElevation(p1, p2, boundaryType)

	// Calculate difference and apply convergence rate
	// This creates an asymptotic approach: delta = (target - current) * rate
	difference := target - currentElev
	delta := difference * TectonicConvergenceRate

	return delta
}

// calculateElevationChange returns the elevation modifier based on boundary type.
// Deprecated: Use calculateEquilibriumElevationChange for equilibrium-based tectonics.
// Kept for backward compatibility with tests.
func calculateElevationChange(p1, p2 TectonicPlate, boundaryType BoundaryType) float64 {
	return GetTargetElevation(p1, p2, boundaryType)
}

// applyBoundaryEffectSpherical applies elevation change with falloff on sphere.
// Uses equilibrium-based approach with hard clamping to physical limits.
func applyBoundaryEffectSpherical(hm *SphereHeightmap, center spatial.Coordinate, elevationChange float64, topology spatial.Topology) {
	pixelRadius := 5

	// Simple falloff using BFS from center
	visited := make(map[spatial.Coordinate]struct{})
	type queueItem struct {
		coord    spatial.Coordinate
		distance int
	}
	queue := []queueItem{{coord: center, distance: 0}}
	visited[center] = struct{}{}

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.distance > pixelRadius {
			continue
		}

		// Calculate falloff factor
		dist := float64(current.distance)
		factor := 1.0 - dist/float64(pixelRadius)
		factor = factor * factor // Square for smoother falloff

		// Apply elevation change with physical limits
		currentElev := hm.Get(current.coord)
		newElev := currentElev + elevationChange*factor

		// Clamp to physical limits
		if newElev > MaxElevation {
			newElev = MaxElevation
		}
		if newElev < MinElevation {
			newElev = MinElevation
		}
		hm.Set(current.coord, newElev)

		// Expand to neighbors
		if current.distance < pixelRadius {
			for _, dir := range directions {
				neighbor := topology.GetNeighbor(current.coord, dir)
				if _, exists := visited[neighbor]; !exists {
					visited[neighbor] = struct{}{}
					queue = append(queue, queueItem{coord: neighbor, distance: current.distance + 1})
				}
			}
		}
	}
}

// SimulateGeologicalAge returns the plate count and surface description for an age
func SimulateGeologicalAge(age GeologicalAge) (int, string) {
	if age == AgeHadean {
		return 0, "molten"
	} else if age == AgeArchean {
		return 3, "cratons" // Small proto-plates
	} else if age == AgeProterozoic {
		return 7, "stable_continents"
	}
	return 12, "modern_plates"
}

// SimulateWilsonCycle determines the tectonic phase based on time
func SimulateWilsonCycle(years int64) string {
	// Wilson Cycle is ~500 million years
	cyclePos := years % 500_000_000

	if cyclePos < 100_000_000 {
		return "Rifting"
	} else if cyclePos < 200_000_000 {
		return "OceanFloorSpreading"
	} else if cyclePos < 400_000_000 {
		return "Subduction"
	}
	return "Orogeny" // Assembly/Collision
}

// SimulateContinentalRift calculates effects of rifting
func SimulateContinentalRift(isDivergent bool) (hasRift bool, volcanicActivity float64) {
	if isDivergent {
		return true, 0.8 // High volcanic activity
	}
	return false, 0.0
}

// CalculateSupercontinentEffects returns climatic impacts
func CalculateSupercontinentEffects(pangaeaIndex float64) (desertPercent float64, speciationRate float64) {
	if pangaeaIndex > 0.8 {
		return 0.6, 0.4 // High desert, low speciation (connected land)
	}
	return 0.1, 1.0
}
