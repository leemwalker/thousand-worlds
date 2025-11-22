package relationship

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	CollectionName = "relationships"
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

func (r *MongoRepository) CreateRelationship(npcID, targetEntityID uuid.UUID) (Relationship, error) {
	// Check if exists first
	filter := bson.M{"npc_id": npcID, "target_entity_id": targetEntityID}
	count, err := r.collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return Relationship{}, err
	}
	if count > 0 {
		return Relationship{}, errors.New("relationship already exists")
	}

	rel := Relationship{
		ID:             uuid.New(),
		NPCID:          npcID,
		TargetEntityID: targetEntityID,
		CurrentAffinity: Affinity{
			Affection: 0,
			Trust:     0,
			Fear:      0,
		},
		// Initialize other fields as empty/zero
	}

	_, err = r.collection.InsertOne(context.Background(), rel)
	if err != nil {
		return Relationship{}, err
	}

	return rel, nil
}

func (r *MongoRepository) GetRelationship(npcID, targetEntityID uuid.UUID) (Relationship, error) {
	filter := bson.M{"npc_id": npcID, "target_entity_id": targetEntityID}
	var rel Relationship
	err := r.collection.FindOne(context.Background(), filter).Decode(&rel)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Relationship{}, errors.New("relationship not found")
		}
		return Relationship{}, err
	}
	return rel, nil
}

func (r *MongoRepository) GetAllRelationships(npcID uuid.UUID) ([]Relationship, error) {
	filter := bson.M{"npc_id": npcID}
	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []Relationship
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *MongoRepository) UpdateAffinity(npcID, targetEntityID uuid.UUID, updates Affinity) error {
	filter := bson.M{"npc_id": npcID, "target_entity_id": targetEntityID}
	update := bson.M{"$set": bson.M{"current_affinity": updates}}

	result, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("relationship not found")
	}
	return nil
}

func (r *MongoRepository) RecordInteraction(npcID, targetEntityID uuid.UUID, interaction Interaction) error {
	filter := bson.M{"npc_id": npcID, "target_entity_id": targetEntityID}

	// Push to recent_interactions and slice to keep last 20
	// MongoDB $push with $slice
	update := bson.M{
		"$push": bson.M{
			"recent_interactions": bson.M{
				"$each":  []Interaction{interaction},
				"$slice": -20, // Keep last 20
			},
		},
		"$set": bson.M{"last_interaction": interaction.Timestamp},
	}

	result, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("relationship not found")
	}
	return nil
}

func (r *MongoRepository) GetBehavioralBaseline(npcID, targetEntityID uuid.UUID) (BehavioralProfile, error) {
	rel, err := r.GetRelationship(npcID, targetEntityID)
	if err != nil {
		return BehavioralProfile{}, err
	}
	return rel.BaselineBehavior, nil
}

func (r *MongoRepository) UpdateRelationship(rel Relationship) error {
	filter := bson.M{"_id": rel.ID}
	update := bson.M{"$set": rel}

	result, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("relationship not found")
	}
	return nil
}
