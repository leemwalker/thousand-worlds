package geography

import (
	"math"
	"runtime"
	"sync"

	"tw-backend/internal/spatial"
)

// =============================================================================
// Phase 10: Glacial Geomorphology
// =============================================================================

// IceData represents ice sheet properties for a single grid cell.
type IceData struct {
	Thickness float64          // Ice thickness in meters (0 = no ice)
	FlowDir   spatial.Vector3D // Flow direction (normalized tangent vector)
	FlowSpeed float64          // Flow speed in m/year
	Age       float64          // Ice age in years (for tracking)
}

// IceSheet represents a continental or regional ice sheet system.
// Optimized for dense grid access (slice instead of map).
type IceSheet struct {
	// Grid of ice data indexed by flat index
	// Index = face*res*res + y*res + x
	Ice []IceData

	// Aggregate statistics
	TotalVolume  float64 // km³ of ice
	TotalArea    float64 // km² covered
	MaxThickness float64 // meters
	Resolution   int     // Grid resolution used for allocation

	// Sediment tracking for moraines
	Sediment []float64 // Accumulated sediment load

	// Scratch buffer for flow calculations (avoids allocation per Update call)
	deltaIce []float64
}

// NewIceSheet creates an empty ice sheet system.
// Requires resolution to allocate dense arrays.
func NewIceSheet(resolution int) *IceSheet {
	totalCells := 6 * resolution * resolution
	return &IceSheet{
		Ice:        make([]IceData, totalCells),
		Sediment:   make([]float64, totalCells),
		deltaIce:   make([]float64, totalCells), // Pre-allocate flow buffer
		Resolution: resolution,
	}
}

// Constants for ice physics
const (
	// IceAccumulationRate is m/year per precipitation unit (simplified)
	IceAccumulationRate = 0.5

	// IceAblationRate is m/year per degree above 0°C
	IceAblationRate = 2.0

	// IceDensity in kg/m³
	IceDensity = 917.0

	// IceFlowConstant (simplified SIA) - relates thickness to flow speed
	// v ~ A * τ^n where τ ~ ρgh * slope, n=3 for ice
	// Simplified: v = C * H^3 * slope
	IceFlowConstant = 1e-16

	// MinIceThickness below which ice disappears
	MinIceThickness = 1.0

	// ErosionCoefficient relates ice flux to bedrock erosion (mm/year per unit flux)
	ErosionCoefficient = 0.001

	// SedimentCapacity is max sediment load per unit ice thickness (arbitrary units)
	SedimentCapacity = 0.1
)

// Update processes ice accumulation, flow, and ablation for one time step.
// dt is in years, temperature is global average, precipitation is mm/year equivalent.
// topology is used for neighbor lookups, heightmap provides bedrock elevation.
func (is *IceSheet) Update(dt float64, tempGrid []float64, precipGrid []float64,
	heightmap *SphereHeightmap, topology spatial.Topology) {

	resolution := is.Resolution // Use stored resolution
	totalCells := 6 * resolution * resolution
	resSq := resolution * resolution

	// Helper to get coordinate from index (inlined for perf, or func)
	indexToCoord := func(idx int) spatial.Coordinate {
		face := idx / resSq
		rem := idx % resSq
		y := rem / resolution
		x := rem % resolution
		return spatial.Coordinate{Face: face, X: x, Y: y}
	}

	// Helper to get index from coordinate
	coordToIndex := func(c spatial.Coordinate) int {
		return c.Face*resSq + c.Y*resolution + c.X
	}

	// Phase 1: Accumulation & Ablation (Parallelized)
	workers := runtime.NumCPU()
	chunkSize := totalCells / workers
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if w == workers-1 {
			end = totalCells
		}

		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for idx := s; idx < e; idx++ {
				temp := tempGrid[idx]
				precip := precipGrid[idx]

				// Ice accumulates where temp < 0 and precipitation > 0
				if temp < 0 && precip > 0 {
					accumulation := precip * IceAccumulationRate * dt
					is.Ice[idx].Thickness += accumulation
				} else if temp > 0 && is.Ice[idx].Thickness > 0 {
					// Ablation (Melting)
					ablation := temp * IceAblationRate * dt
					is.Ice[idx].Thickness -= ablation
					if is.Ice[idx].Thickness < 0 {
						is.Ice[idx].Thickness = 0
					}
				}

				// Update age only if ice exists
				if is.Ice[idx].Thickness > 0 {
					is.Ice[idx].Age += dt
				} else {
					is.Ice[idx].Age = 0
				}
			}
		}(start, end)
	}
	wg.Wait()

	// Phase 2: Flow (SIA - ice flows downhill)
	// Must use double buffering to avoid race conditions or flow bias
	// We'll calculate flow out of each cell into a temporary buffer
	// Flow is local, but can't be purely parallel if we write to neighbors directly.
	// Approach: Calculate Flux OUT of each cell (read-only), store in buffer. Then apply.

	// Reuse pre-allocated delta buffer and zero it
	if len(is.deltaIce) != totalCells {
		is.deltaIce = make([]float64, totalCells)
	}
	for i := range is.deltaIce {
		is.deltaIce[i] = 0
	}

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Serial Flow for correctness (can be parallelized with atomic add)
	for idx := 0; idx < totalCells; idx++ {
		iceThick := is.Ice[idx].Thickness
		if iceThick < MinIceThickness {
			continue
		}

		coord := indexToCoord(idx)
		baseElev := heightmap.Get(coord) + iceThick

		bestIdx := -1
		lowestElev := baseElev

		for _, dir := range directions {
			n := topology.GetNeighbor(coord, dir)
			nIdx := coordToIndex(n)
			nElev := heightmap.Get(n) + is.Ice[nIdx].Thickness
			if nElev < lowestElev {
				lowestElev = nElev
				bestIdx = nIdx
			}
		}

		if bestIdx != -1 {
			slope := (baseElev - lowestElev) / 100000.0
			speed := IceFlowConstant * math.Pow(iceThick, 3) * slope * 1e12
			loss := math.Min(iceThick*0.5, speed*dt) /* Cap 50% */

			// Store flow speed for erosion
			is.Ice[idx].FlowSpeed = speed

			// Apply to buffer
			is.deltaIce[idx] -= loss
			is.deltaIce[bestIdx] += loss
		} else {
			is.Ice[idx].FlowSpeed = 0
		}
	}

	// Apply deltas
	for i := range is.Ice {
		is.Ice[i].Thickness += is.deltaIce[i]
		if is.Ice[i].Thickness < 0 {
			is.Ice[i].Thickness = 0
		}
	}

	// Update stats
	is.updateStats()
}

// ApplyErosion applies glacial erosion to the heightmap based on ice flow.
// Returns total erosion volume for tracking.
func (is *IceSheet) ApplyErosion(heightmap *SphereHeightmap, dt float64, resolution int) float64 {
	totalErosion := 0.0

	// Helper to get coordinate from index
	indexToCoord := func(idx int) spatial.Coordinate {
		resSq := resolution * resolution
		face := idx / resSq
		rem := idx % resSq
		y := rem / resolution
		x := rem % resolution
		return spatial.Coordinate{Face: face, X: x, Y: y}
	}

	for idx := range is.Ice {
		if is.Ice[idx].Thickness < MinIceThickness {
			continue
		}

		// Erosion rate = coefficient * ice flux (thickness * velocity)
		flux := is.Ice[idx].Thickness * is.Ice[idx].FlowSpeed
		erosionDepth := flux * ErosionCoefficient * dt / 1000.0 // Convert mm to m

		if erosionDepth > 0 {
			coord := indexToCoord(idx)
			currentElev := heightmap.Get(coord)
			newElev := currentElev - erosionDepth
			heightmap.Set(coord, newElev)
			totalErosion += erosionDepth

			// Accumulate sediment (for moraine deposition)
			is.Sediment[idx] += erosionDepth * SedimentCapacity
		}
	}

	return totalErosion
}

// DepositMoraines deposits accumulated sediment at ice margins.
// Called when ice retreats.
func (is *IceSheet) DepositMoraines(heightmap *SphereHeightmap, topology spatial.Topology, resolution int) {
	// Helper to get coordinate from index
	indexToCoord := func(idx int) spatial.Coordinate {
		resSq := resolution * resolution
		face := idx / resSq
		rem := idx % resSq
		y := rem / resolution
		x := rem % resolution
		return spatial.Coordinate{Face: face, X: x, Y: y}
	}

	for idx, sediment := range is.Sediment {
		iceThick := is.Ice[idx].Thickness

		// Deposit sediment where ice is thin or absent (margins)
		if iceThick < 10.0 {
			if sediment > 0.1 {
				// Raise terrain by depositing sediment
				coord := indexToCoord(idx)
				currentElev := heightmap.Get(coord)
				heightmap.Set(coord, currentElev+sediment*10.0) // Scale factor
				is.Sediment[idx] = 0
			}
		}
	}
}

// updateStats recalculates aggregate statistics.
func (is *IceSheet) updateStats() {
	is.TotalVolume = 0
	is.TotalArea = 0
	is.MaxThickness = 0

	cellAreaKm2 := 100.0 * 100.0 // Assume 100km grid cells for stats roughly

	for i := range is.Ice {
		thick := is.Ice[i].Thickness
		if thick >= MinIceThickness {
			is.TotalArea += cellAreaKm2
			is.TotalVolume += thick / 1000.0 * cellAreaKm2 // m to km * km²
			if thick > is.MaxThickness {
				is.MaxThickness = thick
			}
		}
	}
}

// GlacialFeatureType represents the type of glacial landform
type GlacialFeatureType string

const (
	FeatureUValley GlacialFeatureType = "u_valley"
	FeatureCirque  GlacialFeatureType = "cirque"
	FeatureFjord   GlacialFeatureType = "fjord"
	FeatureArete   GlacialFeatureType = "arete"
	FeatureMoraine GlacialFeatureType = "moraine"
)

// GlacialFeature represents a detected glacial landform
type GlacialFeature struct {
	Type     GlacialFeatureType
	Location spatial.Coordinate
	Size     float64 // Relative size/importance
}

// DetectGlacialFeatures identifies glacial landforms in the terrain.
// Call after ice retreat to identify U-valleys, cirques, and fjords.
func (is *IceSheet) DetectGlacialFeatures(heightmap *SphereHeightmap, topology spatial.Topology, seaLevel float64) []GlacialFeature {
	features := make([]GlacialFeature, 0)
	resolution := topology.Resolution()
	resSq := resolution * resolution

	indexToCoord := func(idx int) spatial.Coordinate {
		face := idx / resSq
		rem := idx % resSq
		y := rem / resolution
		x := rem % resolution
		return spatial.Coordinate{Face: face, X: x, Y: y}
	}

	for idx, sediment := range is.Sediment {
		coord := indexToCoord(idx)
		if sediment < 0.1 {
			continue
		}

		elev := heightmap.Get(coord)

		// Check for moraine deposits
		if sediment > 0.5 {
			features = append(features, GlacialFeature{
				Type:     FeatureMoraine,
				Location: coord,
				Size:     sediment,
			})
		}

		// Check for U-valley characteristics:
		// - Eroded (had ice)
		// - Neighbors are higher (valley walls)
		// - Below the ice extent
		neighborElevs := make([]float64, 0, 4)
		for _, dir := range []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West} {
			neighbor := topology.GetNeighbor(coord, dir)
			neighborElevs = append(neighborElevs, heightmap.Get(neighbor))
		}

		avgNeighbor := 0.0
		for _, ne := range neighborElevs {
			avgNeighbor += ne
		}
		avgNeighbor /= float64(len(neighborElevs))

		// U-valley: center is lower than neighbors (trough)
		if elev < avgNeighbor-100 && elev > seaLevel {
			features = append(features, GlacialFeature{
				Type:     FeatureUValley,
				Location: coord,
				Size:     avgNeighbor - elev,
			})
		}

		// Fjord: U-valley that's below sea level
		if elev < seaLevel && sediment > 0.3 {
			features = append(features, GlacialFeature{
				Type:     FeatureFjord,
				Location: coord,
				Size:     seaLevel - elev,
			})
		}
	}

	return features
}

// CreateGlacialLakes identifies locations where moraine dams may form lakes.
// Returns coordinates where lakes should form.
func (is *IceSheet) CreateGlacialLakes(heightmap *SphereHeightmap, topology spatial.Topology) []spatial.Coordinate {
	lakes := make([]spatial.Coordinate, 0)
	resolution := topology.Resolution()
	resSq := resolution * resolution

	indexToCoord := func(idx int) spatial.Coordinate {
		face := idx / resSq
		rem := idx % resSq
		y := rem / resolution
		x := rem % resolution
		return spatial.Coordinate{Face: face, X: x, Y: y}
	}

	for idx, sediment := range is.Sediment {
		coord := indexToCoord(idx)
		if sediment < 1.0 {
			continue // Need significant moraine deposit
		}

		// Check if this moraine is damming water (neighbors behind it are lower)
		elev := heightmap.Get(coord)
		dammed := false

		for _, dir := range []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West} {
			neighbor := topology.GetNeighbor(coord, dir)
			neighborElev := heightmap.Get(neighbor)

			// If a neighbor is significantly lower, it may be a lake bed
			if neighborElev < elev-50 {
				lakes = append(lakes, neighbor)
				dammed = true
			}
		}

		if dammed {
			// Mark the moraine location too
			_ = coord
		}
	}

	return lakes
}

// ApplyIsostaticRebound simulates post-glacial rebound after ice removal.
// The crust rises as the weight of ice is removed.
// Note: previousIce needs to be passed as Slice now if possible, or Map if coming from storage.
// Assuming map for now since "previous state" might be sparse?
// Actually let's assume slice for efficiency, but legacy might be problematic.
// Let's assume passed as map for compatibility with potential persistence?
// No, let's update signature to use slice.
// If previousIce represents the state from N years ago, keeping a full copy of 20M cells is expensive (~1GB).
// But for rebound, we just need `iceRemoved`.
// Let's keep it simple: Map is fine here if it's meant to be sparse differences.
func (is *IceSheet) ApplyIsostaticRebound(heightmap *SphereHeightmap, previousIce map[spatial.Coordinate]IceData, reboundRate float64, dt float64) {
	// previousIce uses value struct now
	for coord, oldIce := range previousIce {
		// Look up current ice
		resolution := is.Resolution
		resSq := resolution * resolution
		idx := coord.Face*resSq + coord.Y*resolution + coord.X

		currentThick := 0.0
		if idx >= 0 && idx < len(is.Ice) {
			currentThick = is.Ice[idx].Thickness
		}

		iceRemoved := oldIce.Thickness - currentThick

		if iceRemoved > 0 {
			// Rebound proportional to ice removed
			// Full rebound takes ~10,000 years
			rebound := iceRemoved * reboundRate * dt / 10000.0
			currentElev := heightmap.Get(coord)
			heightmap.Set(coord, currentElev+rebound)
		}
	}
}

// Helper: Convert flat index to coordinate
func iceIndexToCoord(idx, resolution int) spatial.Coordinate {
	resSq := resolution * resolution
	face := idx / resSq
	rem := idx % resSq
	y := rem / resolution
	x := rem % resolution
	return spatial.Coordinate{Face: face, X: x, Y: y}
}

// Helper: Convert coordinate to flat index
func iceCoordToIndex(coord spatial.Coordinate, resolution int) int {
	return coord.Face*resolution*resolution + coord.Y*resolution + coord.X
}
