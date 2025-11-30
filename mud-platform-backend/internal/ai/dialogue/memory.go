package dialogue

import (
	"mud-platform-backend/internal/npc/memory"
	"time"

	"github.com/google/uuid"
)

func (s *DialogueService) createConversationMemory(npcID, speakerID uuid.UUID, input, output, emotion string, weight float64) {
	mem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           npcID,
		Type:            memory.MemoryTypeConversation,
		Timestamp:       time.Now(),
		Clarity:         1.0,
		EmotionalWeight: weight,
		Content: memory.ConversationContent{
			Participants: []uuid.UUID{npcID, speakerID},
			Dialogue: []memory.DialogueLine{
				{Speaker: speakerID, Text: input, Emotion: "neutral"}, // Infer speaker emotion?
				{Speaker: npcID, Text: output, Emotion: emotion},
			},
			// Topic: inferred?, Outcome: inferred?
		},
		Tags: []string{"conversation", "player_interaction"},
	}

	// Fire and forget error handling for now, or log it
	_ = s.memoryRepo.CreateMemory(mem)
}
