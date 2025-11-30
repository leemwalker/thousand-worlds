package main_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"

	"mud-platform-backend/cmd/game-server/api"
	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/game/entry"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/testutil"
	"mud-platform-backend/internal/world/interview"
)

// APIIntegrationSuite tests the complete API integration
type APIIntegrationSuite struct {
	suite.Suite
	db          *sql.DB
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

	// Create repositories
	s.authRepo = auth.NewPostgresRepository(s.db)
	interviewRepo := interview.NewRepository(s.db)

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
	worldHandler := api.NewWorldHandler(nil) // World repo can be nil for basic tests

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
	s.server.Close()
	testutil.CloseDB(s.T(), s.db)
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
	// TODO: Create user first
	// TODO: POST /api/auth/register with same email
	// TODO: Assert 409 status
	// TODO: Assert error code is "CONFLICT"
}

func (s *APIIntegrationSuite) TestAuthRegister_InvalidEmail() {
	// TODO: POST /api/auth/register with invalid email format
	// TODO: Assert 400 status
	// TODO: Assert error code is "INVALID_INPUT"
}

func (s *APIIntegrationSuite) TestAuthRegister_WeakPassword() {
	// TODO: POST /api/auth/register with password < 8 chars
	// TODO: Assert 400 status
	// TODO: Assert error mentions password requirements
}

func (s *APIIntegrationSuite) TestAuthLogin_Success() {
	// TODO: Create user via registration
	// TODO: POST /api/auth/login with correct credentials
	// TODO: Assert 200 status
	// TODO: Assert response contains token
	// TODO: Assert response contains user object
	// TODO: Save token for authenticated tests
}

func (s *APIIntegrationSuite) TestAuthLogin_InvalidCredentials() {
	// TODO: Create user
	// TODO: POST /api/auth/login with wrong password
	// TODO: Assert 401 status
	// TODO: Assert error code is "UNAUTHORIZED"
}

func (s *APIIntegrationSuite) TestAuthLogin_NonExistentUser() {
	// TODO: POST /api/auth/login with non-existent email
	// TODO: Assert 401 status
}

func (s *APIIntegrationSuite) TestAuthGetMe() {
	// TODO: Create user and login to get token
	// TODO: GET /api/auth/me with Authorization header
	// TODO: Assert 200 status
	// TODO: Assert response contains user details
}

func (s *APIIntegrationSuite) TestAuthGetMe_Unauthorized() {
	// TODO: GET /api/auth/me without Authorization header
	// TODO: Assert 401 status
}

func (s *APIIntegrationSuite) TestAuthLogout() {
	// TODO: Create user and login
	// TODO: POST /api/auth/logout with token
	// TODO: Assert 200 status
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
	// TODO: Login and create world
	// TODO: POST /api/game/characters with valid data
	// TODO: Assert 201 status
	// TODO: Assert response contains character with all fields
	// TODO: Verify character in database
}

func (s *APIIntegrationSuite) TestCreateCharacter_WatcherMode() {
	// TODO: Login and create world
	// TODO: POST /api/game/characters with role="watcher"
	// TODO: Assert 201 status
	// TODO: Assert no attributes returned (watchers don't have attributes)
	// TODO: Verify character.role = "watcher" in database
}

func (s *APIIntegrationSuite) TestCreateCharacter_NPCTakeover() {
	// TODO: Login and create world
	// TODO: POST /api/game/characters with appearance, description, occupation
	// TODO: Assert 201 status
	// TODO: Assert all fields saved correctly
}

func (s *APIIntegrationSuite) TestCreateCharacter_InvalidWorldID() {
	// TODO: Login
	// TODO: POST /api/game/characters with non-existent world_id
	// TODO: Assert 400 or 404 status
}

func (s *APIIntegrationSuite) TestCreateCharacter_MissingName() {
	// TODO: Login and create world
	// TODO: POST /api/game/characters without name
	// TODO: Assert 400 status
	// TODO: Assert validation error
}

func (s *APIIntegrationSuite) TestCreateCharacter_InvalidSpecies() {
	// TODO: Login and create world
	// TODO: POST /api/game/characters with invalid species
	// TODO: Assert 400 status
}

func (s *APIIntegrationSuite) TestGetCharacters_Empty() {
	// TODO: Login
	// TODO: GET /api/game/characters
	// TODO: Assert 200 status
	// TODO: Assert empty array
}

func (s *APIIntegrationSuite) TestGetCharacters_MultipleCharacters() {
	// TODO: Login and create 3 characters in different worlds
	// TODO: GET /api/game/characters
	// TODO: Assert 200 status
	// TODO: Assert array contains 3 characters
}

// ============================================================================
// GAME SESSION TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestJoinGame_Success() {
	// TODO: Create character
	// TODO: POST /api/game/join with character_id
	// TODO: Assert 200 status
	// TODO: Assert session established
}

func (s *APIIntegrationSuite) TestJoinGame_InvalidCharacterID() {
	// TODO: Login
	// TODO: POST /api/game/join with non-existent character_id
	// TODO: Assert 404 status
}

func (s *APIIntegrationSuite) TestJoinGame_Unauthorized() {
	// TODO: POST /api/game/join without auth
	// TODO: Assert 401 status
}

// ============================================================================
// WORLD TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestListWorlds_Empty() {
	// TODO: Login
	// TODO: GET /api/game/worlds
	// TODO: Assert 200 status
	// TODO: Assert empty array
}

func (s *APIIntegrationSuite) TestListWorlds_WithWorlds() {
	// TODO: Complete interview to create world
	// TODO: GET /api/game/worlds
	// TODO: Assert 200 status
	// TODO: Assert array contains created world
}

// ============================================================================
// ENTRY OPTIONS TESTS
// ============================================================================

func (s *APIIntegrationSuite) TestGetEntryOptions_Success() {
	// TODO: Login and create world
	// TODO: GET /api/game/entry-options?world_id={worldID}
	// TODO: Assert 200 status
	// TODO: Assert response contains available entry modes
	// TODO: Assert NPCs list is present
}

func (s *APIIntegrationSuite) TestGetEntryOptions_MissingWorldID() {
	// TODO: Login
	// TODO: GET /api/game/entry-options without world_id param
	// TODO: Assert 400 status
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
