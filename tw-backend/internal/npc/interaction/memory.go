package interaction

import (
	"tw-backend/internal/npc/memory"
	"time"

	"github.com/google/uuid"
)

// CreateConversationMemory creates a memory from a conversation
func CreateConversationMemory(conv Conversation, participantID uuid.UUID) memory.Memory {
	// Calculate Emotional Weight
	weight := 0.2 // Default casual
	switch conv.Outcome {
	case OutcomePositive:
		weight = 0.7 // Positive deep/good
	case OutcomeNegative:
		weight = 0.6 // Conflict
	case OutcomeNeutral:
		weight = 0.2 // Casual
	}

	// Create Content
	content := memory.ConversationContent{
		Participants: []uuid.UUID{conv.InitiatorID, conv.ResponderID},
		Outcome:      conv.Outcome,
		// Dialogue: conv.Dialogue, // Assuming mapping exists or struct matches
	}

	// Create Memory
	mem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           participantID,
		Type:            memory.MemoryTypeConversation,
		Timestamp:       time.Now(), // Or conv.EndTime
		Clarity:         1.0,
		EmotionalWeight: weight,
		Content:         content,
		Tags:            []string{"conversation", conv.Outcome},
	}

	return mem
}
