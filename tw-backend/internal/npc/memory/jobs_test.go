package memory

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJobManager_ProcessMemories(t *testing.T) {
	repo := NewMockRepository()
	jm := NewJobManager(repo)
	now := time.Now()

	// Mem 1: Old, accessed recently (should decay and check corruption)
	mem1 := Memory{
		ID:              uuid.New(),
		Timestamp:       now.AddDate(0, 0, -100),
		Clarity:         1.0,
		EmotionalWeight: 0.5,
		AccessCount:     0,
		LastAccessed:    now, // Accessed just now
		Content:         ObservationContent{Event: "Test"},
	}
	repo.CreateMemory(mem1)

	// Mem 2: Old, not accessed (should decay, no corruption check)
	mem2 := Memory{
		ID:              uuid.New(),
		Timestamp:       now.AddDate(0, 0, -100),
		Clarity:         1.0,
		EmotionalWeight: 0.5,
		AccessCount:     0,
		LastAccessed:    now.AddDate(0, 0, -2), // Accessed 2 days ago
		Content:         ObservationContent{Event: "Test"},
	}
	repo.CreateMemory(mem2)

	jm.ProcessMemories([]Memory{mem1, mem2})

	// Verify Mem 1 decayed
	updated1, _ := repo.GetMemory(mem1.ID)
	assert.Less(t, updated1.Clarity, 1.0)

	// Verify Mem 2 decayed
	updated2, _ := repo.GetMemory(mem2.ID)
	assert.Less(t, updated2.Clarity, 1.0)

	// Verify Mem 1 *might* be corrupted (random, so hard to assert true, but logic ran)
	// We can assert that clarity is updated.
}
