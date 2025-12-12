package dialogue

import (
	"tw-backend/internal/npc/relationship"

	"github.com/google/uuid"
)

func (s *DialogueService) updateRelationship(npcID, speakerID uuid.UUID, output string, _ *relationship.Relationship) {
	// Simple sentiment analysis for relationship update
	// In real system, this would be more complex or use the LLM's output metadata

	delta := relationship.Affinity{}

	// Check for positive/negative keywords in OWN response
	// If NPC was friendly, relationship improves? Or depends on Speaker's input?
	// Usually depends on the interaction as a whole.
	// For MVP, let's assume if NPC generated a "joy" response, it went well.

	emotion, _ := s.inferEmotionalReaction(output)

	switch emotion {
	case "joy", "excited":
		delta.Affection = 2
		delta.Trust = 1
	case "anger":
		delta.Affection = -3
		delta.Trust = -2
	case "fear":
		delta.Trust = -1
	}

	if delta.Affection != 0 || delta.Trust != 0 || delta.Fear != 0 {
		_ = s.relationshipRepo.UpdateAffinity(npcID, speakerID, delta)
	}
}
