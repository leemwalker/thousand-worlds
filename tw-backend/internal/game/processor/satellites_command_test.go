package processor

import (
	"context"
	"testing"
	"time"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/repository"
	"tw-backend/internal/worldgen/astronomy"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleDestroyMoon verifies the destroy_moon command logic
func TestHandleDestroyMoon(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(
		mockAuthRepo,
		mockWorldRepo,
		nil, nil, nil, nil, nil, nil, nil, nil,
		ecoSvc,
		nil, nil, nil, nil, nil, nil,
		nil, nil, // Added missing args (Publisher, Ollama)
	)

	// Create user + character + world
	charID := uuid.New()
	userID := uuid.New()
	worldID := uuid.New()
	circ := 40000000.0

	mockWorldRepo.CreateWorld(context.Background(), &repository.World{
		ID:            worldID,
		Name:          "Test World",
		Circumference: &circ,
	})

	mockAuthRepo.CreateCharacter(context.Background(), &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     worldID,
	})

	// Manually initialize Geology with a Moon
	geo := &ecosystem.WorldGeology{
		WorldID: worldID,
		Satellites: []astronomy.Satellite{
			{Name: "Moon1", Mass: 1e20},
		},
	}
	proc.worldGeology[worldID] = geo

	client := &mockClient{
		UserID:      userID,
		CharacterID: charID,
		messages:    make([]websocket.GameMessageData, 0),
	}

	tests := []struct {
		name          string
		commandText   string
		expectedError string
		expectedMsg   string
	}{
		{
			name:          "Valid Destruction",
			commandText:   "destroy_moon Moon1",
			expectedError: "",
			expectedMsg:   "Moon destroyed",
		},
		{
			name:          "Moon Not Found",
			commandText:   "destroy_moon UnknownMoon",
			expectedError: "Moon 'UnknownMoon' not found",
		},
		{
			name:          "Missing Argument",
			commandText:   "destroy_moon",
			expectedError: "Usage: destroy_moon <moon_name>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.messages = make([]websocket.GameMessageData, 0) // Clear messages

			cmd := &websocket.CommandData{
				Text: tt.commandText,
			}

			err := proc.ProcessCommand(context.Background(), client, cmd)
			require.NoError(t, err)

			// Verify messages
			if tt.expectedError != "" {
				found := false
				for _, m := range client.messages {
					if m.Type == "error" && m.Text == tt.expectedError {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error message: %s", tt.expectedError)
			} else {
				// Verify success message and payload
				found := false
				for _, m := range client.messages {
					if m.Type == "game_message" && m.Text == tt.expectedMsg {
						found = true
						// Verify metadata
						// Metadata is already map[string]interface{}
						dataMap := m.Metadata
						assert.NotNil(t, dataMap)
						assert.Equal(t, "moon_destroyed", dataMap["type"])

						meta, ok := dataMap["metadata"].(map[string]interface{})
						assert.True(t, ok, "Inner metadata should be a map")

						assert.Equal(t, "Moon1", meta["moon_id"])
						break
					}
				}
				assert.True(t, found, "Expected success message: %s", tt.expectedMsg)
			}
		})
	}
}
