package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

// Config holds test configuration
type Config struct {
	BaseURL string
	WSURL   string
	DBDSN   string
}

// TestUser represents the test user
type TestUser struct {
	Email    string
	Username string
	Password string
	ID       string
	Token    string
}

// ServerMessage represents a message from server (matches backend format)
type ServerMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// GameMessageData represents the data inside a game_message
type GameMessageData struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "system", "area_description", "tell", etc.
	Text      string                 `json:"text"`
	Timestamp string                 `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Global config
var cfg = Config{
	BaseURL: "http://localhost:8080",
	WSURL:   "ws://localhost:8080",
	DBDSN:   "postgres://admin:password123@localhost:5432/mud_core?sslmode=disable",
}

func init() {
	if url := os.Getenv("TEST_BASE_URL"); url != "" {
		cfg.BaseURL = url
	}
	if url := os.Getenv("TEST_WS_URL"); url != "" {
		cfg.WSURL = url
	}
	if dsn := os.Getenv("TEST_DB_DSN"); dsn != "" {
		cfg.DBDSN = dsn
	}
}

// Helper functions

func isServerRunning(url string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func createTestUser(t *testing.T, db *sql.DB, timestamp int64, prefix string) TestUser {
	user := TestUser{
		Email:    fmt.Sprintf("%s_%d@example.com", prefix, timestamp),
		Username: fmt.Sprintf("%s_%d", prefix, timestamp),
		Password: "SecurePass123!",
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Timeout: 10 * time.Second, Jar: jar}

	reqBody := map[string]string{
		"email":    user.Email,
		"username": user.Username,
		"password": user.Password,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", cfg.BaseURL+"/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Get user ID
	err = db.QueryRow("SELECT user_id FROM users WHERE username = $1", user.Username).Scan(&user.ID)
	require.NoError(t, err)

	return user
}

func loginUser(t *testing.T, user TestUser) string {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Timeout: 10 * time.Second, Jar: jar}

	reqBody := map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", cfg.BaseURL+"/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respData)

	token, ok := respData["token"].(string)
	require.True(t, ok)

	return token
}

func connectWebSocket(t *testing.T, token string) *websocket.Conn {
	u, _ := url.Parse(cfg.WSURL + "/api/game/ws")
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.NoError(t, err)

	return wsConn
}

func sendGameCommand(t *testing.T, ws *websocket.Conn, command string) {
	// Send raw text command (matches refactored frontend)
	msg := map[string]interface{}{
		"type": "command",
		"data": map[string]string{
			"text": command,
		},
	}
	err := ws.WriteJSON(msg)
	require.NoError(t, err)
}

func sendGameCommandWithParams(t *testing.T, ws *websocket.Conn, action string, message, recipient, target *string) {
	// Build raw text command from parameters
	var command string

	switch action {
	case "tell":
		if recipient != nil && message != nil {
			command = fmt.Sprintf("tell %s %s", *recipient, *message)
		}
	case "say":
		if message != nil {
			command = fmt.Sprintf("say %s", *message)
		}
	case "reply":
		if message != nil {
			command = fmt.Sprintf("reply %s", *message)
		}
	default:
		command = action
		if target != nil {
			command += " " + *target
		}
		if message != nil {
			command += " " + *message
		}
	}

	sendGameCommand(t, ws, command)
}

func readNextMessage(ws *websocket.Conn, timeout time.Duration) (string, string, error) {
	ws.SetReadDeadline(time.Now().Add(timeout))
	var msg ServerMessage
	err := ws.ReadJSON(&msg)
	if err != nil {
		return "", "", err
	}

	// For game_message type, parse the nested data
	if msg.Type == "game_message" {
		var gameMsg GameMessageData
		if err := json.Unmarshal(msg.Data, &gameMsg); err != nil {
			return msg.Type, string(msg.Data), nil // Fallback to raw
		}
		// Return the inner type and text
		return gameMsg.Type, gameMsg.Text, nil
	}

	// For other message types, try to extract content
	var dataStruct struct {
		Content string `json:"content"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(msg.Data, &dataStruct); err == nil {
		if dataStruct.Content != "" {
			return msg.Type, dataStruct.Content, nil
		}
		if dataStruct.Message != "" {
			return msg.Type, dataStruct.Message, nil
		}
	}

	// Fallback: try to unmarshal data as string
	var contentString string
	if err := json.Unmarshal(msg.Data, &contentString); err == nil {
		return msg.Type, contentString, nil
	}

	// Final fallback: return raw data as string
	return msg.Type, string(msg.Data), nil
}

func waitForMessage(ws *websocket.Conn, msgType string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout waiting for message type: %s", msgType)
		}

		typ, content, err := readNextMessage(ws, timeout)
		if err != nil {
			return "", err
		}

		if typ == "error" {
			return "", fmt.Errorf("server error: %s", content)
		}

		if typ == msgType {
			return content, nil
		}
	}
}

func waitForMessageContent(ws *websocket.Conn, contentSnippet string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout waiting for message content: %s", contentSnippet)
		}

		_, content, err := readNextMessage(ws, timeout)
		if err != nil {
			return "", err
		}

		if len(content) > 0 && len(contentSnippet) > 0 {
			// Simple contains check
			for i := 0; i <= len(content)-len(contentSnippet); i++ {
				if content[i:i+len(contentSnippet)] == contentSnippet {
					return content, nil
				}
			}
		}
	}
}

func cleanupTestData(t *testing.T, db *sql.DB, username string) {
	// Get User ID
	var userID string
	err := db.QueryRow("SELECT user_id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		return // User might not exist if creation failed
	}

	// Delete related data (in dependency order)
	_, _ = db.Exec("DELETE FROM interview_answers WHERE interview_id IN (SELECT id FROM world_interviews WHERE user_id = $1)", userID)
	_, _ = db.Exec("DELETE FROM world_configurations WHERE created_by = $1", userID)
	_, _ = db.Exec("DELETE FROM characters WHERE user_id = $1", userID)
	_, _ = db.Exec("DELETE FROM worlds WHERE owner_id = $1", userID)
	_, _ = db.Exec("DELETE FROM world_interviews WHERE user_id = $1", userID)
	_, _ = db.Exec("DELETE FROM users WHERE user_id = $1", userID)

	// Also cleanup any other test users (multi1_, multi2_, resume_, etc.)
	// This handles users created in sub-tests
	rows, err := db.Query("SELECT user_id FROM users WHERE username LIKE 'multi%' OR username LIKE 'resume%' OR username LIKE 'invalid%' OR username LIKE 'custom%' OR username LIKE 'suggested%' OR username LIKE 'simul%' OR username LIKE 'chartest%'")
	if err == nil {
		defer rows.Close()
		var testUserIDs []string
		for rows.Next() {
			var uid string
			if err := rows.Scan(&uid); err == nil {
				testUserIDs = append(testUserIDs, uid)
			}
		}

		// Clean up all test users
		for _, uid := range testUserIDs {
			_, _ = db.Exec("DELETE FROM interview_answers WHERE interview_id IN (SELECT id FROM world_interviews WHERE user_id = $1)", uid)
			_, _ = db.Exec("DELETE FROM world_configurations WHERE created_by = $1", uid)
			_, _ = db.Exec("DELETE FROM characters WHERE user_id = $1", uid)
			_, _ = db.Exec("DELETE FROM worlds WHERE owner_id = $1", uid)
			_, _ = db.Exec("DELETE FROM world_interviews WHERE user_id = $1", uid)
			_, _ = db.Exec("DELETE FROM users WHERE user_id = $1", uid)
		}
	}
}
