package personality

import (
	"testing"
)

func TestUpdateMood(t *testing.T) {
	p := NewPersonality()
	p.Neuroticism.Value = 50.0 // Average Neuroticism

	// 1. Trigger New Mood (Positive -> Cheerful)
	events := []string{TriggerPositive}
	mood := UpdateMood(p, nil, 0, events)

	if mood.Type != MoodCheerful {
		t.Errorf("Expected Cheerful mood, got %s", mood.Type)
	}

	// Duration: Base 1.0 * (1 + 0.5) = 1.5 hours
	expectedDuration := 1.5
	if mood.Duration != expectedDuration {
		t.Errorf("Expected duration %f, got %f", expectedDuration, mood.Duration)
	}

	// 2. Decay Mood
	// Advance 1.0 hour
	mood = UpdateMood(p, mood, 1.0, nil)
	if mood.Duration != 0.5 {
		t.Errorf("Expected remaining duration 0.5, got %f", mood.Duration)
	}

	// 3. Expire Mood
	// Advance 1.0 hour (total 2.0)
	mood = UpdateMood(p, mood, 1.0, nil)
	if mood.Type != MoodCalm {
		t.Errorf("Expected Calm mood after expiration, got %s", mood.Type)
	}

	// 4. High Neuroticism Duration
	pHighN := NewPersonality()
	pHighN.Neuroticism.Value = 100.0 // Max

	moodHighN := UpdateMood(pHighN, nil, 0, events)
	// Duration: Base 1.0 * (1 + 1.0) = 2.0 hours
	if moodHighN.Duration != 2.0 {
		t.Errorf("Expected High Neuroticism duration 2.0, got %f", moodHighN.Duration)
	}
}
