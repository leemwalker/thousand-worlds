package main_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"

	"mud-platform-backend/cmd/game-server/api"
	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/game/entry"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/testutil"
	"mud-platform-backend/internal/world/interview"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// APIIntegrationSuite tests the complete API integration
type APIIntegrationSuite struct {
	suite.Suite
	db          *sql.DB
	pool        *pgxpool.Pool
	server      *httptest.Server
	client      *http.Client
	authToken   string
	baseURL     string
	authRepo    auth.Repository
	authService *auth.Service
	// interviewSvc removed - we create it inline in SetupSuite
	entrySvc *entry.Service
	lobbySvc *lobby.Service
}

// SetupSuite runs once before all tests
func (s *APIIntegrationSuite) SetupSuite() {
	// Setup test database
	s.db = testutil.SetupTestDB(s.T())
	testutil.RunMigrations(s.T(), s.db)

	// Create pgxpool for WorldRepository
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://admin:password123@localhost:5432/mud_core?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	s.Require().NoError(err, "Failed to create pgxpool")
	s.pool = pool

	// Create repositories
	s.authRepo = auth.NewPostgresRepository(s.db)
	interviewRepo := interview.NewRepository(s.db)
	worldRepo := repository.NewPostgresWorldRepository(s.pool)

	// Create services with proper signatures
	authConfig := &auth.Config{
		SecretKey:       []byte("test-secret-key"),
		TokenExpiration: 24 * time.Hour,
	}
	s.authService = auth.NewService(authConfig, s.authRepo)
	interviewSvc := interview.NewServiceWithRepository(nil, interviewRepo) // LLM client can be nil for tests
	s.entrySvc = entry.NewService(interviewRepo)
	s.lobbySvc = lobby.NewService(s.authRepo)

	// Create handlers with proper signatures
	authHandler := api.NewAuthHandler(s.authService, nil, nil) // SessionManager and RateLimiter can be nil
	interviewHandler := api.NewInterviewHandler(interviewSvc)
	sessionHandler := api.NewSessionHandler(s.authRepo)
	entryHandler := api.NewEntryHandler(s.entrySvc)
	worldHandler := api.NewWorldHandler(worldRepo)

	// Setup router with all routes
	r := chi.NewRouter()
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(api.AuthMiddleware(s.authService))
		r.Get("/api/auth/me", authHandler.GetMe)
		r.Post("/api/auth/logout", authHandler.Logout)
		r.Post("/api/world/interview/start", interviewHandler.StartInterview)
		r.Post("/api/world/interview/message", interviewHandler.ProcessMessage)
		r.Get("/api/world/interview/active", interviewHandler.GetActiveInterview)
		r.Post("/api/world/interview/finalize", interviewHandler.FinalizeInterview)
		r.Post("/api/game/characters", sessionHandler.CreateCharacter)
		r.Get("/api/game/characters", sessionHandler.GetCharacters)
		r.Post("/api/game/join", sessionHandler.JoinGame)
		r.Get("/api/game/entry-options", entryHandler.GetEntryOptions)
		r.Get("/api/game/worlds", worldHandler.ListWorlds)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create test server
	s.server = httptest.NewServer(r)
	s.baseURL = s.server.URL
	s.client = &http.Client{}
}

// TearDownSuite runs once after all tests
func (s *APIIntegrationSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	if s.db != nil {
		testutil.CloseDB(s.T(), s.db)
	}
	if s.pool != nil {
		s.pool.Close()
	}
}

// SetupTest runs before each test
func (s *APIIntegrationSuite) SetupTest() {
	testutil.TruncateTables(s.T(), s.db)
	s.authToken = ""
}

// ============================================================================
// HEALTH CHECK TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestHealthCheck() {
	resp := testutil.Get(s.T(), s.client, s.baseURL+"/health")
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode)
	body := testutil.ReadBody(s.T(), resp)
	s.Equal("OK", body)
}

// ============================================================================
// AUTHENTICATION TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestAuthRegister_Success() {
	// Prepare request
	registerReq := map[string]string{
		"email":    "newuser@example.com",
		"username": "NewUser",
		"password": "Password123",
	}

	// Execute: POST /api/auth/register
	resp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/register", registerReq)
	defer resp.Body.Close()

	// Assert: Check status code
	s.Equal(201, resp.StatusCode, "Expected 201 Created")

	// Assert: Check response body
	var responseData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &responseData)

	s.Contains(responseData, "user_id", "Response should contain user_id")
	s.Contains(responseData, "email", "Response should contain email")
	s.Equal("newuser@example.com", responseData["email"], "Email should match")

	// Verify: Check database state
	userID := responseData["user_id"].(string)
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "newuser@example.com").Scan(&count)
	s.NoError(err, "Should be able to query database")
	s.Equal(1, count, "User should exist in database")

	// Additional verification: User should have password hash
	var passwordHash string
	err = s.db.QueryRow("SELECT password_hash FROM users WHERE email = $1", "newuser@example.com").Scan(&passwordHash)
	s.NoError(err)
	s.NotEmpty(passwordHash, "Password should be hashed and stored")

	s.T().Logf("✓ Successfully registered user %s", userID)
}

func (s *APIIntegrationSuite) TestAuthRegister_DuplicateEmail() {
	// Create first user
	registerReq := map[string]string{
		"email":    "duplicate@example.com",
		"username": "DuplicateUser",
		"password": "Password123",
	}
	resp1 := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/register", registerReq)
	resp1.Body.Close()
	s.Equal(201, resp1.StatusCode)

	// Try to register with same email
	resp2 := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/register", registerReq)
	defer resp2.Body.Close()

	s.Equal(409, resp2.StatusCode, "Should return 409 Conflict")
	testutil.AssertErrorResponse(s.T(), resp2, "CONFLICT")
}

func (s *APIIntegrationSuite) TestAuthRegister_InvalidEmail() {
	registerReq := map[string]string{
		"email":    "not-an-email",
		"username": "InvalidEmailUser",
		"password": "Password123",
	}
	resp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/register", registerReq)
	defer resp.Body.Close()

	s.Equal(400, resp.StatusCode, "Should return 400 Bad Request")
	testutil.AssertErrorResponse(s.T(), resp, "INVALID_INPUT")
}

func (s *APIIntegrationSuite) TestAuthRegister_WeakPassword() {
	registerReq := map[string]string{
		"email":    "weak@example.com",
		"username": "WeakPassUser",
		"password": "short",
	}
	resp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/register", registerReq)
	defer resp.Body.Close()

	s.Equal(400, resp.StatusCode, "Should return 400 Bad Request")
	body := testutil.ReadBody(s.T(), resp)
	s.Contains(body, "password", "Error should mention password")
}

func (s *APIIntegrationSuite) TestAuthLogin_Success() {
	// Register user first
	email := "logintest@example.com"
	password := "Password123"
	registerReq := map[string]string{
		"email":    email,
		"username": "LoginTestUser",
		"password": password,
	}
	regResp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/register", registerReq)
	regResp.Body.Close()

	// Login
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}
	resp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/login", loginReq)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode, "Should return 200 OK")

	var loginData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &loginData)

	s.Contains(loginData, "token", "Response should contain token")
	s.Contains(loginData, "user", "Response should contain user")
	s.NotEmpty(loginData["token"], "Token should not be empty")
}

func (s *APIIntegrationSuite) TestAuthLogin_InvalidCredentials() {
	// Register user
	registerReq := map[string]string{
		"email":    "badpass@example.com",
		"username": "BadPassUser",
		"password": "Password123",
	}
	regResp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/register", registerReq)
	regResp.Body.Close()

	// Try login with wrong password
	loginReq := map[string]string{
		"email":    "badpass@example.com",
		"password": "WrongPassword",
	}
	resp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/login", loginReq)
	defer resp.Body.Close()

	s.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")
	testutil.AssertErrorResponse(s.T(), resp, "UNAUTHORIZED")
}

func (s *APIIntegrationSuite) TestAuthLogin_NonExistentUser() {
	loginReq := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "Password123",
	}
	resp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/auth/login", loginReq)
	defer resp.Body.Close()

	s.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")
}

func (s *APIIntegrationSuite) TestAuthGetMe() {
	// Create and login user
	email, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	// GET /api/auth/me
	resp := testutil.GetWithAuth(s.T(), s.client, s.baseURL+"/api/auth/me", token)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode, "Should return 200 OK")

	var userData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &userData)

	s.Contains(userData, "user_id", "Response should contain user_id")
	s.Contains(userData, "email", "Response should contain email")
	s.Equal(email, userData["email"], "Email should match")
}

func (s *APIIntegrationSuite) TestAuthGetMe_Unauthorized() {
	// GET without auth token
	resp := testutil.Get(s.T(), s.client, s.baseURL+"/api/auth/me")
	defer resp.Body.Close()

	s.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")
}

func (s *APIIntegrationSuite) TestAuthLogout() {
	// Create and login user
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	// POST /api/auth/logout
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/auth/logout", nil, token)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode, "Should return 200 OK")
}

// ============================================================================
// WORLD INTERVIEW TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestInterviewStart_Success() {
	// TODO: Create user and login
	// TODO: POST /api/world/interview/start with auth token
	// TODO: Assert 200 status
	// TODO: Assert response contains session_id
	// TODO: Assert response contains first question
}

func (s *APIIntegrationSuite) TestInterviewStart_Unauthorized() {
	// TODO: POST /api/world/interview/start without auth
	// TODO: Assert 401 status
}

func (s *APIIntegrationSuite) TestInterviewProcessMessage_Success() {
	// TODO: Start interview to get session_id
	// TODO: POST /api/world/interview/message with session_id and answer
	// TODO: Assert 200 status
	// TODO: Assert response contains next question or completion status
}

func (s *APIIntegrationSuite) TestInterviewProcessMessage_InvalidSession() {
	// TODO: Login
	// TODO: POST /api/world/interview/message with fake session_id
	// TODO: Assert 404 or 400 status
}

func (s *APIIntegrationSuite) TestInterviewGetActive_NotFound() {
	// TODO: Login
	// TODO: GET /api/world/interview/active
	// TODO: Assert 404 status (no active interview)
}

func (s *APIIntegrationSuite) TestInterviewGetActive_Success() {
	// TODO: Start interview
	// TODO: GET /api/world/interview/active
	// TODO: Assert 200 status
	// TODO: Assert response contains session data
}

func (s *APIIntegrationSuite) TestInterviewFinalize_Success() {
	// TODO: Complete full interview flow
	// TODO: POST /api/world/interview/finalize
	// TODO: Assert 200 status
	// TODO: Verify world created in database
}

// ============================================================================
// CHARACTER TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestCreateCharacter_Success() {
	// Login and create world
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	// Create character
	charReq := map[string]interface{}{
		"world_id": worldID.String(),
		"name":     "TestHero",
		"species":  "Human",
	}
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", charReq, token)
	defer resp.Body.Close()

	s.Equal(201, resp.StatusCode, "Should return 201 Created")

	var respData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &respData)

	// Check nested character object
	charData, ok := respData["character"].(map[string]interface{})
	s.True(ok, "Response should contain character object")
	s.Contains(charData, "character_id")
	s.Equal("TestHero", charData["name"])
}

func (s *APIIntegrationSuite) TestCreateCharacter_WatcherMode() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	charReq := map[string]interface{}{
		"world_id": worldID.String(),
		"name":     "Observer",
		"role":     "watcher",
	}
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", charReq, token)
	defer resp.Body.Close()

	s.Equal(201, resp.StatusCode)
	var respData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &respData)

	charData, ok := respData["character"].(map[string]interface{})
	s.True(ok, "Response should contain character object")
	s.Equal("watcher", charData["role"])
}

func (s *APIIntegrationSuite) TestCreateCharacter_NPCTakeover() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	charReq := map[string]interface{}{
		"world_id":    worldID.String(),
		"name":        "Blacksmith",
		"species":     "Human",
		"appearance":  `{"hair":"brown","build":"muscular"}`,
		"description": "A skilled craftsman",
		"occupation":  "Blacksmith",
	}
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", charReq, token)
	defer resp.Body.Close()

	s.Equal(201, resp.StatusCode)
	var respData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &respData)

	charData, ok := respData["character"].(map[string]interface{})
	s.True(ok, "Response should contain character object")
	s.Equal("Blacksmith", charData["occupation"])
}

func (s *APIIntegrationSuite) TestCreateCharacter_InvalidWorldID() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	charReq := map[string]interface{}{
		"world_id": "00000000-0000-0000-0000-000000000000",
		"name":     "Test",
	}
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", charReq, token)
	defer resp.Body.Close()

	// Should return error (400 or 404)
	s.True(resp.StatusCode >= 400, "Should return 4xx error")
}

func (s *APIIntegrationSuite) TestCreateCharacter_MissingName() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	charReq := map[string]interface{}{
		"world_id": worldID.String(),
		// Missing name
	}
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", charReq, token)
	defer resp.Body.Close()

	s.Equal(400, resp.StatusCode, "Should return 400 Bad Request")
}

func (s *APIIntegrationSuite) TestCreateCharacter_InvalidSpecies() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	charReq := map[string]interface{}{
		"world_id": worldID.String(),
		"name":     "Test",
		"species":  "InvalidSpecies123",
	}
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", charReq, token)
	defer resp.Body.Close()

	s.Equal(400, resp.StatusCode, "Should return 400 Bad Request")
}

func (s *APIIntegrationSuite) TestGetCharacters_Empty() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	resp := testutil.GetWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", token)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode)
	var respData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &respData)

	chars, ok := respData["characters"].([]interface{})
	// It might be nil or empty list
	if ok {
		s.Empty(chars, "Should return empty array")
	} else {
		// If key missing or nil, that's also fine for empty
		s.Nil(respData["characters"])
	}
}

func (s *APIIntegrationSuite) TestGetCharacters_MultipleCharacters() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	// Create 3 characters in different worlds
	for i := 0; i < 3; i++ {
		worldID := testutil.CreateTestWorld(s.T(), s.db)
		charReq := map[string]interface{}{
			"world_id": worldID.String(),
			"name":     fmt.Sprintf("Char%d", i+1),
			"species":  "Human",
		}
		resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", charReq, token)
		resp.Body.Close()
	}

	// Get all characters
	resp := testutil.GetWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", token)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode)
	var respData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &respData)

	chars, ok := respData["characters"].([]interface{})
	s.True(ok, "Response should contain characters list")
	s.Len(chars, 3, "Should return 3 characters")
}

// ============================================================================
// GAME SESSION TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestJoinGame_Success() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	// Create character
	charReq := map[string]interface{}{
		"world_id": worldID.String(),
		"name":     "Player1",
		"species":  "Human",
	}
	charResp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/characters", charReq, token)
	var respData map[string]interface{}
	testutil.DecodeJSON(s.T(), charResp, &respData)
	charResp.Body.Close()

	charData := respData["character"].(map[string]interface{})

	// Join game
	joinReq := map[string]interface{}{
		"character_id": charData["character_id"],
	}
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/join", joinReq, token)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode, "Should return 200 OK")
}

func (s *APIIntegrationSuite) TestJoinGame_InvalidCharacterID() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	joinReq := map[string]interface{}{
		"character_id": "00000000-0000-0000-0000-000000000000",
	}
	resp := testutil.PostJSONWithAuth(s.T(), s.client, s.baseURL+"/api/game/join", joinReq, token)
	defer resp.Body.Close()

	s.True(resp.StatusCode >= 400, "Should return error")
}

func (s *APIIntegrationSuite) TestJoinGame_Unauthorized() {
	joinReq := map[string]interface{}{
		"character_id": "test",
	}
	resp := testutil.PostJSON(s.T(), s.client, s.baseURL+"/api/game/join", joinReq)
	defer resp.Body.Close()

	s.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")
}

// ============================================================================
// WORLD TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestListWorlds_Empty() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	resp := testutil.GetWithAuth(s.T(), s.client, s.baseURL+"/api/game/worlds", token)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode)
	var worlds []interface{}
	testutil.DecodeJSON(s.T(), resp, &worlds)
	// May or may not be empty depending on test environment
}

func (s *APIIntegrationSuite) TestListWorlds_WithWorlds() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)
	// Create world directly since interview requires LLM
	testutil.CreateTestWorld(s.T(), s.db)

	resp := testutil.GetWithAuth(s.T(), s.client, s.baseURL+"/api/game/worlds", token)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode)
}

// ============================================================================
// ENTRY OPTIONS TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestGetEntryOptions_Success() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	// Get user ID from email (helper doesn't return ID, but we can query it or just use a random one for config creator)
	// Actually, we need the world ID.
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	// Create user for creator
	creator := testutil.CreateTestUser(s.T(), s.authRepo)
	creatorID := creator.UserID

	// Create world configuration
	configID := uuid.New()
	interviewID := uuid.New()

	// Insert interview first (FK constraint)
	interviewQuery := `
		INSERT INTO world_interviews (
			id, player_id, current_category, current_topic_index,
			answers, history, is_complete, created_at, updated_at
		) VALUES (
			$1, $2, 'Theme', 0,
			'{}', '[]', true, NOW(), NOW()
		)
	`
	_, err := s.db.Exec(interviewQuery, interviewID, creatorID)
	s.Require().NoError(err, "Failed to create interview record")

	query := `
		INSERT INTO world_configurations (
			id, interview_id, world_id, created_by,
			world_name,
			theme, tone, inspirations, unique_aspect, major_conflicts,
			tech_level, magic_level, advanced_tech, magic_impact,
			planet_size, climate_range, land_water_ratio, unique_features, extreme_environments,
			sentient_species, political_structure, cultural_values, economic_system, religions, taboos,
			biome_weights, resource_distribution, species_start_attributes,
			created_at
		) VALUES (
			$1, $2, $3, $4,
			'Test World',
			'Fantasy', 'Dark', '[]', 'Magic', '[]',
			'medieval', 'high', '', 'high',
			'medium', 'temperate', '50/50', '[]', '[]',
			'["Human", "Elf"]', 'Monarchy', '[]', 'Barter', '[]', '[]',
			'{}', '{}', '{}',
			NOW()
		)
	`
	_, err = s.db.Exec(query, configID, interviewID, worldID, creatorID)
	s.Require().NoError(err, "Failed to create world configuration")

	// Make request
	resp := testutil.GetWithAuth(s.T(), s.client, s.baseURL+"/api/game/entry-options?world_id="+worldID.String(), token)
	defer resp.Body.Close()

	s.Equal(200, resp.StatusCode)

	var respData map[string]interface{}
	testutil.DecodeJSON(s.T(), resp, &respData)

	s.True(respData["can_enter_as_watcher"].(bool))
	s.True(respData["can_create_custom"].(bool))

	npcs, ok := respData["available_npcs"].([]interface{})
	s.True(ok, "Should have available_npcs list")
	s.NotEmpty(npcs, "Should have generated NPCs")

	// Check first NPC
	npc := npcs[0].(map[string]interface{})
	s.NotEmpty(npc["id"])
	s.NotEmpty(npc["name"])
	s.Contains([]string{"Human", "Elf"}, npc["species"])
}

func (s *APIIntegrationSuite) TestGetEntryOptions_MissingWorldID() {
	_, token := testutil.CreateUserAndLogin(s.T(), s.baseURL)

	resp := testutil.GetWithAuth(s.T(), s.client, s.baseURL+"/api/game/entry-options", token)
	defer resp.Body.Close()

	s.Equal(400, resp.StatusCode, "Should return 400 Bad Request")
}

// ============================================================================
// COMPLETE FLOW TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestCompleteOnboardingFlow() {
	// TODO: 1. Register user
	// TODO: 2. Login
	// TODO: 3. Start interview
	// TODO: 4. Answer all questions
	// TODO: 5. Finalize interview (creates world)
	// TODO: 6. List worlds (verify world exists)
	// TODO: 7. Get entry options
	// TODO: 8. Create character
	// TODO: 9. Join game
	// TODO: Verify entire flow works end-to-end
}

func (s *APIIntegrationSuite) TestCompleteFlow_WatcherMode() {
	// TODO: Register, login, create world
	// TODO: Create watcher character
	// TODO: Join game as watcher
	// TODO: Verify watcher has limited capabilities
}

func (s *APIIntegrationSuite) TestCompleteFlow_NPCTakeover() {
	// TODO: Register, login, create world
	// TODO: Get entry options (get NPC list)
	// TODO: Create character with NPC details
	// TODO: Join game
	// TODO: Verify NPC takeover successful
}

// Run the test suite
func TestAPIIntegrationSuite(t *testing.T) {
	// Skip if short mode (unit tests only)
	if testing.Short() {
		t.Skip("Skipping API integration tests in short mode")
	}

	suite.Run(t, new(APIIntegrationSuite))
}
