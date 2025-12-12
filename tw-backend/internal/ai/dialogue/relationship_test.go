package dialogue

import (
	"testing"
	"tw-backend/internal/npc/relationship"

	"github.com/google/uuid"
)

func TestUpdateRelationship(t *testing.T) {
	mockRelRepo := new(MockRelRepo)
	s := &DialogueService{
		relationshipRepo: mockRelRepo,
	}

	npcID := uuid.New()
	speakerID := uuid.New()

	// Case 1: Joy -> Positive Affinity
	mockRelRepo.On("UpdateAffinity", npcID, speakerID, relationship.Affinity{
		Affection: 2,
		Trust:     1,
	}).Return(nil).Once()

	s.updateRelationship(npcID, speakerID, "This is wonderful and makes me happy!", nil)

	// Case 2: Anger -> Negative Affinity
	mockRelRepo.On("UpdateAffinity", npcID, speakerID, relationship.Affinity{
		Affection: -3,
		Trust:     -2,
	}).Return(nil).Once()

	s.updateRelationship(npcID, speakerID, "This is unacceptable! I am furious!", nil)

	// Case 3: Neutral/Unknown -> No Update
	// Strict explanation: UpdateAffinity should NOT be called
	// We rely on .Once() above to ensure calls are matched, and here we expect no call
	s.updateRelationship(npcID, speakerID, "Just a normal statement.", nil)

	mockRelRepo.AssertExpectations(t)
}
