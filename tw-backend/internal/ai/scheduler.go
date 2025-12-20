// Package ai provides AI scheduling and behavior systems.
package ai

import (
	"sort"
	"sync"

	"github.com/google/uuid"
)

// Scheduler distributes AI processing across ticks to prevent CPU spikes.
// Uses modulo-based bucketing: entities are processed when (entityIndex % buckets) == (tick % buckets)
type Scheduler struct {
	mu           sync.RWMutex
	totalBuckets int
	currentTick  int64

	// Sorted entity IDs for consistent bucketing
	entityOrder []uuid.UUID
	entitySet   map[uuid.UUID]int // ID -> index in entityOrder
}

// NewScheduler creates a new AI scheduler with the specified number of buckets.
// More buckets = more ticks to process all entities = smoother CPU usage.
// Recommended: 4-10 buckets depending on entity count and tick rate.
func NewScheduler(buckets int) *Scheduler {
	if buckets < 1 {
		buckets = 4 // Default to 4 buckets
	}
	return &Scheduler{
		totalBuckets: buckets,
		entityOrder:  make([]uuid.UUID, 0),
		entitySet:    make(map[uuid.UUID]int),
	}
}

// RegisterEntity adds an entity to the scheduler.
// Entities are assigned to buckets based on their insertion order.
func (s *Scheduler) RegisterEntity(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entitySet[id]; exists {
		return // Already registered
	}

	s.entityOrder = append(s.entityOrder, id)
	s.entitySet[id] = len(s.entityOrder) - 1
}

// UnregisterEntity removes an entity from the scheduler.
func (s *Scheduler) UnregisterEntity(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, exists := s.entitySet[id]
	if !exists {
		return
	}

	// Remove from order slice (swap with last, then truncate)
	lastIdx := len(s.entityOrder) - 1
	if idx != lastIdx {
		s.entityOrder[idx] = s.entityOrder[lastIdx]
		s.entitySet[s.entityOrder[idx]] = idx
	}
	s.entityOrder = s.entityOrder[:lastIdx]
	delete(s.entitySet, id)
}

// GetEntitiesForTick returns the entity IDs that should be processed this tick.
// Uses modulo-based distribution for even spread across buckets.
func (s *Scheduler) GetEntitiesForTick(tick int64) []uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.currentTick = tick
	bucket := int(tick % int64(s.totalBuckets))

	var result []uuid.UUID
	for idx, id := range s.entityOrder {
		if idx%s.totalBuckets == bucket {
			result = append(result, id)
		}
	}
	return result
}

// ShouldProcessEntity checks if a specific entity should be processed this tick.
// More efficient than GetEntitiesForTick when checking individual entities.
func (s *Scheduler) ShouldProcessEntity(tick int64, entityID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, exists := s.entitySet[entityID]
	if !exists {
		return false
	}

	bucket := int(tick % int64(s.totalBuckets))
	return idx%s.totalBuckets == bucket
}

// SetBuckets updates the number of buckets.
// Note: This redistributes all entities.
func (s *Scheduler) SetBuckets(buckets int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if buckets < 1 {
		buckets = 1
	}
	s.totalBuckets = buckets
}

// GetStats returns scheduler statistics.
func (s *Scheduler) GetStats() (totalEntities, buckets int, entitiesPerBucket []int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalEntities = len(s.entityOrder)
	buckets = s.totalBuckets
	entitiesPerBucket = make([]int, s.totalBuckets)

	for idx := range s.entityOrder {
		bucket := idx % s.totalBuckets
		entitiesPerBucket[bucket]++
	}
	return
}

// RebuildFromEntities rebuilds the scheduler from a map of entities.
// Sorts by UUID for consistent ordering across restarts.
func (s *Scheduler) RebuildFromEntities(entityIDs []uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Sort for consistent bucketing
	sorted := make([]uuid.UUID, len(entityIDs))
	copy(sorted, entityIDs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].String() < sorted[j].String()
	})

	s.entityOrder = sorted
	s.entitySet = make(map[uuid.UUID]int, len(sorted))
	for idx, id := range sorted {
		s.entitySet[id] = idx
	}
}
