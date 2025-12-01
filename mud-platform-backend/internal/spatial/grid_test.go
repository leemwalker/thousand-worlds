package spatial

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewSpatialGrid(t *testing.T) {
	grid := NewSpatialGrid(100.0)
	assert.NotNil(t, grid)
	assert.Equal(t, 100.0, grid.cellSize)
	assert.Equal(t, 0, grid.Count())
}

func TestNewSpatialGrid_DefaultCellSize(t *testing.T) {
	grid := NewSpatialGrid(0)
	assert.Equal(t, 100.0, grid.cellSize) // Default
}

func TestSpatialGrid_InsertAndRemove(t *testing.T) {
	grid := NewSpatialGrid(100.0)

	entityID := uuid.New()
	pos := Position{X: 50, Y: 50}

	// Insert
	grid.Insert(entityID, pos)
	assert.Equal(t, 1, grid.Count())

	// Verify position
	retrievedPos, exists := grid.GetPosition(entityID)
	assert.True(t, exists)
	assert.Equal(t, pos, retrievedPos)

	// Remove
	grid.Remove(entityID)
	assert.Equal(t, 0, grid.Count())

	_, exists = grid.GetPosition(entityID)
	assert.False(t, exists)
}

func TestSpatialGrid_UpdatePosition(t *testing.T) {
	grid := NewSpatialGrid(100.0)

	entityID := uuid.New()
	pos1 := Position{X: 50, Y: 50}
	pos2 := Position{X: 250, Y: 250} // Different cell

	// Insert at first position
	grid.Insert(entityID, pos1)
	assert.Equal(t, 1, grid.Count())

	// Update to second position
	grid.Insert(entityID, pos2)
	assert.Equal(t, 1, grid.Count()) // Still only 1 entity

	retrievedPos, exists := grid.GetPosition(entityID)
	assert.True(t, exists)
	assert.Equal(t, pos2, retrievedPos)
}

func TestSpatialGrid_QueryRadius(t *testing.T) {
	grid := NewSpatialGrid(100.0)

	// Insert entities at various positions
	entities := map[uuid.UUID]Position{
		uuid.New(): {X: 0, Y: 0},     // At center
		uuid.New(): {X: 50, Y: 0},    // 50m away
		uuid.New(): {X: 0, Y: 100},   // 100m away
		uuid.New(): {X: 200, Y: 200}, // Far away (~283m)
	}

	for id, pos := range entities {
		grid.Insert(id, pos)
	}

	// Query 60m radius from origin
	results := grid.QueryRadius(Position{X: 0, Y: 0}, 60.0)

	// Should return 2 entities (at 0,0 and 50,0)
	assert.Len(t, results, 2)

	// Query 150m radius
	results = grid.QueryRadius(Position{X: 0, Y: 0}, 150.0)

	// Should return 3 entities (0,0 + 50,0 + 0,100)
	assert.Len(t, results, 3)
}

func TestSpatialGrid_QueryArea(t *testing.T) {
	grid := NewSpatialGrid(100.0)

	entities := []struct {
		id  uuid.UUID
		pos Position
	}{
		{uuid.New(), Position{X: 50, Y: 50}},
		{uuid.New(), Position{X: 150, Y: 150}},
		{uuid.New(), Position{X: 250, Y: 250}},
		{uuid.New(), Position{X: -50, Y: -50}},
	}

	for _, e := range entities {
		grid.Insert(e.id, e.pos)
	}

	// Query area covering first two entities
	results := grid.QueryArea(0, 0, 200, 200)
	assert.Len(t, results, 2)

	// Query area covering all positive quadrant
	results = grid.QueryArea(0, 0, 300, 300)
	assert.Len(t, results, 3)

	// Query area covering negative quadrant
	results = grid.QueryArea(-100, -100, 0, 0)
	assert.Len(t, results, 1)
}

func TestSpatialGrid_PositionToCell(t *testing.T) {
	grid := NewSpatialGrid(100.0)

	tests := []struct {
		pos      Position
		expected GridCell
	}{
		{Position{X: 0, Y: 0}, GridCell{X: 0, Y: 0}},
		{Position{X: 50, Y: 50}, GridCell{X: 0, Y: 0}},
		{Position{X: 100, Y: 100}, GridCell{X: 1, Y: 1}},
		{Position{X: 150, Y: 250}, GridCell{X: 1, Y: 2}},
		{Position{X: -50, Y: -50}, GridCell{X: -1, Y: -1}},
		{Position{X: 99.9, Y: 99.9}, GridCell{X: 0, Y: 0}},
	}

	for _, tt := range tests {
		cell := grid.positionToCell(tt.pos)
		assert.Equal(t, tt.expected, cell, "Position %v", tt.pos)
	}
}

func TestSpatialGrid_EmptyQuery(t *testing.T) {
	grid := NewSpatialGrid(100.0)

	// Query on empty grid
	results := grid.QueryRadius(Position{X: 0, Y: 0}, 100.0)
	assert.Empty(t, results)

	results = grid.QueryArea(0, 0, 100, 100)
	assert.Empty(t, results)
}

func TestSpatialGrid_ConcurrentOperations(t *testing.T) {
	grid := NewSpatialGrid(100.0)

	const numEntities = 100

	// Concurrent inserts
	done := make(chan bool)
	for i := 0; i < numEntities; i++ {
		go func(idx int) {
			entityID := uuid.New()
			pos := Position{
				X: float64(idx * 10),
				Y: float64(idx * 10),
			}
			grid.Insert(entityID, pos)
			done <- true
		}(i)
	}

	// Wait for all inserts
	for i := 0; i < numEntities; i++ {
		<-done
	}

	assert.Equal(t, numEntities, grid.Count())
}

// Benchmark tests to verify O(k) vs O(N) performance

func BenchmarkSpatialGrid_Insert(b *testing.B) {
	grid := NewSpatialGrid(100.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := uuid.New()
		pos := Position{
			X: float64(i % 1000),
			Y: float64(i / 1000),
		}
		grid.Insert(entityID, pos)
	}
}

func BenchmarkSpatialGrid_QueryRadius_SmallArea(b *testing.B) {
	grid := NewSpatialGrid(100.0)

	// Populate grid with 10,000 entities
	for i := 0; i < 10000; i++ {
		entityID := uuid.New()
		pos := Position{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		}
		grid.Insert(entityID, pos)
	}

	center := Position{X: 500, Y: 500}
	radius := 150.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = grid.QueryRadius(center, radius)
	}
}

func BenchmarkSpatialGrid_QueryRadius_LargeArea(b *testing.B) {
	grid := NewSpatialGrid(100.0)

	// Populate grid
	for i := 0; i < 10000; i++ {
		entityID := uuid.New()
		pos := Position{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		}
		grid.Insert(entityID, pos)
	}

	center := Position{X: 500, Y: 500}
	radius := 500.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = grid.QueryRadius(center, radius)
	}
}
