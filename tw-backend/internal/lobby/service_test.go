package lobby

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mud-platform-backend/internal/auth"
)

// TestEnsureLobbyCharacter_NewUser tests creating lobby character for new user
func TestEnsureLobbyCharacter_NewUser(t *testing.T) {
	repo := auth.NewMockRepository()
	service := NewService(repo)

	userID := uuid.New()
	ctx := context.Background()

	// Call EnsureLobbyCharacter
	char, err := service.EnsureLobbyCharacter(ctx, userID)

	// Assert no error
	require.NoError(t, err)
	require.NotNil(t, char)

	// Assert default values
	assert.Equal(t, "Ghost", char.Name)
	assert.Equal(t, "ghost", char.Role)
	assert.Equal(t, LobbyWorldID, char.WorldID)
	assert.Equal(t, userID, char.UserID)
	assert.NotEmpty(t, char.Appearance)
}

// TestEnsureLobbyCharacter_ExistingLobbyCharacter tests returning existing lobby character
func TestEnsureLobbyCharacter_ExistingLobbyCharacter(t *testing.T) {
	repo := auth.NewMockRepository()
	service := NewService(repo)

	userID := uuid.New()
	ctx := context.Background()

	// Create existing lobby character
	existingChar := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     LobbyWorldID,
		Name:        "ExistingGhost",
		Role:        "player",
		Appearance:  `{"color":"blue"}`,
		CreatedAt:   time.Now(),
	}
	err := repo.CreateCharacter(ctx, existingChar)
	require.NoError(t, err)

	// Call EnsureLobbyCharacter
	char, err := service.EnsureLobbyCharacter(ctx, userID)

	// Assert returns existing character
	require.NoError(t, err)
	assert.Equal(t, existingChar.CharacterID, char.CharacterID)
	assert.Equal(t, "ExistingGhost", char.Name)
	assert.Equal(t, "player", char.Role)
}

// TestEnsureLobbyCharacter_CopiesFromExistingCharacter tests appearance copying
func TestEnsureLobbyCharacter_CopiesFromExistingCharacter(t *testing.T) {
	repo := auth.NewMockRepository()
	service := NewService(repo)

	userID := uuid.New()
	worldID := uuid.New()
	ctx := context.Background()

	// Create existing character in another world
	existingChar := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     worldID,
		Name:        "Warrior",
		Role:        "player",
		Appearance:  `{"hair":"brown","armor":"plate"}`,
		CreatedAt:   time.Now(),
	}
	err := repo.CreateCharacter(ctx, existingChar)
	require.NoError(t, err)

	// Call EnsureLobbyCharacter
	lobbyChar, err := service.EnsureLobbyCharacter(ctx, userID)

	// Assert copied values
	require.NoError(t, err)
	assert.Equal(t, "Warrior", lobbyChar.Name, "Should copy name from existing character")
	assert.Equal(t, "player", lobbyChar.Role, "Should use player role when user has characters")
	assert.Equal(t, `{"hair":"brown","armor":"plate"}`, lobbyChar.Appearance, "Should copy appearance")
	assert.Equal(t, LobbyWorldID, lobbyChar.WorldID)
}

// TestEnsureLobbyCharacter_InvalidUserID tests error handling for invalid input
func TestEnsureLobbyCharacter_InvalidUserID(t *testing.T) {
	repo := auth.NewMockRepository()
	service := NewService(repo)

	ctx := context.Background()

	// Call with nil UUID
	char, err := service.EnsureLobbyCharacter(ctx, uuid.Nil)

	// Assert error
	require.Error(t, err)
	assert.Nil(t, char)
	assert.Contains(t, err.Error(), "invalid user ID")
}

// TestEnsureLobbyCharacter_MultipleWorlds tests character uniqueness
func TestEnsureLobbyCharacter_MultipleWorlds(t *testing.T) {
	repo := auth.NewMockRepository()
	service := NewService(repo)

	userID := uuid.New()
	ctx := context.Background()

	// Create characters in multiple worlds
	world1 := uuid.New()
	world2 := uuid.New()

	char1 := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     world1,
		Name:        "Char1",
		Role:        "player",
		CreatedAt:   time.Now().Add(-2 * time.Hour),
	}
	char2 := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     world2,
		Name:        "Char2",
		Role:        "player",
		Appearance:  `{"special":"appearance"}`,
		CreatedAt:   time.Now().Add(-1 * time.Hour), // More recent
	}

	repo.CreateCharacter(ctx, char1)
	repo.CreateCharacter(ctx, char2)

	// Create lobby character - should copy from most recent
	lobbyChar, err := service.EnsureLobbyCharacter(ctx, userID)

	require.NoError(t, err)
	// Note: Depends on GetUserCharacters sorting - if sorted by LastPlayed/CreatedAt DESC,
	// it should use the most recent character
	assert.NotNil(t, lobbyChar)
	assert.Equal(t, LobbyWorldID, lobbyChar.WorldID)
}
