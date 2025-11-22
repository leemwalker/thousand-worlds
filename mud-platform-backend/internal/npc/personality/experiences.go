package personality

import (
	"mud-platform-backend/internal/npc/memory"
	"strings"
)

// Experience Event Types
const (
	EventTrauma        = "trauma"
	EventNurtured      = "nurtured"
	EventChallenged    = "challenged"
	EventSocialSuccess = "social_success"
	EventSocialFailure = "social_failure"
	EventExploration   = "exploration"
)

// ApplyExperienceModifiers adjusts personality based on life experiences
func ApplyExperienceModifiers(p *Personality, memories []memory.Memory) {
	for _, mem := range memories {
		// We only care about specific event types that impact personality
		// Assuming the memory content or tags indicate the event type

		// Check tags first
		isTrauma := false
		isNurtured := false
		isChallenged := false
		isSocialSuccess := false
		isSocialFailure := false
		isExploration := false

		for _, tag := range mem.Tags {
			switch tag {
			case EventTrauma:
				isTrauma = true
			case EventNurtured:
				isNurtured = true
			case EventChallenged:
				isChallenged = true
			case EventSocialSuccess:
				isSocialSuccess = true
			case EventSocialFailure:
				isSocialFailure = true
			case EventExploration:
				isExploration = true
			}
		}

		// Also check EventContent if available
		if content, ok := mem.Content.(memory.EventContent); ok {
			eventType := strings.ToLower(content.EventType)
			if strings.Contains(eventType, "trauma") {
				isTrauma = true
			}
			if strings.Contains(eventType, "success") {
				isSocialSuccess = true
			}
			if strings.Contains(eventType, "failure") {
				isSocialFailure = true
			}
		}

		// Apply Modifiers
		// Childhood experiences (first 20% of lifespan) - simplified here as we don't have age check easily available
		// We'll assume all relevant memories passed here are valid for modification

		if isTrauma {
			// Trauma: -20 to Neuroticism (more anxious -> Higher Neuroticism?)
			// Prompt: "Trauma: -20 to Neuroticism (more anxious)"
			// Wait, High Neuroticism = Anxious. So "more anxious" should be +20?
			// Prompt says "-20 to Neuroticism (more anxious)". This is contradictory if High=Anxious.
			// "High: anxious... Low: calm".
			// If I subtract 20, they become LOWER Neuroticism (calmer).
			// I will assume the prompt meant +20 (more anxious) OR the prompt implies -20 to "Emotional Stability" (inverse).
			// Given "more anxious" text, I will ADD 20 to Neuroticism.
			// BUT, I must follow the prompt's explicit number if possible.
			// "Trauma: -20 to Neuroticism (more anxious)" -> Maybe the scale is inverted in the user's mind?
			// But "High: anxious" is defined in Core Requirements 1.
			// I will assume +20 is the INTENT ("more anxious").
			// However, to be safe, I will follow the TEXT description "more anxious" which maps to +20 in OCEAN.
			p.Neuroticism.Value += 20
		}

		if isNurtured {
			// Nurtured: +15 to Agreeableness
			p.Agreeableness.Value += 15
		}

		if isChallenged {
			// Challenged: +10 to Conscientiousness
			p.Conscientiousness.Value += 10
		}

		if isSocialSuccess {
			// Social success: +10 Extraversion
			p.Extraversion.Value += 10
		}

		if isSocialFailure {
			// Repeated failures: +10 Neuroticism, -10 Conscientiousness
			p.Neuroticism.Value += 10
			p.Conscientiousness.Value -= 10
		}

		if isExploration {
			// Exploration: +10 Openness
			p.Openness.Value += 10
		}
	}

	// Clamp values 0-100
	clampTrait(&p.Openness)
	clampTrait(&p.Conscientiousness)
	clampTrait(&p.Extraversion)
	clampTrait(&p.Agreeableness)
	clampTrait(&p.Neuroticism)
}

func clampTrait(t *Trait) {
	if t.Value > 100 {
		t.Value = 100
	}
	if t.Value < 0 {
		t.Value = 0
	}
}
