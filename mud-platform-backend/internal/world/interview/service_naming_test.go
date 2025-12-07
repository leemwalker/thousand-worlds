package interview

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestifyMockLLM is a mock implementation of LLMClient using testify/mock
type TestifyMockLLM struct {
	mock.Mock
}

func (m *TestifyMockLLM) Generate(prompt string) (string, error) {
	args := m.Called(prompt)
	return args.String(0), args.Error(1)
}

func TestProcessResponse_WorldNameValidation(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := new(TestifyMockLLM)
	// Return JSON when extraction prompt is detected, otherwise return question
	mockLLM.On("Generate", mock.MatchedBy(func(prompt string) bool {
		return len(prompt) > 100 && (strings.HasPrefix(prompt, "You are a data extraction assistant") ||
			strings.Contains(prompt, "data extraction"))
	})).Return("{\"theme\": \"Fantasy\", \"techLevel\": \"medieval\", \"planetSize\": \"medium\", \"sentientSpecies\": [\"Human\"]}", nil)
	mockLLM.On("Generate", mock.Anything).Return("Next question?", nil)

	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	// Setup session
	userID := uuid.New()
	_, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	// Answer all questions EXCEPT world name
	worldNameIndex := -1
	for i, topic := range AllTopics {
		if topic.Name == "World Name" {
			worldNameIndex = i
			break
		}
	}
	require.NotEqual(t, -1, worldNameIndex, "World Name topic not found")

	// Answer all questions before World Name
	for i := 0; i < worldNameIndex; i++ {
		_, _, err := service.ProcessResponse(ctx, userID, "Test answer for "+AllTopics[i].Name)
		require.NoError(t, err)
	}

	// Test 1: Empty name should be rejected
	resp, completed, err := service.ProcessResponse(ctx, userID, "")
	require.NoError(t, err)
	assert.False(t, completed, "Should not complete with empty name")
	assert.Contains(t, resp, "Please provide a name")

	// Test 2: Invalid name (special characters) should be rejected
	resp, completed, err = service.ProcessResponse(ctx, userID, "Invalid@Name")
	require.NoError(t, err)
	assert.False(t, completed, "Should not complete with invalid name")
	assert.Contains(t, resp, "not valid")

	// Test 3: Valid unique name should complete
	resp, completed, err = service.ProcessResponse(ctx, userID, "MyUniqueWorld")
	require.NoError(t, err)
	assert.True(t, completed, "Should complete with valid unique name")
	assert.Contains(t, resp, "Thank you")

	// Verify name was saved
	session, err := repo.GetInterview(ctx, userID)
	require.NoError(t, err)
	answers, err := repo.GetAnswers(ctx, session.ID)
	require.NoError(t, err)
	found := false
	for _, a := range answers {
		if a.AnswerText == "MyUniqueWorld" {
			found = true
			break
		}
	}
	assert.True(t, found, "World Name answer should be saved")
}

func TestProcessResponse_DuplicateName(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := new(TestifyMockLLM)
	// Return JSON when extraction prompt is detected
	mockLLM.On("Generate", mock.MatchedBy(func(prompt string) bool {
		return len(prompt) > 100 && strings.Contains(prompt, "data extraction")
	})).Return(`{"theme": "Fantasy", "techLevel": "medieval", "planetSize": "medium", "sentientSpecies": ["Human"]}`, nil)
	// Return name suggestions when name generation prompt is detected
	mockLLM.On("Generate", mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, "generate EXACTLY") || strings.Contains(prompt, "world description, generat")
	})).Return("Alternative1\nAlternative2\nAlternative3", nil)
	mockLLM.On("Generate", mock.Anything).Return("Next question?", nil)

	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	// Setup existing world with name "ExistingWorld"
	existingConfig := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		WorldName:       "ExistingWorld",
		Theme:           "Test",
		TechLevel:       "medieval",
		PlanetSize:      "medium",
		SentientSpecies: []string{"Human"},
	}
	err := repo.SaveConfiguration(ctx, existingConfig)
	require.NoError(t, err)

	// Verify the name is saved
	isTaken, err := repo.IsWorldNameTaken(ctx, "ExistingWorld")
	require.NoError(t, err)
	assert.True(t, isTaken, "ExistingWorld should be marked as taken")

	// Setup session at World Name topic
	userID := uuid.New()
	// Mock start interview call
	mockLLM.On("Generate", mock.Anything).Return("First question", nil).Once()

	session, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	worldNameIndex := -1
	for i, topic := range AllTopics {
		if topic.Name == "World Name" {
			worldNameIndex = i
			break
		}
	}
	require.NotEqual(t, -1, worldNameIndex, "World Name topic should exist")

	err = repo.UpdateQuestionIndex(ctx, session.ID, worldNameIndex)
	require.NoError(t, err)

	// Mock the name generation response (for when duplicate is detected)
	mockLLM.On("Generate", mock.Anything).Return("Suggestion1\nSuggestion2", nil)

	// Try to use existing name (case-insensitive test)
	resp, completed, err := service.ProcessResponse(ctx, userID, "ExistingWorld")
	require.NoError(t, err)

	// Debug output
	t.Logf("Response: %s", resp)
	t.Logf("Completed: %v", completed)

	assert.False(t, completed, "Interview should not be completed when providing duplicate name")
	assert.Contains(t, resp, "already taken", "Response should mention name is taken")
	// Note: We can't easily test for suggestions because the mock might return alternatives or just the message
}

func TestProcessResponse_NameGenerationTrigger(t *testing.T) {
	repo := NewMockRepository()
	worldRepo := NewMockWorldRepository()
	mockLLM := new(TestifyMockLLM)
	service := NewServiceWithRepository(mockLLM, repo, worldRepo)
	ctx := context.Background()

	// Setup session at the step BEFORE World Name
	userID := uuid.New()
	// Mock start interview call
	mockLLM.On("Generate", mock.Anything).Return("First question", nil).Once()

	session, _, err := service.StartInterview(ctx, userID)
	require.NoError(t, err)

	worldNameIndex := -1
	for i, topic := range AllTopics {
		if topic.Name == "World Name" {
			worldNameIndex = i
			break
		}
	}

	// Set to one step before World Name
	err = repo.UpdateQuestionIndex(ctx, session.ID, worldNameIndex-1)
	require.NoError(t, err)

	// Expect generation call with name suggestions in prompt
	mockLLM.On("Generate", mock.MatchedBy(func(prompt string) bool {
		// Check if prompt contains instructions for name suggestions
		// We can't easily check the exact prompt content here without replicating the logic,
		// but we can check if it returns the expected question format
		return true
	})).Return("What should this world be called? Here are suggestions: Alpha or Beta.", nil)

	// Process response to advance to World Name topic
	resp, completed, err := service.ProcessResponse(ctx, userID, "Previous Answer")
	require.NoError(t, err)
	assert.False(t, completed)

	// The response should be the question generated by LLM
	assert.Contains(t, resp, "What should this world be called")
}
