package ai

import (
	"context"
	"encoding/json"
	"testing"

	"tw-backend/internal/repository"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventBus mocks the NATS interaction
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error) {
	args := m.Called(subject, cb)
	// Return a dummy subscription or nil if error
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*nats.Subscription), args.Error(1)
}

func (m *MockEventBus) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

// MockMemoryRepository mocks the database interaction
type MockMemoryRepository struct {
	mock.Mock
}

func (m *MockMemoryRepository) GetMemoriesByWorldID(ctx context.Context, worldID string) ([]repository.Memory, error) {
	args := m.Called(ctx, worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.Memory), args.Error(1)
}

func TestDesireEngine_ListenForDecisions(t *testing.T) {
	mockBus := new(MockEventBus)
	mockRepo := new(MockMemoryRepository)
	engine := NewDesireEngine(mockBus, mockRepo)

	// Expect subscriptions
	mockBus.On("Subscribe", "npc.command.decide_action", mock.AnythingOfType("nats.MsgHandler")).Return(&nats.Subscription{}, nil)
	mockBus.On("Subscribe", "ai.response.decision.*", mock.AnythingOfType("nats.MsgHandler")).Return(&nats.Subscription{}, nil)

	err := engine.ListenForDecisions()
	assert.NoError(t, err)
	mockBus.AssertExpectations(t)
}

func TestDesireEngine_HandleDecideAction(t *testing.T) {
	mockBus := new(MockEventBus)
	mockRepo := new(MockMemoryRepository)
	engine := NewDesireEngine(mockBus, mockRepo)

	// Setup payload
	cmd := DecideActionCommand{
		EntityID:   "npc-123",
		WorldID:    "world-1",
		WorldState: "You are standing in a tavern.",
	}
	data, _ := json.Marshal(cmd)
	msg := &nats.Msg{Data: data}

	// Mock DB response
	memories := []repository.Memory{
		{NPCID: "npc-123", Content: "I hate loud noises.", ImportanceScore: 0.8},
		{NPCID: "npc-123", Content: "I love ale.", ImportanceScore: 0.9},
		{NPCID: "other", Content: "Irrelevant.", ImportanceScore: 0.9},
	}
	mockRepo.On("GetMemoriesByWorldID", mock.Anything, "world-1").Return(memories, nil)

	// Expect Publish to AI Gateway
	mockBus.On("Publish", "ai.request.decision", mock.MatchedBy(func(data []byte) bool {
		var req AIRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return false
		}
		return req.EntityID == "npc-123" && len(req.Prompt) > 0
	})).Return(nil)

	// Since handleDecideAction is a private method called by the handler,
	// we normally rely on ListenForDecisions to register it.
	// But for testing logic, we can call the method directly if we make it public (not ideal)
	// or capture the handler in Subscribe.
	// Given we can't easily capture the handler without refactoring test setup,
	// let's just make the handler logic testable?
	// Actually, Go allows calling private methods in the same package (Test file is package ai).
	engine.handleDecideAction(msg)

	mockRepo.AssertExpectations(t)
	mockBus.AssertExpectations(t)
}

func TestDesireEngine_HandleAIResponse(t *testing.T) {
	mockBus := new(MockEventBus)
	mockRepo := new(MockMemoryRepository)
	engine := NewDesireEngine(mockBus, mockRepo)

	resp := AIResponse{
		RequestID: "req-1",
		Response:  "DRINK ALE",
		EntityID:  "npc-123",
	}
	data, _ := json.Marshal(resp)
	msg := &nats.Msg{Data: data}

	// Expect Publish to Spatial Service
	mockBus.On("Publish", "spatial.command.action", []byte("DRINK ALE")).Return(nil)

	engine.handleAIResponse(msg)

	mockBus.AssertExpectations(t)
}
