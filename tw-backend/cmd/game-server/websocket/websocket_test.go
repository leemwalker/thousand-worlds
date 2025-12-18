package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketHandler(t *testing.T) {
	// Setup
	upgrader := websocket.Upgrader{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Connect
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Send message
	err = ws.WriteMessage(websocket.TextMessage, []byte("hello"))
	assert.NoError(t, err)

	// Receive echo
	_, p, err := ws.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(p))
}
