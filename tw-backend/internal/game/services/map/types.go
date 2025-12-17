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
	RenderQuality RenderQuality `json:"render_quality"`
	GridSize      int           `json:"grid_size"` // 9 on ground, 17 when flying
	Scale         int           `json:"scale"`     // World units per tile (1 = ground, >1 = flying)
	WorldID       uuid.UUID     `json:"world_id"`
}
