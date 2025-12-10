package lobby_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/testutil"
)

// LobbyIntegrationSuite tests lobby character creation with real database
type LobbyIntegrationSuite struct {
	suite.Suite
	db      *sql.DB
	repo    auth.Repository
	service *lobby.Service
}

// SetupSuite runs once before all tests
func (s *LobbyIntegrationSuite) SetupSuite() {
	s.db = testutil.SetupTestDB(s.T())
	testutil.RunMigrations(s.T(), s.db)
	s.repo = auth.NewPostgresRepository(s.db)
	s.service = lobby.NewService(s.repo)
}

// TearDownSuite runs once after all tests
func (s *LobbyIntegrationSuite) TearDownSuite() {
	testutil.CloseDB(s.T(), s.db)
}

// SetupTest runs before each test
func (s *LobbyIntegrationSuite) SetupTest() {
	// Clean tables before each test to avoid conflicts
	testutil.TruncateTables(s.T(), s.db)
}

// TestLobbyIntegration_NewUserConnection tests creating lobby character for first-time user
func (s *LobbyIntegrationSuite) TestLobbyIntegration_NewUserConnection() {
	ctx := context.Background()

	// Create a user (simulating registration)
	user := &auth.User{
		UserID:       uuid.New(),
		Email:        testutil.GenerateTestEmail(),
		PasswordHash: "hashed_password",
	}
	err := s.repo.CreateUser(ctx, user)
	s.Require().NoError(err, "Failed to create test user")

	// Simulate WebSocket connection without character - should create lobby character
	lobbyChar, err := s.service.EnsureLobbyCharacter(ctx, user.UserID)

	// Assertions
	s.Require().NoError(err, "Should create lobby character without foreign key error")
	s.NotNil(lobbyChar)
	s.Equal(user.UserID, lobbyChar.UserID)
	s.Equal(lobby.LobbyWorldID, lobbyChar.WorldID, "Should be assigned to lobby world")
	s.Equal("Ghost", lobbyChar.Name, "Default name for new user")
	s.Equal("ghost", lobbyChar.Role, "Default role for new user")
	s.NotEmpty(lobbyChar.Appearance)

	s.T().Log("✓ Lobby character created successfully for new user without foreign key violation")
}

// TestLobbyIntegration_ForeignKeyConstraint verifies the foreign key constraint is satisfied
func (s *LobbyIntegrationSuite) TestLobbyIntegration_ForeignKeyConstraint() {
	ctx := context.Background()

	// Create a user
	user := &auth.User{
		UserID:       uuid.New(),
		Email:        testutil.GenerateTestEmail(),
		PasswordHash: "hashed_password",
	}
	err := s.repo.CreateUser(ctx, user)
	s.Require().NoError(err)

	// Verify lobby world exists in database
	var lobbyWorldExists bool
	err = s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM worlds WHERE id = $1)
	`, lobby.LobbyWorldID).Scan(&lobbyWorldExists)
	s.Require().NoError(err)
	s.True(lobbyWorldExists, "Lobby world should exist in database (created by migration 000018)")

	// Create lobby character - this should satisfy the foreign key constraint
	lobbyChar, err := s.service.EnsureLobbyCharacter(ctx, user.UserID)
	s.Require().NoError(err, "Foreign key constraint should be satisfied")

	// Verify character exists in database with correct world_id
	var worldID uuid.UUID
	err = s.db.QueryRow(`
		SELECT world_id FROM characters WHERE character_id = $1
	`, lobbyChar.CharacterID).Scan(&worldID)
	s.Require().NoError(err)
	s.Equal(lobby.LobbyWorldID, worldID, "Character should reference lobby world")

	s.T().Log("✓ Foreign key constraint satisfied - lobby character references valid world")
}

// TestLobbyIntegration_MultipleUsers tests multiple users connecting to lobby simultaneously
func (s *LobbyIntegrationSuite) TestLobbyIntegration_MultipleUsers() {
	ctx := context.Background()

	// Create 5 users with unique emails
	userIDs := make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		user := &auth.User{
			UserID:       uuid.New(),
			Email:        testutil.GenerateTestEmail(), // Each call generates a unique email
			PasswordHash: "hashed_password",
		}
		err := s.repo.CreateUser(ctx, user)
		s.Require().NoError(err, "Failed to create user %d", i+1)
		userIDs[i] = user.UserID
	}

	// Each user connects to lobby
	lobbyChars := make([]*auth.Character, 5)
	for i, userID := range userIDs {
		char, err := s.service.EnsureLobbyCharacter(ctx, userID)
		s.Require().NoError(err, "User %d should create lobby character successfully", i+1)
		s.Equal(lobby.LobbyWorldID, char.WorldID)
		lobbyChars[i] = char
	}

	// Verify all characters exist in database
	var lobbyCharCount int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM characters WHERE world_id = $1
	`, lobby.LobbyWorldID).Scan(&lobbyCharCount)
	s.Require().NoError(err)
	s.Equal(5, lobbyCharCount, "All 5 users should have lobby characters")

	// Verify unique character IDs
	uniqueIDs := make(map[uuid.UUID]bool)
	for _, char := range lobbyChars {
		s.False(uniqueIDs[char.CharacterID], "Each character should have unique ID")
		uniqueIDs[char.CharacterID] = true
	}

	s.T().Log("✓ Multiple users connected to lobby successfully with unique characters")
}

// TestLobbyIntegration_ExistingCharacterRetrieval tests returning existing lobby character
func (s *LobbyIntegrationSuite) TestLobbyIntegration_ExistingCharacterRetrieval() {
	ctx := context.Background()

	// Create a user
	user := &auth.User{
		UserID:       uuid.New(),
		Email:        testutil.GenerateTestEmail(),
		PasswordHash: "hashed_password",
	}
	err := s.repo.CreateUser(ctx, user)
	s.Require().NoError(err)

	// First connection - creates lobby character
	char1, err := s.service.EnsureLobbyCharacter(ctx, user.UserID)
	s.Require().NoError(err)

	// Second connection - should return same character, not create new one
	char2, err := s.service.EnsureLobbyCharacter(ctx, user.UserID)
	s.Require().NoError(err)

	// Verify same character returned
	s.Equal(char1.CharacterID, char2.CharacterID, "Should return existing lobby character")
	s.Equal(char1.Name, char2.Name)
	s.Equal(char1.Role, char2.Role)

	// Verify only one lobby character in database for this user
	var count int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM characters 
		WHERE user_id = $1 AND world_id = $2
	`, user.UserID, lobby.LobbyWorldID).Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count, "Should have exactly one lobby character per user")

	s.T().Log("✓ Existing lobby character retrieved correctly without duplication")
}

// TestLobbyIntegration_CharacterWithExistingWorldCharacter tests appearance copying
func (s *LobbyIntegrationSuite) TestLobbyIntegration_CharacterWithExistingWorldCharacter() {
	ctx := context.Background()

	// Create a user
	user := &auth.User{
		UserID:       uuid.New(),
		Email:        testutil.GenerateTestEmail(),
		PasswordHash: "hashed_password",
	}
	err := s.repo.CreateUser(ctx, user)
	s.Require().NoError(err)

	// Create a world for the user's existing character
	testWorldID := testutil.CreateTestWorld(s.T(), s.db)

	// Create an existing character in another world
	existingChar := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      user.UserID,
		WorldID:     testWorldID,
		Name:        "Hero",
		Role:        "player",
		Appearance:  `{"hair":"blonde","armor":"steel"}`,
	}
	err = s.repo.CreateCharacter(ctx, existingChar)
	s.Require().NoError(err)

	// Now user connects to lobby - should copy appearance from existing character
	lobbyChar, err := s.service.EnsureLobbyCharacter(ctx, user.UserID)
	s.Require().NoError(err)

	// Verify appearance copied (use JSONEq to handle formatting differences)
	s.Equal("Hero", lobbyChar.Name, "Should copy name from existing character")
	s.Equal("player", lobbyChar.Role, "Should use player role")
	s.JSONEq(`{"hair":"blonde","armor":"steel"}`, lobbyChar.Appearance, "Should copy appearance")
	s.Equal(lobby.LobbyWorldID, lobbyChar.WorldID, "Should be in lobby world")

	s.T().Log("✓ Lobby character copied appearance from existing world character")
}

// TestLobbyIntegration_MigrationOrder verifies migrations run in correct order
func (s *LobbyIntegrationSuite) TestLobbyIntegration_MigrationOrder() {
	ctx := context.Background()

	// Verify lobby world was created by migration 000018
	var lobbyWorldName string
	err := s.db.QueryRowContext(ctx, `
		SELECT name FROM worlds WHERE id = $1
	`, lobby.LobbyWorldID).Scan(&lobbyWorldName)
	s.Require().NoError(err, "Lobby world should exist after migrations")
	s.Equal("Lobby", lobbyWorldName)

	// Verify characters table has foreign key constraint
	var constraintExists bool
	err = s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints 
			WHERE constraint_name = 'characters_world_id_fkey'
			AND table_name = 'characters'
		)
	`).Scan(&constraintExists)
	s.Require().NoError(err)
	s.True(constraintExists, "Foreign key constraint should exist")

	s.T().Log("✓ Migrations ran in correct order - lobby world exists before character creation")
}

// Run the integration test suite
func TestLobbyIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping lobby integration tests in short mode")
	}

	suite.Run(t, new(LobbyIntegrationSuite))
}
