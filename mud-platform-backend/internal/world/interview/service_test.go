package interview

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartInterview_Success tests starting a new interview session
func TestStartInterview_Success(t *testing.T) {
	repo := NewMockRepository()
	service := NewServiceWithRepository(nil, repo) // LLM client can be nil for basic tests

	userID := uuid.New()

	// Start interview (old signature returns session, question, error)
	session, _, err := service.StartInterview(userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, userID, session.PlayerID)
	assert.NotEmpty(t, session.ID)
	assert.False(t, session.State.IsComplete)
}

// TestProcessResponse_ValidSession tests processing a response
func TestProcessResponse_ValidSession(t *testing.T) {
	repo := NewMockRepository()
	service := NewServiceWithRepository(nil, repo)

	userID := uuid.New()

	// Start interview
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Process response (will fail without LLM but tests the flow)
	_, _, err = service.ProcessResponse(session.ID, "I want a fantasy world")

	// Expected to fail without LLM client
	assert.Error(t, err)
}

// TestProcessResponse_InvalidSession tests handling of invalid session
func TestProcessResponse_InvalidSession(t *testing.T) {
	repo := NewMockRepository()
	service := NewServiceWithRepository(nil, repo)

	fakeSessionID := uuid.New()

	// Try to process response for non-existent session
	_, _, err := service.ProcessResponse(fakeSessionID, "test")

	// Assert error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestWorldConfiguration_Creation tests configuration saving
func TestWorldConfiguration_Creation(t *testing.T) {
	repo := NewMockRepository()

	worldID := uuid.New()
	config := &WorldConfiguration{
		WorldID:         &worldID,
		Theme:           "Fantasy",
		TechLevel:       "medieval",
		SentientSpecies: []string{"Human", "Elf", "Dwarf"},
	}

	// Save configuration (new signature takes config only)
	err := repo.SaveConfiguration(config)
	require.NoError(t, err)

	// Retrieve configuration
	retrieved, err := repo.GetConfigurationByWorldID(worldID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Fantasy", retrieved.Theme)
	assert.Equal(t, "medieval", retrieved.TechLevel)
	assert.Len(t, retrieved.SentientSpecies, 3)
}

// TestGetActiveSessionForUser_None tests when no session exists
func TestGetActiveSessionForUser_None(t *testing.T) {
	repo := NewMockRepository()

	userID := uuid.New()
	ctx := context.Background()

	// Get active session (none exists)
	session, err := repo.GetActiveSessionForUser(ctx, userID)

	// Assert no error but nil session
	require.NoError(t, err)
	assert.Nil(t, session)
}

// TestMultipleSessions_Isolation tests session isolation
func TestMultipleSessions_Isolation(t *testing.T) {
	repo := NewMockRepository()
	service := NewServiceWithRepository(nil, repo)

	user1 := uuid.New()
	user2 := uuid.New()

	// Start sessions for both users
	session1, _, err := service.StartInterview(user1)
	require.NoError(t, err)

	session2, _, err := service.StartInterview(user2)
	require.NoError(t, err)

	// Assert different sessions
	assert.NotEqual(t, session1.ID, session2.ID)
	assert.Equal(t, user1, session1.PlayerID)
	assert.Equal(t, user2, session2.PlayerID)
}
