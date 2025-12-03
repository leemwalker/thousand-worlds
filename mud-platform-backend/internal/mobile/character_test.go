package mobile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_CreateCharacter_Success tests successful character creation
func TestClient_CreateCharacter_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/game/characters", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "world-123", req["world_id"])
		assert.Equal(t, "TestHero", req["name"])
		assert.Equal(t, "Human", req["species"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"character": map[string]interface{}{
				"character_id": "char-123",
				"user_id":      "user-123",
				"world_id":     "world-123",
				"name":         "TestHero",
				"role":         "player",
			},
			"attributes": map[string]interface{}{
				"vitality": 10,
				"strength": 10,
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	req := &CreateCharacterRequest{
		WorldID: "world-123",
		Name:    "TestHero",
		Species: "Human",
	}

	char, err := client.CreateCharacter(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "char-123", char.CharacterID)
	assert.Equal(t, "TestHero", char.Name)
	assert.Equal(t, "player", char.Role)
}

// TestClient_CreateCharacter_WatcherMode tests creating a watcher character
func TestClient_CreateCharacter_WatcherMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "watcher", req["role"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"character": map[string]interface{}{
				"character_id": "char-watcher",
				"user_id":      "user-123",
				"world_id":     "world-123",
				"name":         "Observer",
				"role":         "watcher",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	req := &CreateCharacterRequest{
		WorldID: "world-123",
		Name:    "Observer",
		Role:    "watcher",
	}

	char, err := client.CreateCharacter(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "watcher", char.Role)
}

// TestClient_CreateCharacter_MissingName tests validation error
func TestClient_CreateCharacter_MissingName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "INVALID_INPUT",
				"message": "Name is required",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	req := &CreateCharacterRequest{
		WorldID: "world-123",
		// Missing name
	}

	char, err := client.CreateCharacter(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, char)
	assert.Contains(t, err.Error(), "required")
}

// TestClient_CreateCharacter_Unauthorized tests unauthorized request
func TestClient_CreateCharacter_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "UNAUTHORIZED",
				"message": "Missing or invalid token",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	// No token set

	req := &CreateCharacterRequest{
		WorldID: "world-123",
		Name:    "Test",
		Species: "Human",
	}

	char, err := client.CreateCharacter(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, char)
}

// TestClient_GetCharacters_Empty tests listing characters when none exist
func TestClient_GetCharacters_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/game/characters", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"characters": []interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	chars, err := client.GetCharacters(context.Background())

	require.NoError(t, err)
	assert.Empty(t, chars)
}

// TestClient_GetCharacters_Multiple tests listing multiple characters
func TestClient_GetCharacters_Multiple(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"characters": []interface{}{
				map[string]interface{}{
					"character_id": "char-1",
					"name":         "Char1",
					"world_id":     "world-1",
					"role":         "player",
				},
				map[string]interface{}{
					"character_id": "char-2",
					"name":         "Char2",
					"world_id":     "world-2",
					"role":         "player",
				},
				map[string]interface{}{
					"character_id": "char-3",
					"name":         "Char3",
					"world_id":     "world-3",
					"role":         "watcher",
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	chars, err := client.GetCharacters(context.Background())

	require.NoError(t, err)
	assert.Len(t, chars, 3)
	assert.Equal(t, "Char1", chars[0].Name)
	assert.Equal(t, "Char2", chars[1].Name)
	assert.Equal(t, "watcher", chars[2].Role)
}

// TestClient_JoinGame_Success tests successful game join
func TestClient_JoinGame_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/game/join", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "char-123", req["character_id"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"character": map[string]interface{}{
				"character_id": "char-123",
				"name":         "TestHero",
			},
			"world_id": "world-123",
			"message":  "Successfully joined world",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	joinResp, err := client.JoinGame(context.Background(), "char-123")

	require.NoError(t, err)
	assert.Equal(t, "world-123", joinResp.WorldID)
	assert.Contains(t, joinResp.Message, "joined")
}

// TestClient_JoinGame_InvalidCharacter tests joining with invalid character ID
func TestClient_JoinGame_InvalidCharacter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "NOT_FOUND",
				"message": "Character not found",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	joinResp, err := client.JoinGame(context.Background(), "invalid-char")

	require.Error(t, err)
	assert.Nil(t, joinResp)
	assert.Contains(t, err.Error(), "not found")
}
