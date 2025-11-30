package websocket

import (
	"log"
	"net/http"

	"mud-platform-backend/internal/lobby"

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
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, lobbyService *lobby.Service) *Handler {
	return &Handler{
		Hub:          hub,
		LobbyService: lobbyService,
	}
}

// ServeHTTP upgrades HTTP connections to WebSocket
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
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

	// If no character ID, join lobby
	if characterID == uuid.Nil {
		lobbyChar, err := h.LobbyService.EnsureLobbyCharacter(r.Context(), userID)
		if err != nil {
			log.Printf("Failed to join lobby: %v", err)
			http.Error(w, "Failed to join lobby", http.StatusInternalServerError)
			return
		}
		characterID = lobbyChar.CharacterID
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create client
	client := NewClient(h.Hub, conn, userID, characterID)

	// Register client
	h.Hub.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	if characterID != uuid.Nil {
		log.Printf("WebSocket connection established for user %s, character %s", userID, characterID)
	} else {
		log.Printf("WebSocket connection established for user %s (no character)", userID)
	}
}
