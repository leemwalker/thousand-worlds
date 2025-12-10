package action

import (
	"testing"
	"time"
)

func TestCalculateReactionTime(t *testing.T) {
	tests := []struct {
		name          string
		actionType    ActionType
		attackVariant AttackType
		agility       int
		modifiers     float64
		expected      time.Duration
		tolerance     time.Duration
	}{
		{
			name:          "Normal Attack, Agility 0",
			actionType:    ActionAttack,
			attackVariant: AttackNormal,
			agility:       0,
			modifiers:     1.0,
			expected:      1000 * time.Millisecond,
			tolerance:     0,
		},
		{
			name:          "Normal Attack, Agility 50",
			actionType:    ActionAttack,
			attackVariant: AttackNormal,
			agility:       50,
			modifiers:     1.0,
			expected:      850 * time.Millisecond, // 1000 * (1 - 0.15)
			tolerance:     1 * time.Millisecond,
		},
		{
			name:          "Quick Attack, Agility 90",
			actionType:    ActionAttack,
			attackVariant: AttackQuick,
			agility:       90,
			modifiers:     1.0,
			expected:      584 * time.Millisecond, // 800 * (1 - 0.27) = 800 * 0.73 = 584
			tolerance:     1 * time.Millisecond,
		},
		{
			name:          "Heavy Attack, Agility 20",
			actionType:    ActionAttack,
			attackVariant: AttackHeavy,
			agility:       20,
			modifiers:     1.0,
			expected:      1410 * time.Millisecond, // 1500 * (1 - 0.06) = 1500 * 0.94 = 1410
			tolerance:     1 * time.Millisecond,
		},
		{
			name:          "Defend, Agility 100 (Max)",
			actionType:    ActionDefend,
			attackVariant: "",
			agility:       100,
			modifiers:     1.0,
			expected:      350 * time.Millisecond, // 500 * (1 - 0.30) = 350
			tolerance:     1 * time.Millisecond,
		},
		{
			name:          "Slowed (x1.5)",
			actionType:    ActionAttack,
			attackVariant: AttackNormal,
			agility:       0,
			modifiers:     1.5,
			expected:      1500 * time.Millisecond,
			tolerance:     0,
		},
		{
			name:          "Hasted (x0.7)",
			actionType:    ActionAttack,
			attackVariant: AttackNormal,
			agility:       0,
			modifiers:     0.7,
			expected:      700 * time.Millisecond,
			tolerance:     0,
		},
		{
			name:          "Min Cap",
			actionType:    ActionDefend,
			attackVariant: "",
			agility:       100,
			modifiers:     0.1,                    // Extreme haste
			expected:      200 * time.Millisecond, // Capped at 200
			tolerance:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateReactionTime(tt.actionType, tt.attackVariant, tt.agility, tt.modifiers)
			diff := got - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}
