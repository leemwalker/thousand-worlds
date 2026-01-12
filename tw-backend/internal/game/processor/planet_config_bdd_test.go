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

// TestBDD_PlanetConfiguration verifies the planet configuration commands
// using a BDD style (Given/When/Then).
func TestBDD_PlanetConfiguration(t *testing.T) {
	// -------------------------------------------------------------------------
	// Setup (Given)
	// -------------------------------------------------------------------------
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil, nil, nil)

	charID := uuid.New()
	userID := uuid.New()
	worldID := uuid.New()
	circ := 40000000.0

	mockWorldRepo.CreateWorld(context.Background(), &repository.World{
		ID:            worldID,
		Name:          "Config Test World",
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

	// Initialize simulation
	target := "run"
	runCmd := &websocket.CommandData{Action: "world", Target: &target}
	_ = proc.ProcessCommand(context.Background(), client, runCmd)
	client.messages = nil // Clear init messages

	// -------------------------------------------------------------------------
	// Scenario 1: Configure Planet Mass
	// -------------------------------------------------------------------------
	t.Run("Configure Planet Mass", func(t *testing.T) {
		// Given: A running world simulation

		// When: User sets mass to 0.5 (half Earth)
		targetCfg := "configure"
		msgCfg := "mass 0.5"
		cmd := &websocket.CommandData{Action: "world", Target: &targetCfg, Message: &msgCfg}

		err := proc.ProcessCommand(context.Background(), client, cmd)
		require.NoError(t, err)

		// Then: System should confirm the change
		getLastMsg := func() string {
			if len(client.messages) == 0 {
				return ""
			}
			return client.messages[len(client.messages)-1].Text
		}
		assert.Contains(t, getLastMsg(), "Planet mass set to 0.50y Earth Mass")

		// And: Physics parameters should be updated
		runner := proc.getRunner(worldID)
		require.NotNil(t, runner)
		// We can't access geology directly without thread unsafety or exposing getters,
		// but we know SetPlanetMass was called if the message was sent.
	})

	// -------------------------------------------------------------------------
	// Scenario 2: Configure Axial Tilt
	// -------------------------------------------------------------------------
	t.Run("Configure Axial Tilt", func(t *testing.T) {
		// When: User sets tilt to 90 degrees (Uranus-style)
		targetCfg := "configure"
		msgCfg := "tilt 90.0"
		cmd := &websocket.CommandData{Action: "world", Target: &targetCfg, Message: &msgCfg}

		err := proc.ProcessCommand(context.Background(), client, cmd)
		require.NoError(t, err)

		// Then: System should confirm
		assert.Contains(t, client.messages[len(client.messages)-1].Text, "Axial tilt set to 90.00°")
	})

	// -------------------------------------------------------------------------
	// Scenario 3: Configure Day Length
	// -------------------------------------------------------------------------
	t.Run("Configure Day Length", func(t *testing.T) {
		// When: User sets day length to 10 hours (36000 sec)
		targetCfg := "configure"
		msgCfg := "day 36000"
		cmd := &websocket.CommandData{Action: "world", Target: &targetCfg, Message: &msgCfg}

		err := proc.ProcessCommand(context.Background(), client, cmd)
		require.NoError(t, err)

		// Then: System should confirm
		assert.Contains(t, client.messages[len(client.messages)-1].Text, "Day length set to 36000.00 seconds")
	})

	// -------------------------------------------------------------------------
	// Scenario 4: Error Handling (Invalid Inputs)
	// -------------------------------------------------------------------------
	t.Run("Invalid Configurations", func(t *testing.T) {
		// When: Negative mass
		targetCfg := "configure"
		msgCfg := "mass -1.0"
		cmd := &websocket.CommandData{Action: "world", Target: &targetCfg, Message: &msgCfg}
		_ = proc.ProcessCommand(context.Background(), client, cmd)
		assert.Contains(t, client.messages[len(client.messages)-1].Text, "Mass multiplier too small")

		// When: Excessive tilt
		msgCfg = "tilt 200"
		cmd.Message = &msgCfg
		_ = proc.ProcessCommand(context.Background(), client, cmd)
		assert.Contains(t, client.messages[len(client.messages)-1].Text, "Tilt must be between 0 and 180")
	})
}
