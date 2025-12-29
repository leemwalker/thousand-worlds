package geography

import (
	"sort"
	"tw-backend/internal/spatial"
)

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
