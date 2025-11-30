package websocket

import (
	// Add missing import
	"log"
	"net/http"

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
	Hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		Hub: hub,
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
