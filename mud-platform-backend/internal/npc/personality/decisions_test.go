package personality

import (
	"testing"
)

func TestCalculateDecisionScore(t *testing.T) {
	// Setup Archetype: Adventurer (High Openness, High Extraversion)
	p := GetArchetype("Adventurer")
	// O=90, C=40, E=75, A=60, N=30

	// Case 1: Novel Option (High Openness should boost)
	optNovel := DecisionOption{Tags: []string{TagNovel}}
	scoreNovel := CalculateDecisionScore(optNovel, p, nil)

	// Base 0 + 20 (High Openness) + Random(0-20) -> 20-40
	if scoreNovel < 20 {
		t.Errorf("Expected score >= 20 for Novel option with High Openness, got %f", scoreNovel)
	}

	// Case 2: Social Option (High Extraversion should boost)
	optSocial := DecisionOption{Tags: []string{TagSocial}}
	scoreSocial := CalculateDecisionScore(optSocial, p, nil)

	// Base 0 + 30 (High Extraversion) + Random(0-20) -> 30-50
	if scoreSocial < 30 {
		t.Errorf("Expected score >= 30 for Social option with High Extraversion, got %f", scoreSocial)
	}

	// Case 3: Mood Influence
	// Add Anxious Mood (+15 Neuroticism)
	// Adventurer N=30. +15 = 45. Still not High (>70).
	// Let's use Hermit (N=70). +15 = 85 (High).
	pHermit := GetArchetype("Hermit")
	moodAnxious := NewMood(MoodAnxious, 1.0)

	optSafe := DecisionOption{Tags: []string{TagSafe}}
	scoreSafe := CalculateDecisionScore(optSafe, pHermit, moodAnxious)

	// Hermit N=70. +15 = 85. High Neuroticism (>70) -> +30 score.
	if scoreSafe < 30 {
		t.Errorf("Expected score >= 30 for Safe option with High Neuroticism (Mood boosted), got %f", scoreSafe)
	}
}
