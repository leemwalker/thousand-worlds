package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateImportanceScore_ByType(t *testing.T) {
	// Event should have highest base
	eventMem := Memory{
		Type:            MemoryTypeEvent,
		EmotionalWeight: 0.0,
	}
	eventScore := CalculateImportanceScore(eventMem)
	assert.Equal(t, 0.7, eventScore) // base 0.7, no emotion boost

	// Observation should have lowest base
	obsMem := Memory{
		Type:            MemoryTypeObservation,
		EmotionalWeight: 0.0,
	}
	obsScore := CalculateImportanceScore(obsMem)
	assert.Equal(t, 0.3, obsScore) // base 0.3, no emotion boost
}

func TestCalculateImportanceScore_EmotionalBoost(t *testing.T) {
	// High emotion event
	highEmotionMemory := Memory{
		Type:            MemoryTypeEvent,
		EmotionalWeight: 0.9,
	}
	score := CalculateImportanceScore(highEmotionMemory)
	// Base 0.7 * (1 + 0.9) = 0.7 * 1.9 = 1.33, but capped at 1.0
	assert.Equal(t, 1.0, score)

	// Low emotion observation
	lowEmotionMemory := Memory{
		Type:            MemoryTypeObservation,
		EmotionalWeight: 0.1,
	}
	score2 := CalculateImportanceScore(lowEmotionMemory)
	// Base 0.3 * (1 + 0.1) = 0.3 * 1.1 = 0.33
	assert.InDelta(t, 0.33, score2, 0.01)
}

func TestCalculateImportanceScore_Relationship(t *testing.T) {
	relMem := Memory{
		Type:            MemoryTypeRelationship,
		EmotionalWeight: 0.5,
	}
	score := CalculateImportanceScore(relMem)
	// Base 0.6 * (1 + 0.5) = 0.6 * 1.5 = 0.9
	assert.InDelta(t, 0.9, score, 0.01)
}

func TestCalculateImportanceScore_Conversation(t *testing.T) {
	convMem := Memory{
		Type:            MemoryTypeConversation,
		EmotionalWeight: 0.3,
	}
	score := CalculateImportanceScore(convMem)
	// Base 0.4 * (1 + 0.3) = 0.4 * 1.3 = 0.52
	assert.InDelta(t, 0.52, score, 0.01)
}
