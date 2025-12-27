package underground

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoundaryGrid_CreatesEmptyGrid(t *testing.T) {
	boundaries := []TectonicBoundary{}
	grid := NewBoundaryGrid(boundaries, 10)

	assert.NotNil(t, grid)
	assert.Equal(t, 10, grid.ChunkSize)
	assert.NotNil(t, grid.Grid)
}

func TestBoundaryGrid_PopulatesChunks(t *testing.T) {
	boundaries := []TectonicBoundary{
		{X: 5, Y: 5, BoundaryType: "divergent", Intensity: 0.8},
		{X: 15, Y: 5, BoundaryType: "convergent", Intensity: 0.9},
		{X: 25, Y: 25, BoundaryType: "transform", Intensity: 0.5},
	}
	grid := NewBoundaryGrid(boundaries, 10)

	// Chunk (0,0) should contain boundary at (5,5)
	chunk00 := grid.Grid[GridKey{X: 0, Y: 0}]
	assert.Len(t, chunk00, 1)
	assert.Equal(t, 5, chunk00[0].X)

	// Chunk (1,0) should contain boundary at (15,5)
	chunk10 := grid.Grid[GridKey{X: 1, Y: 0}]
	assert.Len(t, chunk10, 1)
	assert.Equal(t, 15, chunk10[0].X)

	// Chunk (2,2) should contain boundary at (25,25)
	chunk22 := grid.Grid[GridKey{X: 2, Y: 2}]
	assert.Len(t, chunk22, 1)
	assert.Equal(t, 25, chunk22[0].X)
}

func TestBoundaryGrid_QueryBoundaries_ReturnsLocalAndNeighbors(t *testing.T) {
	// Create boundaries in a 3x3 grid of chunks
	boundaries := []TectonicBoundary{
		{X: 5, Y: 5, BoundaryType: "divergent", Intensity: 0.8},   // Chunk (0,0)
		{X: 15, Y: 5, BoundaryType: "convergent", Intensity: 0.9}, // Chunk (1,0)
		{X: 25, Y: 5, BoundaryType: "transform", Intensity: 0.5},  // Chunk (2,0)
		{X: 15, Y: 15, BoundaryType: "divergent", Intensity: 0.7}, // Chunk (1,1)
	}
	grid := NewBoundaryGrid(boundaries, 10)

	// Query at (15, 5) which is in chunk (1,0)
	// Should return boundaries from chunks (0,0), (1,0), (2,0), (0,1), (1,1), (2,1), etc.
	result := grid.QueryBoundaries(15, 5)

	// Should find at least the 4 boundaries (all are in the 3x3 neighborhood)
	assert.GreaterOrEqual(t, len(result), 3, "Should find boundaries in nearby chunks")

	// At minimum should include the boundary at (15, 5) itself
	found := false
	for _, b := range result {
		if b.X == 15 && b.Y == 5 {
			found = true
			break
		}
	}
	assert.True(t, found, "Should include boundary at query location")
}

func TestBoundaryGrid_QueryBoundaries_EmptyChunk(t *testing.T) {
	boundaries := []TectonicBoundary{
		{X: 5, Y: 5, BoundaryType: "divergent", Intensity: 0.8}, // Chunk (0,0)
	}
	grid := NewBoundaryGrid(boundaries, 10)

	// Query at (100, 100) - far from any boundary
	result := grid.QueryBoundaries(100, 100)

	assert.Empty(t, result, "Should return empty for chunks with no nearby boundaries")
}

func TestCompactChambers_RemovesSolidified(t *testing.T) {
	chambers := []*MagmaChamber{
		{Temperature: 1500, Solidified: false}, // Active
		{Temperature: 500, Solidified: true},   // Solidified - should be removed
		{Temperature: 1200, Solidified: false}, // Active
		{Temperature: 800, Solidified: true},   // Solidified - should be removed
	}

	active, solidified := CompactChambers(chambers)

	assert.Len(t, active, 2, "Should have 2 active chambers")
	assert.Len(t, solidified, 2, "Should have 2 solidified chambers")

	for _, c := range active {
		assert.False(t, c.Solidified, "Active slice should only contain non-solidified")
	}
	for _, c := range solidified {
		assert.True(t, c.Solidified, "Solidified slice should only contain solidified")
	}
}

func TestCompactChambers_EmptyInput(t *testing.T) {
	chambers := []*MagmaChamber{}

	active, solidified := CompactChambers(chambers)

	assert.Empty(t, active)
	assert.Empty(t, solidified)
}

func TestCompactChambers_AllActive(t *testing.T) {
	chambers := []*MagmaChamber{
		{Temperature: 1500, Solidified: false},
		{Temperature: 1400, Solidified: false},
	}

	active, solidified := CompactChambers(chambers)

	assert.Len(t, active, 2)
	assert.Empty(t, solidified)
}

func TestCompactChambers_AllSolidified(t *testing.T) {
	chambers := []*MagmaChamber{
		{Temperature: 500, Solidified: true},
		{Temperature: 400, Solidified: true},
	}

	active, solidified := CompactChambers(chambers)

	assert.Empty(t, active)
	assert.Len(t, solidified, 2)
}
