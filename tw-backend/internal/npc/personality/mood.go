package personality

// Mood Trigger Events
const (
	TriggerPositive    = "positive_event"
	TriggerThreat      = "threat"
	TriggerBetrayal    = "betrayal"
	TriggerLoss        = "loss"
	TriggerAchievement = "achievement"
)

// UpdateMood updates the current mood based on time and events
// Returns the new mood (or the updated current mood)
func UpdateMood(p *Personality, currentMood *Mood, timeDeltaHours float64, events []string) *Mood {
	// 1. Check for new triggers (highest priority last)
	var newMoodType string

	for _, event := range events {
		switch event {
		case TriggerPositive:
			newMoodType = MoodCheerful
		case TriggerThreat:
			newMoodType = MoodAnxious
		case TriggerBetrayal:
			newMoodType = MoodAngry
		case TriggerLoss:
			newMoodType = MoodMelancholy
		case TriggerAchievement:
			newMoodType = MoodExcited
		}
	}

	// If a new mood is triggered, create it
	if newMoodType != "" {
		// Calculate Duration
		// Base Duration varies by type
		baseDuration := 1.0
		switch newMoodType {
		case MoodCheerful:
			baseDuration = 1.0 // 30m - 2h (avg 1.25?) Let's use 1.0
		case MoodAnxious:
			baseDuration = 2.0 // 1-3h
		case MoodAngry:
			baseDuration = 4.0 // 2-6h
		case MoodMelancholy:
			baseDuration = 8.0 // 4-12h
		case MoodExcited:
			baseDuration = 2.5 // 1-4h
		}

		// Neuroticism Modifier: Duration * (1 + Neuroticism/100)
		// High Neuroticism = moods last longer
		duration := baseDuration * (1.0 + p.Neuroticism.Value/100.0)

		return NewMood(newMoodType, duration)
	}

	// 2. If no new trigger, update existing mood
	if currentMood != nil && currentMood.Type != MoodCalm {
		currentMood.Duration -= timeDeltaHours
		if currentMood.Duration <= 0 {
			// Mood expired, return to Calm
			return NewMood(MoodCalm, 0)
		}
		return currentMood
	}

	// Default to Calm if nothing else
	if currentMood == nil {
		return NewMood(MoodCalm, 0)
	}

	return currentMood
}
