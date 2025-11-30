package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
)

// ClientMessageWrapper wraps a client message with the client reference
type ClientMessageWrapper struct {
	Client  *Client
	Message *ClientMessage
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	// Registered clients by character ID
	Clients map[uuid.UUID]*Client

	// Inbound messages from clients
	HandleMessage chan *ClientMessageWrapper

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Message processor
	Processor MessageProcessor

	mu sync.RWMutex
}

// MessageProcessor handles game logic for messages
type MessageProcessor interface {
	ProcessCommand(ctx context.Context, client *Client, cmd *CommandData) error
}

// NewHub creates a new WebSocket hub
func NewHub(processor MessageProcessor) *Hub {
	return &Hub{
		Clients:       make(map[uuid.UUID]*Client),
		HandleMessage: make(chan *ClientMessageWrapper, 256),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Processor:     processor,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.CharacterID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s (character: %s)", client.ID, client.CharacterID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.CharacterID]; ok {
				delete(h.Clients, client.CharacterID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.ID)

		case wrapper := <-h.HandleMessage:
			h.handleClientMessage(ctx, wrapper)
		}
	}
}

// handleClientMessage processes a message from a client
func (h *Hub) handleClientMessage(ctx context.Context, wrapper *ClientMessageWrapper) {
	switch wrapper.Message.Type {
	case MessageTypeCommand:
		var cmd CommandData
		if err := json.Unmarshal(wrapper.Message.Data, &cmd); err != nil {
			wrapper.Client.SendError("Invalid command format")
			return
		}

		if err := h.Processor.ProcessCommand(ctx, wrapper.Client, &cmd); err != nil {
			wrapper.Client.SendError(err.Error())
			log.Printf("Error processing command: %v", err)
		}

	default:
		wrapper.Client.SendError("Unknown message type")
	}
}

// BroadcastToCharacter sends a message to a specific character
func (h *Hub) BroadcastToCharacter(characterID uuid.UUID, msgType string, data interface{}) {
	h.mu.RLock()
	client, ok := h.Clients[characterID]
	h.mu.RUnlock()

	if ok {
		client.SendMessage(msgType, data)
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(msgType string, data interface{}) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.Clients))
	for _, client := range h.Clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		client.SendMessage(msgType, data)
	}
}

// GetClientByCharacter returns a client by character ID
func (h *Hub) GetClientByCharacter(characterID uuid.UUID) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.Clients[characterID]
	return client, ok
}
