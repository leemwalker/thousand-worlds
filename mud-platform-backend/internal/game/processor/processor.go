package processor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"mud-platform-backend/cmd/game-server/websocket"
	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/repository"
)

var (
	ErrInvalidAction = errors.New("invalid action")
	ErrNoCharacter   = errors.New("no active character")
)

// GameProcessor handles game command processing
type GameProcessor struct {
	Hub         *websocket.Hub
	authRepo    auth.Repository
	worldRepo   repository.WorldRepository
	lookService *lobby.LookService
}

// NewGameProcessor creates a new game processor
func NewGameProcessor(authRepo auth.Repository, worldRepo repository.WorldRepository, lookService *lobby.LookService) *GameProcessor {
	return &GameProcessor{
		authRepo:    authRepo,
		worldRepo:   worldRepo,
		lookService: lookService,
	}
}

// SetHub sets the websocket hub
func (p *GameProcessor) SetHub(hub *websocket.Hub) {
	p.Hub = hub
}

// ProcessCommand processes a game command from a client
func (p *GameProcessor) ProcessCommand(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// Validate character exists
	charID := client.GetCharacterID()
	if charID == uuid.Nil {
		return ErrNoCharacter
	}

	// Check if character is in Lobby
	// We can check this by querying the character's location or by checking if the client is in "lobby mode"
	// For now, let's assume if the character's WorldID is LobbyWorldID, they are in the lobby.
	// However, the client interface doesn't expose WorldID directly.
	// We might need to fetch the character or rely on the client knowing its state.
	// But wait, the client just sends commands.
	// Let's fetch the character to be sure, or optimize by caching WorldID in the client.
	// For this implementation, let's fetch the character from AuthRepo to check WorldID.
	// Optimization: The client struct in websocket package has CharacterID.
	// We should probably add WorldID to the GameClient interface or fetch it.
	// Let's fetch it for now to be safe and stateless.

	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	if lobby.IsLobby(char.WorldID) {
		return p.processLobbyCommand(ctx, client, cmd)
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

// processLobbyCommand handles commands specifically for the lobby
func (p *GameProcessor) processLobbyCommand(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	switch cmd.Action {
	case "look", "l":
		return p.handleLobbyLook(ctx, client, cmd)
	case "create":
		// Handle "create world"
		if cmd.Target != nil && strings.ToLower(*cmd.Target) == "world" {
			return p.handleCreateWorld(ctx, client)
		}
		return fmt.Errorf("unknown create command")
	case "enter":
		return p.handleLobbyEnter(ctx, client, cmd)
	case "say":
		return p.handleSay(ctx, client, cmd) // Reuse generic say for now
	case "who":
		return p.handleWho(ctx, client) // Reuse generic who for now
	case "help":
		return p.handleLobbyHelp(ctx, client)
	default:
		client.SendGameMessage("error", "Unknown lobby command. Type 'help' for available commands.", nil)
		return nil
	}
}

func (p *GameProcessor) handleLobbyHelp(ctx context.Context, client websocket.GameClient) error {
	helpText := `
Lobby Commands:
  look                      - See the lobby description, available worlds, and other players.
  look <player_name>        - Inspect another player.
  look <portal_name>        - Inspect a world portal.
  look statue               - Examine the central statue.
  create world              - Start the process of creating a new world.
  enter <world_id>          - Enter a specific world.
  say <message>             - Chat with other players in the lobby.
  who                       - See who is online.
`
	client.SendGameMessage("system", helpText, nil)
	return nil
}

func (p *GameProcessor) handleLobbyLook(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// Get user ID for personalized descriptions
	charID := client.GetCharacterID()
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Check if there's a target
	if cmd.Target != nil && *cmd.Target != "" {
		target := strings.ToLower(strings.TrimSpace(*cmd.Target))

		// Check for statue
		if target == "statue" {
			desc, err := p.lookService.DescribeStatue(ctx, char.UserID)
			if err != nil {
				client.SendGameMessage("error", "You can't see that here.", nil)
				return nil
			}
			client.SendGameMessage("look_result", desc, nil)
			return nil
		}

		// Try to look at a player first
		playerDesc, err := p.lookService.DescribePlayer(ctx, target)
		if err == nil {
			client.SendGameMessage("look_result", playerDesc, nil)
			return nil
		}

		// Try to look at a portal/world
		portalDesc, err := p.lookService.DescribePortal(ctx, target)
		if err == nil {
			client.SendGameMessage("look_result", portalDesc, nil)
			return nil
		}

		// Nothing found
		client.SendGameMessage("error", "You don't see that here.", nil)
		return nil
	}

	// No target - show general lobby description
	// 1. Description
	desc := "You are in the Grand Lobby of Thousand Worlds. A vast, ethereal hall with portals shimmering in the distance.\n\n"

	// 2. Available Worlds
	worlds, err := p.worldRepo.ListWorlds(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worlds: %w", err)
	}

	desc += "Available Worlds:\n"
	for _, w := range worlds {
		desc += fmt.Sprintf("- %s (ID: %s)\n", w.Name, w.ID)
	}
	desc += "\n"

	// 3. Other Players
	clients := p.Hub.GetClientsByWorldID(lobby.LobbyWorldID)
	if len(clients) > 1 { // Exclude self if possible, or just list all
		desc += "Other spirits here:\n"
		for _, c := range clients {
			if c.GetCharacterID() != client.GetCharacterID() {
				// Use Username if available, fallback to CharacterID
				name := c.GetUsername()
				if name == "" {
					name = c.GetCharacterID().String()
				}
				desc += fmt.Sprintf("- %s\n", name)
			}
		}
	} else {
		desc += "You are alone here.\n"
	}

	client.SendGameMessage("area_description", desc, nil)
	return nil
}

func (p *GameProcessor) handleLobbyEnter(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target world ID required")
	}

	worldIDStr := strings.TrimSpace(*cmd.Target)
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		client.SendGameMessage("error", "Invalid world ID format.", nil)
		return nil
	}

	// Verify world exists
	_, err = p.worldRepo.GetWorld(ctx, worldID)
	if err != nil {
		client.SendGameMessage("error", "World not found.", nil)
		return nil
	}

	// Trigger entry options on client
	// The client will then fetch entry options and show the modal
	client.SendGameMessage("trigger_entry_options", worldIDStr, map[string]interface{}{
		"world_id": worldIDStr,
	})
	return nil
}

func (p *GameProcessor) handleCreateWorld(ctx context.Context, client websocket.GameClient) error {
	// Trigger interview start on client
	client.SendGameMessage("start_interview", "Starting world creation interview...", nil)
	return nil
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
	// Validate message is not empty
	if cmd.Message == nil || strings.TrimSpace(*cmd.Message) == "" {
		client.SendGameMessage("error", "What do you want to say?", nil)
		return nil
	}

	message := strings.TrimSpace(*cmd.Message)
	senderUsername := client.GetUsername()
	senderCharID := client.GetCharacterID()

	// Get all clients in the lobby
	lobbyClients := p.Hub.GetClientsByWorldID(lobby.LobbyWorldID)

	// Send to sender with special formatting
	client.SendGameMessage("speech_self", fmt.Sprintf("You say, '%s'", message), map[string]interface{}{
		"sender_id":   senderCharID.String(),
		"sender_name": senderUsername,
		"message":     message,
	})

	// Broadcast to all other players in lobby
	for _, c := range lobbyClients {
		if c.GetCharacterID() != senderCharID {
			c.SendGameMessage("speech", fmt.Sprintf("%s says, '%s'", senderUsername, message), map[string]interface{}{
				"sender_id":   senderCharID.String(),
				"sender_name": senderUsername,
				"message":     message,
			})
		}
	}

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

// handleTell sends a private message to any online player
func (p *GameProcessor) handleTell(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// Validate message is not empty
	if cmd.Message == nil || strings.TrimSpace(*cmd.Message) == "" {
		client.SendGameMessage("error", "What do you want to say?", nil)
		return nil
	}

	// Validate recipient is specified
	if cmd.Recipient == nil || strings.TrimSpace(*cmd.Recipient) == "" {
		client.SendGameMessage("error", "Tell whom?", nil)
		return nil
	}

	message := strings.TrimSpace(*cmd.Message)
	targetUsername := strings.TrimSpace(*cmd.Recipient)
	senderUsername := client.GetUsername()
	senderCharID := client.GetCharacterID()

	// Get ALL connected clients (not just lobby)
	allClients := p.Hub.GetAllClients()

	// Find target client by username (case-insensitive)
	var targetClient websocket.GameClient
	targetLower := strings.ToLower(targetUsername)
	for _, c := range allClients {
		if strings.ToLower(c.GetUsername()) == targetLower {
			targetClient = c
			break
		}
	}

	// Check if target was found
	if targetClient == nil {
		client.SendGameMessage("error", "That player is not online.", nil)
		return nil
	}

	// Send to sender
	client.SendGameMessage("tell_self", fmt.Sprintf("You tell %s, '%s'", targetClient.GetUsername(), message), map[string]interface{}{
		"sender_id":    senderCharID.String(),
		"sender_name":  senderUsername,
		"recipient_id": targetClient.GetCharacterID().String(),
		"recipient":    targetClient.GetUsername(),
		"message":      message,
	})

	// Send to recipient
	targetClient.SendGameMessage("tell", fmt.Sprintf("%s tells you, '%s'", senderUsername, message), map[string]interface{}{
		"sender_id":   senderCharID.String(),
		"sender_name": senderUsername,
		"message":     message,
	})

	return nil
}

// handleWho lists all currently online players
func (p *GameProcessor) handleWho(ctx context.Context, client websocket.GameClient) error {
	if p.Hub == nil {
		return errors.New("game server not fully initialized")
	}

	worldID := client.GetWorldID()
	clients := p.Hub.GetClientsByWorldID(worldID)

	playerList := fmt.Sprintf("=== Online Players (Count: %d) ===\n", len(clients))
	for _, c := range clients {
		// Use Username if available, fallback to CharacterID
		name := c.Username
		if name == "" {
			name = c.CharacterID.String()
		}
		playerList += fmt.Sprintf("- %s\n", name)
	}

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
		"quantity": 1,
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
