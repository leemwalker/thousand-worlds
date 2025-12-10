package memory

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionName = "memories"
)

// MongoRepository implements Repository using MongoDB
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoRepository
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection(CollectionName),
	}
}

// CreateMemory stores a new memory
func (r *MongoRepository) CreateMemory(memory Memory) (uuid.UUID, error) {
	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}

	_, err := r.collection.InsertOne(context.Background(), memory)
	if err != nil {
		return uuid.Nil, err
	}

	return memory.ID, nil
}

// GetMemory retrieves a memory by ID
func (r *MongoRepository) GetMemory(id uuid.UUID) (Memory, error) {
	var raw bson.M
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&raw)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Memory{}, errors.New("memory not found")
		}
		return Memory{}, err
	}

	return decodeMemory(raw)
}

// GetMemoriesByNPC retrieves memories for a specific NPC
func (r *MongoRepository) GetMemoriesByNPC(npcID uuid.UUID, limit, offset int) ([]Memory, error) {
	opts := options.Find().SetSort(bson.M{"timestamp": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(context.Background(), bson.M{"npc_id": npcID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	return decodeMemories(cursor)
}

// GetMemoriesByType retrieves memories by type
func (r *MongoRepository) GetMemoriesByType(npcID uuid.UUID, memoryType string, limit int) ([]Memory, error) {
	opts := options.Find().SetSort(bson.M{"timestamp": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(context.Background(), bson.M{"npc_id": npcID, "type": memoryType}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	return decodeMemories(cursor)
}

// GetMemoriesByTimeframe retrieves memories within a time range
func (r *MongoRepository) GetMemoriesByTimeframe(npcID uuid.UUID, startTime, endTime time.Time) ([]Memory, error) {
	filter := bson.M{
		"npc_id": npcID,
		"timestamp": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	return decodeMemories(cursor)
}

// GetMemoriesByEntity retrieves memories involving a specific entity
func (r *MongoRepository) GetMemoriesByEntity(npcID uuid.UUID, entityID uuid.UUID) ([]Memory, error) {
	// This query is complex because entityID can be in different fields depending on content type
	// We use $or to check all possible fields
	filter := bson.M{
		"npc_id": npcID,
		"$or": []bson.M{
			{"content.entities_present": entityID},
			{"content.participants": entityID},
			{"content.target_entity_id": entityID},
		},
	}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	return decodeMemories(cursor)
}

// GetMemoriesByEmotion retrieves memories with emotional weight above threshold
func (r *MongoRepository) GetMemoriesByEmotion(npcID uuid.UUID, minEmotionalWeight float64, limit int) ([]Memory, error) {
	opts := options.Find().SetSort(bson.M{"emotional_weight": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	filter := bson.M{
		"npc_id":           npcID,
		"emotional_weight": bson.M{"$gte": minEmotionalWeight},
	}

	cursor, err := r.collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	return decodeMemories(cursor)
}

// GetMemoriesByTags retrieves memories matching tags
func (r *MongoRepository) GetMemoriesByTags(npcID uuid.UUID, tags []string, matchAll bool) ([]Memory, error) {
	var tagFilter bson.M
	if matchAll {
		tagFilter = bson.M{"$all": tags}
	} else {
		tagFilter = bson.M{"$in": tags}
	}

	filter := bson.M{
		"npc_id": npcID,
		"tags":   tagFilter,
	}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	return decodeMemories(cursor)
}

// GetAllMemories retrieves all memories for an NPC
func (r *MongoRepository) GetAllMemories(npcID uuid.UUID) ([]Memory, error) {
	cursor, err := r.collection.Find(context.Background(), bson.M{"npc_id": npcID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	return decodeMemories(cursor)
}

// UpdateMemory updates a memory document
func (r *MongoRepository) UpdateMemory(memory Memory) error {
	_, err := r.collection.ReplaceOne(context.Background(), bson.M{"_id": memory.ID}, memory)
	return err
}

// DeleteMemory deletes a memory
func (r *MongoRepository) DeleteMemory(id uuid.UUID) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

// Helper functions for decoding

func decodeMemories(cursor *mongo.Cursor) ([]Memory, error) {
	var results []Memory
	for cursor.Next(context.Background()) {
		var raw bson.M
		if err := cursor.Decode(&raw); err != nil {
			return nil, err
		}
		mem, err := decodeMemory(raw)
		if err != nil {
			return nil, err
		}
		results = append(results, mem)
	}
	return results, nil
}

func decodeMemory(raw bson.M) (Memory, error) {
	// Marshal back to bytes to decode into struct (inefficient but safe for mixed types)
	// Or manually map fields.
	// Better approach: Decode into Memory struct but Content will be map[string]interface{}
	// Then re-decode Content based on Type.

	// Let's use the bson codec to decode the basic fields
	bytes, err := bson.Marshal(raw)
	if err != nil {
		return Memory{}, err
	}

	var base Memory
	if err := bson.Unmarshal(bytes, &base); err != nil {
		return Memory{}, err
	}

	// Now handle Content polymorphism
	// base.Content is currently a bson.D or map
	contentBytes, err := bson.Marshal(base.Content)
	if err != nil {
		return Memory{}, err
	}

	var realContent interface{}
	switch base.Type {
	case MemoryTypeObservation:
		var c ObservationContent
		err = bson.Unmarshal(contentBytes, &c)
		realContent = c
	case MemoryTypeConversation:
		var c ConversationContent
		err = bson.Unmarshal(contentBytes, &c)
		realContent = c
	case MemoryTypeEvent:
		var c EventContent
		err = bson.Unmarshal(contentBytes, &c)
		realContent = c
	case MemoryTypeRelationship:
		var c RelationshipContent
		err = bson.Unmarshal(contentBytes, &c)
		realContent = c
	default:
		// Unknown type, keep as map
		realContent = base.Content
	}

	if err != nil {
		return Memory{}, err
	}

	base.Content = realContent
	return base, nil
}
