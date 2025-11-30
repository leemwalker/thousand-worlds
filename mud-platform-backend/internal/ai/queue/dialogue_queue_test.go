package queue

import (
	"mud-platform-backend/internal/ai/dialogue"
	"testing"
)

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
