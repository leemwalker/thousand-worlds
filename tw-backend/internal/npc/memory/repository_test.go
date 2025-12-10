package memory

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRepository_CRUD(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()

	mem := Memory{
		NPCID:           npcID,
		Type:            MemoryTypeObservation,
		Timestamp:       time.Now(),
		Clarity:         1.0,
		EmotionalWeight: 0.5,
		Content:         ObservationContent{Event: "Test Event"},
	}

	// Create
	id, err := repo.CreateMemory(mem)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)

	// Get
	fetched, err := repo.GetMemory(id)
	assert.NoError(t, err)
	assert.Equal(t, mem.Content.(ObservationContent).Event, fetched.Content.(ObservationContent).Event)

	// Update
	fetched.Clarity = 0.8
	err = repo.UpdateMemory(fetched)
	assert.NoError(t, err)

	fetched, _ = repo.GetMemory(id)
	assert.Equal(t, 0.8, fetched.Clarity)

	// Delete
	err = repo.DeleteMemory(id)
	assert.NoError(t, err)

	_, err = repo.GetMemory(id)
	assert.Error(t, err)
}

func TestRepository_Filtering(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()
	otherNPC := uuid.New()
	entityID := uuid.New()
	now := time.Now()

	// Mem 1: Observation, High Emotion
	repo.CreateMemory(Memory{
		NPCID: npcID, Type: MemoryTypeObservation, Timestamp: now, EmotionalWeight: 0.9,
		Content: ObservationContent{EntitiesPresent: []uuid.UUID{entityID}},
	})

	// Mem 2: Conversation, Low Emotion
	repo.CreateMemory(Memory{
		NPCID: npcID, Type: MemoryTypeConversation, Timestamp: now.Add(-1 * time.Hour), EmotionalWeight: 0.2,
		Content: ConversationContent{Participants: []uuid.UUID{entityID}},
	})

	// Mem 3: Other NPC
	repo.CreateMemory(Memory{
		NPCID: otherNPC, Type: MemoryTypeObservation,
	})

	// By NPC
	mems, _ := repo.GetMemoriesByNPC(npcID, 10, 0)
	assert.Len(t, mems, 2)

	// By Type
	mems, _ = repo.GetMemoriesByType(npcID, MemoryTypeObservation, 10)
	assert.Len(t, mems, 1)
	assert.Equal(t, MemoryTypeObservation, mems[0].Type)

	// By Entity
	mems, _ = repo.GetMemoriesByEntity(npcID, entityID)
	assert.Len(t, mems, 2)

	// By Emotion
	mems, _ = repo.GetMemoriesByEmotion(npcID, 0.5, 10)
	assert.Len(t, mems, 1)
	assert.Equal(t, 0.9, mems[0].EmotionalWeight)
}
