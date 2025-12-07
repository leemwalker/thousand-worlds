package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputFormatter follows TDD - tests written first
func TestOutputFormatter_FormatMovement(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		location  string
		want      *GameMessage
	}{
		{
			name:      "north movement",
			direction: "north",
			location:  "Forest Path",
			want: &GameMessage{
				Type: MessageTypeMovement,
				Segments: []TextSegment{
					{Text: "You move ", Color: "#888"},
					{Text: "north", Color: "#4a9eff", Bold: true},
					{Text: " to ", Color: "#888"},
					{Text: "Forest Path", Color: "#ffeb3b", EntityType: EntityTypeLocation},
				},
			},
		},
		{
			name:      "south movement",
			direction: "south",
			location:  "Dark Cave",
			want: &GameMessage{
				Type: MessageTypeMovement,
				Segments: []TextSegment{
					{Text: "You move ", Color: "#888"},
					{Text: "south", Color: "#4a9eff", Bold: true},
					{Text: " to ", Color: "#888"},
					{Text: "Dark Cave", Color: "#ffeb3b", EntityType: EntityTypeLocation},
				},
			},
		},
	}

	formatter := NewOutputFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatMovement(tt.direction, tt.location)

			require.NotNil(t, got)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, len(tt.want.Segments), len(got.Segments))

			for i, wantSeg := range tt.want.Segments {
				assert.Equal(t, wantSeg.Text, got.Segments[i].Text)
				assert.Equal(t, wantSeg.Color, got.Segments[i].Color)
				assert.Equal(t, wantSeg.Bold, got.Segments[i].Bold)
				assert.Equal(t, wantSeg.EntityType, got.Segments[i].EntityType)
			}
		})
	}
}

func TestOutputFormatter_FormatCombat(t *testing.T) {
	tests := []struct {
		name     string
		attacker string
		target   string
		damage   int
		targetID string
		want     *GameMessage
	}{
		{
			name:     "player hits enemy",
			attacker: "You",
			target:   "Goblin",
			damage:   15,
			targetID: "mob-123",
			want: &GameMessage{
				Type: MessageTypeCombat,
				Segments: []TextSegment{
					{Text: "You", Color: "#4caf50", Bold: true},
					{Text: " hit ", Color: "#888"},
					{Text: "Goblin", Color: "#f44336", EntityID: "mob-123", EntityType: EntityTypeNPC},
					{Text: " for ", Color: "#888"},
					{Text: "15", Color: "#ff9800", Bold: true},
					{Text: " damage", Color: "#888"},
				},
			},
		},
	}

	formatter := NewOutputFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatCombat(tt.attacker, tt.target, tt.damage, tt.targetID)

			require.NotNil(t, got)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, len(tt.want.Segments), len(got.Segments))

			for i, wantSeg := range tt.want.Segments {
				assert.Equal(t, wantSeg.Text, got.Segments[i].Text)
				assert.Equal(t, wantSeg.Color, got.Segments[i].Color)
			}
		})
	}
}

func TestOutputFormatter_FormatDialogue(t *testing.T) {
	tests := []struct {
		name      string
		speaker   string
		speakerID string
		message   string
		want      *GameMessage
	}{
		{
			name:      "NPC dialogue",
			speaker:   "Guard",
			speakerID: "npc-guard-1",
			message:   "Halt! Who goes there?",
			want: &GameMessage{
				Type: MessageTypeDialogue,
				Segments: []TextSegment{
					{Text: "Guard", Color: "#9c27b0", Bold: true, EntityID: "npc-guard-1", EntityType: EntityTypeNPC},
					{Text: " says: ", Color: "#888"},
					{Text: "\"Halt! Who goes there?\"", Color: "#e1bee7", Italic: true},
				},
			},
		},
	}

	formatter := NewOutputFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatDialogue(tt.speaker, tt.speakerID, tt.message)

			require.NotNil(t, got)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, len(tt.want.Segments), len(got.Segments))
		})
	}
}

func TestOutputFormatter_FormatSystem(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantType MessageType
	}{
		{
			name:     "system message",
			message:  "Server shutting down in 5 minutes",
			wantType: MessageTypeSystem,
		},
		{
			name:     "info message",
			message:  "You gained 100 experience",
			wantType: MessageTypeSystem,
		},
	}

	formatter := NewOutputFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatSystem(tt.message)

			require.NotNil(t, got)
			assert.Equal(t, tt.wantType, got.Type)
			assert.Len(t, got.Segments, 1)
			assert.Equal(t, tt.message, got.Segments[0].Text)
		})
	}
}

func TestOutputFormatter_FormatError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantType MessageType
	}{
		{
			name:     "error message",
			message:  "Command not found",
			wantType: MessageTypeError,
		},
		{
			name:     "validation error",
			message:  "Invalid target",
			wantType: MessageTypeError,
		},
	}

	formatter := NewOutputFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatError(tt.message)

			require.NotNil(t, got)
			assert.Equal(t, tt.wantType, got.Type)
			assert.Len(t, got.Segments, 1)
			assert.Contains(t, got.Segments[0].Color, "f44336") // Red color
		})
	}
}
