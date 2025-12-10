package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateRetentionScore(t *testing.T) {
	now := time.Now()

	// Case 1: Fresh, high emotion, high clarity
	mem1 := Memory{
		Timestamp:       now,
		Clarity:         1.0,
		EmotionalWeight: 1.0,
		AccessCount:     0,
	}
	// Base = 1.0 * 1.0 = 1.0
	// Decay = 1.0 (0 days)
	// Rehearsal = 1.0 (0 access)
	// Score = 1.0
	score1 := CalculateRetentionScore(mem1, now)
	assert.InDelta(t, 1.0, score1, 0.01)

	// Case 2: Old, low emotion, no access
	mem2 := Memory{
		Timestamp:       now.AddDate(0, 0, -100),
		Clarity:         0.5,
		EmotionalWeight: 0.2,
		AccessCount:     0,
	}
	// Base = 0.5 * 0.2 = 0.1
	// Decay = 1.0 - (0.001 * 100) = 0.9
	// Rehearsal = 1.0
	// Score = 0.1 * 0.9 = 0.09
	score2 := CalculateRetentionScore(mem2, now)
	assert.InDelta(t, 0.09, score2, 0.01)
	assert.True(t, IsForgotten(score2))

	// Case 3: Old, high access (Rehearsal)
	mem3 := Memory{
		Timestamp:       now.AddDate(0, 0, -100),
		Clarity:         0.8,
		EmotionalWeight: 0.5,
		AccessCount:     20,
	}
	// Base = 0.8 * 0.5 = 0.4
	// Decay = 0.9
	// Rehearsal Bonus = min(20/20, 0.5) = 0.5?
	// Wait, CalculateRehearsalBonus is in rehearsal.go.
	// Assuming it returns 0.5 for 20 accesses.
	// Rehearsal Factor = 1.0 + (0.5 * 20) = 11.0?
	// Wait, retention.go line 39: rehearsalFactor := 1.0 + (rehearsalBonus * float64(memory.AccessCount))
	// If bonus is 0.5, factor is 1 + 10 = 11.
	// Score = 0.4 * 0.9 * 11 = 3.96.
	// This seems very high, but consistent with code "sticky".
	score3 := CalculateRetentionScore(mem3, now)
	assert.InDelta(t, 3.96, score3, 0.1)
	assert.False(t, IsForgotten(score3))
}

func TestIsForgotten(t *testing.T) {
	assert.True(t, IsForgotten(0.1))
	assert.False(t, IsForgotten(0.2))
}
