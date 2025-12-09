package processor

import (
	"context"
	"testing"

	"mud-platform-backend/cmd/game-server/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleFace_Direction(t *testing.T) {
	processor, client, authRepo, _ := setupTest(t)
	// Character is created in setupTest
	// Ensure it's in a test world (Lobby is fine, or create a standard world)
	// Lobby world should work if DescribeView works for Lobby too.
	// But Lobby `DescribeView` might just call `DescribeRoom` which returns "Lobby".
	// Let's rely on standard world behavior if possible, or MockWorld.
	// setupTest uses Lobby.

	target := "north"
	cmd := &websocket.CommandData{
		Action: "face",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Check messages
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Equal(t, "info", lastMsg.Type)
	assert.Contains(t, lastMsg.Text, "You face North")

	// Verify orientation update in DB
	updatedChar, _ := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	assert.Equal(t, 0.0, updatedChar.OrientationX)
	assert.Equal(t, 1.0, updatedChar.OrientationY)
	assert.Equal(t, 0.0, updatedChar.OrientationZ)
}

func TestHandleFace_Status(t *testing.T) {
	processor, client, authRepo, _ := setupTest(t)
	char, _ := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	// Set initial orientation
	char.OrientationX = 1
	char.OrientationY = 0
	authRepo.UpdateCharacter(context.Background(), char)

	cmd := &websocket.CommandData{
		Action: "face",
		Text:   "face",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	lastMsg := client.messages[len(client.messages)-1]
	assert.Equal(t, "info", lastMsg.Type)
	assert.Contains(t, lastMsg.Text, "You are facing East")
}

func TestHandleFace_InvalidDirection(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "invalid"
	cmd := &websocket.CommandData{
		Action: "face",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	lastMsg := client.messages[len(client.messages)-1]
	assert.Equal(t, "error", lastMsg.Type)
	assert.Contains(t, lastMsg.Text, "Invalid direction")
}

func TestHandleFace_ImplicitText(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	// "face north" via text parsing
	cmd := &websocket.CommandData{
		Action: "face",
		Text:   "face north",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "You face North")
}
