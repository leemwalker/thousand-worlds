package processor

import (
	"context"
	"testing"
	"time"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/game/constants"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/game/services/look"
	"tw-backend/internal/player"
	"tw-backend/internal/repository"
	"tw-backend/internal/world/interview"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockInterviewRepository for testing
type MockInterviewRepository struct{}

func (m *MockInterviewRepository) GetConfigurationByWorldID(ctx context.Context, worldID uuid.UUID) (*interview.WorldConfiguration, error) {
	return nil, nil // Return nil config for basic testing
}
func (m *MockInterviewRepository) GetConfigurationByUserID(ctx context.Context, userID uuid.UUID) (*interview.WorldConfiguration, error) {
	return nil, nil
}
func (m *MockInterviewRepository) CreateInterview(ctx context.Context, userID uuid.UUID) (*interview.Interview, error) {
	return &interview.Interview{ID: uuid.New(), UserID: userID}, nil
}
func (m *MockInterviewRepository) GetActiveInterview(ctx context.Context, userID uuid.UUID) (*interview.Interview, error) {
	return nil, nil
}
func (m *MockInterviewRepository) GetInterview(ctx context.Context, userID uuid.UUID) (*interview.Interview, error) {
	return nil, nil
}
func (m *MockInterviewRepository) GetAnswers(ctx context.Context, interviewID uuid.UUID) ([]interview.Answer, error) {
	return []interview.Answer{}, nil
}
func (m *MockInterviewRepository) UpdateInterview(ctx context.Context, interview *interview.Interview) error {
	return nil
}
func (m *MockInterviewRepository) CreateConfiguration(ctx context.Context, config *interview.WorldConfiguration) error {
	return nil
}
func (m *MockInterviewRepository) SaveConfiguration(ctx context.Context, config *interview.WorldConfiguration) error {
	return nil
}
func (m *MockInterviewRepository) IsWorldNameTaken(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (m *MockInterviewRepository) UpdateInterviewStatus(ctx context.Context, id uuid.UUID, status interview.Status) error {
	return nil
}
func (m *MockInterviewRepository) UpdateQuestionIndex(ctx context.Context, id uuid.UUID, index int) error {
	return nil
}
func (m *MockInterviewRepository) SaveAnswer(ctx context.Context, interviewID uuid.UUID, index int, text string) error {
	return nil
}

// MockWorldRepository for testing
type MockWorldRepository struct {
	worlds map[uuid.UUID]*repository.World
}

func NewMockWorldRepository() *MockWorldRepository {
	return &MockWorldRepository{
		worlds: make(map[uuid.UUID]*repository.World),
	}
}

func (m *MockWorldRepository) CreateWorld(ctx context.Context, world *repository.World) error {
	m.worlds[world.ID] = world
	return nil
}

func (m *MockWorldRepository) GetWorld(ctx context.Context, worldID uuid.UUID) (*repository.World, error) {
	if w, ok := m.worlds[worldID]; ok {
		return w, nil
	}
	return nil, assert.AnError
}

func (m *MockWorldRepository) ListWorlds(ctx context.Context) ([]repository.World, error) {
	var worlds []repository.World
	for _, w := range m.worlds {
		worlds = append(worlds, *w)
	}
	return worlds, nil
}

func (m *MockWorldRepository) GetWorldsByOwner(ctx context.Context, ownerID uuid.UUID) ([]repository.World, error) {
	var worlds []repository.World
	for _, w := range m.worlds {
		if w.OwnerID == ownerID {
			worlds = append(worlds, *w)
		}
	}
	return worlds, nil
}

func (m *MockWorldRepository) UpdateWorld(ctx context.Context, world *repository.World) error {
	m.worlds[world.ID] = world
	return nil
}

func (m *MockWorldRepository) DeleteWorld(ctx context.Context, worldID uuid.UUID) error {
	delete(m.worlds, worldID)
	return nil
}

// Mock client for testing
type mockClient struct {
	CharacterID  uuid.UUID
	UserID       uuid.UUID
	Username     string
	WorldID      uuid.UUID
	messages     []websocket.GameMessageData
	stateUpdates int
}

func (m *mockClient) GetCharacterID() uuid.UUID {
	return m.CharacterID
}

func (m *mockClient) GetWorldID() uuid.UUID {
	return m.WorldID
}

func (m *mockClient) SetWorldID(id uuid.UUID) {
	m.WorldID = id
}

func (m *mockClient) GetUserID() uuid.UUID {
	return m.UserID
}

func (m *mockClient) GetUsername() string {
	return m.Username
}

func (m *mockClient) SendGameMessage(msgType, text string, metadata map[string]interface{}) {
	m.messages = append(m.messages, websocket.GameMessageData{
		Type:     msgType,
		Text:     text,
		Metadata: metadata, // Capture metadata for verification
	})
}

func (m *mockClient) SendStateUpdate(state *websocket.StateUpdateData) {
	m.stateUpdates++
}

func (m *mockClient) GetLastTellSender() string {
	return ""
}

func (m *mockClient) SetLastTellSender(username string) {
	// No-op for basic mock
}

func (m *mockClient) ClearLastTellSender() {
	// No-op for basic mock
}

func (m *mockClient) SetCharacterID(id uuid.UUID) {
	m.CharacterID = id
}

func newMockClient() *mockClient {
	return &mockClient{
		CharacterID: uuid.New(),
		UserID:      uuid.New(),
		Username:    "TestUser",
		messages:    make([]websocket.GameMessageData, 0),
	}
}

// TestHandleWatcher tests the watcher command
func TestHandleWatcher(t *testing.T) {
	processor, client, authRepo, _ := setupTest(t)

	// Ensure character is in the Lobby so "watcher" command is valid
	char, _ := authRepo.GetCharacter(context.Background(), client.GetCharacterID())
	char.WorldID = constants.LobbyWorldID
	// We need to update the character in the mock repo
	// Since MockRepository.UpdateCharacter doesn't exist or we can just re-create
	// Let's just create a new character correctly or use a helper if available.
	// Looking at setupTest, it calls CreateCharacter.
	// Let's just overwrite it in the map if we can, or calling CreateCharacter with same ID might work if distinct?
	// Actually, let's just use a new client/character setup if needed, OR:
	// The simplest way is to fetch, modify, and rely on pointer if it's stored by pointer?
	// MockRepository likely stores by value or pointer.
	// Let's assume we can just overwrite it.
	authRepo.CreateCharacter(context.Background(), char) // Overwrite

	// Create a target world UUID
	worldID := uuid.New()
	target := worldID.String()

	cmd := &websocket.CommandData{
		Action: "watcher",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "system", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "watcher")

	// Verify SetWorldID was called with the correct WorldID
	assert.Equal(t, worldID, client.WorldID, "Client WorldID should be updated to target world")
}

func setupTest(t *testing.T) (*GameProcessor, *mockClient, *auth.MockRepository, *MockWorldRepository) {
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository()

	// Add Lobby world to mock repository for movement tests
	lobbyWorld := &repository.World{
		ID:   constants.LobbyWorldID,
		Name: "Lobby",
	}
	mockWorldRepo.worlds[lobbyWorld.ID] = lobbyWorld

	interviewRepo := &MockInterviewRepository{}
	// New LookService requires EntityService
	entityService := entity.NewService()
	lookService := look.NewLookService(mockWorldRepo, nil, entityService, interviewRepo, mockAuthRepo, nil, nil)
	interviewService := interview.NewServiceWithRepository(nil, interviewRepo, mockWorldRepo)
	spatialService := player.NewSpatialService(mockAuthRepo, mockWorldRepo, nil)

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, lookService, entityService, interviewService, spatialService, nil, nil, nil, nil, nil, nil, nil, nil)

	// Create and set up the hub
	hub := websocket.NewHub(proc)
	proc.SetHub(hub)

	client := newMockClient()

	// Create a character for the client in the mock repo
	// Use Lobby world so movement tests work
	char := &auth.Character{
		CharacterID: client.CharacterID,
		UserID:      uuid.New(),
		WorldID:     constants.LobbyWorldID,
		Name:        "TestChar",
		CreatedAt:   time.Now(),
		PositionX:   5.0, // Lobby center
		PositionY:   5.0, // Lobby center
	}
	err := mockAuthRepo.CreateCharacter(context.Background(), char)
	require.NoError(t, err)

	return proc, client, mockAuthRepo, mockWorldRepo
}

// TestCardinalDirections tests the basic cardinal direction commands in lobby
func TestCardinalDirections(t *testing.T) {
	// Only n, s, e, w are supported in the lobby (10m x 1000m space)
	directions := []struct {
		action   string
		expected string
	}{
		{"north", "north"},
		{"n", "north"},
		{"east", "east"},
		{"e", "east"},
		{"south", "south"},
		{"s", "south"},
		{"west", "west"},
		{"w", "west"},
	}

	for _, tt := range directions {
		t.Run(tt.action, func(t *testing.T) {
			processor, client, _, _ := setupTest(t)
			cmd := &websocket.CommandData{
				Action: tt.action,
			}

			err := processor.ProcessCommand(context.Background(), client, cmd)

			require.NoError(t, err)
			require.GreaterOrEqual(t, len(client.messages), 1)
			assert.Contains(t, client.messages[0].Text, tt.expected)
		})
	}
}

// TestHandleOpen tests the open command - not supported in lobby
func TestHandleOpen_Door(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "door"
	cmd := &websocket.CommandData{
		Action: "open",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "You open the door")
}

func TestHandleOpen_Container(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "chest"
	cmd := &websocket.CommandData{
		Action: "open",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "You open the chest")
}

func TestHandleOpen_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "open",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// Expect error due to missing target
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target required")
}

// TestHandleEnter tests the enter command in lobby
func TestHandleEnter_Portal(t *testing.T) {
	processor, client, mockAuth, worldRepo := setupTest(t)

	// Add world to mock repository
	portalWorld := &repository.World{
		ID:   uuid.New(),
		Name: "TestWorld",
	}
	worldRepo.worlds[portalWorld.ID] = portalWorld

	// Create dummy current world for portal location calculation
	currentWorld := &repository.World{
		BoundsMin: &repository.Vector3{X: 0, Y: 0, Z: 0},
		BoundsMax: &repository.Vector3{X: 10, Y: 10, Z: 0},
	}

	// Calculate portal location and move character there
	// We need a spatial service instance to calculate the deterministic location
	spatialSvc := player.NewSpatialService(auth.NewMockRepository(), worldRepo, nil)
	px, py := spatialSvc.GetPortalLocation(currentWorld, portalWorld.ID)

	// Move character
	char, _ := mockAuth.GetCharacter(context.Background(), client.GetCharacterID())
	char.PositionX = px
	char.PositionY = py
	mockAuth.UpdateCharacter(context.Background(), char)

	target := "TestWorld"
	cmd := &websocket.CommandData{
		Action: "enter",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	// Lobby's handleLobbyEnter sends trigger_entry_options, not transition
	assert.Equal(t, "trigger_entry_options", client.messages[0].Type)
}

func TestHandleEnter_Doorway(t *testing.T) {
	processor, client, _, _ := setupTest(t)

	target := "doorway"
	cmd := &websocket.CommandData{
		Action: "enter",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	// Should get "no portal to doorway here" message
	assert.Contains(t, client.messages[0].Text, "doorway")
}

func TestHandleEnter_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "enter",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// handleEnter returns nil but sends error message
	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "Enter what")
}

// TestHandle Say tests the say command
func TestHandleSay_BroadcastsMessage(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	// Mock the Hub dependency
	hub := websocket.NewHub(processor)
	processor.SetHub(hub)

	message := "Hello everyone!"
	cmd := &websocket.CommandData{
		Action:  "say",
		Message: &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "speech_self", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, message)
}

func TestHandleSay_NoMessage(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	hub := websocket.NewHub(processor)
	processor.SetHub(hub)

	cmd := &websocket.CommandData{
		Action: "say",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// No error returned, but client should receive error message
	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "error", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "What do you want to say")
}

// TestHandleWhisper tests the whisper command - not supported in lobby
func TestHandleWhisper_ToNearbyPlayer(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	recipient := "Bob"
	message := "psst, secret"
	cmd := &websocket.CommandData{
		Action:    "whisper",
		Recipient: &recipient,
		Message:   &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// Whisper is supported now
	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	// It sends "whisper" type
	assert.Equal(t, "whisper", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "You whisper to Bob")
}

func TestHandleWhisper_NoRecipient(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	message := "secret"
	cmd := &websocket.CommandData{
		Action:  "whisper",
		Message: &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "recipient required")
}

func TestHandleWhisper_NoMessage(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	recipient := "Bob"
	cmd := &websocket.CommandData{
		Action:    "whisper",
		Recipient: &recipient,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "message required")
}

// TestHandleTell tests the tell command
func TestHandleTell_OnlinePlayer(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	recipient := "Alice"
	message := "Are you there?"
	cmd := &websocket.CommandData{
		Action:    "tell",
		Recipient: &recipient,
		Message:   &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// Player not found, should send error message to client
	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "error", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "not online")
}

func TestHandleTell_NoRecipient(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	message := "hello"
	cmd := &websocket.CommandData{
		Action:  "tell",
		Message: &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// No error returned, but client receives error message
	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "error", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "Tell whom")
}

func TestHandleTell_NoMessage(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	recipient := "Alice"
	cmd := &websocket.CommandData{
		Action:    "tell",
		Recipient: &recipient,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// No error returned, but client receives error message
	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "error", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "What do you want to say")
}

// TestHandleWho tests the who command
func TestHandleWho_ListsPlayers(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	// Mock the Hub dependency
	hub := websocket.NewHub(processor)
	processor.SetHub(hub)

	cmd := &websocket.CommandData{
		Action: "who",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "player_list", client.messages[0].Type)
	// Message should contain formatted player list
}

func TestHandleWho_NoHub(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	// Explicitly clear the Hub to test error condition
	processor.SetHub(nil)

	cmd := &websocket.CommandData{
		Action: "who",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "game server not fully initialized")
}

// TestHandleHelp tests the help command
func TestHandleHelp(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "help",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "system", client.messages[0].Type)
	assert.Equal(t, "system", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "Available Commands")
}

// TestHandleLook tests look command (should still work)
func TestHandleLook_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "look",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(client.messages), 1)
	assert.Equal(t, "area_description", client.messages[0].Type)
}

func TestHandleLook_WithTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "sword"
	cmd := &websocket.CommandData{
		Action: "look",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// In lobby, looking at unknown items returns "don't see that"
	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "don't see")
}

// TestHandleGet tests the get command
func TestHandleGet(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "sword"
	cmd := &websocket.CommandData{
		Action: "get",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// With nil worldEntityService, falls back to legacy behavior
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(client.messages), 1)
	// Legacy fallback sends action message "You pick up the sword."
	assert.Equal(t, "action", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "pick up")
}

func TestHandleGet_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "get",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	// handleGetObject sends error to client, not error return
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(client.messages), 1)
	assert.Equal(t, "error", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "Get what?")
}

// TestHandleDrop tests the drop command - not supported in lobby
func TestHandleDrop(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "sword"
	cmd := &websocket.CommandData{
		Action: "drop",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "You drop the sword")
}

func TestHandleDrop_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "drop",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target item required")
}

// TestHandleAttack tests the attack command - not supported in lobby
func TestHandleAttack(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "goblin"
	cmd := &websocket.CommandData{
		Action: "attack",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "combat", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "You attack")
}

func TestHandleAttack_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "attack",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target required")
}

// TestHandleTalk tests the talk command - not supported in lobby
func TestHandleTalk(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "merchant"
	cmd := &websocket.CommandData{
		Action: "talk",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "dialogue", client.messages[0].Type)
}

func TestHandleTalk_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "talk",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target required")
}

// TestHandleInventory tests the inventory command - not supported in lobby
func TestHandleInventory(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "inventory",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	// Inventory calls sendStateUpdate which sends "state_update"
	// It doesn't strictly send a message back, but update.
	// We check for state update count
	assert.Equal(t, 1, client.stateUpdates)
}

// TestHandleCraft tests the craft command - not supported in lobby
func TestHandleCraft(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "sword"
	cmd := &websocket.CommandData{
		Action: "craft",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "crafted sword")
}

func TestHandleCraft_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "craft",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "item required")
}

// TestHandleUse tests the use command - not supported in lobby
func TestHandleUse(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	target := "potion"
	cmd := &websocket.CommandData{
		Action: "use",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "use the potion")
}

func TestHandleUse_NoTarget(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "use",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "item required")
}

// TestInvalidCommand tests unknown commands
func TestInvalidCommand(t *testing.T) {
	processor, client, _, _ := setupTest(t)
	cmd := &websocket.CommandData{
		Action: "invalid_action",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

// TestNoCharacter tests command with no character
func TestNoCharacter(t *testing.T) {
	processor, _, _, _ := setupTest(t)
	client := &mockClient{
		CharacterID: uuid.Nil,
		messages:    make([]websocket.GameMessageData, 0),
	}
	cmd := &websocket.CommandData{
		Action: "north",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active character")
}
