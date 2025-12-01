package mobile

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGameWebSocket_NewGameWebSocket tests WebSocket client creation
func TestGameWebSocket_NewGameWebSocket(t *testing.T) {
	ws := NewGameWebSocket("ws://localhost:8080", "test-token")

	assert.NotNil(t, ws)
	assert.Equal(t, "ws://localhost:8080", ws.baseURL)
	assert.Equal(t, "test-token", ws.token)
	assert.False(t, ws.IsConnected())
}

// TestGameWebSocket_Connect_Success tests successful WebSocket connection
func TestGameWebSocket_Connect_Success(t *testing.T) {
	// Create WebSocket test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Upgrade to WebSocket
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade: %v", err)
			return
		}
		defer conn.Close()

		// Verify token in query params
		token := r.URL.Query().Get("token")
		assert.Equal(t, "test-token", token)

		// Keep connection alive for a bit
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws://" + strings.TrimPrefix(server.URL, "http://")

	ws := NewGameWebSocket(wsURL, "test-token")
	err := ws.Connect(context.Background())

	require.NoError(t, err)
	assert.True(t, ws.IsConnected())

	// Cleanup
	ws.Disconnect()
}

// TestGameWebSocket_SendCommand_Success tests sending a command
func TestGameWebSocket_SendCommand_Success(t *testing.T) {
	receivedCommand := make(chan *Command, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read message from client
		var msg struct {
			Type string   `json:"type"`
			Data *Command `json:"data"`
		}
		err = conn.ReadJSON(&msg)
		if err == nil {
			receivedCommand <- msg.Data
		}

		// Keep connection alive
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws://" + strings.TrimPrefix(server.URL, "http://")
	ws := NewGameWebSocket(wsURL, "test-token")

	err := ws.Connect(context.Background())
	require.NoError(t, err)

	// Send command
	cmd := &Command{
		Action:  "say",
		Message: "hello world",
	}
	err = ws.SendCommand(cmd)
	require.NoError(t, err)

	// Verify command was received
	select {
	case received := <-receivedCommand:
		assert.Equal(t, "say", received.Action)
		assert.Equal(t, "hello world", received.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for command")
	}

	ws.Disconnect()
}

// TestGameWebSocket_ReceiveMessage_Success tests receiving server messages
func TestGameWebSocket_ReceiveMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send message to client
		msg := map[string]interface{}{
			"type": "movement",
			"data": map[string]interface{}{
				"direction": "north",
				"message":   "You move north",
			},
		}
		conn.WriteJSON(msg)

		// Keep connection alive
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws://" + strings.TrimPrefix(server.URL, "http://")
	ws := NewGameWebSocket(wsURL, "test-token")

	// Set up message handler
	receivedMessage := make(chan *ServerMessage, 1)
	unsubscribe := ws.OnMessage(func(msg *ServerMessage) {
		receivedMessage <- msg
	})
	defer unsubscribe()

	err := ws.Connect(context.Background())
	require.NoError(t, err)

	// Wait for message
	select {
	case msg := <-receivedMessage:
		assert.Equal(t, "movement", msg.Type)
		assert.NotNil(t, msg.Data)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	ws.Disconnect()
}

// TestGameWebSocket_Disconnect tests disconnection
func TestGameWebSocket_Disconnect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Keep connection alive
		time.Sleep(1 * time.Second)
	}))
	defer server.Close()

	wsURL := "ws://" + strings.TrimPrefix(server.URL, "http://")
	ws := NewGameWebSocket(wsURL, "test-token")

	err := ws.Connect(context.Background())
	require.NoError(t, err)
	assert.True(t, ws.IsConnected())

	err = ws.Disconnect()
	require.NoError(t, err)
	assert.False(t, ws.IsConnected())
}

// TestGameWebSocket_SendCommand_NotConnected tests sending when not connected
func TestGameWebSocket_SendCommand_NotConnected(t *testing.T) {
	ws := NewGameWebSocket("ws://localhost:8080", "test-token")

	cmd := &Command{Action: "move"}
	err := ws.SendCommand(cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// TestGameWebSocket_MultipleHandlers tests multiple message handlers
func TestGameWebSocket_MultipleHandlers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send message
		msg := map[string]interface{}{
			"type": "test",
			"data": map[string]string{"message": "hello"},
		}
		conn.WriteJSON(msg)

		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws://" + strings.TrimPrefix(server.URL, "http://")
	ws := NewGameWebSocket(wsURL, "test-token")

	// Set up two handlers
	received1 := make(chan bool, 1)
	received2 := make(chan bool, 1)

	unsub1 := ws.OnMessage(func(msg *ServerMessage) {
		received1 <- true
	})
	defer unsub1()

	unsub2 := ws.OnMessage(func(msg *ServerMessage) {
		received2 <- true
	})
	defer unsub2()

	err := ws.Connect(context.Background())
	require.NoError(t, err)

	// Both handlers should receive the message
	timeout := time.After(1 * time.Second)

	select {
	case <-received1:
		// OK
	case <-timeout:
		t.Fatal("Handler 1 didn't receive message")
	}

	select {
	case <-received2:
		// OK
	case (<-time.After(100 * time.Millisecond)):
		t.Fatal("Handler 2 didn't receive message")
	}

	ws.Disconnect()
}
