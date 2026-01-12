package processor

import (
	"context"
	"testing"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetPOIs(t *testing.T) {
	processor, client, _, _ := setupTest(t)

	cmd := &websocket.CommandData{
		Action: "get_pois",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1) // Should have one response

	msg := client.messages[0]
	assert.Equal(t, "points_of_interest", msg.Type)

	data, ok := msg.Metadata["pois"].([]geography.PointOfInterest)
	if !ok {
		// It might be unmarshaled as map[string]interface{} in a real scenario if passing through JSON
		// But here we are checking the struct directly passed to SendGameMessage
		t.Logf("Metadata: %+v", msg.Metadata)
	}

	// Since we are mocking, we expect the slice of structs
	assert.True(t, ok, "Expected 'pois' to be []geography.PointOfInterest")
	assert.GreaterOrEqual(t, len(data), 1, "Should generate at least some POIs")

	// Validate System POIs
	foundSys := false
	for _, p := range data {
		if p.Name == "RAM Capital" || p.Name == "CPU Core" {
			foundSys = true
		}
		// ID should be valid UUID (not empty)
		assert.NotEqual(t, uuid.Nil, p.ID)
		assert.NotEmpty(t, p.Name)
		assert.NotEmpty(t, p.Type)
	}
	assert.True(t, foundSys, "Should find system POIs like RAM Capital or CPU Core")
}
