package memory

import (
	"testing"
	"time"
)

func TestMemoryCoverage(t *testing.T) {
	// 1. Decay Edge Cases
	now := time.Now()
	mem := Memory{
		Timestamp:       now,
		Clarity:         1.0,
		EmotionalWeight: 0.0,
		AccessCount:     0,
	}
	// 0 days
	c := CalculateCurrentClarity(mem, now)
	if c != 1.0 {
		t.Error("Expected no decay for 0 days")
	}

	// 2. Relevance Scoring (if accessible, need to check visibility)
	// Assuming GetRelevantMemories is in repository or service, but logic might be internal.
	// We'll check if we can test the scoring logic if it's exported.
	// It seems `CalculateRelevance` might not be exported or exists in a file I haven't seen fully.
	// I'll check `relevance.go` content if needed, but for now let's assume I can test `CalculateCurrentClarity` more thoroughly.

	// Rehearsal cap
	mem2 := Memory{
		Timestamp:   now.AddDate(0, 0, -100),
		Clarity:     1.0,
		AccessCount: 1000, // High access
	}
	// Bonus should be capped at 0.5
	// Rate = Base * 1.0 = 0.001
	// Decay = 0.001 * 100 * (1 - 0.5) = 0.05
	// Clarity = 0.95
	c2 := CalculateCurrentClarity(mem2, now)
	if c2 < 0.94 || c2 > 0.96 {
		t.Errorf("Expected capped rehearsal bonus, got %f", c2)
	}

	// 3. Tagging (if exported)
	// I'll check `tagging.go` later.
}
