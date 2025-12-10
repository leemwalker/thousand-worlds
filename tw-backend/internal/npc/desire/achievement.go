package desire

// AchievementContext holds data for achievement needs
type AchievementContext struct {
	ActiveTasks      int
	TaskCompleted    bool
	SkillXPGained    bool
	WealthPercentile float64 // 0-100
}

// UpdateAchievementNeeds updates task completion, skill improvement, and resource acquisition
func UpdateAchievementNeeds(profile *DesireProfile, timeDeltaHours float64, traits PersonalityTraits, ctx AchievementContext) {
	// Task Completion
	// Increases while tasks are pending
	// Each active task adds 10 to need (instant? or rate?)
	// Prompt: "Increases while tasks are pending... Each active task adds 10 to need"
	// This implies a static value based on tasks? Or rate?
	// "Decreases by 30 when task completed" implies dynamic.
	// Let's assume rate = ActiveTasks * 1.0 per hour?
	// Or maybe the need IS the urgency of tasks.
	// Let's try: Rate = ActiveTasks * 2.0 per hour.
	// Plus Conscientiousness multiplier.

	rate := float64(ctx.ActiveTasks) * 2.0
	// Conscientiousness multiplies urgency: needValue * (1 + conscientiousness)
	// We apply this multiplier when calculating Priority, not raw need value usually.
	// But prompt says "Conscientiousness personality trait multiplies urgency".
	// Let's apply it to the rate here for accumulation.
	rate = rate * (1.0 + traits.Conscientiousness)

	profile.Needs[NeedTaskCompletion].Value += rate * timeDeltaHours

	if ctx.TaskCompleted {
		profile.Needs[NeedTaskCompletion].Value -= 30.0
	}
	clamp(profile.Needs[NeedTaskCompletion])

	// Skill Improvement
	// Increases by 0.3 per hour for high Openness
	// Prompt: "Increases by 0.3 per hour for NPCs with high Openness"
	// Let's scale by Openness.
	skillRate := 0.3 * traits.Openness
	profile.Needs[NeedSkillImprovement].Value += skillRate * timeDeltaHours

	if ctx.SkillXPGained {
		profile.Needs[NeedSkillImprovement].Value -= 10.0 // Arbitrary reduction for practice
	}
	clamp(profile.Needs[NeedSkillImprovement])

	// Resource Acquisition
	// Based on wealth comparison
	// Wealth percentile < 50%: need increases
	if ctx.WealthPercentile < 50.0 {
		// Lower percentile = Higher need increase
		// Rate = (50 - percentile) * 0.1 per hour?
		wealthRate := (50.0 - ctx.WealthPercentile) * 0.05
		profile.Needs[NeedResourceAcquisition].Value += wealthRate * timeDeltaHours
	}
	clamp(profile.Needs[NeedResourceAcquisition])
}
