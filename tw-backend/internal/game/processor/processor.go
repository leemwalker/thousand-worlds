package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/character"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/game/constants"
	"tw-backend/internal/game/formatter"
	"tw-backend/internal/game/services/combat"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/game/services/inventory"
	"tw-backend/internal/game/services/look"
	gamemap "tw-backend/internal/game/services/map"
	"tw-backend/internal/player"
	"tw-backend/internal/repository"
	"tw-backend/internal/skills"
	"tw-backend/internal/world/interview"
	"tw-backend/internal/worldentity"
	"tw-backend/internal/worldgen/weather"
)

var (
	ErrInvalidAction = errors.New("invalid action")
	ErrNoCharacter   = errors.New("no active character")
)

// GameProcessor handles game command processing
type GameProcessor struct {
	Hub                *websocket.Hub
	authRepo           auth.Repository
	worldRepo          repository.WorldRepository
	characterRepo      character.CharacterRepository // Full character data (event sourced)
	lookService        *look.LookService
	entityService      *entity.Service
	interviewService   *interview.InterviewService
	spatialService     *player.SpatialService
	weatherService     *weather.Service
	mapService         *gamemap.Service
	skillsRepo         skills.Repository
	worldEntityService *worldentity.Service
	ecosystemService   *ecosystem.Service
	combatService      *combat.Service
	inventoryService   *inventory.Service

	// WorldGeology stores geological state per world (worldID -> geology)
	worldGeology map[uuid.UUID]*ecosystem.WorldGeology

	// worldRunners stores async simulation runners per world
	worldRunners map[uuid.UUID]*ecosystem.SimulationRunner
}

// NewGameProcessor creates a new game processor
func NewGameProcessor(
	authRepo auth.Repository,
	worldRepo repository.WorldRepository,
	characterRepo character.CharacterRepository,
	lookService *look.LookService,
	entityService *entity.Service,
	interviewService *interview.InterviewService,
	spatialService *player.SpatialService,
	weatherService *weather.Service,
	skillsRepo skills.Repository,
	worldEntityService *worldentity.Service,
	ecosystemService *ecosystem.Service,
	combatService *combat.Service,
	inventoryService *inventory.Service,
) *GameProcessor {
	// Create map service with available dependencies
	mapSvc := gamemap.NewService(worldRepo, skillsRepo, entityService, lookService, worldEntityService, ecosystemService)

	return &GameProcessor{
		authRepo:           authRepo,
		worldRepo:          worldRepo,
		characterRepo:      characterRepo,
		lookService:        lookService,
		entityService:      entityService,
		interviewService:   interviewService,
		spatialService:     spatialService,
		weatherService:     weatherService,
		mapService:         mapSvc,
		skillsRepo:         skillsRepo,
		worldEntityService: worldEntityService,
		ecosystemService:   ecosystemService,
		combatService:      combatService,
		inventoryService:   inventoryService,
		worldGeology:       make(map[uuid.UUID]*ecosystem.WorldGeology),
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

	// Cardinal directions (pass cmd for watcher distance movement)
	case "north", "n":
		return p.handleDirection(ctx, client, cmd, "north")
	case "northeast", "ne":
		return p.handleDirection(ctx, client, cmd, "northeast")
	case "east", "e":
		return p.handleDirection(ctx, client, cmd, "east")
	case "southeast", "se":
		return p.handleDirection(ctx, client, cmd, "southeast")
	case "south", "s":
		return p.handleDirection(ctx, client, cmd, "south")
	case "southwest", "sw":
		return p.handleDirection(ctx, client, cmd, "southwest")
	case "west", "w":
		return p.handleDirection(ctx, client, cmd, "west")
	case "northwest", "nw":
		return p.handleDirection(ctx, client, cmd, "northwest")
	case "up", "u":
		return p.handleDirection(ctx, client, cmd, "up")
	case "down", "d":
		return p.handleDirection(ctx, client, cmd, "down")

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

	// Object interaction commands (get/take/grab/pick all resolve to "get")
	case "get":
		return p.handleGetObject(ctx, client, cmd)
	case "push":
		return p.handlePushObject(ctx, client, cmd)

	// Existing commands
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
	case "weather":
		return p.handleWeather(ctx, client, cmd)
	case "ecosystem":
		return p.handleEcosystem(ctx, client, cmd)
	case "world":
		return p.handleWorld(ctx, client, cmd)
	case "fly":
		return p.handleFly(ctx, client, cmd)

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
	return fmt.Errorf("unknown create command: try 'create world'")
}

func (p *GameProcessor) handleCreateWorld(_ context.Context, client websocket.GameClient) error {
	// Trigger interview start on client
	client.SendGameMessage("start_interview", "Starting world creation interview...", nil)
	return nil
}

// Command handlers
func (p *GameProcessor) handleHelp(_ context.Context, client websocket.GameClient) error {
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
// Watchers can move long distances using format: "<direction> <distance>" (e.g., "w 500")
func (p *GameProcessor) handleDirection(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData, direction string) error {
	charID := client.GetCharacterID()

	// Check if watcher with distance specified
	distance := 1 // Default movement distance
	if cmd != nil && cmd.Target != nil {
		// Check if character is a watcher
		char, err := p.authRepo.GetCharacter(ctx, charID)
		if err == nil && char != nil && char.Role == "watcher" {
			// Parse distance from target
			if parsedDist, parseErr := strconv.Atoi(*cmd.Target); parseErr == nil && parsedDist > 0 {
				distance = parsedDist
				// Cap distance at reasonable maximum
				if distance > 10000 {
					distance = 10000
				}
			}
		} else if cmd.Target != nil {
			// Non-watcher tried to specify distance
			client.SendGameMessage("info", "Only watchers can travel great distances instantly.", nil)
		}
	}

	// Move the specified distance
	if distance > 1 {
		// Watcher long-distance movement
		msg, err := p.spatialService.HandleMovementCommandWithDistance(ctx, charID, direction, float64(distance))
		if err != nil {
			client.SendGameMessage("error", err.Error(), nil)
			return nil
		}
		client.SendGameMessage("movement", msg, nil)
	} else {
		// Normal single-step movement
		msg, err := p.spatialService.HandleMovementCommand(ctx, charID, direction)
		if err != nil {
			client.SendGameMessage("error", err.Error(), nil)
			return nil
		}
		client.SendGameMessage("movement", msg, nil)
	}

	// Send map update after movement
	p.sendMapUpdate(ctx, client)

	return nil
}

// getDirectionVector returns the x,y offset for a direction
func (p *GameProcessor) getDirectionVector(direction string) (int, int) {
	switch direction {
	case "north":
		return 0, 1
	case "south":
		return 0, -1
	case "east":
		return 1, 0
	case "west":
		return -1, 0
	case "northeast":
		return 1, 1
	case "northwest":
		return -1, 1
	case "southeast":
		return 1, -1
	case "southwest":
		return -1, -1
	default:
		return 0, 0
	}
}

// handleOpen opens doors or containers
func (p *GameProcessor) handleOpen(_ context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
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
		client.SendGameMessage("error", "Enter what? Try 'enter <portal name>' or 'enter <world name>'.", nil)
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

	// 1. First check if target matches a nearby portal entity with destination_world_id
	var destWorld *repository.World
	if p.worldEntityService != nil {
		entities, err := p.worldEntityService.GetEntitiesAt(ctx, char.WorldID, char.PositionX, char.PositionY, 5.0)
		if err == nil {
			targetLower := strings.ToLower(targetStr)
			for _, e := range entities {
				// Check if entity name matches target
				if !strings.Contains(strings.ToLower(e.Name), targetLower) {
					continue
				}
				// Check if it has a destination_world_id in metadata
				if e.Metadata != nil {
					if destID, ok := e.Metadata["destination_world_id"].(string); ok {
						worldID, parseErr := uuid.Parse(destID)
						if parseErr == nil {
							destWorld, err = p.worldRepo.GetWorld(ctx, worldID)
							if err == nil && destWorld != nil {
								// Found portal with valid destination
								break
							}
						}
					}
				}
			}
		}
	}

	// 2. If no portal entity found, fall back to world name/UUID lookup
	if destWorld == nil {
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
	isLobby := constants.IsLobby(currentWorld.ID)
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
func (p *GameProcessor) handleSay(_ context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// Validate message is not empty
	if cmd.Message == nil || strings.TrimSpace(*cmd.Message) == "" {
		client.SendGameMessage("error", "What do you want to say?", nil)
		return nil
	}

	message := strings.TrimSpace(*cmd.Message)
	senderUsername := client.GetUsername()
	senderCharID := client.GetCharacterID()

	// Get all clients in the lobby
	lobbyClients := p.Hub.GetClientsByWorldID(constants.LobbyWorldID)

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
func (p *GameProcessor) handleWhisper(_ context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
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
func (p *GameProcessor) handleWho(_ context.Context, client websocket.GameClient) error {
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
	target := ""
	if cmd.Target != nil {
		target = *cmd.Target
	}

	charID := client.GetCharacterID()
	worldID := client.GetWorldID()

	// Fetch character to get position data
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		log.Printf("[PROCESSOR] Failed to get character for look: %v", err)
		return err
	}

	// Determine orientation name from vector if not stored?
	// SpatialService has helper for this.
	orientation := p.spatialService.GetDirectionName(char.OrientationX, char.OrientationY, char.OrientationZ)

	// If specific target (not room/here)
	if target != "" && strings.ToLower(target) != "here" && strings.ToLower(target) != "room" {
		description, err := p.lookService.DescribeEntity(ctx, char, target)
		if err != nil {
			client.SendGameMessage("error", fmt.Sprintf("You don't see any '%s' here.", target), nil)
			return nil
		}
		client.SendGameMessage("area_description", description, nil)
		return nil
	}

	// Describe Room
	dc := look.DescribeContext{
		WorldID:     worldID,
		Character:   char,
		Orientation: orientation,
		DetailLevel: 1, // Default to basic
	}

	description, err := p.lookService.Describe(ctx, dc)
	if err != nil {
		log.Printf("[PROCESSOR] Failed to describe room: %v", err)
		description = "You are in a mysterious place. The mist conceals everything."
	}

	client.SendGameMessage("area_description", description, map[string]interface{}{
		"character_id": charID.String(),
		"world_id":     worldID.String(),
	})

	// Also send map update when looking at the room
	p.sendMapUpdate(ctx, client)

	return nil
}

func (p *GameProcessor) handleDrop(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target item required")
	}

	charID := client.GetCharacterID()
	itemName := *cmd.Target

	// Remove from inventory
	item, err := p.inventoryService.RemoveItem(ctx, charID, itemName)
	if err != nil {
		client.SendGameMessage("error", fmt.Sprintf("You don't have a %s.", itemName), nil)
		return nil
	}

	// Add to world entities
	authChar, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		return fmt.Errorf("failed to get character location: %w", err)
	}

	// Create world entity
	droppedEntity := worldentity.WorldEntity{
		ID:           uuid.New(),
		WorldID:      authChar.WorldID,
		Name:         item.Name,
		Description:  item.Description,
		EntityType:   worldentity.EntityTypeItem,
		X:            authChar.PositionX,
		Y:            authChar.PositionY,
		Z:            authChar.PositionZ,
		Interactable: true,
	}

	if err := p.worldEntityService.Create(ctx, &droppedEntity); err != nil {
		log.Printf("Failed to create dropped entity: %v", err)
		return fmt.Errorf("failed to drop item")
	}

	client.SendGameMessage("system", fmt.Sprintf("You drop the %s.", item.Name), nil)
	p.sendStateUpdate(client)
	return nil
}

func (p *GameProcessor) handleInventory(ctx context.Context, client websocket.GameClient) error {
	charID := client.GetCharacterID()
	items, err := p.inventoryService.GetInventory(ctx, charID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	if len(items) == 0 {
		client.SendGameMessage("system", "Your inventory is empty.", nil)
		return nil
	}

	var itemNames []string
	for _, i := range items {
		itemNames = append(itemNames, fmt.Sprintf("- %s (%d)", i.Name, i.Quantity))
	}

	msg := fmt.Sprintf("Inventory:\n%s", strings.Join(itemNames, "\n"))
	client.SendGameMessage("system", msg, nil)
	return nil
}

// Tick processes periodic game updates (combat)
func (p *GameProcessor) Tick(dt time.Duration) {
	events := p.combatService.Tick(dt)
	for _, evt := range events {
		// Broadcast combat events
		// precise broadcasting would require mapped locations, here we broadcast to world or participants
		// For P0, we broadcast to the participants

		data := evt.Data
		actorIDStr, _ := data["actor_id"].(uuid.UUID)
		targetIDStr, _ := data["target_id"].(uuid.UUID)

		// Simplify: Find clients for these IDs and send message
		// In real impl, we'd use p.Hub.GetClientByCharacterID() if it existed, or loop

		// Construct message
		var msg string
		if evt.Type == "combat_action" {
			actionType := data["type"].(string)
			msg = fmt.Sprintf("Combat: %s performs %s on %s", actorIDStr, actionType, targetIDStr)
		}

		// Send to world (spammy but works for P0 verification)
		// Better: Send to Hub broadcast for that world? We don't have event location easily available in event yet
		// We can get location from actor if we looked them up.

		// For now, just log and try to notify players if possible
		log.Printf("[COMBAT-EVENT] %s: %s", evt.Type, msg)

		// Broadcast to all clients (global broadcast for debug/proof)
		// p.Hub.Broadcast(websocket.GameMessage{Type: "combat_event", Payload: msg})
	}
}

func (p *GameProcessor) handleAttack(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("target required for attack")
	}

	targetName := strings.ToLower(*cmd.Target)
	attackerID := client.GetCharacterID()

	// Try to get full character data (with attributes)
	attackerChar, err := p.characterRepo.Load(ctx, attackerID)
	if err != nil {
		// FALLBACK: If event-sourced character not found, load basic auth data and mock attributes
		authChar, authErr := p.authRepo.GetCharacter(ctx, attackerID)
		if authErr != nil {
			return authErr
		}

		// Use default template (Human) since Species is not yet in Auth DB
		// TODO: Add Species to Auth DB or migrate fully to Event Sourcing
		template := character.GetSpeciesTemplate(character.SpeciesHuman)
		attackerChar = &character.Character{
			ID:        authChar.CharacterID,
			Name:      authChar.Name,
			BaseAttrs: template.BaseAttrs,
			SecAttrs:  character.CalculateSecondaryAttributes(template.BaseAttrs),
		}
	}

	// Ensure attacker is in combat state
	p.combatService.JoinCombatFromCharacter(attackerChar)

	// fast lookup: is target a player?
	// Get clients in same world
	var targetClientID uuid.UUID
	// Use authChar worldID if attackerChar might be fresh/empty on some fields?
	// attackerChar from Load should have WorldID if event sourced properly?
	// The Load() implementation rehydrates from events. events have PlayerID/Name/Attributes.
	// They might NOT have WorldID updates (since movement is likely not event sourced yet?).
	// Safest is to rely on client/auth state for location.

	// Get worldID from client state or authRepo
	authChar, _ := p.authRepo.GetCharacter(ctx, attackerID)
	if authChar == nil {
		return errors.New("failed to resolve attacker location")
	}

	roomClients := p.Hub.GetClientsByWorldID(authChar.WorldID)
	var targetChar *character.Character

	for _, c := range roomClients {
		if strings.ToLower(c.GetUsername()) == targetName {
			tID := c.GetCharacterID()
			targetClientID = tID

			// Load target full char
			tChar, err := p.characterRepo.Load(ctx, tID)
			if err != nil {
				// Fallback for target
				tAuthChar, tAuthErr := p.authRepo.GetCharacter(ctx, tID)
				if tAuthErr == nil {
					template := character.GetSpeciesTemplate(character.SpeciesHuman)
					tChar = &character.Character{
						ID:        tAuthChar.CharacterID,
						Name:      tAuthChar.Name,
						BaseAttrs: template.BaseAttrs,
						SecAttrs:  character.CalculateSecondaryAttributes(template.BaseAttrs),
					}
				}
			}
			targetChar = tChar
			break
		}
	}

	if targetChar != nil {
		// Target is a player
		p.combatService.JoinCombatFromCharacter(targetChar)
		err := p.combatService.QueueAttack(attackerID, targetClientID)
		if err != nil {
			client.SendGameMessage("error", fmt.Sprintf("Failed to attack: %v", err), nil)
			return nil
		}

		client.SendGameMessage("combat", fmt.Sprintf("You attack %s!", targetChar.Name), nil)
		return nil
	}

	// TODO: Handle NPC targets
	return fmt.Errorf("target '%s' not found", *cmd.Target)
}

func (p *GameProcessor) handleTalk(_ context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
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

func (p *GameProcessor) handleCraft(_ context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
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

func (p *GameProcessor) handleUse(_ context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
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

// sendMapUpdate sends the mini-map data to the client
func (p *GameProcessor) sendMapUpdate(ctx context.Context, client websocket.GameClient) {
	if p.mapService == nil {
		return
	}

	charID := client.GetCharacterID()
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil || char == nil {
		return
	}

	mapData, err := p.mapService.GetMapData(ctx, char)
	if err != nil {
		log.Printf("[PROCESSOR] Failed to get map data: %v", err)
		return
	}

	// Convert MapData to map for SendGameMessage
	// Using JSON marshal/unmarshal for type conversion
	jsonBytes, err := json.Marshal(mapData)
	if err != nil {
		log.Printf("[PROCESSOR] Failed to serialize map data: %v", err)
		return
	}

	var mapPayload map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &mapPayload); err != nil {
		log.Printf("[PROCESSOR] Failed to convert map data: %v", err)
		return
	}

	// Send map_update message
	log.Printf("[PROCESSOR] Sending map_update to client with %d tiles at position (%.1f, %.1f), quality=%s",
		len(mapData.Tiles), mapData.PlayerX, mapData.PlayerY, mapData.RenderQuality)
	client.SendGameMessage("map_update", "", mapPayload)
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
	worldID := constants.LobbyWorldID

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

// handleWeather allows forcing weather states (God Mode)
func (p *GameProcessor) handleWeather(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		return errors.New("usage: weather <storm|rain|snow|clear>")
	}

	weatherTypeStr := strings.ToLower(*cmd.Target)

	// Map string to WeatherState enum
	var weatherState weather.WeatherType
	switch weatherTypeStr {
	case "clear", "sunny":
		weatherState = weather.WeatherClear
	case "cloudy", "overcast":
		weatherState = weather.WeatherCloudy
	case "rain", "rainy":
		weatherState = weather.WeatherRain
	case "storm":
		weatherState = weather.WeatherStorm
	case "snow", "snowy":
		weatherState = weather.WeatherSnow
	default:
		return fmt.Errorf("unknown weather type: %s", weatherTypeStr)
	}

	worldID := client.GetWorldID()

	// For now, we apply this to the WHOLE world.
	// But the service logic usually works per cell.
	// We'll add a ForceWeather method to the service that handles this.

	if p.weatherService == nil {
		return errors.New("weather service not available")
	}

	if err := p.weatherService.ForceWorldWeather(ctx, worldID, weatherState); err != nil {
		return fmt.Errorf("failed to set weather: %w", err)
	}

	client.SendGameMessage("system", fmt.Sprintf("Weather changed to %s.", weatherTypeStr), nil)
	return nil
}
