package integration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/game/processor"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/game/services/look"
	"tw-backend/internal/player"
	"tw-backend/internal/repository"
	"tw-backend/internal/world/interview"
)

// StatefulMockWorldRepository for E2E testing
type StatefulMockWorldRepository struct {
	worlds map[uuid.UUID]*repository.World
}

func NewStatefulMockWorldRepository() *StatefulMockWorldRepository {
	return &StatefulMockWorldRepository{
		worlds: make(map[uuid.UUID]*repository.World),
	}
}

func (m *StatefulMockWorldRepository) CreateWorld(ctx context.Context, world *repository.World) error {
	m.worlds[world.ID] = world
	return nil
}

func (m *StatefulMockWorldRepository) GetWorld(ctx context.Context, worldID uuid.UUID) (*repository.World, error) {
	if w, ok := m.worlds[worldID]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("world not found")
}

func (m *StatefulMockWorldRepository) ListWorlds(ctx context.Context) ([]repository.World, error) {
	var list []repository.World
	for _, w := range m.worlds {
		list = append(list, *w)
	}
	return list, nil
}

func (m *StatefulMockWorldRepository) UpdateWorld(ctx context.Context, world *repository.World) error {
	m.worlds[world.ID] = world
	return nil
}

func (m *StatefulMockWorldRepository) DeleteWorld(ctx context.Context, worldID uuid.UUID) error {
	delete(m.worlds, worldID)
	return nil
}

func (m *StatefulMockWorldRepository) GetWorldsByOwner(ctx context.Context, ownerID uuid.UUID) ([]repository.World, error) {
	var list []repository.World
	for _, w := range m.worlds {
		if w.OwnerID == ownerID {
			list = append(list, *w)
		}
	}
	return list, nil
}

// MockLLM local definition
type MockLLM struct {
	GenerateFunc func(prompt string) (string, error)
}

func (m *MockLLM) Generate(prompt string) (string, error) {
	return m.GenerateFunc(prompt)
}

// TestGameClient implementation
type TestGameClient struct {
	UserID         uuid.UUID
	CharacterID    uuid.UUID
	WorldID        uuid.UUID
	Username       string
	LastTellSender string
	Messages       []struct {
		Type string
		Text string
		Data interface{}
	}
}

func (c *TestGameClient) GetUserID() uuid.UUID                            { return c.UserID }
func (c *TestGameClient) GetCharacterID() uuid.UUID                       { return c.CharacterID }
func (c *TestGameClient) GetWorldID() uuid.UUID                           { return c.WorldID }
func (c *TestGameClient) GetUsername() string                             { return c.Username }
func (c *TestGameClient) GetLastTellSender() string                       { return c.LastTellSender }
func (c *TestGameClient) SetLastTellSender(s string)                      { c.LastTellSender = s }
func (c *TestGameClient) ClearLastTellSender()                            { c.LastTellSender = "" }
func (c *TestGameClient) SendStateUpdate(data *websocket.StateUpdateData) {}
func (c *TestGameClient) SetCharacterID(id uuid.UUID)                     { c.CharacterID = id }
func (c *TestGameClient) SetWorldID(id uuid.UUID)                         { c.WorldID = id }

func (c *TestGameClient) SendGameMessage(msgType, text string, data map[string]interface{}) {
	c.Messages = append(c.Messages, struct {
		Type string
		Text string
		Data interface{}
	}{Type: msgType, Text: text, Data: data})
}

func TestWatcherEntryE2E(t *testing.T) {
	ctx := context.Background()

	// 1. Setup Dependencies
	authRepo := auth.NewMockRepository()
	worldRepo := NewStatefulMockWorldRepository()
	interviewRepo := interview.NewMockRepository()

	// Mock LLM for interview
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			if strings.Contains(prompt, "You are a data extraction assistant") {
				return `{"theme": "Cyberpunk", "worldName": "Neon City"}`, nil
			}
			return "Next question?", nil
		},
	}

	interviewService := interview.NewServiceWithRepository(mockLLM, interviewRepo, worldRepo)

	// Initialize look service and processor
	entitySvc := entity.NewService()
	lookService := look.NewLookService(worldRepo, nil, entitySvc, interviewRepo) // Use interviewRepo (repository interface) not service
	spatialSvc := player.NewSpatialService(authRepo, worldRepo)
	gameProcessor := processor.NewGameProcessor(authRepo, worldRepo, lookService, entitySvc, interviewService, spatialSvc, nil)

	// Assign to `proc` if used later, or rename later usages to gameProcessor
	proc := gameProcessor

	// 2. User creates world "Neon City"
	userID := uuid.New()
	creatorCharID := uuid.New()
	userClient := &TestGameClient{
		UserID:      userID,
		CharacterID: creatorCharID,
		Username:    "Creator",
		WorldID:     uuid.Nil, // Lobby
	}

	// Create character in authRepo
	err := authRepo.CreateCharacter(ctx, &auth.Character{
		CharacterID: creatorCharID,
		UserID:      userID,
		WorldID:     uuid.Nil,
		Name:        "Creator",
	})
	require.NoError(t, err)

	// 2a. Start Interview (tell statue create world)
	cmd := &websocket.CommandData{Action: "tell", Recipient: ptr("statue"), Message: ptr("create world")}
	err = proc.ProcessCommand(ctx, userClient, cmd)
	require.NoError(t, err)

	// Verify "Thinking" and "Emote" and "Question" messages
	// Note: Messages are appended.
	// 0: Thinking emote
	// 1: Eyes glow emote
	// 2: Question tell
	require.GreaterOrEqual(t, len(userClient.Messages), 3)
	assert.Equal(t, "emote", userClient.Messages[0].Type) // Thinking
	assert.Equal(t, "emote", userClient.Messages[1].Type) // Statue eyes glow
	assert.Equal(t, "tell", userClient.Messages[2].Type)  // Question

	// 2b. Answer questions
	// We need to answer questions until completion.
	// We don't know exactly how many loops, but mocked interview has AllTopics.
	// NOTE: We need to access interview.AllTopics. It is exported.

	for i := 0; i < len(interview.AllTopics); i++ {
		// Clear messages to track new ones
		userClient.Messages = nil

		replyMsg := "Answer"
		if i == len(interview.AllTopics)-1 {
			replyMsg = "Neon City"
		}

		cmd := &websocket.CommandData{Action: "reply", Message: &replyMsg}
		err := proc.ProcessCommand(ctx, userClient, cmd)
		require.NoError(t, err)

		// If it's the last question (Neon City), implementation might send "Here is the vision..." + "Review phase"
		// If it's intermediate, it sends next question.
		// We just process until we are ready for confirmation.
	}

	// 2c. Confirm review
	userClient.Messages = nil
	confirmMsg := "yes"
	cmd = &websocket.CommandData{Action: "reply", Message: &confirmMsg}
	err = proc.ProcessCommand(ctx, userClient, cmd)
	require.NoError(t, err)

	// Verify creation
	foundCreationMsg := false
	for _, m := range userClient.Messages {
		if m.Type == "tell" && strings.Contains(m.Text, "forged") {
			foundCreationMsg = true
		}
	}
	assert.True(t, foundCreationMsg, "Should find creation message")

	// 3. User enters world
	// 3a. Use logic 'enter Neon City'
	userClient.Messages = nil
	target := "Neon City"
	cmd = &websocket.CommandData{Action: "enter", Target: &target}
	err = proc.ProcessCommand(ctx, userClient, cmd)
	require.NoError(t, err)

	// Verify trigger_entry_options
	require.NotEmpty(t, userClient.Messages)
	assert.Equal(t, "trigger_entry_options", userClient.Messages[0].Type)

	// 4. Mimic joining as watcher
	worldList, _ := worldRepo.ListWorlds(ctx)
	require.NotEmpty(t, worldList)
	worldID := worldList[0].ID

	watcherID := uuid.New()
	watcherClient := &TestGameClient{
		UserID:      userID,
		CharacterID: watcherID,
		WorldID:     worldID, // In the new world
		Username:    "Creator",
	}

	err = authRepo.CreateCharacter(ctx, &auth.Character{
		CharacterID: watcherID,
		UserID:      userID,
		WorldID:     worldID,
		Name:        "Watcher",
	})
	require.NoError(t, err)

	// 5. Watcher looks around
	cmd = &websocket.CommandData{Action: "look"}
	err = proc.ProcessCommand(ctx, watcherClient, cmd)
	require.NoError(t, err)

	// Verify description
	require.NotEmpty(t, watcherClient.Messages)
	// Should receive area_description
	descMsg := watcherClient.Messages[0]
	assert.Equal(t, "area_description", descMsg.Type)
	assert.NotEmpty(t, descMsg.Text)
	fmt.Printf("Watcher sees: %s\n", descMsg.Text)

	// Check for proper description (not error, not fallback if possible)
	assert.NotContains(t, descMsg.Text, "mysterious place")
}

func ptr(s string) *string {
	return &s
}
