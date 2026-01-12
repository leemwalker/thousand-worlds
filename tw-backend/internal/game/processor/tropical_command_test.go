package processor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/repository"
)

func TestHandleEnterTropicalWorld(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	// Create map service dependencies
	// Note: We need a functional processor to test this
	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil, nil, nil)

	// Initialize map service manually if NewGameProcessor didn't (it does)
	// But we might need to mock things if map service depends on them?
	// NewGameProcessor initializes Real map service.
	// MapService needs WorldRepo, which is mocked.

	// Setup User & Character
	userID := uuid.New()
	charID := uuid.New()
	initialWorldID := uuid.New()

	// Create Initial World
	circ := 40000000.0
	mockWorldRepo.CreateWorld(context.Background(), &repository.World{
		ID:            initialWorldID,
		Name:          "Lobby",
		Circumference: &circ,
	})

	mockAuthRepo.CreateCharacter(context.Background(), &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     initialWorldID,
		PositionX:   100,
		PositionY:   100,
	})

	client := &mockClient{
		UserID:      userID,
		CharacterID: charID,
		WorldID:     initialWorldID,
	}

	// 1. First Call: Generate World
	cmd := &websocket.CommandData{
		Action: "enter_tropical_world",
	}

	err := proc.handleEnterTropicalWorld(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify Character Moved
	char, err := mockAuthRepo.GetCharacter(context.Background(), charID)
	require.NoError(t, err)
	assert.NotEqual(t, initialWorldID, char.WorldID)

	// Verify World Parameters
	newWorld, err := mockWorldRepo.GetWorld(context.Background(), char.WorldID)
	require.NoError(t, err)
	assert.Equal(t, "Tropical Test World", newWorld.Name)

	// Check Metadata
	dayLen, ok := newWorld.Metadata["day_length_hours"].(float64)
	assert.True(t, ok)
	assert.Equal(t, 30.0, dayLen)

	tilt, ok := newWorld.Metadata["axial_tilt"].(float64)
	assert.True(t, ok)
	assert.Equal(t, 2.0, tilt)

	// Verify Geology Config
	geo, exists := proc.worldGeology[newWorld.ID]
	assert.True(t, exists, "Geology should be stored in memory")
	assert.NotNil(t, geo.Heightmap, "Heightmap should be generated")

	// Verify Physics Params
	// Day length: 30 hours * 3600 = 108000
	assert.Equal(t, 108000.0, geo.Params.DayLengthSec, "Day length should be 30 hours")
	// Tilt: 2.0 degrees
	assert.Equal(t, 2.0, geo.Params.AxialTiltDeg, "Axial tilt should be 2.0 degrees")

	// Verify Welcome Message
	foundWelcome := false
	for _, msg := range client.messages {
		if strings.Contains(msg.Text, "Welcome to Tropical Test World") {
			foundWelcome = true
			break
		}
	}
	assert.True(t, foundWelcome, "Should send welcome message")

	// 2. Second Call: Enter Existing World
	// Reset character position to lobby
	char.WorldID = initialWorldID
	mockAuthRepo.UpdateCharacter(context.Background(), char)
	client.messages = nil // Clear messages

	err = proc.handleEnterTropicalWorld(context.Background(), client, cmd)
	require.NoError(t, err)

	// Should be same world ID
	charAfter, _ := mockAuthRepo.GetCharacter(context.Background(), charID)
	assert.Equal(t, newWorld.ID, charAfter.WorldID)

	foundExisting := false
	// We might log "Found existing..." but client gets same welcome message
	// Let's check log output? Hard in test.
	// Check client message again
	for _, msg := range client.messages {
		if strings.Contains(msg.Text, "Welcome to Tropical Test World") {
			foundExisting = true
			break
		}
	}
	assert.True(t, foundExisting)
}
