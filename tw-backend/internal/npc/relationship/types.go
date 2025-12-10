package relationship

import (
	"time"

	"github.com/google/uuid"
)

// Affinity represents the emotional bond between entities
type Affinity struct {
	Affection int `json:"affection" bson:"affection"` // -100 to +100
	Trust     int `json:"trust" bson:"trust"`         // -100 to +100
	Fear      int `json:"fear" bson:"fear"`           // -100 to +100
}

// BehavioralProfile tracks personality traits (0.0 to 1.0)
type BehavioralProfile struct {
	Aggression   float64 `json:"aggression" bson:"aggression"`
	Generosity   float64 `json:"generosity" bson:"generosity"`
	Honesty      float64 `json:"honesty" bson:"honesty"`
	Sociability  float64 `json:"sociability" bson:"sociability"`
	Recklessness float64 `json:"recklessness" bson:"recklessness"`
	Loyalty      float64 `json:"loyalty" bson:"loyalty"`
}

// Interaction represents a single event affecting the relationship
type Interaction struct {
	Timestamp         time.Time         `json:"timestamp" bson:"timestamp"`
	ActionType        string            `json:"action_type" bson:"action_type"`
	AffinityDelta     Affinity          `json:"affinity_delta" bson:"affinity_delta"`
	BehavioralContext BehavioralProfile `json:"behavioral_context" bson:"behavioral_context"`
}

// Relationship is the core document tracking the connection
type Relationship struct {
	ID                 uuid.UUID         `json:"id" bson:"_id,omitempty"`
	NPCID              uuid.UUID         `json:"npc_id" bson:"npc_id"`
	TargetEntityID     uuid.UUID         `json:"target_entity_id" bson:"target_entity_id"`
	CurrentAffinity    Affinity          `json:"current_affinity" bson:"current_affinity"`
	BaselineBehavior   BehavioralProfile `json:"baseline_behavior" bson:"baseline_behavior"`
	RecentInteractions []Interaction     `json:"recent_interactions" bson:"recent_interactions"`
	LastInteraction    time.Time         `json:"last_interaction" bson:"last_interaction"`
}

// DriftMetrics captures how much current behavior deviates from baseline
type DriftMetrics struct {
	DriftScore     float64  `json:"drift_score"`
	DriftDirection int      `json:"drift_direction"` // +1 (Improved), -1 (Worsened)
	AffectedTraits []string `json:"affected_traits"`
	DriftLevel     string   `json:"drift_level"` // None, Subtle, Moderate, Severe
}
