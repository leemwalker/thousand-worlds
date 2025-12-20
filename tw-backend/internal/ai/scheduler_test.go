package ai

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScheduler(t *testing.T) {
	s := NewScheduler(4)
	require.NotNil(t, s)
	assert.Equal(t, 4, s.totalBuckets)
}

func TestNewScheduler_DefaultBuckets(t *testing.T) {
	s := NewScheduler(0)
	assert.Equal(t, 4, s.totalBuckets, "Should default to 4 buckets")

	s2 := NewScheduler(-1)
	assert.Equal(t, 4, s2.totalBuckets, "Should default to 4 buckets for negative")
}

func TestScheduler_RegisterUnregister(t *testing.T) {
	s := NewScheduler(4)
	id1 := uuid.New()
	id2 := uuid.New()

	s.RegisterEntity(id1)
	s.RegisterEntity(id2)

	total, buckets, _ := s.GetStats()
	assert.Equal(t, 2, total)
	assert.Equal(t, 4, buckets)

	s.UnregisterEntity(id1)
	total, _, _ = s.GetStats()
	assert.Equal(t, 1, total)

	s.UnregisterEntity(id2)
	total, _, _ = s.GetStats()
	assert.Equal(t, 0, total)
}

func TestScheduler_GetEntitiesForTick(t *testing.T) {
	s := NewScheduler(4)

	// Register 8 entities (2 per bucket)
	ids := make([]uuid.UUID, 8)
	for i := 0; i < 8; i++ {
		ids[i] = uuid.New()
		s.RegisterEntity(ids[i])
	}

	// Tick 0 should get entities at indices 0, 4 (bucket 0)
	tick0 := s.GetEntitiesForTick(0)
	assert.Len(t, tick0, 2, "Tick 0 should process 2 entities")

	// Tick 1 should get entities at indices 1, 5 (bucket 1)
	tick1 := s.GetEntitiesForTick(1)
	assert.Len(t, tick1, 2, "Tick 1 should process 2 entities")

	// Verify all entities processed over 4 ticks
	allProcessed := make(map[uuid.UUID]bool)
	for tick := int64(0); tick < 4; tick++ {
		for _, id := range s.GetEntitiesForTick(tick) {
			allProcessed[id] = true
		}
	}
	assert.Len(t, allProcessed, 8, "All 8 entities should be processed over 4 ticks")
}

func TestScheduler_ShouldProcessEntity(t *testing.T) {
	s := NewScheduler(4)
	id := uuid.New()
	s.RegisterEntity(id)

	// First entity (index 0) should process on ticks 0, 4, 8...
	assert.True(t, s.ShouldProcessEntity(0, id))
	assert.False(t, s.ShouldProcessEntity(1, id))
	assert.False(t, s.ShouldProcessEntity(2, id))
	assert.False(t, s.ShouldProcessEntity(3, id))
	assert.True(t, s.ShouldProcessEntity(4, id))
}

func TestScheduler_Distribution(t *testing.T) {
	s := NewScheduler(4)

	// Register 100 entities
	for i := 0; i < 100; i++ {
		s.RegisterEntity(uuid.New())
	}

	_, _, entitiesPerBucket := s.GetStats()

	// Each bucket should have ~25 entities
	for i, count := range entitiesPerBucket {
		assert.GreaterOrEqual(t, count, 24, "Bucket %d should have at least 24 entities", i)
		assert.LessOrEqual(t, count, 26, "Bucket %d should have at most 26 entities", i)
	}
}

func TestScheduler_RebuildFromEntities(t *testing.T) {
	s := NewScheduler(4)

	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	s.RebuildFromEntities(ids)

	total, _, _ := s.GetStats()
	assert.Equal(t, 3, total)

	// Verify deterministic ordering (sorted by UUID string)
	s2 := NewScheduler(4)
	s2.RebuildFromEntities(ids)

	// Should produce same results
	for tick := int64(0); tick < 4; tick++ {
		e1 := s.GetEntitiesForTick(tick)
		e2 := s2.GetEntitiesForTick(tick)
		assert.Equal(t, e1, e2, "Same tick should produce same entities")
	}
}

func TestScheduler_UnregisterNonExistent(t *testing.T) {
	s := NewScheduler(4)
	// Should not panic
	s.UnregisterEntity(uuid.New())
}

func TestScheduler_DuplicateRegister(t *testing.T) {
	s := NewScheduler(4)
	id := uuid.New()

	s.RegisterEntity(id)
	s.RegisterEntity(id) // Duplicate

	total, _, _ := s.GetStats()
	assert.Equal(t, 1, total, "Duplicate registration should be ignored")
}
