package processor

import (
	"context"
	"fmt"
	"math"
	"strings"
	"tw-backend/cmd/game-server/websocket"
)

// handleFly toggles flight mode for the character
// handleFly toggles flight mode or adjusts altitude
func (p *GameProcessor) handleFly(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	charID := client.GetCharacterID()
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Get world bounds to calculate max altitude
	maxAltitude := 1000.0 // Default fallback
	if p.worldRepo != nil {
		world, err := p.worldRepo.GetWorld(ctx, char.WorldID)
		if err == nil && world != nil && world.BoundsMin != nil && world.BoundsMax != nil {
			// Calculate world diameter (use larger dimension)
			widthX := world.BoundsMax.X - world.BoundsMin.X
			widthY := world.BoundsMax.Y - world.BoundsMin.Y
			diameter := math.Max(widthX, widthY)
			// Max altitude = 0.4% of diameter (Earth stratosphere ratio)
			maxAltitude = diameter * 0.004
			if maxAltitude < 100 {
				maxAltitude = 100 // Minimum useful altitude
			}
		}
	}

	arg := ""
	if cmd.Target != nil {
		arg = strings.ToLower(strings.TrimSpace(*cmd.Target))
	}

	// Case 1: No argument - Toggle flight
	if arg == "" {
		char.IsFlying = !char.IsFlying
		if char.IsFlying {
			if char.PositionZ < 1 {
				char.PositionZ = 10 // Default takeoff altitude
			}
			client.SendGameMessage("system", fmt.Sprintf("🦅 You take to the skies! (Altitude: %.0fm, Max: %.0fm)", char.PositionZ, maxAltitude), nil)
		} else {
			char.PositionZ = 0
			client.SendGameMessage("system", "🚶 You descend and land gently on the ground.", nil)
		}
	} else {
		// Case 2: Arguments provided
		change := 0.0

		switch arg {
		case "up":
			change = 10.0
		case "down":
			change = -10.0
		default:
			// Try to parse number
			var alt float64
			if _, err := fmt.Sscanf(arg, "%f", &alt); err == nil {
				// Absolute altitude - clamp to max
				if alt > maxAltitude {
					alt = maxAltitude
					client.SendGameMessage("system", fmt.Sprintf("⚠️ Maximum altitude for this world is %.0fm.", maxAltitude), nil)
				}
				char.PositionZ = alt
				if char.PositionZ > 0 {
					char.IsFlying = true
				} else {
					char.PositionZ = 0
					char.IsFlying = false
				}
				client.SendGameMessage("system", fmt.Sprintf("🦅 Altitude set to %.0fm.", char.PositionZ), nil)
				goto Update
			} else {
				client.SendGameMessage("error", "Usage: fly [up|down|<altitude>]", nil)
				return nil
			}
		}

		// Apply relative change
		char.PositionZ += change

		// Handle landing logic
		if char.PositionZ <= 0 {
			char.PositionZ = 0
			char.IsFlying = false
			client.SendGameMessage("system", "🚶 You touch down on the ground.", nil)
		} else {
			// Clamp to max altitude
			if char.PositionZ > maxAltitude {
				char.PositionZ = maxAltitude
				client.SendGameMessage("system", fmt.Sprintf("⚠️ You've reached the maximum altitude of %.0fm for this world.", maxAltitude), nil)
			} else if !char.IsFlying {
				char.IsFlying = true // Auto-takeoff
				client.SendGameMessage("system", "🦅 You lift off from the ground!", nil)
			} else if change > 0 {
				client.SendGameMessage("system", fmt.Sprintf("🦅 Climbing... (Altitude: %.0fm)", char.PositionZ), nil)
			} else {
				client.SendGameMessage("system", fmt.Sprintf("🦅 Descending... (Altitude: %.0fm)", char.PositionZ), nil)
			}
		}
	}

Update:
	// Update character in database
	if err := p.authRepo.UpdateCharacter(ctx, char); err != nil {
		client.SendGameMessage("error", "Failed to update flight status", nil)
		return nil
	}

	// Send updated map with expanded view
	p.sendMapUpdate(ctx, client)

	return nil
}
