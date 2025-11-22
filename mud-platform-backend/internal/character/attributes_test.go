package character

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSecondaryAttributes(t *testing.T) {
	tests := []struct {
		name     string
		input    Attributes
		expected SecondaryAttributes
	}{
		{
			name:  "All Zeros",
			input: Attributes{},
			expected: SecondaryAttributes{
				MaxHP:      0,
				MaxStamina: 0,
				MaxFocus:   0,
				MaxMana:    0,
				MaxNerve:   0,
			},
		},
		{
			name: "Balanced 10s",
			input: Attributes{
				Might:     10,
				Agility:   10,
				Endurance: 10,
				Reflexes:  10,
				Vitality:  10,
				Intellect: 10,
				Willpower: 10,
				Presence:  10,
				Intuition: 10,
			},
			expected: SecondaryAttributes{
				MaxHP:      100, // 10 * 10
				MaxStamina: 100, // (10*7) + (10*3) = 70 + 30
				MaxFocus:   100, // (10*6) + (10*4) = 60 + 40
				MaxMana:    100, // (10*6) + (10*4) = 60 + 40
				MaxNerve:   100, // (10*5) + (10*3) + (10*2) = 50 + 30 + 20
			},
		},
		{
			name: "High Physical",
			input: Attributes{
				Might:     20,
				Endurance: 20,
				Vitality:  20,
			},
			expected: SecondaryAttributes{
				MaxHP:      200, // 20 * 10
				MaxStamina: 200, // (20*7) + (20*3) = 140 + 60
				MaxFocus:   0,
				MaxMana:    0,
				MaxNerve:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSecondaryAttributes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
