// Package geography provides hex grid and tectonic plate systems for geographic isolation.
// This enables realistic continental configuration, species isolation, and regional effects.
package geography

import (
	"math"

	"github.com/google/uuid"
)

// HexCoord represents axial coordinates for a hexagonal grid
// Using axial coordinates (q, r) which simplifies many hex operations
// See: https://www.redblobgames.com/grids/hexagons/ for reference
type HexCoord struct {
	Q int `json:"q"` // Column
	R int `json:"r"` // Row
}

// HexDirection represents the 6 directions from a hex
type HexDirection int

const (
	HexDirE  HexDirection = 0 // East (+q)
	HexDirNE HexDirection = 1 // Northeast (+q, -r)
	HexDirNW HexDirection = 2 // Northwest (-r)
	HexDirW  HexDirection = 3 // West (-q)
	HexDirSW HexDirection = 4 // Southwest (-q, +r)
	HexDirSE HexDirection = 5 // Southeast (+r)
)

// Direction offsets for axial coordinates
var hexDirectionOffsets = []HexCoord{
	{Q: 1, R: 0},  // E
	{Q: 1, R: -1}, // NE
	{Q: 0, R: -1}, // NW
	{Q: -1, R: 0}, // W
	{Q: -1, R: 1}, // SW
	{Q: 0, R: 1},  // SE
}

// NewHexCoord creates a new hex coordinate
func NewHexCoord(q, r int) HexCoord {
	return HexCoord{Q: q, R: r}
}

// Add returns the sum of two hex coordinates
func (h HexCoord) Add(other HexCoord) HexCoord {
	return HexCoord{Q: h.Q + other.Q, R: h.R + other.R}
}

// Subtract returns the difference of two hex coordinates
func (h HexCoord) Subtract(other HexCoord) HexCoord {
	return HexCoord{Q: h.Q - other.Q, R: h.R - other.R}
}

// Scale multiplies the coordinate by a scalar
func (h HexCoord) Scale(factor int) HexCoord {
	return HexCoord{Q: h.Q * factor, R: h.R * factor}
}

// Neighbor returns the hex coordinate in the given direction
func (h HexCoord) Neighbor(dir HexDirection) HexCoord {
	return h.Add(hexDirectionOffsets[dir])
}

// AllNeighbors returns all 6 neighboring hex coordinates
func (h HexCoord) AllNeighbors() []HexCoord {
	neighbors := make([]HexCoord, 6)
	for i := 0; i < 6; i++ {
		neighbors[i] = h.Neighbor(HexDirection(i))
	}
	return neighbors
}

// S returns the implicit third cube coordinate (q + r + s = 0)
func (h HexCoord) S() int {
	return -h.Q - h.R
}

// Distance returns the hex distance between two coordinates
func (h HexCoord) Distance(other HexCoord) int {
	diff := h.Subtract(other)
	return (abs(diff.Q) + abs(diff.R) + abs(diff.S())) / 2
}

// Ring returns all hex coordinates at exactly the given distance
func (h HexCoord) Ring(radius int) []HexCoord {
	if radius == 0 {
		return []HexCoord{h}
	}

	results := make([]HexCoord, 0, 6*radius)
	// Start at the southwest corner
	current := h.Add(hexDirectionOffsets[HexDirSW].Scale(radius))

	for dir := 0; dir < 6; dir++ {
		for step := 0; step < radius; step++ {
			results = append(results, current)
			current = current.Neighbor(HexDirection(dir))
		}
	}
	return results
}

// Spiral returns all hex coordinates within the given radius (inclusive)
func (h HexCoord) Spiral(radius int) []HexCoord {
	results := []HexCoord{h}
	for r := 1; r <= radius; r++ {
		results = append(results, h.Ring(r)...)
	}
	return results
}

// ToPixel converts hex coordinates to pixel/world coordinates
// Uses pointy-top hex orientation with size = distance from center to corner
func (h HexCoord) ToPixel(hexSize float64) (x, y float64) {
	x = hexSize * (math.Sqrt(3)*float64(h.Q) + math.Sqrt(3)/2*float64(h.R))
	y = hexSize * (3.0 / 2 * float64(h.R))
	return x, y
}

// FromPixel converts pixel/world coordinates to the nearest hex coordinate
func FromPixel(x, y, hexSize float64) HexCoord {
	q := (math.Sqrt(3)/3*x - 1.0/3*y) / hexSize
	r := (2.0 / 3 * y) / hexSize
	return roundHex(q, r)
}

// roundHex rounds fractional cube coordinates to nearest hex
func roundHex(q, r float64) HexCoord {
	s := -q - r

	rq := math.Round(q)
	rr := math.Round(r)
	rs := math.Round(s)

	qDiff := math.Abs(rq - q)
	rDiff := math.Abs(rr - r)
	sDiff := math.Abs(rs - s)

	if qDiff > rDiff && qDiff > sDiff {
		rq = -rr - rs
	} else if rDiff > sDiff {
		rr = -rq - rs
	}
	// else rs = -rq - rr (not needed for axial coords)

	return HexCoord{Q: int(rq), R: int(rr)}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Scale multiplies direction offset by a factor
func (o HexCoord) ScaleDir(factor int) HexCoord {
	return HexCoord{Q: o.Q * factor, R: o.R * factor}
}

// --- HexCell represents a single cell on the world hex grid ---

// TerrainType represents the type of terrain in a hex cell
type TerrainType string

const (
	TerrainOcean      TerrainType = "ocean"
	TerrainShallows   TerrainType = "shallows"
	TerrainCoast      TerrainType = "coast"
	TerrainPlains     TerrainType = "plains"
	TerrainHills      TerrainType = "hills"
	TerrainMountain   TerrainType = "mountain"
	TerrainVolcanic   TerrainType = "volcanic"
	TerrainRift       TerrainType = "rift"       // Tectonic rift zone
	TerrainSubduction TerrainType = "subduction" // Subduction zone
)

// HexCell represents a single hexagonal cell in the world grid
type HexCell struct {
	ID          uuid.UUID   `json:"id"`
	Coord       HexCoord    `json:"coord"`
	PlateID     uuid.UUID   `json:"plate_id"`  // Tectonic plate this cell belongs to
	RegionID    *uuid.UUID  `json:"region_id"` // Ecological region (nil if unassigned)
	BiomeID     *uuid.UUID  `json:"biome_id"`  // Biome type
	Terrain     TerrainType `json:"terrain"`
	Elevation   float32     `json:"elevation"`   // -1.0 (deep ocean) to 1.0 (high mountain)
	Temperature float32     `json:"temperature"` // Annual average, in standard units
	Moisture    float32     `json:"moisture"`    // 0.0 (desert) to 1.0 (rainforest)
	IsLand      bool        `json:"is_land"`     // Convenience flag
}

// NewHexCell creates a new hex cell
func NewHexCell(coord HexCoord, terrain TerrainType, elevation float32) *HexCell {
	return &HexCell{
		ID:        uuid.New(),
		Coord:     coord,
		Terrain:   terrain,
		Elevation: elevation,
		IsLand:    terrain != TerrainOcean && terrain != TerrainShallows,
	}
}

// IsPassable returns true if organisms can traverse this cell
func (c *HexCell) IsPassable() bool {
	switch c.Terrain {
	case TerrainOcean:
		return false // Deep ocean blocks land animals
	case TerrainMountain:
		return false // Mountains are barriers
	case TerrainVolcanic:
		return false // Active volcanic zones are impassable
	default:
		return true
	}
}

// IsWater returns true if this cell is water
func (c *HexCell) IsWater() bool {
	return c.Terrain == TerrainOcean || c.Terrain == TerrainShallows
}

// --- HexGrid manages the collection of hex cells ---

// HexGrid represents the complete world hex grid
type HexGrid struct {
	WorldID uuid.UUID             `json:"world_id"`
	Cells   map[HexCoord]*HexCell `json:"-"`        // Keyed by coord for O(1) lookup
	Width   int                   `json:"width"`    // Grid width in hexes
	Height  int                   `json:"height"`   // Grid height in hexes
	HexSize float64               `json:"hex_size"` // Size for coordinate conversion
}

// NewHexGrid creates a new hex grid
func NewHexGrid(worldID uuid.UUID, width, height int, hexSize float64) *HexGrid {
	return &HexGrid{
		WorldID: worldID,
		Cells:   make(map[HexCoord]*HexCell),
		Width:   width,
		Height:  height,
		HexSize: hexSize,
	}
}

// GetCell returns the cell at the given coordinate, or nil if not found
func (g *HexGrid) GetCell(coord HexCoord) *HexCell {
	return g.Cells[coord]
}

// SetCell sets the cell at the given coordinate
func (g *HexGrid) SetCell(cell *HexCell) {
	g.Cells[cell.Coord] = cell
}

// GetNeighbors returns all neighboring cells that exist in the grid
func (g *HexGrid) GetNeighbors(coord HexCoord) []*HexCell {
	neighbors := make([]*HexCell, 0, 6)
	for _, nc := range coord.AllNeighbors() {
		if cell := g.GetCell(nc); cell != nil {
			neighbors = append(neighbors, cell)
		}
	}
	return neighbors
}

// GetPassableNeighbors returns neighboring cells that are passable
func (g *HexGrid) GetPassableNeighbors(coord HexCoord) []*HexCell {
	neighbors := make([]*HexCell, 0, 6)
	for _, nc := range coord.AllNeighbors() {
		if cell := g.GetCell(nc); cell != nil && cell.IsPassable() {
			neighbors = append(neighbors, cell)
		}
	}
	return neighbors
}

// GetLandNeighbors returns neighboring land cells
func (g *HexGrid) GetLandNeighbors(coord HexCoord) []*HexCell {
	neighbors := make([]*HexCell, 0, 6)
	for _, nc := range coord.AllNeighbors() {
		if cell := g.GetCell(nc); cell != nil && cell.IsLand {
			neighbors = append(neighbors, cell)
		}
	}
	return neighbors
}

// CellsInRegion returns all cells belonging to a region
func (g *HexGrid) CellsInRegion(regionID uuid.UUID) []*HexCell {
	cells := make([]*HexCell, 0)
	for _, cell := range g.Cells {
		if cell.RegionID != nil && *cell.RegionID == regionID {
			cells = append(cells, cell)
		}
	}
	return cells
}

// CountLandCells returns the number of land cells in the grid
func (g *HexGrid) CountLandCells() int {
	count := 0
	for _, cell := range g.Cells {
		if cell.IsLand {
			count++
		}
	}
	return count
}

// FindPath finds a path between two coordinates using BFS (land-only)
// Returns nil if no path exists
func (g *HexGrid) FindPath(start, end HexCoord) []HexCoord {
	startCell := g.GetCell(start)
	endCell := g.GetCell(end)
	if startCell == nil || endCell == nil {
		return nil
	}
	if !startCell.IsLand || !endCell.IsLand {
		return nil
	}

	// BFS
	visited := make(map[HexCoord]bool)
	parent := make(map[HexCoord]HexCoord)
	queue := []HexCoord{start}
	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == end {
			// Reconstruct path
			path := []HexCoord{end}
			for path[len(path)-1] != start {
				path = append(path, parent[path[len(path)-1]])
			}
			// Reverse
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path
		}

		for _, neighbor := range g.GetPassableNeighbors(current) {
			if !visited[neighbor.Coord] {
				visited[neighbor.Coord] = true
				parent[neighbor.Coord] = current
				queue = append(queue, neighbor.Coord)
			}
		}
	}

	return nil // No path
}
