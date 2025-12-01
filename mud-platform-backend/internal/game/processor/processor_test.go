package processor

import (
	"context"
	"testing"

	"mud-platform-backend/cmd/game-server/websocket"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock client for testing
type mockClient struct {
	CharacterID  uuid.UUID
	messages     []websocket.GameMessageData
	stateUpdates int
}

func (m *mockClient) GetCharacterID() uuid.UUID {
	return m.CharacterID
}

func (m *mockClient) SendGameMessage(msgType, text string, metadata map[string]interface{}) {
	m.messages = append(m.messages, websocket.GameMessageData{
		Type: msgType,
		Text: text,
	})
}

func (m *mockClient) SendStateUpdate(state *websocket.StateUpdateData) {
	m.stateUpdates++
}

func newMockClient() *mockClient {
	return &mockClient{
		CharacterID: uuid.New(),
		messages:    make([]websocket.GameMessageData, 0),
	}
}

// TestCardinalDirections tests all 10 cardinal direction commands
func TestCardinalDirections(t *testing.T) {
	directions := []struct {
		action   string
		expected string
	}{
		{"north", "north"},
		{"n", "north"},
		{"northeast", "northeast"},
		{"ne", "northeast"},
		{"east", "east"},
		{"e", "east"},
		{"southeast", "southeast"},
		{"se", "southeast"},
		{"south", "south"},
		{"s", "south"},
		{"southwest", "southwest"},
		{"sw", "southwest"},
		{"west", "west"},
		{"w", "west"},
		{"northwest", "northwest"},
		{"nw", "northwest"},
		{"up", "up"},
		{"u", "up"},
		{"down", "down"},
		{"d", "down"},
	}

	for _, tt := range directions {
		t.Run(tt.action, func(t *testing.T) {
			processor := NewGameProcessor()
			client := newMockClient()
			cmd := &websocket.CommandData{
				Action: tt.action,
			}

			err := processor.ProcessCommand(context.Background(), client, cmd)

			require.NoError(t, err)
			require.Len(t, client.messages, 1)
			assert.Contains(t, client.messages[0].Text, tt.expected)
			assert.Equal(t, 1, client.stateUpdates)
		})
	}
}

// TestHandleOpen tests the open command
func TestHandleOpen_Door(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "door"
	cmd := &websocket.CommandData{
		Action: "open",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "open")
	assert.Contains(t, client.messages[0].Text, target)
}

func TestHandleOpen_Container(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "chest"
	cmd := &websocket.CommandData{
		Action: "open",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "chest")
}

func TestHandleOpen_NoTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "open",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target required")
}

// TestHandleEnter tests the enter command
func TestHandleEnter_Portal(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "portal"
	cmd := &websocket.CommandData{
		Action: "enter",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "enter")
	assert.Contains(t, client.messages[0].Text, target)
}

func TestHandleEnter_Doorway(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "doorway"
	cmd := &websocket.CommandData{
		Action: "enter",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	assert.Contains(t, client.messages[0].Text, "doorway")
}

func TestHandleEnter_NoTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "enter",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target required")
}

// TestHandleSay tests the say command
func TestHandleSay_BroadcastsMessage(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	message := "Hello everyone!"
	cmd := &websocket.CommandData{
		Action:  "say",
		Message: &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "speech", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, message)
}

func TestHandleSay_NoMessage(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "say",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "message required")
}

// TestHandleWhisper tests the whisper command
func TestHandleWhisper_ToNearbyPlayer(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	recipient := "Bob"
	message := "psst, secret"
	cmd := &websocket.CommandData{
		Action:    "whisper",
		Recipient: &recipient,
		Message:   &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "whisper", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, recipient)
}

func TestHandleWhisper_NoRecipient(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
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
	processor := NewGameProcessor()
	client := newMockClient()
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
	processor := NewGameProcessor()
	client := newMockClient()
	recipient := "Alice"
	message := "Are you there?"
	cmd := &websocket.CommandData{
		Action:    "tell",
		Recipient: &recipient,
		Message:   &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "tell", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, recipient)
}

func TestHandleTell_NoRecipient(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	message := "hello"
	cmd := &websocket.CommandData{
		Action:  "tell",
		Message: &message,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "recipient required")
}

func TestHandleTell_NoMessage(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	recipient := "Alice"
	cmd := &websocket.CommandData{
		Action:    "tell",
		Recipient: &recipient,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "message required")
}

// TestHandleWho tests the who command
func TestHandleWho_ListsPlayers(t *testing.T) {
	processor := NewGameProcessor()
	// Mock the Hub dependency
	hub := websocket.NewHub(processor)
	processor.SetHub(hub)

	client := newMockClient()
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
	processor := NewGameProcessor()
	// No Hub set

	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "who",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "game server not fully initialized")
}

// TestHandleHelp tests the help command
func TestHandleHelp(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "help",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "system", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, "Available Commands")
}

// TestHandleLook tests look command (should still work)
func TestHandleLook_NoTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "look",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "area_description", client.messages[0].Type)
}

func TestHandleLook_WithTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "sword"
	cmd := &websocket.CommandData{
		Action: "look",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, target)
}

// TestHandleTake tests the take command
func TestHandleTake(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "sword"
	cmd := &websocket.CommandData{
		Action: "take",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "item_acquired", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, target)
	assert.Equal(t, 1, client.stateUpdates)
}

func TestHandleTake_NoTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "take",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target item required")
}

// TestHandleDrop tests the drop command
func TestHandleDrop(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "sword"
	cmd := &websocket.CommandData{
		Action: "drop",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "drop")
	assert.Contains(t, client.messages[0].Text, target)
	assert.Equal(t, 1, client.stateUpdates)
}

func TestHandleDrop_NoTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "drop",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target item required")
}

// TestHandleAttack tests the attack command
func TestHandleAttack(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "goblin"
	cmd := &websocket.CommandData{
		Action: "attack",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "combat", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, target)
	assert.Equal(t, 1, client.stateUpdates)
}

func TestHandleAttack_NoTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "attack",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target required")
}

// TestHandleTalk tests the talk command
func TestHandleTalk(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
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
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "talk",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target required")
}

// TestHandleInventory tests the inventory command
func TestHandleInventory(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "inventory",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	assert.Equal(t, 1, client.stateUpdates)
}

// TestHandleCraft tests the craft command
func TestHandleCraft(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "sword"
	cmd := &websocket.CommandData{
		Action: "craft",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Equal(t, "crafting_success", client.messages[0].Type)
	assert.Contains(t, client.messages[0].Text, target)
	assert.Equal(t, 1, client.stateUpdates)
}

func TestHandleCraft_NoTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "craft",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "item required")
}

// TestHandleUse tests the use command
func TestHandleUse(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	target := "potion"
	cmd := &websocket.CommandData{
		Action: "use",
		Target: &target,
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.NoError(t, err)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Text, "use")
	assert.Contains(t, client.messages[0].Text, target)
	assert.Equal(t, 1, client.stateUpdates)
}

func TestHandleUse_NoTarget(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "use",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "item required")
}

// TestInvalidCommand tests unknown commands
func TestInvalidCommand(t *testing.T) {
	processor := NewGameProcessor()
	client := newMockClient()
	cmd := &websocket.CommandData{
		Action: "invalid_action",
	}

	err := processor.ProcessCommand(context.Background(), client, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

// TestNoCharacter tests command with no character
func TestNoCharacter(t *testing.T) {
	processor := NewGameProcessor()
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
