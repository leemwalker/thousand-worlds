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
	GetWorldID() uuid.UUID
	GetUserID() uuid.UUID
	GetUsername() string
	SendGameMessage(msgType, text string, metadata map[string]interface{})
	SendStateUpdate(state *StateUpdateData)

	// Reply command support
	GetLastTellSender() string
	SetLastTellSender(username string)
	ClearLastTellSender()

	SetCharacterID(id uuid.UUID)
	SetWorldID(id uuid.UUID)
}

// Client represents a WebSocket client connection
type Client struct {
	ID          uuid.UUID
	CharacterID uuid.UUID
	UserID      uuid.UUID
	Username    string
	WorldID     uuid.UUID
	Hub         *Hub
	Conn        *websocket.Conn
	Send        chan []byte
	mu          sync.Mutex
	isClosed    bool

	// Reply command state
	LastTellSender   string       // Username of last player who sent us a tell
	LastTellSenderMu sync.RWMutex // Protects LastTellSender for thread-safe access
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID, characterID, worldID uuid.UUID, username string) *Client {
	return &Client{
		ID:          uuid.New(),
		CharacterID: characterID,
		UserID:      userID,
		Username:    username,
		WorldID:     worldID,
		Hub:         hub,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		isClosed:    false,
	}
}

// GetCharacterID returns the character ID associated with the client
func (c *Client) GetCharacterID() uuid.UUID {
	return c.CharacterID
}

// SetCharacterID sets the character ID associated with the client
func (c *Client) SetCharacterID(id uuid.UUID) {
	c.CharacterID = id
}

// GetWorldID returns the world ID associated with the client
func (c *Client) GetWorldID() uuid.UUID {
	return c.WorldID
}

// SetWorldID sets the world ID associated with the client
func (c *Client) SetWorldID(id uuid.UUID) {
	c.WorldID = id
}

// GetUserID returns the user ID associated with the client
func (c *Client) GetUserID() uuid.UUID {
	return c.UserID
}

// GetUsername returns the username associated with the client
func (c *Client) GetUsername() string {
	return c.Username
}

// SetLastTellSender updates who last sent this client a tell (thread-safe)
func (c *Client) SetLastTellSender(username string) {
	c.LastTellSenderMu.Lock()
	defer c.LastTellSenderMu.Unlock()
	c.LastTellSender = username
}

// GetLastTellSender retrieves who last sent this client a tell (thread-safe)
func (c *Client) GetLastTellSender() string {
	c.LastTellSenderMu.RLock()
	defer c.LastTellSenderMu.RUnlock()
	return c.LastTellSender
}

// ClearLastTellSender clears the last tell sender state
func (c *Client) ClearLastTellSender() {
	c.LastTellSenderMu.Lock()
	defer c.LastTellSenderMu.Unlock()
	c.LastTellSender = ""
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
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
		// DEBUG logging
		log.Printf("[WS] DEBUG: Received raw message: %s", string(message))

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
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SafeClose safely closes the send channel
func (c *Client) SafeClose() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.isClosed {
		close(c.Send)
		c.isClosed = true
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

	if c.isClosed {
		return websocket.ErrCloseSent
	}

	select {
	case c.Send <- jsonData:
		return nil
	default:
		// Channel is full, client too slow
		log.Printf("[WS] WARNING: Dropped message for client %s (character: %s) - Send channel full (256 buffer). Message type: %s", c.ID, c.CharacterID, msgType)
		return websocket.ErrCloseSent
	}

}

// SendError sends an error message to the client
func (c *Client) SendError(message string) {
	_ = c.SendMessage(MessageTypeError, ErrorData{
		Message: message,
	})
}

// SendGameMessage sends a game message to the client
func (c *Client) SendGameMessage(msgType, text string, metadata map[string]interface{}) {
	_ = c.SendMessage(MessageTypeGameMessage, GameMessageData{
		ID:        uuid.New().String(),
		Type:      msgType,
		Text:      text,
		Timestamp: time.Now(),
		Metadata:  metadata,
	})
}

// SendStateUpdate sends a state update to the client
func (c *Client) SendStateUpdate(state *StateUpdateData) {
	_ = c.SendMessage(MessageTypeStateUpdate, state)
}
