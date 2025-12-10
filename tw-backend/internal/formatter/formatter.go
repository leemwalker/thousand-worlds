package formatter

import "fmt"

// OutputFormatter formats game events into structured messages for the client
type OutputFormatter struct{}

// NewOutputFormatter creates a new OutputFormatter instance
func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{}
}

// FormatMovement formats a movement event
func (f *OutputFormatter) FormatMovement(direction, location string) *GameMessage {
	segments := []TextSegment{
		{Text: "You move ", Color: "#888"},
		{Text: direction, Color: "#4a9eff", Bold: true},
		{Text: " to ", Color: "#888"},
		{Text: location, Color: "#ffeb3b", EntityType: EntityTypeLocation},
	}

	return NewGameMessage(MessageTypeMovement, segments)
}

// FormatCombat formats a combat event
func (f *OutputFormatter) FormatCombat(attacker, target string, damage int, targetID string) *GameMessage {
	segments := []TextSegment{
		{Text: attacker, Color: "#4caf50", Bold: true},
		{Text: " hit ", Color: "#888"},
		{Text: target, Color: "#f44336", EntityID: targetID, EntityType: EntityTypeNPC},
		{Text: " for ", Color: "#888"},
		{Text: fmt.Sprintf("%d", damage), Color: "#ff9800", Bold: true},
		{Text: " damage", Color: "#888"},
	}

	return NewGameMessage(MessageTypeCombat, segments)
}

// FormatDialogue formats NPC/player dialogue
func (f *OutputFormatter) FormatDialogue(speaker, speakerID, message string) *GameMessage {
	segments := []TextSegment{
		{Text: speaker, Color: "#9c27b0", Bold: true, EntityID: speakerID, EntityType: EntityTypeNPC},
		{Text: " says: ", Color: "#888"},
		{Text: fmt.Sprintf("\"%s\"", message), Color: "#e1bee7", Italic: true},
	}

	return NewGameMessage(MessageTypeDialogue, segments)
}

// FormatSystem formats a system message
func (f *OutputFormatter) FormatSystem(message string) *GameMessage {
	segments := []TextSegment{
		{Text: message, Color: "#00bcd4"},
	}

	return NewGameMessage(MessageTypeSystem, segments)
}

// FormatError formats an error message
func (f *OutputFormatter) FormatError(message string) *GameMessage {
	segments := []TextSegment{
		{Text: message, Color: "#f44336", Bold: true},
	}

	return NewGameMessage(MessageTypeError, segments)
}
