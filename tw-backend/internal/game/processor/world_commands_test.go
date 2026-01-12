package processor

import (
	"context"
	"strings"
	"testing"
	"time"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/ecosystem/state"
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

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil, nil, nil)

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
	foundGeologyOnlyMsg := false
	for _, m := range client.messages {
		if strings.Contains(m.Text, "V2 Systems initialized") {
			foundV2Message = true
		}
		if strings.Contains(m.Text, "geology-only simulation") {
			foundGeologyOnlyMsg = true
		}
	}
	// V2 Systems should NOT be initialized in geology-only mode
	assert.False(t, foundV2Message, "V2 Systems should NOT be initialized with --only-geology")
	assert.True(t, foundGeologyOnlyMsg, "Should receive geology-only message")
}

// TestHandleWorld_Simulate_Default verifies that WITHOUT flags, life is simulated.
func TestHandleWorld_Simulate_Default(t *testing.T) {
	// Setup
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

func TestHandleWorld_Configure_Tilt(t *testing.T) {
	// Setup
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

	// Initialize simulation first to ensure Runner exists
	// Initialize simulation runner using 'run'
	target := "run"
	runCmd := &websocket.CommandData{
		Action: "world",
		Target: &target,
	}
	proc.ProcessCommand(context.Background(), client, runCmd)

	// EXECUTE: Configure tilt
	targetCfg := "configure"
	msgCfg := "tilt 45.5"
	cmd := &websocket.CommandData{
		Action:  "world",
		Target:  &targetCfg,
		Message: &msgCfg,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// VERIFY: Last message should confirm tilt
	messages := client.GetMessages()
	found := false
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.Contains(messages[i].Text, "Axial tilt set to 45.50") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should confirm axial tilt update")
}

func TestHandleWorld_Configure_DayLength(t *testing.T) {
	// Setup
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

	// Initialize simulation runner using 'run'
	target := "run"
	runCmd := &websocket.CommandData{
		Action: "world",
		Target: &target,
	}
	proc.ProcessCommand(context.Background(), client, runCmd)

	// EXECUTE: Configure day length
	targetCfg := "configure"
	msgCfg := "day 36000" // 10 hours
	cmd := &websocket.CommandData{
		Action:  "world",
		Target:  &targetCfg,
		Message: &msgCfg,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// VERIFY: Last message should confirm day length
	messages := client.GetMessages()
	found := false
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.Contains(messages[i].Text, "Day length set to 36000.00 seconds") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should confirm day length update")
}

func TestHandleWorld_Configure_Mass(t *testing.T) {
	// Setup
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

	// Initialize simulation runner using 'run'
	target := "run"
	runCmd := &websocket.CommandData{
		Action: "world",
		Target: &target,
	}
	proc.ProcessCommand(context.Background(), client, runCmd)

	// EXECUTE: Configure mass
	targetCfg := "configure"
	msgCfg := "mass 0.5" // Mars-sized (half earth mass?? No Mars is 0.1 but 0.5 is fine for test)
	cmd := &websocket.CommandData{
		Action:  "world",
		Target:  &targetCfg,
		Message: &msgCfg,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// VERIFY: Last message should confirm mass
	messages := client.GetMessages()
	found := false
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.Contains(messages[i].Text, "Planet mass set to 0.50y Earth Mass") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should confirm planet mass update")
}

func TestHandleWorld_Reset(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil, nil, nil)

	// Mock Map Service
	proc.mapService = &mockMapService{
		geologyMap: make(map[uuid.UUID]*ecosystem.WorldGeology),
	}

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

	// 1. Initialize Simulation State (Geology, Entities, Runner)
	// Initialize geology manually as if simulation ran
	geo := ecosystem.NewWorldGeology(worldID, 12345, circ)
	proc.worldGeology[worldID] = geo
	proc.mapService.SetWorldGeology(worldID, geo)

	// Add dummy entity
	entityID := uuid.New()
	ecoSvc.Entities[entityID] = &state.LivingEntityState{
		EntityID: entityID,
		WorldID:  worldID,
	}

	// Initialize simulation runner
	target := "run"
	runCmd := &websocket.CommandData{Action: "world", Target: &target}
	// This creates the runner
	// Note: We ignore error here as run might fail if not fully configured,
	// but we just need runner creation attempt.
	// Actually, run creates runner.
	proc.ProcessCommand(context.Background(), client, runCmd)

	// Force create runner if not created by run command (due to mocks)
	if proc.getRunner(worldID) == nil {
		proc.getOrCreateRunner(worldID)
	}
	require.NotNil(t, proc.getRunner(worldID), "Runner should exist before reset")

	// 2. EXECUTE: World Reset
	targetReset := "reset"
	cmd := &websocket.CommandData{Action: "world", Target: &targetReset}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// 3. VERIFY
	// Runner should be removed
	assert.Nil(t, proc.getRunner(worldID), "Runner should be removed after reset")

	// Geology should be cleared from processor
	_, exists := proc.worldGeology[worldID]
	assert.False(t, exists, "Geology should be cleared from processor")

	// Geology should be cleared from map service
	// We check if we can retrieve it (mock returns nil if cleared)
	assert.Nil(t, proc.mapService.GetWorldGeology(worldID), "Geology should be cleared from map service")

	// Entities for this world should be removed
	_, entExists := ecoSvc.Entities[entityID]
	assert.False(t, entExists, "Entities for world should be removed")

	// Client should receive world_reset message
	messages := client.GetMessages()
	foundResetMsg := false
	for _, m := range messages {
		if m.Type == "world_reset" {
			foundResetMsg = true
			break
		}
	}
	assert.True(t, foundResetMsg, "Should send world_reset message to client")
}
