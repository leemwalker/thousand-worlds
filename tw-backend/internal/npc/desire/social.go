package desire

import (
	"mud-platform-backend/internal/character"
)

// PersonalityTraits holds relevant traits for social needs
type PersonalityTraits struct {
	Extraversion      float64 // 0-1 (Low to High)
	Neuroticism       float64
	Conscientiousness float64
	Openness          float64
}

// SocialContext holds social environment data
type SocialContext struct {
	Alone              bool
	WithFriends        bool
	Talking            bool
	RelationshipAction bool // Did a positive interaction occur?
}

// UpdateSocialNeeds updates companionship, conversation, and affection
func UpdateSocialNeeds(profile *DesireProfile, timeDeltaHours float64, traits PersonalityTraits, ctx SocialContext) {
	// Companionship
	// Increases when alone, decreases when with friends
	if ctx.Alone {
		rate := 0.5
		// Extraverts get lonely faster (+50%)
		// Introverts slower (-50%)
		// Base is 0.5. Extraversion 1.0 -> 0.75. Extraversion 0.0 -> 0.25.
		// Formula: rate * (0.5 + 0.5 * Extraversion) ?
		// Prompt: "Extraverted: +50% increase rate... Introverted: -50% increase rate"
		// Let's assume Extraversion 0.5 is baseline? Or 0.0 is introverted?
		// Let's assume 0.0 = Introvert (-50%), 1.0 = Extravert (+50%).
		// Rate = 0.5 * (0.5 + traits.Extraversion) ->
		// If 0.0: 0.5 * 0.5 = 0.25.
		// If 1.0: 0.5 * 1.5 = 0.75.
		// If 0.5: 0.5 * 1.0 = 0.5. Matches.
		rate = rate * (0.5 + traits.Extraversion)
		profile.Needs[NeedCompanionship].Value += rate * timeDeltaHours
	} else if ctx.WithFriends {
		profile.Needs[NeedCompanionship].Value -= 5.0 * timeDeltaHours
	}
	clamp(profile.Needs[NeedCompanionship])

	// Conversation
	// Increases when not talking, decreases when talking
	if !ctx.Talking {
		rate := 1.0
		// Extraverts need talk more (+100% rate?)
		// Prompt: "Extraverted: +100% increase rate"
		// Rate = 1.0 * (1.0 + traits.Extraversion) -> 1.0 to 2.0?
		// Or relative to baseline?
		// Let's assume baseline is for average person.
		// Rate = 1.0 * (0.5 + traits.Extraversion)? No.
		// Let's just add Extraversion factor.
		// Rate = 1.0 + (traits.Extraversion * 1.0) -> 1.0 to 2.0.
		rate = 1.0 + traits.Extraversion
		profile.Needs[NeedConversation].Value += rate * timeDeltaHours
	} else {
		// Decrease per conversation event usually, but here per hour of talking?
		// Prompt: "Decreases by 20 per meaningful conversation".
		// If ctx.Talking is true, assume continuous conversation?
		// Let's assume -20 per hour for now, or handle discrete events elsewhere.
		profile.Needs[NeedConversation].Value -= 20.0 * timeDeltaHours
	}
	clamp(profile.Needs[NeedConversation])

	// Affection
	// Slow increase
	profile.Needs[NeedAffection].Value += 0.2 * timeDeltaHours
	if ctx.RelationshipAction {
		profile.Needs[NeedAffection].Value -= 10.0 // Discrete event
	}
	clamp(profile.Needs[NeedAffection])
}

// ApplySocialPenalties
func ApplySocialPenalties(profile *DesireProfile) character.Attributes {
	mods := character.Attributes{}

	// Companionship
	if profile.Needs[NeedCompanionship].Value >= 80 {
		// Lonely: -10 Presence
		mods.Presence -= 10
	}

	return mods
}
