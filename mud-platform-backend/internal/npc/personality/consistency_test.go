package personality

import (
	"testing"
)

func TestPersonalityConsistency(t *testing.T) {
	// 1. Adventurer (High Openness) vs Novelty
	// Should choose Novel option > 70% of time vs Familiar
	adventurer := GetArchetype("Adventurer")

	novelCount := 0
	trials := 100

	optNovel := DecisionOption{Name: "Novel", Tags: []string{TagNovel}}
	optFamiliar := DecisionOption{Name: "Familiar", Tags: []string{TagFamiliar}}

	for i := 0; i < trials; i++ {
		scoreNovel := CalculateDecisionScore(optNovel, adventurer, nil)
		scoreFamiliar := CalculateDecisionScore(optFamiliar, adventurer, nil)

		if scoreNovel > scoreFamiliar {
			novelCount++
		}
	}

	if novelCount < 70 {
		t.Errorf("Adventurer chose Novel only %d/%d times (expected > 70)", novelCount, trials)
	}

	// 2. Leader (High Extraversion) vs Social
	// Should choose Social option > 70% of time vs Solitary
	leader := GetArchetype("Leader")

	socialCount := 0

	optSocial := DecisionOption{Name: "Social", Tags: []string{TagSocial}}
	optSolitary := DecisionOption{Name: "Solitary", Tags: []string{TagSolitary}}

	for i := 0; i < trials; i++ {
		scoreSocial := CalculateDecisionScore(optSocial, leader, nil)
		scoreSolitary := CalculateDecisionScore(optSolitary, leader, nil)

		if scoreSocial > scoreSolitary {
			socialCount++
		}
	}

	if socialCount < 70 {
		t.Errorf("Leader chose Social only %d/%d times (expected > 70)", socialCount, trials)
	}

	// 3. Hermit (High Neuroticism) vs Safety
	// Should choose Safe option > 70% of time vs Risky
	hermit := GetArchetype("Hermit")

	safeCount := 0

	optSafe := DecisionOption{Name: "Safe", Tags: []string{TagSafe}}
	optRisky := DecisionOption{Name: "Risky", Tags: []string{TagRisky}}

	for i := 0; i < trials; i++ {
		scoreSafe := CalculateDecisionScore(optSafe, hermit, nil)
		scoreRisky := CalculateDecisionScore(optRisky, hermit, nil)

		if scoreSafe > scoreRisky {
			safeCount++
		}
	}

	if safeCount < 70 {
		t.Errorf("Hermit chose Safe only %d/%d times (expected > 70)", safeCount, trials)
	}
}
