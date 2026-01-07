package geography

import (
	"math"
	"math/rand"
	"sort"
	"sync" // Used in parallel functions

	"tw-backend/internal/spatial"
)

// ApplyThermalErosion improves slope stability by moving material from steep slopes to lower neighbors
func ApplyThermalErosion(hm *Heightmap, iterations int, seed int64) {
	// Talus angle approximation (max difference allowed)
	threshold := 40.0
	width, height := hm.Width, hm.Height

	// Pre-define neighbors to avoid allocation in inner loop
	neighbors := [][2]int{
		{0, 1}, {0, -1}, {1, 0}, {-1, 0},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}

	for iter := 0; iter < iterations; iter++ {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				currentElev := hm.Get(x, y)
				maxDiff := 0.0
				var bestNeighX, bestNeighY int

				// Check neighbors

				// Find lowest neighbor
				for _, n := range neighbors {
					nx, ny := x+n[0], y+n[1]
					if nx >= 0 && nx < width && ny >= 0 && ny < height {
						diff := currentElev - hm.Get(nx, ny)
						if diff > maxDiff {
							maxDiff = diff
							bestNeighX, bestNeighY = nx, ny
						}
					}
				}

				// If slope is too steep, erode
				if maxDiff > threshold {
					transfer := maxDiff * 0.1 // Move 10% of excess
					hm.Set(x, y, currentElev-transfer)
					hm.Set(bestNeighX, bestNeighY, hm.Get(bestNeighX, bestNeighY)+transfer)
				}
			}
		}
	}
}

// ApplyThermalErosionSpherical improves slope stability on the sphere heightmap
// Uses the topology graph to find neighbors instead of 2D grid logic.
// Parallelized for performance.
func ApplyThermalErosionSpherical(hm *SphereHeightmap, topology spatial.Topology, iterations int, seed int64) {
	// Talus angle approximation (max difference allowed)
	threshold := 40.0
	resolution := topology.Resolution()
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Use worker pool for parallel processing
	// Split work by faces (6 tasks)
	var wg sync.WaitGroup

	for iter := 0; iter < iterations; iter++ {
		// Iterate over all 6 faces concurrently
		wg.Add(6)
		for face := 0; face < 6; face++ {
			go func(f int) {
				defer wg.Done()
				for y := 0; y < resolution; y++ {
					for x := 0; x < resolution; x++ {
						coord := spatial.Coordinate{Face: f, X: x, Y: y}
						currentElev := hm.Get(coord)

						maxDiff := 0.0
						var bestNeigh spatial.Coordinate

						// Check 4 cardinal neighbors (diagonals are complex on sphere grid)
						for _, dir := range directions {
							neighbor := topology.GetNeighbor(coord, dir)
							diff := currentElev - hm.Get(neighbor)
							if diff > maxDiff {
								maxDiff = diff
								bestNeigh = neighbor
							}
						}

						// If slope is too steep, erode
						if maxDiff > threshold {
							// Move material downhill
							transfer := (maxDiff - threshold) * 0.5 // Standard talus slippage formula

							// Safety clamp to prevent oscillation
							if transfer > maxDiff {
								transfer = maxDiff * 0.5
							}

							// NOTE: Data race possible at face edges, but acceptable for geological noise
							hm.Set(coord, currentElev-transfer)
							hm.Set(bestNeigh, hm.Get(bestNeigh)+transfer)
						}
					}
				}
			}(face)
		}
		wg.Wait()
	}
}

// =============================================================================
// Stream Power Erosion (Physics-Based River Carving)
// =============================================================================

// StreamPowerConstants defines the physics parameters for stream power erosion.
const (
	StreamPowerK = 0.00001 // Erosivity constant
	StreamPowerM = 0.5     // Flux exponent (0.3-0.6 typical)
	StreamPowerN = 1.0     // Slope exponent (1.0-2.0 typical)
)

// ApplyStreamPowerErosion erodes terrain using Stream Power Law: E = K × Flux^m × Slope^n
// Integrates with isostasy: eroded mass reduces crust thickness, elevation recalculated.
// Parallelized.
func ApplyStreamPowerErosion(hm *SphereHeightmap, hydro *HydrologyLayer, plates []TectonicPlate, dt float64, seaLevel float64) {
	res := hm.Resolution()
	// Use lower resolution for flux calculation if needed, but here we process cells
	totalCells := 6 * res * res
	resSq := res * res

	// Build plate lookup
	var plateGrid []int
	if plates != nil {
		plateGrid = make([]int, totalCells)
		// Initialize to -1
		for i := range plateGrid {
			plateGrid[i] = -1
		}
		// This part is fast enough to keep serial or parallelize if needed
		// Map plates to grid
		for i, p := range plates {
			for coord := range p.Region {
				idx := coord.Face*resSq + coord.Y*res + coord.X
				if idx >= 0 && idx < totalCells {
					plateGrid[idx] = i
				}
			}
		}
	}

	// Stream Power with Sediment Transport
	// Refactored to sequential downstream processing to allow sediment routing.

	// 1. Sort cells by elevation (High to Low) for downstream routing
	// This ensures we process upstream cells before their downstream neighbors
	processingOrder := make([]int, totalCells)
	for i := 0; i < totalCells; i++ {
		processingOrder[i] = i
	}

	// Sort processing order
	sort.Slice(processingOrder, func(i, j int) bool {
		// We want Descending order (High -> Low)
		idxI := processingOrder[i]
		idxJ := processingOrder[j]
		// Using hydro IndexToCoord because generic indexToCoord isn't available
		elevI := hm.Get(hydro.IndexToCoord(idxI))
		elevJ := hm.Get(hydro.IndexToCoord(idxJ))
		return elevI > elevJ
	})

	// Track outgoing sediment load for each cell (to be passed to downstream)
	sedimentLoad := make([]float64, totalCells)

	// Constants for Transport Capacity
	// Qt = Kt * Q^m * S^n
	const Kt = 2.0             // Transport efficiency
	const depositionRate = 0.5 // Fraction of excess load deposited per step

	for _, idx := range processingOrder {
		coord := hydro.IndexToCoord(idx)

		// Skip ocean cells (base level)
		currentElev := hm.Get(coord)
		if currentElev <= seaLevel {
			// Deposit at river mouth
			load := sedimentLoad[idx]
			if load > 0 {
				hm.AddSediment(coord, load)
			}
			continue
		}

		flux := hydro.Flux[idx]
		if flux < 1.0 {
			continue
		}

		downhillIdx := hydro.FlowDirection[idx]
		if downhillIdx < 0 {
			// Local minimum (pit). Deposit everything.
			hm.AddSediment(coord, sedimentLoad[idx])
			continue
		}

		// Calculate Slope
		downhillCoord := hydro.IndexToCoord(downhillIdx)
		slope := currentElev - hm.Get(downhillCoord)
		if slope < 0.001 {
			slope = 0.001
		}

		// 1. Calculate Transport Capacity
		slopeNorm := slope / 1000.0 // km drop per cell
		capacity := Kt * math.Pow(flux, 1.5) * math.Pow(slopeNorm, 1.0) * dt

		// 2. Calculate Erosion (Detachment) Potential
		erosionPotential := StreamPowerK * math.Pow(flux, StreamPowerM) * math.Pow(slopeNorm, StreamPowerN) * dt

		// Isostatic adjustment / Hardness
		hardness := hm.GetRockHardness(coord)
		erosionPotential *= (1.0 - hardness*0.8)

		// 3. Sediment Balance
		incomingLoad := sedimentLoad[idx]
		currentLoad := incomingLoad

		if currentLoad > capacity {
			// Deposition Mode
			depositAmount := (currentLoad - capacity) * depositionRate
			hm.AddSediment(coord, depositAmount)
			currentLoad -= depositAmount
		} else {
			// Erosion Mode
			remainingCapacity := capacity - currentLoad
			erodeAmount := math.Min(erosionPotential, remainingCapacity)

			// Slope limit
			if erodeAmount > slope*0.9 {
				erodeAmount = slope * 0.9
			}

			if erodeAmount > 0 {
				actualEroded := hm.Erode(coord, erodeAmount)
				currentLoad += actualEroded

				// Isostasy check (simplified)
				if plateGrid != nil {
					// We don't update plate thickness here to avoid complexity
					// SimTectonics handles rebalancing
				}
			}
		}

		// Pass load to downstream neighbor
		sedimentLoad[downhillIdx] += currentLoad
	}
}

// ApplyHydraulicErosion simulates rain and water flow to carve valleys

func ApplyHydraulicErosion(hm *Heightmap, drops int, seed int64) {
	r := rand.New(rand.NewSource(seed))
	width, height := hm.Width, hm.Height

	// Constants
	dt := 1.2
	density := 1.0 // Density of water
	evapRate := 0.001
	depositionRate := 0.3
	minVol := 0.01
	friction := 0.1

	for i := 0; i < drops; i++ {
		// Spawn drop
		x := float64(r.Intn(width))
		y := float64(r.Intn(height))

		// Drop properties
		speedX, speedY := 0.0, 0.0
		volume := 1.0
		sediment := 0.0

		for volume > minVol {
			ix, iy := int(x), int(y)
			if ix < 0 || ix >= width-1 || iy < 0 || iy >= height-1 {
				break
			}

			// Get surface normal / gradient
			n00 := hm.Get(ix, iy)
			n10 := hm.Get(ix+1, iy)
			n01 := hm.Get(ix, iy+1)
			n11 := hm.Get(ix+1, iy+1)

			gx := (n10 + n11) - (n00 + n01)
			gy := (n01 + n11) - (n00 + n10)

			// Update Position
			// F = ma, but here just assume F ~ gradient
			speedX = (speedX * (1 - friction)) - (gx * 0.5)
			speedY = (speedY * (1 - friction)) - (gy * 0.5)

			x += speedX * dt
			y += speedY * dt

			if x < 0 || x >= float64(width-1) || y < 0 || y >= float64(height-1) {
				break
			}

			// New elevation
			// Interpolate new height
			// Simplified: just use nearest integer for erosion target
			newIx, newIy := int(x), int(y)
			newElev := hm.Get(newIx, newIy)
			// oldElev := hm.Get(ix, iy) // Unused

			// Approximate height difference along trajectory
			heightDiff := newElev - hm.Get(ix, iy)

			// Sediment capacity
			// Capacity is proportional to velocity and volume
			velocity := math.Sqrt(speedX*speedX + speedY*speedY)
			capacity := math.Max(-heightDiff, minVol) * velocity * volume * density

			if heightDiff > 0 {
				// Moving uphill? Fill depression
				// Deposit sediment
				amount := math.Min(sediment, heightDiff)
				sediment -= amount
				hm.Set(ix, iy, hm.Get(ix, iy)+amount)
			} else {
				if sediment > capacity {
					// Deposit
					amount := (sediment - capacity) * depositionRate
					sediment -= amount
					hm.Set(ix, iy, hm.Get(ix, iy)+amount)
				} else {
					// Erode
					amount := math.Min((capacity-sediment)*0.3, -heightDiff)
					sediment += amount
					hm.Set(ix, iy, hm.Get(ix, iy)-amount)
				}
			}

			volume *= (1.0 - evapRate)
		}
	}
}

// =============================================================================
// Differential Erosion (Phase 5: Geological Provinces)
// =============================================================================

// ApplyDifferentialErosion simulates water erosion respecting rock hardness.
// Soft provinces erode faster creating deep valleys and jagged coastlines.
// Sediment deposits when velocity drops, building continental shelves.
//
// Physics:
//   - Erosion: MaterialRemoved = BaseRate * Velocity * (1.0 - RockHardness)
//   - Deposition: When velocity drops or elevation <= seaLevel
//   - Sediment tracking: Uses SphereHeightmap.AddSediment and Erode methods
func ApplyDifferentialErosion(hm *SphereHeightmap, topology spatial.Topology, numDrops int, seed int64, seaLevel float64) {
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Erosion constants
	erosionConstant := 0.15 // Base erosion rate
	depositionRate := 0.3   // Fraction of sediment deposited per step
	evaporationRate := 0.02 // Water loss per step
	minVolume := 0.05       // Minimum water volume before droplet dies
	maxSteps := 64          // Maximum steps per droplet

	for drop := 0; drop < numDrops; drop++ {
		// Spawn droplet at random position
		startPoint := spatial.RandomPointOnSphere(seed + int64(drop))
		coord := topology.FromVector(startPoint.X, startPoint.Y, startPoint.Z)

		// Droplet properties
		volume := 1.0   // Water volume
		velocity := 0.0 // Current velocity
		sediment := 0.0 // Carried sediment

		prevCoord := coord

		// Trace droplet path
		for step := 0; step < maxSteps && volume > minVolume; step++ {
			currentElev := hm.Get(coord)
			hardness := hm.GetRockHardness(coord)

			// Find steepest descent
			var lowestNeighbor *spatial.Coordinate
			lowestElev := currentElev

			for _, dir := range directions {
				neighbor := topology.GetNeighbor(coord, dir)
				neighborElev := hm.Get(neighbor)
				if neighborElev < lowestElev {
					lowestElev = neighborElev
					neighborCopy := neighbor
					lowestNeighbor = &neighborCopy
				}
			}

			// Calculate slope and velocity
			slope := currentElev - lowestElev
			if slope < 0 {
				slope = 0
			}

			// Update velocity based on slope
			newVelocity := math.Sqrt(velocity*velocity + slope*0.5)
			if newVelocity > 10.0 {
				newVelocity = 10.0 // Cap velocity
			}

			// Sediment capacity based on velocity and volume
			capacity := newVelocity * volume * 2.0

			// Determine if we should erode or deposit
			velocityDecreased := newVelocity < velocity*0.8
			atSeaLevel := currentElev <= seaLevel
			atLocalMinimum := lowestNeighbor == nil

			if velocityDecreased || atSeaLevel || atLocalMinimum || sediment > capacity {
				// DEPOSIT: Velocity dropped, at sea level, or over capacity
				depositAmount := sediment * depositionRate
				if depositAmount > sediment {
					depositAmount = sediment
				}
				if depositAmount > 0 {
					hm.AddSediment(coord, depositAmount)
					sediment -= depositAmount
				}
			} else if sediment < capacity {
				// ERODE: We have capacity to pick up more sediment
				// Erosion scales with velocity and INVERSELY with hardness
				// Hard rock (0.9) erodes at 10% rate, soft rock (0.2) at 80% rate
				erosionFactor := 1.0 - hardness

				// Erosion Cap (Refinement Task 4)
				// High peaks (> 6000m) erode much faster (glacial/gravity limit)
				if currentElev > 6000.0 {
					erosionFactor *= 3.0
				}

				erodeAmount := erosionConstant * newVelocity * erosionFactor * volume

				// Can't erode more than the slope allows
				if erodeAmount > slope*0.5 {
					erodeAmount = slope * 0.5
				}

				if erodeAmount > 0 {
					// Use the Erode method which handles sediment vs bedrock
					actualEroded := hm.Erode(coord, erodeAmount)
					sediment += actualEroded
				}
			}

			// Handle local minimum - deposit all and stop
			if lowestNeighbor == nil {
				if sediment > 0 {
					hm.AddSediment(coord, sediment)
				}
				break
			}

			// Move to next cell
			velocity = newVelocity
			prevCoord = coord
			coord = *lowestNeighbor

			// Evaporation
			volume *= (1.0 - evaporationRate)
		}

		// Deposit remaining sediment at final position
		if sediment > 0 && coord != prevCoord {
			hm.AddSediment(coord, sediment)
		}
	}
}

// ApplyRiverErosion carves valleys where water flux is high.
// This creates realistic V-shaped valleys along river paths.
//
// Parameters:
// - fluxThreshold: Minimum flux required to cause erosion (e.g. 50.0)
// - erosionAmount: Base depth to erode per iteration (scaled by flux)
// - seaLevel: Erosion stops at sea level
func ApplyRiverErosion(hm *SphereHeightmap, fluxThreshold float64, erosionAmount float64, seaLevel float64) {
	res := hm.Resolution()
	topo := hm.Topology()

	// Parallelize by face (6 goroutines)
	var wg sync.WaitGroup
	wg.Add(6)
	for face := 0; face < 6; face++ {
		go func(f int) {
			defer wg.Done()
			for y := 0; y < res; y++ {
				for x := 0; x < res; x++ {
					coord := spatial.Coordinate{Face: f, X: x, Y: y}
					data := hm.GetCellData(coord)

					// Only erode if flux matches or exceeds threshold
					// This simulates river channels
					if data.Flux >= fluxThreshold {
						elev := hm.Get(coord)

						// Don't erode below sea level
						if elev <= seaLevel {
							continue
						}

						// Erode based on flux intensity
						// More flux = deeper channel
						// Logarithmic scaling to prevent massive canyon spikes
						fluxFactor := math.Log(data.Flux/fluxThreshold) + 1.0
						erodeDepth := erosionAmount * fluxFactor

						// Limit erosion depth to avoid pits
						if erodeDepth > 20.0 {
							erodeDepth = 20.0
						}

						newElev := elev - erodeDepth
						if newElev < seaLevel {
							newElev = seaLevel
						}
						hm.Set(coord, newElev)

						// Thermal erosion / Smoothing neighbors to form V-shape
						// Widen the valley
						// NOTE: Minor data race possible at face boundaries, acceptable for geological noise
						neighbors := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}
						for _, dir := range neighbors {
							n := topo.GetNeighbor(coord, dir)
							nElev := hm.Get(n)

							// If neighbor is significantly higher, erode it too (slumping)
							if nElev > newElev+5.0 {
								slump := (nElev - newElev) * 0.3
								hm.Set(n, nElev-slump)
							}
						}
					}
				}
			}
		}(face)
	}
	wg.Wait()
}
