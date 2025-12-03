package interview_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mud-platform-backend/internal/world/interview"
)

// TestEndToEndInterviewWithNaming tests the complete interview flow:
// 1. Start interview
// 2. Answer all questions
// 3. Provide a world name
// 4. Complete interview
// 5. Extract and save configuration
func TestEndToEndInterviewWithNaming(t *testing.T) {
	// Setup
	repo := interview.NewMockRepository()
	mockLLM := &interview.MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			// For extraction, return full JSON with world name
			// Use strings.Contains to be more robust
			if strings.Contains(prompt, "You are a data extraction assistant") {
				return `{
					"theme": "High Fantasy",
					"tone": "Epic",
					"techLevel": "medieval",
					"magicLevel": "common",
					"planetSize": "Earth-sized",
					"sentientSpecies": ["Humans", "Elves", "Dwarves"]
				}`, nil
			}
			// For questions, return mock questions
			return "Next question about your world?", nil
		},
	}
	service := interview.NewServiceWithRepository(mockLLM, repo)
	userID := uuid.New()

	// Step 1: Start interview
	session, question, err := service.StartInterview(userID)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, question)
	assert.False(t, session.State.IsComplete)

	// Step 2: Answer all questions except world name
	totalTopics := len(interview.AllTopics)
	for i := 0; i < totalTopics-1; i++ {
		resp, completed, err := service.ProcessResponse(session.ID, "Test answer for topic "+interview.AllTopics[i].Name)
		require.NoError(t, err)
		assert.False(t, completed, "Interview should not complete until all questions answered")
		assert.NotEmpty(t, resp)
	}

	// Step 3: Provide world name (final question)
	worldName := "Aethoria - Realm of Magic"
	resp, completed, err := service.ProcessResponse(session.ID, worldName)
	require.NoError(t, err)
	assert.True(t, completed, "Interview should complete after world name")
	assert.Contains(t, resp, "Thank you")

	// Step 4: Verify session state
	updatedSession, err := repo.GetInterview(session.ID)
	require.NoError(t, err)
	assert.True(t, updatedSession.State.IsComplete)
	assert.Equal(t, worldName, updatedSession.State.Answers["World Name"])

	// Step 5: Complete interview and extract configuration
	config, err := service.CompleteInterview(session.ID)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Step 6: Verify configuration
	assert.Equal(t, worldName, config.WorldName)
	assert.Equal(t, "High Fantasy", config.Theme)
	assert.Equal(t, "medieval", config.TechLevel)
	assert.Len(t, config.SentientSpecies, 3)
	assert.Contains(t, config.SentientSpecies, "Humans")
}

// TestEndToEndWithDuplicateName tests handling of duplicate world names
func TestEndToEndWithDuplicateName(t *testing.T) {
	// Setup
	repo := interview.NewMockRepository()
	mockLLM := &interview.MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			// For name generation, return suggestions
			if strings.Contains(prompt, "generate EXACTLY") || strings.Contains(prompt, "Based on the following world description") {
				return "Alt Name 1\nAlt Name 2", nil
			}
			return "Next question?", nil
		},
	}
	service := interview.NewServiceWithRepository(mockLLM, repo)

	// Create existing world with a name
	existingConfig := &interview.WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "Taken World",
		Theme:           "Fantasy",
		TechLevel:       "medieval",
		PlanetSize:      "medium",
		SentientSpecies: []string{"Human"},
	}
	err := repo.SaveConfiguration(existingConfig)
	require.NoError(t, err)

	// Start new interview
	userID := uuid.New()
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Answer all questions except world name
	for i := 0; i < len(interview.AllTopics)-1; i++ {
		_, _, err := service.ProcessResponse(session.ID, "Answer "+interview.AllTopics[i].Name)
		require.NoError(t, err)
	}

	// Try to use taken name
	resp, completed, err := service.ProcessResponse(session.ID, "Taken World")
	require.NoError(t, err)
	assert.False(t, completed, "Should not complete with duplicate name")
	assert.Contains(t, resp, "already taken")
	assert.Contains(t, resp, "Alt Name")

	// Provide unique name
	resp, completed, err = service.ProcessResponse(session.ID, "My Unique World")
	require.NoError(t, err)
	assert.True(t, completed, "Should complete with unique name")

	// Verify
	updatedSession, err := repo.GetInterview(session.ID)
	require.NoError(t, err)
	assert.Equal(t, "My Unique World", updatedSession.State.Answers["World Name"])
}

// TestEndToEndInvalidWorldName tests validation of invalid world names
func TestEndToEndInvalidWorldName(t *testing.T) {
	repo := interview.NewMockRepository()
	mockLLM := &interview.MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Next question?", nil
		},
	}
	service := interview.NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Answer all questions except world name
	for i := 0; i < len(interview.AllTopics)-1; i++ {
		_, _, err := service.ProcessResponse(session.ID, "Answer")
		require.NoError(t, err)
	}

	// Test cases for invalid names
	testCases := []struct {
		name          string
		worldName     string
		expectedError string
	}{
		{"empty", "", "Please provide a name"},
		{"special chars", "World@#$%", "not valid"},
		{"too long", string(make([]byte, 101)), "not valid"}, // 101 characters
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, completed, err := service.ProcessResponse(session.ID, tc.worldName)
			require.NoError(t, err)
			assert.False(t, completed, "Should not complete with invalid name")
			assert.Contains(t, resp, tc.expectedError)
		})

		//  Reload session for next test case
		session, err = repo.GetInterview(session.ID)
		require.NoError(t, err)
	}

	// Valid name should work
	_, completed, err := service.ProcessResponse(session.ID, "Valid World Name")
	require.NoError(t, err)
	assert.True(t, completed)
}

// TestInterviewResumption tests resuming an interview and completing with world name
func TestInterviewResumption(t *testing.T) {
	repo := interview.NewMockRepository()
	mockLLM := &interview.MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Resumed question?", nil
		},
	}
	service := interview.NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Answer half the questions
	midPoint := len(interview.AllTopics) / 2
	for i := 0; i < midPoint; i++ {
		_, _, err := service.ProcessResponse(session.ID, "Answer "+interview.AllTopics[i].Name)
		require.NoError(t, err)
	}

	// Resume interview
	resumedSession, question, err := service.ResumeInterview(session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, resumedSession.ID)
	assert.NotEmpty(t, question)
	assert.Equal(t, midPoint, resumedSession.State.CurrentTopicIndex)

	// Continue answering
	for i := midPoint; i < len(interview.AllTopics)-1; i++ {
		_, _, err := service.ProcessResponse(session.ID, "Answer")
		require.NoError(t, err)
	}

	// Answer world name question
	resp, completed, err := service.ProcessResponse(session.ID, "Resumed World")
	require.NoError(t, err)
	assert.True(t, completed)
	assert.Contains(t, resp, "Thank you")
}

// TestGetActiveInterview tests retrieving active interview for a user
func TestGetActiveInterviewFlow(t *testing.T) {
	repo := interview.NewMockRepository()
	mockLLM := &interview.MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Question?", nil
		},
	}
	service := interview.NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()

	// No active interview initially
	session, err := service.GetActiveInterview(userID)
	assert.NoError(t, err)
	assert.Nil(t, session)

	// Start interview
	startedSession, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Should now have active interview
	session, err = service.GetActiveInterview(userID)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, startedSession.ID, session.ID)

	// Complete interview (answer all questions)
	for i := 0; i < len(interview.AllTopics); i++ {
		if i == len(interview.AllTopics)-1 {
			// Last question is world name
			_, _, err = service.ProcessResponse(session.ID, "Final World")
		} else {
			_, _, err = service.ProcessResponse(session.ID, "Answer")
		}
		require.NoError(t, err)
	}

	// Note: After completion, the interview may or may not be retrievable depending on implementation
	// The main goal of this test is to verify retrieval of active incomplete interviews
}

// TestEditAnswerDuringInterview tests editing a previous answer
func TestEditAnswerDuringInterview(t *testing.T) {
	repo := interview.NewMockRepository()
	mockLLM := &interview.MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Question?", nil
		},
	}
	service := interview.NewServiceWithRepository(mockLLM, repo)

	userID := uuid.New()
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Answer first question
	firstTopicName := interview.AllTopics[0].Name
	_, _, err = service.ProcessResponse(session.ID, "Original Answer")
	require.NoError(t, err)

	// Edit answer
	err = service.EditAnswer(session.ID, firstTopicName, "Edited Answer")
	require.NoError(t, err)

	// Verify
	updated, err := repo.GetInterview(session.ID)
	require.NoError(t, err)
	assert.Equal(t, "Edited Answer", updated.State.Answers[firstTopicName])
}

// TestMultipleUsersWithSameWorldName tests that world names are globally unique
func TestMultipleUsersWithSameWorldName(t *testing.T) {
	repo := interview.NewMockRepository()
	mockLLM := &interview.MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			if len(prompt) > 50 && prompt[:50] == "Based on the following world description, generat" {
				return "Alternative1\nAlternative2", nil
			}
			return "Question?", nil
		},
	}
	service := interview.NewServiceWithRepository(mockLLM, repo)

	// User 1 creates world with name
	user1 := uuid.New()
	session1, _, err := service.StartInterview(user1)
	require.NoError(t, err)

	for i := 0; i < len(interview.AllTopics)-1; i++ {
		_, _, err := service.ProcessResponse(session1.ID, "Answer")
		require.NoError(t, err)
	}
	_, _, err = service.ProcessResponse(session1.ID, "Shared World Name")
	require.NoError(t, err)

	// Complete and save
	config1, err := service.CompleteInterview(session1.ID)
	require.NoError(t, err)
	assert.Equal(t, "Shared World Name", config1.WorldName)

	// User 2 tries to use same name
	user2 := uuid.New()
	session2, _, err := service.StartInterview(user2)
	require.NoError(t, err)

	for i := 0; i < len(interview.AllTopics)-1; i++ {
		_, _, err := service.ProcessResponse(session2.ID, "Answer")
		require.NoError(t, err)
	}

	// Should be rejected
	resp, completed, err := service.ProcessResponse(session2.ID, "Shared World Name")
	require.NoError(t, err)
	assert.False(t, completed)
	assert.Contains(t, resp, "already taken")

	// Alternative name should work
	resp, completed, err = service.ProcessResponse(session2.ID, "User2 World")
	require.NoError(t, err)
	assert.True(t, completed)
}
