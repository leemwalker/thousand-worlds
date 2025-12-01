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

// TestClient_ListWorlds_Success tests successful world listing
func TestClient_ListWorlds_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/game/worlds", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{
			map[string]interface{}{
				"id":         "world-1",
				"name":       "Fantasy Realm",
				"shape":      "sphere",
				"created_at": "2024-01-01T00:00:00Z",
			},
			map[string]interface{}{
				"id":         "world-2",
				"name":       "Sci-Fi Galaxy",
				"shape":      "sphere",
				"created_at": "2024-01-02T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	worlds, err := client.ListWorlds(context.Background())

	require.NoError(t, err)
	assert.Len(t, worlds, 2)
	assert.Equal(t, "Fantasy Realm", worlds[0].Name)
	assert.Equal(t, "Sci-Fi Galaxy", worlds[1].Name)
}

// TestClient_ListWorlds_Empty tests empty world list
func TestClient_ListWorlds_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	worlds, err := client.ListWorlds(context.Background())

	require.NoError(t, err)
	assert.Empty(t, worlds)
}

// TestClient_ListWorlds_Unauthorized tests unauthorized access
func TestClient_ListWorlds_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "UNAUTHORIZED",
			"message": "Missing or invalid token",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	// No token set

	worlds, err := client.ListWorlds(context.Background())

	require.Error(t, err)
	assert.Nil(t, worlds)
}
