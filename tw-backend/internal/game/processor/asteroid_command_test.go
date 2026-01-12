package processor

import (
	"context"
	"testing"
	"time"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleSpawnAsteroid verifies the spawn_asteroid command logic
func TestHandleSpawnAsteroid(t *testing.T) {
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
		nil, nil,
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
			name:          "Valid Asteroid",
			commandText:   "spawn_asteroid 10 20 1000",
			expectedError: "",
			expectedMsg:   "asteroid_impact",
		},
		{
			name:          "Invalid args",
			commandText:   "spawn_asteroid 0",
			expectedError: "", // Command returns nil error but sends error message
			expectedMsg:   "error",
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
			found := false
			for _, m := range client.messages {
				if m.Type == tt.expectedMsg {
					found = true
					if tt.expectedMsg == "asteroid_impact" {
						dataMap := m.Metadata
						// impactData has "type", "location" (map[string]float64), "mass", "impact_time", "origin".
						// t.Logf("Full Metadata: %+v", dataMap)

						assert.Equal(t, "asteroid_impact", dataMap["type"])
						assert.NotNil(t, dataMap["location"], "Location data should be present")

						// Location was defined as map[string]float64 in handler
						loc, ok := dataMap["location"].(map[string]float64)
						assert.True(t, ok, "Location should be a map[string]float64")

						if ok {
							// Verify values match input
							assert.Equal(t, 10.0, loc["lat"])
							assert.Equal(t, 20.0, loc["lon"])
						}
					}
					break
				}
				// Also check if it's an error message when expectedMsg is "error"
				if tt.expectedMsg == "error" && m.Type == "error" {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected message type: %s", tt.expectedMsg)
		})
	}
}
