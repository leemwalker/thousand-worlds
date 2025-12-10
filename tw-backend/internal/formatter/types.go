package formatter

// MessageType defines the type of game message
type MessageType string

const (
	MessageTypeMovement MessageType = "movement"
	MessageTypeCombat   MessageType = "combat"
	MessageTypeDialogue MessageType = "dialogue"
	MessageTypeSystem   MessageType = "system"
	MessageTypeError    MessageType = "error"
)

// EntityType defines the type of entity referenced in a text segment
type EntityType string

const (
	EntityTypeNPC      EntityType = "npc"
	EntityTypePlayer   EntityType = "player"
	EntityTypeItem     EntityType = "item"
	EntityTypeLocation EntityType = "location"
	EntityTypeResource EntityType = "resource"
)

// TextSegment represents a formatted piece of text with optional styling and entity references
type TextSegment struct {
	Text       string     `json:"text"`
	Color      string     `json:"color,omitempty"`
	Bold       bool       `json:"bold,omitempty"`
	Italic     bool       `json:"italic,omitempty"`
	Underline  bool       `json:"underline,omitempty"`
	EntityID   string     `json:"entityID,omitempty"`
	EntityType EntityType `json:"entityType,omitempty"`
}

// GameMessage represents a structured message sent to the client
type GameMessage struct {
	Type      MessageType   `json:"type"`
	Segments  []TextSegment `json:"segments"`
	Timestamp int64         `json:"timestamp"` // Unix timestamp in milliseconds
}
