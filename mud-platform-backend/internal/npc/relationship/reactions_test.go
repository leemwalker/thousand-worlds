package relationship

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateReaction_Subtle(t *testing.T) {
	metrics := DriftMetrics{
		DriftLevel:     "Subtle",
		DriftScore:     0.3,
		DriftDirection: -1,
		AffectedTraits: []string{"honesty"},
	}

	r := GenerateReaction(metrics)

	assert.Contains(t, r.Comment, "seem different")
	assert.True(t, r.MemoryTrigger)
	assert.Equal(t, 0.3, r.MemoryEmotion)

	// Modifier: -5 (specific) + (-1 * 0.3 * 50 = -15) = -20 affection?
	// Or just formula?
	// Implementation adds them.
	// -5 + -15 = -20.
	assert.Equal(t, -20, r.AffinityModifier.Affection)
}

func TestGenerateReaction_Severe(t *testing.T) {
	metrics := DriftMetrics{
		DriftLevel:     "Severe",
		DriftScore:     0.8,
		DriftDirection: -1,
		AffectedTraits: []string{"aggression"},
	}

	r := GenerateReaction(metrics)

	assert.Contains(t, r.Comment, "not yourself")
	assert.Equal(t, 1.0, r.MemoryEmotion)

	// Trust: -25 (specific) + (-1 * 0.8 * 30 = -24) = -49
	assert.Equal(t, -49, r.AffinityModifier.Trust)
}
