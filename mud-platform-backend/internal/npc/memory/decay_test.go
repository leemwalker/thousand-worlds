package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateCurrentClarity_LinearDecay(t *testing.T) {
	now := time.Now()
	// 100 days ago
	created := now.AddDate(0, 0, -100)

	mem := Memory{
		Timestamp:       created,
		Clarity:         1.0,
		EmotionalWeight: 0.5,
		AccessCount:     0,
	}

	// Formula: effectiveRate = BaseDecayRate * (1.0 - (weight * 0.5))
	// Rate = 0.001 * (1.0 - 0.25) = 0.00075
	// Decay = 0.00075 * 100 = 0.075
	// Clarity = 1.0 - 0.075 = 0.925

	clarity := CalculateCurrentClarity(mem, now)
	assert.InDelta(t, 0.925, clarity, 0.01)
}

func TestCalculateCurrentClarity_EmotionImpact(t *testing.T) {
	now := time.Now()
	created := now.AddDate(0, 0, -100)

	// High Emotion (1.0) -> Rate = Base * (1 - 0.5) = 0.0005
	memHigh := Memory{
		Timestamp:       created,
		Clarity:         1.0,
		EmotionalWeight: 1.0,
		AccessCount:     0,
	}
	// Decay = 0.0005 * 100 = 0.05
	// Clarity = 0.95
	clarityHigh := CalculateCurrentClarity(memHigh, now)
	assert.InDelta(t, 0.95, clarityHigh, 0.01)

	// Low Emotion (0.0) -> Rate = Base * (1 - 0) = 0.001
	memLow := Memory{
		Timestamp:       created,
		Clarity:         1.0,
		EmotionalWeight: 0.0,
		AccessCount:     0,
	}
	// Decay = 0.001 * 100 = 0.1
	// Clarity = 0.90
	clarityLow := CalculateCurrentClarity(memLow, now)
	assert.InDelta(t, 0.90, clarityLow, 0.01)
}

func TestCalculateCurrentClarity_RehearsalBonus(t *testing.T) {
	now := time.Now()
	created := now.AddDate(0, 0, -100)

	// Max Rehearsal (Access >= 10 for 0.5 bonus)
	mem := Memory{
		Timestamp:       created,
		Clarity:         1.0,
		EmotionalWeight: 0.5, // Rate 0.00075
		AccessCount:     10,  // Bonus 0.5
	}

	// Rate = 0.00075
	// Decay = 0.00075 * 100 * (1 - 0.5) = 0.075 * 0.5 = 0.0375
	// Clarity = 1.0 - 0.0375 = 0.9625

	clarity := CalculateCurrentClarity(mem, now)
	assert.InDelta(t, 0.9625, clarity, 0.01)
}

func TestCalculateCurrentClarity_Floor(t *testing.T) {
	now := time.Now()
	created := now.AddDate(-10, 0, 0) // 10 years ago

	mem := Memory{
		Timestamp:       created,
		Clarity:         1.0,
		EmotionalWeight: 0.5,
		AccessCount:     0,
	}

	// Should hit floor 0.1
	clarity := CalculateCurrentClarity(mem, now)
	assert.Equal(t, 0.1, clarity)
}
