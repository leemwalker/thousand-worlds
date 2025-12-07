package formatter

import "time"

// NewGameMessage creates a new GameMessage with the current timestamp
func NewGameMessage(msgType MessageType, segments []TextSegment) *GameMessage {
	return &GameMessage{
		Type:      msgType,
		Segments:  segments,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewTextSegment creates a plain text segment
func NewTextSegment(text string) TextSegment {
	return TextSegment{Text: text}
}

// NewColoredSegment creates a colored text segment
func NewColoredSegment(text, color string) TextSegment {
	return TextSegment{Text: text, Color: color}
}

// NewEntitySegment creates a text segment with an entity reference
func NewEntitySegment(text, entityID string, entityType EntityType) TextSegment {
	return TextSegment{
		Text:       text,
		EntityID:   entityID,
		EntityType: entityType,
	}
}
