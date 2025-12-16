// Package ecosystem provides real-time event broadcasting to connected clients.
package ecosystem

import (
	"sync"

	"tw-backend/internal/ecosystem/geography"

	"github.com/google/uuid"
)

// BroadcastClient interface for sending messages (avoids import cycle with websocket)
type BroadcastClient interface {
	SendGameMessage(msgType, content string, data interface{})
}

// BroadcastEventType categorizes broadcast events (separate from logger EventType)
type BroadcastEventType string

const (
	BroadcastSimulation BroadcastEventType = "simulation"
	BroadcastGeological BroadcastEventType = "geological"
	BroadcastSpeciation BroadcastEventType = "speciation"
	BroadcastExtinction BroadcastEventType = "extinction"
	BroadcastTurning    BroadcastEventType = "turning_point"
	BroadcastMinimap    BroadcastEventType = "minimap"
)

// BroadcastEvent represents an event to broadcast
type BroadcastEvent struct {
	Type     BroadcastEventType     `json:"type"`
	Year     int64                  `json:"year"`
	Message  string                 `json:"message"`
	Severity float64                `json:"severity,omitempty"`
	Position *geography.HexCoord    `json:"position,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// MinimapCell represents a single cell update for the minimap
type MinimapCell struct {
	Q       int       `json:"q"`
	R       int       `json:"r"`
	Emoji   string    `json:"emoji"`
	BiomeID uuid.UUID `json:"biome_id,omitempty"`
}

// MinimapUpdate contains a batch of cell updates
type MinimapUpdate struct {
	Cells []MinimapCell `json:"cells"`
	Year  int64         `json:"year"`
}

// SimulationBroadcaster handles real-time event broadcasting to connected clients
type SimulationBroadcaster struct {
	worldID uuid.UUID
	clients map[uuid.UUID]BroadcastClient
	mu      sync.RWMutex
}

// NewSimulationBroadcaster creates a new broadcaster for a world
func NewSimulationBroadcaster(worldID uuid.UUID) *SimulationBroadcaster {
	return &SimulationBroadcaster{
		worldID: worldID,
		clients: make(map[uuid.UUID]BroadcastClient),
	}
}

// AddClient registers a client to receive broadcasts
func (sb *SimulationBroadcaster) AddClient(clientID uuid.UUID, client BroadcastClient) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.clients[clientID] = client
}

// RemoveClient unregisters a client
func (sb *SimulationBroadcaster) RemoveClient(clientID uuid.UUID) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	delete(sb.clients, clientID)
}

// BroadcastEvent sends an event to all connected clients
func (sb *SimulationBroadcaster) BroadcastEvent(event BroadcastEvent) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	for _, client := range sb.clients {
		client.SendGameMessage("simulation", event.Message, map[string]interface{}{
			"type":     event.Type,
			"year":     event.Year,
			"severity": event.Severity,
			"data":     event.Data,
		})
	}
}

// BroadcastMinimapUpdate sends minimap cell updates to all clients
func (sb *SimulationBroadcaster) BroadcastMinimapUpdate(update MinimapUpdate) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	for _, client := range sb.clients {
		client.SendGameMessage("minimap", "", map[string]interface{}{
			"type":  "minimap_update",
			"year":  update.Year,
			"cells": update.Cells,
		})
	}
}

// BroadcastProximityEvent sends an event only to clients within range
func (sb *SimulationBroadcaster) BroadcastProximityEvent(event BroadcastEvent, maxDistance int, getPlayerPos func(uuid.UUID) *geography.HexCoord) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if event.Position == nil {
		// No position - broadcast to all
		for _, client := range sb.clients {
			client.SendGameMessage("simulation", event.Message, map[string]interface{}{
				"type":     event.Type,
				"year":     event.Year,
				"severity": event.Severity,
			})
		}
		return
	}

	for clientID, client := range sb.clients {
		playerPos := getPlayerPos(clientID)
		if playerPos == nil {
			continue
		}

		distance := playerPos.Distance(*event.Position)
		if distance <= maxDistance {
			client.SendGameMessage("simulation", event.Message, map[string]interface{}{
				"type":     event.Type,
				"year":     event.Year,
				"severity": event.Severity,
				"distance": distance,
			})
		}
	}
}

// GetBroadcastDistance returns the appropriate broadcast distance for an event type
func GetBroadcastDistance(eventType BroadcastEventType) int {
	switch eventType {
	case BroadcastGeological:
		return 50 // Regional events like volcanoes
	case BroadcastSpeciation, BroadcastExtinction:
		return 10 // Local - in that biome
	default:
		return -1 // Global (no limit)
	}
}
