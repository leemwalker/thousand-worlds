package emotion

import (
	"testing"
	"time"

	"mud-platform-backend/internal/npc/memory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConsolidateMemories(t *testing.T) {
	repo := memory.NewMockRepository()
	consolidator := NewConsolidator(repo)
	npcID := uuid.New()

	// High emotion, recent -> Should boost
	mem1 := memory.Memory{
		ID:              uuid.New(),
		NPCID:           npcID,
		Timestamp:       time.Now().Add(-1 * time.Hour),
		EmotionalWeight: 0.8,
		Clarity:         0.5,
	}
	repo.CreateMemory(mem1)

	// Low emotion, recent -> No boost
	mem2 := memory.Memory{
		ID:              uuid.New(),
		NPCID:           npcID,
		Timestamp:       time.Now().Add(-2 * time.Hour),
		EmotionalWeight: 0.3,
		Clarity:         0.5,
	}
	repo.CreateMemory(mem2)

	// High emotion, old -> No boost
	mem3 := memory.Memory{
		ID:              uuid.New(),
		NPCID:           npcID,
		Timestamp:       time.Now().Add(-25 * time.Hour),
		EmotionalWeight: 0.9,
		Clarity:         0.5,
	}
	repo.CreateMemory(mem3)

	err := consolidator.ConsolidateMemories(npcID)
	assert.NoError(t, err)

	// Check Mem1
	updated1, _ := repo.GetMemory(mem1.ID)
	// Boost = 0.1 * 0.8 = 0.08. New Clarity = 0.58
	assert.InDelta(t, 0.58, updated1.Clarity, 0.001)

	// Check Mem2
	updated2, _ := repo.GetMemory(mem2.ID)
	assert.Equal(t, 0.5, updated2.Clarity)

	// Check Mem3
	updated3, _ := repo.GetMemory(mem3.ID)
	assert.Equal(t, 0.5, updated3.Clarity)
}
