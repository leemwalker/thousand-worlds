package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// GameClient defines the interface for a client connection
type GameClient interface {
	GetCharacterID() uuid.UUID
	SendGameMessage(msgType, text string, metadata map[string]interface{})
	SendStateUpdate(state *StateUpdateData)
}

// Client represents a WebSocket client connection
type Client struct {
	ID          uuid.UUID
	CharacterID uuid.UUID
	UserID      uuid.UUID
	Hub         *Hub
	Conn        *websocket.Conn
	Send        chan []byte
	mu          sync.Mutex
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID, characterID uuid.UUID) *Client {
	return &Client{
		ID:          uuid.New(),
		CharacterID: characterID,
		UserID:      userID,
		Hub:         hub,
		Conn:        conn,
		Send:        make(chan []byte, 256),
	}
}

// GetCharacterID returns the character ID associated with the client
func (c *Client) GetCharacterID() uuid.UUID {
	return c.CharacterID
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// Parse client message
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			c.SendError("Invalid message format")
			continue
		}

		// Handle message
		c.Hub.HandleMessage <- &ClientMessageWrapper{
			Client:  c,
			Message: &clientMsg,
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a server message to the client
func (c *Client) SendMessage(msgType string, data interface{}) error {
	msg := ServerMessage{
		Type: msgType,
		Data: data,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case c.Send <- jsonData:
		return nil
	default:
		// Channel is full, client too slow
		return websocket.ErrCloseSent
	}
}

// SendError sends an error message to the client
func (c *Client) SendError(message string) {
	c.SendMessage(MessageTypeError, ErrorData{
		Message: message,
	})
}

// SendGameMessage sends a game message to the client
func (c *Client) SendGameMessage(msgType, text string, metadata map[string]interface{}) {
	c.SendMessage(MessageTypeGameMessage, GameMessageData{
		ID:        uuid.New().String(),
		Type:      msgType,
		Text:      text,
		Timestamp: time.Now(),
		Metadata:  metadata,
	})
}

// SendStateUpdate sends a state update to the client
func (c *Client) SendStateUpdate(state *StateUpdateData) {
	c.SendMessage(MessageTypeStateUpdate, state)
}
