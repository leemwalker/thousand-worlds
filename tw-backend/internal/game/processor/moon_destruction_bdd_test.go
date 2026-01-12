package processor

import (
	"context"
	"testing"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/worldgen/astronomy"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBDD_MoonDestruction verifies the full flow of destroying a moon.
// Feature: Moon Fragmentation
// Scenario: Player destroys a moon
func TestBDD_MoonDestruction(t *testing.T) {
	// --- GIVEN ---
	// A World with a Moon named "TargetMoon"
	proc, client, authRepo, _ := setupTest(t)

	// Access geology directly to inject moon (white-box setup)
	// Get character's world ID
	char, err := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	require.NoError(t, err)
	worldID := char.WorldID

	// Create a new world for this test to specific ID?
	// setupTest puts char in LobbyWorldID. Lobby doesn't have geology usually?
	// Let's create a new world ID and move char there.
	// But wait, setupTest creates services.
	// We need to initialize `worldGeology` map in proc.

	geo := &ecosystem.WorldGeology{
		WorldID: worldID,
		Satellites: []astronomy.Satellite{
			{Name: "TargetMoon", Mass: 5e20},
			{Name: "SurvivorMoon", Mass: 3e20},
		},
		Heightmap: &geography.Heightmap{Width: 10, Height: 10, Elevations: make([]float64, 100)},
	}
	// Inject geology
	proc.worldGeology[worldID] = geo

	// --- TO VERIFY SETUP ---
	// Client asks for satellites
	cmdGet := &websocket.CommandData{Action: "get_satellites"}
	client.messages = make([]websocket.GameMessageData, 0)
	err = proc.ProcessCommand(context.Background(), client, cmdGet)
	require.NoError(t, err)

	// Check we got the list
	foundInfo := false
	for _, m := range client.messages {
		if m.Type == "satellites_info" {
			foundInfo = true
			data := m.Metadata
			sats := data["satellites"].([]astronomy.Satellite)
			assert.Len(t, sats, 2)
			break
		}
	}
	require.True(t, foundInfo, "Setup failed: Could not get satellites")

	// --- WHEN ---
	// Player sends destroy_moon command
	client.messages = make([]websocket.GameMessageData, 0) // Clear messages
	cmdDestroy := &websocket.CommandData{
		Text: "destroy_moon TargetMoon",
	}
	err = proc.ProcessCommand(context.Background(), client, cmdDestroy)
	require.NoError(t, err)

	// --- THEN ---
	// 1. Receive moon_destroyed event
	var destroyedMsg *websocket.GameMessageData
	for _, m := range client.messages {
		if m.Type == "game_message" && m.Text == "Moon destroyed" {
			destroyedMsg = &m
			break
		}
	}
	require.NotNil(t, destroyedMsg, "Did not receive destruction confirmation")

	meta := destroyedMsg.Metadata["metadata"].(map[string]interface{})
	assert.Equal(t, "TargetMoon", meta["moon_id"])

	// 2. Moon should be removed from backend state
	// Verify by calling get_satellites again
	client.messages = make([]websocket.GameMessageData, 0)
	err = proc.ProcessCommand(context.Background(), client, cmdGet)
	require.NoError(t, err)

	foundInfo = false
	for _, m := range client.messages {
		if m.Type == "satellites_info" {
			foundInfo = true
			data := m.Metadata
			sats := data["satellites"].([]astronomy.Satellite)

			// Should only have SurvivorMoon
			assert.Len(t, sats, 1)
			assert.Equal(t, "SurvivorMoon", sats[0].Name)
			break
		}
	}
	require.True(t, foundInfo, "Verification failed: Could not get updated satellites")
}
