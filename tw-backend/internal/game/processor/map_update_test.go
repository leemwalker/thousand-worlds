package processor

import (
	"context"
	"testing"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/game/constants"
	"tw-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMapUpdateOnConnect verifies that a map_update is sent when client connects
// RED: This test should fail initially because OnClientConnected doesn't send map_update
func TestMapUpdateOnConnect(t *testing.T) {
	// Setup
	processor, client, authRepo, _ := setupTest(t)

	// Get the character for context
	char, err := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	require.NoError(t, err)
	require.NotNil(t, char)

	// Clear any messages from setup
	client.messages = nil

	// Simulate client connection callback
	processor.OnClientConnected(context.Background(), client)

	// Verify map_update was sent
	var foundMapUpdate bool
	for _, msg := range client.messages {
		if msg.Type == "map_update" {
			foundMapUpdate = true
			break
		}
	}

	assert.True(t, foundMapUpdate, "Expected map_update message to be sent on client connect")
}

// TestMapUpdateAfterWorldChange verifies that a map_update is sent after entering a new world
// RED: This test should fail initially because handleWatcher doesn't send map_update
func TestMapUpdateAfterWorldChange(t *testing.T) {
	processor, client, authRepo, worldRepo := setupTest(t)

	// Setup: Create character in lobby
	char, err := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	require.NoError(t, err)
	char.WorldID = constants.LobbyWorldID
	_ = authRepo.UpdateCharacter(context.Background(), char)

	// Create target world
	targetWorld := &repository.World{
		ID:   uuid.New(),
		Name: "TestWorld",
	}
	worldRepo.worlds[targetWorld.ID] = targetWorld

	// Clear messages from setup
	client.messages = nil

	// Execute: Enter as watcher
	target := targetWorld.ID.String()
	cmd := &websocket.CommandData{
		Action: "watcher",
		Target: &target,
	}

	err = processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify: map_update should be sent after world change
	var foundMapUpdate bool
	for _, msg := range client.messages {
		if msg.Type == "map_update" {
			foundMapUpdate = true
			break
		}
	}

	assert.True(t, foundMapUpdate, "Expected map_update message after entering new world")
}

// TestMapUpdateWatcherLongDistance verifies map_update is sent after watcher moves 100+ units
// Tests Issue 7: Watcher movement >10 units should update minimap correctly
func TestMapUpdateWatcherLongDistance(t *testing.T) {
	processor, client, authRepo, worldRepo := setupTest(t)

	// Create a spherical world (no bounds = spherical)
	circumference := 10000.0
	targetWorld := &repository.World{
		ID:            uuid.New(),
		Name:          "TestWorld",
		Circumference: &circumference,
		// No BoundsMin/BoundsMax = spherical world with wrapping
	}
	worldRepo.worlds[targetWorld.ID] = targetWorld

	// Setup: Character in lobby, enter as watcher
	char, _ := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	char.WorldID = constants.LobbyWorldID
	_ = authRepo.UpdateCharacter(context.Background(), char)

	// Enter world as watcher first
	target := targetWorld.ID.String()
	_ = processor.ProcessCommand(context.Background(), client, &websocket.CommandData{
		Action: "watcher",
		Target: &target,
	})

	// Get initial position
	watcherChar, _ := authRepo.GetCharacterByUserAndWorld(context.Background(), client.GetUserID(), targetWorld.ID)
	initialX := watcherChar.PositionX
	initialY := watcherChar.PositionY
	t.Logf("Initial position: (%.1f, %.1f)", initialX, initialY)

	// Clear messages
	client.messages = nil

	// Move 500 units east as watcher
	distance := "500"
	cmd := &websocket.CommandData{
		Action: "east",
		Target: &distance,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify: map_update should be sent
	var foundMapUpdate bool
	var mapMetadata map[string]interface{}
	for _, msg := range client.messages {
		if msg.Type == "map_update" {
			foundMapUpdate = true
			mapMetadata = msg.Metadata
			break
		}
	}

	assert.True(t, foundMapUpdate, "Expected map_update after long-distance movement")

	// Verify player position changed
	if mapMetadata != nil {
		playerX, _ := mapMetadata["player_x"].(float64)
		playerY, _ := mapMetadata["player_y"].(float64)
		t.Logf("Map update position: (%.1f, %.1f)", playerX, playerY)

		// Position should have changed from initial (0,0) to (500, 0)
		assert.NotEqual(t, initialX, playerX, "Player X should have changed after movement")
	}
}
