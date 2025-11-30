package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"mud-platform-backend/internal/testutil"
)

// RepositoryIntegrationSuite tests the PostgresRepository with a real database
type RepositoryIntegrationSuite struct {
	suite.Suite
	db   *sql.DB
	repo *PostgresRepository
}

// SetupSuite runs once before all tests
func (s *RepositoryIntegrationSuite) SetupSuite() {
	s.db = testutil.SetupTestDB(s.T())
	testutil.RunMigrations(s.T(), s.db)
	s.repo = NewPostgresRepository(s.db)
}

// TearDownSuite runs once after all tests
func (s *RepositoryIntegrationSuite) TearDownSuite() {
	testutil.CloseDB(s.T(), s.db)
}

// SetupTest runs before each test
func (s *RepositoryIntegrationSuite) SetupTest() {
	testutil.TruncateTables(s.T(), s.db)
}

// TestCreateAndRetrieveUser tests user creation and retrieval
func (s *RepositoryIntegrationSuite) TestCreateAndRetrieveUser() {
	ctx := context.Background()

	user := &User{
		UserID:    uuid.New(),
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	// Create user
	err := s.repo.CreateUser(ctx, user)
	s.NoError(err)

	// Retrieve by ID
	retrieved, err := s.repo.GetUserByID(ctx, user.UserID)
	s.NoError(err)
	s.Equal(user.Email, retrieved.Email)
	s.Equal(user.UserID, retrieved.UserID)

	// Retrieve by email
	retrievedByEmail, err := s.repo.GetUserByEmail(ctx, user.Email)
	s.NoError(err)
	s.Equal(user.UserID, retrievedByEmail.UserID)
}

// createTestUser is a helper to create a test user
func (s *RepositoryIntegrationSuite) createTestUser() *User {
	user := &User{
		UserID:    uuid.New(),
		Email:     testutil.GenerateTestEmail(),
		CreatedAt: time.Now(),
	}
	err := s.repo.CreateUser(context.Background(), user)
	s.Require().NoError(err)
	return user
}

// createTestWorld creates a world for testing (to satisfy foreign key)
func (s *RepositoryIntegrationSuite) createTestWorld() uuid.UUID {
	worldID := uuid.New()
	// Insert minimal world record matching actual schema
	_, err := s.db.Exec(`
		INSERT INTO worlds (id, name, shape, created_at)
		VALUES ($1, $2, $3, $4)
	`, worldID, "Test World", "sphere", time.Now())
	s.Require().NoError(err)
	return worldID
}

// createTestCharacter is a helper to create a test character
// Note: Creates its own world to respect UNIQUE(user_id, world_id) constraint
func (s *RepositoryIntegrationSuite) createTestCharacter(userID uuid.UUID) *Character {
	worldID := s.createTestWorld()
	char := &Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     worldID,
		Name:        testutil.GenerateTestName("TestChar"),
		Role:        "player",
		Appearance:  `{"hair":"brown","eyes":"blue"}`,
		Description: "A test character",
		Occupation:  "Tester",
		CreatedAt:   time.Now(),
	}
	err := s.repo.CreateCharacter(context.Background(), char)
	s.Require().NoError(err)
	return char
}

// TestCreateCharacterWithAllFields tests character creation with all new fields
func (s *RepositoryIntegrationSuite) TestCreateCharacterWithAllFields() {
	ctx := context.Background()

	// Create user first
	user := s.createTestUser()

	// Create character with all fields
	char := &Character{
		CharacterID: uuid.New(),
		UserID:      user.UserID,
		WorldID:     s.createTestWorld(),
		Name:        "TestCharacter",
		Role:        "player",
		Appearance:  `{"hair":"brown","eyes":"blue","height":"tall"}`,
		Description: "A brave adventurer seeking fortune and glory",
		Occupation:  "Warrior",
		Position: &Position{
			Latitude:  45.5,
			Longitude: -122.6,
		},
	}

	// Create character
	err := s.repo.CreateCharacter(ctx, char)
	s.NoError(err)

	// Retrieve and verify all fields
	retrieved, err := s.repo.GetCharacter(ctx, char.CharacterID)
	s.NoError(err)
	s.Equal(char.Name, retrieved.Name)
	s.Equal(char.Role, retrieved.Role)
	s.NotEmpty(retrieved.Appearance) // JSONB may reorder keys, just check it exists
	s.Equal(char.Description, retrieved.Description)
	s.Equal(char.Occupation, retrieved.Occupation)
	s.NotNil(retrieved.Position)
	s.InDelta(char.Position.Latitude, retrieved.Position.Latitude, 0.0001)
	s.InDelta(char.Position.Longitude, retrieved.Position.Longitude, 0.0001)
}

// TestWatcherCharacterCreation tests creating a character with watcher role
func (s *RepositoryIntegrationSuite) TestWatcherCharacterCreation() {
	ctx := context.Background()

	user := s.createTestUser()

	watcher := &Character{
		CharacterID: uuid.New(),
		UserID:      user.UserID,
		WorldID:     s.createTestWorld(),
		Name:        "Watcher",
		Role:        "watcher",
		Appearance:  "",  // Watchers don't need appearance
		Description: "An invisible observer",
		Occupation:  "Watcher",
		CreatedAt:   time.Now(),
	}

	err := s.repo.CreateCharacter(ctx, watcher)
	s.NoError(err)

	// Verify watcher fields
	retrieved, err := s.repo.GetCharacter(ctx, watcher.CharacterID)
	s.NoError(err)
	s.Equal("watcher", retrieved.Role)
	s.Equal("Watcher", retrieved.Occupation)
}

// TestGetUserCharacters tests retrieving all characters for a user
func (s *RepositoryIntegrationSuite) TestGetUserCharacters() {
	ctx := context.Background()

	user := s.createTestUser()

	// Create multiple characters (each gets its own world)
	char1 := s.createTestCharacter(user.UserID)
	char2 := s.createTestCharacter(user.UserID)
	char3 := s.createTestCharacter(user.UserID)

	// Retrieve all characters
	chars, err := s.repo.GetUserCharacters(ctx, user.UserID)
	s.NoError(err)
	s.Len(chars, 3)

	// Verify character IDs are present
	charIDs := make(map[uuid.UUID]bool)
	for _, c := range chars {
		charIDs[c.CharacterID] = true
	}
	s.True(charIDs[char1.CharacterID])
	s.True(charIDs[char2.CharacterID])
	s.True(charIDs[char3.CharacterID])
}

// TestGetCharacterByUserAndWorld tests retrieving character by user and world
func (s *RepositoryIntegrationSuite) TestGetCharacterByUserAndWorld() {
	ctx := context.Background()

	user := s.createTestUser()
	worldID := s.createTestWorld()

	// Create character in specific world
	char := &Character{
		CharacterID: uuid.New(),
		UserID:      user.UserID,
		WorldID:     worldID,
		Name:        "TestChar",
		Role:        "player",
		Appearance:  `{"hair":"brown"}`,
		CreatedAt:   time.Now(),
	}
	err := s.repo.CreateCharacter(ctx, char)
	s.NoError(err)

	// Retrieve by user and world
	retrieved, err := s.repo.GetCharacterByUserAndWorld(ctx, user.UserID, worldID)
	s.NoError(err)
	s.Equal(char.CharacterID, retrieved.CharacterID)

	// Test non-existent combination
	otherWorld := uuid.New()
	notFound, err := s.repo.GetCharacterByUserAndWorld(ctx, user.UserID, otherWorld)
	s.Error(err)
	s.Nil(notFound)
}

// TestUpdateCharacter tests character updates
func (s *RepositoryIntegrationSuite) TestUpdateCharacter() {
	ctx := context.Background()

	user := s.createTestUser()
	char := s.createTestCharacter(user.UserID)

	// Update character
	char.Name = "UpdatedName"
	char.Position = &Position{
		Latitude:  40.7,
		Longitude: -74.0,
	}
	now := time.Now()
	char.LastPlayed = &now

	err := s.repo.UpdateCharacter(ctx, char)
	s.NoError(err)

	// Verify update
	retrieved, err := s.repo.GetCharacter(ctx, char.CharacterID)
	s.NoError(err)
	s.Equal("UpdatedName", retrieved.Name)
	s.NotNil(retrieved.Position)
	s.InDelta(40.7, retrieved.Position.Latitude, 0.0001)
	s.NotNil(retrieved.LastPlayed)
}

// TestDuplicateEmailError tests that duplicate emails are rejected
func (s *RepositoryIntegrationSuite) TestDuplicateEmailError() {
	ctx := context.Background()

	email := "duplicate@example.com"

	user1 := &User{
		UserID:    uuid.New(),
		Email:     email,
		CreatedAt: time.Now(),
	}

	err := s.repo.CreateUser(ctx, user1)
	s.NoError(err)

	// Try to create another user with same email
	user2 := &User{
		UserID:    uuid.New(),
		Email:     email,
		CreatedAt: time.Now(),
	}

	err = s.repo.CreateUser(ctx, user2)
	s.Error(err)
	s.Equal(ErrDuplicateEmail, err)
}

// TestCharacterNotFound tests error handling for non-existent characters
func (s *RepositoryIntegrationSuite) TestCharacterNotFound() {
	ctx := context.Background()

	nonExistentID := uuid.New()

	char, err := s.repo.GetCharacter(ctx, nonExistentID)
	s.Error(err)
	s.Nil(char)
	s.Equal(ErrCharacterNotFound, err)
}

// TestRunRepositoryIntegrationSuite runs the integration test suite
func TestRepositoryIntegrationSuite(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(RepositoryIntegrationSuite))
}
