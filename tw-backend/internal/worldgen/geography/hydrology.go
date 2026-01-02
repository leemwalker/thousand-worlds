package geography

import (
	"runtime"
	"sort"
	"sync"
	"tw-backend/internal/spatial"
)

// =============================================================================
// HydrologyLayer: Flow Direction and Flux Accumulation
// =============================================================================

// HydrologyLayer stores the computed flow field for a heightmap.
// Used for deterministic river routing and stream power erosion.
type HydrologyLayer struct {
	FlowDirection []int     // Index of downhill neighbor (-1 = Sink)
	Flux          []float64 // Total water volume passing through each cell
	Resolution    int       // Grid resolution for coordinate conversion
}

// CalculateFlowField computes flow directions and flux accumulation for the entire sphere.
// Algorithm:
//  1. Steepest Descent: For each cell, find the neighbor with lowest elevation
//  2. Topological Sort: Sort cells by elevation (highest to lowest)
//  3. Accumulation: Push flux from each cell to its downhill neighbor
//
// If rainfall is nil, uses uniform 1.0 (backward compatibility).
// Otherwise, Flux[i] is initialized from rainfall[i].
// Returns a HydrologyLayer with FlowDirection and accumulated Flux.
func CalculateFlowField(hm *SphereHeightmap, rainfall []float64) *HydrologyLayer {
	topology := hm.Topology()
	res := hm.Resolution()
	totalCells := 6 * res * res
	resSq := res * res

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	hydro := &HydrologyLayer{
		FlowDirection: make([]int, totalCells),
		Flux:          make([]float64, totalCells),
		Resolution:    res,
	}

	// Initialize all to sink (-1) and base rainfall
	useRainfall := rainfall != nil && len(rainfall) == totalCells
	for i := range hydro.FlowDirection {
		hydro.FlowDirection[i] = -1
		if useRainfall {
			hydro.Flux[i] = rainfall[i]
		} else {
			hydro.Flux[i] = 1.0 // Uniform rainfall fallback
		}
	}

	// Step 1: Calculate flow directions (steepest descent)
	// Also build sorted list for topological processing
	type cellNode struct {
		idx  int
		elev float64
	}
	nodes := make([]cellNode, totalCells)

	// Worker pool for parallel flow calculation
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
				// Reconstruct coordinate from index
				face := idx / resSq
				rem := idx % resSq
				y := rem / res
				x := rem % res
				coord := spatial.Coordinate{Face: face, X: x, Y: y}

				currentElev := hm.Get(coord)
				nodes[idx] = cellNode{idx: idx, elev: currentElev}

				// Find steepest descent neighbor
				lowestElev := currentElev
				lowestIdx := -1

				for _, dir := range directions {
					neighbor := topology.GetNeighbor(coord, dir)
					neighborElev := hm.Get(neighbor)

					if neighborElev < lowestElev {
						lowestElev = neighborElev
						lowestIdx = neighbor.Face*resSq + neighbor.Y*res + neighbor.X
					}
				}

				hydro.FlowDirection[idx] = lowestIdx
			}
		}(start, end)
	}
	wg.Wait()

	// Step 2: Sort cells by elevation (highest to lowest)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].elev > nodes[j].elev
	})

	// Step 3: Accumulate flux (process highest cells first)
	// This ensures upstream flux is added before the cell is processed
	for _, node := range nodes {
		idx := node.idx
		downhillIdx := hydro.FlowDirection[idx]

		// Pass flux to downhill neighbor
		if downhillIdx >= 0 && downhillIdx < totalCells {
			hydro.Flux[downhillIdx] += hydro.Flux[idx]
		}
		// If downhillIdx == -1, this is a sink (lake/ocean), flux stays here
	}

	return hydro
}

// CalculateFlowFieldWithRainfall computes flow with spatially-variable rainfall.
// Use GenerateRainfallMap from weather package to create the rainfall input.
// DEPRECATED: Use CalculateFlowField(hm, rainfall) directly instead.
func CalculateFlowFieldWithRainfall(hm *SphereHeightmap, rainfall []float64) *HydrologyLayer {
	return CalculateFlowField(hm, rainfall)
}

// CoordToIndex converts a spherical coordinate to a flat index
func (h *HydrologyLayer) CoordToIndex(coord spatial.Coordinate) int {
	resSq := h.Resolution * h.Resolution
	return coord.Face*resSq + coord.Y*h.Resolution + coord.X
}

// IndexToCoord converts a flat index to a spherical coordinate
func (h *HydrologyLayer) IndexToCoord(idx int) spatial.Coordinate {
	resSq := h.Resolution * h.Resolution
	face := idx / resSq
	rem := idx % resSq
	y := rem / h.Resolution
	x := rem % h.Resolution
	return spatial.Coordinate{Face: face, X: x, Y: y}
}

// GetFlux returns the flux at a coordinate
func (h *HydrologyLayer) GetFlux(coord spatial.Coordinate) float64 {
	idx := h.CoordToIndex(coord)
	if idx >= 0 && idx < len(h.Flux) {
		return h.Flux[idx]
	}
	return 0
}

// IsSink returns true if the cell has no lower neighbor
func (h *HydrologyLayer) IsSink(coord spatial.Coordinate) bool {
	idx := h.CoordToIndex(coord)
	if idx >= 0 && idx < len(h.FlowDirection) {
		return h.FlowDirection[idx] == -1
	}
	return true
}

// CalculateGlobalFluxWithRainfall computes flow accumulation using spatially-variable rainfall.
// It uses the rainfall array to initialize Flux, then accumulates downhill.
// If rainfall is nil or wrong length, falls back to uniform 1.0.
func CalculateGlobalFluxWithRainfall(hm *SphereHeightmap, rainfall []float64) {
	topology := hm.Topology()
	res := hm.Resolution()
	totalCells := 6 * res * res
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	useRainfall := rainfall != nil && len(rainfall) == totalCells

	// 1. Initialize Flux from rainfall array
	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				idx := face*res*res + y*res + x
				if useRainfall {
					hm.cellData[face][y*res+x].Flux = rainfall[idx]
				} else {
					hm.cellData[face][y*res+x].Flux = 1.0
				}
			}
		}
	}

	// 2. Create a list of all cells to sort by elevation
	type cellNode struct {
		coord spatial.Coordinate
		elev  float64
	}

	nodes := make([]cellNode, 0, totalCells)

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				c := spatial.Coordinate{Face: face, X: x, Y: y}
				nodes = append(nodes, cellNode{
					coord: c,
					elev:  hm.Get(c),
				})
			}
		}
	}

	// 3. Sort by Elevation (Highest to Lowest)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].elev > nodes[j].elev
	})

	// 4. Distribute Flux
	for _, node := range nodes {
		currentCoord := node.coord
		currentElev := node.elev
		currentData := hm.GetCellData(currentCoord)

		// Find lowest neighbor
		var lowestNeighbor *spatial.Coordinate
		lowestElev := currentElev

		for _, dir := range directions {
			n := topology.GetNeighbor(currentCoord, dir)
			nElev := hm.Get(n)

			if nElev < lowestElev {
				lowestElev = nElev
				nCopy := n
				lowestNeighbor = &nCopy
			}
		}

		// If a lower neighbor exists, pass flux to it
		if lowestNeighbor != nil {
			neighborData := hm.GetCellData(*lowestNeighbor)
			neighborData.Flux += currentData.Flux
			hm.SetCellData(*lowestNeighbor, neighborData)
		}
	}
}

// CalculateGlobalFlux computes flow accumulation for every cell.
// It simulates water flowing downhill from rainfall.
// Flux represents the volume of water passing through a cell.
func CalculateGlobalFlux(hm *SphereHeightmap) {
	topology := hm.Topology()
	res := hm.Resolution()
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// 1. Initialize Flux with base rainfall (currently uniform 1.0)
	// TODO: Use actual rainfall map from weather system later
	for face := 0; face < 6; face++ {
		for i := range hm.cellData[face] {
			hm.cellData[face][i].Flux = 1.0
		}
	}

	// 2. Create a list of all cells to sort by elevation
	// Struct to hold coordinate and elevation for sorting
	type cellNode struct {
		coord spatial.Coordinate
		elev  float64
	}

	totalCells := 6 * res * res
	nodes := make([]cellNode, 0, totalCells)

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				c := spatial.Coordinate{Face: face, X: x, Y: y}
				nodes = append(nodes, cellNode{
					coord: c,
					elev:  hm.Get(c),
				})
			}
		}
	}

	// 3. Sort by Elevation (Highest to Lowest)
	// Processing highest cells first ensures upstream flux is pushed downstream correctly
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].elev > nodes[j].elev
	})

	// 4. Distribute Flux
	for _, node := range nodes {
		currentCoord := node.coord
		currentElev := node.elev
		currentData := hm.GetCellData(currentCoord)

		// Find lowest neighbor
		var lowestNeighbor *spatial.Coordinate
		lowestElev := currentElev

		for _, dir := range directions {
			n := topology.GetNeighbor(currentCoord, dir)
			nElev := hm.Get(n)

			// Must be strictly lower to flow
			if nElev < lowestElev {
				lowestElev = nElev
				nCopy := n // copy
				lowestNeighbor = &nCopy
			}
		}

		// If a lower neighbor exists, pass flux to it
		if lowestNeighbor != nil {
			neighborData := hm.GetCellData(*lowestNeighbor)
			neighborData.Flux += currentData.Flux
			hm.SetCellData(*lowestNeighbor, neighborData)
		}
		// If no lower neighbor (Sink), flux stays here (Lake candidate)
	}
}

// ApplyRiverErosion lowers the heightmap along paths of high flux (rivers),
// creating V-shaped valleys.
// It respects sea level (won't erode below it).
