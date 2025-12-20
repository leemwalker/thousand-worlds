// Package ai provides AI scheduling and behavior systems.
package ai

import (
	"sort"
	"sync"

	"github.com/google/uuid"
)

// Scheduler distributes AI processing across ticks to prevent CPU spikes.
// Uses pre-computed bucket slices for O(N/buckets) GetEntitiesForTick access.
type Scheduler struct {
	mu           sync.RWMutex
	totalBuckets int
	currentTick  int64

	// Pre-computed bucket slices for O(N/buckets) access
	buckets   [][]uuid.UUID     // bucket index -> entity IDs in that bucket
	entitySet map[uuid.UUID]int // ID -> bucket index
}

// NewScheduler creates a new AI scheduler with the specified number of buckets.
// More buckets = more ticks to process all entities = smoother CPU usage.
// Recommended: 4-10 buckets depending on entity count and tick rate.
func NewScheduler(buckets int) *Scheduler {
	if buckets < 1 {
		buckets = 4 // Default to 4 buckets
	}
	s := &Scheduler{
		totalBuckets: buckets,
		buckets:      make([][]uuid.UUID, buckets),
		entitySet:    make(map[uuid.UUID]int),
	}
	for i := range s.buckets {
		s.buckets[i] = make([]uuid.UUID, 0)
	}
	return s
}

// RegisterEntity adds an entity to the scheduler.
// Entities are assigned to buckets using round-robin based on total count.
func (s *Scheduler) RegisterEntity(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entitySet[id]; exists {
		return // Already registered
	}

	// Assign to bucket with fewest entities for even distribution
	minBucket := 0
	minCount := len(s.buckets[0])
	for i := 1; i < s.totalBuckets; i++ {
		if len(s.buckets[i]) < minCount {
			minCount = len(s.buckets[i])
			minBucket = i
		}
	}

	s.buckets[minBucket] = append(s.buckets[minBucket], id)
	s.entitySet[id] = minBucket
}

// UnregisterEntity removes an entity from the scheduler.
func (s *Scheduler) UnregisterEntity(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucketIdx, exists := s.entitySet[id]
	if !exists {
		return
	}

	// Find and remove from bucket slice
	bucket := s.buckets[bucketIdx]
	for i, eid := range bucket {
		if eid == id {
			// Swap with last, then truncate
			bucket[i] = bucket[len(bucket)-1]
			s.buckets[bucketIdx] = bucket[:len(bucket)-1]
			break
		}
	}
	delete(s.entitySet, id)
}

// GetEntitiesForTick returns the entity IDs that should be processed this tick.
// O(N/buckets) - returns pre-computed bucket slice directly.
func (s *Scheduler) GetEntitiesForTick(tick int64) []uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.currentTick = tick
	bucketIdx := int(tick % int64(s.totalBuckets))

	// Return a copy to prevent external mutation
	result := make([]uuid.UUID, len(s.buckets[bucketIdx]))
	copy(result, s.buckets[bucketIdx])
	return result
}

// ShouldProcessEntity checks if a specific entity should be processed this tick.
// O(1) lookup using entitySet.
func (s *Scheduler) ShouldProcessEntity(tick int64, entityID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucketIdx, exists := s.entitySet[entityID]
	if !exists {
		return false
	}

	currentBucket := int(tick % int64(s.totalBuckets))
	return bucketIdx == currentBucket
}

// SetBuckets updates the number of buckets.
// Note: This redistributes all entities.
func (s *Scheduler) SetBuckets(buckets int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if buckets < 1 {
		buckets = 1
	}

	// Collect all entities
	var allEntities []uuid.UUID
	for _, bucket := range s.buckets {
		allEntities = append(allEntities, bucket...)
	}

	// Rebuild with new bucket count
	s.totalBuckets = buckets
	s.buckets = make([][]uuid.UUID, buckets)
	for i := range s.buckets {
		s.buckets[i] = make([]uuid.UUID, 0)
	}
	s.entitySet = make(map[uuid.UUID]int)

	// Re-register all entities (round-robin)
	for i, id := range allEntities {
		bucketIdx := i % buckets
		s.buckets[bucketIdx] = append(s.buckets[bucketIdx], id)
		s.entitySet[id] = bucketIdx
	}
}

// GetStats returns scheduler statistics.
func (s *Scheduler) GetStats() (totalEntities, buckets int, entitiesPerBucket []int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buckets = s.totalBuckets
	entitiesPerBucket = make([]int, s.totalBuckets)

	for i, bucket := range s.buckets {
		entitiesPerBucket[i] = len(bucket)
		totalEntities += len(bucket)
	}
	return
}

// RebuildFromEntities rebuilds the scheduler from a slice of entity IDs.
// Sorts by UUID for consistent ordering across restarts, then distributes round-robin.
func (s *Scheduler) RebuildFromEntities(entityIDs []uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Sort for consistent bucketing across restarts
	sorted := make([]uuid.UUID, len(entityIDs))
	copy(sorted, entityIDs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].String() < sorted[j].String()
	})

	// Clear and rebuild buckets
	s.buckets = make([][]uuid.UUID, s.totalBuckets)
	for i := range s.buckets {
		s.buckets[i] = make([]uuid.UUID, 0, len(sorted)/s.totalBuckets+1)
	}
	s.entitySet = make(map[uuid.UUID]int, len(sorted))

	// Distribute round-robin
	for i, id := range sorted {
		bucketIdx := i % s.totalBuckets
		s.buckets[bucketIdx] = append(s.buckets[bucketIdx], id)
		s.entitySet[id] = bucketIdx
	}
}
