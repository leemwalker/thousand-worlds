package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// NotificationTypeNewWorld is sent when a new world is created
	NotificationTypeNewWorld NotificationType = "new_world"
	// NotificationTypeUserSignIn is sent when a user signs in
	NotificationTypeUserSignIn NotificationType = "user_signin"
)

// Notification represents a push notification
type Notification struct {
	ID        string           `json:"id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Data      interface{}      `json:"data,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
	Read      bool             `json:"read"`
}

// NotificationPreferences represents user's notification preferences
type NotificationPreferences struct {
	UserID                 string `json:"user_id"`
	EnableNewWorldAlerts   bool   `json:"enable_new_world_alerts"`
	EnableUserSignInAlerts bool   `json:"enable_user_signin_alerts"`
	PushEnabled            bool   `json:"push_enabled"`
}

// RegisterForPushNotifications registers a device for push notifications
func (c *Client) RegisterForPushNotifications(ctx context.Context, deviceToken string, platform string) error {
	reqBody := map[string]string{
		"device_token": deviceToken,
		"platform":     platform, // "ios" or "android"
	}

	resp, err := c.doRequest(ctx, "POST", "/api/notifications/register", reqBody, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return handleErrorResponse(resp)
	}

	return nil
}

// GetNotifications retrieves notifications for the authenticated user
func (c *Client) GetNotifications(ctx context.Context, unreadOnly bool) ([]*Notification, error) {
	path := "/api/notifications"
	if unreadOnly {
		path += "?unread=true"
	}

	resp, err := c.doRequest(ctx, "GET", path, nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	var result struct {
		Notifications []*Notification `json:"notifications"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %w", err)
	}

	return result.Notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func (c *Client) MarkNotificationAsRead(ctx context.Context, notificationID string) error {
	path := fmt.Sprintf("/api/notifications/%s/read", notificationID)

	resp, err := c.doRequest(ctx, "POST", path, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}

	return nil
}

// GetNotificationPreferences retrieves the user's notification preferences
func (c *Client) GetNotificationPreferences(ctx context.Context) (*NotificationPreferences, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/notifications/preferences", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	var prefs NotificationPreferences
	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		return nil, fmt.Errorf("failed to decode preferences: %w", err)
	}

	return &prefs, nil
}

// UpdateNotificationPreferences updates the user's notification preferences
func (c *Client) UpdateNotificationPreferences(ctx context.Context, prefs *NotificationPreferences) error {
	resp, err := c.doRequest(ctx, "PUT", "/api/notifications/preferences", prefs, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}

	return nil
}
