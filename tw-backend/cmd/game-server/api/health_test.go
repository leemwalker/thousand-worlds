package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler()
	assert.NotNil(t, handler)
}

func TestHealthHandler_SetReady(t *testing.T) {
	handler := NewHealthHandler()

	handler.SetReady(true)
	assert.True(t, handler.isReady.Load())

	handler.SetReady(false)
	assert.False(t, handler.isReady.Load())
}

func TestHealthHandler_SetConnectedUsers(t *testing.T) {
	handler := NewHealthHandler()

	handler.SetConnectedUsers(42)
	assert.Equal(t, int64(42), handler.connectedUsers.Load())
}

func TestHealthHandler_LivenessProbe(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()

	handler.LivenessProbe(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "alive")
}

func TestHealthHandler_ReadinessProbe(t *testing.T) {
	handler := NewHealthHandler()

	t.Run("not ready", func(t *testing.T) {
		handler.SetReady(false)

		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		handler.ReadinessProbe(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		assert.Contains(t, w.Body.String(), "not_ready")
	})

	t.Run("ready", func(t *testing.T) {
		handler.SetReady(true)

		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		handler.ReadinessProbe(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, w.Body.String(), "ready")
	})
}

func TestHealthHandler_HealthCheck(t *testing.T) {
	handler := NewHealthHandler()
	handler.SetReady(true)
	handler.SetConnectedUsers(10)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := w.Body.String()
	assert.Contains(t, body, "healthy")
	assert.Contains(t, body, "connected_users")
	assert.Contains(t, body, "10")
}

func TestHealthHandler_RegisterRoutes(t *testing.T) {
	handler := NewHealthHandler()
	mux := http.NewServeMux()

	handler.RegisterRoutes(mux)

	// Verify routes were registered by making requests
	tests := []struct {
		path string
		code int
	}{
		{"/health", http.StatusOK},
		{"/health/live", http.StatusOK},
		{"/health/ready", http.StatusOK}, // Ready by default
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			assert.Equal(t, tt.code, w.Result().StatusCode)
		})
	}
}
