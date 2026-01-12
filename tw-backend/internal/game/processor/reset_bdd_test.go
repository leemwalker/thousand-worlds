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

// TestBDD_WorldReset verifies the world reset functionality
// using a BDD style (Given/When/Then).
func TestBDD_WorldReset(t *testing.T) {
	// -------------------------------------------------------------------------
	// Setup (Given)
	// -------------------------------------------------------------------------
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil, nil, nil)

	// Mock Map Service needs to be set to check geology clearing
	proc.mapService = &mockMapService{
		geologyMap: make(map[uuid.UUID]*ecosystem.WorldGeology),
	}

	charID := uuid.New()
	userID := uuid.New()
	worldID := uuid.New()
	circ := 40000000.0

	mockWorldRepo.CreateWorld(context.Background(), &repository.World{
		ID:            worldID,
		Name:          "Reset Scenario World",
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
	}

	// Given: A running world simulation with data
	geo := ecosystem.NewWorldGeology(worldID, 12345, circ)
	proc.worldGeology[worldID] = geo
	proc.mapService.SetWorldGeology(worldID, geo)

	// Create runner
	runner := proc.getOrCreateRunner(worldID)
	require.NotNil(t, runner)

	// -------------------------------------------------------------------------
	// Scenario: Reset Active World
	// -------------------------------------------------------------------------
	t.Run("Reset Active World", func(t *testing.T) {
		// When: User executes 'world reset' command
		target := "reset"
		cmd := &websocket.CommandData{Action: "world", Target: &target}

		err := proc.ProcessCommand(context.Background(), client, cmd)
		require.NoError(t, err)

		// Then: Simulation runner should be stopped and removed
		assert.Nil(t, proc.getRunner(worldID), "Runner should be removed")

		// And: World data (geology) should be cleared
		_, hasGeo := proc.worldGeology[worldID]
		assert.False(t, hasGeo, "Geology should be removed from processor")
		assert.Nil(t, proc.mapService.GetWorldGeology(worldID), "Geology should be cleared from map service")

		// And: Client should receive reset confirmation and signal
		messages := client.GetMessages()
		foundReset := false
		foundSignal := false

		for _, m := range messages {
			if m.Type == "world_reset" {
				foundSignal = true
			}
			if m.Type == "system" && assert.ObjectsAreEqual(m.Text, "⏹️ Async simulation stopped.") {
				foundReset = true
			}
		}

		assert.True(t, foundSignal, "Should send world_reset signal")
		assert.True(t, foundReset, "Should confirm simulation stop")
	})
}
