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

func TestCalculateDecisionScore_Conscientiousness(t *testing.T) {
	// Scholar has high conscientiousness (C=90)
	scholar := GetArchetype("Scholar")

	// Planned option should score high
	plannedOpt := DecisionOption{Tags: []string{TagPlanned}}
	score := CalculateDecisionScore(plannedOpt, scholar, nil)
	if score < 25 {
		t.Errorf("Expected score >= 25 for planned option with high C, got %f", score)
	}

	// Impulsive option should score low (negative modifier)
	impulsiveOpt := DecisionOption{Tags: []string{TagImpulsive}}
	score2 := CalculateDecisionScore(impulsiveOpt, scholar, nil)
	// Gets -15 penalty, but +20 random can compensate somewhat
	if score2 > 20 {
		t.Errorf("Expected lower score for impulsive with high C, got %f", score2)
	}
}

func TestCalculateDecisionScore_Agreeableness(t *testing.T) {
	// Leader has high agreeableness (A=70)
	leader := GetArchetype("Leader")

	// Cooperative option should boost
	coopOpt := DecisionOption{Tags: []string{TagCooperative}}
	score := CalculateDecisionScore(coopOpt, leader, nil)
	if score < 20 {
		t.Errorf("Expected score >= 20 for cooperative with high A, got %f", score)
	}

	// Competitive option should be lower
	compOpt := DecisionOption{Tags: []string{TagCompetitive}}
	score2 := CalculateDecisionScore(compOpt, leader, nil)
	if score2 > 20 {
		t.Errorf("Expected lower score for competitive with high A, got %f", score2)
	}
}

func TestCalculateDecisionScore_LowOpenness(t *testing.T) {
	// Create personality with low openness
	p := NewPersonality()
	p.Openness.Value = 20

	// Familiar option should be preferred
	familiarOpt := DecisionOption{Tags: []string{TagFamiliar}}
	score := CalculateDecisionScore(familiarOpt, p, nil)
	if score < 20 {
		t.Errorf("Expected score >= 20 for familiar with low O, got %f", score)
	}
}

func TestCalculateDecisionScore_LowExtraversion(t *testing.T) {
	// Hermit has low extraversion (E=10)
	hermit := GetArchetype("Hermit")

	// Solitary option should boost
	solitaryOpt := DecisionOption{Tags: []string{TagSolitary}}
	score := CalculateDecisionScore(solitaryOpt, hermit, nil)
	if score < 25 {
		t.Errorf("Expected score >= 25 for solitary with low E, got %f", score)
	}
}

func TestCalculateDecisionScore_LowNeuroticism(t *testing.T) {
	// Leader has low neuroticism (N=25)
	leader := GetArchetype("Leader")

	// Risky option should boost
	riskyOpt := DecisionOption{Tags: []string{TagRisky}}
	score := CalculateDecisionScore(riskyOpt, leader, nil)
	if score < 15 {
		t.Errorf("Expected score >= 15 for risky with low N, got %f", score)
	}
}

func TestCalculateDecisionScore_MultipleTags(t *testing.T) {
	// Adventurer: O=90, E=75
	adventurer := GetArchetype("Adventurer")

	// Option with both novel and social tags
	opt := DecisionOption{Tags: []string{TagNovel, TagSocial}}
	score := CalculateDecisionScore(opt, adventurer, nil)

	// Should get +20 (novel) + 30 (social) + random = 50+
	if score < 50 {
		t.Errorf("Expected score >= 50 for novel+social with high O+E, got %f", score)
	}
}

func TestCalculateDecisionScore_MoodModifiers(t *testing.T) {
	p := NewPersonality()
	p.Extraversion.Value = 60 // Not quite high

	// Add cheerful mood (+5 Extraversion)
	mood := NewMood(MoodCheerful, 1.0)

	// With mood, should reach 65, still not >=70 for bonus
	// But let's test a case where it pushes over
	p.Extraversion.Value = 68
	socialOpt := DecisionOption{Tags: []string{TagSocial}}
	score := CalculateDecisionScore(socialOpt, p, mood)

	// 68 + 5 = 73, now >= 70, gets +30 bonus
	if score < 30 {
		t.Errorf("Expected mood to push E over threshold for bonus, got %f", score)
	}
}
