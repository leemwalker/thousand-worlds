package processor

import (
	"context"
	"tw-backend/cmd/game-server/websocket"
)

// handleFly toggles flight mode for the character
func (p *GameProcessor) handleFly(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	charID := client.GetCharacterID()
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Toggle flight mode
	char.IsFlying = !char.IsFlying

	// Update character in database
	if err := p.authRepo.UpdateCharacter(ctx, char); err != nil {
		client.SendGameMessage("error", "Failed to update flight mode", nil)
		return nil
	}

	// Send appropriate message
	if char.IsFlying {
		client.SendGameMessage("system", "🦅 You take to the skies! The ground shrinks below you as you gain altitude. Your vision expands to take in the vast landscape.", nil)
	} else {
		client.SendGameMessage("system", "🚶 You descend and land gently on the ground. The world returns to its normal perspective.", nil)
	}

	// Send updated map with expanded view
	p.sendMapUpdate(ctx, client)

	return nil
}
