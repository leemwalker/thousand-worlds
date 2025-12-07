package formatter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewGameMessage(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		segments []TextSegment
		wantType MessageType
	}{
		{
			name:     "creates movement message",
			msgType:  MessageTypeMovement,
			segments: []TextSegment{{Text: "You move north"}},
			wantType: MessageTypeMovement,
		},
		{
			name:     "creates combat message",
			msgType:  MessageTypeCombat,
			segments: []TextSegment{{Text: "You hit the goblin", Color: "red"}},
			wantType: MessageTypeCombat,
		},
		{
			name:     "creates dialogue message",
			msgType:  MessageTypeDialogue,
			segments: []TextSegment{{Text: "Hello", EntityID: "npc-1", EntityType: EntityTypeNPC}},
			wantType: MessageTypeDialogue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewGameMessage(tt.msgType, tt.segments)

			assert.Equal(t, tt.wantType, msg.Type)
			assert.Equal(t, tt.segments, msg.Segments)
			assert.True(t, msg.Timestamp > 0, "Timestamp should be set")
			assert.True(t, msg.Timestamp <= time.Now().UnixMilli(), "Timestamp should not be in future")
		})
	}
}

func TestTextSegment(t *testing.T) {
	tests := []struct {
		name    string
		segment TextSegment
		wantErr bool
	}{
		{
			name:    "plain text segment",
			segment: TextSegment{Text: "Hello"},
			wantErr: false,
		},
		{
			name:    "colored text segment",
			segment: TextSegment{Text: "Error!", Color: "#ff0000"},
			wantErr: false,
		},
		{
			name:    "bold text segment",
			segment: TextSegment{Text: "Important", Bold: true},
			wantErr: false,
		},
		{
			name:    "entity reference segment",
			segment: TextSegment{Text: "Goblin", EntityID: "mob-123", EntityType: EntityTypeNPC},
			wantErr: false,
		},
		{
			name:    "fully styled segment",
			segment: TextSegment{Text: "Legendary Sword", Color: "#ffd700", Bold: true, Italic: true, EntityID: "item-456", EntityType: EntityTypeItem},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.segment.Text, "Text should not be empty")
		})
	}
}
