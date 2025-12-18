package processor

import (
	"context"
	"fmt"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/errors"
	"tw-backend/internal/game/services/inventory"
	"tw-backend/internal/worldentity"
)

// handleGetObject attempts to pick up a world entity
func (p *GameProcessor) handleGetObject(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil || *cmd.Target == "" {
		client.SendGameMessage("error", "Get what?", nil)
		return nil
	}

	charID := client.GetCharacterID()

	authChar, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		client.SendGameMessage("error", "Failed to find your character.", nil)
		return nil
	}

	// Check if we have the worldEntityService
	if p.worldEntityService == nil {
		client.SendGameMessage("error", "World interaction unavailable.", nil)
		return nil
	}

	// Try to find the entity by name
	entity, err := p.worldEntityService.GetEntityByName(ctx, authChar.WorldID, *cmd.Target)
	if err != nil || entity == nil {
		client.SendGameMessage("error", fmt.Sprintf("You don't see any '%s' here.", *cmd.Target), nil)
		return nil
	}

	// Check distance
	dx := entity.X - authChar.PositionX
	dy := entity.Y - authChar.PositionY

	distSq := dx*dx + dy*dy
	if distSq > 9.0 { // 3 meters squared
		client.SendGameMessage("error", fmt.Sprintf("You are too far away from the %s.", entity.Name), nil)
		return nil
	}

	// Check if entity can be interacted with
	allowed, msg := p.worldEntityService.CanInteract(entity, "get")
	if !allowed {
		client.SendGameMessage("error", msg, nil)
		return nil
	}

	// Remove from world
	if err := p.worldEntityService.Delete(ctx, entity.ID); err != nil {
		return fmt.Errorf("failed to pick up item: %w", err)
	}

	// Add to inventory
	invItem := inventory.Item{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
	}
	metadata := map[string]interface{}{
		"name":        invItem.Name,
		"description": invItem.Description,
	}
	if err := p.inventoryService.AddItem(ctx, charID, invItem.ID, 1, metadata); err != nil {
		return errors.NewInternalError("failed to add item to inventory: %v", err)
	}

	client.SendGameMessage("action", fmt.Sprintf("You pick up the %s.", entity.Name), nil)
	p.sendStateUpdate(client)
	return nil
}

// handlePushObject attempts to push/move a world entity
func (p *GameProcessor) handlePushObject(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil || *cmd.Target == "" {
		client.SendGameMessage("error", "Push what?", nil)
		return nil
	}

	target := *cmd.Target
	charID := client.GetCharacterID()

	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		client.SendGameMessage("error", "Failed to find your character.", nil)
		return nil
	}

	// Check if we have the worldEntityService
	if p.worldEntityService == nil {
		// Fall back to legacy behavior
		client.SendGameMessage("action", fmt.Sprintf("You push the %s.", target), nil)
		return nil
	}

	// Try to find the entity by name
	entity, err := p.worldEntityService.GetEntityByName(ctx, char.WorldID, target)
	if err != nil {
		client.SendGameMessage("error", fmt.Sprintf("You don't see any '%s' here.", target), nil)
		return nil
	}

	// Check if entity can be interacted with
	allowed, msg := p.worldEntityService.CanInteract(entity, "push")
	if !allowed {
		client.SendGameMessage("error", msg, nil)
		return nil
	}

	// Check entity type - static objects can't be pushed even if unlocked
	if entity.EntityType == worldentity.EntityTypeStatic {
		client.SendGameMessage("error", fmt.Sprintf("The %s is too heavy to move.", entity.Name), nil)
		return nil
	}

	// TODO: Implement actual pushing logic (moving entity position)
	client.SendGameMessage("action", fmt.Sprintf("You push the %s.", entity.Name), nil)
	return nil
}
