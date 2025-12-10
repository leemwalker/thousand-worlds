package websocket

import (
	"log"
	"net/http"
	"time"

	"tw-backend/internal/auth"
	"tw-backend/internal/lobby"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking for production
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Handler handles WebSocket upgrade requests
type Handler struct {
	Hub          *Hub
	LobbyService *lobby.Service
	AuthRepo     auth.Repository
	DescGen      *lobby.DescriptionGenerator
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, lobbyService *lobby.Service, authRepo auth.Repository, descGen *lobby.DescriptionGenerator) *Handler {
	return &Handler{
		Hub:          hub,
		LobbyService: lobbyService,
		AuthRepo:     authRepo,
		DescGen:      descGen,
	}
}

// ServeHTTP upgrades HTTP connections to WebSocket
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[WS] Incoming WebSocket connection request from %s", r.RemoteAddr)

	// Get user ID from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Printf("[WS] ERROR: No userID in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[WS] User authenticated: %s", userIDStr)

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[WS] ERROR: Invalid user ID format: %s", userIDStr)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get character ID from query parameter
	characterID := uuid.Nil
	if charIDStr := r.URL.Query().Get("character_id"); charIDStr != "" {
		if parsed, err := uuid.Parse(charIDStr); err == nil {
			characterID = parsed
		}
	}

	log.Printf("[WS] Character ID from query: %s", characterID)

	// If no character ID, join lobby
	if characterID == uuid.Nil {
		log.Printf("[WS] No character ID, joining lobby for user %s", userID)
		lobbyChar, err := h.LobbyService.EnsureLobbyCharacter(r.Context(), userID)
		if err != nil {
			log.Printf("[WS] Failed to join lobby: %v", err)
			http.Error(w, "Failed to join lobby", http.StatusInternalServerError)
			return
		}
		characterID = lobbyChar.CharacterID
		log.Printf("[WS] Lobby character: %s", characterID)
	}

	// Fetch character to get WorldID
	char, err := h.AuthRepo.GetCharacter(r.Context(), characterID)
	if err != nil {
		log.Printf("[WS] Failed to fetch character: %v", err)
		http.Error(w, "Failed to fetch character", http.StatusInternalServerError)
		return
	}

	log.Printf("[WS] Character fetched successfully: %s (world: %s)", char.CharacterID, char.WorldID)

	// Fetch user to get username
	user, err := h.AuthRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("[WS] Failed to fetch user: %v", err)
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Failed to upgrade connection: %v", err)
		return
	}

	log.Printf("[WS] Connection upgraded successfully")

	// Create client
	client := NewClient(h.Hub, conn, userID, characterID, char.WorldID, user.Username)

	// Register client
	h.Hub.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	if characterID != uuid.Nil {
		log.Printf("[WS] WebSocket connection established for user %s, character %s, world %s", userID, characterID, char.WorldID)
	} else {
		log.Printf("[WS] WebSocket connection established for user %s (no character)", userID)
	}

	// WELCOME LOGIC
	isFirstTime := user.LastLogin == nil

	// Only send lobby welcome messages if actually in the lobby
	if lobby.IsLobby(char.WorldID) {
		// Send Welcome Message
		if isFirstTime {
			welcomeMsg := "Welcome to Thousand Worlds, you are in the lobby where you can meet other players who aren't in a world, see the worlds you can visit, and create your own world! Please *look* around at other players and portals to the worlds, *say* hello, and don't forget to *look statue*"
			client.SendGameMessage("system", welcomeMsg, nil)
		} else {
			client.SendGameMessage("system", "Welcome to the Lobby.", nil)
		}
	} else {
		// If in a world, we might want to send a brief welcome or just let the description handle it
		// For now, let's just log it
		log.Printf("[WS] User %s entering world %s", user.Username, char.WorldID)
	}

	// Send Lobby Description if in Lobby
	if lobby.IsLobby(char.WorldID) {
		hubClients := h.Hub.GetClientsByWorldID(lobby.LobbyWorldID)

		// Adapt clients to lobby.WebsocketClient interface
		var lobbyClients []lobby.WebsocketClient
		for _, c := range hubClients {
			lobbyClients = append(lobbyClients, c)
		}

		desc, err := h.DescGen.GenerateDescription(r.Context(), user, char, lobbyClients)
		if err != nil {
			log.Printf("[WS] Failed to generate lobby description: %v", err)
		} else {
			client.SendGameMessage("area_description", desc, nil)
		}
	}

	// Update LastLogin
	now := time.Now()
	user.LastLogin = &now
	// We don't update LastWorldID here, that happens on Enter.
	// But if they are in Lobby, LastWorldID should probably remain what it was (for atmosphere),
	// or maybe we don't update it until they leave the lobby?
	// The requirement says "The base lobby description should be influenced by the last world the player visited".
	// So we should NOT overwrite LastWorldID with LobbyID.

	if err := h.AuthRepo.UpdateUser(r.Context(), user); err != nil {
		log.Printf("[WS] Failed to update user last login: %v", err)
	}
}
