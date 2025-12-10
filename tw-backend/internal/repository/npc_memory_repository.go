package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Memory represents a single memory unit for an NPC.
type Memory struct {
	ID                      string  `bson:"_id,omitempty"`
	WorldID                 string  `bson:"world_id"`
	NPCID                   string  `bson:"npc_id"`
	Content                 string  `bson:"content"`
	DetailLevel             int     `bson:"detail_level"` // 1=Fact, 2=Summary, 3=Vivid
	ImportanceScore         float64 `bson:"importance_score"`
	Source                  string  `bson:"source"` // 'Player' or 'NPC'
	LastAccessedVirtualTime float64 `bson:"last_accessed_virtual_time"`
}

type NPCMemoryRepository struct {
	collection *mongo.Collection
}

func NewNPCMemoryRepository(db *mongo.Database) *NPCMemoryRepository {
	return &NPCMemoryRepository{
		collection: db.Collection("memories"),
	}
}

// StoreMemory inserts a new memory document into MongoDB.
func (r *NPCMemoryRepository) StoreMemory(ctx context.Context, memory Memory) error {
	_, err := r.collection.InsertOne(ctx, memory)
	if err != nil {
		return fmt.Errorf("npcMemory.StoreMemory: insert failed: %w", err)
	}
	return nil
}

// EnsureIndexes creates necessary indexes for the memories collection.
func (r *NPCMemoryRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "world_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "npc_id", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("npcMemory.EnsureIndexes: failed to create indexes: %w", err)
	}
	return nil
}

// GetMemoriesByWorldID retrieves all memories for a specific world.
func (r *NPCMemoryRepository) GetMemoriesByWorldID(ctx context.Context, worldID string) ([]Memory, error) {
	filter := bson.M{"world_id": worldID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("npcMemory.GetMemoriesByWorldID: find failed: %w", err)
	}
	defer cursor.Close(ctx)

	var memories []Memory
	if err := cursor.All(ctx, &memories); err != nil {
		return nil, fmt.Errorf("npcMemory.GetMemoriesByWorldID: decode failed: %w", err)
	}
	return memories, nil
}

// UpdateMemory updates an existing memory document.
func (r *NPCMemoryRepository) UpdateMemory(ctx context.Context, memory Memory) error {
	filter := bson.M{"_id": memory.ID}
	update := bson.M{"$set": memory}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("npcMemory.UpdateMemory: update failed: %w", err)
	}
	return nil
}
