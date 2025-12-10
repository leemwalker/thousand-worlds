package memory

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateImportance(t *testing.T) {
	now := time.Now()

	// High importance: High emotion, high clarity, recent, accessed often
	memHigh := Memory{
		EmotionalWeight: 0.9,
		Clarity:         0.9,
		Timestamp:       now,
		AccessCount:     20,
	}
	// Base = 0.81
	// Recency = 1.0
	// Access = 1.0
	// Total = 0.81 * (1 + 1 + 1) = 2.43
	impHigh := CalculateImportance(memHigh)
	assert.InDelta(t, 2.43, impHigh, 0.01)

	// Low importance: Low emotion, low clarity, old, never accessed
	memLow := Memory{
		EmotionalWeight: 0.1,
		Clarity:         0.1,
		Timestamp:       now.AddDate(-1, 0, 0), // 1 year old
		AccessCount:     0,
	}
	// Base = 0.01
	// Recency = 0.0 (approx)
	// Access = 0.0
	// Total = 0.01 * 1 = 0.01
	impLow := CalculateImportance(memLow)
	assert.InDelta(t, 0.01, impLow, 0.01)
}

func TestCalculateRelevance(t *testing.T) {
	now := time.Now()
	ctx := RelevanceContext{
		CurrentTime: now,
		Tags:        []string{"combat", "forest"},
	}

	// Relevant memory: Recent, emotional, matching tags
	memRel := Memory{
		Timestamp:       now,
		EmotionalWeight: 0.8,
		AccessCount:     5,
		Tags:            []string{"combat", "forest", "enemy"},
	}
	// Recency (1.0 * 0.3) = 0.3
	// Emotion (0.8 * 0.4) = 0.32
	// Access (0.5 * 0.1) = 0.05
	// Match (1.0 * 0.2) = 0.2
	// Total = 0.87
	scoreRel := CalculateRelevance(memRel, ctx)
	assert.InDelta(t, 0.87, scoreRel, 0.01)

	// Irrelevant memory: Old, neutral, no tags
	memIrrel := Memory{
		Timestamp:       now.AddDate(-1, 0, 0),
		EmotionalWeight: 0.1,
		AccessCount:     0,
		Tags:            []string{"cooking"},
	}
	// Recency (0.0 * 0.3) = 0.0
	// Emotion (0.1 * 0.4) = 0.04
	// Access (0.0 * 0.1) = 0.0
	// Match (0.0 * 0.2) = 0.0
	// Total = 0.04
	scoreIrrel := CalculateRelevance(memIrrel, ctx)
	assert.InDelta(t, 0.04, scoreIrrel, 0.01)
}

func TestGetRelevantMemories(t *testing.T) {
	now := time.Now()
	ctx := RelevanceContext{CurrentTime: now}

	m1 := Memory{ID: uuid.New(), Timestamp: now, EmotionalWeight: 0.9}                   // Score ~0.66
	m2 := Memory{ID: uuid.New(), Timestamp: now, EmotionalWeight: 0.1}                   // Score ~0.34
	m3 := Memory{ID: uuid.New(), Timestamp: now.AddDate(-1, 0, 0), EmotionalWeight: 0.1} // Score ~0.04

	memories := []Memory{m3, m1, m2}
	sorted := GetRelevantMemories(memories, ctx)

	assert.Equal(t, m1.ID, sorted[0].ID)
	assert.Equal(t, m2.ID, sorted[1].ID)
	assert.Equal(t, m3.ID, sorted[2].ID)
}
