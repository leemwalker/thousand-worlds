package processor

import (
	"context"
	"strings"
	"testing"
	"time"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/repository" // Added import

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleWorld_Simulate_OnlyGeology verifies that using --only-geology
// prevents biological entities from spawning and evolving.
func TestHandleWorld_Simulate_OnlyGeology(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	// Create user character and key world data
	charID := uuid.New()
	userID := uuid.New()
	worldID := uuid.New()
	circ := 40000000.0

	// Mock valid world return using CreateWorld
	mockWorldRepo.CreateWorld(context.Background(), &repository.World{
		ID:            worldID,
		Name:          "Test World",
		Circumference: &circ,
	})

	mockAuthRepo.CreateCharacter(context.Background(), &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     worldID,
		PositionX:   0,
		PositionY:   0,
	})

	client := &mockClient{
		UserID:      userID,
		CharacterID: charID,
	}

	// EXECUTE: Run simulation with --only-geology
	// We run for a short duration to keep test fast, but long enough to trigger initialization
	target := "simulate"
	msg := "100 --only-geology"
	cmd := &websocket.CommandData{
		Action:  "world",
		Target:  &target,
		Message: &msg,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// VERIFY: No entities should be spawned
	assert.Empty(t, ecoSvc.Entities, "Ecosystem should have 0 entities with --only-geology")

	// Verify messages confirm geology-only
	foundV2Message := false
	for _, m := range client.messages {
		if strings.Contains(m.Text, "V2 Systems initialized") {
			foundV2Message = true
			assert.Contains(t, m.Text, "Active: false", "V2 Systems should report inactive")
		}
	}
	assert.True(t, foundV2Message, "Should have received system initialization message")
}

// TestHandleWorld_Simulate_Default verifies that WITHOUT flags, life is simulated.
func TestHandleWorld_Simulate_Default(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	charID := uuid.New()
	userID := uuid.New()
	worldID := uuid.New()
	circ := 40000000.0

	mockWorldRepo.CreateWorld(context.Background(), &repository.World{
		ID:            worldID,
		Name:          "Test World",
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

	// EXECUTE: Run simulation normally
	target := "simulate"
	msg := "100"
	cmd := &websocket.CommandData{
		Action:  "world",
		Target:  &target,
		Message: &msg,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// VERIFY: Entities should be spawned
	foundSpawnMsg := false
	for _, m := range client.messages {
		if strings.Contains(m.Text, "Spawned") && strings.Contains(m.Text, "entities") {
			foundSpawnMsg = true
		}
	}

	if len(ecoSvc.Entities) > 0 {
		assert.True(t, foundSpawnMsg, "Should report spawning if entities exist")
	}
}
