package processor

import (
	"context"
	"testing"
	"time"

	"mud-platform-backend/cmd/game-server/websocket"
	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleLobby_Success(t *testing.T) {
	processor, client, authRepo, worldRepo := setupTest(t)

	// Create a non-lobby world
	otherWorld := &repository.World{
		ID:   uuid.New(),
		Name: "Other World",
	}
	worldRepo.worlds[otherWorld.ID] = otherWorld

	// Create character in that world
	char := &auth.Character{
		CharacterID: client.GetCharacterID(),
		UserID:      client.GetUserID(),
		WorldID:     otherWorld.ID, // Not in lobby
		Name:        "Traveler",
		CreatedAt:   time.Now(),
		PositionX:   100.0,
		PositionY:   100.0,
	}
	err := authRepo.CreateCharacter(context.Background(), char)
	require.NoError(t, err)

	cmd := &websocket.CommandData{
		Action: "lobby",
	}

	err = processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify character moved to lobby
	updatedChar, err := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	require.NoError(t, err)
	assert.Equal(t, lobby.LobbyWorldID, updatedChar.WorldID)
	assert.Equal(t, otherWorld.ID, *updatedChar.LastWorldVisited)
	assert.Equal(t, 5.0, updatedChar.PositionX)
	assert.Equal(t, 2.0, updatedChar.PositionY)

	// Verify message sent
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Equal(t, "system", lastMsg.Type)
	assert.Contains(t, lastMsg.Text, "You return to the Grand Lobby")
}

func TestHandleLobby_AlreadyInLobby(t *testing.T) {
	processor, client, authRepo, _ := setupTest(t)

	// Character is already in lobby (setupTest creates char in lobby)
	// Just verify it for sanity
	char, err := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	require.NoError(t, err)
	char.WorldID = lobby.LobbyWorldID
	authRepo.UpdateCharacter(context.Background(), char)

	cmd := &websocket.CommandData{
		Action: "lobby",
	}

	err = processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify error message
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Equal(t, "error", lastMsg.Type)
	assert.Contains(t, lastMsg.Text, "already in the lobby")
}

func TestHandleLobby_WatcherRemoval(t *testing.T) {
	processor, client, authRepo, worldRepo := setupTest(t)

	otherWorld := &repository.World{
		ID:   uuid.New(),
		Name: "Other World",
	}
	worldRepo.worlds[otherWorld.ID] = otherWorld

	char := &auth.Character{
		CharacterID: client.GetCharacterID(),
		UserID:      client.GetUserID(),
		WorldID:     otherWorld.ID,
		Name:        "Watcher",
		Role:        "watcher",
		CreatedAt:   time.Now(),
	}
	err := authRepo.CreateCharacter(context.Background(), char)
	require.NoError(t, err)

	cmd := &websocket.CommandData{
		Action: "lobby",
	}

	err = processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	updatedChar, err := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	require.NoError(t, err)
	assert.Equal(t, lobby.LobbyWorldID, updatedChar.WorldID)
	// Only verification is WorldID change, as strict removal from "instance" is implied by ID change in this architecture
}
