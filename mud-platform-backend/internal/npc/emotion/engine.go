package emotion

import (
	"math"
)

// EmotionEngine analyzes events to produce emotional profiles
type EmotionEngine struct{}

func NewEmotionEngine() *EmotionEngine {
	return &EmotionEngine{}
}

// AnalyzeEvent determines the emotional profile and aggregate weight of an event
func (e *EmotionEngine) AnalyzeEvent(context EventContext, personality PersonalityTraits) (EmotionProfile, float64) {
	profile := make(EmotionProfile)

	// Base emotions from event type
	if context.IsThreat {
		profile[Fear] = 0.95
		profile[Anger] = 0.5
	} else if context.IsBetrayal {
		profile[Anger] = 0.8
		profile[Sadness] = 0.6
	} else if context.IsDeath {
		profile[Sadness] = 0.9
		profile[Anger] = 0.3
	} else if context.IsFirstMeeting {
		profile[Surprise] = 0.5
		profile[Joy] = 0.2 // Mild positive bias? Or neutral? Prompt says "First meeting: 0.5" weight.
		// Let's assume mild curiosity/surprise.
	} else if context.GiftValue > 0 {
		// "0.3 + (giftValue / wealthLevel) * 0.4" for weight.
		// Joy = 0.6, Surprise = 0.3 (from prompt example)
		// Let's scale Joy by value.
		ratio := 0.0
		if context.WealthLevel > 0 {
			ratio = context.GiftValue / context.WealthLevel
		}
		if ratio > 1.0 {
			ratio = 1.0
		}
		profile[Joy] = 0.3 + (ratio * 0.7)
		profile[Surprise] = 0.3
	} else if context.DamageTaken > 0 {
		// Combat
		ratio := 0.0
		if context.MaxHP > 0 {
			ratio = context.DamageTaken / context.MaxHP
		}
		if ratio > 1.0 {
			ratio = 1.0
		}
		profile[Fear] = 0.5 + (ratio * 0.5)
		profile[Anger] = 0.5
	} else {
		// Casual/Default
		profile[Joy] = 0.1
	}

	// Apply Personality Modifiers
	// Neurotic: Fear/Sadness +20%
	if personality.Neuroticism > 0 {
		if v, ok := profile[Fear]; ok {
			profile[Fear] = math.Min(1.0, v*1.2)
		}
		if v, ok := profile[Sadness]; ok {
			profile[Sadness] = math.Min(1.0, v*1.2)
		}
	}
	// Aggressive: Anger +30%
	if personality.Aggression > 0 {
		if v, ok := profile[Anger]; ok {
			profile[Anger] = math.Min(1.0, v*1.3)
		}
	}
	// Optimistic: Joy +20%, Sadness -20%
	if personality.Optimism > 0 {
		if v, ok := profile[Joy]; ok {
			profile[Joy] = math.Min(1.0, v*1.2)
		}
		if v, ok := profile[Sadness]; ok {
			profile[Sadness] = v * 0.8
		}
	}

	// Calculate Complex Emotions
	e.CalculateComplexEmotions(profile)

	// Calculate Aggregate Weight (Max intensity? Or specific formulas from prompt?)
	// Prompt: "Calculated from event context: Combat: 0.7 + ..., First meeting: 0.5, etc."
	// These seem to be the *EmotionalWeight* of the memory, which is a single scalar.
	// The profile is the breakdown.
	// Let's calculate the scalar weight based on the prompt's formulas, using the context.

	weight := 0.0
	if context.IsThreat {
		weight = 0.95
	} else if context.IsBetrayal {
		weight = 0.9
	} else if context.IsDeath {
		weight = 0.8
	} else if context.IsFirstMeeting {
		weight = 0.5
	} else if context.GiftValue > 0 {
		ratio := 0.0
		if context.WealthLevel > 0 {
			ratio = context.GiftValue / context.WealthLevel
		}
		weight = 0.3 + (ratio * 0.4)
	} else if context.DamageTaken > 0 {
		ratio := 0.0
		if context.MaxHP > 0 {
			ratio = context.DamageTaken / context.MaxHP
		}
		weight = 0.7 + (ratio * 0.3)
	} else {
		weight = 0.1 // Casual
	}

	if weight > 1.0 {
		weight = 1.0
	}

	return profile, weight
}

// CalculateComplexEmotions adds derived emotions to the profile
func (e *EmotionEngine) CalculateComplexEmotions(profile EmotionProfile) {
	// Anticipation = Joy * 0.5 + Surprise * 0.5
	if j, ok := profile[Joy]; ok {
		if s, ok := profile[Surprise]; ok {
			profile[Anticipation] = (j * 0.5) + (s * 0.5)
		}
	}

	// Contempt = Anger * 0.6 + Disgust * 0.4
	if a, ok := profile[Anger]; ok {
		if d, ok := profile[Disgust]; ok {
			profile[Contempt] = (a * 0.6) + (d * 0.4)
		}
	}

	// Anxiety = Fear * 0.7 + Surprise * 0.3
	if f, ok := profile[Fear]; ok {
		if s, ok := profile[Surprise]; ok {
			profile[Anxiety] = (f * 0.7) + (s * 0.3)
		}
	}
}
