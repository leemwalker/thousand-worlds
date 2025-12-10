package batch

import (
	"context"
	"errors"
	"testing"

	"tw-backend/internal/ai/dialogue"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockGenerator struct {
	mock.Mock
}

func (m *MockGenerator) GenerateDialogue(ctx context.Context, npcID, speakerID uuid.UUID, input string) (*dialogue.DialogueResponse, error) {
	args := m.Called(ctx, npcID, speakerID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dialogue.DialogueResponse), args.Error(1)
}

func TestProcessBatch(t *testing.T) {
	mockGen := new(MockGenerator)
	processor := NewBatchProcessor(mockGen)

	npc1 := uuid.New()
	npc2 := uuid.New()
	speaker := uuid.New()

	reqs := []*dialogue.DialogueRequest{
		{NPCID: npc1, SpeakerID: speaker, Input: "Hello 1"},
		{NPCID: npc2, SpeakerID: speaker, Input: "Hello 2"},
	}

	// Setup Expectations
	mockGen.On("GenerateDialogue", mock.Anything, npc1, speaker, "Hello 1").Return(&dialogue.DialogueResponse{Text: "Response 1"}, nil)
	mockGen.On("GenerateDialogue", mock.Anything, npc2, speaker, "Hello 2").Return(nil, errors.New("error 2"))

	// Execute
	batchReq := BatchRequest{Requests: reqs}
	resp := processor.ProcessBatch(context.Background(), batchReq)

	// Verify
	if len(resp.Responses) != 2 {
		t.Errorf("Expected 2 responses slots, got %d", len(resp.Responses))
	}

	if resp.Responses[0].Text != "Response 1" {
		t.Errorf("Expected Response 1, got %v", resp.Responses[0])
	}

	if resp.Responses[1] != nil {
		t.Errorf("Expected nil response for error, got %v", resp.Responses[1])
	}

	if len(resp.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(resp.Errors))
	}

	if resp.Errors[1].Error() != "error 2" {
		t.Errorf("Expected 'error 2', got %v", resp.Errors[1])
	}
}
