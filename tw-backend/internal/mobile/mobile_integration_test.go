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

	"tw-backend/cmd/game-server/api"
	gameWS "tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/character"
	"tw-backend/internal/game/entry"
	"tw-backend/internal/game/processor"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/game/services/look"
	"tw-backend/internal/mobile"
	"tw-backend/internal/player"
	"tw-backend/internal/repository"
	"tw-backend/internal/testutil"
	"tw-backend/internal/world/interview"
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
	dbURL := "postgres://admin:test_password_123456@localhost:5432/mud_core?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	s.Require().NoError(err)
	s.pool = pool

	// Create repositories
	authRepo := auth.NewPostgresRepository(s.db)
	interviewRepo := interview.NewRepository(s.pool)
	worldRepo := repository.NewPostgresWorldRepository(s.pool)

	// Create services
	authConfig := &auth.Config{
		SecretKey:       []byte("test-secret-key-mobile-integration"),
		TokenExpiration: 24 * time.Hour,
	}
	s.authService = auth.NewService(authConfig, authRepo)
	_ = interview.NewServiceWithRepository(nil, interviewRepo, worldRepo) // Not used directly in tests
	entrySvc := entry.NewService(interviewRepo)
	// Services
	entitySvc := entity.NewService()
	lookService := look.NewLookService(worldRepo, nil, entitySvc, interviewRepo, authRepo, nil, nil)
	spatialSvc := player.NewSpatialService(authRepo, worldRepo, nil)
	interviewService := interview.NewServiceWithRepository(nil, interviewRepo, worldRepo)
	gameProcessor := processor.NewGameProcessor(authRepo, worldRepo, lookService, entitySvc, interviewService, spatialSvc, nil, nil, nil, nil)
	hub := gameWS.NewHub(gameProcessor)
	go hub.Run(context.Background())

	// Create handlers
	authHandler := api.NewAuthHandler(s.authService, nil, nil)
	sessionHandler := api.NewSessionHandler(authRepo, lookService)
	entryHandler := api.NewEntryHandler(entrySvc)
	worldHandler := api.NewWorldHandler(worldRepo)
	// Create CreationService
	creationService := character.NewCreationService(authRepo)

	// Create handler
	wsHandler := gameWS.NewHandler(hub, creationService, authRepo, lookService)

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

	user, err := s.client.Register(ctx, email, "TestUser", password)
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
	s.Contains(joinResp.Message, "Welcome")

	s.T().Log("✓ Complete onboarding flow successful")
}

// TestMultipleCharactersInDifferentWorlds tests creating multiple characters
func (s *MobileSDKIntegrationSuite) TestMultipleCharactersInDifferentWorlds() {
	ctx := context.Background()

	// Register and login
	email := testutil.GenerateTestEmail()
	_, err := s.client.Register(ctx, email, "MultiCharUser", "Password123")
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
	s.client.Register(ctx, email, "TestUser", "Password123")
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

// TestReturningPlayerGreeting tests the welcome message for returning players
func (s *MobileSDKIntegrationSuite) TestReturningPlayerGreeting() {
	ctx := context.Background()

	// 1. Register and Login
	email := testutil.GenerateTestEmail()
	_, err := s.client.Register(ctx, email, "ReturningPlayer", "Password123")
	s.Require().NoError(err)

	_, err = s.client.Login(ctx, email, "Password123")
	s.Require().NoError(err)

	// 2. Create World and Character
	worldID := testutil.CreateTestWorld(s.T(), s.db)
	charReq := &mobile.CreateCharacterRequest{
		WorldID: worldID.String(),
		Name:    "Returnee",
		Species: "Human",
	}
	char, err := s.client.CreateCharacter(ctx, charReq)
	s.Require().NoError(err)

	// 3. Join Game (First time -> New Player Message)
	joinResp, err := s.client.JoinGame(ctx, char.CharacterID)
	s.Require().NoError(err)
	s.Contains(joinResp.Message, "Welcome to Thousand Worlds", "Should receive new player welcome message")

	// 4. Join Game AGAIN (Returning -> Welcome Back Message)
	// We simulate a re-join. For the purpose of this test, calling JoinGame again is sufficient
	// as the first JoinGame should have set LastWorldID.
	joinResp2, err := s.client.JoinGame(ctx, char.CharacterID)
	s.Require().NoError(err)
	s.Contains(joinResp2.Message, "Welcome back", "Should receive returning player welcome message")
}

// TestTokenPersistenceAcrossRequests tests that the token persists
func (s *MobileSDKIntegrationSuite) TestTokenPersistenceAcrossRequests() {
	ctx := context.Background()

	email := testutil.GenerateTestEmail()
	s.client.Register(ctx, email, "TestUser", "Password123")

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
	s.client.Register(ctx, email, "TestUser", "Password123")

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
	s.client.Register(ctx, email, "TestUser", "Password123")
	s.client.Login(ctx, email, "Password123")

	// Create a world
	worldID := testutil.CreateTestWorld(s.T(), s.db)

	// Make concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			// Create a new client for this goroutine to avoid race conditions on token
			client := mobile.NewClient(s.server.URL)

			// Create unique user for this request
			email := testutil.GenerateTestEmail()
			password := "Password123"
			_, err := client.Register(ctx, email, testutil.GenerateTestName("User"), password)
			s.NoError(err)

			_, err = client.Login(ctx, email, password)
			s.NoError(err)

			// Get user info
			_, err = client.GetMe(ctx)
			s.NoError(err)

			// List worlds
			_, err = client.ListWorlds(ctx)
			s.NoError(err)

			// Create character
			req := &mobile.CreateCharacterRequest{
				WorldID: worldID.String(),
				Name:    testutil.GenerateTestName("Concurrent"),
				Species: "Human",
			}
			_, err = client.CreateCharacter(ctx, req)
			s.NoError(err)

			done <- true
		}(i)
	}

	// Wait for all to complete
	timeout := time.After(10 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// OK
		case <-timeout:
			s.Fail("Timeout waiting for concurrent requests")
		}
	}

	// Verify we can login as one of the users and see their character
	// Note: We can't easily check total characters count via client as it only returns user's characters
}

// Run the test suite
func TestMobileSDKIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mobile SDK integration tests in short mode")
	}

	suite.Run(t, new(MobileSDKIntegrationSuite))
}
