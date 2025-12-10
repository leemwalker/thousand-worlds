package prompt

import (
	"tw-backend/internal/character"
	"tw-backend/internal/npc/personality"
	"tw-backend/internal/npc/relationship"
	"strings"
	"testing"
)

func TestPromptBuilder_Build(t *testing.T) {
	// Setup
	p := personality.NewPersonality()
	mood := personality.NewMood(personality.MoodCalm, 1.0)
	attr := character.Attributes{Might: 10, Agility: 10, Intellect: 10}
	rel := &relationship.Relationship{
		CurrentAffinity: relationship.Affinity{Affection: 50, Trust: 50},
	}

	builder := NewPromptBuilder().
		WithNPC("Gandalf", 2000, "Maiar", "Wizard", p, mood, "Smoke pipe", 10.0, attr).
		WithContext("The Shire", "Morning", "Sunny", "Frodo").
		WithSpeaker("Frodo", rel).
		WithConversation("Adventure", "Hello Gandalf!")

	// Build
	prompt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify Content
	checks := []string{
		"You are Gandalf",
		"PERSONALITY:",
		"Openness:",
		"Mood: calm",
		"Top Desire: Smoke pipe",
		"Location: The Shire",
		"RELATIONSHIP WITH Frodo:",
		"Affection: 50/100",
		"Frodo says: \"Hello Gandalf!\"",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("Prompt missing expected content: %s", check)
		}
	}
}

func TestPromptBuilder_WithDrift(t *testing.T) {
	p := personality.NewPersonality()
	mood := personality.NewMood(personality.MoodCalm, 1.0)
	attr := character.Attributes{Might: 10}
	rel := &relationship.Relationship{CurrentAffinity: relationship.Affinity{Affection: 50}}

	drift := &relationship.DriftMetrics{DriftLevel: "Severe"}
	base := relationship.BehavioralProfile{Aggression: 0.1}
	curr := relationship.BehavioralProfile{Aggression: 0.9}

	builder := NewPromptBuilder().
		WithNPC("Bilbo", 111, "Hobbit", "Burglar", p, mood, "Rest", 5.0, attr).
		WithSpeaker("Gandalf", rel).
		WithDrift(drift, base, curr)

	prompt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !strings.Contains(prompt, "PERSONALITY DRIFT DETECTED") {
		t.Error("Prompt missing drift section")
	}
	if !strings.Contains(prompt, "Drift Level: Severe") {
		t.Error("Prompt missing drift level")
	}
	if !strings.Contains(prompt, "alarmed by this drastic personality change") {
		t.Error("Prompt missing severe drift instruction")
	}
}
