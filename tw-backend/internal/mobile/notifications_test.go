package mobile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_RegisterForPushNotifications tests device registration
func TestClient_RegisterForPushNotifications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notifications/register", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "device-token-123", req["device_token"])
		assert.Equal(t, "ios", req["platform"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "registered",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	err := client.RegisterForPushNotifications(context.Background(), "device-token-123", "ios")
	require.NoError(t, err)
}

// TestClient_GetNotifications tests fetching notifications
func TestClient_GetNotifications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notifications", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"notifications": []interface{}{
				map[string]interface{}{
					"id":        "notif-1",
					"type":      "new_world",
					"title":     "New World Created",
					"message":   "A new world 'Fantasy Realm' has been created",
					"timestamp": time.Now().Format(time.RFC3339),
					"read":      false,
				},
				map[string]interface{}{
					"id":        "notif-2",
					"type":      "user_signin",
					"title":     "User Signed In",
					"message":   "User TestUser123 has signed in",
					"timestamp": time.Now().Format(time.RFC3339),
					"read":      true,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	notifications, err := client.GetNotifications(context.Background(), false)
	require.NoError(t, err)
	assert.Len(t, notifications, 2)
	assert.Equal(t, "notif-1", notifications[0].ID)
	assert.Equal(t, NotificationTypeNewWorld, notifications[0].Type)
	assert.False(t, notifications[0].Read)
}

// TestClient_GetNotifications_UnreadOnly tests fetching only unread notifications
func TestClient_GetNotifications_UnreadOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notifications", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("unread"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"notifications": []interface{}{
				map[string]interface{}{
					"id":        "notif-1",
					"type":      "new_world",
					"title":     "New World",
					"message":   "Check it out!",
					"timestamp": time.Now().Format(time.RFC3339),
					"read":      false,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	notifications, err := client.GetNotifications(context.Background(), true)
	require.NoError(t, err)
	assert.Len(t, notifications, 1)
	assert.False(t, notifications[0].Read)
}

// TestClient_MarkNotificationAsRead tests marking a notification as read
func TestClient_MarkNotificationAsRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notifications/notif-123/read", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	err := client.MarkNotificationAsRead(context.Background(), "notif-123")
	require.NoError(t, err)
}

// TestClient_GetNotificationPreferences tests fetching notification preferences
func TestClient_GetNotificationPreferences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notifications/preferences", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":                   "user-123",
			"enable_new_world_alerts":   true,
			"enable_user_signin_alerts": false,
			"push_enabled":              true,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	prefs, err := client.GetNotificationPreferences(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "user-123", prefs.UserID)
	assert.True(t, prefs.EnableNewWorldAlerts)
	assert.False(t, prefs.EnableUserSignInAlerts)
	assert.True(t, prefs.PushEnabled)
}

// TestClient_UpdateNotificationPreferences tests updating notification preferences
func TestClient_UpdateNotificationPreferences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notifications/preferences", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var prefs NotificationPreferences
		json.NewDecoder(r.Body).Decode(&prefs)
		assert.True(t, prefs.EnableNewWorldAlerts)
		assert.True(t, prefs.EnableUserSignInAlerts)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "updated",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	prefs := &NotificationPreferences{
		UserID:                 "user-123",
		EnableNewWorldAlerts:   true,
		EnableUserSignInAlerts: true,
		PushEnabled:            true,
	}

	err := client.UpdateNotificationPreferences(context.Background(), prefs)
	require.NoError(t, err)
}
