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

func TestUpdateMood_AllTriggerTypes(t *testing.T) {
	p := NewPersonality()
	p.Neuroticism.Value = 50.0

	tests := []struct {
		trigger      string
		expectedMood string
	}{
		{TriggerPositive, MoodCheerful},
		{TriggerThreat, MoodAnxious},
		{TriggerBetrayal, MoodAngry},
		{TriggerLoss, MoodMelancholy},
		{TriggerAchievement, MoodExcited},
	}

	for _, tt := range tests {
		mood := UpdateMood(p, nil, 0, []string{tt.trigger})
		if mood.Type != tt.expectedMood {
			t.Errorf("Trigger %s: expected %s, got %s", tt.trigger, tt.expectedMood, mood.Type)
		}
	}
}

func TestUpdateMood_LowNeuroticism(t *testing.T) {
	p := NewPersonality()
	p.Neuroticism.Value = 0.0 // Minimum

	mood := UpdateMood(p, nil, 0, []string{TriggerPositive})

	// Duration: Base 1.0 * (1 + 0.0) = 1.0 hours
	if mood.Duration != 1.0 {
		t.Errorf("Expected low N duration 1.0, got %f", mood.Duration)
	}
}

func TestUpdateMood_DecayToZero(t *testing.T) {
	p := NewPersonality()
	p.Neuroticism.Value = 50.0

	mood := UpdateMood(p, nil, 0, []string{TriggerPositive})
	initialDuration := mood.Duration

	// Decay by exactly the duration
	mood = UpdateMood(p, mood, initialDuration, nil)

	// Should expire to Calm
	if mood.Type != MoodCalm {
		t.Errorf("Expected Calm after exact decay, got %s", mood.Type)
	}
	if mood.Duration != 0 {
		t.Errorf("Expected 0 duration, got %f", mood.Duration)
	}
}

func TestUpdateMood_PartialDecay(t *testing.T) {
	p := NewPersonality()
	p.Neuroticism.Value = 50.0

	mood := UpdateMood(p, nil, 0, []string{TriggerPositive})
	initialDuration := mood.Duration // 1.5

	// Decay by half
	mood = UpdateMood(p, mood, 0.75, nil)

	// Should still be cheerful
	if mood.Type != MoodCheerful {
		t.Errorf("Expected Cheerful to persist, got %s", mood.Type)
	}
	if mood.Duration != initialDuration-0.75 {
		t.Errorf("Expected duration %f, got %f", initialDuration-0.75, mood.Duration)
	}
}

func TestUpdateMood_NoEvents(t *testing.T) {
	p := NewPersonality()

	// No current mood, no events
	mood := UpdateMood(p, nil, 0, nil)

	// Should default to Calm
	if mood.Type != MoodCalm {
		t.Errorf("Expected Calm with no events, got %s", mood.Type)
	}
}

func TestUpdateMood_OverrideMood(t *testing.T) {
	p := NewPersonality()
	p.Neuroticism.Value = 50.0

	// Start with cheerful
	mood := UpdateMood(p, nil, 0, []string{TriggerPositive})
	if mood.Type != MoodCheerful {
		t.Fatalf("Setup failed, expected Cheerful")
	}

	// Trigger threat while cheerful
	mood = UpdateMood(p, mood, 0, []string{TriggerThreat})

	// Should override to anxious
	if mood.Type != MoodAnxious {
		t.Errorf("Expected new trigger to override mood, got %s", mood.Type)
	}
}

func TestUpdateMood_NeuroticismEffect(t *testing.T) {
	// Test that high neuroticism extends duration significantly
	pLow := NewPersonality()
	pLow.Neuroticism.Value = 10.0

	pHigh := NewPersonality()
	pHigh.Neuroticism.Value = 90.0

	moodLow := UpdateMood(pLow, nil, 0, []string{TriggerPositive})
	moodHigh := UpdateMood(pHigh, nil, 0, []string{TriggerPositive})

	// High N should have significantly longer duration
	if moodHigh.Duration <= moodLow.Duration {
		t.Errorf("Expected high N mood (%f) to last longer than low N mood (%f)", moodHigh.Duration, moodLow.Duration)
	}
}

func TestNewMood_AllTypes(t *testing.T) {
	moodTypes := []string{
		MoodCheerful,
		MoodMelancholy,
		MoodAnxious,
		MoodAngry,
		MoodExcited,
		MoodCalm,
	}

	for _, moodType := range moodTypes {
		mood := NewMood(moodType, 1.0)
		if mood.Type != moodType {
			t.Errorf("Expected mood type %s, got %s", moodType, mood.Type)
		}
		if mood.Duration != 1.0 {
			t.Errorf("Expected duration 1.0, got %f", mood.Duration)
		}
		if mood.Modifiers == nil {
			t.Errorf("Expected modifiers to be initialized for %s", moodType)
		}
	}
}
