package geography

import (
	"tw-backend/internal/spatial"
)

// HierarchicalPlateGrid reduces memory for plate lookups by using a sparse sparse/dense hybrid approach.
// Instead of a flat array of potentially millions of cells (1.6GB at 8k res),
// it uses a coarse Level 0 grid where each cell represents a large block (e.g., 64x64 or 32x32).
// - If a Level 0 block is entirely within one plate, it stores the PlateID.
// - If a Level 0 block contains a boundary (Mixed), it stores a sentinel (-1).
// Lookups hit Level 0 first. If -1, they fall back to a precise check (checking the Plate.Region maps).
type HierarchicalPlateGrid struct {
	// Level0 is a coarse grid for each of the 6 faces.
	// Dimensions: [Face][BlockY][BlockX]
	// Block size is relative to the full resolution.
	Level0     [][][]int8
	Resolution int // Full world resolution
	BlockSize  int // Dimension of a Level 0 block (e.g. 64)

	// Reference to plates for fallback lookups
	Plates []TectonicPlate
}

// NewHierarchicalPlateGrid creates a new grid for fast, low-memory plate lookups.
// blockSize determines the granularity of the coarse grid. 32 or 64 are good defaults.
func NewHierarchicalPlateGrid(plates []TectonicPlate, resolution int, blockSize int) *HierarchicalPlateGrid {
	if blockSize <= 0 {
		blockSize = 64
	}

	// Calculate grid dimensions
	blocksPerDim := (resolution + blockSize - 1) / blockSize

	grid := &HierarchicalPlateGrid{
		Level0:     make([][][]int8, 6),
		Resolution: resolution,
		BlockSize:  blockSize,
		Plates:     plates,
	}

	// Initialize Level 0 grids
	for f := 0; f < 6; f++ {
		grid.Level0[f] = make([][]int8, blocksPerDim)
		for y := 0; y < blocksPerDim; y++ {
			grid.Level0[f][y] = make([]int8, blocksPerDim)
			// Initialize to "Unknown" (-2) or "Mixed" (-1)?
			// Let's use -2 for Empty/Unassigned (ocean/default) and -1 for Mixed.
			// Actually, plates usually cover everything. Let's start with -2 (Uninitialized).
			for x := 0; x < blocksPerDim; x++ {
				grid.Level0[f][y][x] = -2
			}
		}
	}

	grid.build(plates)
	return grid
}

func (h *HierarchicalPlateGrid) build(plates []TectonicPlate) {
	// 1. Iterate over all plates and their regions
	// We want to fill the Level0 grid.
	// Since identifying "pure" blocks requires verifying ALL cells in the block,
	// an efficient way is to iterate plates and "paint" the blocks.
	// If a block gets painted by more than one plate, it becomes "Mixed" (-1).

	// Optimization: This build step could be slow if we iterate every cell of every plate.
	// But we only run this once or rarely.
	// Faster approach:
	// Iterate through every BLOCK. For each block, sample density or check boundaries.
	// Even better: The Plate struct has a map of cells.
	// We can iterate the map of cells for each plate.

	for plateIdx, plate := range plates {
		pid := int8(plateIdx)

		for coord := range plate.Region {
			// Determine Block Coordinates
			bx := coord.X / h.BlockSize
			by := coord.Y / h.BlockSize

			// Boundary check (safety)
			if bx < 0 || by < 0 || bx >= len(h.Level0[0][0]) || by >= len(h.Level0[0]) {
				continue
			}

			currentVal := h.Level0[coord.Face][by][bx]

			if currentVal == -2 {
				// Block was empty, now assigned to this plate
				h.Level0[coord.Face][by][bx] = pid
			} else if currentVal != -1 && currentVal != pid {
				// Block was assigned to another plate, now we define it as Mixed
				h.Level0[coord.Face][by][bx] = -1 // Mixed
			}
			// If already Mixed (-1) or already this plate (pid), do nothing.
		}
	}

	// After iterating all active plate cells, we might have -2 blocks (void? deep ocean?).
	// If the world is fully covered by plates (which it should be), -2 shouldn't happen much
	// unless gaps exist. Treat -2 as Mixed (Fallback) to be safe.
}

// GetPlateID returns the index of the plate at the given coordinate.
// Returns -1 if not found.
func (h *HierarchicalPlateGrid) GetPlateID(coord spatial.Coordinate) int {
	bx := coord.X / h.BlockSize
	by := coord.Y / h.BlockSize

	// Bounds check
	if coord.Face < 0 || coord.Face >= 6 {
		return -1
	}
	if by < 0 || by >= len(h.Level0[0]) {
		return -1
	}
	if bx < 0 || bx >= len(h.Level0[0][0]) {
		return -1
	}

	val := h.Level0[coord.Face][by][bx]

	// Fast path: Homogeneous block
	if val >= 0 {
		return int(val)
	}

	// Slow path: Mixed block (-1) or Uninitialized (-2)
	// Check exact plate regions
	// This is O(NumPlates) * O(1) map lookup.
	// Number of plates is small (~10-20).
	// So worst case is fine for only 10% of cells.
	for i, p := range h.Plates {
		if _, ok := p.Region[coord]; ok {
			return i
		}
	}

	return -1
}
