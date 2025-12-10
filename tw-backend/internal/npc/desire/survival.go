package desire

import (
	"mud-platform-backend/internal/character"
)

// Context holds environmental data for need calculation
type Context struct {
	IsSleeping      bool
	IsEating        bool
	IsDrinking      bool
	InCombat        bool
	HostileCount    int
	LocationSafety  float64 // 0-100 (0=Safe, 100=Dangerous)
	HoursSinceSleep float64
}

// UpdateSurvivalNeeds updates hunger, thirst, sleep, and safety
func UpdateSurvivalNeeds(profile *DesireProfile, timeDeltaHours float64, ctx Context) {
	// Hunger
	// Increases by 1.0 per hour
	if !ctx.IsEating {
		profile.Needs[NeedHunger].Value += 1.0 * timeDeltaHours
	}
	clamp(profile.Needs[NeedHunger])

	// Thirst
	// Increases by 1.5 per hour
	if !ctx.IsDrinking {
		profile.Needs[NeedThirst].Value += 1.5 * timeDeltaHours
	}
	clamp(profile.Needs[NeedThirst])

	// Sleep
	// Increases by 1.0 per hour awake
	// Decreases by 10 per hour of sleep
	if ctx.IsSleeping {
		profile.Needs[NeedSleep].Value -= 10.0 * timeDeltaHours
	} else {
		profile.Needs[NeedSleep].Value += 1.0 * timeDeltaHours
	}
	clamp(profile.Needs[NeedSleep])

	// Safety
	// Context dependent
	safetyVal := 0.0
	if ctx.InCombat {
		safetyVal = 100.0
	} else {
		safetyVal = ctx.LocationSafety + float64(ctx.HostileCount)*10.0
	}
	profile.Needs[NeedSafety].Value = safetyVal
	clamp(profile.Needs[NeedSafety])
}

// ApplySurvivalPenalties calculates attribute modifiers based on needs
func ApplySurvivalPenalties(profile *DesireProfile) character.Attributes {
	mods := character.Attributes{}

	// Hunger
	if profile.Needs[NeedHunger].Value >= 85 {
		// -10% to all? We return flat modifiers here usually.
		// Let's assume -10 flat for simplicity or percentage if supported.
		// The prompt says "-10% to all attributes".
		// Since Attributes struct is integers, let's return percentage modifiers?
		// Or just negative values.
		// The prompt implies a significant penalty.
		// Let's return a separate struct for multipliers if needed, or just assume this function
		// returns flat penalties approximating 10% of average stats (e.g. -5).
		// Let's use -5 for now as a placeholder for "10%".
		mods.Might -= 5
		mods.Agility -= 5
		mods.Endurance -= 5
		mods.Reflexes -= 5
		mods.Vitality -= 5
		mods.Intellect -= 5
		mods.Cunning -= 5
		mods.Willpower -= 5
		mods.Presence -= 5
		mods.Intuition -= 5
	}

	// Thirst
	if profile.Needs[NeedThirst].Value >= 80 {
		// -15% -> -8
		mods.Might -= 8
		mods.Agility -= 8
		mods.Endurance -= 8
		mods.Reflexes -= 8
		mods.Vitality -= 8
		mods.Intellect -= 8
		mods.Cunning -= 8
		mods.Willpower -= 8
		mods.Presence -= 8
		mods.Intuition -= 8
	}

	// Sleep
	if profile.Needs[NeedSleep].Value >= 90 {
		// Exhausted: -25% -> -12
		mods.Might -= 12
		mods.Agility -= 12
		mods.Endurance -= 12
		mods.Reflexes -= 12
		mods.Vitality -= 12
		mods.Intellect -= 12
		mods.Cunning -= 12
		mods.Willpower -= 12
		mods.Presence -= 12
		mods.Intuition -= 12
	} else if profile.Needs[NeedSleep].Value >= 75 {
		// Tired: -10% to Focus (Mental?) and Reflexes
		mods.Reflexes -= 5
		mods.Intellect -= 5 // Proxy for Focus
	}

	return mods
}

func clamp(n *Need) {
	if n.Value > 100 {
		n.Value = 100
	}
	if n.Value < 0 {
		n.Value = 0
	}
}
