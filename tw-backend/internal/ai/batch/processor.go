package batch

import (
	"context"
	"sync"

	"mud-platform-backend/internal/ai/dialogue"

	"github.com/google/uuid"
)

// DialogueGenerator defines the interface for generating dialogue
type DialogueGenerator interface {
	GenerateDialogue(ctx context.Context, npcID, speakerID uuid.UUID, input string) (*dialogue.DialogueResponse, error)
}

// BatchProcessor handles multiple dialogue requests in parallel
type BatchProcessor struct {
	generator DialogueGenerator
}

// NewBatchProcessor creates a new processor
func NewBatchProcessor(gen DialogueGenerator) *BatchProcessor {
	return &BatchProcessor{
		generator: gen,
	}
}

// BatchRequest represents a collection of dialogue requests
type BatchRequest struct {
	Requests []*dialogue.DialogueRequest
}

// BatchResponse represents the aggregated results
type BatchResponse struct {
	Responses []*dialogue.DialogueResponse
	Errors    map[int]error
}

// ProcessBatch executes requests in parallel
func (p *BatchProcessor) ProcessBatch(ctx context.Context, req BatchRequest) BatchResponse {
	count := len(req.Requests)
	responses := make([]*dialogue.DialogueResponse, count)
	errors := make(map[int]error)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, r := range req.Requests {
		wg.Add(1)
		go func(index int, request *dialogue.DialogueRequest) {
			defer wg.Done()

			// Call DialogueService
			resp, err := p.generator.GenerateDialogue(ctx, request.NPCID, request.SpeakerID, request.Input)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors[index] = err
			} else {
				responses[index] = resp
			}
		}(i, r)
	}

	wg.Wait()

	return BatchResponse{
		Responses: responses,
		Errors:    errors,
	}
}
