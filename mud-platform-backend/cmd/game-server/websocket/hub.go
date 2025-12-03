package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"mud-platform-backend/internal/metrics"
	"mud-platform-backend/internal/spatial"

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

	// Spatial index for efficient area queries
	// Reduces nearby player lookups from O(N) to O(k)
	SpatialIndex *spatial.SpatialGrid

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
	ProcessCommand(ctx context.Context, client GameClient, cmd *CommandData) error
}

// NewHub creates a new WebSocket hub
func NewHub(processor MessageProcessor) *Hub {
	hub := &Hub{
		Clients:       make(map[uuid.UUID]*Client),
		SpatialIndex:  spatial.NewSpatialGrid(100.0), // 100m cells
		HandleMessage: make(chan *ClientMessageWrapper, 256),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Processor:     processor,
	}
	return hub
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
			// Add to spatial index with default position (0, 0)
			// Position will be updated when character moves
			h.SpatialIndex.Insert(client.CharacterID, spatial.Position{X: 0, Y: 0})
			h.mu.Unlock()
			metrics.SetActiveConnections(len(h.Clients))
			log.Printf("Client registered: %s (character: %s)", client.ID, client.CharacterID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.CharacterID]; ok {
				delete(h.Clients, client.CharacterID)
				// Remove from spatial index
				h.SpatialIndex.Remove(client.CharacterID)
				close(client.Send)
			}
			h.mu.Unlock()
			metrics.SetActiveConnections(len(h.Clients))
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

// UpdateCharacterPosition updates a character's position in the spatial index
// This should be called whenever a character moves
func (h *Hub) UpdateCharacterPosition(characterID uuid.UUID, x, y float64) {
	h.SpatialIndex.Insert(characterID, spatial.Position{X: x, Y: y})
}

// BroadcastToArea sends a message to all clients within a radius
// Performance: O(k/W) where k = clients in area, W = worker count
// Uses concurrent workers for parallel message sending
func (h *Hub) BroadcastToArea(center spatial.Position, radius float64, msgType string, data interface{}) {
	start := time.Now()
	defer func() {
		metrics.RecordHubBroadcast(time.Since(start))
		metrics.RecordMessageProcessed(msgType)
	}()

	// Query spatial index for nearby entities (O(k))
	entityIDs := h.SpatialIndex.QueryRadius(center, radius)

	h.mu.RLock()
	// Build list of clients to notify
	clients := make([]*Client, 0, len(entityIDs))
	for _, entityID := range entityIDs {
		if client, ok := h.Clients[entityID]; ok {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()

	if len(clients) == 0 {
		return
	}

	// Use concurrent workers for large broadcasts
	const workerThreshold = 10
	if len(clients) < workerThreshold {
		// Small broadcast - send serially (avoid goroutine overhead)
		for _, client := range clients {
			client.SendMessage(msgType, data)
		}
		return
	}

	// Large broadcast - use worker pool for parallelization
	h.broadcastConcurrent(clients, msgType, data)
}

// broadcastConcurrent sends messages to clients using a worker pool
// Parallelizes message sending to reduce total broadcast time
func (h *Hub) broadcastConcurrent(clients []*Client, msgType string, data interface{}) {
	const numWorkers = 4 // Tune based on CPU cores

	// Create job channel
	jobs := make(chan *Client, len(clients))
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for client := range jobs {
				client.SendMessage(msgType, data)
			}
		}()
	}

	// Distribute jobs to workers
	for _, client := range clients {
		jobs <- client
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()
}

// GetClientCount returns the current number of connected clients
// Thread-safe method for health checks and metrics
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

// GetClientsByWorldID returns all clients in a specific world
func (h *Hub) GetClientsByWorldID(worldID uuid.UUID) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	for _, client := range h.Clients {
		if client.WorldID == worldID {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetAllClients returns all connected clients
func (h *Hub) GetAllClients() []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*Client, 0, len(h.Clients))
	for _, client := range h.Clients {
		clients = append(clients, client)
	}
	return clients
}
