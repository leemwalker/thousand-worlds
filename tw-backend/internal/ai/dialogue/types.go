package dialogue

import (
	"tw-backend/internal/ai/cache"
	"tw-backend/internal/ai/ollama"
	"tw-backend/internal/ai/prompt"
	"tw-backend/internal/character"
	"tw-backend/internal/npc/desire"
	"tw-backend/internal/npc/memory"
	"tw-backend/internal/npc/personality"
	"tw-backend/internal/npc/relationship"

	"github.com/google/uuid"
)

// Intent Constants
const (
	IntentSeekingFood       = "seeking_food"
	IntentSeekingConnection = "seeking_connection"
	IntentSeekingSafety     = "seeking_safety"
	IntentFocusedOnGoal     = "focused_on_goal"
	IntentNeutral           = "neutral"
)

// DialogueRequest represents the payload sent to the AI Gateway (or used internally)
type DialogueRequest struct {
	NPCID          uuid.UUID                  `json:"npc_id"`
	SpeakerID      uuid.UUID                  `json:"speaker_id"`
	Input          string                     `json:"input"`
	NPCState       NPCState                   `json:"npc_state"`
	Relationship   *relationship.Relationship `json:"relationship"`
	RecentMemories []memory.Memory            `json:"recent_memories"`
	DriftMetrics   *relationship.DriftMetrics `json:"drift_metrics,omitempty"`
	Intent         string                     `json:"intent"`
}

// NPCState aggregates necessary NPC data
type NPCState struct {
	Name        string                   `json:"name"`
	Personality *personality.Personality `json:"personality"`
	Mood        *personality.Mood        `json:"mood"`
	Attributes  character.Attributes     `json:"attributes"`
}

// DialogueResponse represents the generated dialogue and metadata
type DialogueResponse struct {
	Text              string  `json:"text"`
	EmotionalReaction string  `json:"emotional_reaction"`
	EmotionalWeight   float64 `json:"emotional_weight"`
	UsedFallback      bool    `json:"used_fallback"`
}

// Repository Interfaces
type NPCRepository interface {
	GetNPC(id uuid.UUID) (*character.Character, error) // Assuming Character holds basic info
	GetPersonality(id uuid.UUID) (*personality.Personality, error)
	GetMood(id uuid.UUID) (*personality.Mood, error)
	// Add other getters as needed based on actual repo structure
}

type MemoryRepository interface {
	GetMemories(npcID uuid.UUID, limit int) ([]memory.Memory, error)
	CreateMemory(mem memory.Memory) error
}

type RelationshipRepository interface {
	GetRelationship(npcID, targetID uuid.UUID) (*relationship.Relationship, error)
	UpdateAffinity(npcID, targetID uuid.UUID, affinity relationship.Affinity) error
	GetDriftMetrics(npcID uuid.UUID) (*relationship.DriftMetrics, error)
	GetBehavioralProfile(npcID uuid.UUID) (relationship.BehavioralProfile, error) // Baseline
	GetCurrentBehavior(npcID uuid.UUID) (relationship.BehavioralProfile, error)   // Current
}

type DesireRepository interface {
	GetDesireProfile(npcID uuid.UUID) (*desire.DesireProfile, error)
}

// DialogueService orchestrates the dialogue generation
type DialogueService struct {
	npcRepo          NPCRepository
	memoryRepo       MemoryRepository
	relationshipRepo RelationshipRepository
	desireRepo       DesireRepository
	promptBuilder    *prompt.PromptBuilder
	ollamaClient     *ollama.OllamaClient
	dialogueCache    *cache.DialogueCache
}
