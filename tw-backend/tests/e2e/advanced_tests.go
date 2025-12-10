package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdvancedLobbyCommands tests advanced lobby commands
func TestAdvancedLobbyCommands(t *testing.T) {
	// Setup
	if !isServerRunning(cfg.BaseURL) {
		t.Skipf("Server not running at %s. Skipping E2E test.", cfg.BaseURL)
	}

	db, err := sql.Open("pgx", cfg.DBDSN)
	require.NoError(t, err)
	defer db.Close()

	// Create two test users for multi-player testing
	timestamp := time.Now().Unix()
	user1 := createTestUser(t, db, timestamp, "user1")
	user2 := createTestUser(t, db, timestamp+1, "user2")

	defer cleanupTestData(t, db, user1.Username)
	defer cleanupTestData(t, db, user2.Username)

	// Login both users
	user1Token := loginUser(t, user1)
	user2Token := loginUser(t, user2)

	// Connect user1 to WebSocket
	wsConn1 := connectWebSocket(t, user1Token)
	defer wsConn1.Close()

	// Read welcome message
	_, _ = waitForMessage(wsConn1, "system", 5*time.Second)

	// Test 1: Look at statue
	t.Run("LookStatue", func(t *testing.T) {
		target := "statue"
		sendGameCommandWithParams(t, wsConn1, "look", nil, nil, &target)

		msg, err := waitForMessage(wsConn1, "look_result", 3*time.Second)
		require.NoError(t, err)
		t.Logf("Statue description: %s", msg)
		assert.Contains(t, strings.ToLower(msg), "statue", "Should contain statue description")
	})

	// Test 2: Look at another player
	t.Run("LookPlayer", func(t *testing.T) {
		// Connect user2
		wsConn2 := connectWebSocket(t, user2Token)
		defer wsConn2.Close()

		// Read welcome message for user2
		_, _ = waitForMessage(wsConn2, "system", 5*time.Second)

		// Give time for both connections to register
		time.Sleep(100 * time.Millisecond)

		// User1 looks at user2
		target := user2.Username
		sendGameCommandWithParams(t, wsConn1, "look", nil, nil, &target)

		msg, err := waitForMessage(wsConn1, "look_result", 3*time.Second)
		require.NoError(t, err)
		t.Logf("Player description: %s", msg)
		assert.Contains(t, msg, user2.Username, "Should describe the player")
	})
}

// TestErrorConditions tests various error scenarios
func TestErrorConditions(t *testing.T) {
	if !isServerRunning(cfg.BaseURL) {
		t.Skipf("Server not running at %s. Skipping E2E test.", cfg.BaseURL)
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Timeout: 10 * time.Second, Jar: jar}

	// Test 1: Invalid login credentials
	t.Run("InvalidCredentials", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "WrongPassword123!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", cfg.BaseURL+"/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should reject invalid credentials")
	})

	// Test 2: Malformed JSON
	t.Run("MalformedJSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", cfg.BaseURL+"/api/auth/login", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should reject malformed JSON")
	})

	// Test 3: Missing required fields
	t.Run("MissingFields", func(t *testing.T) {
		reqBody := map[string]string{
			"email": "test@example.com",
			// Missing password
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", cfg.BaseURL+"/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should reject missing fields")
	})

	// Test 4: WebSocket with invalid token
	t.Run("InvalidWebSocketToken", func(t *testing.T) {
		u, _ := url.Parse(cfg.WSURL + "/api/game/ws")
		q := u.Query()
		q.Set("token", "invalid_token_12345")
		u.RawQuery = q.Encode()

		_, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			t.Error("Should fail to connect with invalid token")
		}
		if resp != nil {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should return 401 for invalid token")
		}
	})

	// Test 5: Invalid command
	t.Run("InvalidCommand", func(t *testing.T) {
		db, _ := sql.Open("pgx", cfg.DBDSN)
		defer db.Close()

		timestamp := time.Now().Unix()
		user := createTestUser(t, db, timestamp, "errtest")
		defer cleanupTestData(t, db, user.Username)

		token := loginUser(t, user)
		wsConn := connectWebSocket(t, token)
		defer wsConn.Close()

		// Read welcome
		_, _ = waitForMessage(wsConn, "system", 5*time.Second)

		// Send invalid command
		sendGameCommand(t, wsConn, "invalidcommand123")

		// Should receive error message
		msg, err := waitForMessage(wsConn, "error", 2*time.Second)
		require.NoError(t, err)
		t.Logf("Error message: %s", msg)
		assert.Contains(t, strings.ToLower(msg), "unknown", "Should indicate unknown command")
	})
}

// TestPerformanceBenchmark measures response times
func TestPerformanceBenchmark(t *testing.T) {
	if !isServerRunning(cfg.BaseURL) {
		t.Skipf("Server not running at %s. Skipping E2E test.", cfg.BaseURL)
	}

	db, _ := sql.Open("pgx", cfg.DBDSN)
	defer db.Close()

	timestamp := time.Now().Unix()
	user := createTestUser(t, db, timestamp, "perftest")
	defer cleanupTestData(t, db, user.Username)

	token := loginUser(t, user)

	t.Run("RegistrationLatency", func(t *testing.T) {
		start := time.Now()
		timestamp := time.Now().Unix()
		_ = createTestUser(t, db, timestamp, "perfuser")
		duration := time.Since(start)

		defer cleanupTestData(t, db, fmt.Sprintf("perfuser_%d", timestamp))

		t.Logf("Registration latency: %v", duration)
		assert.Less(t, duration.Milliseconds(), int64(1000), "Registration should complete in < 1s")
	})

	t.Run("LoginLatency", func(t *testing.T) {
		start := time.Now()
		_ = loginUser(t, user)
		duration := time.Since(start)

		t.Logf("Login latency: %v", duration)
		assert.Less(t, duration.Milliseconds(), int64(500), "Login should complete in < 500ms")
	})

	t.Run("WebSocketLatency", func(t *testing.T) {
		wsConn := connectWebSocket(t, token)
		defer wsConn.Close()

		// Read welcome
		_, _ = waitForMessage(wsConn, "system", 5*time.Second)

		// Measure command response time
		start := time.Now()
		sendGameCommand(t, wsConn, "who")
		_, err := waitForMessage(wsConn, "player_list", 2*time.Second)
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("Command response latency: %v", duration)
		assert.Less(t, duration.Milliseconds(), int64(200), "Command response should be < 200ms")
	})

	t.Run("ConcurrentCommands", func(t *testing.T) {
		wsConn := connectWebSocket(t, token)
		defer wsConn.Close()

		// Read welcome
		_, _ = waitForMessage(wsConn, "system", 5*time.Second)

		// Send 10 commands rapidly
		start := time.Now()
		for i := 0; i < 10; i++ {
			sendGameCommand(t, wsConn, "who")
		}

		// Wait for all 10 responses
		for i := 0; i < 10; i++ {
			_, _ = waitForMessage(wsConn, "player_list", 2*time.Second)
		}
		duration := time.Since(start)

		t.Logf("10 concurrent commands completed in: %v", duration)
		assert.Less(t, duration.Milliseconds(), int64(2000), "10 commands should complete in < 2s")
	})
}

// TestMultipleUsers tests concurrent users
func TestMultipleUsers(t *testing.T) {
	if !isServerRunning(cfg.BaseURL) {
		t.Skipf("Server not running at %s. Skipping E2E test.", cfg.BaseURL)
	}

	db, _ := sql.Open("pgx", cfg.DBDSN)
	defer db.Close()

	const numUsers = 5
	users := make([]TestUser, numUsers)
	connections := make([]*websocket.Conn, numUsers)

	// Create and connect all users
	timestamp := time.Now().Unix()
	for i := 0; i < numUsers; i++ {
		users[i] = createTestUser(t, db, timestamp+int64(i), fmt.Sprintf("multi%d", i))
		defer cleanupTestData(t, db, users[i].Username)

		token := loginUser(t, users[i])
		connections[i] = connectWebSocket(t, token)
		defer connections[i].Close()

		// Read welcome
		_, _ = waitForMessage(connections[i], "system", 5*time.Second)
	}

	t.Run("AllUsersSeenInWho", func(t *testing.T) {
		// User 0 sends 'who' command
		sendGameCommand(t, connections[0], "who")
		msg, err := waitForMessage(connections[0], "player_list", 2*time.Second)
		require.NoError(t, err)

		t.Logf("Who result: %s", msg)

		// Check that all users appear in the list
		for _, user := range users {
			assert.Contains(t, msg, user.Username, "User list should contain %s", user.Username)
		}
	})

	t.Run("BroadcastSayCommand", func(t *testing.T) {
		// User 0 says something
		sayMsg := "Hello from user 0!"
		sendGameCommandWithParams(t, connections[0], "say", &sayMsg, nil, nil)

		// User 0 should receive speech_self
		msg, err := waitForMessage(connections[0], "speech_self", 2*time.Second)
		require.NoError(t, err)
		assert.Contains(t, msg, "You say")

		// Other users should receive speech
		for i := 1; i < numUsers; i++ {
			msg, err := waitForMessage(connections[i], "speech", 2*time.Second)
			require.NoError(t, err)
			assert.Contains(t, msg, users[0].Username)
			assert.Contains(t, msg, "Hello from user 0!")
		}
	})
}
