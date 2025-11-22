package character

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePointBuy(t *testing.T) {
	base := Attributes{Might: 50}    // Simplified base
	variance := Attributes{Might: 0} // No variance for simplicity

	tests := []struct {
		name      string
		increases map[string]int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid - No Increases",
			increases: map[string]int{},
			wantErr:   false,
		},
		{
			name:      "Valid - Tier 1 Cost (10 points for +10)",
			increases: map[string]int{AttrMight: 10},
			wantErr:   false,
		},
		{
			name:      "Valid - Tier 2 Cost (30 points for +20)",
			increases: map[string]int{AttrMight: 20},
			wantErr:   false,
			// Cost: 10*1 + 10*2 = 30
		},
		{
			name:      "Valid - Tier 3 Cost (60 points for +30)",
			increases: map[string]int{AttrMight: 30},
			wantErr:   false,
			// Cost: 10*1 + 10*2 + 10*3 = 10 + 20 + 30 = 60
		},
		{
			name:      "Invalid - Exceeds Max Increase (+31)",
			increases: map[string]int{AttrMight: 31},
			wantErr:   true,
			errMsg:    "exceeds max of +30",
		},
		{
			name:      "Invalid - Negative Increase",
			increases: map[string]int{AttrMight: -1},
			wantErr:   true,
			errMsg:    "cannot decrease",
		},
		{
			name: "Invalid - Exceeds Budget (101 points)",
			// +30 Might = 60 points
			// +20 Agility = 30 points
			// +10 Endurance = 10 points
			// Total = 100 points (Valid)
			// Add +1 Reflexes = 1 point -> 101 points (Invalid)
			increases: map[string]int{
				AttrMight:     30,
				AttrAgility:   20,
				AttrEndurance: 10,
				AttrReflexes:  1,
			},
			wantErr: true,
			errMsg:  "exceeds budget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePointBuy(base, variance, tt.increases)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplyPointBuy(t *testing.T) {
	base := Attributes{Might: 50}
	variance := Attributes{Might: 5} // +5 variance
	increases := map[string]int{AttrMight: 10}

	final := ApplyPointBuy(base, variance, increases)

	// Expected: 50 (base) + 5 (variance) + 10 (increase) = 65
	assert.Equal(t, 65, final.Might)
}

func TestApplyPointBuy_AllAttributes(t *testing.T) {
	base := Attributes{}
	variance := Attributes{}
	increases := map[string]int{
		AttrMight: 1, AttrAgility: 1, AttrEndurance: 1, AttrReflexes: 1, AttrVitality: 1,
		AttrIntellect: 1, AttrCunning: 1, AttrWillpower: 1, AttrPresence: 1, AttrIntuition: 1,
		AttrSight: 1, AttrHearing: 1, AttrSmell: 1, AttrTaste: 1, AttrTouch: 1,
	}

	final := ApplyPointBuy(base, variance, increases)

	assert.Equal(t, 1, final.Might)
	assert.Equal(t, 1, final.Agility)
	assert.Equal(t, 1, final.Endurance)
	assert.Equal(t, 1, final.Reflexes)
	assert.Equal(t, 1, final.Vitality)
	assert.Equal(t, 1, final.Intellect)
	assert.Equal(t, 1, final.Cunning)
	assert.Equal(t, 1, final.Willpower)
	assert.Equal(t, 1, final.Presence)
	assert.Equal(t, 1, final.Intuition)
	assert.Equal(t, 1, final.Sight)
	assert.Equal(t, 1, final.Hearing)
	assert.Equal(t, 1, final.Smell)
	assert.Equal(t, 1, final.Taste)
	assert.Equal(t, 1, final.Touch)
}
