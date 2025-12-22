package gamemap

import (
	"github.com/google/uuid"
)

// RenderQuality determines the visual fidelity of map tiles
type RenderQuality string

const (
	QualityLow    RenderQuality = "low"    // ASCII characters: . # @ ~
	QualityMedium RenderQuality = "medium" // Simple icons with colors
	QualityHigh   RenderQuality = "high"   // Full emoji icons
)

// GetRenderQuality determines quality based on perception skill level (0-100)
func GetRenderQuality(perception int) RenderQuality {
	if perception >= 71 {
		return QualityHigh
	}
	if perception >= 31 {
		return QualityMedium
	}
	return QualityLow
}

// MapTile represents a single tile in the mini-map grid
type MapTile struct {
	X           int         `json:"x"`
	Y           int         `json:"y"`
	Biome       string      `json:"biome"`
	Elevation   float64     `json:"elevation"`
	Entities    []MapEntity `json:"entities,omitempty"`
	Portal      *PortalInfo `json:"portal,omitempty"`
	IsPlayer    bool        `json:"is_player,omitempty"`
	OutOfBounds bool        `json:"out_of_bounds,omitempty"` // True if tile is outside world bounds
	Occluded    bool        `json:"occluded,omitempty"`      // True if tile is hidden by terrain (LOS)
}

// MapEntity represents an entity visible on the map
type MapEntity struct {
	ID     uuid.UUID `json:"id"`
	Type   string    `json:"type"` // "player", "npc", "resource", "monster", "static", "item", etc
	Name   string    `json:"name,omitempty"`
	Status string    `json:"status,omitempty"` // "friendly", "neutral", "hostile"
	Glyph  string    `json:"glyph,omitempty"`  // Custom glyph for rendering (e.g., "🗿" for statue)
}

// PortalInfo represents a portal visible on the map
type PortalInfo struct {
	WorldName   string `json:"world_name,omitempty"`
	Destination string `json:"destination,omitempty"`
	Active      bool   `json:"active"`
}

// MapData contains all data needed to render the mini-map
type MapData struct {
	Tiles         []MapTile     `json:"tiles"`
	PlayerX       float64       `json:"player_x"`
	PlayerY       float64       `json:"player_y"`
	PlayerZ       float64       `json:"player_z"` // Player altitude for zoom
	RenderQuality RenderQuality `json:"render_quality"`
	GridSize      int           `json:"grid_size"` // 9 on ground, 17 when flying
	Scale         int           `json:"scale"`     // World units per tile (1 = ground, >1 = flying)
	WorldID       uuid.UUID     `json:"world_id"`
	IsSimulated   bool          `json:"is_simulated"` // False for lobby/unsimulated worlds
}

// WorldMapTile represents an aggregated tile for the full world map
// Each tile represents a region of the world (e.g., 100x100 world units)
type WorldMapTile struct {
	GridX        int     `json:"grid_x"`              // Grid X position (0-based)
	GridY        int     `json:"grid_y"`              // Grid Y position (0-based)
	Biome        string  `json:"biome"`               // Dominant biome in this region
	AvgElevation float64 `json:"avg_elevation"`       // Average elevation
	IsPlayer     bool    `json:"is_player,omitempty"` // Player is in this region
}

// WorldMapData contains aggregated data for full world map display
type WorldMapData struct {
	Tiles       []WorldMapTile `json:"tiles"`
	GridWidth   int            `json:"grid_width"`   // Number of columns
	GridHeight  int            `json:"grid_height"`  // Number of rows
	WorldWidth  float64        `json:"world_width"`  // World width in units
	WorldHeight float64        `json:"world_height"` // World height in units
	PlayerX     float64        `json:"player_x"`     // Player X position
	PlayerY     float64        `json:"player_y"`     // Player Y position
	WorldID     uuid.UUID      `json:"world_id"`
	WorldName   string         `json:"world_name,omitempty"`
	IsSimulated bool           `json:"is_simulated"` // False for lobby/unsimulated worlds

	// Simulation summary data (populated after simulation)
	AvgTemperature float64 `json:"avg_temperature,omitempty"` // Average temperature in Celsius
	MaxElevation   float64 `json:"max_elevation,omitempty"`   // Maximum elevation in meters
	SeaLevel       float64 `json:"sea_level,omitempty"`       // Sea level in meters
	LandCoverage   float64 `json:"land_coverage,omitempty"`   // Percentage of land above sea level
	SimulatedYears int64   `json:"simulated_years,omitempty"` // Total years simulated
	Seed           int64   `json:"seed,omitempty"`            // Simulation seed used
}
