package memory

import (
	"time"

	"github.com/google/uuid"
)

// Memory Types
const (
	MemoryTypeObservation  = "observation"
	MemoryTypeConversation = "conversation"
	MemoryTypeEvent        = "event"
	MemoryTypeRelationship = "relationship"
)

// Location represents a point in the world
type Location struct {
	X       float64   `json:"x" bson:"x"`
	Y       float64   `json:"y" bson:"y"`
	Z       float64   `json:"z" bson:"z"`
	WorldID uuid.UUID `json:"world_id" bson:"world_id"`
}

// Memory represents a stored memory for an NPC
type Memory struct {
	ID              uuid.UUID          `json:"id" bson:"_id"`
	NPCID           uuid.UUID          `json:"npc_id" bson:"npc_id"`
	Type            string             `json:"type" bson:"type"` // observation, conversation, event, relationship
	Timestamp       time.Time          `json:"timestamp" bson:"timestamp"`
	Clarity         float64            `json:"clarity" bson:"clarity"`                   // 0.0-1.0
	EmotionalWeight float64            `json:"emotional_weight" bson:"emotional_weight"` // 0.0-1.0
	DominantEmotion string             `json:"dominant_emotion" bson:"dominant_emotion"`
	EmotionProfile  map[string]float64 `json:"emotion_profile" bson:"emotion_profile"`
	AccessCount     int                `json:"access_count" bson:"access_count"`
	LastAccessed    time.Time          `json:"last_accessed" bson:"last_accessed"`
	Content         interface{}        `json:"content" bson:"content"`
	OriginalContent interface{}        `json:"original_content,omitempty" bson:"original_content,omitempty"`
	Corrupted       bool               `json:"corrupted" bson:"corrupted"`
	Tags            []string           `json:"tags" bson:"tags"`
	RelatedMemories []uuid.UUID        `json:"related_memories" bson:"related_memories"`
}

// ObservationContent represents witnessing an event
type ObservationContent struct {
	Event             string      `json:"event" bson:"event"`
	Location          Location    `json:"location" bson:"location"`
	EntitiesPresent   []uuid.UUID `json:"entities_present" bson:"entities_present"`
	WeatherConditions string      `json:"weather_conditions" bson:"weather_conditions"`
	TimeOfDay         string      `json:"time_of_day" bson:"time_of_day"`
}

// DialogueLine represents a single line in a conversation
type DialogueLine struct {
	Speaker uuid.UUID `json:"speaker" bson:"speaker"`
	Text    string    `json:"text" bson:"text"`
	Emotion string    `json:"emotion" bson:"emotion"`
}

// RelationshipImpact represents how a conversation affected a relationship
type RelationshipImpact struct {
	EntityID      uuid.UUID `json:"entity_id" bson:"entity_id"`
	AffinityDelta int       `json:"affinity_delta" bson:"affinity_delta"`
}

// ConversationContent represents a dialogue
type ConversationContent struct {
	Participants       []uuid.UUID        `json:"participants" bson:"participants"`
	Dialogue           []DialogueLine     `json:"dialogue" bson:"dialogue"`
	Location           Location           `json:"location" bson:"location"`
	Outcome            string             `json:"outcome" bson:"outcome"`
	RelationshipImpact RelationshipImpact `json:"relationship_impact" bson:"relationship_impact"`
}

// EventContent represents a significant personal experience
type EventContent struct {
	EventType         string      `json:"event_type" bson:"event_type"`
	Description       string      `json:"description" bson:"description"`
	Location          Location    `json:"location" bson:"location"`
	Participants      []uuid.UUID `json:"participants" bson:"participants"`
	Consequences      string      `json:"consequences" bson:"consequences"`
	EmotionalResponse string      `json:"emotional_response" bson:"emotional_response"`
}

// RelationshipContent represents a connection with another entity
type RelationshipContent struct {
	TargetEntityID    uuid.UUID   `json:"target_entity_id" bson:"target_entity_id"`
	Affinity          int         `json:"affinity" bson:"affinity"` // -100 to +100 (derived/aggregate)
	Trust             int         `json:"trust" bson:"trust"`
	Fear              int         `json:"fear" bson:"fear"`
	FirstImpression   string      `json:"first_impression" bson:"first_impression"`
	SharedExperiences []uuid.UUID `json:"shared_experiences" bson:"shared_experiences"`
	RelationshipType  string      `json:"relationship_type" bson:"relationship_type"`
}
