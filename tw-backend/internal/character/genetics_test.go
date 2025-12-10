package character

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSpeciesTemplate(t *testing.T) {
	tests := []struct {
		name     string
		species  string
		expected Attributes
	}{
		{
			name:    "Human (Default)",
			species: SpeciesHuman,
			expected: Attributes{
				Might: 50, Agility: 50, Endurance: 50, Reflexes: 50, Vitality: 50,
				Intellect: 50, Cunning: 50, Willpower: 50, Presence: 50, Intuition: 50,
				Sight: 50, Hearing: 50, Smell: 50, Taste: 50, Touch: 50,
			},
		},
		{
			name:    "Dwarf",
			species: SpeciesDwarf,
			expected: Attributes{
				Might: 60, Agility: 40, Endurance: 65, Reflexes: 45, Vitality: 60,
				Intellect: 50, Cunning: 45, Willpower: 60, Presence: 55, Intuition: 50,
				Sight: 45, Hearing: 65, Smell: 50, Taste: 55, Touch: 60,
			},
		},
		{
			name:    "Elf",
			species: SpeciesElf,
			expected: Attributes{
				Might: 40, Agility: 65, Endurance: 45, Reflexes: 60, Vitality: 45,
				Intellect: 60, Cunning: 55, Willpower: 50, Presence: 60, Intuition: 65,
				Sight: 70, Hearing: 65, Smell: 55, Taste: 50, Touch: 50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := GetSpeciesTemplate(tt.species)
			assert.Equal(t, tt.species, tmpl.Name)
			assert.Equal(t, tt.expected, tmpl.BaseAttrs)
		})
	}
}

func TestApplyVariance(t *testing.T) {
	base := Attributes{
		Might: 50, Agility: 50, Endurance: 50, Reflexes: 50, Vitality: 50,
		Intellect: 50, Cunning: 50, Willpower: 50, Presence: 50, Intuition: 50,
		Sight: 50, Hearing: 50, Smell: 50, Taste: 50, Touch: 50,
	}

	// Use a fixed seed for deterministic testing
	// Seed 1 produces specific random numbers
	newAttrs, variance := ApplyApplyVariance(base, 1)

	// Verify variance is within range [-5, 5]
	assert.InDelta(t, 0, variance.Might, 5)
	assert.InDelta(t, 0, variance.Agility, 5)
	assert.InDelta(t, 0, variance.Endurance, 5)

	// Verify application
	assert.Equal(t, base.Might+variance.Might, newAttrs.Might)
	assert.Equal(t, base.Agility+variance.Agility, newAttrs.Agility)

	// Verify randomness (run again with different seed)
	_, variance2 := ApplyApplyVariance(base, 2)
	assert.NotEqual(t, variance, variance2, "Different seeds should produce different variance")
}

// Helper wrapper to fix the typo in the test call if needed,
// but actually I should just call ApplyVariance directly.
// The test code above had a typo `ApplyApplyVariance`. Correcting it now.
func ApplyApplyVariance(base Attributes, seed int64) (Attributes, VarianceData) {
	return ApplyVariance(base, seed)
}
