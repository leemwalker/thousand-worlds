package emotion

import (
	"testing"

	"mud-platform-backend/internal/npc/memory"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSimilarity(t *testing.T) {
	current := EmotionProfile{Fear: 0.8, Anger: 0.2}
	mem := EmotionProfile{Fear: 0.7, Anger: 0.3}

	// Fear sim: 1 - |0.8 - 0.7| = 0.9
	// Anger sim: 1 - |0.2 - 0.3| = 0.9
	// Avg: 0.9

	sim := CalculateSimilarity(current, mem)
	assert.InDelta(t, 0.9, sim, 0.01)
}

func TestGetSimilarMemories(t *testing.T) {
	current := EmotionProfile{Joy: 0.9}

	mem1 := memory.Memory{
		EmotionProfile: EmotionProfile{Joy: 0.8},  // Sim 0.9
		Clarity:        1.0, EmotionalWeight: 1.0, // Imp 2.0
	}
	mem2 := memory.Memory{
		EmotionProfile: EmotionProfile{Joy: 0.1}, // Sim 0.2 (Threshold fail)
	}
	mem3 := memory.Memory{
		EmotionProfile: EmotionProfile{Joy: 0.9},  // Sim 1.0
		Clarity:        0.5, EmotionalWeight: 0.5, // Imp 0.75
	}

	mems := []memory.Memory{mem1, mem2, mem3}

	results := GetSimilarMemories(current, mems)

	assert.Len(t, results, 2)
	assert.Equal(t, mem1, results[0]) // Higher importance should win?
	// Mem1 Score: 0.9*0.5 + 2.0*0.3 + 0.1 = 0.45 + 0.6 + 0.1 = 1.15
	// Mem3 Score: 1.0*0.5 + 0.75*0.3 + 0.1 = 0.5 + 0.225 + 0.1 = 0.825
}
