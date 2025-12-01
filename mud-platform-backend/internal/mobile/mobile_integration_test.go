package mobile_test

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"mud-platform-backend/cmd/game-server/api"
	gameWS "mud-platform-backend/cmd/game-server/websocket"
	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/game/entry"
	"mud-platform-backend/internal/game/processor"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/mobile"
	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/testutil"
	"mud-platform-backend/internal/world/interview"
)

// MobileSDKIntegrationSuite tests the mobile SDK end-to-end
type MobileSDKIntegrationSuite struct {
	suite.Suite
	db          *sql.DB
	pool        *pgxpool.Pool
	server      *httptest.Server
	client      *mobile.Client
	wsUpgrader  websocket.Upgrader
	authService *auth.Service
}

// SetupSuite runs once before all tests
func (s *MobileSDKIntegrationSuite) SetupSuite() {
	// Setup test database
	s.db = testutil.SetupTestDB(s.T())
	testutil.RunMigrations(s.T(), s.db)

	// Create pgxpool
	dbURL := "postgres://admin:password123@localhost:5432/mud_core?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	s.Require().NoError(err)
	s.pool = pool

	// Create repositories
	authRepo := auth.NewPostgresRepository(s.db)
	interviewRepo := interview.NewRepository(s.db)
	worldRepo := repository.NewPostgresWorldRepository(s.pool)

	// Create services
	authConfig := &auth.Config{
		SecretKey:       []byte("test-secret-key-mobile-integration"),
		TokenExpiration: 24 * time.Hour,
	}
	s.authService = auth.NewService(authConfig, authRepo)
	_ = interview.NewServiceWithRepository(nil, interviewRepo) // Not used directly in tests
	entrySvc := entry.NewService(interviewRepo)
	lobbySvc := lobby.NewService(authRepo)

	// Create game processor and WebSocket hub
	gameProcessor := processor.NewGameProcessor()
	hub := gameWS.NewHub(gameProcessor)
	go hub.Run(context.Background())

	// Create handlers
	authHandler := api.NewAuthHandler(s.authService, nil, nil)
	sessionHandler := api.NewSessionHandler(authRepo)
	entryHandler := api.NewEntryHandler(entrySvc)
	worldHandler := api.NewWorldHandler(worldRepo)
	wsHandler := gameWS.NewHandler(hub, lobbySvc)

	// Setup router
	r := chi.NewRouter()
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(api.AuthMiddleware(s.authService))
		r.Get("/api/auth/me", authHandler.GetMe)
		r.Post("/api/auth/logout", authHandler.Logout)
		r.Post("/api/game/characters", sessionHandler.CreateCharacter)
		r.Get("/api/game/characters", sessionHandler.GetCharacters)
		r.Post("/api/game/join", sessionHandler.JoinGame)
		r.Get("/api/game/entry-options", entryHandler.GetEntryOptions)
		r.Get("/api/game/worlds", worldHandler.ListWorlds)
		r.Get("/api/game/ws", wsHandler.ServeHTTP)
	})

	// Create test server
	s.server = httptest.NewServer(r)
	s.client = mobile.NewClient(s.server.URL)
}

// TearDownSuite runs once after all tests
func (s *MobileSDKIntegrationSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	if s.pool != nil {
		s.pool.Close()
	}
	if s.db != nil {
		testutil.CloseDB(s.T(), s.db)
	}
}

// SetupTest runs before each test
func (s *MobileSDKIntegrationSuite) SetupTest() {
	testutil.TruncateTables(s.T(), s.db)
	s.client.ClearToken()
}

// TestCompleteOnboardingFlow tests the complete user onboarding workflow
func (s *MobileSDKIntegrationSuite) TestCompleteOnboardingFlow() {
	ctx := context.Background()

	// Step 1: Register a new user
	email := testutil.GenerateTestEmail()
	password := "TestPassword123"

	user, err := s.client.Register(ctx, email, password)
	s.Require().NoError(err)
	s.NotEmpty(user.UserID)
	s.Equal(email, user.Email)

	// Step 2: Login
	loginResp, err := s.client.Login(ctx, email, password)
	s.Require().NoError(err)
	s.NotEmpty(loginResp.Token)
	s.Equal(user.UserID, loginResp.User.UserID)

	// Verify token is automatically set
	s.NotEmpty(s.client.GetToken())

	// Step 3: Verify authenticated access
	me, err := s.client.GetMe(ctx)
	s.Require().NoError(err)
	s.Equal(user.UserID, me.UserID)
	s.Equal(email, me.Email)

	// Step 4: List available worlds
	worlds, err := s.client.ListWorlds(ctx)
	s.Require().NoError(err)
	// May be empty initially

	// Step 5: Create a test world for character creation
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	// Step 6: List worlds again to verify
	worlds, err = s.client.ListWorlds(ctx)
	s.Require().NoError(err)
	s.NotEmpty(worlds)
	s.Equal(worldID.String(), worlds[0].ID)

	// Step 7: Create a character
	charReq := &mobile.CreateCharacterRequest{
		WorldID: worldID.String(),
		Name:    "IntegrationTestHero",
		Species: "Human",
	}
	char, err := s.client.CreateCharacter(ctx, charReq)
	s.Require().NoError(err)
	s.NotEmpty(char.CharacterID)
	s.Equal("IntegrationTestHero", char.Name)
	s.Equal(worldID.String(), char.WorldID)

	// Step 8: List characters
	chars, err := s.client.GetCharacters(ctx)
	s.Require().NoError(err)
	s.Len(chars, 1)
	s.Equal(char.CharacterID, chars[0].CharacterID)

	// Step 9: Join game with character
	joinResp, err := s.client.JoinGame(ctx, char.CharacterID)
	s.Require().NoError(err)
	s.Equal(worldID.String(), joinResp.WorldID)
	s.Contains(joinResp.Message, "joined")

	s.T().Log("✓ Complete onboarding flow successful")
}

// TestMultipleCharactersInDifferentWorlds tests creating multiple characters
func (s *MobileSDKIntegrationSuite) TestMultipleCharactersInDifferentWorlds() {
	ctx := context.Background()

	// Register and login
	email := testutil.GenerateTestEmail()
	_, err := s.client.Register(ctx, email, "Password123")
	s.Require().NoError(err)

	_, err = s.client.Login(ctx, email, "Password123")
	s.Require().NoError(err)

	// Create 3 different worlds
	world1 := testutil.CreateTestWorld(s.T(), s.db)
	world2 := testutil.CreateTestWorld(s.T(), s.db)
	world3 := testutil.CreateTestWorld(s.T(), s.db)

	// Create character in each world
	chars := make([]*mobile.Character, 0)

	for i, worldID := range []string{world1.String(), world2.String(), world3.String()} {
		req := &mobile.CreateCharacterRequest{
			WorldID: worldID,
			Name:    testutil.GenerateTestName("Hero"),
			Species: "Human",
		}
		char, err := s.client.CreateCharacter(ctx, req)
		s.Require().NoError(err)
		chars = append(chars, char)
		s.T().Logf("Created character %d: %s in world %s", i+1, char.Name, char.WorldID)
	}

	// List all characters
	allChars, err := s.client.GetCharacters(ctx)
	s.Require().NoError(err)
	s.Len(allChars, 3)
}

// TestWatcherCharacterCreation tests creating a watcher character
func (s *MobileSDKIntegrationSuite) TestWatcherCharacterCreation() {
	ctx := context.Background()

	// Register and login
	email := testutil.GenerateTestEmail()
	s.client.Register(ctx, email, "Password123")
	s.client.Login(ctx, email, "Password123")

	worldID := testutil.CreateTestWorld(s.T(), s.db)

	// Create watcher character
	req := &mobile.CreateCharacterRequest{
		WorldID: worldID.String(),
		Name:    "Observer",
		Role:    "watcher",
	}
	char, err := s.client.CreateCharacter(ctx, req)
	s.Require().NoError(err)
	s.Equal("watcher", char.Role)
	s.T().Log("✓ Watcher character created successfully")
}

// TestTokenPersistenceAcrossRequests tests that the token persists
func (s *MobileSDKIntegrationSuite) TestTokenPersistenceAcrossRequests() {
	ctx := context.Background()

	email := testutil.GenerateTestEmail()
	s.client.Register(ctx, email, "Password123")

	loginResp, err := s.client.Login(ctx, email, "Password123")
	s.Require().NoError(err)

	token := loginResp.Token

	// Make multiple authenticated requests
	for i := 0; i < 3; i++ {
		me, err := s.client.GetMe(ctx)
		s.Require().NoError(err)
		s.NotNil(me)

		// Token should remain the same
		s.Equal(token, s.client.GetToken())
	}

	// Logout should clear token
	s.client.Logout()
	s.Empty(s.client.GetToken())
}

// TestInvalidCredentialsHandling tests error handling
func (s *MobileSDKIntegrationSuite) TestInvalidCredentialsHandling() {
	ctx := context.Background()

	// Try to login with non-existent user
	_, err := s.client.Login(ctx, "nonexistent@example.com", "WrongPassword")
	s.Require().Error(err)
	s.Contains(err.Error(), "credentials")

	// Register user
	email := testutil.GenerateTestEmail()
	s.client.Register(ctx, email, "Password123")

	// Try to login with wrong password
	_, err = s.client.Login(ctx, email, "WrongPassword")
	s.Require().Error(err)

	// Try authenticated request without token
	s.client.ClearToken()
	_, err = s.client.GetMe(ctx)
	s.Require().Error(err)
}

// TestConcurrentClientRequests tests thread safety
func (s *MobileSDKIntegrationSuite) TestConcurrentClientRequests() {
	ctx := context.Background()

	// Create and login
	email := testutil.GenerateTestEmail()
	s.client.Register(ctx, email, "Password123")
	s.client.Login(ctx, email, "Password123")

	// Create a world
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	// Make concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			// Get user info
			_, err := s.client.GetMe(ctx)
			s.NoError(err)

			// List worlds
			_, err = s.client.ListWorlds(ctx)
			s.NoError(err)

			// Create character
			req := &mobile.CreateCharacterRequest{
				WorldID: worldID.String(),
				Name:    testutil.GenerateTestName("Concurrent"),
				Species: "Human",
			}
			_, err = s.client.CreateCharacter(ctx, req)
			s.NoError(err)

			done <- true
		}(i)
	}

	// Wait for all to complete
	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// OK
		case <-timeout:
			s.Fail("Timeout waiting for concurrent requests")
		}
	}

	// Verify all characters were created
	chars, err := s.client.GetCharacters(ctx)
	s.Require().NoError(err)
	s.Len(chars, 10)
}

// Run the test suite
func TestMobileSDKIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mobile SDK integration tests in short mode")
	}

	suite.Run(t, new(MobileSDKIntegrationSuite))
}
