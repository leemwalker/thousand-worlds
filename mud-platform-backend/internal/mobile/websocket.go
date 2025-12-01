package mobile

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// GameWebSocket represents a WebSocket client for real-time game communication
type GameWebSocket struct {
	baseURL  string
	token    string
	conn     *websocket.Conn
	handlers []func(*ServerMessage)
	mu       sync.RWMutex
	done     chan struct{}
}

// Command represents a game command to send to the server
type Command struct {
	Action    string `json:"action"`
	Target    string `json:"target,omitempty"`
	Message   string `json:"message,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Quantity  int    `json:"quantity,omitempty"`
}

// ServerMessage represents a message received from the server
type ServerMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NewGameWebSocket creates a new WebSocket client
func NewGameWebSocket(baseURL, token string) *GameWebSocket {
	return &GameWebSocket{
		baseURL:  baseURL,
		token:    token,
		handlers: make([]func(*ServerMessage), 0),
		done:     make(chan struct{}),
	}
}

// Connect establishes a WebSocket connection to the game server
func (g *GameWebSocket) Connect(ctx context.Context) error {
	// Build WebSocket URL with token
	u, err := url.Parse(g.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	// Set path to WebSocket endpoint
	u.Path = "/api/game/ws"

	// Add token as query parameter
	q := u.Query()
	q.Set("token", g.token)
	u.RawQuery = q.Encode()

	// Establish connection
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	g.mu.Lock()
	g.conn = conn
	g.done = make(chan struct{})
	g.mu.Unlock()

	// Start reading messages
	go g.readMessages()

	return nil
}

// Disconnect closes the WebSocket connection
func (g *GameWebSocket) Disconnect() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.conn == nil {
		return nil
	}

	// Signal done
	close(g.done)

	// Close connection
	err := g.conn.Close()
	g.conn = nil

	return err
}

// IsConnected returns whether the WebSocket is currently connected
func (g *GameWebSocket) IsConnected() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.conn != nil
}

// SendCommand sends a game command to the server
func (g *GameWebSocket) SendCommand(cmd *Command) error {
	g.mu.RLock()
	conn := g.conn
	g.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	// Wrap command in message envelope
	message := map[string]interface{}{
		"type": "command",
		"data": cmd,
	}

	return conn.WriteJSON(message)
}

// OnMessage registers a handler for incoming server messages
// Returns an unsubscribe function
func (g *GameWebSocket) OnMessage(handler func(*ServerMessage)) func() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.handlers = append(g.handlers, handler)

	// Return unsubscribe function
	return func() {
		g.mu.Lock()
		defer g.mu.Unlock()

		for i, h := range g.handlers {
			// Compare function pointers
			if &h == &handler {
				g.handlers = append(g.handlers[:i], g.handlers[i+1:]...)
				break
			}
		}
	}
}

// readMessages continuously reads messages from the WebSocket
func (g *GameWebSocket) readMessages() {
	for {
		select {
		case <-g.done:
			return
		default:
			g.mu.RLock()
			conn := g.conn
			g.mu.RUnlock()

			if conn == nil {
				return
			}

			var msg ServerMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				// Connection closed or error
				return
			}

			// Notify all handlers
			g.mu.RLock()
			handlers := make([]func(*ServerMessage), len(g.handlers))
			copy(handlers, g.handlers)
			g.mu.RUnlock()

			for _, handler := range handlers {
				go handler(&msg)
			}
		}
	}
}
