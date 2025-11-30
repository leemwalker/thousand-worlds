package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"mud-platform-backend/internal/auth"
)

func TestSessionHandler_CreateCharacter(t *testing.T) {
	// Setup
	repo := auth.NewMockRepository()
	handler := NewSessionHandler(repo)

	// Create a user first
	userID := uuid.New()
	user := &auth.User{
		UserID:    userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}
	repo.CreateUser(context.Background(), user)

	t.Run("Create Watcher Character", func(t *testing.T) {
		worldID := uuid.New()
		payload := CreateCharacterRequest{
			WorldID:     worldID,
			Name:        "Watcher",
			Species:     "Spirit",
			Role:        "watcher",
			Description: "An invisible observer.",
			Occupation:  "Watcher",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/game/characters", bytes.NewBuffer(body))
		
		// Inject user ID into context (mocking AuthMiddleware)
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.CreateCharacter(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp CreateCharacterResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, "Watcher", resp.Character.Name)
		assert.Equal(t, "watcher", resp.Character.Role)
		assert.Equal(t, "An invisible observer.", resp.Character.Description)
		assert.Equal(t, "Watcher", resp.Character.Occupation)
	})

	t.Run("Create NPC Takeover Character", func(t *testing.T) {
		worldID := uuid.New()
		payload := CreateCharacterRequest{
			WorldID:     worldID,
			Name:        "Villager",
			Species:     "Human",
			Role:        "player",
			Description: "A humble villager.",
			Occupation:  "Farmer",
			Appearance:  `{"hair": "brown"}`,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/game/characters", bytes.NewBuffer(body))
		
		// Inject user ID into context (mocking AuthMiddleware)
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.CreateCharacter(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp CreateCharacterResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, "Villager", resp.Character.Name)
		assert.Equal(t, "player", resp.Character.Role)
		assert.Equal(t, "A humble villager.", resp.Character.Description)
		assert.Equal(t, "Farmer", resp.Character.Occupation)
		assert.Equal(t, `{"hair": "brown"}`, resp.Character.Appearance)
	})
}
