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
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Mock question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo)

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
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Next question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()

	// Start interview
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Process response
	_, _, err = service.ProcessResponse(session.ID, "I want a fantasy world")

	// Should succeed with mock LLM
	assert.NoError(t, err)
}

// TestProcessResponse_InvalidSession tests handling of invalid session
func TestProcessResponse_InvalidSession(t *testing.T) {
	repo := NewMockRepository()
	mockLLM := &MockLLM{}
	service := NewServiceWithRepository(mockLLM, repo)

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
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Mock question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo)

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

// TestResumeInterview tests resuming an existing interview
func TestResumeInterview(t *testing.T) {
	repo := NewMockRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Resumed question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()

	// Start interview
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Resume interview
	resumedSession, question, err := service.ResumeInterview(session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, resumedSession.ID)
	assert.Equal(t, "Resumed question?", question)
}

// TestGetActiveInterview tests retrieving active interview via service
func TestGetActiveInterview(t *testing.T) {
	repo := NewMockRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()

	// No active interview initially
	session, err := service.GetActiveInterview(userID)
	assert.NoError(t, err)
	assert.Nil(t, session)

	// Start interview
	startedSession, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Get active interview
	session, err = service.GetActiveInterview(userID)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, startedSession.ID, session.ID)
}

// TestEditAnswer tests editing a previous answer
func TestEditAnswer(t *testing.T) {
	repo := NewMockRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Answer first question
	_, _, err = service.ProcessResponse(session.ID, "Original Answer")
	require.NoError(t, err)

	// Edit answer
	// Note: We need to know the topic name. First topic is usually "Theme" or similar from AllTopics[0]
	firstTopicName := AllTopics[0].Name
	err = service.EditAnswer(session.ID, firstTopicName, "Edited Answer")
	require.NoError(t, err)

	// Verify update
	updatedSession, err := repo.GetInterview(session.ID)
	require.NoError(t, err)
	assert.Equal(t, "Edited Answer", updatedSession.State.Answers[firstTopicName])
}

// TestCompleteInterview tests completing an interview and extracting config
func TestCompleteInterview(t *testing.T) {
	repo := NewMockRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return `{"theme": "Fantasy", "techLevel": "medieval", "planetSize": "medium", "sentientSpecies": ["Human"]}`, nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Add world name answer (required for validation)
	session.State.Answers["World Name"] = "Test Fantasy World"

	// Manually mark session as complete for testing
	session.State.IsComplete = true
	repo.UpdateInterview(session)

	// Complete interview
	config, err := service.CompleteInterview(session.ID)
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "Fantasy", config.Theme)
	assert.Equal(t, "Test Fantasy World", config.WorldName)
}
