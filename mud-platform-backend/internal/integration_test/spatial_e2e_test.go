package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mud-platform-backend/cmd/game-server/websocket"
	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/game/processor"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/player"
	"mud-platform-backend/internal/repository"
)

// TestSpatialMovementE2E tests spatial navigation end-to-end
func TestSpatialMovementE2E(t *testing.T) {
	t.Run("Happy Path: Move in all directions in lobby", func(t *testing.T) {
		proc, client, authRepo, worldRepo := setupSpatialTest(t)

		// Add lobby world
		lobbyWorld := &repository.World{
			ID:   lobby.LobbyWorldID,
			Name: "Lobby",
		}
		worldRepo.worlds[lobbyWorld.ID] = lobbyWorld

		// Create character at center of lobby
		char := &auth.Character{
			CharacterID: client.CharacterID,
			UserID:      client.UserID,
			WorldID:     lobby.LobbyWorldID,
			PositionX:   5.0,
			PositionY:   500.0,
			Name:        "TestCharacter",
		}
		authRepo.CreateCharacter(context.Background(), char)

		directions := []struct {
			direction string
			expectedY float64
		}{
			{"n", 501.0},
			{"s", 500.0}, // Back to 500
			{"e", 500.0}, // Y stays same
			{"w", 500.0}, // Y stays same
		}

		for _, d := range directions {
			client.Messages = nil // Reset messages
			cmd := &websocket.CommandData{Action: d.direction}
			err := proc.ProcessCommand(context.Background(), client, cmd)
			require.NoError(t, err)
			assert.Len(t, client.Messages, 1)
			assert.Contains(t, client.Messages[0].Text, "You move")
		}
	})

	t.Run("Sad Path: Move into wall (boundary)", func(t *testing.T) {
		proc, client, authRepo, worldRepo := setupSpatialTest(t)

		lobbyWorld := &repository.World{
			ID:   lobby.LobbyWorldID,
			Name: "Lobby",
		}
		worldRepo.worlds[lobbyWorld.ID] = lobbyWorld

		// Create character at north boundary
		char := &auth.Character{
			CharacterID: client.CharacterID,
			UserID:      client.UserID,
			WorldID:     lobby.LobbyWorldID,
			PositionX:   5.0,
			PositionY:   1000.0, // At north wall
			Name:        "TestCharacter",
		}
		authRepo.CreateCharacter(context.Background(), char)

		// Try to move north into wall
		cmd := &websocket.CommandData{Action: "n"}
		err := proc.ProcessCommand(context.Background(), client, cmd)
		require.NoError(t, err)
		assert.Contains(t, client.Messages[0].Text, "cannot go further")
	})

	t.Run("Sad Path: Move at east boundary", func(t *testing.T) {
		proc, client, authRepo, worldRepo := setupSpatialTest(t)

		lobbyWorld := &repository.World{
			ID:   lobby.LobbyWorldID,
			Name: "Lobby",
		}
		worldRepo.worlds[lobbyWorld.ID] = lobbyWorld

		// Create character at east boundary
		char := &auth.Character{
			CharacterID: client.CharacterID,
			UserID:      client.UserID,
			WorldID:     lobby.LobbyWorldID,
			PositionX:   10.0, // At east wall
			PositionY:   500.0,
			Name:        "TestCharacter",
		}
		authRepo.CreateCharacter(context.Background(), char)

		// Try to move east into wall
		cmd := &websocket.CommandData{Action: "e"}
		err := proc.ProcessCommand(context.Background(), client, cmd)
		require.NoError(t, err)
		assert.Contains(t, client.Messages[0].Text, "cannot go further")
	})
}

// TestPortalTransitionE2E tests portal entry and exit
func TestPortalTransitionE2E(t *testing.T) {
	t.Run("Happy Path: Enter world from lobby", func(t *testing.T) {
		proc, client, authRepo, worldRepo := setupSpatialTest(t)

		// Add lobby and target world
		lobbyWorld := &repository.World{
			ID:   lobby.LobbyWorldID,
			Name: "Lobby",
		}
		targetWorld := &repository.World{
			ID:   uuid.New(),
			Name: "TestWorld",
		}
		worldRepo.worlds[lobbyWorld.ID] = lobbyWorld
		worldRepo.worlds[targetWorld.ID] = targetWorld

		char := &auth.Character{
			CharacterID: client.CharacterID,
			UserID:      client.UserID,
			WorldID:     lobby.LobbyWorldID,
			PositionX:   5.0,
			PositionY:   500.0,
			Name:        "TestCharacter",
		}
		authRepo.CreateCharacter(context.Background(), char)

		// Enter world by name
		target := "TestWorld"
		cmd := &websocket.CommandData{
			Action: "enter",
			Target: &target,
		}
		err := proc.ProcessCommand(context.Background(), client, cmd)
		require.NoError(t, err)

		// Lobby's enter command triggers entry_options modal
		assert.Len(t, client.Messages, 1)
		assert.Equal(t, "trigger_entry_options", client.Messages[0].Type)
	})

	t.Run("Sad Path: Enter non-existent world", func(t *testing.T) {
		proc, client, authRepo, worldRepo := setupSpatialTest(t)

		lobbyWorld := &repository.World{
			ID:   lobby.LobbyWorldID,
			Name: "Lobby",
		}
		worldRepo.worlds[lobbyWorld.ID] = lobbyWorld

		char := &auth.Character{
			CharacterID: client.CharacterID,
			UserID:      client.UserID,
			WorldID:     lobby.LobbyWorldID,
			PositionX:   5.0,
			PositionY:   500.0,
			Name:        "TestCharacter",
		}
		authRepo.CreateCharacter(context.Background(), char)

		// Try to enter non-existent world
		target := "NonExistentWorld"
		cmd := &websocket.CommandData{
			Action: "enter",
			Target: &target,
		}
		err := proc.ProcessCommand(context.Background(), client, cmd)
		require.NoError(t, err)

		assert.Len(t, client.Messages, 1)
		assert.Contains(t, client.Messages[0].Text, "no portal")
	})

	t.Run("Sad Path: Enter without target", func(t *testing.T) {
		proc, client, authRepo, worldRepo := setupSpatialTest(t)

		lobbyWorld := &repository.World{
			ID:   lobby.LobbyWorldID,
			Name: "Lobby",
		}
		worldRepo.worlds[lobbyWorld.ID] = lobbyWorld

		char := &auth.Character{
			CharacterID: client.CharacterID,
			UserID:      client.UserID,
			WorldID:     lobby.LobbyWorldID,
			PositionX:   5.0,
			PositionY:   500.0,
			Name:        "TestCharacter",
		}
		authRepo.CreateCharacter(context.Background(), char)

		// Try to enter without specifying target
		cmd := &websocket.CommandData{
			Action: "enter",
		}
		err := proc.ProcessCommand(context.Background(), client, cmd)

		// Should return error for missing target
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target world ID required")
	})
}

// Helper to set up spatial tests
func setupSpatialTest(t *testing.T) (*processor.GameProcessor, *TestGameClient, *auth.MockRepository, *StatefulMockWorldRepository) {
	authRepo := auth.NewMockRepository()
	worldRepo := NewStatefulMockWorldRepository()

	lookService := lobby.NewLookService(authRepo, worldRepo, nil, nil)
	spatialSvc := player.NewSpatialService(authRepo, worldRepo)

	proc := processor.NewGameProcessor(authRepo, worldRepo, lookService, nil, spatialSvc)

	hub := websocket.NewHub(proc)
	proc.SetHub(hub)

	client := &TestGameClient{
		UserID:      uuid.New(),
		CharacterID: uuid.New(),
		WorldID:     lobby.LobbyWorldID,
		Username:    "TestUser",
	}

	return proc, client, authRepo, worldRepo
}
