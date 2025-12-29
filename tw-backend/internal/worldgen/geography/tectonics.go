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
	// ContinentalRigidity controls how far continental crust effects propagate (rings)
	ContinentalRigidity = 3
	// OceanicRigidity controls how far oceanic crust effects propagate (rings)
	OceanicRigidity = 1
)

// FeatureType describes the tectonic feature created at a boundary
type FeatureType string

const (
	FeatureNone            FeatureType = "none"
	FeatureTrench          FeatureType = "trench"
	FeatureIslandArc       FeatureType = "island_arc"
	FeatureCoastalMountain FeatureType = "coastal_mountain"
	FeatureOrogeny         FeatureType = "orogeny"
	FeatureMidOceanRidge   FeatureType = "mid_ocean_ridge"
	FeatureRiftValley      FeatureType = "rift_valley"
)

// CollisionResult describes the tectonic outcome at a specific cell
type CollisionResult struct {
	TargetElevation float64     // Target elevation for this cell
	Feature         FeatureType // Type of tectonic feature created
	RigidityRings   int         // How many rings the effect should propagate
}

// GetPlateDensity returns the density of a plate in g/cm³
// Oceanic crust is denser (~3.0 g/cm³) than continental (~2.7 g/cm³)
// Older oceanic crust is also denser due to cooling
func GetPlateDensity(p TectonicPlate) float64 {
	if p.Type == PlateOceanic {
		// Base oceanic density + age factor (older = denser due to cooling)
		// Age is in million years, add 0.001 g/cm³ per million years
		return 3.0 + (p.Age * 0.001)
	}
	// Continental crust has lower density
	return 2.7
}

// CalculateCollisionResult determines the tectonic outcome for a specific cell
// at a plate boundary based on crust physics.
// cellPlate is the plate the cell belongs to, neighborPlate is the adjacent plate.
func CalculateCollisionResult(cellPlate, neighborPlate TectonicPlate, boundaryType BoundaryType) CollisionResult {
	// Divergent boundaries - spreading/rifting
	if boundaryType == BoundaryDivergent {
		if cellPlate.Type == PlateOceanic && neighborPlate.Type == PlateOceanic {
			return CollisionResult{
				TargetElevation: -2500, // Mid-ocean ridge (elevated from -4000 ocean floor)
				Feature:         FeatureMidOceanRidge,
				RigidityRings:   OceanicRigidity,
			}
		}
		if cellPlate.Type == PlateContinental && neighborPlate.Type == PlateContinental {
			return CollisionResult{
				TargetElevation: -200, // Continental rift valley
				Feature:         FeatureRiftValley,
				RigidityRings:   ContinentalRigidity,
			}
		}
		// Mixed: use cell's type to determine rigidity
		rings := OceanicRigidity
		if cellPlate.Type == PlateContinental {
			rings = ContinentalRigidity
		}
		return CollisionResult{
			TargetElevation: 100,
			Feature:         FeatureNone,
			RigidityRings:   rings,
		}
	}

	// Transform boundaries - minimal elevation change
	if boundaryType == BoundaryTransform {
		return CollisionResult{
			TargetElevation: 0,
			Feature:         FeatureNone,
			RigidityRings:   OceanicRigidity,
		}
	}

	// Convergent boundaries - the complex case
	cellDensity := GetPlateDensity(cellPlate)
	neighborDensity := GetPlateDensity(neighborPlate)

	// Ocean vs Ocean: Denser (older) plate subducts
	if cellPlate.Type == PlateOceanic && neighborPlate.Type == PlateOceanic {
		if cellDensity >= neighborDensity {
			// This cell's plate subducts -> this cell becomes a trench
			return CollisionResult{
				TargetElevation: -8000 - (cellPlate.Age * 20), // Older = deeper trench, -8000 to -10000m
				Feature:         FeatureTrench,
				RigidityRings:   OceanicRigidity,
			}
		}
		// Neighbor subducts -> this cell becomes island arc (volcanic island chain)
		// Island arcs are volcanic peaks above subduction zones
		// Base elevation 500m creates emergent islands; noise variation adds peaks (+1000m)
		// and rigidity falloff (1 ring only) keeps features narrow, creating gaps between peaks
		return CollisionResult{
			TargetElevation: 500, // Above sea level - volcanic peaks emerge
			Feature:         FeatureIslandArc,
			RigidityRings:   OceanicRigidity,
		}
	}

	// Ocean vs Continent: Ocean always subducts
	if cellPlate.Type == PlateOceanic && neighborPlate.Type == PlateContinental {
		// This cell is oceanic, it subducts -> trench
		return CollisionResult{
			TargetElevation: -6000,
			Feature:         FeatureTrench,
			RigidityRings:   OceanicRigidity,
		}
	}
	if cellPlate.Type == PlateContinental && neighborPlate.Type == PlateOceanic {
		// This cell is continental, neighbor subducts -> coastal mountains (Andes-style)
		return CollisionResult{
			TargetElevation: 3500 + (cellPlate.Thickness * 30), // Thicker crust = higher mountains, 3000-5000m
			Feature:         FeatureCoastalMountain,
			RigidityRings:   ContinentalRigidity,
		}
	}

	// Continent vs Continent: Buckling -> massive orogeny
	return CollisionResult{
		TargetElevation: 6000 + (cellPlate.Thickness * 50), // Thicker = higher, 6000-8800m
		Feature:         FeatureOrogeny,
		RigidityRings:   ContinentalRigidity,
	}
}

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

		// Randomly assign type (30% continental, 70% oceanic)
		// Previously: first N plates were always continental, now truly random
		plateType := PlateOceanic
		thickness := 5 + r.Float64()*5 // 5-10km oceanic crust
		if r.Float64() < 0.3 {
			plateType = PlateContinental
			thickness = 30 + r.Float64()*20 // 30-50km continental crust
		}

		// Age range 0-200 million years for better density variation
		// (older oceanic crust = denser = more likely to subduct)
		age := r.Float64() * 200

		plates[i] = TectonicPlate{
			ID:        uuid.New(),
			Type:      plateType,
			Centroid:  centroid,
			Position:  position,
			Velocity:  velocity,
			Region:    make(map[spatial.Coordinate]struct{}),
			Thickness: thickness,
			Age:       age,
		}

		if debug.Is(debug.Geology) {
			log.Printf("[PLATE INIT] Plate %d: Type=%v Age=%.1fMy Thickness=%.1fkm",
				i, plateType, age, thickness)
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

	// IMPORTANT: Clear existing regions before reassignment to prevent memory leak
	for i := range plates {
		plates[i].Region = make(map[spatial.Coordinate]struct{})
	}

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

		// Apply equilibrium-based elevation change with collision physics
		currentElev := heightmap.Get(bc.Coord)
		elevationDelta, collisionResult := calculateEquilibriumElevationChangeV2(currentPlate, neighborPlate, bc.BoundaryType, currentElev)

		// Apply scale factor for variable time steps
		elevationDelta *= scaleFactor

		// Use rigidity-aware boundary effect
		applyBoundaryEffectWithRigidity(heightmap, bc.Coord, elevationDelta, collisionResult.RigidityRings, topology)
	}

	return heightmap
}

// PassiveMarginDecayRate controls how fast old boundaries erode back to base elevation
// Value of 0.02 means 2% of remaining difference per tectonic step (very slow)
const PassiveMarginDecayRate = 0.02

// ApplyBoundaryDecay erodes cells that are NO LONGER at plate boundaries toward base elevation.
// This prevents "phantom mountains" from persisting after plate boundaries move away.
// Should be called after SimulateTectonicsWithCache to handle passive margins.
func ApplyBoundaryDecay(plates []TectonicPlate, heightmap *SphereHeightmap, cache *BoundaryCache, topology spatial.Topology, scaleFactor float64) {
	if debug.Is(debug.Perf) {
		defer debug.Time(debug.Perf, "ApplyBoundaryDecay")()
	}

	resolution := topology.Resolution()

	// Build a set of boundary cells for O(1) lookup
	boundarySet := make(map[spatial.Coordinate]struct{}, len(cache.Cells))
	for _, bc := range cache.Cells {
		boundarySet[bc.Coord] = struct{}{}
	}

	// Build plate lookup grid for fast plate assignment
	plateGrid := make([]int, 6*resolution*resolution)
	for i := range plateGrid {
		plateGrid[i] = -1
	}
	for i, p := range plates {
		for coord := range p.Region {
			idx := coord.Face*resolution*resolution + coord.Y*resolution + coord.X
			if idx >= 0 && idx < len(plateGrid) {
				plateGrid[idx] = i
			}
		}
	}

	// Iterate all cells
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}

				// Skip cells that ARE at boundaries - they get tectonic uplift
				if _, isBoundary := boundarySet[coord]; isBoundary {
					continue
				}

				// Get plate for this cell
				idx := face*resolution*resolution + y*resolution + x
				plateIdx := plateGrid[idx]
				if plateIdx < 0 {
					continue
				}
				plate := plates[plateIdx]

				// Determine base elevation for this plate type
				baseElev := -4000.0 // Ocean floor
				if plate.Type == PlateContinental {
					baseElev = 100.0 // Continental shelf
				}

				// Get current elevation
				currentElev := heightmap.Get(coord)

				// Apply slow decay toward base elevation (isostatic rebound)
				// This makes old mountains erode and old ocean ridges sink
				difference := baseElev - currentElev
				delta := difference * PassiveMarginDecayRate * scaleFactor

				// Apply decay
				newElev := currentElev + delta
				newElev = clampElevation(newElev)
				heightmap.Set(coord, newElev)
			}
		}
	}
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
// Deprecated: Use calculateEquilibriumElevationChangeV2 for collision-aware physics.
func calculateEquilibriumElevationChange(p1, p2 TectonicPlate, boundaryType BoundaryType, currentElev float64) float64 {
	result := CalculateCollisionResult(p1, p2, boundaryType)
	target := result.TargetElevation

	// Calculate difference and apply convergence rate
	// This creates an asymptotic approach: delta = (target - current) * rate
	difference := target - currentElev
	delta := difference * TectonicConvergenceRate

	return delta
}

// calculateEquilibriumElevationChangeV2 returns both the elevation delta and collision result.
// This enables the caller to apply rigidity-aware boundary effects.
func calculateEquilibriumElevationChangeV2(cellPlate, neighborPlate TectonicPlate, boundaryType BoundaryType, currentElev float64) (float64, CollisionResult) {
	result := CalculateCollisionResult(cellPlate, neighborPlate, boundaryType)
	target := result.TargetElevation

	// Calculate difference and apply convergence rate
	difference := target - currentElev
	delta := difference * TectonicConvergenceRate

	return delta, result
}

// calculateElevationChange returns the elevation modifier based on boundary type.
// Deprecated: Use calculateEquilibriumElevationChange for equilibrium-based tectonics.
// Kept for backward compatibility with tests.
func calculateElevationChange(p1, p2 TectonicPlate, boundaryType BoundaryType) float64 {
	return GetTargetElevation(p1, p2, boundaryType)
}

// applyBoundaryEffectSpherical applies elevation change at a boundary cell.
// OPTIMIZED: Uses 2-ring falloff for smoother terrain without full BFS.
func applyBoundaryEffectSpherical(hm *SphereHeightmap, center spatial.Coordinate, elevationChange float64, topology spatial.Topology) {
	// Ring 0: Center (100% effect)
	currentElev := hm.Get(center)
	newElev := currentElev + elevationChange
	newElev = clampElevation(newElev)
	hm.Set(center, newElev)

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Track visited to prevent double-application in rings
	// Since we only go 2 rings out, a small map or simply iterating carefully is needed.
	// For efficiency/simplicity with just 2 rings, we can just use a visited map.
	visited := map[spatial.Coordinate]struct{}{
		center: {},
	}

	// Ring 1: Immediate neighbors (50% effect)
	ring1 := make([]spatial.Coordinate, 0, 4)
	for _, dir := range directions {
		neighbor := topology.GetNeighbor(center, dir)
		if _, exists := visited[neighbor]; !exists {
			visited[neighbor] = struct{}{}
			ring1 = append(ring1, neighbor)

			nElev := hm.Get(neighbor)
			nNewElev := nElev + elevationChange*0.5
			nNewElev = clampElevation(nNewElev)
			hm.Set(neighbor, nNewElev)
		}
	}

	// Ring 2: Neighbors of Ring 1 (25% effect)
	for _, r1Coord := range ring1 {
		for _, dir := range directions {
			neighbor := topology.GetNeighbor(r1Coord, dir)
			if _, exists := visited[neighbor]; !exists {
				visited[neighbor] = struct{}{}

				nElev := hm.Get(neighbor)
				nNewElev := nElev + elevationChange*0.25
				nNewElev = clampElevation(nNewElev)
				hm.Set(neighbor, nNewElev)
			}
		}
	}
}

// applyBoundaryEffectWithRigidity applies elevation change with variable propagation based on crustal rigidity.
// Continental crust is rigid: effects propagate further (3 rings).
// Oceanic crust is plastic: effects are localized (1 ring).
// This creates realistic coastal mountain ranges while keeping oceanic features narrow.
func applyBoundaryEffectWithRigidity(hm *SphereHeightmap, center spatial.Coordinate, elevationChange float64, rigidityRings int, topology spatial.Topology) {
	// Ring 0: Center (100% effect)
	currentElev := hm.Get(center)
	newElev := currentElev + elevationChange
	newElev = clampElevation(newElev)
	hm.Set(center, newElev)

	if rigidityRings <= 0 {
		return
	}

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Track visited to prevent double-application
	visited := map[spatial.Coordinate]struct{}{
		center: {},
	}

	// Build rings dynamically based on rigidity
	currentRing := []spatial.Coordinate{center}

	for ring := 1; ring <= rigidityRings; ring++ {
		// Calculate falloff: each ring gets proportionally less effect
		// Ring 1: 50%, Ring 2: 25%, Ring 3: 12.5%, etc.
		falloff := 1.0 / float64(uint(1)<<uint(ring)) // 0.5, 0.25, 0.125...

		nextRing := make([]spatial.Coordinate, 0, len(currentRing)*4)

		for _, coord := range currentRing {
			for _, dir := range directions {
				neighbor := topology.GetNeighbor(coord, dir)
				if _, exists := visited[neighbor]; !exists {
					visited[neighbor] = struct{}{}
					nextRing = append(nextRing, neighbor)

					nElev := hm.Get(neighbor)
					nNewElev := nElev + elevationChange*falloff
					nNewElev = clampElevation(nNewElev)
					hm.Set(neighbor, nNewElev)
				}
			}
		}

		currentRing = nextRing
	}
}

func clampElevation(elev float64) float64 {
	if elev > MaxElevation {
		return MaxElevation
	}
	if elev < MinElevation {
		return MinElevation
	}
	return elev
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
