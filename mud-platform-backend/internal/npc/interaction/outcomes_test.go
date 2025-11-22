package interaction

import (
	"mud-platform-backend/internal/npc/relationship"
	"testing"

	"github.com/google/uuid"
)

func TestApplyInteractionOutcome(t *testing.T) {
	// Positive Outcome
	update := ApplyInteractionOutcome(OutcomePositive)
	if update.AffectionDelta != 3 || update.TrustDelta != 2 {
		t.Errorf("Expected +3/+2 for positive, got %+d/%+d", update.AffectionDelta, update.TrustDelta)
	}

	// Apply to Relationship
	rel := relationship.Relationship{
		CurrentAffinity: relationship.Affinity{Affection: 50, Trust: 50},
	}
	UpdateRelationship(&rel, update)

	if rel.CurrentAffinity.Affection != 53 {
		t.Errorf("Expected Affection 53, got %d", rel.CurrentAffinity.Affection)
	}
}

func TestCreateConversationMemory(t *testing.T) {
	conv := Conversation{
		ID:          uuid.New(),
		InitiatorID: uuid.New(),
		ResponderID: uuid.New(),
		Outcome:     OutcomePositive,
	}

	mem := CreateConversationMemory(conv, conv.InitiatorID)

	if mem.EmotionalWeight != 0.7 {
		t.Errorf("Expected emotional weight 0.7 for positive outcome, got %f", mem.EmotionalWeight)
	}

	if mem.Type != "conversation" {
		t.Errorf("Expected memory type conversation, got %s", mem.Type)
	}
}
