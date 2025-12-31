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
// RouteFluxThroughLakes updates the SphereHeightmap to route accumulated flux through lakes.
// For each lake, sums the inflow flux from all tributary cells and assigns it to the outlet.
// This ensures rivers continue from lake outlets to the ocean.
func RouteFluxThroughLakes(hm *SphereHeightmap, lakes []*Lake) {
	if len(lakes) == 0 {
		return
	}

	for _, lake := range lakes {
		if len(lake.Cells) == 0 {
			continue
		}

		// Calculate total influx from cells that flow INTO the lake
		// Since we don't have a flow graph here easily, we can just sum the Flux of all lake cells.
		// Wait, Flux is "passing through".
		// In CalculateGlobalFlux, flux accumulates downhill.
		// At a sink, flux is trapped.
		// So checking the Flux of every cell in the lake is correct?
		// No, we only want the Flux entering the lake.
		// Use the Flux of the Sink cell(s)?
		// The sink cell has accumulated flux from its basin.
		// But a lake might have multiple local minima if huge?
		// IdentifySinks finds local minima. FillDepressions merges them?
		// FillDepressions uses "start" which is a sink.
		// So we can just take the flux of the lake cells.
		// But Flux accumulates.
		// If A flows to B (Lake). B has Flux(A) + Rainfall.
		// So if we sum all lake cells, we might double count?
		// No, flow stops at the sink.
		// So only the Sink cell needs to be read?
		// What if the lake covers multiple former sinks?
		// We should sum flux of all cells in the lake that are LOCAL MINIMA (sinks).
		// Or simpler: Any cell where water stops.
		// Given CalculateGlobalFlux implementation:
		// "If no lower neighbor (Sink), flux stays here".
		// So yes, sum flux of all "Sink" cells within the lake.
		// But `data.Flux` might be updated.
		// Let's just sum Flux of all cells in Lake.Cells?
		// If a lake cell A flows to B (also in lake).
		// CalculatingGlobalFlux:
		// A finds B lower?
		// Since Lake is flat, FillDepressions flattens it.
		// But CalculateGlobalFlux runs BEFORE FillDepressions usually?
		// In `geology.go`, it runs BEFORE. So water flows to deep points.
		// So yes, the Sinks have the flux. Non-sinks flow to sinks.
		// So we loop lake.Cells, find which ones were sinks?
		// Or just iterate all lake.Cells, and if `Flux > 1.0` (rain), add it?
		// Safer: Just sum everything.
		// Flux is conserved. If A->B, A passes flux to B. A is reset?
		// No, `neighborData.Flux += currentData.Flux`. A keeps its flux in `CalculateGlobalFlux`?
		// Let's check `CalculateGlobalFlux` in `hydrology.go` line 271.
		// It adds to neighbor. It does NOT clear current.
		// So `Flux` variable is "Volume passing through".
		// So if A->B->C.
		// A=1. B=2. C=3.
		// If {A,B,C} are all in lake.
		// Sum = 6. Incorrect. Total is 3.
		// We only want the Flux at the "End" of the flow chains within the lake.
		// Since CalculateGlobalFlux leaves flux in Sinks.
		// And all flow ends in Sinks.
		// We just need to sum Flux of "Sink" cells inside the lake.
		// How to identify Sinks?
		// `IdentifySinks` did it.
		// But we don't have that list here easily.
		// We can scan lake.Cells and check neighbors?
		// Or better: `FillDepressions` logic ensures the lake is a basin.
		// The flux is effectively caught in the bottom.
		// Just find the max flux in the lake cells?
		// If multiple prongs, max might miss one.
		// Summing maxes of independent branches is hard.

		// Alternative: Iterate all lake cells. If flow direction points NOWHERE (no lower neighbor), add flux.
		// We can check `CalculateGlobalFlux` logic again: "If no lower neighbor... flux stays here".
		// So we reuse that check.

		totalInflux := 0.0
		topology := hm.Topology()
		directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

		for _, cell := range lake.Cells {
			elev := hm.Get(cell)
			isSink := true
			for _, dir := range directions {
				n := topology.GetNeighbor(cell, dir)
				if hm.Get(n) < elev {
					isSink = false
					break
				}
			}

			if isSink {
				// This cell trapped flux.
				data := hm.GetCellData(cell)
				totalInflux += data.Flux
			}
		}

		// Assign total flux to outlet
		outletData := hm.GetCellData(lake.Outlet)
		outletData.Flux += totalInflux
		hm.SetCellData(lake.Outlet, outletData)
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
