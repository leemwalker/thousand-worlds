package desire

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSimulation_24HourCycle(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	traits := PersonalityTraits{
		Extraversion:      0.5,
		Neuroticism:       0.5,
		Conscientiousness: 0.5,
		Openness:          0.5,
	}

	// Simulate 24 hours
	// 0-8: Sleep
	// 8-12: Work
	// 12-13: Eat
	// 13-17: Work
	// 17-20: Social
	// 20-24: Leisure

	for hour := 0; hour < 24; hour++ {
		ctx := Context{}
		socialCtx := SocialContext{}
		achCtx := AchievementContext{}
		pleasureCtx := PleasureContext{}

		// Context Logic
		if hour < 8 {
			ctx.IsSleeping = true
		} else {
			ctx.IsSleeping = false
			ctx.HoursSinceSleep = float64(hour - 8)
		}

		if hour == 12 || hour == 19 {
			ctx.IsEating = true
		}

		if hour >= 8 && hour < 17 && hour != 12 {
			achCtx.ActiveTasks = 3
		}

		if hour >= 17 && hour < 20 {
			socialCtx.WithFriends = true
			socialCtx.Talking = true
		}

		if hour >= 20 {
			pleasureCtx.BoredHours = 1.0
		}

		// Update Needs
		UpdateSurvivalNeeds(dp, 1.0, ctx)
		UpdateSocialNeeds(dp, 1.0, traits, socialCtx)
		UpdateAchievementNeeds(dp, 1.0, traits, achCtx)
		UpdatePleasureNeeds(dp, 1.0, traits, pleasureCtx)

		// Check Priorities
		priorities := CalculatePriorities(dp, traits)
		topNeed := priorities[0]

		// Debug Output
		// t.Logf("Hour %d: Hunger=%.2f, Sleep=%.2f, Top=%s", hour, dp.Needs[NeedHunger].Value, dp.Needs[NeedSleep].Value, topNeed.NeedName)

		// Assertions for key times
		if hour == 7 {
			// Waking up (Hour 7 is 7:00-8:00, sleep ends at 8:00)
			// Hunger accumulates 1.0/hr. 7 hours = 7.0?
			// Wait, simulation loop starts at 0.
			// Hour 0: Sleep. Hunger +1.
			// Hour 7: Sleep. Hunger +1. Total 8.0 hunger?
			// Ah, initial value is 0.
			// 8 hours sleep -> 8.0 hunger.
			// Assertion > 50 is wrong if rate is 1.0/hr.
			// Needs scale 0-100. 1.0/hr means 100 hours to starve.
			// Prompt: "Hunger... Increases by 1.0 per hour... At 70+: hungry"
			// So 3 days to get hungry?
			// Prompt says: "Increases by 1.0 per hour".
			// Maybe initial values should be randomized or higher?
			// Or test expectations are wrong.
			// Let's check logic.
			// If rate is 1.0, then 8 hours is just 8.0.
			// Assertion expects > 50.
			// I should adjust the test expectation to match the spec (1.0/hr).
			// Or maybe the spec implies faster rate?
			// "At 70+: hungry". 70 hours = 3 days. Realistic for starvation.
			// But "Hungry" status usually means "Time to eat".
			// Maybe 1.0 is too slow for game time?
			// Prompt: "Increases by 1.0 per hour (game time)".
			// If 1 hour game time = X real time, it matters.
			// But here we simulate game hours.
			// So 3 days without food to be "Hungry".
			// That seems fine for survival mechanics.
			// But for daily cycle test, we won't see high hunger.
			// Unless we start with some hunger.

			// Let's adjust test to expect realistic values.
			// At hour 7 (morning), hunger should be > 0.
			assert.True(t, dp.Needs[NeedHunger].Value > 5, "Hunger should have increased during sleep")

			// Sleep: Decreases by 10/hr.
			// If started at 0, it stays at 0 (clamped).
			// So sleep should be 0.
			assert.True(t, dp.Needs[NeedSleep].Value < 5, "Sleep should be low after rest")
		}

		if hour == 16 {
			// End of work: Achievement needs might be high if tasks pending, or low if done.
			// Here we simulated active tasks, so TaskCompletion should be high.
			// But we didn't "complete" them in context logic (TaskCompleted=false).
			assert.True(t, dp.Needs[NeedTaskCompletion].Value > 20, "Task need should accumulate")
		}

		if hour == 23 {
			// End of day (Hour 23 is 23:00-00:00)
			// Awake from 8 to 23 = 15 hours.
			// Sleep increases 1.0/hr.
			// Total 15.0 sleep.
			// Assertion > 50 is too high.
			assert.True(t, dp.Needs[NeedSleep].Value > 10, "Sleep should be rising at night")
		}

		// Verify Action Generation
		action := GetBestAction(dp.Needs[topNeed.NeedName], dp.NPCID)
		assert.NotEmpty(t, action.Name)
	}
}
