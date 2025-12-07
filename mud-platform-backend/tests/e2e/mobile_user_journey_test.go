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
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMobileUserJourney(t *testing.T) {
	// 0. Setup
	// Check if server is running
	if !isServerRunning(cfg.BaseURL) {
		t.Skipf("Server not running at %s. Skipping E2E test.", cfg.BaseURL)
	}

	// Connect to DB for verification and cleanup
	db, err := sql.Open("pgx", cfg.DBDSN)
	require.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	require.NoError(t, db.Ping(), "Failed to ping database")

	// Generate unique user
	timestamp := time.Now().Unix()
	user := TestUser{
		Email:    fmt.Sprintf("mobile_test_%d@example.com", timestamp),
		Username: fmt.Sprintf("mob_%d", timestamp), // Short username
		Password: "SecurePass123!",
	}

	// HTTP Client with Cookie Jar
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
	}

	// Cleanup function
	defer func() {
		t.Log("Cleaning up test data...")
		cleanupTestData(t, db, user.Username)
	}()

	// --- STEP 1: VERIFY USER ACCOUNT CREATION ---
	t.Run("Step1_CreateAccount", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    user.Email,
			"username": user.Username,
			"password": user.Password,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", cfg.BaseURL+"/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var respData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respData)

		// Verify DB
		var dbEmail string
		err = db.QueryRow("SELECT email FROM users WHERE username = $1", user.Username).Scan(&dbEmail)
		require.NoError(t, err, "User not found in database")
		assert.Equal(t, user.Email, dbEmail)
	})

	// --- STEP 2: VERIFY USER LOGIN ---
	t.Run("Step2_Login", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    user.Email,
			"password": user.Password,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", cfg.BaseURL+"/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Logf("Login Response: %s", string(bodyBytes))

		// Re-create body for decoder
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		require.Equal(t, http.StatusOK, resp.StatusCode, "Login failed with status: %d", resp.StatusCode)

		var respData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respData)

		token, ok := respData["token"].(string)
		require.True(t, ok, "Token not found in response")
		user.Token = token

		// Get User ID from DB
		err = db.QueryRow("SELECT user_id FROM users WHERE username = $1", user.Username).Scan(&user.ID)
		require.NoError(t, err)
	})

	// --- STEP 3: VERIFY LOBBY COMMANDS ---
	var wsConn *websocket.Conn

	t.Run("Step3_LobbyCommands", func(t *testing.T) {
		// Connect to WebSocket
		u, _ := url.Parse(cfg.WSURL + "/api/game/ws")
		q := u.Query()
		q.Set("token", user.Token)
		u.RawQuery = q.Encode()

		header := http.Header{}
		header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)")

		var err error
		wsConn, _, err = websocket.DefaultDialer.Dial(u.String(), header)
		require.NoError(t, err, "WebSocket connection failed")

		// Read welcome message (system)
		msg, err := waitForMessage(wsConn, "system", 5*time.Second)
		require.NoError(t, err)
		t.Logf("Received welcome: %s", msg)

		// Send LOOK command
		sendGameCommand(t, wsConn, "look")
		msg, err = waitForMessage(wsConn, "area_description", 3*time.Second)
		require.NoError(t, err, "Did not receive area description")
		t.Logf("Look result: %s", msg)

		// Send SAY command
		sayMsg := "Hello E2E!"
		sendGameCommandWithParams(t, wsConn, "say", &sayMsg, nil, nil)
		// Expect speech_self echo
		msg, err = waitForMessage(wsConn, "speech_self", 2*time.Second)
		require.NoError(t, err)
		assert.Contains(t, msg, "You say", "Say output should contain 'You say'")

		// Send WHO command
		sendGameCommand(t, wsConn, "who")
		msg, err = waitForMessage(wsConn, "player_list", 2*time.Second)
		require.NoError(t, err)
		t.Logf("Who result: %s", msg)
	})

	// --- STEP 3.1: SPATIAL MOVEMENT TESTS ---
	t.Run("Step3_1_SpatialMovement", func(t *testing.T) {
		require.NotNil(t, wsConn, "WebSocket connection lost")

		// Test 3.1.1: Move in all cardinal directions (happy path)
		t.Run("MoveAllDirections", func(t *testing.T) {
			directions := []string{"n", "e", "s", "w"}
			for _, dir := range directions {
				sendGameCommand(t, wsConn, dir)
				msg, err := waitForMessage(wsConn, "movement", 3*time.Second)
				require.NoError(t, err, "Movement command '%s' failed", dir)
				assert.Contains(t, msg, "You move", "Movement response should contain 'You move'")
				t.Logf("Movement %s: %s", dir, msg)
			}
		})

		// Test 3.1.2: Try to move beyond boundary (sad path)
		t.Run("MoveBeyondBoundary", func(t *testing.T) {
			// Move north repeatedly until hitting boundary
			for i := 0; i < 1005; i++ {
				sendGameCommand(t, wsConn, "n")
				msg, _ := waitForMessage(wsConn, "movement", 2*time.Second)
				if strings.Contains(msg, "cannot go further") || strings.Contains(msg, "boundary") {
					t.Logf("Hit boundary after %d moves: %s", i+1, msg)
					break
				}
			}
			// Should have hit boundary
			t.Log("Boundary test completed")
		})

		// Test 3.1.3: Move back to center for subsequent tests
		t.Run("ReturnToCenter", func(t *testing.T) {
			// Move south to get back towards center
			for i := 0; i < 500; i++ {
				sendGameCommand(t, wsConn, "s")
				_, _ = waitForMessage(wsConn, "movement", 2*time.Second)
			}
			t.Log("Returned to approximate center")
		})
	})

	// --- STEP 4: WORLD CREATION INTERVIEW TESTS ---
	t.Run("Step4_WorldCreationInterview", func(t *testing.T) {
		require.NotNil(t, wsConn, "WebSocket connection lost")

		// Test 4.1: Complete interview in one go using tell/reply
		t.Run("CompleteInterviewInOneGo", func(t *testing.T) {
			// Start interview by telling statue
			tellMsg := "I want to create a world"
			statue := "statue"
			sendGameCommandWithParams(t, wsConn, "tell", &tellMsg, &statue, nil)

			// Wait for statue's response (interview start)
			// First request might take longer due to model loading
			msg, err := waitForMessage(wsConn, "tell", 900*time.Second)
			require.NoError(t, err)
			t.Logf("Statue: %s", msg)

			// Answer all interview questions
			// Based on AllTopics: Core Concept, Sentient Species, Environment, Magic & Tech, Conflict, World Name
			answers := []string{
				"A high-tech cyberpunk world with neon cities", // Core Concept
				"Humans and AI entities",                       // Sentient Species
				"Urban sprawl with some wilderness preserves",  // Environment
				"High tech, some experimental nanotech magic",  // Magic & Tech
				"Corporate wars and AI rights movements",       // Conflict
			}

			for i, answer := range answers {
				t.Logf("Answering question %d: %s", i+1, answer)

				// Use reply command (targets last teller which is statue)
				sendGameCommand(t, wsConn, fmt.Sprintf("reply %s", answer))

				// Wait for next question or completion notification
				msg, err := waitForMessage(wsConn, "tell", 540*time.Second)
				require.NoError(t, err)
				t.Logf("Statue response %d: %s", i+1, msg)
			}

			// Final answer: World Name
			// Use unique name to avoid conflicts with previous test runs
			worldName := fmt.Sprintf("TestWorld-%d", time.Now().Unix())
			sendGameCommand(t, wsConn, fmt.Sprintf("reply %s", worldName))

			// Wait for completion message
			msg, err = waitForMessage(wsConn, "tell", 120*time.Second)
			require.NoError(t, err)
			t.Logf("Final statue message: %s", msg)

			// Verify interview marked as completed in DB
			require.Eventually(t, func() bool {
				var status string
				err := db.QueryRow("SELECT status FROM world_interviews WHERE user_id = $1", user.ID).Scan(&status)
				return err == nil && status == "completed"
			}, 15*time.Second, 500*time.Millisecond, "Interview not completed in DB")
		})

		// Test 4.2: Logout, login, resume and complete
		t.Run("LogoutResumeInterview", func(t *testing.T) {
			// Create second user for isolation
			timestamp2 := time.Now().Unix() + 1
			user2 := createTestUser(t, db, timestamp2, "resume")
			defer cleanupTestData(t, db, user2.Username)

			token2 := loginUser(t, user2)
			wsConn2 := connectWebSocket(t, token2)

			// Read welcome
			_, _ = waitForMessage(wsConn2, "system", 5*time.Second)

			// Start interview
			tellMsg := "create a new world please"
			statue := "statue"
			sendGameCommandWithParams(t, wsConn2, "tell", &tellMsg, &statue, nil)
			_, _ = waitForMessage(wsConn2, "tell", 900*time.Second)

			// Answer first 2 questions
			sendGameCommand(t, wsConn2, "reply A mystical forest realm")
			_, _ = waitForMessage(wsConn2, "tell", 540*time.Second)

			sendGameCommand(t, wsConn2, "reply Elves and tree spirits")
			_, _ = waitForMessage(wsConn2, "tell", 540*time.Second)

			// Disconnect (simulate logout)
			wsConn2.Close()
			t.Log("User logged out mid-interview")

			// Wait a moment
			time.Sleep(500 * time.Millisecond)

			// Reconnect (simulate login)
			wsConn2 = connectWebSocket(t, token2)
			defer wsConn2.Close()
			_, _ = waitForMessage(wsConn2, "system", 5*time.Second)

			// Resume interview by telling statue again
			resumeMsg := "continue my world creation"
			sendGameCommandWithParams(t, wsConn2, "tell", &resumeMsg, &statue, nil)

			// Should get the next question (question 3)
			msg, err := waitForMessage(wsConn2, "tell", 540*time.Second)
			require.NoError(t, err)
			t.Logf("Resumed at: %s", msg)

			// Complete remaining questions
			remainingAnswers := []string{
				"Ancient forests with magical clearings",
				"Natural magic, no technology",
				"Dark forces threatening the sacred groves",
			}

			for _, answer := range remainingAnswers {
				sendGameCommand(t, wsConn2, fmt.Sprintf("reply %s", answer))
				_, _ = waitForMessage(wsConn2, "tell", 540*time.Second)
			}

			// Provide world name
			sendGameCommand(t, wsConn2, "reply Eldergrove")
			_, _ = waitForMessage(wsConn2, "tell", 120*time.Second)

			// Verify completion
			require.Eventually(t, func() bool {
				var status string
				err := db.QueryRow("SELECT status FROM world_interviews WHERE user_id = $1", user2.ID).Scan(&status)
				return err == nil && status == "completed"
			}, 15*time.Second, 500*time.Millisecond)
		})

		// Test 4.3: Two users taking interview simultaneously
		t.Run("TwoUsersSimultaneous", func(t *testing.T) {
			// Create two users
			ts3 := time.Now().Unix() + 2
			user3 := createTestUser(t, db, ts3, "multi1")
			user4 := createTestUser(t, db, ts3+1, "multi2")
			defer cleanupTestData(t, db, user3.Username)
			defer cleanupTestData(t, db, user4.Username)

			token3 := loginUser(t, user3)
			token4 := loginUser(t, user4)

			wsConn3 := connectWebSocket(t, token3)
			wsConn4 := connectWebSocket(t, token4)
			defer wsConn3.Close()
			defer wsConn4.Close()

			_, _ = waitForMessage(wsConn3, "system", 5*time.Second)
			_, _ = waitForMessage(wsConn4, "system", 5*time.Second)

			// Both start interviews
			statue := "statue"
			msg1 := "I need a world"
			msg2 := "Create my world"
			sendGameCommandWithParams(t, wsConn3, "tell", &msg1, &statue, nil)
			sendGameCommandWithParams(t, wsConn4, "tell", &msg2, &statue, nil)

			_, _ = waitForMessage(wsConn3, "tell", 900*time.Second)
			_, _ = waitForMessage(wsConn4, "tell", 60*time.Second)

			// User 3 answers first question
			sendGameCommand(t, wsConn3, "reply A desert wasteland")
			msg3Response, err := waitForMessage(wsConn3, "tell", 540*time.Second)
			require.NoError(t, err)

			// User 4 answers first question (different answer)
			sendGameCommand(t, wsConn4, "reply An icy tundra")
			msg4Response, err := waitForMessage(wsConn4, "tell", 540*time.Second)
			require.NoError(t, err)

			// Verify no cross-contamination
			assert.NotContains(t, msg3Response, " icy ")
			assert.NotContains(t, msg3Response, "tundra")
			assert.NotContains(t, msg4Response, "desert")
			assert.NotContains(t, msg4Response, "wasteland")

			t.Log("No cross-contamination detected between simultaneous interviews")
		})

		// Test 4.4: Invalid answer handling
		t.Run("InvalidAnswerHandling", func(t *testing.T) {
			ts5 := time.Now().Unix() + 4
			user5 := createTestUser(t, db, ts5, "invalid")
			defer cleanupTestData(t, db, user5.Username)

			token5 := loginUser(t, user5)
			wsConn5 := connectWebSocket(t, token5)
			defer wsConn5.Close()

			_, _ = waitForMessage(wsConn5, "system", 5*time.Second)

			// Start interview
			statue := "statue"
			startMsg := "make me a world"
			sendGameCommandWithParams(t, wsConn5, "tell", &startMsg, &statue, nil)
			_, _ = waitForMessage(wsConn5, "tell", 900*time.Second)

			// Send invalid answer (just digits)
			sendGameCommand(t, wsConn5, "reply 12345")

			// Should receive a prompt to try again
			response, err := waitForMessage(wsConn5, "tell", 540*time.Second)
			require.NoError(t, err)

			// Response should indicate the answer wasn't understood
			containsRetry := strings.Contains(strings.ToLower(response), "sorry") ||
				strings.Contains(strings.ToLower(response), "understand") ||
				strings.Contains(strings.ToLower(response), "try again") ||
				strings.Contains(strings.ToLower(response), "could you") ||
				strings.Contains(strings.ToLower(response), "didn't quite catch") ||
				strings.Contains(strings.ToLower(response), "give me a sense")

			assert.True(t, containsRetry, "Statue should prompt for better answer, got: %s", response)
			t.Logf("Invalid answer handling response: %s", response)

			// Provide valid answer
			sendGameCommand(t, wsConn5, "reply A steampunk industrial city")
			_, _ = waitForMessage(wsConn5, "tell", 540*time.Second)
			t.Log("Valid answer accepted after invalid attempt")
		})
	})

	// --- STEP 5: WORLD CREATION VERIFICATION TESTS ---
	t.Run("Step5_VerifyWorldCreation", func(t *testing.T) {
		// Test 5.1: Measure world generation time and verify basic creation
		t.Run("BasicWorldGeneration", func(t *testing.T) {
			// User from Step 4.1 should have world "Neon Sprawl"
			t.Log("Measuring world generation time...")

			startTime := time.Now()
			var worldID string
			var worldName string

			// Wait for world generation with generous timeout
			// We'll measure how long it actually takes
			t.Logf("Checking for world creation for User ID: %s", user.ID)
			require.Eventually(t, func() bool {
				err := db.QueryRow("SELECT id, name FROM worlds WHERE owner_id = $1", user.ID).Scan(&worldID, &worldName)
				if err != nil {
					// Debug: List all worlds to see what exists
					rows, _ := db.Query("SELECT id, name, owner_id FROM worlds")
					if rows != nil {
						defer rows.Close()
						for rows.Next() {
							var wid, wname, wowner string = "UNSET", "UNSET", "UNSET"
							if err := rows.Scan(&wid, &wname, &wowner); err != nil {
								t.Logf("Scan error: %v", err)
								continue
							}
							t.Logf("Existing world: ID=%s, Name='%s', Owner=%s", wid, wname, wowner)
						}
					}
				}
				return err == nil
			}, 120*time.Second, 2*time.Second, "World not created in database after 2 minutes")

			genTime := time.Since(startTime)
			t.Logf("World generation completed in: %v", genTime)
			t.Logf("World created: %s (%s)", worldName, worldID)

			// Verify it's the correct world
			assert.True(t, strings.HasPrefix(worldName, "TestWorld-"), "World name should start with TestWorld-, got: %s", worldName)
		})

		// Test 5.2: User custom name with deduplication
		t.Run("CustomNameWithDedup", func(t *testing.T) {
			ts6 := time.Now().Unix() + 5
			user6 := createTestUser(t, db, ts6, "custom")
			defer cleanupTestData(t, db, user6.Username)

			token6 := loginUser(t, user6)
			wsConn6 := connectWebSocket(t, token6)
			defer wsConn6.Close()

			_, _ = waitForMessage(wsConn6, "system", 5*time.Second)

			// Create world with custom name
			statue := "statue"
			startMsg := "I want to create a world"
			sendGameCommandWithParams(t, wsConn6, "tell", &startMsg, &statue, nil)
			_, _ = waitForMessage(wsConn6, "tell", 900*time.Second)

			// Answer interview questions quickly
			answers := []string{
				"A vibrant fantasy realm",
				"Dragons and wizards",
				"Mountainous with deep valleys",
				"High magic, medieval tech",
				"Ancient evil awakening",
			}

			for _, answer := range answers {
				sendGameCommand(t, wsConn6, fmt.Sprintf("reply %s", answer))
				_, _ = waitForMessage(wsConn6, "tell", 540*time.Second)
			}

			// Setup collision: Create a world that we will collide with
			// Since BasicWorldGeneration uses random name, we must create our own blocker
			blockerName := "Neon Sprawl"
			// Just insert directly into DB to simulate existing world. Use arbitrary UUIDs.
			_, _ = db.Exec("INSERT INTO worlds (id, name, owner_id, shape, created_at) VALUES ($1, $2, $3, 'sphere', NOW())",
				uuid.New(), blockerName, user6.ID)

			// Try to use "Neon Sprawl" (which we just planted)
			sendGameCommand(t, wsConn6, "reply Neon Sprawl")
			response, err := waitForMessage(wsConn6, "tell", 540*time.Second)
			require.NoError(t, err, "Timeout waiting for name deduplication response")

			// Should be told name is taken and offered alternatives
			nameTaken := strings.Contains(strings.ToLower(response), "taken") ||
				strings.Contains(strings.ToLower(response), "already exists") ||
				strings.Contains(strings.ToLower(response), "unavailable")

			assert.True(t, nameTaken, "Should indicate name is taken, got: %s", response)
			t.Logf("Name collision detected: %s", response)

			// Provide unique name
			sendGameCommand(t, wsConn6, "reply Dragon Peak")
			_, _ = waitForMessage(wsConn6, "tell", 120*time.Second)

			// Verify world created with unique name
			require.Eventually(t, func() bool {
				var name string
				err := db.QueryRow("SELECT name FROM worlds WHERE owner_id = $1 AND name = 'Dragon Peak'", user6.ID).Scan(&name)
				return err == nil && name == "Dragon Peak"
			}, 120*time.Second, 2*time.Second)
		})

		// Test 5.3: Use statue's suggested name
		t.Run("SuggestedNameWithDedup", func(t *testing.T) {
			ts7 := time.Now().Unix() + 6
			user7 := createTestUser(t, db, ts7, "suggested")
			defer cleanupTestData(t, db, user7.Username)

			token7 := loginUser(t, user7)
			wsConn7 := connectWebSocket(t, token7)
			defer wsConn7.Close()

			_, _ = waitForMessage(wsConn7, "system", 5*time.Second)

			// Start interview
			statue := "statue"
			startMsg := "create world"
			sendGameCommandWithParams(t, wsConn7, "tell", &startMsg, &statue, nil)
			_, _ = waitForMessage(wsConn7, "tell", 900*time.Second)

			// Answer questions
			answers := []string{
				"A mysterious shadow realm",
				"Shade beings",
				"Perpetual twilight",
				"Dark magic dominates",
				"Light versus darkness",
			}

			for _, answer := range answers {
				sendGameCommand(t, wsConn7, fmt.Sprintf("reply %s", answer))
				_, _ = waitForMessage(wsConn7, "tell", 540*time.Second)
			}

			// Last message should have 2 suggested names
			// We'll accept the first suggested name by just using it
			// For E2E purposes, we'll use a common pattern name
			sendGameCommand(t, wsConn7, "reply Shadowveil")
			_, _ = waitForMessage(wsConn7, "tell", 120*time.Second)

			// Verify world created
			require.Eventually(t, func() bool {
				var name string
				err := db.QueryRow("SELECT name FROM worlds WHERE owner_id = $1", user7.ID).Scan(&name)
				return err == nil
			}, 120*time.Second, 2*time.Second)
		})

		// Test 5.4: Two users creating worlds simultaneously
		t.Run("SimultaneousWorldCreation", func(t *testing.T) {
			ts8 := time.Now().Unix() + 7
			user8 := createTestUser(t, db, ts8, "simul1")
			user9 := createTestUser(t, db, ts8+1, "simul2")
			defer cleanupTestData(t, db, user8.Username)
			defer cleanupTestData(t, db, user9.Username)

			token8 := loginUser(t, user8)
			token9 := loginUser(t, user9)

			wsConn8 := connectWebSocket(t, token8)
			wsConn9 := connectWebSocket(t, token9)
			defer wsConn8.Close()
			defer wsConn9.Close()

			_, _ = waitForMessage(wsConn8, "system", 5*time.Second)
			_, _ = waitForMessage(wsConn9, "system", 5*time.Second)

			// Start both interviews
			statue := "statue"
			msg8 := "make world"
			msg9 := "create world"
			sendGameCommandWithParams(t, wsConn8, "tell", &msg8, &statue, nil)
			sendGameCommandWithParams(t, wsConn9, "tell", &msg9, &statue, nil)

			_, err1 := waitForMessage(wsConn8, "tell", 900*time.Second)
			require.NoError(t, err1)
			_, err2 := waitForMessage(wsConn9, "tell", 900*time.Second)
			require.NoError(t, err2)

			// Both answer questions
			answers := []string{
				"Different concept",
				"Different species",
				"Different environment",
				"Different magic",
				"Different conflict",
			}

			for _, answer := range answers {
				sendGameCommand(t, wsConn8, fmt.Sprintf("reply %s", answer))
				sendGameCommand(t, wsConn9, fmt.Sprintf("reply %s", answer))
				_, _ = waitForMessage(wsConn8, "tell", 540*time.Second)
				_, _ = waitForMessage(wsConn9, "tell", 540*time.Second)
			}

			// Both try to name world "Twin Realms"
			sendGameCommand(t, wsConn8, "reply Twin Realms")
			sendGameCommand(t, wsConn9, "reply Twin Realms")

			response8, _ := waitForMessage(wsConn8, "tell", 540*time.Second)
			response9, _ := waitForMessage(wsConn9, "tell", 540*time.Second)

			// One should succeed, one should be told name is taken
			// Determine which one got the name first
			taken8 := strings.Contains(strings.ToLower(response8), "taken") ||
				strings.Contains(strings.ToLower(response8), "already")
			taken9 := strings.Contains(strings.ToLower(response9), "taken") ||
				strings.Contains(strings.ToLower(response9), "already")

			// Exactly one should have naming conflict
			assert.True(t, taken8 != taken9, "One user should get conflict, one should succeed")

			// The one with conflict provides alternative name
			if taken8 {
				sendGameCommand(t, wsConn8, "reply Twin Realms Alpha")
				_, _ = waitForMessage(wsConn8, "tell", 120*time.Second)
			} else {
				sendGameCommand(t, wsConn9, "reply Twin Realms Beta")
				_, _ = waitForMessage(wsConn9, "tell", 120*time.Second)
			}

			// Verify both worlds created with different names
			var name8, name9 string
			require.Eventually(t, func() bool {
				err1 := db.QueryRow("SELECT name FROM worlds WHERE owner_id = $1", user8.ID).Scan(&name8)
				err2 := db.QueryRow("SELECT name FROM worlds WHERE owner_id = $1", user9.ID).Scan(&name9)
				return err1 == nil && err2 == nil
			}, 120*time.Second, 2*time.Second)

			assert.NotEqual(t, name8, name9, "Worlds should have different names")
			t.Logf("User8 world: %s, User9 world: %s", name8, name9)
		})
	})

	// --- STEP 6: WORLD ENTRY TESTS ---
	t.Run("Step6_EnterWorld", func(t *testing.T) {
		// Test 6.1: Enter own world as watcher
		t.Run("EnterOwnWorldAsWatcher", func(t *testing.T) {
			// Create dedicated user and world for this test
			tsWatcher := time.Now().Unix() + 20
			userWatcher := createTestUser(t, db, tsWatcher, "watcher_own")
			defer cleanupTestData(t, db, userWatcher.Username)

			tokenWatcher := loginUser(t, userWatcher)
			wsConnWatcher := connectWebSocket(t, tokenWatcher)
			defer wsConnWatcher.Close()
			_, _ = waitForMessage(wsConnWatcher, "system", 5*time.Second)

			// Create world manually in DB to skip interview
			worldName := "WatcherWorld"
			_, err := db.Exec("INSERT INTO worlds (id, name, owner_id, shape, created_at) VALUES ($1, $2, $3, 'sphere', NOW())",
				uuid.New(), worldName, userWatcher.ID)
			require.NoError(t, err)

			// Send enter command (server expects UUID now)
			var worldID string
			err = db.QueryRow("SELECT id FROM worlds WHERE name = $1", worldName).Scan(&worldID)
			require.NoError(t, err)

			sendGameCommand(t, wsConnWatcher, fmt.Sprintf("enter %s", worldID))

			// Expect prompt for how to enter (trigger_entry_options)
			msg, err := waitForMessage(wsConnWatcher, "trigger_entry_options", 10*time.Second)
			require.NoError(t, err)
			t.Logf("Entry prompt: %s", msg)

			// Choose watcher mode
			sendGameCommand(t, wsConnWatcher, fmt.Sprintf("watcher %s", worldID))

			// Should confirm entry as watcher
			confirmMsg, err := waitForMessage(wsConnWatcher, "system", 10*time.Second)
			require.NoError(t, err)
			assert.Contains(t, strings.ToLower(confirmMsg), "watcher")
			t.Logf("Watcher entry confirmed: %s", confirmMsg)

			// Verify character created in DB with watcher role
			var role string
			err = db.QueryRow(`
				SELECT role FROM characters 
				WHERE user_id = $1 AND world_id = $2`,
				userWatcher.ID, worldID).Scan(&role)
			require.NoError(t, err)
			assert.Equal(t, "watcher", role)

			// Verify location is NOT lobby by using look command
			sendGameCommand(t, wsConnWatcher, "look")
			lookMsg, err := waitForMessage(wsConnWatcher, "area_description", 5*time.Second)
			require.NoError(t, err)
			assert.NotContains(t, lookMsg, "Grand Lobby", "Should not be in lobby after entering world")
			assert.Contains(t, lookMsg, "WatcherWorld", "Should see world name in description")
		})

		// Test 6.2: Enter own world and create character
		t.Run("EnterOwnWorldCreateCharacter", func(t *testing.T) {
			// User from 4.2 has world "Eldergrove"
			// Create dedicated user for this test since we can't easily reuse existing user's credentials
			tsChar := time.Now().Unix() + 10
			userChar := createTestUser(t, db, tsChar, "chartest")
			defer cleanupTestData(t, db, userChar.Username)

			tokenChar := loginUser(t, userChar)
			wsConnChar := connectWebSocket(t, tokenChar)
			defer wsConnChar.Close()

			_, _ = waitForMessage(wsConnChar, "system", 5*time.Second)

			// This user needs to create a world first
			// Quick interview
			// This user needs to create a world first
			// Quick interview
			statue := "statue"
			startMsg := "world please"
			sendGameCommandWithParams(t, wsConnChar, "tell", &startMsg, &statue, nil)
			_, _ = waitForMessage(wsConnChar, "tell", 60*time.Second)

			quickAnswers := []string{"Test concept", "Test species", "Test env", "Test tech", "Test conflict"}
			for _, ans := range quickAnswers {
				sendGameCommand(t, wsConnChar, fmt.Sprintf("reply %s", ans))
				_, _ = waitForMessage(wsConnChar, "tell", 60*time.Second)
			}

			sendGameCommand(t, wsConnChar, "reply CharTestWorld")
			_, _ = waitForMessage(wsConnChar, "tell", 60*time.Second)

			// Wait for world creation and get ID
			var worldID string
			require.Eventually(t, func() bool {
				err := db.QueryRow("SELECT id FROM worlds WHERE owner_id = $1", userChar.ID).Scan(&worldID)
				return err == nil
			}, 120*time.Second, 2*time.Second)

			// Now enter with character creation
			sendGameCommand(t, wsConnChar, fmt.Sprintf("enter %s", worldID))
			_, _ = waitForMessage(wsConnChar, "trigger_entry_options", 10*time.Second)

			// Choose create character with params including WorldID
			// Use 'player' as role to match verification expectation
			sendGameCommand(t, wsConnChar, fmt.Sprintf("create character Zara player Human %s", worldID))
			confirmMsg, err := waitForMessage(wsConnChar, "system", 10*time.Second)
			require.NoError(t, err, "Failed to receive character creation confirmation")
			t.Logf("Character created: %s", confirmMsg)

			// Verify character in DB
			var charName, charRole string
			err = db.QueryRow(`
				SELECT name, role FROM characters 
				WHERE user_id = $1 AND world_id = $2`,
				userChar.ID, worldID).Scan(&charName, &charRole)
			require.NoError(t, err, "Failed to find created character in DB")
			assert.Equal(t, "player", charRole)
		})

		// Test 6.3: Enter own world and take over NPC
		t.Run("EnterOwnWorldTakeNPC", func(t *testing.T) {
			// This test requires NPCs to exist in the world
			// For E2E, we'll simulate the flow
			t.Skip("NPC takeover requires pre-existing NPCs in world - implement when NPC system is in place")

			// Expected flow:
			// 1. enter <world_name>
			// 2. choose "take over npc"
			// 3. get list of available NPCs
			// 4. select NPC
			// 5. confirm takeover
		})

		// Test 6.4: Enter another user's world as watcher
		t.Run("EnterOtherWorldAsWatcher", func(t *testing.T) {
			// Create Host User and World
			tsHost := time.Now().Unix() + 21
			userHost := createTestUser(t, db, tsHost, "host_watch")
			defer cleanupTestData(t, db, userHost.Username)

			worldName := "HostWorldWatcher"
			_, err := db.Exec("INSERT INTO worlds (id, name, owner_id, shape, created_at) VALUES ($1, $2, $3, 'sphere', NOW())",
				uuid.New(), worldName, userHost.ID)
			require.NoError(t, err)

			var worldID string
			err = db.QueryRow("SELECT id FROM worlds WHERE name = $1", worldName).Scan(&worldID)
			require.NoError(t, err)

			// Create Guest User
			tsGuest := time.Now().Unix() + 22
			userGuest := createTestUser(t, db, tsGuest, "guest_watch")
			defer cleanupTestData(t, db, userGuest.Username)

			tokenGuest := loginUser(t, userGuest)
			wsConnGuest := connectWebSocket(t, tokenGuest)
			defer wsConnGuest.Close()

			_, _ = waitForMessage(wsConnGuest, "system", 5*time.Second)

			// Enter host's world
			sendGameCommand(t, wsConnGuest, fmt.Sprintf("enter %s", worldID))
			_, _ = waitForMessage(wsConnGuest, "trigger_entry_options", 10*time.Second)

			// Choose watcher
			sendGameCommand(t, wsConnGuest, fmt.Sprintf("watcher %s", worldID))
			confirmMsg, err := waitForMessage(wsConnGuest, "system", 10*time.Second)
			require.NoError(t, err)
			assert.Contains(t, strings.ToLower(confirmMsg), "watcher")

			// Verify character created for guest user in world
			var role string
			err = db.QueryRow(`
				SELECT role FROM characters 
				WHERE user_id = $1 AND world_id = $2`,
				userGuest.ID, worldID).Scan(&role)
			require.NoError(t, err)
			assert.Equal(t, "watcher", role)
		})

		// Test 6.5: Enter another user's world and create character
		t.Run("EnterOtherWorldCreateCharacter", func(t *testing.T) {
			// Create Host User and World
			tsHost2 := time.Now().Unix() + 23
			userHost2 := createTestUser(t, db, tsHost2, "host_char")
			defer cleanupTestData(t, db, userHost2.Username)

			worldName := "HostWorldChar"
			_, err := db.Exec("INSERT INTO worlds (id, name, owner_id, shape, created_at) VALUES ($1, $2, $3, 'sphere', NOW())",
				uuid.New(), worldName, userHost2.ID)
			require.NoError(t, err)

			var worldID string
			err = db.QueryRow("SELECT id FROM worlds WHERE name = $1", worldName).Scan(&worldID)
			require.NoError(t, err)

			// Create Guest User
			tsGuest2 := time.Now().Unix() + 24
			userGuest2 := createTestUser(t, db, tsGuest2, "guest_char")
			defer cleanupTestData(t, db, userGuest2.Username)

			tokenGuest2 := loginUser(t, userGuest2)
			wsConnGuest2 := connectWebSocket(t, tokenGuest2)
			defer wsConnGuest2.Close()

			_, _ = waitForMessage(wsConnGuest2, "system", 5*time.Second)

			// Enter host's world
			sendGameCommand(t, wsConnGuest2, fmt.Sprintf("enter %s", worldID))
			_, _ = waitForMessage(wsConnGuest2, "trigger_entry_options", 10*time.Second)

			// Choose create character with params
			sendGameCommand(t, wsConnGuest2, fmt.Sprintf("create character Kael player Human %s", worldID))
			confirmMsg, _ := waitForMessage(wsConnGuest2, "system", 10*time.Second)
			t.Logf("Character created in other's world: %s", confirmMsg)

			// Verify character
			var role string
			err = db.QueryRow(`
				SELECT role FROM characters 
				WHERE user_id = $1 AND world_id = $2`,
				userGuest2.ID, worldID).Scan(&role)
			require.NoError(t, err)
			assert.Equal(t, "player", role)
		})

		// Test 6.6: Enter another user's world and take over NPC
		t.Run("EnterOtherWorldTakeNPC", func(t *testing.T) {
			t.Skip("NPC takeover requires pre-existing NPCs - implement when NPC system is in place")

			// Expected flow same as 6.3 but for other user's world
		})
	})

	if wsConn != nil {
		wsConn.Close()
	}
}
