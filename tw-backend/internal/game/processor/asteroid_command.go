package processor

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"tw-backend/cmd/game-server/websocket"
)

// handleSpawnAsteroid processes the 'spawn_asteroid' command
// Syntax: spawn_asteroid <lat> <lon> <mass_kg>
// Example: spawn_asteroid 0 0 1000000
func (p *GameProcessor) handleSpawnAsteroid(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// 1. Check permissions (optional, for now open to all for testing)
	// if !client.IsAdmin() { ... }

	// 2. Parse arguments from cmd.Text or cmd.Target
	// Since generic commands might put arguments in Target or just be in Text
	// We'll trust cmd.Text if available, otherwise reconstruct
	args := strings.Fields(cmd.Text)
	if len(args) < 4 { // spawn_asteroid lat lon mass
		client.SendGameMessage("error", "Usage: spawn_asteroid <lat> <lon> <mass>", nil)
		return nil
	}

	lat, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		client.SendGameMessage("error", "Invalid latitude", nil)
		return nil
	}

	lon, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		client.SendGameMessage("error", "Invalid longitude", nil)
		return nil
	}

	mass, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		client.SendGameMessage("error", "Invalid mass", nil)
		return nil
	}

	// 3. Create Impact Event Data
	impactData := map[string]interface{}{
		"type": "asteroid_impact",
		"location": map[string]float64{
			"lat": lat,
			"lon": lon,
		},
		"mass":        mass,
		"impact_time": time.Now().UnixMilli() + 5000, // Impact in 5 seconds
		"origin": map[string]float64{
			// Random start point in space for visual approach
			"distance": 10.0,     // Planet radii
			"phi":      lat + 45, // Angle offset
			"theta":    lon + 45,
		},
	}

	// 4. Broadcast to all clients in the world
	// If the server was fully strictly ECS, we'd emit an event to the simulation loop.
	// For visualization phase, we broadcast directly.
	if p.Hub != nil {
		clients := p.Hub.GetClientsByWorldID(client.GetWorldID())
		for _, c := range clients {
			c.SendGameMessage("asteroid_impact", "WARNING: Asteroid impact detected!", impactData)
		}
	} else {
		// Fallback for testing or when Hub not available, send to self
		client.SendGameMessage("asteroid_impact", "WARNING: Asteroid impact detected!", impactData)
	}

	log.Printf("[ASTEROID] Spawned asteroid targeting %.2f, %.2f with mass %.0f kg", lat, lon, mass)
	return nil
}
