package processor

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/ecosystem/state"
)

func TestHandleEcosystem_Status(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	// Pre-populate ecosystem
	ecoSvc.Spawner.CreateEntity(state.SpeciesRabbit, 1)

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	client := &mockClient{
		UserID:      uuid.New(),
		CharacterID: uuid.New(),
	}

	// Execute Status
	target := "status"
	cmd := &websocket.CommandData{
		Action: "ecosystem",
		Target: &target,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "Ecosystem Status")
}

func TestHandleEcosystem_Spawn(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	// Create user character so we have a location
	charID := uuid.New()
	userID := uuid.New()
	mockAuthRepo.CreateCharacter(context.Background(), &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     uuid.New(),
		PositionX:   100,
		PositionY:   200,
	})

	client := &mockClient{
		UserID:      userID,
		CharacterID: charID,
	}

	// Execute Spawn
	target := "spawn"
	msg := "wolf"
	cmd := &websocket.CommandData{
		Action:  "ecosystem",
		Target:  &target,
		Message: &msg,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify entity created
	require.NotEmpty(t, ecoSvc.Entities)
	var spawned *state.LivingEntityState
	for _, e := range ecoSvc.Entities {
		spawned = e
		break
	}
	assert.Equal(t, state.SpeciesWolf, spawned.Species)
	assert.Equal(t, 100.0, spawned.PositionX)
	assert.Equal(t, 200.0, spawned.PositionY)

	// Verify message
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "Spawned wolf")
}

func TestHandleEcosystem_WithParsing(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	// Create user character so we have a location
	charID := uuid.New()
	userID := uuid.New()
	mockAuthRepo.CreateCharacter(context.Background(), &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     uuid.New(),
		PositionX:   50,
		PositionY:   50,
	})

	client := &mockClient{
		UserID:      userID,
		CharacterID: charID,
	}

	// Execute via Text (Triggering ParseText)
	cmd := &websocket.CommandData{
		Text: "ecosystem spawn rabbit",
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify entity created
	require.NotEmpty(t, ecoSvc.Entities)
	var spawned *state.LivingEntityState
	for _, e := range ecoSvc.Entities {
		spawned = e
		break
	}
	assert.Equal(t, state.SpeciesRabbit, spawned.Species)

	// Verify message
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "Spawned rabbit")
}

func TestHandleEcosystem_Alias(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	client := &mockClient{
		UserID:      uuid.New(),
		CharacterID: uuid.New(),
	}

	// Execute via Alias "eco"
	cmd := &websocket.CommandData{
		Text: "eco status",
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "Ecosystem Status")
}

func TestHandleEcosystem_Log(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	client := &mockClient{
		UserID:      uuid.New(),
		CharacterID: uuid.New(),
	}

	// Create entity with logs
	ent := ecoSvc.Spawner.CreateEntity(state.SpeciesWolf, 1)
	ent.AddLog("Sleep", "Low energy")
	ecoSvc.Entities[ent.EntityID] = ent

	idStr := ent.EntityID.String()

	// Execute Log with partial ID
	target := "log"
	msg := idStr[:5] // Partial ID
	cmd := &websocket.CommandData{
		Action:  "ecosystem",
		Target:  &target,
		Message: &msg,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "Decision Logs for")
	assert.Contains(t, lastMsg.Text, "Sleep: Low energy")
}

func TestHandleEcosystem_Breed(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	client := &mockClient{
		UserID:      uuid.New(),
		CharacterID: uuid.New(),
	}

	// Create two parent entities
	parent1 := ecoSvc.Spawner.CreateEntity(state.SpeciesRabbit, 1)
	parent1.WorldID = uuid.New()
	parent1.PositionX = 10
	parent1.PositionY = 10
	ecoSvc.Entities[parent1.EntityID] = parent1

	parent2 := ecoSvc.Spawner.CreateEntity(state.SpeciesRabbit, 1)
	parent2.WorldID = parent1.WorldID
	parent2.PositionX = 12
	parent2.PositionY = 10
	ecoSvc.Entities[parent2.EntityID] = parent2

	// Execute breed
	target := "breed"
	msg := parent1.EntityID.String()[:5] + " " + parent2.EntityID.String()[:5]
	cmd := &websocket.CommandData{
		Action:  "ecosystem",
		Target:  &target,
		Message: &msg,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify offspring created
	assert.Equal(t, 3, len(ecoSvc.Entities))

	// Verify message
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "Bred rabbit!")
	assert.Contains(t, lastMsg.Text, "Generation 2")
}

func TestHandleEcosystem_Lineage(t *testing.T) {
	// Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()
	ecoSvc := ecosystem.NewService(time.Now().Unix())

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, nil, nil, nil, nil, ecoSvc, nil, nil, nil, nil, nil, nil)

	client := &mockClient{
		UserID:      uuid.New(),
		CharacterID: uuid.New(),
	}

	// Create parent and child
	parent1 := ecoSvc.Spawner.CreateEntity(state.SpeciesWolf, 1)
	parent2 := ecoSvc.Spawner.CreateEntity(state.SpeciesWolf, 1)
	ecoSvc.Entities[parent1.EntityID] = parent1
	ecoSvc.Entities[parent2.EntityID] = parent2

	// Reproduce
	child, err := ecoSvc.GetEvolutionManager().Reproduce(parent1, parent2)
	require.NoError(t, err)
	ecoSvc.Entities[child.EntityID] = child

	// Execute lineage on child
	target := "lineage"
	msg := child.EntityID.String()[:5]
	cmd := &websocket.CommandData{
		Action:  "ecosystem",
		Target:  &target,
		Message: &msg,
	}

	err = proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// Verify
	require.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "Lineage for")
	assert.Contains(t, lastMsg.Text, "Generation: 2")
	assert.Contains(t, lastMsg.Text, "Parent 1:")
	assert.Contains(t, lastMsg.Text, "Parent 2:")
}
