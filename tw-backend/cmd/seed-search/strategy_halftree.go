package main

import (
	"container/heap"
	"fmt"
	"os"
)

// HalfTreeStrategy implements an interval-based priority search
// It subdivides the search space into intervals and prioritizes them based on the quality of their midpoint.
type HalfTreeStrategy struct{}

func (s *HalfTreeStrategy) Name() string {
	return "halftree"
}

func (s *HalfTreeStrategy) Explanation() string {
	return "Interval Priority Search: Checks interval midpoints. High scores boost interval priority. Recursively drills down into promising ranges."
}

// Run executes the half-tree search
func (s *HalfTreeStrategy) Run(seeds chan<- int64, results <-chan SeedResult, config SearchConfig) <-chan SeedResult {
	const MaxWorldSeed = 10_000_000_000
	rangeStart := config.StartSeed
	rangeEnd := int64(MaxWorldSeed)
	searchBudget := config.Count

	fmt.Fprintf(os.Stderr, "🌲 Half-Tree Strategy: exploring range [%d, %d] with budget of %d candidates\n", rangeStart, rangeEnd, searchBudget)

	output := make(chan SeedResult)

	// Manager Goroutine
	go func() {
		defer close(output)

		// Priority Queue
		pq := &IntervalPQ{}
		heap.Init(pq)

		// Initial Interval
		heap.Push(pq, &SearchInterval{
			Min:      rangeStart,
			Max:      rangeEnd,
			Priority: 50.0, // Base priority
			Depth:    0,
		})

		// State
		checked := make(map[int64]bool)
		pending := make(map[int64]*SearchInterval)
		completedCount := 0

		// Seed Buffer for select loop
		var nextSeed int64
		var nextInterval *SearchInterval
		var seedsChan chan<- int64 // nil when no seed ready

		// We consider "checked" as "dispatched".

		for {
			// 1. Prepare next seed if needed and possible
			if seedsChan == nil {
				// Check if we are completely done (budget met and no pending results)
				if completedCount >= searchBudget && len(pending) == 0 {
					return
				}

				// Determine if we can feed more seeds based on budget
				canFeed := completedCount+len(pending) < searchBudget

				if canFeed && pq.Len() > 0 {
					item := heap.Pop(pq).(*SearchInterval)

					// Calculate midpoint
					mid := item.Min + (item.Max-item.Min)/2

					// Handle edge cases for intervals
					if item.Max <= item.Min { // Interval is empty or single point
						continue // Skip this interval
					}

					if checked[mid] {
						// If midpoint was already checked, it means we've processed this exact seed before.
						// This can happen if intervals overlap at edges or due to integer division.
						// Instead of re-checking, we immediately split this interval using its current priority.
						// This ensures we still explore the sub-ranges.
						left := &SearchInterval{Min: item.Min, Max: mid, Priority: item.Priority, Depth: item.Depth + 1}
						right := &SearchInterval{Min: mid + 1, Max: item.Max, Priority: item.Priority, Depth: item.Depth + 1}
						if left.Max >= left.Min {
							heap.Push(pq, left)
						}
						if right.Max >= right.Min {
							heap.Push(pq, right)
						}
						continue // Try to pop another interval
					}

					// This midpoint is new and can be dispatched
					nextSeed = mid
					nextInterval = item
					seedsChan = seeds // Enable send case in select
				} else if !canFeed && len(pending) == 0 {
					// Budget exhausted and no pending results.
					return
				} else if !canFeed {
					// Budget exhausted, just wait for pending results.
					// seedsChan remains nil, so the send case is disabled.
				} else if pq.Len() == 0 && len(pending) == 0 {
					// Empty queue and no pending. Search space exhausted.
					return
				}
			}

			// 2. Select Loop
			select {
			case seedsChan <- nextSeed:
				// Successfully dispatched a seed
				checked[nextSeed] = true
				pending[nextSeed] = nextInterval

				// Disable send until we prepare the next seed
				seedsChan = nil
				nextInterval = nil
				nextSeed = 0 // Clear nextSeed

			case res, ok := <-results:
				if !ok {
					// The results channel was closed, meaning workers are done.
					// We should also stop.
					return
				}

				completedCount++

				// Forward result synchronously to ensure order and avoid race on close
				output <- res

				// Process Feedback: Update PQ based on result
				parent, found := pending[res.Seed]
				if found {
					delete(pending, res.Seed) // Remove from pending as it's now processed

					// Priority Logic:
					// Good score means we should prioritize drilling down into this region.
					// Score is typically 0-100.
					priority := float64(res.Score)
					if priority < 10 {
						priority = 10
					} // Ensure a minimum priority to keep exploring

					// Split the parent interval into two children based on the midpoint (res.Seed)
					mid := res.Seed
					left := &SearchInterval{Min: parent.Min, Max: mid, Priority: priority, Depth: parent.Depth + 1}
					right := &SearchInterval{Min: mid + 1, Max: parent.Max, Priority: priority, Depth: parent.Depth + 1}

					// Push valid sub-intervals to the priority queue
					if left.Max >= left.Min {
						heap.Push(pq, left)
					}
					if right.Max >= right.Min {
						heap.Push(pq, right)
					}
				}
			}
		}
	}()

	return output
}

// SearchInterval represents a range of seeds to explore
type SearchInterval struct {
	Min      int64
	Max      int64
	Priority float64 // Higher is better
	Depth    int
	index    int // Loop index for heap
}

// IntervalPQ implements heap.Interface
type IntervalPQ []*SearchInterval

func (pq IntervalPQ) Len() int { return len(pq) }

func (pq IntervalPQ) Less(i, j int) bool {
	// Higher priority comes first (Pop returns MAX priority)
	return pq[i].Priority > pq[j].Priority
}

func (pq IntervalPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *IntervalPQ) Push(x interface{}) {
	n := len(*pq)
	item := x.(*SearchInterval)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *IntervalPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // Avoid memory leak
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
