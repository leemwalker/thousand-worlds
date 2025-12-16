// Package geography provides ecological region management for species isolation.
// Regions group hex cells and track isolation duration for gigantism/dwarfism effects.
package geography

import (
	"github.com/google/uuid"
)

// ObstacleType represents barriers that hinder migration
type ObstacleType string

const (
	ObstacleNone     ObstacleType = "none"
	ObstacleMountain ObstacleType = "mountain"
	ObstacleOcean    ObstacleType = "ocean"
	ObstacleDesert   ObstacleType = "desert"
	ObstacleRiver    ObstacleType = "river"
	ObstacleIce      ObstacleType = "ice"
	ObstacleVolcanic ObstacleType = "volcanic"
)

// Region represents an ecological region composed of hex cells
type Region struct {
	ID             uuid.UUID          `json:"id"`
	Name           string             `json:"name"`
	WorldID        uuid.UUID          `json:"world_id"`
	BiomeID        *uuid.UUID         `json:"biome_id,omitempty"` // Primary biome type
	Cells          []HexCoord         `json:"cells"`              // All cells in this region
	Connections    []RegionConnection `json:"connections"`        // Links to adjacent regions
	IsolationYears int64              `json:"isolation_years"`    // Years since connected to mainland
	IsIsland       bool               `json:"is_island"`          // True if surrounded by water
	Area           int                `json:"area"`               // Number of cells
	Perimeter      int                `json:"perimeter"`          // Number of edge cells
	CreatedYear    int64              `json:"created_year"`
}

// RegionConnection represents a connection between two regions
type RegionConnection struct {
	TargetRegionID uuid.UUID    `json:"target_region_id"`
	Obstacle       ObstacleType `json:"obstacle"`      // Primary barrier type
	Difficulty     float32      `json:"difficulty"`    // 0.0 (easy) to 1.0 (nearly impassable)
	Width          int          `json:"width"`         // Number of cells at narrowest point
	BridgeCoords   []HexCoord   `json:"bridge_coords"` // Cells forming the connection
}

// NewRegion creates a new ecological region
func NewRegion(name string, worldID uuid.UUID, createdYear int64) *Region {
	return &Region{
		ID:          uuid.New(),
		Name:        name,
		WorldID:     worldID,
		Cells:       make([]HexCoord, 0),
		Connections: make([]RegionConnection, 0),
		CreatedYear: createdYear,
	}
}

// AddCell adds a hex cell to the region
func (r *Region) AddCell(coord HexCoord) {
	r.Cells = append(r.Cells, coord)
	r.Area = len(r.Cells)
}

// ContainsCell returns true if the region contains the given cell
func (r *Region) ContainsCell(coord HexCoord) bool {
	for _, c := range r.Cells {
		if c == coord {
			return true
		}
	}
	return false
}

// AddConnection adds or updates a connection to another region
func (r *Region) AddConnection(targetID uuid.UUID, obstacle ObstacleType, difficulty float32, width int, bridgeCoords []HexCoord) {
	// Check if connection already exists
	for i, conn := range r.Connections {
		if conn.TargetRegionID == targetID {
			r.Connections[i] = RegionConnection{
				TargetRegionID: targetID,
				Obstacle:       obstacle,
				Difficulty:     difficulty,
				Width:          width,
				BridgeCoords:   bridgeCoords,
			}
			return
		}
	}

	r.Connections = append(r.Connections, RegionConnection{
		TargetRegionID: targetID,
		Obstacle:       obstacle,
		Difficulty:     difficulty,
		Width:          width,
		BridgeCoords:   bridgeCoords,
	})
}

// RemoveConnection removes a connection to another region
func (r *Region) RemoveConnection(targetID uuid.UUID) {
	for i, conn := range r.Connections {
		if conn.TargetRegionID == targetID {
			r.Connections = append(r.Connections[:i], r.Connections[i+1:]...)
			return
		}
	}
}

// IsIsolated returns true if the region has no easy connections
func (r *Region) IsIsolated() bool {
	if len(r.Connections) == 0 {
		return true
	}
	// Check if all connections are difficult
	for _, conn := range r.Connections {
		if conn.Difficulty < 0.8 {
			return false
		}
	}
	return true
}

// GetIsolationModifier returns trait modifiers based on isolation duration
// Long isolation leads to island effects (gigantism for small species, dwarfism for large)
func (r *Region) GetIsolationModifier() IsolationModifier {
	if r.IsolationYears < 100000 {
		return IsolationModifier{} // No effect for short isolation
	}

	// Strength of effect scales with isolation time (capped at 10M years)
	years := r.IsolationYears
	if years > 10000000 {
		years = 10000000
	}
	strength := float32(years) / 10000000.0

	// Islands have stronger effects
	if r.IsIsland {
		strength *= 1.5
		if strength > 1.0 {
			strength = 1.0
		}
	}

	return IsolationModifier{
		Strength:         strength,
		SmallSizeMod:     1.0 + strength*0.5, // Small species get bigger (island gigantism)
		LargeSizeMod:     1.0 - strength*0.4, // Large species get smaller (island dwarfism)
		FertilityMod:     1.0 - strength*0.1, // Lower fertility in isolation
		AggressionMod:    1.0 - strength*0.3, // Island species often less aggressive
		FearOfNoveltyMod: 1.0 - strength*0.5, // Island species often naive (tameness)
	}
}

// IsolationModifier contains trait modifiers for isolated regions
type IsolationModifier struct {
	Strength         float32 `json:"strength"`       // 0.0-1.0
	SmallSizeMod     float32 `json:"small_size_mod"` // Multiplier for small species size
	LargeSizeMod     float32 `json:"large_size_mod"` // Multiplier for large species size
	FertilityMod     float32 `json:"fertility_mod"`
	AggressionMod    float32 `json:"aggression_mod"`
	FearOfNoveltyMod float32 `json:"fear_novelty_mod"` // Island tameness effect
}

// ApplyToSize applies island rule size modification
// Returns the modified size based on original size
func (m IsolationModifier) ApplyToSize(originalSize float64) float64 {
	if m.Strength == 0 {
		return originalSize
	}

	// Island rule: small become larger, large become smaller
	// Threshold around size 3.0 (medium animal)
	if originalSize < 3.0 {
		return originalSize * float64(m.SmallSizeMod)
	} else if originalSize > 6.0 {
		return originalSize * float64(m.LargeSizeMod)
	}
	// Medium-sized species are less affected
	return originalSize
}

// RegionSystem manages all regions in a world
type RegionSystem struct {
	WorldID      uuid.UUID             `json:"world_id"`
	Regions      map[uuid.UUID]*Region `json:"regions"`
	CurrentYear  int64                 `json:"current_year"`
	UpdatePeriod int64                 `json:"update_period"` // Years between re-evaluations
}

// NewRegionSystem creates a new region management system
func NewRegionSystem(worldID uuid.UUID) *RegionSystem {
	return &RegionSystem{
		WorldID:      worldID,
		Regions:      make(map[uuid.UUID]*Region),
		UpdatePeriod: 10000, // Re-evaluate every 10,000 years
	}
}

// AddRegion adds a region to the system
func (rs *RegionSystem) AddRegion(region *Region) {
	rs.Regions[region.ID] = region
}

// GetRegion retrieves a region by ID
func (rs *RegionSystem) GetRegion(id uuid.UUID) *Region {
	return rs.Regions[id]
}

// Update advances the region system and updates isolation counters
func (rs *RegionSystem) Update(years int64, grid *HexGrid, tectonics *TectonicSystem) {
	rs.CurrentYear += years

	for _, region := range rs.Regions {
		// Update isolation years for isolated regions
		if region.IsIsolated() {
			region.IsolationYears += years
		} else {
			// Reset isolation counter for connected regions
			// (with decay - some effects persist)
			if region.IsolationYears > 0 {
				region.IsolationYears -= years / 10 // Slow decay
				if region.IsolationYears < 0 {
					region.IsolationYears = 0
				}
			}
		}

		// Update island status
		region.IsIsland = rs.checkIsIsland(region, grid)
	}
}

// checkIsIsland determines if a region is an island (surrounded by water)
func (rs *RegionSystem) checkIsIsland(region *Region, grid *HexGrid) bool {
	if len(region.Cells) == 0 {
		return false
	}

	// Check if any cell has a non-water neighbor outside the region
	for _, coord := range region.Cells {
		for _, neighbor := range grid.GetNeighbors(coord) {
			if !region.ContainsCell(neighbor.Coord) && neighbor.IsLand {
				return false // Has land connection outside region
			}
		}
	}

	return true // Surrounded by water or region boundary only
}

// IdentifyRegions scans the hex grid and creates regions from connected land masses
func (rs *RegionSystem) IdentifyRegions(grid *HexGrid) {
	// Clear existing regions
	rs.Regions = make(map[uuid.UUID]*Region)

	visited := make(map[HexCoord]bool)
	regionNum := 1

	for coord, cell := range grid.Cells {
		if visited[coord] || !cell.IsLand {
			continue
		}

		// BFS to find connected land mass
		region := NewRegion(
			generateRegionName(regionNum),
			rs.WorldID,
			rs.CurrentYear,
		)
		regionNum++

		queue := []HexCoord{coord}
		visited[coord] = true

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			region.AddCell(current)

			// Update cell's region ID
			if c := grid.GetCell(current); c != nil {
				c.RegionID = &region.ID
			}

			// Visit land neighbors
			for _, neighbor := range grid.GetLandNeighbors(current) {
				if !visited[neighbor.Coord] {
					visited[neighbor.Coord] = true
					queue = append(queue, neighbor.Coord)
				}
			}
		}

		rs.AddRegion(region)
	}

	// Identify connections between regions
	rs.identifyConnections(grid)
}

// identifyConnections finds connections between adjacent regions
func (rs *RegionSystem) identifyConnections(grid *HexGrid) {
	// For each region, find neighboring regions
	for _, region := range rs.Regions {
		neighborRegions := make(map[uuid.UUID][]HexCoord) // Region ID -> bridge cells

		for _, coord := range region.Cells {
			for _, neighbor := range grid.GetNeighbors(coord) {
				if neighbor.RegionID != nil && *neighbor.RegionID != region.ID {
					neighborRegions[*neighbor.RegionID] = append(
						neighborRegions[*neighbor.RegionID],
						coord,
					)
				}
			}
		}

		// Create connections
		for targetID, bridgeCoords := range neighborRegions {
			// Determine obstacle type from terrain between regions
			obstacle := ObstacleNone
			maxDifficulty := float32(0)

			for _, coord := range bridgeCoords {
				for _, neighbor := range grid.GetNeighbors(coord) {
					if !region.ContainsCell(neighbor.Coord) {
						switch neighbor.Terrain {
						case TerrainMountain:
							obstacle = ObstacleMountain
							maxDifficulty = max32(maxDifficulty, 0.8)
						case TerrainOcean:
							obstacle = ObstacleOcean
							maxDifficulty = max32(maxDifficulty, 1.0)
						case TerrainVolcanic:
							obstacle = ObstacleVolcanic
							maxDifficulty = max32(maxDifficulty, 0.9)
						}
					}
				}
			}

			region.AddConnection(targetID, obstacle, maxDifficulty, len(bridgeCoords), bridgeCoords)
		}
	}
}

// generateRegionName creates a name for a region
func generateRegionName(num int) string {
	prefixes := []string{"North", "South", "East", "West", "Central", "Greater", "Lesser"}
	names := []string{"Lands", "Territories", "Expanse", "Domain", "Realm", "Province"}

	prefix := prefixes[num%len(prefixes)]
	name := names[(num/len(prefixes))%len(names)]

	return prefix + " " + name
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
