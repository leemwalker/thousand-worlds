package geography

import (
	"container/heap"
	"tw-backend/internal/spatial"
)

// Lake represents a generated body of water
type Lake struct {
	ID            int
	SurfaceHeight float64
	Cells         []spatial.Coordinate
	Outlet        spatial.Coordinate
}

// IdentifySinks finds all local minima (cells with no lower neighbors).
// Ignores sinks below seaLevel (assumed to be ocean).
func IdentifySinks(hm *SphereHeightmap, seaLevel float64) []spatial.Coordinate {
	res := hm.Resolution()
	topo := hm.Topology()
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}
	sinks := []spatial.Coordinate{}

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				c := spatial.Coordinate{Face: face, X: x, Y: y}
				elev := hm.Get(c)

				// Skip ocean sinks
				if elev < seaLevel {
					continue
				}

				isSink := true

				for _, dir := range directions {
					n := topo.GetNeighbor(c, dir)
					if hm.Get(n) <= elev {
						isSink = false
						break
					}
				}

				if isSink {
					sinks = append(sinks, c)
				}
			}
		}
	}
	return sinks
}

// FillDepressions identifies lakes by filling depressions from sinks upwards
// until a spillover point is reached.
// It modifies the heightmap to flatten lake surfaces and marks IsLake in CellData.
func FillDepressions(hm *SphereHeightmap, seaLevel float64) []*Lake {
	sinks := IdentifySinks(hm, seaLevel)
	lakes := []*Lake{}
	lakeIDCounter := 1

	// Visited set for this pass to avoid reprocessing same lake from multiple local minima
	globalVisited := make(map[spatial.Coordinate]bool)

	for _, sink := range sinks {
		if globalVisited[sink] {
			continue // Already processed this basin
		}

		// Try to fill this depression
		lake, cells := fillBasin(hm, sink, globalVisited)
		if lake != nil {
			lake.ID = lakeIDCounter
			lakeIDCounter++
			lakes = append(lakes, lake)

			// Mark cells
			for _, c := range cells {
				originalHeight := hm.Get(c)
				depth := 0.0

				// Flatten surface to lake height and calculate depth
				if originalHeight < lake.SurfaceHeight {
					depth = lake.SurfaceHeight - originalHeight
					hm.Set(c, lake.SurfaceHeight)
				}

				data := hm.GetCellData(c)
				data.IsLake = true
				data.LakeID = lake.ID
				data.LakeDepth = depth
				hm.SetCellData(c, data)
			}
		}
	}
	return lakes
}

// RouteFluxThroughLakes updates the HydrologyLayer to route accumulated flux through lakes.
// For each lake, sums the inflow flux from all tributary cells and assigns it to the outlet.
// This ensures rivers continue from lake outlets to the ocean.
func RouteFluxThroughLakes(hm *SphereHeightmap, hydro *HydrologyLayer, lakes []*Lake) {
	res := hm.Resolution()
	resSq := res * res

	for _, lake := range lakes {
		if len(lake.Cells) == 0 {
			continue
		}

		// Calculate total influx from cells that flow INTO the lake
		totalInflux := 0.0

		for _, lakeCoord := range lake.Cells {
			lakeIdx := lakeCoord.Face*resSq + lakeCoord.Y*res + lakeCoord.X
			if lakeIdx >= 0 && lakeIdx < len(hydro.Flux) {
				totalInflux += hydro.Flux[lakeIdx]
			}
		}

		// Assign total flux to outlet
		outletIdx := lake.Outlet.Face*resSq + lake.Outlet.Y*res + lake.Outlet.X
		if outletIdx >= 0 && outletIdx < len(hydro.Flux) {
			hydro.Flux[outletIdx] = totalInflux

			// Update flow direction to point downstream from outlet
			// (The outlet should already have a flow direction from CalculateFlowField)
		}
	}
}

// Priority Queue for flood filling
type Item struct {
	coord spatial.Coordinate
	elev  float64
	index int
}
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].elev < pq[j].elev } // Min-heap
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].index = i; pq[j].index = j }
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func fillBasin(hm *SphereHeightmap, start spatial.Coordinate, globalVisited map[spatial.Coordinate]bool) (*Lake, []spatial.Coordinate) {
	topo := hm.Topology()
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Priority Queue stores boundary of current lake
	pq := &PriorityQueue{}
	heap.Init(pq)

	initialElev := hm.Get(start)
	heap.Push(pq, &Item{coord: start, elev: initialElev})

	// Correct algorithm:
	// 1. Maintain a Region (set of cells).
	// 2. Maintain a Boundary (min-heap).
	// 3. Start with Sink in Region. Add Neighbors to Boundary.
	// 4. Pop LOWEST cell from Boundary (candidate spillover).
	// 5. If Lowest Boundary < Region Max Height.... wait.

	waterLevel := initialElev
	region := []spatial.Coordinate{}

	// Reset
	heap.Init(pq)
	visited := make(map[spatial.Coordinate]bool)
	visited[start] = true

	// Push initial neighbors to boundary
	for _, dir := range directions {
		n := topo.GetNeighbor(start, dir)
		if !visited[n] {
			visited[n] = true
			heap.Push(pq, &Item{coord: n, elev: hm.Get(n)})
		}
	}

	region = append(region, start)

	const MaxLakeSize = 1000 // Safety break
	lakeSize := 1
	var outlet spatial.Coordinate
	hasOutlet := false

	for pq.Len() > 0 {
		minEdge := heap.Pop(pq).(*Item)

		if minEdge.elev < waterLevel {
			// Found a way down! Spillover.
			// The water runs out here.
			outlet = minEdge.coord
			hasOutlet = true
			break
		}

		// If edge is higher or equal, expanding lake means raising water level
		waterLevel = minEdge.elev

		// This cell becomes part of the lake (or rather, the basin is filled UP TO this cell)
		// Wait, if minEdge.elev is the new water level, this cell is the rim.
		// If we include it, we keep expanding.

		// Add new neighbors
		region = append(region, minEdge.coord)
		lakeSize++
		if lakeSize > MaxLakeSize {
			return nil, nil // Too big (ocean?)
		}

		for _, dir := range directions {
			n := topo.GetNeighbor(minEdge.coord, dir)
			if !visited[n] {
				visited[n] = true
				heap.Push(pq, &Item{coord: n, elev: hm.Get(n)})
			}
		}
	}

	if !hasOutlet {
		// No outlet found (e.g. hit max size or map edge?), discard
		return nil, nil
	}

	// Finalize:
	// All cells in 'region' are submerged up to 'waterLevel'.
	// Only include those strictly below waterLevel? Or equal?
	// Usually water surface is constant. All cells in 'region' are effectively "under" that level
	// except the last one which is the spillover.

	// Filter global visited
	for _, c := range region {
		globalVisited[c] = true
	}

	return &Lake{SurfaceHeight: waterLevel, Outlet: outlet}, region
}
