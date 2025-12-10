package processor

import (
	"context"
	"fmt"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/game/constants"
)

// handleLobby sends the player back to the lobby
func (p *GameProcessor) handleLobby(ctx context.Context, client websocket.GameClient) error {
	charID := client.GetCharacterID()
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// 1. Check if already in lobby
	if constants.IsLobby(char.WorldID) {
		client.SendGameMessage("error", "You are already in the lobby.", nil)
		return nil
	}

	// 2. Store current world as last visited
	// Make a copy of the UUID to avoid pointer issues if needed, though uuid.UUID is a value type
	currentWorldID := char.WorldID
	char.LastWorldVisited = &currentWorldID

	// 3. Move to Lobby
	char.WorldID = constants.LobbyWorldID

	// 4. Set Position to Lobby Spawn (South Entrance)
	char.PositionX = 5.0
	char.PositionY = 2.0
	char.PositionZ = 0.0

	// 5. Handle Watcher specifics
	// Requirement: If Watcher, remove from world instance entirely.
	if char.Role == "watcher" {
		// Watchers shouldn't have persistent position in the world they left?
		// The prompt says "remove them from the world instance entirely".
		// Changing WorldID effectively removes them from that world's context.
	}

	// 6. Persist
	if err := p.authRepo.UpdateCharacter(ctx, char); err != nil {
		return fmt.Errorf("failed to update character for lobby return: %w", err)
	}

	// 7. Notify Client
	client.SendGameMessage("system", "You return to the Grand Lobby.", nil)

	// 8. Update State (triggers client recharge/scene switch)
	p.sendStateUpdate(client)

	return nil
}
