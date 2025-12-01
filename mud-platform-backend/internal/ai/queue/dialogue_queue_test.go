package queue

import (
	"context"
	"mud-platform-backend/internal/ai/dialogue"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// MockDialogueGenerator for testing
type MockDialogueGenerator struct {
	mu           sync.Mutex
	GenerateFunc func(ctx context.Context, npcID, speakerID uuid.UUID, input string) (*dialogue.DialogueResponse, error)
	CallCount    int
}

func (m *MockDialogueGenerator) GenerateDialogue(ctx context.Context, npcID, speakerID uuid.UUID, input string) (*dialogue.DialogueResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CallCount++
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, npcID, speakerID, input)
	}
	return &dialogue.DialogueResponse{Text: "Mock response"}, nil
}

func TestDialogueQueue_Enqueue(t *testing.T) {
	q := NewDialogueQueue(2) // Small size for testing

	req1 := &dialogue.DialogueRequest{Input: "1"}
	req2 := &dialogue.DialogueRequest{Input: "2"}
	req3 := &dialogue.DialogueRequest{Input: "3"}

	// 1. Enqueue Critical
	if err := q.Enqueue(req1, PriorityCritical); err != nil {
		t.Errorf("Failed to enqueue critical: %v", err)
	}
	if err := q.Enqueue(req2, PriorityCritical); err != nil {
		t.Errorf("Failed to enqueue critical 2: %v", err)
	}

	// 2. Overflow Critical
	if err := q.Enqueue(req3, PriorityCritical); err == nil {
		t.Error("Expected overflow error for critical")
	}

	// 3. Enqueue Low
	if err := q.Enqueue(req1, PriorityLow); err != nil {
		t.Errorf("Failed to enqueue low: %v", err)
	}
}

func TestWorkerPool_FetchOrder(t *testing.T) {
	q := NewDialogueQueue(10)
	wp := &WorkerPool{queue: q}

	crit := &dialogue.DialogueRequest{Input: "Critical"}
	high := &dialogue.DialogueRequest{Input: "High"}
	low := &dialogue.DialogueRequest{Input: "Low"}

	q.Enqueue(low, PriorityLow)
	q.Enqueue(high, PriorityHigh)
	q.Enqueue(crit, PriorityCritical)

	// Should fetch Critical first
	req := wp.fetchNextRequest()
	if req.Input != "Critical" {
		t.Errorf("Expected Critical, got %s", req.Input)
	}

	// Should fetch High next
	req = wp.fetchNextRequest()
	if req.Input != "High" {
		t.Errorf("Expected High, got %s", req.Input)
	}

	// Should fetch Low next
	req = wp.fetchNextRequest()
	if req.Input != "Low" {
		t.Errorf("Expected Low, got %s", req.Input)
	}
}

func TestWorkerPool_Processing(t *testing.T) {
	q := NewDialogueQueue(10)
	mockService := &MockDialogueGenerator{}

	// Use wait group to wait for processing in test
	var wg sync.WaitGroup
	mockService.GenerateFunc = func(ctx context.Context, npcID, speakerID uuid.UUID, input string) (*dialogue.DialogueResponse, error) {
		defer wg.Done()
		return &dialogue.DialogueResponse{Text: "Processed " + input}, nil
	}

	wp := NewWorkerPool(q, 2, mockService)
	wp.Start()
	defer wp.Stop()

	// Enqueue requests
	count := 5
	wg.Add(count)
	for i := 0; i < count; i++ {
		q.Enqueue(&dialogue.DialogueRequest{Input: "test"}, PriorityNormal)
	}

	// Wait for processing with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for requests to be processed")
	}

	mockService.mu.Lock()
	if mockService.CallCount != count {
		t.Errorf("Expected %d calls, got %d", count, mockService.CallCount)
	}
	mockService.mu.Unlock()
}
