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
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Mock question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New()

	// Start interview (old signature returns session, question, error)
	session, _, err := service.StartInterview(ctx, userID)

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
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Next question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New()

	// Start interview
	_, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	// Process response
	_, _, err = service.ProcessResponse(ctx, userID, "I want a fantasy world")

	// Should succeed with mock LLM
	assert.NoError(t, err)
}

// TestProcessResponse_InvalidSession tests handling of invalid session
func TestProcessResponse_InvalidSession(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New() // User with no session

	// Try to process response for user with no session
	_, _, err := service.ProcessResponse(ctx, userID, "test")

	// Assert error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active interview")
}

// TestWorldConfiguration_Creation tests configuration saving
func TestWorldConfiguration_Creation(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	worldID := uuid.New()
	config := &WorldConfiguration{
		WorldID:         &worldID,
		Theme:           "Fantasy",
		TechLevel:       "medieval",
		SentientSpecies: []string{"Human", "Elf", "Dwarf"},
	}

	// Save configuration (new signature takes config only)
	err := repo.SaveConfiguration(ctx, config)
	require.NoError(t, err)

	// Retrieve configuration
	retrieved, err := repo.GetConfigurationByWorldID(ctx, worldID)

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
	// Use service.GetActiveInterview instead of repo method if repo method doesn't exist or is different
	mockLLM := &MockLLM{}
	worldRepo := NewMockWorldRepository()
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	session, err := service.GetActiveInterview(ctx, userID)

	// Assert no error but nil session
	require.NoError(t, err)
	assert.Nil(t, session)
}

// TestMultipleSessions_Isolation tests session isolation
func TestMultipleSessions_Isolation(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Mock question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	user1 := uuid.New()
	user2 := uuid.New()

	// Start sessions for both users
	session1, _, err := service.StartInterview(ctx, user1)
	require.NoError(t, err)

	session2, _, err := service.StartInterview(ctx, user2)
	require.NoError(t, err)

	// Assert different sessions
	assert.NotEqual(t, session1.ID, session2.ID)
	assert.Equal(t, user1, session1.PlayerID)
	assert.Equal(t, user2, session2.PlayerID)
}

// TestResumeInterview tests resuming an existing interview
func TestResumeInterview(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Resumed question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New()

	// Start interview
	session, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	// Resume interview
	resumedSession, question, err := service.ResumeInterview(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, resumedSession.ID)
	assert.Equal(t, "Resumed question?", question)
}

// TestGetActiveInterview tests retrieving active interview via service
func TestGetActiveInterview(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New()

	// No active interview initially
	session, err := service.GetActiveInterview(ctx, userID)
	assert.NoError(t, err)
	assert.Nil(t, session)

	// Start interview
	startedSession, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	// Get active interview
	session, err = service.GetActiveInterview(ctx, userID)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, startedSession.ID, session.ID)
}

// TestEditAnswer tests editing a previous answer
func TestEditAnswer(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New()
	session, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	// Answer first question
	_, _, err = service.ProcessResponse(ctx, userID, "Original Answer")
	require.NoError(t, err)

	// Edit answer
	// Note: We need to know the topic name. First topic is usually "Theme" or similar from AllTopics[0]
	firstTopicName := AllTopics[0].Name
	err = service.EditAnswer(ctx, userID, session.ID, firstTopicName, "Edited Answer")
	require.NoError(t, err)

	// Verify update
	updatedSession, err := repo.GetInterview(ctx, userID)
	require.NoError(t, err)

	// Check answers via repo
	answers, err := repo.GetAnswers(ctx, updatedSession.ID)
	require.NoError(t, err)
	found := false
	for _, a := range answers {
		if a.QuestionIndex == 0 && a.AnswerText == "Edited Answer" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// TestCompleteInterview tests completing an interview and extracting config
func TestCompleteInterview(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			// Return raw JSON without escaping
			return "{\"theme\": \"Fantasy\", \"techLevel\": \"medieval\", \"planetSize\": \"medium\", \"sentientSpecies\": [\"Human\"]}", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New()
	session, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	// Answer all questions to trigger configuration extraction
	for i := 0; i < len(AllTopics)-1; i++ {
		_, _, err := service.ProcessResponse(ctx, userID, "Test answer "+AllTopics[i].Name)
		require.NoError(t, err)
	}

	// Answer final question (World Name)
	resp, completed, err := service.ProcessResponse(ctx, userID, "Test Fantasy World")
	require.NoError(t, err)
	assert.False(t, completed, "Should not complete yet - review phase")
	assert.Contains(t, resp, "Here is the vision for your world")

	// Confirm
	resp, completed, err = service.ProcessResponse(ctx, userID, "yes")
	require.NoError(t, err)
	assert.True(t, completed, "Should complete after confirmation")
	assert.Contains(t, resp, "Thank you")

	// Complete interview - configuration should already be saved by ProcessResponse
	config, err := service.CompleteInterview(ctx, userID, session.ID)
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "Fantasy", config.Theme)
	assert.Equal(t, "Test Fantasy World", config.WorldName)
}

// TestStartInterview_Instructions tests that the first message contains instructions
func TestStartInterview_Instructions(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Mock question?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Start interview
	_, question, err := service.StartInterview(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, question, "Welcome, creator")
	assert.Contains(t, question, "using the \"reply\" command")
	assert.Contains(t, question, "Mock question?")
}

// TestProcessResponse_ChangeCommand tests the change command handling
func TestProcessResponse_ChangeCommand(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Review?", nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New()
	session, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	// Answer first question
	_, _, err = service.ProcessResponse(ctx, userID, "Original Answer")
	require.NoError(t, err)

	// Send change command
	// First topic is "Core Concept"
	resp, _, err := service.ProcessResponse(ctx, userID, "change Core Concept to Sci-Fi")
	require.NoError(t, err)
	assert.Contains(t, resp, "Updated Core Concept to 'Sci-Fi'")

	// Check DB
	answers, err := repo.GetAnswers(ctx, session.ID)
	require.NoError(t, err)
	found := false
	for _, a := range answers {
		if a.QuestionIndex == 0 && a.AnswerText == "Sci-Fi" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find updated answer")
}

// TestProcessResponse_ReviewMode tests the review mode flow
func TestProcessResponse_ReviewMode(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			// Extract config returns JSON
			return `{"theme": "Fantasy", "worldName": "MyWorld"}`, nil
		},
	}
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	userID := uuid.New()
	_, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	// Answer all questions
	// Note: AllTopics len is 14 in previous files, but we should rely on dynamic len
	for i := 0; i < len(AllTopics); i++ {
		_, completed, err := service.ProcessResponse(ctx, userID, "Answer")
		require.NoError(t, err)
		assert.False(t, completed)
	}

	// Now in review mode. Next response "yes" should trigger completion.
	// But first response entered review mode summary.
	// So calling ProcessResponse with "yes"
	resp, completed, err := service.ProcessResponse(ctx, userID, "yes")
	require.NoError(t, err)
	assert.True(t, completed)
	assert.Contains(t, resp, "Thank you")
}
