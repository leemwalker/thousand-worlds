package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"mud-platform-backend/internal/auth"
)

// MockLobbyService (we can't easily mock the struct methods unless we interface it,
// but Handler uses *lobby.Service directly. For now we'll skip deep integration tests
// of ServeHTTP that require full service mocking and focus on what we can test)

// Since Handler.ServeHTTP does a lot of work including DB calls and WebSocket upgrades,
// it's hard to unit test without extensive mocking.
// However, we can test the error conditions that happen before upgrade.

func TestNewHandler(t *testing.T) {
	hub := &Hub{}
	handler := NewHandler(hub, nil, nil, nil)
	assert.Equal(t, hub, handler.Hub)
}

func TestHandler_ServeHTTP_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	// No userID in context
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_ServeHTTP_InvalidUserID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/ws", nil)
	// Add invalid userID to context
	ctx := context.WithValue(req.Context(), "userID", "invalid-uuid")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// To test success paths, we'd need to mock AuthRepo, LobbyService, and DescriptionGenerator.
// AuthRepo is an interface, so that's easy.
// LobbyService and DescriptionGenerator are structs, which is harder.
// We might need to refactor Handler to use interfaces for these services to fully test it.
// For now, let's test what we can.

func TestHandler_ServeHTTP_FetchUserError(t *testing.T) {
	authRepo := auth.NewMockRepository()
	handler := NewHandler(nil, nil, authRepo, nil)

	userID := uuid.New()
	charID := uuid.New()

	req := httptest.NewRequest("GET", "/ws?character_id="+charID.String(), nil)
	ctx := context.WithValue(req.Context(), "userID", userID.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Don't create character in repo, so GetCharacter returns error

	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
