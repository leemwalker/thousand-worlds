package interview

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLLMClient is a mock implementation of LLMClient
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) Generate(prompt string) (string, error) {
	args := m.Called(prompt)
	return args.String(0), args.Error(1)
}

func TestGenerateWorldNames(t *testing.T) {
	mockClient := new(MockLLMClient)
	generator := NewNameGenerator(mockClient)

	session := &InterviewSession{
		State: InterviewState{
			Answers: map[string]string{
				"Core Concept":     "High fantasy world",
				"Sentient Species": "Elves, Humans",
				"Environment":      "Forests",
				"Magic & Tech":     "High magic",
				"Conflict":         "War against dark lord",
			},
		},
	}

	// Mock response with names
	mockResponse := `
1. Aethoria
2. Sylvaris
`
	mockClient.On("Generate", mock.Anything).Return(mockResponse, nil)

	names, err := generator.GenerateWorldNames(session, 2)
	assert.NoError(t, err)
	assert.Len(t, names, 2)
	assert.Equal(t, "Aethoria", names[0])
	assert.Equal(t, "Sylvaris", names[1])
}

func TestGenerateWorldNames_Deduplication(t *testing.T) {
	mockClient := new(MockLLMClient)
	generator := NewNameGenerator(mockClient)

	session := &InterviewSession{
		State: InterviewState{
			Answers: map[string]string{},
		},
	}

	// Mock response with duplicate names (case insensitive)
	mockResponse := `
- Aethoria
- aethoria
- Sylvaris
`
	mockClient.On("Generate", mock.Anything).Return(mockResponse, nil)

	names, err := generator.GenerateWorldNames(session, 2)
	assert.NoError(t, err)
	// We expect 2 unique names, but the mock returns "Aethoria", "aethoria", "Sylvaris".
	// "aethoria" is a duplicate of "Aethoria".
	// So we get "Aethoria" and "Sylvaris".
	assert.Len(t, names, 2)
	assert.Equal(t, "Aethoria", names[0])
	assert.Equal(t, "Sylvaris", names[1])
}

func TestGenerateWorldNames_Cleaning(t *testing.T) {
	mockClient := new(MockLLMClient)
	generator := NewNameGenerator(mockClient)

	session := &InterviewSession{
		State: InterviewState{
			Answers: map[string]string{},
		},
	}

	// Mock response with various formatting
	mockResponse := `
1. NameOne
2) NameTwo
- NameThree
* NameFour
`
	mockClient.On("Generate", mock.Anything).Return(mockResponse, nil)

	names, err := generator.GenerateWorldNames(session, 4)
	assert.NoError(t, err)
	assert.Len(t, names, 4)
	assert.Equal(t, "NameOne", names[0])
	assert.Equal(t, "NameTwo", names[1])
	assert.Equal(t, "NameThree", names[2])
	assert.Equal(t, "NameFour", names[3])
}
