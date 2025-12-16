package processor

import (
	"context"
	"fmt"
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
			client.SendGameMessage("system", fmt.Sprintf("🦅 You take to the skies! (Altitude: %.0fm)", char.PositionZ), nil)
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
				// Absolute altitude
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
			if !char.IsFlying {
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
