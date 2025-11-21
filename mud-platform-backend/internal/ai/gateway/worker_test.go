package gateway

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

// MockAIClient is a mock implementation of AIClient
type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) Generate(prompt string, model string) (string, error) {
	args := m.Called(prompt, model)
	return args.String(0), args.Error(1)
}

// MockPublisher is a mock implementation of Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(subj string, data []byte) error {
	args := m.Called(subj, data)
	return args.Error(0)
}

func TestProcessRequest_Success(t *testing.T) {
	mockClient := new(MockAIClient)
	mockPublisher := new(MockPublisher)

	req := AIRequest{
		ID:              "123",
		Prompt:          "test prompt",
		Model:           "test-model",
		ResponseSubject: "response.subject",
	}

	mockClient.On("Generate", "test prompt", "test-model").Return("test response", nil)

	mockPublisher.On("Publish", "response.subject", mock.MatchedBy(func(data []byte) bool {
		var resp AIResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return false
		}
		return resp.ID == "123" && resp.Response == "test response" && resp.Error == ""
	})).Return(nil)

	processRequest(mockPublisher, mockClient, req)

	mockClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestProcessRequest_AIError(t *testing.T) {
	mockClient := new(MockAIClient)
	mockPublisher := new(MockPublisher)

	req := AIRequest{
		ID:              "123",
		Prompt:          "test prompt",
		Model:           "test-model",
		ResponseSubject: "response.subject",
	}

	mockClient.On("Generate", "test prompt", "test-model").Return("", errors.New("ai error"))

	mockPublisher.On("Publish", "response.subject", mock.MatchedBy(func(data []byte) bool {
		var resp AIResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return false
		}
		return resp.ID == "123" && resp.Error == "ai error"
	})).Return(nil)

	processRequest(mockPublisher, mockClient, req)

	mockClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestProcessRequest_NoResponseSubject(t *testing.T) {
	mockClient := new(MockAIClient)
	mockPublisher := new(MockPublisher)

	req := AIRequest{
		ID:     "123",
		Prompt: "test prompt",
		Model:  "test-model",
		// No ResponseSubject
	}

	mockClient.On("Generate", "test prompt", "test-model").Return("test response", nil)

	// Publish should NOT be called

	processRequest(mockPublisher, mockClient, req)

	mockClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
