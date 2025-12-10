package relationship

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyActionModifier(t *testing.T) {
	rel := Relationship{
		CurrentAffinity: Affinity{Affection: 0, Trust: 0, Fear: 0},
	}

	// Gift
	ApplyActionModifier(&rel, "gift", 50)
	assert.Equal(t, 5, rel.CurrentAffinity.Affection)
	assert.Equal(t, 5, rel.CurrentAffinity.Trust)

	// Threat
	ApplyActionModifier(&rel, "threat", 0)
	assert.Equal(t, -10, rel.CurrentAffinity.Affection) // 5 - 15
	assert.Equal(t, -5, rel.CurrentAffinity.Trust)      // 5 - 10
	assert.Equal(t, 20, rel.CurrentAffinity.Fear)

	// Bounds Check
	rel.CurrentAffinity.Affection = 95
	ApplyActionModifier(&rel, "help", 0)
	assert.Equal(t, 100, rel.CurrentAffinity.Affection) // Capped at 100
}

func TestClampAffinity(t *testing.T) {
	assert.Equal(t, 100, ClampAffinity(150))
	assert.Equal(t, -100, ClampAffinity(-150))
	assert.Equal(t, 50, ClampAffinity(50))
}
