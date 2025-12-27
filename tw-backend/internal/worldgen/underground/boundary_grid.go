package underground

// GridKey identifies a spatial chunk for O(1) boundary lookup
type GridKey struct {
	X, Y int
}

// BoundaryGrid provides O(1) spatial lookup for tectonic boundaries
// Instead of checking every boundary for every chamber (O(N*M)),
// we can now check only nearby chunks (O(N * ~9))
type BoundaryGrid struct {
	ChunkSize int                            // Grid cell size in map units
	Grid      map[GridKey][]TectonicBoundary // Chunk -> boundaries in that chunk
}

// NewBoundaryGrid creates a spatial index from a boundary list
// ChunkSize determines the granularity of the grid (typically 10-50 map units)
func NewBoundaryGrid(boundaries []TectonicBoundary, chunkSize int) *BoundaryGrid {
	bg := &BoundaryGrid{
		ChunkSize: chunkSize,
		Grid:      make(map[GridKey][]TectonicBoundary),
	}

	for _, b := range boundaries {
		key := GridKey{
			X: b.X / chunkSize,
			Y: b.Y / chunkSize,
		}
		bg.Grid[key] = append(bg.Grid[key], b)
	}

	return bg
}

// QueryBoundaries returns all boundaries in the chamber's chunk and 8 neighbors
// This reduces O(N) boundary iteration to O(1) spatial lookup
func (bg *BoundaryGrid) QueryBoundaries(x, y int) []TectonicBoundary {
	centerX := x / bg.ChunkSize
	centerY := y / bg.ChunkSize

	var result []TectonicBoundary

	// Check 3x3 neighborhood of chunks
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			key := GridKey{X: centerX + dx, Y: centerY + dy}
			if boundaries, ok := bg.Grid[key]; ok {
				result = append(result, boundaries...)
			}
		}
	}

	return result
}

// CompactChambers separates active chambers from solidified ones
// Returns (active, solidified) where active are still simulated and
// solidified can be archived to InactiveMagmaDeposits
func CompactChambers(chambers []*MagmaChamber) (active, solidified []*MagmaChamber) {
	active = make([]*MagmaChamber, 0, len(chambers)/2)
	solidified = make([]*MagmaChamber, 0, len(chambers)/2)

	for _, c := range chambers {
		if c.Solidified {
			solidified = append(solidified, c)
		} else {
			active = append(active, c)
		}
	}

	return active, solidified
}
