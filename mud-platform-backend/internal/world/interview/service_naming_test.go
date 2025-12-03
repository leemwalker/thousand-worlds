package interview

import (
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
	mockLLM := new(TestifyMockLLM)
	mockLLM.On("Generate", mock.Anything).Return("Next question?", nil)

	service := NewServiceWithRepository(mockLLM, repo)

	// Setup session at the step before World Name
	userID := uuid.New()
	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	// Fast forward to World Name topic
	// Find index of World Name topic
	worldNameIndex := -1
	for i, topic := range AllTopics {
		if topic.Name == "World Name" {
			worldNameIndex = i
			break
		}
	}
	require.NotEqual(t, -1, worldNameIndex, "World Name topic not found")

	// Set session state to be at World Name topic
	session.State.CurrentTopicIndex = worldNameIndex
	session.State.CurrentCategory = CategoryTheme
	repo.UpdateInterview(session)

	// Test 1: Empty name
	resp, completed, err := service.ProcessResponse(session.ID, "")
	require.NoError(t, err)
	assert.False(t, completed)
	assert.Contains(t, resp, "Please provide a name")

	// Test 2: Invalid name (special characters)
	resp, completed, err = service.ProcessResponse(session.ID, "Invalid@Name")
	require.NoError(t, err)
	assert.False(t, completed)
	assert.Contains(t, resp, "not valid")

	// Test 3: Valid unique name
	resp, completed, err = service.ProcessResponse(session.ID, "MyUniqueWorld")
	require.NoError(t, err)

	// Should advance or complete (depending on if it's the last topic)
	// Since World Name is the last topic, it should complete
	assert.True(t, completed)
	assert.Contains(t, resp, "Thank you")

	// Verify name was saved in session answers
	updatedSession, _ := repo.GetInterview(session.ID)
	assert.Equal(t, "MyUniqueWorld", updatedSession.State.Answers["World Name"])
}

func TestProcessResponse_DuplicateName(t *testing.T) {
	repo := NewMockRepository()
	mockLLM := new(TestifyMockLLM)

	service := NewServiceWithRepository(mockLLM, repo)

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
	err := repo.SaveConfiguration(existingConfig)
	require.NoError(t, err)

	// Verify the name is saved
	isTaken, err := repo.IsWorldNameTaken("ExistingWorld")
	require.NoError(t, err)
	assert.True(t, isTaken, "ExistingWorld should be marked as taken")

	// Setup session at World Name topic
	userID := uuid.New()
	// Mock start interview call
	mockLLM.On("Generate", mock.Anything).Return("First question", nil).Once()

	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	worldNameIndex := -1
	for i, topic := range AllTopics {
		if topic.Name == "World Name" {
			worldNameIndex = i
			break
		}
	}
	require.NotEqual(t, -1, worldNameIndex, "World Name topic should exist")

	session.State.CurrentTopicIndex = worldNameIndex
	err = repo.UpdateInterview(session)
	require.NoError(t, err)

	// Mock the name generation response (for when duplicate is detected)
	mockLLM.On("Generate", mock.Anything).Return("Suggestion1\nSuggestion2", nil)

	// Try to use existing name (case-insensitive test)
	resp, completed, err := service.ProcessResponse(session.ID, "ExistingWorld")
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
	mockLLM := new(TestifyMockLLM)
	service := NewServiceWithRepository(mockLLM, repo)

	// Setup session at the step BEFORE World Name
	userID := uuid.New()
	// Mock start interview call
	mockLLM.On("Generate", mock.Anything).Return("First question", nil).Once()

	session, _, err := service.StartInterview(userID)
	require.NoError(t, err)

	worldNameIndex := -1
	for i, topic := range AllTopics {
		if topic.Name == "World Name" {
			worldNameIndex = i
			break
		}
	}

	// Set to one step before World Name
	session.State.CurrentTopicIndex = worldNameIndex - 1
	repo.UpdateInterview(session)

	// Expect generation call with name suggestions in prompt
	mockLLM.On("Generate", mock.MatchedBy(func(prompt string) bool {
		// Check if prompt contains instructions for name suggestions
		// We can't easily check the exact prompt content here without replicating the logic,
		// but we can check if it returns the expected question format
		return true
	})).Return("What should this world be called? Here are suggestions: Alpha or Beta.", nil)

	// Process response to advance to World Name topic
	resp, completed, err := service.ProcessResponse(session.ID, "Previous Answer")
	require.NoError(t, err)
	assert.False(t, completed)

	// The response should be the question generated by LLM
	assert.Contains(t, resp, "What should this world be called")
}
