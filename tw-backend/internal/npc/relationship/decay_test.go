package relationship

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDecay_Normal(t *testing.T) {
	rel := Relationship{
		CurrentAffinity: Affinity{Affection: 50, Trust: 50},
	}

	// 30 days -> 0.5 decay
	CalculateDecay(&rel, 30)

	// 50 - 0.5 = 49.5 -> rounds to 50? Or 49?
	// math.Round(49.5) = 50 (Go rounds to nearest even or away from zero? Go's math.Round rounds away from zero for .5)
	// Let's check implementation: int(math.Round(val))
	// 49.5 -> 50.
	// 60 days -> 1.0 decay.

	rel.CurrentAffinity.Affection = 50
	CalculateDecay(&rel, 60)
	assert.Equal(t, 49, rel.CurrentAffinity.Affection)
}

func TestCalculateDecay_Strong(t *testing.T) {
	rel := Relationship{
		CurrentAffinity: Affinity{Affection: 80},
	}

	// 60 days -> 1.0 base decay.
	// Strong (>75) -> 0.5 multiplier -> 0.5 decay.
	// 80 - 0.5 = 79.5 -> 80.
	// 120 days -> 2.0 base -> 1.0 actual.

	CalculateDecay(&rel, 120)
	assert.Equal(t, 79, rel.CurrentAffinity.Affection)
}

func TestCalculateDecay_Negative(t *testing.T) {
	rel := Relationship{
		CurrentAffinity: Affinity{Affection: -80},
	}

	// 60 days -> 1.0 base.
	// Negative (<-50) -> 2.0 multiplier -> 2.0 decay.
	// -80 + 2.0 = -78.

	CalculateDecay(&rel, 60)
	assert.Equal(t, -78, rel.CurrentAffinity.Affection)
}
