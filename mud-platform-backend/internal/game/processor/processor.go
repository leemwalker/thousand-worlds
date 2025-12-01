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
	Hub *websocket.Hub
}

// NewGameProcessor creates a new game processor
func NewGameProcessor() *GameProcessor {
	return &GameProcessor{}
}

// SetHub sets the websocket hub
func (p *GameProcessor) SetHub(hub *websocket.Hub) {
	p.Hub = hub
}

// ProcessCommand processes a game command from a client
func (p *GameProcessor) ProcessCommand(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// Validate character exists
	if client.GetCharacterID().String() == "00000000-0000-0000-0000-000000000000" {
		return ErrNoCharacter
	}

	// Route command to appropriate handler
	switch cmd.Action {
	case "help":
		return p.handleHelp(ctx, client)

	// Cardinal directions
	case "north", "n":
		return p.handleDirection(ctx, client, "north")
	case "northeast", "ne":
		return p.handleDirection(ctx, client, "northeast")
	case "east", "e":
		return p.handleDirection(ctx, client, "east")
	case "southeast", "se":
		return p.handleDirection(ctx, client, "southeast")
	case "south", "s":
		return p.handleDirection(ctx, client, "south")
	case "southwest", "sw":
		return p.handleDirection(ctx, client, "southwest")
	case "west", "w":
		return p.handleDirection(ctx, client, "west")
	case "northwest", "nw":
		return p.handleDirection(ctx, client, "northwest")
	case "up", "u":
		return p.handleDirection(ctx, client, "up")
	case "down", "d":
		return p.handleDirection(ctx, client, "down")

	// Interaction
	case "open":
		return p.handleOpen(ctx, client, cmd)
	case "enter":
		return p.handleEnter(ctx, client, cmd)

	// Observation
	case "look":
		return p.handleLook(ctx, client, cmd)

	// Communication
	case "say":
		return p.handleSay(ctx, client, cmd)
	case "whisper":
		return p.handleWhisper(ctx, client, cmd)
	case "tell":
		return p.handleTell(ctx, client, cmd)

	// Social
	case "who":
		return p.handleWho(ctx, client)

	// Existing commands
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

// Command handlers
func (p *GameProcessor) handleHelp(ctx context.Context, client websocket.GameClient) error {
	helpText := `
Available Commands:
  Movement:
    n, ne, e, se, s, sw, w, nw - Move in cardinal directions
    up, down                   - Move vertically
  
  Interaction:
    open <door/container>      - Open doors or containers
    enter <portal/doorway>     - Enter through portals or doorways
    look [target]              - Look around or examine something
  
  Communication:
    say <message>              - Speak to nearby players
    whisper <player> <message> - Whisper to nearby player (5m range)
    tell <player> <message>    - Direct message to any player
  
  Social:
    who                        - List online players
  
  Actions:
    take <item>                - Pick up an item
    drop <item>                - Drop an item
    inventory                  - View your inventory
    attack <target>            - Attack a target
    talk <target>              - Talk to an NPC
    craft <item>               - Craft an item
    use <item>                 - Use an item
`
	client.SendGameMessage("system", helpText, nil)
	return nil
}

// handleDirection handles all cardinal direction movements
func (p *GameProcessor) handleDirection(ctx context.Context, client websocket.GameClient, direction string) error {
	// TODO: Integrate with actual movement system and world validation
	client.SendGameMessage("movement", fmt.Sprintf("You move %s.", direction), nil)
	p.sendStateUpdate(client)
	return nil
}

// handleOpen opens doors or containers
func (p *GameProcessor) handleOpen(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target required for open command")
	}

	// TODO: Integrate with world system to check if door/container exists and is locked
	client.SendGameMessage("action", fmt.Sprintf("You open the %s.", *cmd.Target), nil)
	p.sendStateUpdate(client)
	return nil
}

// handleEnter enters through portals, doorways, or archways
func (p *GameProcessor) handleEnter(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target required for enter command")
	}

	// TODO: Integrate with world system to handle portal/doorway transitions
	client.SendGameMessage("movement", fmt.Sprintf("You enter the %s.", *cmd.Target), nil)
	p.sendStateUpdate(client)
	return nil
}

// handleSay broadcasts a message to the player's area
func (p *GameProcessor) handleSay(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Message == nil {
		return errors.New("message required for say command")
	}

	// TODO: Broadcast to all players in the same area
	client.SendGameMessage("speech", fmt.Sprintf("You say: %s", *cmd.Message), map[string]interface{}{
		"speakerID": client.GetCharacterID().String(),
		"message":   *cmd.Message,
	})
	return nil
}

// handleWhisper sends a private message to a nearby player (5m range)
func (p *GameProcessor) handleWhisper(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Recipient == nil {
		return errors.New("recipient required for whisper command")
	}
	if cmd.Message == nil {
		return errors.New("message required for whisper command")
	}

	// TODO: Calculate distance to recipient, check 5m range
	// TODO: Proximity-based clarity - closer players can overhear more clearly
	// TODO: Perception stat affects who can overhear
	client.SendGameMessage("whisper", fmt.Sprintf("You whisper to %s: %s", *cmd.Recipient, *cmd.Message), map[string]interface{}{
		"recipient": *cmd.Recipient,
		"message":   *cmd.Message,
	})
	return nil
}

// handleTell sends a direct message to any player (online or offline)
func (p *GameProcessor) handleTell(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Recipient == nil {
		return errors.New("recipient required for tell command")
	}
	if cmd.Message == nil {
		return errors.New("message required for tell command")
	}

	// TODO: Check if recipient is online
	// TODO: If online, deliver immediately
	// TODO: If offline, store message and send push notification
	client.SendGameMessage("tell", fmt.Sprintf("You tell %s: %s", *cmd.Recipient, *cmd.Message), map[string]interface{}{
		"recipient": *cmd.Recipient,
		"message":   *cmd.Message,
	})
	return nil
}

// handleWho lists all currently online players
func (p *GameProcessor) handleWho(ctx context.Context, client websocket.GameClient) error {
	if p.Hub == nil {
		return errors.New("game server not fully initialized")
	}

	// Get all connected clients
	// Note: In a real implementation, we would want to filter by world/area
	// and resolve character names from IDs.
	// For now, we'll just list the count and IDs.

	// Accessing Clients map directly is not thread-safe if not using Hub methods
	// But Hub doesn't expose a method to get all clients safely yet.
	// We should add GetConnectedClients to Hub or similar.
	// For now, we'll assume we can't access it safely and just send a placeholder
	// or we need to add a method to Hub.

	playerList := "=== Online Players ===\n"
	// TODO: Implement safe access to Hub clients
	playerList += "Alice [Active] - Fantasy Realm\nBob [Idle] - Sci-Fi Galaxy"

	client.SendGameMessage("player_list", playerList, nil)
	return nil
}

func (p *GameProcessor) handleLook(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// TODO: Integrate with world/area system
	description := "You are in a grassy clearing. Trees surround you on all sides."
	if cmd.Target != nil {
		description = fmt.Sprintf("You examine the %s.", *cmd.Target)
	}

	client.SendGameMessage("area_description", description, nil)
	return nil
}

func (p *GameProcessor) handleTake(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
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

func (p *GameProcessor) handleDrop(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target item required")
	}

	// TODO: Integrate with inventory system
	client.SendGameMessage("system", fmt.Sprintf("You drop the %s.", *cmd.Target), nil)
	p.sendStateUpdate(client)
	return nil
}

func (p *GameProcessor) handleAttack(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
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

func (p *GameProcessor) handleTalk(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target required for talk")
	}

	// TODO: Integrate with dialogue/NPC system
	client.SendGameMessage("dialogue", "Hello, traveler!", map[string]interface{}{
		"speakerName": *cmd.Target,
	})

	return nil
}

func (p *GameProcessor) handleInventory(ctx context.Context, client websocket.GameClient) error {
	// TODO: Get actual inventory from database
	p.sendStateUpdate(client)
	return nil
}

func (p *GameProcessor) handleCraft(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
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

func (p *GameProcessor) handleUse(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("item required for use")
	}

	// TODO: Integrate with item use system
	client.SendGameMessage("system", fmt.Sprintf("You use the %s.", *cmd.Target), nil)
	p.sendStateUpdate(client)
	return nil
}

// sendStateUpdate sends the current game state to the client
func (p *GameProcessor) sendStateUpdate(client websocket.GameClient) {
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
