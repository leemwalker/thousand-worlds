//go:build integration
// +build integration

package memory

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Integration tests for MongoDB repository
// Run with: go test -tags=integration -v ./internal/npc/memory/...
// Requires: MongoDB running on localhost:27017 or TEST_MONGODB_URI env var

func getMongoClient(t *testing.T) *mongo.Client {
	mongoURI := os.Getenv("TEST_MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err)

	// Ping to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Skip("MongoDB not available, skipping integration tests")
	}

	return client
}

func setupTestDB(t *testing.T) (*MongoRepository, func()) {
	client := getMongoClient(t)

	// Use a test database
	db := client.Database("test_thousand_worlds_memory")
	repo := NewMongoRepository(db)

	// Cleanup function
	cleanup := func() {
		ctx := context.Background()
		db.Drop(ctx)
		client.Disconnect(ctx)
	}

	return repo, cleanup
}

func TestMongoRepository_Integration_CreateAndGet(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	npcID := uuid.New()
	mem := Memory{
		NPCID:           npcID,
		Type:            MemoryTypeObservation,
		Timestamp:       time.Now(),
		Clarity:         1.0,
		EmotionalWeight: 0.5,
		Content: ObservationContent{
			Event: "Test observation",
			Location: Location{
				X:       100,
				Y:       200,
				Z:       0,
				WorldID: uuid.New(),
			},
		},
		Tags: []string{"test", "integration"},
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

	// Verify content
	content, ok := retrieved.Content.(ObservationContent)
	require.True(t, ok, "Content should be ObservationContent")
	assert.Equal(t, "Test observation", content.Event)
}

func TestMongoRepository_Integration_GetByNPC(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

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
		Content:   ObservationContent{Event: "Other"},
	}
	_, err := repo.CreateMemory(mem)
	require.NoError(t, err)

	// Query
	memories, err := repo.GetMemoriesByNPC(npcID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, memories, 3)
}

func TestMongoRepository_Integration_GetByType(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	npcID := uuid.New()

	// Create observation
	obs := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Observation"},
	}
	_, err := repo.CreateMemory(obs)
	require.NoError(t, err)

	// Create conversation
	conv := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeConversation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ConversationContent{Participants: []uuid.UUID{npcID}},
	}
	_, err = repo.CreateMemory(conv)
	require.NoError(t, err)

	// Query observations
	memories, err := repo.GetMemoriesByType(npcID, MemoryTypeObservation, 10)
	require.NoError(t, err)
	assert.Len(t, memories, 1)
	assert.Equal(t, MemoryTypeObservation, memories[0].Type)
}

func TestMongoRepository_Integration_GetByEntity(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	npcID := uuid.New()
	entityID := uuid.New()

	// Create memory with entity in observation
	obs := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content: ObservationContent{
			Event:           "Saw entity",
			EntitiesPresent: []uuid.UUID{entityID},
		},
	}
	_, err := repo.CreateMemory(obs)
	require.NoError(t, err)

	// Create memory with entity in conversation
	conv := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeConversation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content: ConversationContent{
			Participants: []uuid.UUID{npcID, entityID},
		},
	}
	_, err = repo.CreateMemory(conv)
	require.NoError(t, err)

	// Query
	memories, err := repo.GetMemoriesByEntity(npcID, entityID)
	require.NoError(t, err)
	assert.Len(t, memories, 2)
}

func TestMongoRepository_Integration_Update(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	npcID := uuid.New()
	mem := Memory{
		NPCID:     npcID,
		Type:      MemoryTypeObservation,
		Timestamp: time.Now(),
		Clarity:   1.0,
		Content:   ObservationContent{Event: "Original"},
	}

	id, err := repo.CreateMemory(mem)
	require.NoError(t, err)

	// Update
	retrieved, err := repo.GetMemory(id)
	require.NoError(t, err)
	retrieved.Clarity = 0.5

	err = repo.UpdateMemory(retrieved)
	require.NoError(t, err)

	// Verify
	updated, err := repo.GetMemory(id)
	require.NoError(t, err)
	assert.Equal(t, 0.5, updated.Clarity)
}

func TestMongoRepository_Integration_Delete(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	npcID := uuid.New()
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

	// Verify
	_, err = repo.GetMemory(id)
	assert.Error(t, err)
}

func TestMongoRepository_Integration_PolymorphicContent(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	npcID := uuid.New()
	worldID := uuid.New()

	testCases := []struct {
		name   string
		memory Memory
		verify func(t *testing.T, content interface{})
	}{
		{
			name: "Observation",
			memory: Memory{
				NPCID:     npcID,
				Type:      MemoryTypeObservation,
				Timestamp: time.Now(),
				Clarity:   1.0,
				Content: ObservationContent{
					Event: "Test event",
					Location: Location{
						X:       50,
						Y:       60,
						Z:       0,
						WorldID: worldID,
					},
					EntitiesPresent: []uuid.UUID{uuid.New()},
				},
			},
			verify: func(t *testing.T, content interface{}) {
				obs, ok := content.(ObservationContent)
				require.True(t, ok)
				assert.Equal(t, "Test event", obs.Event)
				assert.Equal(t, 50.0, obs.Location.X)
			},
		},
		{
			name: "Conversation",
			memory: Memory{
				NPCID:     npcID,
				Type:      MemoryTypeConversation,
				Timestamp: time.Now(),
				Clarity:   1.0,
				Content: ConversationContent{
					Participants: []uuid.UUID{npcID, uuid.New()},
					Outcome:      "positive",
				},
			},
			verify: func(t *testing.T, content interface{}) {
				conv, ok := content.(ConversationContent)
				require.True(t, ok)
				assert.Equal(t, "positive", conv.Outcome)
				assert.Len(t, conv.Participants, 2)
			},
		},
		{
			name: "Event",
			memory: Memory{
				NPCID:     npcID,
				Type:      MemoryTypeEvent,
				Timestamp: time.Now(),
				Clarity:   1.0,
				Content: EventContent{
					Description:  "Important event",
					Participants: []uuid.UUID{npcID},
				},
			},
			verify: func(t *testing.T, content interface{}) {
				evt, ok := content.(EventContent)
				require.True(t, ok)
				assert.Equal(t, "Important event", evt.Description)
			},
		},
		{
			name: "Relationship",
			memory: Memory{
				NPCID:     npcID,
				Type:      MemoryTypeRelationship,
				Timestamp: time.Now(),
				Clarity:   1.0,
				Content: RelationshipContent{
					TargetEntityID: uuid.New(),
					Affinity:       75,
				},
			},
			verify: func(t *testing.T, content interface{}) {
				rel, ok := content.(RelationshipContent)
				require.True(t, ok)
				assert.Equal(t, 75, rel.Affinity)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := repo.CreateMemory(tc.memory)
			require.NoError(t, err)

			retrieved, err := repo.GetMemory(id)
			require.NoError(t, err)
			tc.verify(t, retrieved.Content)
		})
	}
}

func TestMongoRepository_Integration_TagsAndEmotion(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	npcID := uuid.New()

	// High emotion memory with tags
	highEmotion := Memory{
		NPCID:           npcID,
		Type:            MemoryTypeEvent,
		Timestamp:       time.Now(),
		Clarity:         1.0,
		EmotionalWeight: 0.9,
		Content:         EventContent{Description: "High emotion"},
		Tags:            []string{"important", "emotional"},
	}
	_, err := repo.CreateMemory(highEmotion)
	require.NoError(t, err)

	// Low emotion memory
	lowEmotion := Memory{
		NPCID:           npcID,
		Type:            MemoryTypeEvent,
		Timestamp:       time.Now(),
		Clarity:         1.0,
		EmotionalWeight: 0.2,
		Content:         EventContent{Description: "Low emotion"},
		Tags:            []string{"mundane"},
	}
	_, err = repo.CreateMemory(lowEmotion)
	require.NoError(t, err)

	// Query by emotion
	memories, err := repo.GetMemoriesByEmotion(npcID, 0.5, 10)
	require.NoError(t, err)
	assert.Len(t, memories, 1)
	assert.Equal(t, 0.9, memories[0].EmotionalWeight)

	// Query by tags (match all)
	taggedMemories, err := repo.GetMemoriesByTags(npcID, []string{"important", "emotional"}, true)
	require.NoError(t, err)
	assert.Len(t, taggedMemories, 1)

	// Query by tags (match any)
	anyTagMemories, err := repo.GetMemoriesByTags(npcID, []string{"important", "mundane"}, false)
	require.NoError(t, err)
	assert.Len(t, anyTagMemories, 2)
}
