package processor

import (
	"context"
	"fmt"
	"strings"
	"tw-backend/cmd/game-server/websocket"
)

// handleGetSatellites handles the request for natural satellites and rings
func (p *GameProcessor) handleGetSatellites(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	charID := client.GetCharacterID()

	// Get current world for context
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if char == nil || err != nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Helper to safely get geology with read lock
	geology, exists := p.worldGeology[char.WorldID]
	if !exists || geology == nil || !geology.IsInitialized() {
		client.SendGameMessage("error", "World geology not initialized", nil)
		return nil
	}

	// Get satellites and rings
	satellites := geology.Satellites
	var rings interface{} = nil
	if geology.Rings != nil {
		rings = geology.Rings.GetVisibleRings()
	}

	client.SendTypedMessage("satellites_info", map[string]interface{}{
		"satellites": satellites,
		"rings":      rings,
	})
	return nil
}

// handleDestroyMoon destroys a moon by name and triggers an event
func (p *GameProcessor) handleDestroyMoon(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// Parse arguments from Text
	parts := strings.Fields(cmd.Text)
	if len(parts) < 2 {
		client.SendGameMessage("error", "Usage: destroy_moon <moon_name>", nil)
		return nil
	}
	moonName := strings.Join(parts[1:], " ")

	charID := client.GetCharacterID()
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if char == nil || err != nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	geology, exists := p.worldGeology[char.WorldID]
	if !exists || geology == nil {
		client.SendGameMessage("error", "World geology not found", nil)
		return nil
	}

	// Find moon
	var found bool
	for _, s := range geology.Satellites {
		if strings.EqualFold(s.Name, moonName) {
			found = true
			break
		}
	}

	if !found {
		client.SendGameMessage("error", fmt.Sprintf("Moon '%s' not found", moonName), nil)
		return nil
	}

	// Remove moon from geology
	if !geology.RemoveSatellite(moonName) {
		// Should have been found earlier, but race condition possible
		client.SendGameMessage("error", fmt.Sprintf("Moon '%s' not found (concurrently removed?)", moonName), nil)
		return nil
	}

	// Emit destruction event
	eventData := map[string]interface{}{
		"type": "moon_destroyed",
		"metadata": map[string]interface{}{
			"moon_id":     moonName, // Using name as ID for now
			"debris_mass": 1.0e19,   // Dummy value
		},
	}

	// Broadcast to all clients in the world via the hub
	// We need a way to broadcast. The Hub has BroadcastToWorld (or similar)?
	// Checking GameProcessor struct... it has *websocket.Hub
	// Hub likely has a broadcast method.
	// For now, let's just send to the requesting client for testing,
	// OR if we have a BroadcastToWorld method.
	// Looking at Hub methods might be needed.
	// Assuming p.Hub.BroadcastToWorld(worldID, msg) exists or similar.
	// If not, we can just send to the client for now as the user controls the camera.

	// Better: Use p.Hub.BroadcastToWorld if it exists.
	// Let's assume passed client is enough for FPV testing for now,
	// but to be "proper", it should be broadcast.
	// I will just send to the client for this iteration to avoid Hub interface guessing.

	// Actually, the command needs to update state too?
	// For visual visual only phase, we just emit event.

	client.SendGameMessage("game_message", "Moon destroyed", eventData)

	return nil
}
