package memory

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests use MockRepository which implements the same interface
// This allows us to test the interface contract without requiring MongoDB

func TestRepository_CreateAndGetMemory(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	mem := Memory{
		NPCID:           npcID,
		Type:            MemoryTypeObservation,
		Timestamp:       time.Now(),
		Clarity:         1.0,
		EmotionalWeight: 0.5,
		Content: ObservationContent{
			Event: "Test event",
			Location: Location{
				X: 100, Y: 200, Z: 0,
				WorldID: uuid.New(),
			},
		},
		Tags: []string{"tag1", "tag2"},
	}

	// Create
	id, err := repo.CreateMemory(mem)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	// Get
	retrieved, err := repo.GetMemory(id)
	require.NoError(t, err)
	assert.Equal(t, npcID, retrieved.NPCID)
	assert.Equal(t, MemoryTypeObservation, retrieved.Type)
	assert.Equal(t, 1.0, retrieved.Clarity)
}

func TestRepository_GetMemory_NotFound(t *testing.T) {
	repo := NewMockRepository()

	_, err := repo.GetMemory(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRepository_GetMemoriesByNPC(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()
	otherNPCID := uuid.New()

	// Create 3 memories for npcID
	for i := 0; i < 3; i++ {
		mem := Memory{
			NPCID:     npcID,
			Type:      MemoryTypeObservation,
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Clarity:   1.0,
			Content:   ObservationContent{Event: "Event"},
		}
		_, err := repo.CreateMemory(mem)
		require.NoError(t, err)
	}

	// Create 1 memory for different NPC
	mem := Memory{
		NPCID:     otherNPCID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Event"},
	}
	_, err := repo.CreateMemory(mem)
	require.NoError(t, err)

	// Query
	memories, err := repo.GetMemoriesByNPC(npcID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, memories, 3)
}

func TestRepository_GetMemoriesByNPC_Pagination(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	// Create 5 memories
	for i := 0; i < 5; i++ {
		mem := Memory{
			NPCID:     npcID,
			Type:      MemoryTypeObservation,
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Clarity:   1.0,
			Content:   ObservationContent{Event: "Event"},
		}
		_, err := repo.CreateMemory(mem)
		require.NoError(t, err)
	}

	// Page 1 (limit 2, offset 0)
	page1, err := repo.GetMemoriesByNPC(npcID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Page 2 (limit 2, offset 2)
	page2, err := repo.GetMemoriesByNPC(npcID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2)
}

func TestRepository_GetMemoriesByType(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	// Create different types
	obsMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Observation"},
	}
	_, err := repo.CreateMemory(obsMemory)
	require.NoError(t, err)

	convMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeConversation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ConversationContent{},
	}
	_, err = repo.CreateMemory(convMemory)
	require.NoError(t, err)

	// Query by type
	observations, err := repo.GetMemoriesByType(npcID, MemoryTypeObservation, 10)
	require.NoError(t, err)
	assert.Len(t, observations, 1)
	assert.Equal(t, MemoryTypeObservation, observations[0].Type)

	conversations, err := repo.GetMemoriesByType(npcID, MemoryTypeConversation, 10)
	require.NoError(t, err)
	assert.Len(t, conversations, 1)
	assert.Equal(t, MemoryTypeConversation, conversations[0].Type)
}

func TestRepository_GetMemoriesByTimeframe(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()
	now := time.Now()

	// Create memory in the past (2 hours ago)
	pastMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: now.Add(-2 * time.Hour),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Past"},
	}
	_, err := repo.CreateMemory(pastMemory)
	require.NoError(t, err)

	// Create memory now
	currentMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: now,
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Current"},
	}
	_, err = repo.CreateMemory(currentMemory)
	require.NoError(t, err)

	// Query last hour (should only get currentMemory)
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)
	memories, err := repo.GetMemoriesByTimeframe(npcID, startTime, endTime)
	require.NoError(t, err)
	assert.Len(t, memories, 1)
}

func TestRepository_GetMemoriesByEntity(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()
	entityID := uuid.New()
	otherEntityID := uuid.New()

	// Observation with entity
	obsMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content: ObservationContent{
			Event:           "Saw entity",
			EntitiesPresent: []uuid.UUID{entityID},
		},
	}
	_, err := repo.CreateMemory(obsMemory)
	require.NoError(t, err)

	// Conversation with entity
	convMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeConversation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content: ConversationContent{
			Participants: []uuid.UUID{npcID, entityID},
		},
	}
	_, err = repo.CreateMemory(convMemory)
	require.NoError(t, err)

	// Memory without entity
	otherMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content: ObservationContent{
			Event:           "Alone",
			EntitiesPresent: []uuid.UUID{otherEntityID},
		},
	}
	_, err = repo.CreateMemory(otherMemory)
	require.NoError(t, err)

	// Query
	memories, err := repo.GetMemoriesByEntity(npcID, entityID)
	require.NoError(t, err)
	assert.Len(t, memories, 2)
}

func TestRepository_GetMemoriesByEmotion(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	// High emotion memory
	highMemory := Memory{
		NPCID:           npcID,
		Type:            MemoryTypeEvent,
		Timestamp:       time.Now(),
		Clarity:         1.0,
		EmotionalWeight: 0.9,
		Content:         EventContent{Description: "High emotion"},
	}
	_, err := repo.CreateMemory(highMemory)
	require.NoError(t, err)

	// Low emotion memory
	lowMemory := Memory{
		NPCID:           npcID,
		Type:            MemoryTypeEvent,
		Timestamp:       time.Now(),
		Clarity:         1.0,
		EmotionalWeight: 0.2,
		Content:         EventContent{Description: "Low emotion"},
	}
	_, err = repo.CreateMemory(lowMemory)
	require.NoError(t, err)

	// Query with threshold 0.5
	memories, err := repo.GetMemoriesByEmotion(npcID, 0.5, 10)
	require.NoError(t, err)
	assert.Len(t, memories, 1)
	assert.Equal(t, 0.9, memories[0].EmotionalWeight)
}

func TestRepository_GetMemoriesByTags_MatchAll(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	// Memory with both tags
	bothMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Both tags"},
		Tags:      []string{"tag1", "tag2"},
	}
	_, err := repo.CreateMemory(bothMemory)
	require.NoError(t, err)

	// Memory with only one tag
	oneMemory := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "One tag"},
		Tags:      []string{"tag1"},
	}
	_, err = repo.CreateMemory(oneMemory)
	require.NoError(t, err)

	// Match all tags (should only get bothMemory)
	memories, err := repo.GetMemoriesByTags(npcID, []string{"tag1", "tag2"}, true)
	require.NoError(t, err)
	assert.Len(t, memories, 1)
	assert.Contains(t, memories[0].Tags, "tag1")
	assert.Contains(t, memories[0].Tags, "tag2")
}

func TestRepository_GetMemoriesByTags_MatchAny(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	// Memory with tag1
	mem1 := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Tag1"},
		Tags:      []string{"tag1"},
	}
	_, err := repo.CreateMemory(mem1)
	require.NoError(t, err)

	// Memory with tag2
	mem2 := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Tag2"},
		Tags:      []string{"tag2"},
	}
	_, err = repo.CreateMemory(mem2)
	require.NoError(t, err)

	// Match any (should get both)
	memories, err := repo.GetMemoriesByTags(npcID, []string{"tag1", "tag2"}, false)
	require.NoError(t, err)
	assert.Len(t, memories, 2)
}

func TestRepository_UpdateMemory(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	// Create memory
	mem := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Original"},
	}
	id, err := repo.CreateMemory(mem)
	require.NoError(t, err)

	// Update clarity
	retrieved, err := repo.GetMemory(id)
	require.NoError(t, err)
	retrieved.Clarity = 0.5

	err = repo.UpdateMemory(retrieved)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetMemory(id)
	require.NoError(t, err)
	assert.Equal(t, 0.5, updated.Clarity)
}

func TestRepository_DeleteMemory(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	// Create memory
	mem := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "To delete"},
	}
	id, err := repo.CreateMemory(mem)
	require.NoError(t, err)

	// Delete
	err = repo.DeleteMemory(id)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetMemory(id)
	assert.Error(t, err)
}

func TestRepository_GetAllMemories(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()
	otherNPCID := uuid.New()

	// Create 3 memories for npcID
	for i := 0; i < 3; i++ {
		mem := Memory{
			NPCID:     npcID,
			Type:      MemoryTypeObservation,
			Timestamp: time.Now(),
			Clarity:   1.0,
			Content:   ObservationContent{Event: "Event"},
		}
		_, err := repo.CreateMemory(mem)
		require.NoError(t, err)
	}

	// Create 2 memories for other NPC
	for i := 0; i < 2; i++ {
		mem := Memory{
			NPCID:     otherNPCID,
			Type:      MemoryTypeObservation,
			Timestamp: time.Now(),
			Clarity:   1.0,
			Content:   ObservationContent{Event: "Event"},
		}
		_, err := repo.CreateMemory(mem)
		require.NoError(t, err)
	}

	// Get all for npcID
	memories, err := repo.GetAllMemories(npcID)
	require.NoError(t, err)
	assert.Len(t, memories, 3)
}
