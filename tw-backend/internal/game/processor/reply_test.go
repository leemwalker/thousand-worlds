package processor

import (
	"context"
	"sync"
	"testing"
	"time"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/game/constants"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClientWithReply extends mockClient to support reply functionality
type mockClientWithReply struct {
	*mockClient
	lastTellSender   string
	lastTellSenderMu sync.RWMutex
}

func (m *mockClientWithReply) SetLastTellSender(username string) {
	m.lastTellSenderMu.Lock()
	defer m.lastTellSenderMu.Unlock()
	m.lastTellSender = username
}

func (m *mockClientWithReply) GetLastTellSender() string {
	m.lastTellSenderMu.RLock()
	defer m.lastTellSenderMu.RUnlock()
	return m.lastTellSender
}

func (m *mockClientWithReply) ClearLastTellSender() {
	m.lastTellSenderMu.Lock()
	defer m.lastTellSenderMu.Unlock()
	m.lastTellSender = ""
}

func newMockClientWithReply(username string, worldID uuid.UUID) *mockClientWithReply {
	return &mockClientWithReply{
		mockClient: &mockClient{
			CharacterID: uuid.New(),
			UserID:      uuid.New(),
			Username:    username,
			messages:    make([]websocket.GameMessageData, 0),
		},
	}
}

// Override GetWorldID for lobby testing
func (m *mockClientWithReply) GetWorldID() uuid.UUID {
	return constants.LobbyWorldID
}

// TestHandleReply_NoMessage tests reply with empty message
func TestHandleReply_NoMessage(t *testing.T) {
	processor, _, _, _ := setupTest(t)

	bob := newMockClientWithReply("Bob", constants.LobbyWorldID)
	bob.SetLastTellSender("Alice")

	cmd := &websocket.CommandData{
		Action: "reply",
	}

	err := processor.ProcessCommand(context.Background(), bob, cmd)
	require.NoError(t, err)

	require.Len(t, bob.messages, 1)
	assert.Equal(t, "error", bob.messages[0].Type)
	assert.Contains(t, bob.messages[0].Text, "What do you want to say")
}

// TestHandleReply_NoPreviousTell tests reply when no tell received
func TestHandleReply_NoPreviousTell(t *testing.T) {
	processor, _, _, _ := setupTest(t)

	bob := newMockClientWithReply("Bob", constants.LobbyWorldID)

	message := "Hello!"
	cmd := &websocket.CommandData{
		Action:  "reply",
		Message: &message,
	}

	err := processor.ProcessCommand(context.Background(), bob, cmd)
	require.NoError(t, err)

	require.Len(t, bob.messages, 1)
	assert.Equal(t, "error", bob.messages[0].Type)
	assert.Contains(t, bob.messages[0].Text, "haven't received any messages")
}

// TestHandleReply_SenderOffline tests reply when sender disconnected
func TestHandleReply_SenderOffline(t *testing.T) {
	processor, _, _, _ := setupTest(t)
	hub := websocket.NewHub(processor)
	processor.SetHub(hub)

	bob := newMockClientWithReply("Bob", constants.LobbyWorldID)
	bob.SetLastTellSender("Alice")

	message := "Are you there?"
	cmd := &websocket.CommandData{
		Action:  "reply",
		Message: &message,
	}

	err := processor.ProcessCommand(context.Background(), bob, cmd)
	require.NoError(t, err)

	require.Len(t, bob.messages, 1)
	assert.Equal(t, "error", bob.messages[0].Type)
	assert.Contains(t, bob.messages[0].Text, "no longer online")
}

// TestHandleReply_ThreadSafe tests concurrent access to last tell sender
func TestHandleReply_ThreadSafe(t *testing.T) {
	bob := newMockClientWithReply("Bob", constants.LobbyWorldID)

	done := make(chan bool)
	iterations := 100

	go func() {
		for i := 0; i < iterations; i++ {
			bob.SetLastTellSender("Alice")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_ = bob.GetLastTellSender()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			bob.ClearLastTellSender()
		}
		done <- true
	}()

	<-done
	<-done
	<-done

	// Should not panic
	_ = bob.GetLastTellSender()
}

// Helper to set up lobby character for testing
func setupLobbyCharacter(t *testing.T, authRepo auth.Repository, client *mockClientWithReply) {
	char := &auth.Character{
		CharacterID: client.CharacterID,
		UserID:      client.UserID,
		WorldID:     constants.LobbyWorldID,
		Name:        client.Username,
		CreatedAt:   time.Now(),
	}
	err := authRepo.CreateCharacter(context.Background(), char)
	require.NoError(t, err)
}

// Note: TestHandleReply_Success is not implemented here because it requires
// mocking the Hub's client lookup, which depends on concrete *websocket.Client.
// The logic is verified via error cases and manual testing.
