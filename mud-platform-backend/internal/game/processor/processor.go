package processor

import (
	"context"
	"errors"
	"fmt"

	"mud-platform-backend/cmd/game-server/websocket"
)

var (
	ErrInvalidAction = errors.New("invalid action")
	ErrNoCharacter   = errors.New("no active character")
)

// GameProcessor handles game command processing
type GameProcessor struct {
	// Add game logic dependencies here
	// e.g., characterRepo, worldRepo, combatSystem, etc.
}

// NewGameProcessor creates a new game processor
func NewGameProcessor() *GameProcessor {
	return &GameProcessor{}
}

// ProcessCommand processes a game command from a client
func (p *GameProcessor) ProcessCommand(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	// Validate character exists
	if client.CharacterID.String() == "00000000-0000-0000-0000-000000000000" {
		return ErrNoCharacter
	}

	// Route command to appropriate handler
	switch cmd.Action {
	case "help":
		return p.handleHelp(ctx, client)
	case "move":
		return p.handleMove(ctx, client, cmd)
	case "look":
		return p.handleLook(ctx, client, cmd)
	case "take":
		return p.handleTake(ctx, client, cmd)
	case "drop":
		return p.handleDrop(ctx, client, cmd)
	case "attack":
		return p.handleAttack(ctx, client, cmd)
	case "talk":
		return p.handleTalk(ctx, client, cmd)
	case "inventory":
		return p.handleInventory(ctx, client)
	case "craft":
		return p.handleCraft(ctx, client, cmd)
	case "use":
		return p.handleUse(ctx, client, cmd)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidAction, cmd.Action)
	}
}

// Command handlers (placeholders for full game logic integration)
func (p *GameProcessor) handleHelp(ctx context.Context, client *websocket.Client) error {
	helpText := `
Available Commands:
  help                 - Show this help message
  move <direction>     - Move in a direction (north, south, east, west)
  look [target]        - Look around or examine something
  take <item>          - Pick up an item
  drop <item>          - Drop an item from inventory
  inventory            - View your inventory
  attack <target>      - Attack a target
  talk <target>        - Talk to an NPC
  craft <item>         - Craft an item
  use <item>          - Use an item
`
	client.SendGameMessage("system", helpText, nil)
	return nil
}

func (p *GameProcessor) handleMove(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	if cmd.Direction == nil {
		return errors.New("direction required for move command")
	}

	// TODO: Integrate with actual movement system
	client.SendGameMessage("movement", fmt.Sprintf("You move %s.", *cmd.Direction), nil)

	// Send state update
	p.sendStateUpdate(client)

	return nil
}

func (p *GameProcessor) handleLook(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	// TODO: Integrate with world/area system
	description := "You are in a grassy clearing. Trees surround you on all sides."
	if cmd.Target != nil {
		description = fmt.Sprintf("You examine the %s.", *cmd.Target)
	}

	client.SendGameMessage("area_description", description, nil)
	return nil
}

func (p *GameProcessor) handleTake(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target item required")
	}

	// TODO: Integrate with inventory system
	client.SendGameMessage("item_acquired", fmt.Sprintf("You pick up the %s.", *cmd.Target), map[string]interface{}{
		"itemName": *cmd.Target,
		"quantity": 1,
	})

	p.sendStateUpdate(client)
	return nil
}

func (p *GameProcessor) handleDrop(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target item required")
	}

	// TODO: Integrate with inventory system
	client.SendGameMessage("system", fmt.Sprintf("You drop the %s.", *cmd.Target), nil)
	p.sendStateUpdate(client)
	return nil
}

func (p *GameProcessor) handleAttack(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target required for attack")
	}

	// TODO: Integrate with combat system
	client.SendGameMessage("combat", fmt.Sprintf("You attack %s for 10 damage!", *cmd.Target), map[string]interface{}{
		"targetName": *cmd.Target,
		"damage":     10,
	})

	p.sendStateUpdate(client)
	return nil
}

func (p *GameProcessor) handleTalk(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target required for talk")
	}

	// TODO: Integrate with dialogue/NPC system
	client.SendGameMessage("dialogue", "Hello, traveler!", map[string]interface{}{
		"speakerName": *cmd.Target,
	})

	return nil
}

func (p *GameProcessor) handleInventory(ctx context.Context, client *websocket.Client) error {
	// TODO: Get actual inventory from database
	p.sendStateUpdate(client)
	return nil
}

func (p *GameProcessor) handleCraft(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("item required for crafting")
	}

	// TODO: Integrate with crafting system
	client.SendGameMessage("crafting_success", fmt.Sprintf("You successfully crafted %s!", *cmd.Target), map[string]interface{}{
		"itemName": *cmd.Target,
		"quality":  "common",
	})

	p.sendStateUpdate(client)
	return nil
}

func (p *GameProcessor) handleUse(ctx context.Context, client *websocket.Client, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("item required for use")
	}

	// TODO: Integrate with item use system
	client.SendGameMessage("system", fmt.Sprintf("You use the %s.", *cmd.Target), nil)
	p.sendStateUpdate(client)
	return nil
}

// sendStateUpdate sends the current game state to the client
func (p *GameProcessor) sendStateUpdate(client *websocket.Client) {
	// TODO: Get actual state from database
	state := &websocket.StateUpdateData{
		HP:         100,
		MaxHP:      100,
		Stamina:    80,
		MaxStamina: 100,
		Focus:      60,
		MaxFocus:   100,
		Position: websocket.Position{
			X: 0,
			Y: 0,
		},
		Inventory:    []websocket.InventoryItem{},
		Equipment:    websocket.Equipment{},
		VisibleTiles: []websocket.VisibleTile{},
	}

	client.SendStateUpdate(state)
}
