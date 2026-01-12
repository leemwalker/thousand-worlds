package processor

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"tw-backend/cmd/game-server/websocket"

	"github.com/google/uuid"
)

// handleSpawnAsteroid processes the 'spawn_asteroid' command
// Syntax: spawn_asteroid <lat> <lon> <mass_kg> [target_type] [target_id]
// Example: spawn_asteroid 0 0 1000000
// Example Moon Formation: spawn_asteroid 0 0 1e20
// Example Moon Destruction: spawn_asteroid 0 0 1e15 moon moon_1
func (p *GameProcessor) handleSpawnAsteroid(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// 1. Parse arguments
	args := strings.Fields(cmd.Text)
	if len(args) < 4 {
		client.SendGameMessage("error", "Usage: spawn_asteroid <lat> <lon> <mass> [target_type] [target_id]", nil)
		return nil
	}

	lat, _ := strconv.ParseFloat(args[1], 64)
	lon, _ := strconv.ParseFloat(args[2], 64)
	mass, _ := strconv.ParseFloat(args[3], 64)

	targetType := "planet"
	targetID := ""
	if len(args) >= 6 {
		targetType = args[4]
		targetID = args[5]
	}

	// 2. Determine Outcome based on Mass and Target
	const MOON_FORMATION_THRESHOLD = 5e19 // kg (approx 1/1500th Earth Moon)

	// Case A: Moon Destruction / Ring Formation
	if targetType == "moon" && targetID != "" {
		// If large asteroid hits moon -> Moon Destroyed -> Ring Formation
		// We assume ANY asteroid explicitly targeting a moon destroys it for gameplay drama,
		// or check mass ratio if we want realism. Let's say mass > 1e12 is enough to crack it visually.
		if mass > 1e12 {
			p.handleMoonDestruction(client, targetID, mass)
			return nil
		}
	}

	// Case B: Moon Formation (Protoplanet capture)
	if targetType == "planet" && mass >= MOON_FORMATION_THRESHOLD {
		p.handleMoonFormation(client, mass)
		return nil
	}

	// Case C: Planet Surface Impact
	// Standard impact event
	impactData := map[string]interface{}{
		"type": "asteroid_impact",
		"location": map[string]float64{
			"lat": lat,
			"lon": lon,
		},
		"mass":        mass,
		"impact_time": time.Now().UnixMilli() + 5000, // Impact in 5 seconds
		"origin": map[string]float64{
			"distance": 10.0,
			"phi":      lat + 45,
			"theta":    lon + 45,
		},
	}

	p.broadcastToWorld(client.GetWorldID(), "asteroid_impact", "WARNING: Asteroid impact detected!", impactData, client)
	log.Printf("[ASTEROID] Spawned asteroid targeting %.2f, %.2f with mass %.0f kg", lat, lon, mass)
	return nil
}

func (p *GameProcessor) handleMoonFormation(client websocket.GameClient, mass float64) {
	// Create a new moon entry
	// In a real ECS system we'd add an entity. Here we'll just broadcast the update.
	newMoon := map[string]interface{}{
		"name":     fmt.Sprintf("Protoplanet-%d", time.Now().Unix()%1000),
		"mass":     mass,
		"distance": 20000 + rand.Float64()*50000, // km
	}

	// Update World State (Placeholder - assumes an in-memory or DB update exists)
	// For visualization, we just verify the event is sent.

	p.broadcastToWorld(client.GetWorldID(), "satellites_info", "New moon captured!", map[string]interface{}{
		"new_moon": newMoon,
		"action":   "add",
	}, client)
	log.Printf("[ASTEROID] Protoplanet captured as new moon: %+v", newMoon)
}

func (p *GameProcessor) handleMoonDestruction(client websocket.GameClient, moonID string, impactMass float64) {
	// 1. Broadcast Moon Destruction
	p.broadcastToWorld(client.GetWorldID(), "moon_destroyed", fmt.Sprintf("Moon %s destroyed by impact!", moonID), map[string]interface{}{
		"moon_id":     moonID,
		"debris_mass": impactMass * 5, // Debris is impact + moon mass chunk
	}, client)

	// 2. Form Ring (Gameplay rule: Moon destruction forms ring)
	p.broadcastToWorld(client.GetWorldID(), "satellites_info", "Debris field formed a ring!", map[string]interface{}{
		"rings": map[string]interface{}{
			"innerRadius": 1.2,
			"outerRadius": 2.5,
			"opacity":     0.8,
			"color":       "#A0A0A0",
		},
		"action": "update_ring",
	}, client)
	log.Printf("[ASTEROID] Moon %s destroyed, ring formed.", moonID)
}

func (p *GameProcessor) broadcastToWorld(worldID uuid.UUID, msgType string, content string, data map[string]interface{}, fallbackClient websocket.GameClient) {
	if p.Hub != nil {
		clients := p.Hub.GetClientsByWorldID(worldID)
		for _, c := range clients {
			c.SendGameMessage(msgType, content, data)
		}
	} else if fallbackClient != nil {
		// Fallback for testing or when Hub is not available
		fallbackClient.SendGameMessage(msgType, content, data)
	}
}
