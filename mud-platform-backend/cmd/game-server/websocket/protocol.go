package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Message types
const (
	MessageTypeCommand     = "command"
	MessageTypeGameMessage = "game_message"
	MessageTypeStateUpdate = "state_update"
	MessageTypeError       = "error"
)

// ClientMessage represents a message from client to server
type ClientMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// CommandData represents a player command
type CommandData struct {
	Action    string  `json:"action"`
	Target    *string `json:"target,omitempty"`
	Direction *string `json:"direction,omitempty"`
	Quantity  *int    `json:"quantity,omitempty"`
}

// ServerMessage represents a message from server to client
type ServerMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// GameMessageData represents a game event/message
type GameMessageData struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Text      string                 `json:"text"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// StateUpdateData represents the current game state
type StateUpdateData struct {
	HP           int             `json:"hp"`
	MaxHP        int             `json:"maxHP"`
	Stamina      int             `json:"stamina"`
	MaxStamina   int             `json:"maxStamina"`
	Focus        int             `json:"focus"`
	MaxFocus     int             `json:"maxFocus"`
	Position     Position        `json:"position"`
	Inventory    []InventoryItem `json:"inventory"`
	Equipment    Equipment       `json:"equipment"`
	VisibleTiles []VisibleTile   `json:"visibleTiles"`
}

// Position represents a location in the world
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// InventoryItem represents an item in inventory
type InventoryItem struct {
	ItemID    uuid.UUID `json:"itemID"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon,omitempty"`
	Quality   string    `json:"quality"`
	Quantity  int       `json:"quantity"`
	Weight    int       `json:"weight"`
	Equipable bool      `json:"equipable"`
	Slot      *string   `json:"slot,omitempty"`
}

// Equipment represents equipped items
type Equipment struct {
	Head     *InventoryItem `json:"head"`
	Chest    *InventoryItem `json:"chest"`
	Legs     *InventoryItem `json:"legs"`
	Feet     *InventoryItem `json:"feet"`
	MainHand *InventoryItem `json:"mainHand"`
	OffHand  *InventoryItem `json:"offHand"`
}

// VisibleTile represents a tile the player can see
type VisibleTile struct {
	X         int         `json:"x"`
	Y         int         `json:"y"`
	Biome     string      `json:"biome"`
	Elevation int         `json:"elevation"`
	Entities  []MapEntity `json:"entities"`
}

// MapEntity represents an entity on the map
type MapEntity struct {
	ID   string `json:"id"`
	Type string `json:"type"` // npc, player, resource, monster
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// ErrorData represents an error message
type ErrorData struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
