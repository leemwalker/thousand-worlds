package damage

import (
	"tw-backend/internal/character"
	"testing"
)

func TestCalculateDamage(t *testing.T) {
	// Setup common objects
	attrs := character.Attributes{
		Might:   50,
		Agility: 50,
		Cunning: 50,
	}

	weapon := &Weapon{
		Name:          "Test Sword",
		Type:          WeaponSlashing,
		BaseDamage:    20,
		Durability:    100,
		MaxDurability: 100,
	}

	armor := &Armor{
		Name:          "Leather Armor",
		Type:          ArmorLeather,
		Durability:    100,
		MaxDurability: 100,
	}

	tests := []struct {
		name        string
		attrs       character.Attributes
		weapon      *Weapon
		skill       int
		armor       *Armor
		roll        int
		isHeavy     bool
		expectedMin int
		expectedMax int
	}{
		{
			name:    "Basic Hit (Roll 50)",
			attrs:   attrs,
			weapon:  weapon,
			skill:   0,
			armor:   nil,
			roll:    50,
			isHeavy: false,
			// Base 20 * Skill 1.0 * Attr(50) 1.25 * Roll 0.5 = 12.5 -> 12
			expectedMin: 12,
			expectedMax: 12,
		},
		{
			name:    "High Skill (100)",
			attrs:   attrs,
			weapon:  weapon,
			skill:   100, // 1.5x
			armor:   nil,
			roll:    50,
			isHeavy: false,
			// Base 20 * Skill 1.5 * Attr 1.25 * Roll 0.5 = 18.75 -> 18
			expectedMin: 18,
			expectedMax: 18,
		},
		{
			name:    "Armor Reduction (Leather vs Slashing 15%)",
			attrs:   attrs,
			weapon:  weapon,
			skill:   0,
			armor:   armor,
			roll:    50,
			isHeavy: false,
			// Raw: 20 * 1.0 * 1.25 * 0.5 = 12.5 -> 12
			// Red: 15% -> 12 * 0.85 = 10.2 -> 10
			expectedMin: 10,
			expectedMax: 10,
		},
		{
			name:    "Critical Hit (Roll 100)",
			attrs:   attrs,
			weapon:  weapon,
			skill:   0,
			armor:   armor,
			roll:    100,
			isHeavy: false,
			// Raw: 20 * 1.0 * 1.25 * 1.0 * 2.0 (Crit) = 50
			// Armor: 15% * 0.5 (Ignore) = 7.5%
			// Final: 50 * 0.925 = 46.25 -> 46
			expectedMin: 46,
			expectedMax: 46,
		},
		{
			name:        "Fumble (Roll 1)",
			attrs:       attrs,
			weapon:      weapon,
			skill:       0,
			armor:       nil,
			roll:        1,
			isHeavy:     false,
			expectedMin: 0,
			expectedMax: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDamage(tt.attrs, tt.weapon, tt.skill, tt.armor, tt.roll, tt.isHeavy)
			if result.FinalDamage < tt.expectedMin || result.FinalDamage > tt.expectedMax {
				t.Errorf("Expected damage between %d and %d, got %d", tt.expectedMin, tt.expectedMax, result.FinalDamage)
			}
		})
	}
}
