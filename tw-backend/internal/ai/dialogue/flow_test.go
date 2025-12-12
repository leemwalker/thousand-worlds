package dialogue

import (
	"testing"
	"tw-backend/internal/npc/desire"

	"github.com/stretchr/testify/assert"
)

func TestDetermineIntent(t *testing.T) {
	s := &DialogueService{}

	tests := []struct {
		name           string
		needs          map[string]*desire.Need
		expectedIntent string
	}{
		{
			name: "Hunger > 70",
			needs: map[string]*desire.Need{
				desire.NeedHunger: {Name: desire.NeedHunger, Value: 75.0},
				desire.NeedSafety: {Name: desire.NeedSafety, Value: 10.0},
			},
			expectedIntent: IntentSeekingFood,
		},
		{
			name: "Safety > 50",
			needs: map[string]*desire.Need{
				desire.NeedHunger: {Name: desire.NeedHunger, Value: 10.0},
				desire.NeedSafety: {Name: desire.NeedSafety, Value: 60.0},
			},
			expectedIntent: IntentSeekingSafety,
		},
		{
			name: "No dominant need",
			needs: map[string]*desire.Need{
				desire.NeedHunger: {Name: desire.NeedHunger, Value: 10.0},
				desire.NeedSafety: {Name: desire.NeedSafety, Value: 10.0},
			},
			expectedIntent: IntentNeutral,
		},
		{
			name: "Priority check: Hunger vs Safety",
			needs: map[string]*desire.Need{
				desire.NeedHunger: {Name: desire.NeedHunger, Value: 80.0},
				desire.NeedSafety: {Name: desire.NeedSafety, Value: 60.0},
			},
			expectedIntent: IntentSeekingFood,
		},
		{
			name: "Top Need Below Threshold",
			needs: map[string]*desire.Need{
				desire.NeedHunger: {Name: desire.NeedHunger, Value: 50.0}, // Threshold is 70
			},
			expectedIntent: IntentNeutral,
		},
		{
			name: "Companionship > 60",
			needs: map[string]*desire.Need{
				desire.NeedCompanionship: {Name: desire.NeedCompanionship, Value: 65.0},
			},
			expectedIntent: IntentSeekingConnection,
		},
		{
			name: "TaskCompletion > 60",
			needs: map[string]*desire.Need{
				desire.NeedTaskCompletion: {Name: desire.NeedTaskCompletion, Value: 65.0},
			},
			expectedIntent: IntentFocusedOnGoal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &desire.DesireProfile{
				Needs: tt.needs,
			}
			intent, _ := s.determineIntent(profile)
			assert.Equal(t, tt.expectedIntent, intent)
		})
	}
}
