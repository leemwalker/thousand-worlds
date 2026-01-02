package geography

import (
	"container/heap"
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

// Isostatic base elevations (in meters)
// These represent the "natural floating height" of each crust type on the mantle.
const (
	// ContinentalBaseElevation is the natural elevation of continental crust (+20m lowlands)
	// Lowered from 150m to allow for more varied terrain and less "blocky" continents
	ContinentalBaseElevation = 20.0
	// OceanicBaseElevation is the natural elevation of oceanic crust (-4000m abyssal plain)
	OceanicBaseElevation = -4000.0
	// IsostaticRelaxationRate is how quickly crust drifts toward base (5% per step)
	// This allows tectonic deformations to persist while slowly returning toward equilibrium
	IsostaticRelaxationRate = 0.05
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
	TargetElevation float64     // Target elevation for this cell (calculated via isostasy)
	NewThickness    float64     // Resulting crustal thickness in km after collision
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
// at a plate boundary based on crust physics and isostasy.
// Uses mass conservation: collisions cause crustal thickening, which then
// determines elevation via Archimedes' buoyancy principle.
func CalculateCollisionResult(cellPlate, neighborPlate TectonicPlate, boundaryType BoundaryType) CollisionResult {
	// Divergent boundaries - spreading/rifting (thinning)
	if boundaryType == BoundaryDivergent {
		if cellPlate.Type == PlateOceanic && neighborPlate.Type == PlateOceanic {
			// Mid-ocean ridge: new thin crust forms
			newThickness := 4.0 // Fresh, thin oceanic crust at ridge
			elevation := CalculateIsostaticHeight(newThickness, DensityBasalt)
			return CollisionResult{
				TargetElevation: elevation, // Elevated from abyssal (~-2500m)
				NewThickness:    newThickness,
				Feature:         FeatureMidOceanRidge,
				RigidityRings:   OceanicRigidity,
			}
		}
		if cellPlate.Type == PlateContinental && neighborPlate.Type == PlateContinental {
			// Continental rift: crust thins
			newThickness := cellPlate.Thickness * 0.7 // 30% thinning
			elevation := CalculateIsostaticHeight(newThickness, DensityGranite)
			return CollisionResult{
				TargetElevation: elevation, // Lowered rift valley
				NewThickness:    newThickness,
				Feature:         FeatureRiftValley,
				RigidityRings:   ContinentalRigidity,
			}
		}
		// Mixed: use cell's type to determine rigidity and elevation
		rings := OceanicRigidity
		density := DensityBasalt
		if cellPlate.Type == PlateContinental {
			rings = ContinentalRigidity
			density = DensityGranite
		}
		elevation := CalculateIsostaticHeight(cellPlate.Thickness, density)
		return CollisionResult{
			TargetElevation: elevation,
			NewThickness:    cellPlate.Thickness,
			Feature:         FeatureNone,
			RigidityRings:   rings,
		}
	}

	// Transform boundaries - minimal elevation change (no thickening)
	if boundaryType == BoundaryTransform {
		density := DensityBasalt
		if cellPlate.Type == PlateContinental {
			density = DensityGranite
		}
		elevation := CalculateIsostaticHeight(cellPlate.Thickness, density)
		return CollisionResult{
			TargetElevation: elevation,
			NewThickness:    cellPlate.Thickness,
			Feature:         FeatureNone,
			RigidityRings:   OceanicRigidity,
		}
	}

	// =========================================================================
	// Convergent boundaries - Mass Conservation Model
	// =========================================================================
	cellDensity := GetPlateDensity(cellPlate)
	neighborDensity := GetPlateDensity(neighborPlate)

	// Ocean vs Ocean: Denser (older) plate subducts
	if cellPlate.Type == PlateOceanic && neighborPlate.Type == PlateOceanic {
		if cellDensity >= neighborDensity {
			// This cell's plate subducts -> deep trench
			// Trenches are pulled down by subducting slab, use very thin effective thickness
			trenchThickness := 2.0 // Effectively pulled down
			// Manual override for trenches: they go deeper than isostasy alone
			trenchElevation := -8000 - (cellPlate.Age * 20) // -8000 to -10000m
			return CollisionResult{
				TargetElevation: trenchElevation,
				NewThickness:    trenchThickness,
				Feature:         FeatureTrench,
				RigidityRings:   OceanicRigidity,
			}
		}
		// Neighbor subducts -> island arc (volcanic thickening)
		// Volcanism adds ~5km of material to overriding plate
		newThickness := cellPlate.Thickness + 5.0
		elevation := CalculateIsostaticHeight(newThickness, DensityBasalt)
		return CollisionResult{
			TargetElevation: elevation, // Still below sea level but elevated (~-4300m)
			NewThickness:    newThickness,
			Feature:         FeatureIslandArc,
			RigidityRings:   OceanicRigidity,
		}
	}

	// Ocean vs Continent: Ocean always subducts
	if cellPlate.Type == PlateOceanic && neighborPlate.Type == PlateContinental {
		// Oceanic cell subducts -> trench
		trenchElevation := -6000 - (cellPlate.Age * 10) // Deep trench
		return CollisionResult{
			TargetElevation: trenchElevation,
			NewThickness:    2.0, // Pulled down
			Feature:         FeatureTrench,
			RigidityRings:   OceanicRigidity,
		}
	}
	if cellPlate.Type == PlateContinental && neighborPlate.Type == PlateOceanic {
		// Continental cell: crumples as ocean subducts under it (Andes-style)
		// Continental crust thickens by ~10km from compression and volcanic addition
		newThickness := cellPlate.Thickness + 10.0
		elevation := CalculateIsostaticHeight(newThickness, DensityGranite)
		return CollisionResult{
			TargetElevation: elevation, // Typically 2500-4000m
			NewThickness:    newThickness,
			Feature:         FeatureCoastalMountain,
			RigidityRings:   ContinentalRigidity,
		}
	}

	// Continent vs Continent: Massive folding -> orogeny (Himalayas)
	// Plates fold together: NewThickness = (T1 + T2) * 0.8 (mass conservation with compression loss)
	combinedThickness := (cellPlate.Thickness + neighborPlate.Thickness) * 0.8
	elevation := CalculateIsostaticHeight(combinedThickness, DensityGranite)
	return CollisionResult{
		TargetElevation: elevation, // 4000-6500m depending on combined thickness
		NewThickness:    combinedThickness,
		Feature:         FeatureOrogeny,
		RigidityRings:   ContinentalRigidity,
	}
}

// GeneratePlates creates tectonic plates using spherical topology.
// Uses Multi-Source BFS to assign regions efficiently in O(N) time.
// Plate types are assigned by AREA to guarantee ~continentalPerc continental coverage:
// largest plates become continental until target % of total area is covered.
func GeneratePlates(count int, topology spatial.Topology, seed int64, continentalPerc float64) []TectonicPlate {
	r := rand.New(rand.NewSource(seed))
	resolution := topology.Resolution()
	plates := make([]TectonicPlate, count)

	// 1. Initialize plates with random centroids distributed across all faces
	// Type/Thickness will be assigned AFTER BFS based on area
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

		// Age range 0-200 million years for better density variation
		// (older oceanic crust = denser = more likely to subduct)
		age := r.Float64() * 200

		// Initialize as Oceanic (will be reassigned after BFS)
		plates[i] = TectonicPlate{
			ID:          uuid.New(),
			Type:        PlateOceanic, // Default, will be reassigned
			Centroid:    centroid,
			Position:    position,
			Velocity:    velocity,
			Region:      make(map[spatial.Coordinate]struct{}),
			Thickness:   6.0 + r.Float64()*4.0, // 6-10km oceanic crust
			MeanDensity: DensityBasalt,         // 3000 kg/m³
			Age:         age,
		}
	}

	// 2. Multi-Source BFS to assign all cells to nearest plate
	ReassignPlateRegions(plates, topology)

	// 3. Sort plates by area (region size) descending
	// Use a simple index sort to find largest plates
	type plateArea struct {
		index int
		area  int
	}
	areas := make([]plateArea, count)
	totalCells := 0
	for i, plate := range plates {
		areas[i] = plateArea{index: i, area: len(plate.Region)}
		totalCells += len(plate.Region)
	}

	// Sort descending by area (bubble sort for small count, typically 5-15 plates)
	for i := 0; i < len(areas)-1; i++ {
		for j := i + 1; j < len(areas); j++ {
			if areas[j].area > areas[i].area {
				areas[i], areas[j] = areas[j], areas[i]
			}
		}
	}

	// 4. Assign Continental to largest plates until target % area
	targetContinentalArea := float64(totalCells) * continentalPerc
	coveredArea := 0.0

	for _, pa := range areas {
		// If we've met the target (or target is 0.0), stop
		if coveredArea >= targetContinentalArea {
			break
		}

		// Assign this plate as Continental
		plates[pa.index].Type = PlateContinental
		plates[pa.index].Thickness = 30.0 + r.Float64()*10.0 // 30-40km continental crust
		plates[pa.index].MeanDensity = DensityGranite        // 2700 kg/m³
		coveredArea += float64(pa.area)

		if debug.Is(debug.Geology) {
			log.Printf("[PLATE INIT] Plate %d: Type=Continental Area=%d (%.1f%% of target)",
				pa.index, pa.area, coveredArea/targetContinentalArea*100)
		}
	}

	if debug.Is(debug.Geology) {
		log.Printf("[PLATE INIT] Continental coverage: %.1f%% (target: %.1f%%)",
			coveredArea/float64(totalCells)*100, continentalPerc*100)
	}

	// 5. Initialize Dynamic Motion (Phase 8a)
	InitializePlateMotion(plates, r)

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

	// DEBUG: Log cache stats and first few deltas
	if debug.Is(debug.Geology) && len(cache.Cells) > 0 {
		log.Printf("[TECTONICS] BoundaryCache cells: %d, scaleFactor: %.2f", len(cache.Cells), scaleFactor)

		// Sample first convergent boundary
		for i := 0; i < len(cache.Cells) && i < 5; i++ {
			bc := cache.Cells[i]
			currentPlate := plates[bc.PlateIdx]
			neighborPlate := plates[bc.NeighborIdx]
			currentElev := heightmap.Get(bc.Coord)
			delta, result := calculateEquilibriumElevationChangeV2(currentPlate, neighborPlate, bc.BoundaryType, currentElev)
			log.Printf("[TECTONICS SAMPLE %d] Type=%s Cell=%s->%s Elev=%.0f Target=%.0f Delta=%.0f Scaled=%.0f",
				i, bc.BoundaryType, currentPlate.Type, neighborPlate.Type,
				currentElev, result.TargetElevation, delta, delta*scaleFactor)
		}
	}

	// Process only cached boundary cells
	for _, bc := range cache.Cells {
		currentPlate := plates[bc.PlateIdx]
		neighborPlate := plates[bc.NeighborIdx]

		// Apply equilibrium-based elevation change with collision physics
		currentElev := heightmap.Get(bc.Coord)
		elevationDelta, collisionResult := calculateEquilibriumElevationChangeV2(currentPlate, neighborPlate, bc.BoundaryType, currentElev)

		// Phase 8b: Crustal Accretion Logic
		// Transform Oceanic Crust -> Continental (Island Arc) via Magmatic Differentiation
		if collisionResult.Feature == FeatureIslandArc {
			// Calculate accretion flux based on convergence rate
			// Flux ~ Velocity * Time (scaleFactor)
			// Using constant flux for now for simplicity, roughly 0.1 unit per tick
			flux := 0.1 * scaleFactor

			// Add to plate's total accreted mass
			plates[bc.PlateIdx].AccretedMass += flux

			// Check local cell data
			cellData := heightmap.GetCellData(bc.Coord)
			if !cellData.IsContinental {
				// Probability of arc formation increases with total accreted mass
				// or simple threshold. Let's use a local accumulation if we had it,
				// but for now use the global plate mass as a proxy for "active volcanic era"
				// combined with checking if we are THE boundary cell.

				// Simplified: Just use a probability roll scaled by flux
				// This simulates that arcs form at specific points, not everywhere at once
				if rand.Float64() < (0.05 * scaleFactor) {
					cellData.IsContinental = true
					// Raise elevation immediately to sea level to simulate rapid volcano growth
					if currentElev < -500 {
						heightmap.Set(bc.Coord, -500) // Just below surface
					}
					heightmap.SetCellData(bc.Coord, cellData)

					// Update plate's continental area tracking
					plates[bc.PlateIdx].ContinentalArea += 1 // Approx 100km^2
				}
			}
		}

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

// ApplyIsostaticRelaxation drifts each cell toward its base elevation based on plate type.
// This implements the physics of crustal isostasy: denser oceanic crust sinks (-4000m)
// while lighter continental crust floats (+150m).
// Over time, this removes "ocean wall" artifacts from old collisions while preserving
// tectonic deformations (mountains, trenches) which are reapplied each step.
// relaxationRate controls the speed (0.05 = 5% per step toward base).
func ApplyIsostaticRelaxation(plates []TectonicPlate, heightmap *SphereHeightmap, topology spatial.Topology, relaxationRate float64) {
	resolution := topology.Resolution()
	totalCells := 6 * resolution * resolution
	resSq := resolution * resolution

	// Build plate lookup grid (same as SimulateTectonics)
	plateGrid := make([]int, totalCells)
	for i := range plateGrid {
		plateGrid[i] = -1
	}
	for i, p := range plates {
		for coord := range p.Region {
			idx := (coord.Face * resSq) + (coord.Y * resolution) + coord.X
			if idx >= 0 && idx < totalCells {
				plateGrid[idx] = i
			}
		}
	}

	// Apply relaxation to each cell
	for idx := 0; idx < totalCells; idx++ {
		plateIdx := plateGrid[idx]
		if plateIdx == -1 {
			continue
		}

		// Reconstruct coordinate
		face := idx / resSq
		rem := idx % resSq
		y := rem / resolution
		x := rem % resolution
		coord := spatial.Coordinate{Face: face, X: x, Y: y}

		// Determine target base elevation from plate type OR cell data
		var baseElevation float64
		cellData := heightmap.GetCellData(coord)

		// If cell is Continental (Island Arc/Terrane) OR Plate is Continental -> High elevation
		if cellData.IsContinental || plates[plateIdx].Type == PlateContinental {
			baseElevation = ContinentalBaseElevation // +20m
		} else {
			baseElevation = OceanicBaseElevation // -4000m
		}

		// Lerp current elevation toward base
		// newElev = current + (target - current) * rate
		currentElev := heightmap.Get(coord)
		newElev := currentElev + (baseElevation-currentElev)*relaxationRate
		heightmap.Set(coord, newElev)
	}
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
// Continental crust is rigid: effects propagate further (3 rings) with gradual falloff.
// Oceanic crust is plastic: effects are localized (1-2 rings).
// UPDATED: Uses non-geometric falloff for wider, more realistic mountain ranges:
// - Continental (3 rings): 100% -> 60% -> 20% (gradual highlands/foothills)
// - Oceanic (1 ring): 100% -> 50% (narrow ridges)
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

	// Lookup table for falloff percentages (non-geometric for realistic orogeny)
	// Continental (3 rings): gradual transition from peaks to foothills
	// Oceanic (1-2 rings): narrow, localized ridges
	var falloffTable []float64
	if rigidityRings >= 3 {
		// Continental: 60% highlands, 20% foothills (creates wide mountain ranges)
		falloffTable = []float64{0.60, 0.20, 0.10}
	} else if rigidityRings == 2 {
		// Mixed: 50%, 20%
		falloffTable = []float64{0.50, 0.20}
	} else {
		// Oceanic: 50% (narrow ridge)
		falloffTable = []float64{0.50}
	}

	// Build rings dynamically based on rigidity
	currentRing := []spatial.Coordinate{center}

	// Jittered Uplift Logic (Refinement Task 2):
	// Instead of perfectly centered uplift, we distribute force to scatter peaks.
	// 70% to Center/Target, 30% to a Random Neighbor.

	// Apply to Center (70%)
	amountCenter := elevationChange * 0.7
	currentElev = hm.Get(center)
	newElev = currentElev + amountCenter
	newElev = clampElevation(newElev)
	hm.Set(center, newElev)

	// Pick Random Neighbor for Jitter (30%)
	// Use a simple hash or rand based on coordinate to keep it deterministic but "random"
	// Or just use the first neighbor from a shuffled list?
	// To ensure determinism, we use a hash of the coordinate.
	dirIdx := (center.X + center.Y + center.Face) % 4
	jitterDir := directions[dirIdx]
	jitterNeighbor := topology.GetNeighbor(center, jitterDir)

	amountJitter := elevationChange * 0.3
	jitterElev := hm.Get(jitterNeighbor)
	newJitterElev := jitterElev + amountJitter
	newJitterElev = clampElevation(newJitterElev)
	hm.Set(jitterNeighbor, newJitterElev)

	// Continue with standard rigidity propagation for the rest?
	// The prompt implies this "applyBoundaryEffect" is the main mechanism.
	// If we smudge the CENTER, the rings will propagate from the center.
	// But `applyBoundaryEffectWithRigidity` applies to rings AROUND the center.
	// If we split the center force, should we propagate from BOTH?
	// That might be expensive.
	// Let's stick to the prompt: "Refactor applyBoundaryEffect... Apply 70% to Target, 30% to Random Neighbor"

	// For RIGIDITY rings, we will propagate from the MAIN center as before,
	// but using the remaining "impulse"?
	// The original code applied `elevationChange` to center, then `0.6 * elevationChange` to ring 1.
	// If we reduce center to 0.7, does ring 1 still get 0.6 of TOTAL?
	// Assume yes, the ring falloff is separate scaling.

	// Iterate Rings for falloff (using original center for simplicity of propagation)
	visited[jitterNeighbor] = struct{}{} // Mark jitter neighbor as visited so it doesn't get double applied in rings

	for ring := 1; ring <= rigidityRings; ring++ {
		// Use lookup table for falloff
		falloffIdx := ring - 1
		if falloffIdx >= len(falloffTable) {
			falloffIdx = len(falloffTable) - 1
		}
		falloff := falloffTable[falloffIdx]

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

// =============================================================================
// Geological Province Generation (Phase 5)
// =============================================================================

// ProvinceQueue for Dijkstra
type ProvinceItem struct {
	Coordinate spatial.Coordinate
	ProvinceID int
	Cost       float64
	Index      int
}

type ProvinceQueue []*ProvinceItem

func (pq ProvinceQueue) Len() int { return len(pq) }
func (pq ProvinceQueue) Less(i, j int) bool {
	return pq[i].Cost < pq[j].Cost
}
func (pq ProvinceQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *ProvinceQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*ProvinceItem)
	item.Index = n
	*pq = append(*pq, item)
}
func (pq *ProvinceQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

// GenerateProvinces creates geological sub-regions within continental plates.
// Uses Randomized Dijkstra to ensure organic, irregular shapes.
func GenerateProvinces(plates []TectonicPlate, topology spatial.Topology, seed int64) []GeologicalProvince {
	r := rand.New(rand.NewSource(seed))
	provinces := []GeologicalProvince{}
	provinceID := 1

	for _, plate := range plates {
		// Only generate provinces for continental plates
		if plate.Type != PlateContinental {
			continue
		}

		// Skip plates with no region
		if len(plate.Region) == 0 {
			continue
		}

		// Generate 3-5 provinces per continental plate
		numProvinces := 3 + r.Intn(3) // 3, 4, or 5

		// Convert plate region to slice for random access
		regionSlice := make([]spatial.Coordinate, 0, len(plate.Region))
		for coord := range plate.Region {
			regionSlice = append(regionSlice, coord)
		}

		// Pick random seeds within the plate
		for i := 0; i < numProvinces && i < len(regionSlice); i++ {
			// Pick random coordinate from plate region
			idx := r.Intn(len(regionSlice))
			seedCoord := regionSlice[idx]

			// Determine province type with weighted distribution
			// 40% Craton, 30% FoldBelt, 30% Basin
			var provType ProvinceType
			var hardness float64
			var deformation float64

			roll := r.Float64()
			if roll < 0.4 {
				provType = ProvinceCraton
				hardness = CratonHardness
				deformation = 0.1
			} else if roll < 0.7 {
				provType = ProvinceFoldBelt
				hardness = FoldBeltHardness
				deformation = 0.8
			} else {
				provType = ProvinceBasin
				hardness = BasinHardness
				deformation = 0.3
			}

			prov := GeologicalProvince{
				ID:          provinceID,
				Type:        provType,
				PlateID:     plate.ID,
				Hardness:    hardness,
				Deformation: deformation,
				SeedCoord:   seedCoord,
			}
			provinces = append(provinces, prov)
			provinceID++
		}
	}

	return provinces
}

// InitializeProvinceHardness assigns province IDs and hardness values to all cells
// in continental plates using Randomized Dijkstra for organic shapes.
func InitializeProvinceHardness(hm *SphereHeightmap, plates []TectonicPlate, provinces []GeologicalProvince, topology spatial.Topology, seed int64) {
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}
	r := rand.New(rand.NewSource(seed))

	// Build set of all continental cells for fast lookup
	continentalCells := make(map[spatial.Coordinate]struct{})
	for _, plate := range plates {
		if plate.Type == PlateContinental {
			for coord := range plate.Region {
				continentalCells[coord] = struct{}{}
			}
		}
	}

	// Initialize Priority Queue with seeds
	pq := &ProvinceQueue{}
	heap.Init(pq)

	assigned := make(map[spatial.Coordinate]int)

	for _, prov := range provinces {
		// Initial cost 0 for seeds
		heap.Push(pq, &ProvinceItem{
			Coordinate: prov.SeedCoord,
			ProvinceID: prov.ID,
			Cost:       0.0,
		})

		// Note: We don't mark assigned yet, we let Pop handle it to ensure lowest cost wins if seeds are close?
		// Actually for seeds we can just assign.
		assigned[prov.SeedCoord] = prov.ID

		// Set initial hardness
		data := hm.GetCellData(prov.SeedCoord)
		data.RockHardness = prov.Hardness
		data.ProvinceID = prov.ID
		hm.SetCellData(prov.SeedCoord, data)
	}

	// Build province lookup for hardness
	provinceLookup := make(map[int]float64)
	for _, prov := range provinces {
		provinceLookup[prov.ID] = prov.Hardness
	}

	// Randomized Dijkstra Expansion
	for pq.Len() > 0 {
		current := heap.Pop(pq).(*ProvinceItem)

		// Check all 4 neighbors
		for _, dir := range directions {
			neighbor := topology.GetNeighbor(current.Coordinate, dir)

			// Skip if not continental
			if _, isContinental := continentalCells[neighbor]; !isContinental {
				continue
			}

			// Skip if already assigned (Dijkstra guarantees first visit is optimal if weights are non-negative)
			if _, exists := assigned[neighbor]; exists {
				continue
			}

			// Calculate travel cost
			// Base Cost (1.0) + Random Variance (0.0 - 1.0)
			// This creates "wobbly" organic boundaries
			travelCost := 1.0 + r.Float64()
			newCost := current.Cost + travelCost

			// Assign
			assigned[neighbor] = current.ProvinceID

			// Set cell data
			hardness := provinceLookup[current.ProvinceID]
			data := hm.GetCellData(neighbor)
			data.RockHardness = hardness
			data.ProvinceID = current.ProvinceID
			hm.SetCellData(neighbor, data)

			// Push to PQ
			heap.Push(pq, &ProvinceItem{
				Coordinate: neighbor,
				ProvinceID: current.ProvinceID,
				Cost:       newCost,
			})
		}
	}
}
