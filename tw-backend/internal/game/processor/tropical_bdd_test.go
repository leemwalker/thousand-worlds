package processor

import (
	"context"
	"strings"
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

// TestBDD_TropicalWorldEntry verifies the flow of entering the tropical test world
// and ensuring its physics parameters are correctly applied.
func TestBDD_TropicalWorldEntry(t *testing.T) {
	// -------------------------------------------------------------------------
	// Setup (Given)
	// -------------------------------------------------------------------------
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil, nil, nil)

	// User starts in the Lobby
	userID := uuid.New()
	charID := uuid.New()
	lobbyID := uuid.New()
	circ := 40000000.0

	mockWorldRepo.CreateWorld(context.Background(), &repository.World{
		ID:            lobbyID,
		Name:          "Lobby",
		Circumference: &circ,
	})

	mockAuthRepo.CreateCharacter(context.Background(), &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     lobbyID,
	})

	client := &mockClient{
		UserID:      userID,
		CharacterID: charID,
		WorldID:     lobbyID,
	}

	// -------------------------------------------------------------------------
	// Scenario: Enter Tropical World via Command
	// -------------------------------------------------------------------------
	t.Run("Enter Tropical World", func(t *testing.T) {
		// Given: User is in Lobby

		// When: User sends "enter_tropical_world" command (triggered by portal)
		cmd := &websocket.CommandData{Action: "enter_tropical_world"}
		err := proc.handleEnterTropicalWorld(context.Background(), client, cmd)
		require.NoError(t, err)

		// Then: User should be moved to "Tropical Test World"
		char, _ := mockAuthRepo.GetCharacter(context.Background(), charID)
		newWorld, _ := mockWorldRepo.GetWorld(context.Background(), char.WorldID)
		assert.Equal(t, "Tropical Test World", newWorld.Name)

		// And: Welcome message should be sent
		foundWelcome := false
		for _, m := range client.messages {
			if strings.Contains(m.Text, "Welcome to Tropical Test World") {
				foundWelcome = true
				break
			}
		}
		assert.True(t, foundWelcome, "Should receive welcome message")

		// And: World should have specific physics parameters (30h day, 2deg tilt)
		// We verify this by inspecting the in-memory geology instance that the processor holds
		geo := proc.worldGeology[newWorld.ID]
		require.NotNil(t, geo, "Geology should be initialized")

		// 30 hours = 108000 seconds
		assert.Equal(t, 108000.0, geo.Params.DayLengthSec, "Day length should be 30h")
		assert.Equal(t, 2.0, geo.Params.AxialTiltDeg, "Tilt should be 2 degrees")
	})
}
