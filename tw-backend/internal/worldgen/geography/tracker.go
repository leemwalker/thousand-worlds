package geography

import (
	"sync"
	"tw-backend/internal/spatial"
)

// ActiveCellGrid tracks regions requiring simulation updates (e.g., erosion, decay).
// It uses a spatial hashing approach to group active cells into buckets for efficient concurrent access.
type ActiveCellGrid struct {
	buckets    map[int][]spatial.Coordinate
	bucketSize int // Approximate dimension of a spatial bucket (e.g., 32 for 32x32 chunks)
	resolution int // World resolution (e.g., 128, 512, etc.)
	mu         sync.RWMutex
}

// NewActiveCellGrid creates a new tracker.
func NewActiveCellGrid(resolution, bucketSize int) *ActiveCellGrid {
	if bucketSize <= 0 {
		bucketSize = 32
	}
	return &ActiveCellGrid{
		buckets:    make(map[int][]spatial.Coordinate),
		bucketSize: bucketSize,
		resolution: resolution,
	}
}

// hash computes a specialized spatial hash key for a coordinate.
// It groups cells into bucketSize x bucketSize chunks per face.
func (g *ActiveCellGrid) hash(coord spatial.Coordinate) int {
	// Simple spatial hash: Face | ChunkY | ChunkX
	// Face is 0-5.
	// Map X,Y to ChunkX, ChunkY
	cx := coord.X / g.bucketSize
	cy := coord.Y / g.bucketSize

	// Max chunks per dimension
	chunksPerDim := (g.resolution + g.bucketSize - 1) / g.bucketSize

	// Key: Face * (chunksPerDim^2) + cy * chunksPerDim + cx
	return coord.Face*chunksPerDim*chunksPerDim + cy*chunksPerDim + cx
}

// MarkActive marks a cell as requiring processing in the next tick.
// It is safe for concurrent use.
func (g *ActiveCellGrid) MarkActive(coord spatial.Coordinate) {
	key := g.hash(coord)
	g.mu.Lock()
	// Check if this coordinate is arguably already in the bucket?
	// For performance, we might just append and dedup later,
	// or perform a simple check if the bucket is small.
	// For now, simple append. Ideally, we use a set or a dedup pass.
	// To avoid heavy sets, we'll dedup on retrieval.
	g.buckets[key] = append(g.buckets[key], coord)
	g.mu.Unlock()
}

// GetAllActiveCells returns a deduplicated list of all active coordinates.
// This clears the current grid (double-buffering style usage pattern).
func (g *ActiveCellGrid) GetAllActiveCells() []spatial.Coordinate {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Dedup map
	set := make(map[spatial.Coordinate]struct{})

	for _, coords := range g.buckets {
		for _, c := range coords {
			set[c] = struct{}{}
		}
	}

	// Clear buckets
	g.buckets = make(map[int][]spatial.Coordinate)

	// Convert to slice
	result := make([]spatial.Coordinate, 0, len(set))
	for c := range set {
		result = append(result, c)
	}

	return result
}

// Count returns the number of active buckets (approximate activity metric).
func (g *ActiveCellGrid) Count() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.buckets)
}

// InitializeFromHeightmap performs a full scan to find unstable cells.
// conditionFunc should return true if the cell is unstable.
func (g *ActiveCellGrid) InitializeFullScan(resolution int, conditionFunc func(spatial.Coordinate) bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				if conditionFunc(coord) {
					key := g.hash(coord)
					g.buckets[key] = append(g.buckets[key], coord)
				}
			}
		}
	}
}
