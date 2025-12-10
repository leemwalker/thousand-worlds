package worldentity

import (
	"time"

	"github.com/google/uuid"
)

// EntityType defines the category of entity
type EntityType string

const (
	EntityTypeStatic    EntityType = "static"    // Walls, statues, doors
	EntityTypeNPC       EntityType = "npc"       // Non-player characters
	EntityTypeCreature  EntityType = "creature"  // Animals, monsters
	EntityTypeItem      EntityType = "item"      // Dropped/crafted items
	EntityTypePlant     EntityType = "plant"     // Harvestable flora
	EntityTypeStructure EntityType = "structure" // Buildings, player-made
	EntityTypeResource  EntityType = "resource"  // Ore, trees, etc.
)

// WorldEntity represents any entity in the game world
type WorldEntity struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	WorldID      uuid.UUID              `json:"world_id" db:"world_id"`
	EntityType   EntityType             `json:"entity_type" db:"entity_type"`
	Name         string                 `json:"name" db:"name"`
	Description  string                 `json:"description" db:"description"`
	Details      string                 `json:"details" db:"details"` // Shown when close/high perception
	X            float64                `json:"x" db:"x"`
	Y            float64                `json:"y" db:"y"`
	Z            float64                `json:"z" db:"z"`
	Collision    bool                   `json:"collision" db:"collision"`       // Blocks movement
	Locked       bool                   `json:"locked" db:"locked"`             // Prevents Get/Push/Move
	Interactable bool                   `json:"interactable" db:"interactable"` // Can be interacted with
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`         // Type-specific data, rendering
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// CollisionRadius returns the collision radius for this entity
func (e *WorldEntity) CollisionRadius() float64 {
	if e.Metadata != nil {
		if r, ok := e.Metadata["collision_radius"].(float64); ok {
			return r
		}
	}
	return 0.5 // Default 0.5m radius
}

// GetGlyph returns the rendering glyph for this entity
func (e *WorldEntity) GetGlyph() string {
	if e.Metadata != nil {
		if g, ok := e.Metadata["glyph"].(string); ok {
			return g
		}
	}
	// Default glyphs by type
	switch e.EntityType {
	case EntityTypeStatic:
		return "◼"
	case EntityTypeNPC:
		return "👤"
	case EntityTypeCreature:
		return "🐾"
	case EntityTypeItem:
		return "📦"
	case EntityTypePlant:
		return "🌿"
	case EntityTypeStructure:
		return "🏠"
	case EntityTypeResource:
		return "⛏"
	default:
		return "?"
	}
}
