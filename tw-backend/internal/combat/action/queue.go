package action

import (
	"container/heap"
	"sync"
)

// ActionHeap implements heap.Interface for CombatAction
type ActionHeap []*CombatAction

func (h ActionHeap) Len() int           { return len(h) }
func (h ActionHeap) Less(i, j int) bool { return h[i].ExecuteAt.Before(h[j].ExecuteAt) }
func (h ActionHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *ActionHeap) Push(x interface{}) {
	action, _ := x.(*CombatAction) // Type assertion guaranteed by heap.Push caller
	*h = append(*h, action)
}

func (h *ActionHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// CombatQueue manages the priority queue of actions
type CombatQueue struct {
	actions ActionHeap
	mu      sync.RWMutex
}

// NewCombatQueue creates a new queue
func NewCombatQueue() *CombatQueue {
	q := &CombatQueue{
		actions: make(ActionHeap, 0),
	}
	heap.Init(&q.actions)
	return q
}

// Enqueue adds an action to the queue
func (q *CombatQueue) Enqueue(action *CombatAction) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(&q.actions, action)
}

// Dequeue removes and returns the next action
func (q *CombatQueue) Dequeue() *CombatAction {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.actions) == 0 {
		return nil
	}
	action, _ := heap.Pop(&q.actions).(*CombatAction) // Type assertion guaranteed by heap implementation
	return action
}

// Peek returns the next action without removing it
func (q *CombatQueue) Peek() *CombatAction {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.actions) == 0 {
		return nil
	}
	return q.actions[0]
}

// Len returns the number of actions in the queue
func (q *CombatQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.actions)
}
