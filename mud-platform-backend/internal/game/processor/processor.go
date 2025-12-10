package processor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"mud-platform-backend/cmd/game-server/websocket"
	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/game/formatter"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/player"
	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/world/interview"
)

var (
	ErrInvalidAction = errors.New("invalid action")
	ErrNoCharacter   = errors.New("no active character")
)

// GameProcessor handles game command processing
type GameProcessor struct {
	Hub              *websocket.Hub
	authRepo         auth.Repository
	worldRepo        repository.WorldRepository
	lookService      *lobby.LookService
	interviewService *interview.InterviewService
	spatialService   *player.SpatialService
}

// NewGameProcessor creates a new game processor
func NewGameProcessor(authRepo auth.Repository, worldRepo repository.WorldRepository, lookService *lobby.LookService, interviewService *interview.InterviewService, spatialService *player.SpatialService) *GameProcessor {
	return &GameProcessor{
		authRepo:         authRepo,
		worldRepo:        worldRepo,
		lookService:      lookService,
		interviewService: interviewService,
		spatialService:   spatialService,
	}
}

// SetHub sets the websocket hub
func (p *GameProcessor) SetHub(hub *websocket.Hub) {
	p.Hub = hub
}

// ProcessCommand processes a game command from a client
func (p *GameProcessor) ProcessCommand(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	log.Printf("[PROCESSOR] ProcessCommand called with Text='%s', Action='%s'", cmd.Text, cmd.Action)
	// If Text field is provided, parse it into structured format
	if cmd.Text != "" {
		parser := NewCommandParser()
		parsedCmd := parser.ParseText(cmd.Text)
		if parsedCmd == nil {
			return fmt.Errorf("invalid command")
		}
		// Use parsed command for processing
		cmd = parsedCmd
	}

	// Validate character exists
	charID := client.GetCharacterID()
	if charID == uuid.Nil {
		return ErrNoCharacter
	}

	// Check if character is in Lobby - logic removed as requested for generic handling.
	// All commands are now processed via the generic switch.

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

	// Orientation
	case "face":
		return p.handleFace(ctx, client, cmd)

	// Interaction
	case "open":
		return p.handleOpen(ctx, client, cmd)
	case "enter":
		return p.handleEnter(ctx, client, cmd)

	// Observation
	case "look", "l":
		return p.handleLook(ctx, client, cmd)

	// Communication
	case "say":
		return p.handleSay(ctx, client, cmd)
	case "whisper":
		return p.handleWhisper(ctx, client, cmd)
	case "tell":
		return p.handleTell(ctx, client, cmd)
	case "reply", "r":
		return p.handleReply(ctx, client, cmd)

	// Social
	case "who":
		return p.handleWho(ctx, client)

	// Creation / Management
	case "create":
		return p.handleCreate(ctx, client, cmd)
	case "watcher":
		return p.handleWatcher(ctx, client, cmd)

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
	case "lobby":
		return p.handleLobby(ctx, client)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidAction, cmd.Action)
	}
}

func (p *GameProcessor) handleCreate(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target != nil {
		target := strings.ToLower(*cmd.Target)
		if target == "world" {
			return p.handleCreateWorld(ctx, client)
		}
		// Handle "create character Name Role Race"
		if strings.HasPrefix(target, "character") {
			return p.handleCreateCharacter(ctx, client, cmd)
		}
	}
	return fmt.Errorf("unknown create command. Try 'create world'.")
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
	msg, err := p.spatialService.HandleMovementCommand(ctx, client.GetCharacterID(), direction)
	if err != nil {
		client.SendGameMessage("error", err.Error(), nil) // Send error to client, e.g. "blocked by wall"
		return nil
	}

	client.SendGameMessage("movement", msg, nil)
	// p.sendStateUpdate(client) // Temporarily disabled until implemented to read live DB data
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
	if cmd.Target == nil || strings.TrimSpace(*cmd.Target) == "" {
		client.SendGameMessage("error", "Enter what? Try 'enter <world name>'.", nil)
		return nil
	}

	targetStr := strings.TrimSpace(*cmd.Target)
	charID := client.GetCharacterID()
	// Get character
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		client.SendGameMessage("error", "Failed to find your character.", nil)
		return nil
	}

	// 1. Resolve target to a World
	var destWorld *repository.World
	worldID, err := uuid.Parse(targetStr)
	if err == nil {
		// Valid UUID
		destWorld, err = p.worldRepo.GetWorld(ctx, worldID)
		if err != nil {
			client.SendGameMessage("error", "That portal doesn't seem to lead anywhere.", nil)
			return nil
		}
	} else {
		// Try to find world by name
		worlds, errList := p.worldRepo.ListWorlds(ctx)
		if errList != nil {
			client.SendGameMessage("error", "Failed to search for portals.", nil)
			return nil
		}

		targetName := strings.ToLower(targetStr)
		for _, w := range worlds {
			if strings.ToLower(w.Name) == targetName {
				destWorld = &w
				break
			}
		}

		if destWorld == nil {
			client.SendGameMessage("error", fmt.Sprintf("There is no portal to '%s' here.", targetStr), nil)
			return nil
		}
	}

	// 2. Check if already in that world
	if char.WorldID == destWorld.ID {
		client.SendGameMessage("info", "You are already in that world.", nil)
		return nil
	}

	// 3. Proximity Check (Unified)
	// We check proximity to the portal for the destination world
	// Get current world for spatial calculations
	currentWorld, err := p.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		return fmt.Errorf("failed to get current world: %w", err)
	}

	// Logic: If current world has spatially mapped portals (like Lobby or others with 'bounds'), check proximity.
	// For now, we assume if we are entering a generic World from another generic World, there is a portal location.
	// We rely on SpatialService to tell us.
	// Note: Generic "Enter" might not always imply a physical portal (could be a command), but assuming "enter <world>" implies physical travel.
	isLobby := currentWorld.ID.String() == lobby.LobbyWorldID.String()
	portalX, portalY := p.spatialService.GetPortalLocation(currentWorld, destWorld.ID)
	allowed, hint := p.spatialService.CheckPortalProximity(char.PositionX, char.PositionY, portalX, portalY, isLobby)
	if !allowed {
		client.SendGameMessage("error", hint, nil)
		return nil
	}

	// 4. Trigger Entry Options (Unified Flow)
	// Instead of auto-moving, we trigger the entry options modal.
	// This allows the user to choose "New Character" or "Existing Character" or "Watcher".
	// This unifies the flow so "enter <world>" behaves the same everywhere.
	client.SendGameMessage("trigger_entry_options", destWorld.ID.String(), map[string]interface{}{
		"world_id": destWorld.ID.String(),
	})

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
	formattedMessage := fmt.Sprintf("You say, %s", formatter.Format(fmt.Sprintf("'%s'", message), formatter.StyleGreen))
	client.SendGameMessage("speech_self", formattedMessage, map[string]interface{}{
		"sender_id":   senderCharID.String(),
		"sender_name": senderUsername,
		"message":     message,
	})

	// Broadcast to all other players in lobby
	for _, c := range lobbyClients {
		if c.GetCharacterID() != senderCharID {
			formattedSpeech := fmt.Sprintf("%s says, %s",
				formatter.Format(senderUsername, formatter.StyleCyan),
				formatter.Format(fmt.Sprintf("'%s'", message), formatter.StyleGreen))
			c.SendGameMessage("speech", formattedSpeech, map[string]interface{}{
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

	// Special handling for "statue" - this is the world creation NPC in the lobby
	if strings.ToLower(targetUsername) == "statue" {
		log.Printf("[STATUE] User %s (ID: %s) talking to statue with message: %s", senderUsername, client.GetUserID(), message)

		// User ID from client
		userID := client.GetUserID()

		// Try to get existing interview first
		session, err := p.interviewService.GetActiveInterview(ctx, userID)
		log.Printf("[STATUE] GetActiveInterview result: session=%v, err=%v", session != nil, err)

		if err != nil || session == nil {
			log.Printf("[STATUE] Starting new interview for user %s", userID)

			// Send thinking emote
			client.SendGameMessage("emote", "The statue stands motionless, its stone gaze seeming to penetrate your mind...", map[string]interface{}{
				"source": "Statue",
			})

			// No active interview - start a new one
			session, question, err := p.interviewService.StartInterview(ctx, userID)
			if err != nil {
				log.Printf("[STATUE] ERROR starting interview: %v", err)
				client.SendGameMessage("error", fmt.Sprintf("The statue remains silent. %v", err), nil)
				return err
			}

			log.Printf("[STATUE] Interview started successfully, sending first question")

			// Send emote
			client.SendGameMessage("emote", "The statue's eyes glow warmly.", map[string]interface{}{
				"source": "Statue",
			})

			// Send the initial question from the statue
			response := fmt.Sprintf("A voice resonates in your mind:\n\n%s", question)
			client.SendGameMessage("tell", response, map[string]interface{}{
				"sender_name": "Statue",
				"session_id":  session.ID.String(),
			})
		} else {
			log.Printf("[STATUE] Processing response for existing interview")

			// Send thinking emote
			client.SendGameMessage("emote", "The statue listens intently...", map[string]interface{}{
				"source": "Statue",
			})

			// Interview in progress - process the message as a response
			nextQuestion, isComplete, err := p.interviewService.ProcessResponse(ctx, userID, message)
			if err != nil {
				log.Printf("[STATUE] ERROR processing response: %v", err)
				client.SendGameMessage("error", fmt.Sprintf("The statue seems confused. %v", err), nil)
				return err
			}

			if isComplete {
				log.Printf("[STATUE] Interview complete for user %s", userID)
				// Interview complete
				// The service now includes the enter instruction in the response
				response := fmt.Sprintf("The statue's eyes shine with approval.\n\n%s", nextQuestion)
				client.SendGameMessage("tell", response, map[string]interface{}{
					"sender_name": "Statue",
					"completed":   true,
				})
			} else {
				log.Printf("[STATUE] Sending next question")
				// Send next question
				response := fmt.Sprintf("The statue acknowledges your answer with a nod.\n\n%s", nextQuestion)
				client.SendGameMessage("tell", response, map[string]interface{}{
					"sender_name": "Statue",
				})
			}
		}

		log.Printf("[STATUE] Completed statue interaction for user %s", userID)
		// Update last tell sender so they can reply
		client.SetLastTellSender("statue")
		return nil
	}

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

	// Update recipient's last tell sender so they can use reply command
	targetClient.SetLastTellSender(senderUsername)

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
	// If target is specified, it might be an item or feature.
	// For now, let's assume we are mostly looking at the room if no target or "here".
	if cmd.Target != nil && *cmd.Target != "" && strings.ToLower(*cmd.Target) != "here" && strings.ToLower(*cmd.Target) != "room" {
		description := fmt.Sprintf("You examine the %s.", *cmd.Target)
		client.SendGameMessage("area_description", description, nil)
		return nil
	}

	charID := client.GetCharacterID()
	worldID := client.GetWorldID()

	// Fetch character to get position data
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		log.Printf("[PROCESSOR] Failed to get character for look: %v", err)
		return err
	}

	// Get dynamic room description from LookService
	description, err := p.lookService.DescribeRoom(ctx, worldID, char)
	if err != nil {
		log.Printf("[PROCESSOR] Failed to describe room: %v", err)
		description = "You are in a mysterious place. The mist conceals everything."
	}

	client.SendGameMessage("area_description", description, map[string]interface{}{
		"character_id": charID.String(),
		"world_id":     worldID.String(),
	})
	return nil
}

func (p *GameProcessor) handleTake(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target item required")
	}

	// TODO: Integrate with inventory system
	formattedItem := fmt.Sprintf("You pick up %s.", formatter.Item(*cmd.Target, "common"))
	client.SendGameMessage("item_acquired", formattedItem, map[string]interface{}{
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
	formattedCombat := fmt.Sprintf("You attack %s for %s damage!",
		formatter.Target(*cmd.Target),
		formatter.Damage(10))
	client.SendGameMessage("combat", formattedCombat, map[string]interface{}{
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

	// TODO: Integrate with NPC dialogue system
	formattedDialogue := fmt.Sprintf("%s says: %s",
		formatter.Format(*cmd.Target, formatter.StyleYellow),
		formatter.Format("'Hello, traveler!'", formatter.StyleGreen))
	client.SendGameMessage("dialogue", formattedDialogue, map[string]interface{}{
		"npcName":   *cmd.Target,
		"npcID":     "placeholder-npc-id",
		"dialogue":  "Hello, traveler!",
		"available": []string{"quest", "trade", "goodbye"},
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
	// Get character from database for live position
	charID := client.GetCharacterID()
	ctx := context.Background()
	char, err := p.authRepo.GetCharacter(ctx, charID)

	posX, posY := 0.0, 0.0
	if err == nil && char != nil {
		posX = char.PositionX
		posY = char.PositionY
	}

	// TODO: Get actual HP/Stamina/Focus from character stats when implemented
	state := &websocket.StateUpdateData{
		HP:         100,
		MaxHP:      100,
		Stamina:    80,
		MaxStamina: 100,
		Focus:      60,
		MaxFocus:   100,
		Position: websocket.Position{
			X: posX,
			Y: posY,
		},
		Inventory:    []websocket.InventoryItem{},
		Equipment:    websocket.Equipment{},
		VisibleTiles: []websocket.VisibleTile{},
	}

	client.SendStateUpdate(state)
}

func (p *GameProcessor) handleCreateCharacter(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("missing character details")
	}

	// Target format: "character Name Role Race"
	parts := strings.Split(*cmd.Target, " ")
	if len(parts) < 4 {
		return errors.New("usage: create character <Name> <Role> <Race>")
	}

	name := parts[1]
	role := parts[2]
	race := parts[3] // We'll store race in Description for now as Appearance creates complexity

	charID := uuid.New()
	userID := client.GetUserID()
	// Create characters in the lobby world by default if not specified
	worldID := lobby.LobbyWorldID

	// If 5th argument provided, parse as world ID
	if len(parts) >= 5 {
		parsedWorldID, err := uuid.Parse(parts[4])
		if err == nil {
			worldID = parsedWorldID
		}
	}

	now := time.Now()
	// Create character struct
	newChar := &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     worldID,
		Name:        name,
		Role:        role,
		Description: fmt.Sprintf("Race: %s", race),
		CreatedAt:   now,
		LastPlayed:  &now,
		Position: &auth.Position{
			Latitude:  0,
			Longitude: 0,
		},
	}

	if err := p.authRepo.CreateCharacter(ctx, newChar); err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	// Update client state
	client.SetCharacterID(charID)

	client.SendGameMessage("system", fmt.Sprintf("Character %s created successfully! You are a %s %s.", name, race, role), map[string]interface{}{
		"character_id": charID.String(),
		"name":         name,
		"role":         role,
		"race":         race,
	})

	return nil
}

func (p *GameProcessor) handleWatcher(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil || *cmd.Target == "" {
		return errors.New("missing world ID")
	}

	// Try UUID parse first
	worldID, err := uuid.Parse(*cmd.Target)
	if err != nil {
		// Not a UUID, try to find world by name
		targetName := strings.ToLower(strings.TrimSpace(*cmd.Target))
		worlds, errList := p.worldRepo.ListWorlds(ctx)
		if errList != nil {
			client.SendGameMessage("error", "Failed to search for world.", nil)
			return nil
		}

		var foundWorld *repository.World
		for _, w := range worlds {
			if strings.ToLower(w.Name) == targetName {
				foundWorld = &w
				break
			}
		}

		if foundWorld != nil {
			worldID = foundWorld.ID
		} else {
			return fmt.Errorf("invalid world ID or name: %w", err)
		}
	}

	userID := client.GetUserID()
	// Check if already has a character in this world
	existingChar, err := p.authRepo.GetCharacterByUserAndWorld(ctx, userID, worldID)
	// If method not implemented or error, ignore for now and try create?
	// The interface has GetCharacterByUserAndWorld.
	// But let's assume if it exists, we just switch to it.
	if err == nil && existingChar != nil {
		client.SetCharacterID(existingChar.CharacterID)
		client.SetWorldID(worldID)
		client.SendGameMessage("system", "You return to the world as a watcher.", map[string]interface{}{
			"character_id": existingChar.CharacterID.String(),
			"role":         "watcher",
			"world_id":     worldID.String(),
		})
		return nil
	}

	// Create watcher character
	charID := uuid.New()
	now := time.Now()
	newChar := &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     worldID,
		Name:        "Watcher",
		Role:        "watcher",
		Description: "An incorporeal observer.",
		CreatedAt:   now,
		LastPlayed:  &now,
		Position: &auth.Position{
			Latitude:  0,
			Longitude: 0,
		},
	}

	if err := p.authRepo.CreateCharacter(ctx, newChar); err != nil {
		return fmt.Errorf("failed to create watcher character: %w", err)
	}

	client.SetCharacterID(charID)
	// IMPORTANT: Update the client's WorldID so subsequent commands are routed to the world, not lobby
	client.SetWorldID(worldID)

	client.SendGameMessage("system", "You enter the world as a watcher.", map[string]interface{}{
		"character_id": charID.String(),
		"role":         "watcher",
		"world_id":     worldID.String(),
	})

	return nil
}
