package geography

import (
	"sync"
	"testing"
	"tw-backend/internal/spatial"
)

// TestActiveCellGrid tests the ActiveCellGrid functionality
func TestActiveCellGrid(t *testing.T) {
	// Setup
	resolution := 128
	bucketSize := 32
	grid := NewActiveCellGrid(resolution, bucketSize)

	// Test 1: Initialization
	if grid.resolution != resolution {
		t.Errorf("expected resolution %d, got %d", resolution, grid.resolution)
	}
	if grid.bucketSize != bucketSize {
		t.Errorf("expected bucketSize %d, got %d", bucketSize, grid.bucketSize)
	}

	// Test 2: MarkActive
	testCoord1 := spatial.Coordinate{Face: 0, X: 10, Y: 10} // Bucket 0,0
	testCoord2 := spatial.Coordinate{Face: 0, X: 40, Y: 10} // Bucket 1,0 (assuming 32 size)

	// Mark active
	grid.MarkActive(testCoord1)
	grid.MarkActive(testCoord2)

	// Verify bucket marking (implementation detail check)
	// Bucket Index for face 0, x=10 (0), y=10 (0) -> 0
	// 4 faces * (128/32)^2 ? No, bucket logic is simpler.
	// Let's rely on GetActiveCells to verify behavior rather than internal state if possible,
	// checking strictly public API behavior.

	// Test 3: GetActiveCells
	// Should return uniquely marked cells
	activeCells := grid.GetAllActiveCells()

	// Expect unique cells (2 unique coordinates added above)
	if len(activeCells) != 2 {
		t.Errorf("expected 2 active cells, got %d", len(activeCells))
	}

	// Verify clearing (double-buffer behavior)
	activeCells2 := grid.GetAllActiveCells()
	if len(activeCells2) != 0 {
		t.Errorf("expected grid to be cleared after get, got %d cells", len(activeCells2))
	}

	// Test 4: Concurrency
	var wg sync.WaitGroup // Use waitgroup from testing/sync
	// Actually we are in test package, need to import sync

	concurrencyCount := 100
	for i := 0; i < concurrencyCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c := spatial.Coordinate{Face: 0, X: idx, Y: idx}
			grid.MarkActive(c)
		}(i)
	}
	wg.Wait()

	activeCells3 := grid.GetAllActiveCells()
	if len(activeCells3) != concurrencyCount {
		t.Errorf("expected %d concurrently added cells, got %d", concurrencyCount, len(activeCells3))
	}

	// Test 5: Deduplication
	grid.MarkActive(testCoord1)
	grid.MarkActive(testCoord1) // Duplicate

	deduped := grid.GetAllActiveCells()
	if len(deduped) != 1 {
		t.Errorf("expected 1 deduped cell, got %d", len(deduped))
	}
}
