package processor

import (
	"context"
	"testing"
	"tw-backend/cmd/game-server/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSpawnAsteroid_Impact(t *testing.T) {
	processor, client, _, _ := setupTest(t)

	// Normal impact
	cmd := &websocket.CommandData{
		Action: "spawn_asteroid",
		Text:   "spawn_asteroid 10.0 20.0 1000", // lat lon mass
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Since we mock Hub getClients, if Hub is nil, it sends to self.
	// Our setupTest mocks Hub now? Let's check setupTest in processor_test.go
	// It does: hub := websocket.NewHub(proc); proc.SetHub(hub)
	// But hub.GetClientsByWorldID returns empty if no clients registered?
	// The client created in setupTest isn't implicitly registered in the Hub's map unless we Register.
	// However, handleSpawnAsteroid fallback:
	// if p.Hub != nil { ... } else { client.SendGameMessage(...) }

	// To ensure we get the message, we should verify how broadcast works.
	// If Hub is set (which it is in setupTest), but client is not in it, we might get nothing.
	// We should manually register the client or modify setupTest.
	// OR, we can just assume the test environment needs p.Hub to be nil to force self-send,
	// OR we fix our expectation.

	// Let's force fallback for simplicity in THIS specific test wrapper?
	// Or better: assert that Hub broadcasting works.
	// We can manually add the client to the hub or simply check if the logic runs.

	// Hack: Set Hub to nil to force 'send to self' branch for easy testing
	processor.SetHub(nil)

	err = processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(client.messages), 1)
	msg := client.messages[len(client.messages)-1]
	assert.Equal(t, "asteroid_impact", msg.Type)

	data := msg.Metadata
	assert.Equal(t, 1000.0, data["mass"])
}

func TestHandleSpawnAsteroid_MoonFormation(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	processor.SetHub(nil) // Force fallback

	// Mass > Threshold (5e19)
	cmd := &websocket.CommandData{
		Action: "spawn_asteroid",
		Text:   "spawn_asteroid 0 0 6e19 planet",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(client.messages), 1)
	msg := client.messages[len(client.messages)-1]

	// Expect satellites_info update
	assert.Equal(t, "satellites_info", msg.Type)
	assert.Equal(t, "add", msg.Metadata["action"])

	moon := msg.Metadata["new_moon"].(map[string]interface{})
	assert.Equal(t, 6e19, moon["mass"])
}

func TestHandleSpawnAsteroid_MoonDestruction(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	processor.SetHub(nil) // Force fallback

	// Target moon, Mass > 1e12
	cmd := &websocket.CommandData{
		Action: "spawn_asteroid",
		Text:   "spawn_asteroid 0 0 2e12 moon Moon-1",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Should produce multiple messages: Moon Destroyed AND Ring Update
	// check last 2 messages
	require.GreaterOrEqual(t, len(client.messages), 2)

	// 1. Moon Destroyed
	// 2. Ring Update

	// Order might vary depending on broadcast, but typically sequential.
	// Let's find them
	foundDestroy := false
	foundRing := false

	for _, m := range client.messages {
		if m.Type == "moon_destroyed" {
			foundDestroy = true
			assert.Equal(t, "Moon-1", m.Metadata["moon_id"])
		}
		if m.Type == "satellites_info" && m.Metadata["action"] == "update_ring" {
			foundRing = true
		}
	}

	assert.True(t, foundDestroy, "Should have destroyed moon")
	assert.True(t, foundRing, "Should have created ring")
}
