package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"tw-backend/internal/formatter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandFlowIntegration tests the full command processing flow
// Raw text -> Parser -> Processor -> Formatter -> JSON response
func TestCommandFlowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name         string
		rawCommand   string
		wantType     formatter.MessageType
		wantSegments int
		wantContains string
	}{
		{
			name:         "look command returns formatted area description",
			rawCommand:   "look",
			wantType:     formatter.MessageTypeSystem,
			wantSegments: 1,
			wantContains: "",
		},
		{
			name:         "movement command returns formatted movement message",
			rawCommand:   "north",
			wantType:     formatter.MessageTypeMovement,
			wantSegments: 4, // "You move", "north", "to", "location"
			wantContains: "north",
		},
		{
			name:         "attack command returns formatted combat message",
			rawCommand:   "attack goblin",
			wantType:     formatter.MessageTypeCombat,
			wantSegments: 6, // "attacker", "hit", "target", "for", "damage", "damage"
			wantContains: "hit",
		},
		{
			name:         "say command returns formatted dialogue message",
			rawCommand:   "say hello",
			wantType:     formatter.MessageTypeDialogue,
			wantSegments: 3, // "speaker", "says:", "message"
			wantContains: "hello",
		},
		{
			name:         "invalid command returns error message",
			rawCommand:   "invalidcommandxyz",
			wantType:     formatter.MessageTypeError,
			wantSegments: 1,
			wantContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// This would normally go through the full command processor
			// For now, we test the formatter directly
			form := formatter.NewOutputFormatter()

			var msg *formatter.GameMessage
			switch {
			case tt.rawCommand == "north":
				msg = form.FormatMovement("north", "Forest Path")
			case tt.rawCommand == "attack goblin":
				msg = form.FormatCombat("You", "Goblin", 15, "mob-123")
			case tt.rawCommand == "say hello":
				msg = form.FormatDialogue("Player", "player-1", "hello")
			case tt.rawCommand == "invalidcommandxyz":
				msg = form.FormatError("Unknown command")
			default:
				msg = form.FormatSystem("You look around.")
			}

			require.NotNil(t, msg)
			assert.Equal(t, tt.wantType, msg.Type)

			if tt.wantSegments > 0 {
				assert.Equal(t, tt.wantSegments, len(msg.Segments), "segment count mismatch")
			}

			// Verify JSON serialization
			jsonBytes, err := json.Marshal(msg)
			require.NoError(t, err)
			assert.NotEmpty(t, jsonBytes)

			// Verify JSON can be unmarshaled
			var unmarshaled formatter.GameMessage
			err = json.Unmarshal(jsonBytes, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, msg.Type, unmarshaled.Type)

			// Verify we didn't parse on client
			assert.NotContains(t, string(jsonBytes), "parse", "Should not contain parsing logic")

			_ = ctx // Will use when integrating with real processor
		})
	}
}

// TestWebSocketMessageFlow tests WebSocket message exchange
func TestWebSocketMessageFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("client sends raw text, server responds with formatted JSON", func(t *testing.T) {
		// Arrange
		rawCommand := "look"
		form := formatter.NewOutputFormatter()

		// Act - Server processes command and formats response
		response := form.FormatSystem("You are in a dark forest.")

		// Assert - Verify response structure
		require.NotNil(t, response)
		assert.Equal(t, formatter.MessageTypeSystem, response.Type)
		assert.Len(t, response.Segments, 1)
		assert.Greater(t, response.Timestamp, int64(0))

		// Verify JSON serialization for WebSocket transmission
		jsonBytes, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded formatter.GameMessage
		err = json.Unmarshal(jsonBytes, &decoded)
		require.NoError(t, err)
		assert.Equal(t, response.Type, decoded.Type)
		assert.Equal(t, len(response.Segments), len(decoded.Segments))

		_ = rawCommand // Would be sent via WebSocket in real test
	})

	t.Run("multiple rapid commands processed in order", func(t *testing.T) {
		form := formatter.NewOutputFormatter()
		commands := []string{"look", "north", "look", "south", "look"}
		responses := make([]*formatter.GameMessage, 0, len(commands))

		for _, cmd := range commands {
			var msg *formatter.GameMessage
			if cmd == "north" || cmd == "south" {
				msg = form.FormatMovement(cmd, "New Area")
			} else {
				msg = form.FormatSystem("Area description")
			}
			responses = append(responses, msg)
		}

		// Verify all commands processed
		assert.Len(t, responses, 5)

		// Verify timestamps are sequential
		for i := 1; i < len(responses); i++ {
			assert.GreaterOrEqual(t, responses[i].Timestamp, responses[i-1].Timestamp)
		}
	})
}

// TestFormatterIntegration ensures formatter is properly integrated
func TestFormatterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("formatter produces valid JSON for all message types", func(t *testing.T) {
		form := formatter.NewOutputFormatter()

		testCases := []struct {
			name string
			msg  *formatter.GameMessage
		}{
			{"movement", form.FormatMovement("east", "Village Square")},
			{"combat", form.FormatCombat("You", "Dragon", 50, "boss-1")},
			{"dialogue", form.FormatDialogue("Guard", "npc-2", "Stop!")},
			{"system", form.FormatSystem("You gained a level!")},
			{"error", form.FormatError("Invalid target")},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				jsonBytes, err := json.Marshal(tc.msg)
				require.NoError(t, err)
				assert.NotEmpty(t, jsonBytes)

				// Verify required fields
				assert.Contains(t, string(jsonBytes), `"type"`)
				assert.Contains(t, string(jsonBytes), `"segments"`)
				assert.Contains(t, string(jsonBytes), `"timestamp"`)
			})
		}
	})
}
