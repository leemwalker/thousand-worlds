package spatial

import (
	"math"
	"sync"

	"github.com/google/uuid"
)

// Position represents a 2D coordinate in the game world
type Position struct {
	X, Y float64
}

// GridCell represents a cell in the spatial grid
type GridCell struct {
	X, Y int
}

// SpatialGrid implements a grid-based spatial index for fast area queries
// Divides the world into fixed-size cells (default 100m x 100m)
// Performance: O(k) where k = entities in nearby cells vs O(N) for all entities
type SpatialGrid struct {
	cellSize float64
	// Map of grid cell to entity IDs in that cell
	cells map[GridCell]map[uuid.UUID]Position
	// Map of entity ID to current grid cell
	entityToCell map[uuid.UUID]GridCell
	mu           sync.RWMutex
}

// NewSpatialGrid creates a new spatial grid with specified cell size in meters
func NewSpatialGrid(cellSize float64) *SpatialGrid {
	if cellSize <= 0 {
		cellSize = 100.0 // Default 100m cells
	}

	return &SpatialGrid{
		cellSize:     cellSize,
		cells:        make(map[GridCell]map[uuid.UUID]Position),
		entityToCell: make(map[uuid.UUID]GridCell),
	}
}

// positionToCell converts a world position to a grid cell
// Complexity: O(1)
func (g *SpatialGrid) positionToCell(pos Position) GridCell {
	return GridCell{
		X: int(math.Floor(pos.X / g.cellSize)),
		Y: int(math.Floor(pos.Y / g.cellSize)),
	}
}

// Insert adds or updates an entity's position in the grid
// Complexity: O(1) - constant time hash map operations
func (g *SpatialGrid) Insert(entityID uuid.UUID, pos Position) {
	g.mu.Lock()
	defer g.mu.Unlock()

	newCell := g.positionToCell(pos)

	// Remove from old cell if entity already exists
	if oldCell, exists := g.entityToCell[entityID]; exists {
		if oldCell != newCell {
			delete(g.cells[oldCell], entityID)
			if len(g.cells[oldCell]) == 0 {
				delete(g.cells, oldCell)
			}
		}
	}

	// Add to new cell
	if g.cells[newCell] == nil {
		g.cells[newCell] = make(map[uuid.UUID]Position)
	}
	g.cells[newCell][entityID] = pos
	g.entityToCell[entityID] = newCell
}

// Remove removes an entity from the grid
// Complexity: O(1)
func (g *SpatialGrid) Remove(entityID uuid.UUID) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if cell, exists := g.entityToCell[entityID]; exists {
		delete(g.cells[cell], entityID)
		if len(g.cells[cell]) == 0 {
			delete(g.cells, cell)
		}
		delete(g.entityToCell, entityID)
	}
}

// QueryRadius returns all entities within a radius of the given position
// Complexity: O(k) where k = entities in nearby cells (typically ~9 cells)
// vs O(N) for checking all entities
func (g *SpatialGrid) QueryRadius(center Position, radius float64) []uuid.UUID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Calculate which cells to check based on radius
	centerCell := g.positionToCell(center)
	cellRadius := int(math.Ceil(radius / g.cellSize))

	results := make([]uuid.UUID, 0)
	radiusSquared := radius * radius

	// Check cells in a square around the center cell
	for dx := -cellRadius; dx <= cellRadius; dx++ {
		for dy := -cellRadius; dy <= cellRadius; dy++ {
			cell := GridCell{
				X: centerCell.X + dx,
				Y: centerCell.Y + dy,
			}

			if entities, exists := g.cells[cell]; exists {
				// Check each entity in this cell
				for entityID, pos := range entities {
					distSquared := g.distanceSquared(center, pos)
					if distSquared <= radiusSquared {
						results = append(results, entityID)
					}
				}
			}
		}
	}

	return results
}

// QueryArea returns all entities in a rectangular area
// Complexity: O(k) where k = entities in cells intersecting the area
func (g *SpatialGrid) QueryArea(minX, minY, maxX, maxY float64) []uuid.UUID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	minCell := g.positionToCell(Position{X: minX, Y: minY})
	maxCell := g.positionToCell(Position{X: maxX, Y: maxY})

	results := make([]uuid.UUID, 0)

	for x := minCell.X; x <= maxCell.X; x++ {
		for y := minCell.Y; y <= maxCell.Y; y++ {
			cell := GridCell{X: x, Y: y}
			if entities, exists := g.cells[cell]; exists {
				for entityID, pos := range entities {
					// Verify entity is actually within the area bounds
					if pos.X >= minX && pos.X <= maxX && pos.Y >= minY && pos.Y <= maxY {
						results = append(results, entityID)
					}
				}
			}
		}
	}

	return results
}

// GetPosition returns the position of an entity if it exists in the grid
func (g *SpatialGrid) GetPosition(entityID uuid.UUID) (Position, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	cell, exists := g.entityToCell[entityID]
	if !exists {
		return Position{}, false
	}

	if entities, ok := g.cells[cell]; ok {
		if pos, found := entities[entityID]; found {
			return pos, true
		}
	}

	return Position{}, false
}

// Count returns the total number of entities in the grid
func (g *SpatialGrid) Count() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.entityToCell)
}

// distanceSquared calculates squared Euclidean distance
// Avoids sqrt() for performance (we compare against radius squared)
func (g *SpatialGrid) distanceSquared(a, b Position) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

// GetCellCount returns the number of active cells (for debugging/metrics)
func (g *SpatialGrid) GetCellCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.cells)
}
