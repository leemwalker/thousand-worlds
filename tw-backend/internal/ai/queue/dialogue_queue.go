package queue

import (
	"errors"
	"mud-platform-backend/internal/ai/dialogue"
)

// Priority levels
type Priority int

const (
	PriorityCritical Priority = 1
	PriorityHigh     Priority = 2
	PriorityNormal   Priority = 3
	PriorityLow      Priority = 4
)

// DialogueQueue manages requests by priority
type DialogueQueue struct {
	critical chan *dialogue.DialogueRequest
	high     chan *dialogue.DialogueRequest
	normal   chan *dialogue.DialogueRequest
	low      chan *dialogue.DialogueRequest
	// We'll use a semaphore in the worker, not here directly, or here to limit enqueue?
	// The plan says "semaphore chan struct{} // Limit concurrent requests"
	// But usually semaphore limits *processing*, not queueing.
	// Queue size limits queueing.
}

// NewDialogueQueue creates a new queue
func NewDialogueQueue(size int) *DialogueQueue {
	return &DialogueQueue{
		critical: make(chan *dialogue.DialogueRequest, size),
		high:     make(chan *dialogue.DialogueRequest, size),
		normal:   make(chan *dialogue.DialogueRequest, size),
		low:      make(chan *dialogue.DialogueRequest, size),
	}
}

// Enqueue adds a request to the appropriate channel
func (q *DialogueQueue) Enqueue(req *dialogue.DialogueRequest, p Priority) error {
	select {
	case q.getChannel(p) <- req:
		return nil
	default:
		return errors.New("queue full")
	}
}

func (q *DialogueQueue) getChannel(p Priority) chan *dialogue.DialogueRequest {
	switch p {
	case PriorityCritical:
		return q.critical
	case PriorityHigh:
		return q.high
	case PriorityNormal:
		return q.normal
	case PriorityLow:
		return q.low
	default:
		return q.low
	}
}
