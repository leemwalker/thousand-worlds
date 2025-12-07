package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	hub := &Hub{}
	userID := uuid.New()
	charID := uuid.New()
	worldID := uuid.New()
	username := "TestUser"

	client := NewClient(hub, nil, userID, charID, worldID, username)

	assert.NotNil(t, client)
	assert.NotEqual(t, uuid.Nil, client.ID)
	assert.Equal(t, userID, client.UserID)
	assert.Equal(t, charID, client.CharacterID)
	assert.Equal(t, worldID, client.WorldID)
	assert.Equal(t, username, client.Username)
	assert.Equal(t, hub, client.Hub)
	assert.NotNil(t, client.Send)
}

func TestClient_Getters(t *testing.T) {
	userID := uuid.New()
	charID := uuid.New()
	worldID := uuid.New()
	username := "TestUser"

	client := &Client{
		UserID:      userID,
		CharacterID: charID,
		WorldID:     worldID,
		Username:    username,
	}

	assert.Equal(t, userID, client.GetUserID())
	assert.Equal(t, charID, client.GetCharacterID())
	assert.Equal(t, worldID, client.GetWorldID())
	assert.Equal(t, username, client.GetUsername())
}

func TestClient_SendMessage(t *testing.T) {
	client := &Client{
		Send: make(chan []byte, 10),
	}

	err := client.SendMessage("test_type", "test_data")
	require.NoError(t, err)

	select {
	case msg := <-client.Send:
		var serverMsg ServerMessage
		err := json.Unmarshal(msg, &serverMsg)
		require.NoError(t, err)
		assert.Equal(t, "test_type", serverMsg.Type)
		assert.Equal(t, "test_data", serverMsg.Data)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}

func TestClient_SendError(t *testing.T) {
	client := &Client{
		Send: make(chan []byte, 10),
	}

	client.SendError("something went wrong")

	select {
	case msg := <-client.Send:
		var serverMsg ServerMessage
		err := json.Unmarshal(msg, &serverMsg)
		require.NoError(t, err)
		assert.Equal(t, MessageTypeError, serverMsg.Type)

		dataMap, ok := serverMsg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "something went wrong", dataMap["message"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}

func TestClient_SendGameMessage(t *testing.T) {
	client := &Client{
		Send: make(chan []byte, 10),
	}

	meta := map[string]interface{}{"foo": "bar"}
	client.SendGameMessage("info", "hello world", meta)

	select {
	case msg := <-client.Send:
		var serverMsg ServerMessage
		err := json.Unmarshal(msg, &serverMsg)
		require.NoError(t, err)
		assert.Equal(t, MessageTypeGameMessage, serverMsg.Type)

		dataMap, ok := serverMsg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "info", dataMap["type"])
		assert.Equal(t, "hello world", dataMap["text"])
		assert.NotEmpty(t, dataMap["id"])
		assert.NotEmpty(t, dataMap["timestamp"])

		metaMap, ok := dataMap["metadata"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "bar", metaMap["foo"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}

func TestClient_SendStateUpdate(t *testing.T) {
	client := &Client{
		Send: make(chan []byte, 10),
	}

	state := &StateUpdateData{
		HP:       100,
		MaxHP:    100,
		Position: Position{X: 10, Y: 20},
	}
	client.SendStateUpdate(state)

	select {
	case msg := <-client.Send:
		var serverMsg ServerMessage
		err := json.Unmarshal(msg, &serverMsg)
		require.NoError(t, err)
		assert.Equal(t, MessageTypeStateUpdate, serverMsg.Type)

		// Note: JSON unmarshaling converts numbers to float64
		dataMap, ok := serverMsg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(100), dataMap["hp"])

		posMap, ok := dataMap["position"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(10), posMap["x"])
		assert.Equal(t, float64(20), posMap["y"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}
