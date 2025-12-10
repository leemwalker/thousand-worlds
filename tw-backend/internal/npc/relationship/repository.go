package relationship

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Repository defines storage operations for relationships
type Repository interface {
	CreateRelationship(npcID, targetEntityID uuid.UUID) (Relationship, error)
	GetRelationship(npcID, targetEntityID uuid.UUID) (Relationship, error)
	GetAllRelationships(npcID uuid.UUID) ([]Relationship, error)
	UpdateAffinity(npcID, targetEntityID uuid.UUID, updates Affinity) error
	RecordInteraction(npcID, targetEntityID uuid.UUID, interaction Interaction) error
	GetBehavioralBaseline(npcID, targetEntityID uuid.UUID) (BehavioralProfile, error)
	UpdateRelationship(rel Relationship) error
}

// MockRepository is an in-memory implementation for testing
type MockRepository struct {
	rels map[string]Relationship // Key: npcID_targetID
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		rels: make(map[string]Relationship),
	}
}

func (r *MockRepository) key(npcID, targetID uuid.UUID) string {
	return npcID.String() + "_" + targetID.String()
}

func (r *MockRepository) CreateRelationship(npcID, targetEntityID uuid.UUID) (Relationship, error) {
	key := r.key(npcID, targetEntityID)
	if _, exists := r.rels[key]; exists {
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
		LastInteraction: time.Now(),
	}
	r.rels[key] = rel
	return rel, nil
}

func (r *MockRepository) GetRelationship(npcID, targetEntityID uuid.UUID) (Relationship, error) {
	key := r.key(npcID, targetEntityID)
	rel, exists := r.rels[key]
	if !exists {
		return Relationship{}, errors.New("relationship not found")
	}
	return rel, nil
}

func (r *MockRepository) GetAllRelationships(npcID uuid.UUID) ([]Relationship, error) {
	var result []Relationship
	for _, rel := range r.rels {
		if rel.NPCID == npcID {
			result = append(result, rel)
		}
	}
	return result, nil
}

func (r *MockRepository) UpdateAffinity(npcID, targetEntityID uuid.UUID, updates Affinity) error {
	key := r.key(npcID, targetEntityID)
	rel, exists := r.rels[key]
	if !exists {
		return errors.New("relationship not found")
	}

	rel.CurrentAffinity = updates
	r.rels[key] = rel
	return nil
}

func (r *MockRepository) RecordInteraction(npcID, targetEntityID uuid.UUID, interaction Interaction) error {
	key := r.key(npcID, targetEntityID)
	rel, exists := r.rels[key]
	if !exists {
		return errors.New("relationship not found")
	}

	// Add to recent interactions (limit 20)
	rel.RecentInteractions = append(rel.RecentInteractions, interaction)
	if len(rel.RecentInteractions) > 20 {
		rel.RecentInteractions = rel.RecentInteractions[len(rel.RecentInteractions)-20:]
	}
	rel.LastInteraction = interaction.Timestamp

	r.rels[key] = rel
	return nil
}

func (r *MockRepository) GetBehavioralBaseline(npcID, targetEntityID uuid.UUID) (BehavioralProfile, error) {
	key := r.key(npcID, targetEntityID)
	rel, exists := r.rels[key]
	if !exists {
		return BehavioralProfile{}, errors.New("relationship not found")
	}
	return rel.BaselineBehavior, nil
}

func (r *MockRepository) UpdateRelationship(rel Relationship) error {
	key := r.key(rel.NPCID, rel.TargetEntityID)
	if _, exists := r.rels[key]; !exists {
		return errors.New("relationship not found")
	}
	r.rels[key] = rel
	return nil
}
