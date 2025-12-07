package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mud-platform-backend/internal/spatial"
)

// MockMessageProcessor for testing
type MockMessageProcessor struct {
	mock.Mock
}

func (m *MockMessageProcessor) ProcessCommand(ctx context.Context, client GameClient, cmd *CommandData) error {
	args := m.Called(ctx, client, cmd)
	return args.Error(0)
}

func TestNewHub(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.Clients)
	assert.NotNil(t, hub.SpatialIndex)
	assert.Equal(t, processor, hub.Processor)
}

func TestHub_RegisterUnregister(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		ID:          uuid.New(),
		CharacterID: uuid.New(),
		Send:        make(chan []byte, 256),
	}

	// Register
	hub.Register <- client

	// Wait for registration
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.GetClientCount())
	storedClient, ok := hub.GetClientByCharacter(client.CharacterID)
	assert.True(t, ok)
	assert.Equal(t, client, storedClient)

	// Unregister
	hub.Unregister <- client

	// Wait for unregistration
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, hub.GetClientCount())
	_, ok = hub.GetClientByCharacter(client.CharacterID)
	assert.False(t, ok)
}

func TestHub_BroadcastToCharacter(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)

	client := &Client{
		CharacterID: uuid.New(),
		Send:        make(chan []byte, 256),
	}

	// Manually add client to avoid race conditions with Run()
	hub.Clients[client.CharacterID] = client

	hub.BroadcastToCharacter(client.CharacterID, "test_msg", "hello")

	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "test_msg")
		assert.Contains(t, string(msg), "hello")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}

func TestHub_BroadcastToAll(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)

	client1 := &Client{CharacterID: uuid.New(), Send: make(chan []byte, 256)}
	client2 := &Client{CharacterID: uuid.New(), Send: make(chan []byte, 256)}

	hub.Clients[client1.CharacterID] = client1
	hub.Clients[client2.CharacterID] = client2

	hub.BroadcastToAll("broadcast_msg", "hello all")

	// Check client 1
	select {
	case msg := <-client1.Send:
		assert.Contains(t, string(msg), "broadcast_msg")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message client 1")
	}

	// Check client 2
	select {
	case msg := <-client2.Send:
		assert.Contains(t, string(msg), "broadcast_msg")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message client 2")
	}
}

func TestHub_BroadcastToArea(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)

	// Client in range
	client1 := &Client{CharacterID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Clients[client1.CharacterID] = client1
	hub.SpatialIndex.Insert(client1.CharacterID, spatial.Position{X: 10, Y: 10})

	// Client out of range
	client2 := &Client{CharacterID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Clients[client2.CharacterID] = client2
	hub.SpatialIndex.Insert(client2.CharacterID, spatial.Position{X: 1000, Y: 1000})

	// Broadcast at (0,0) with radius 100
	hub.BroadcastToArea(spatial.Position{X: 0, Y: 0}, 100, "area_msg", "nearby")

	// Client 1 should receive
	select {
	case msg := <-client1.Send:
		assert.Contains(t, string(msg), "area_msg")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message client 1")
	}

	// Client 2 should NOT receive
	select {
	case <-client2.Send:
		t.Fatal("Client 2 received message but should be out of range")
	default:
		// OK
	}
}

func TestHub_GetClientsByWorldID(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)

	world1 := uuid.New()
	world2 := uuid.New()

	client1 := &Client{CharacterID: uuid.New(), WorldID: world1}
	client2 := &Client{CharacterID: uuid.New(), WorldID: world1}
	client3 := &Client{CharacterID: uuid.New(), WorldID: world2}

	hub.Clients[client1.CharacterID] = client1
	hub.Clients[client2.CharacterID] = client2
	hub.Clients[client3.CharacterID] = client3

	clients := hub.GetClientsByWorldID(world1)
	assert.Len(t, clients, 2)

	clients2 := hub.GetClientsByWorldID(world2)
	assert.Len(t, clients2, 1)
}

func TestHub_UpdateCharacterPosition(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)

	charID := uuid.New()

	// Initial update
	hub.UpdateCharacterPosition(charID, 100, 200)

	// Verify in spatial index
	results := hub.SpatialIndex.QueryRadius(spatial.Position{X: 100, Y: 200}, 1)
	assert.Contains(t, results, charID)
}

func TestHub_GetAllClients(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)

	client1 := &Client{CharacterID: uuid.New()}
	client2 := &Client{CharacterID: uuid.New()}

	hub.Clients[client1.CharacterID] = client1
	hub.Clients[client2.CharacterID] = client2

	clients := hub.GetAllClients()
	assert.Len(t, clients, 2)
}

func TestHub_HandleClientMessage(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		CharacterID: uuid.New(),
		Send:        make(chan []byte, 10),
	}

	// Test valid command
	cmdData := []byte(`{"action":"say","message":"hello"}`)
	clientMsg := &ClientMessage{
		Type: MessageTypeCommand,
		Data: cmdData,
	}

	processor.On("ProcessCommand", mock.Anything, client, mock.Anything).Return(nil)

	hub.HandleMessage <- &ClientMessageWrapper{
		Client:  client,
		Message: clientMsg,
	}

	// Wait for processing (async)
	time.Sleep(50 * time.Millisecond)
	processor.AssertExpectations(t)

	// Test invalid command format
	invalidMsg := &ClientMessage{
		Type: MessageTypeCommand,
		Data: []byte(`invalid json`),
	}

	hub.HandleMessage <- &ClientMessageWrapper{
		Client:  client,
		Message: invalidMsg,
	}

	// Should receive error
	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "Invalid command format")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for error message")
	}
}

func TestHub_BroadcastConcurrent(t *testing.T) {
	processor := &MockMessageProcessor{}
	hub := NewHub(processor)

	// Create enough clients to trigger concurrent broadcast (threshold is 10)
	numClients := 20
	clients := make([]*Client, numClients)

	for i := 0; i < numClients; i++ {
		client := &Client{
			CharacterID: uuid.New(),
			Send:        make(chan []byte, 10),
		}
		hub.Clients[client.CharacterID] = client
		hub.SpatialIndex.Insert(client.CharacterID, spatial.Position{X: 0, Y: 0})
		clients[i] = client
	}

	// Broadcast to area
	hub.BroadcastToArea(spatial.Position{X: 0, Y: 0}, 100, "test_msg", "data")

	// Verify all clients received message
	for i, client := range clients {
		select {
		case <-client.Send:
			// OK
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("Timeout waiting for message client %d", i)
		}
	}
}
