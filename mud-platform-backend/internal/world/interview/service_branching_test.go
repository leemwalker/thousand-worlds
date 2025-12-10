package interview

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBranchingLLM is a mock implementation of the LLMClient interface for testing.
type MockBranchingLLM struct {
	GenerateFunc func(prompt string) (string, error)
}

// Generate calls the GenerateFunc if it's set, otherwise returns a default mock response.
func (m *MockBranchingLLM) Generate(prompt string) (string, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(prompt)
	}
	return "Mock LLM Response", nil
}

// TestBranchingLogic verifies the behavior at the Branch topic (Q7)
func TestBranchingLogic(t *testing.T) {
	// Setup
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockBranchingLLM{
		GenerateFunc: func(prompt string) (string, error) {
			if strings.Contains(prompt, "Factions") {
				return "What factions exist?", nil
			}
			return "Mock Question", nil
		},
	}
	service := NewService(mockLLM, repo, worldRepo)

	ctx := context.Background()
	playerID := uuid.New()

	// Create a session
	interview, err := repo.CreateInterview(ctx, playerID)
	require.NoError(t, err)

	// Manually advance to Branch topic (Index 6)
	err = repo.UpdateQuestionIndex(ctx, interview.ID, 6)
	require.NoError(t, err)

	// Case 1: "continue" - should advance to next question
	t.Run("Continue", func(t *testing.T) {
		resp, complete, err := service.ProcessResponse(ctx, playerID, "continue")
		assert.NoError(t, err)
		assert.False(t, complete)

		// Check message
		assert.Equal(t, "What factions exist?", resp)

		// Check state in repo
		updatedInterview, _ := repo.GetInterview(ctx, playerID)
		assert.Equal(t, 7, updatedInterview.CurrentQuestionIndex)

		answers, _ := repo.GetAnswers(ctx, interview.ID)
		found := false
		for _, a := range answers {
			if a.QuestionIndex == 6 && a.AnswerText == "continue" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should save 'continue' as answer")
	})

	// Case 2: Invalid input for Branch
	t.Run("InvalidBranchInput", func(t *testing.T) {
		// Reset interview index for this sub-test (since previous run changed it to 7)
		repo.UpdateQuestionIndex(ctx, interview.ID, 6)

		// Also wait a bit to ensure UpdatedAt is different if needed, or just rely on logic
		time.Sleep(1 * time.Millisecond)

		resp, complete, err := service.ProcessResponse(ctx, playerID, "something else")
		assert.NoError(t, err)
		assert.False(t, complete)
		assert.Contains(t, resp, "reply with 'the name is <world name>'", "Should return specific instruction")

		// Verify no state change for invalid input
		updatedInterview, _ := repo.GetInterview(ctx, playerID)
		assert.Equal(t, 6, updatedInterview.CurrentQuestionIndex)
	})
}

// TestGlobalInterrupt verifies the regex interrupt works after Q7
func TestGlobalInterrupt(t *testing.T) {
	// Setup
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockBranchingLLM{
		GenerateFunc: func(prompt string) (string, error) {
			// Mock extraction response for completion
			return `{"theme": "fantasy", "worldName": "Interrupted World", "sentientSpecies": ["humans"]}`, nil
		},
	}
	service := NewService(mockLLM, repo, worldRepo)

	ctx := context.Background()
	playerID := uuid.New()

	interview, err := repo.CreateInterview(ctx, playerID)
	require.NoError(t, err)

	// Advance past Q7 (e.g. at Factions, index 7)
	err = repo.UpdateQuestionIndex(ctx, interview.ID, 7)
	require.NoError(t, err)

	resp, complete, err := service.ProcessResponse(ctx, playerID, "The Name Is Interrupted World")
	assert.NoError(t, err)
	assert.True(t, complete)
	assert.Contains(t, resp, "Your world is being forged")

	// Verify world created
	worlds, _ := worldRepo.GetWorldsByOwner(ctx, playerID)
	assert.NotEmpty(t, worlds)
	assert.Equal(t, "Interrupted World", worlds[0].Name)

	// Verify interview status updated
	updatedInterview, _ := repo.GetInterview(ctx, playerID)
	assert.Equal(t, StatusCompleted, updatedInterview.Status)
}

// TestBranchingLogic_ImmediateName tests naming at branch point
func TestBranchingLogic_ImmediateName(t *testing.T) {
	// Setup
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := &MockBranchingLLM{
		GenerateFunc: func(prompt string) (string, error) {
			// Mock extraction response for completion
			return `{"theme": "fantasy", "worldName": "Immediate World", "sentientSpecies": ["humans"]}`, nil
		},
	}
	service := NewService(mockLLM, repo, worldRepo)

	ctx := context.Background()
	playerID := uuid.New()

	interview, err := repo.CreateInterview(ctx, playerID)
	require.NoError(t, err)

	// At Branch topic (Index 6)
	err = repo.UpdateQuestionIndex(ctx, interview.ID, 6)
	require.NoError(t, err)

	resp, complete, err := service.ProcessResponse(ctx, playerID, "the name is Immediate World")
	assert.NoError(t, err)
	assert.True(t, complete)
	assert.Contains(t, resp, "Your world is being forged")

	// Verify world created
	worlds, _ := worldRepo.GetWorldsByOwner(ctx, playerID)
	assert.NotEmpty(t, worlds)
	assert.Equal(t, "Immediate World", worlds[0].Name)

	// Verify interview status updated
	updatedInterview, _ := repo.GetInterview(ctx, playerID)
	assert.Equal(t, StatusCompleted, updatedInterview.Status)
}
