package interaction

import (
	"tw-backend/internal/npc/personality"
	"testing"
)

func TestShouldInitiateConversation(t *testing.T) {
	// Setup Context
	ctx := InteractionContext{
		Distance:           3.0, // Within 5m
		InitiatorAffection: 50,  // Neutral/Positive
	}

	// Setup Personality (High Extraversion)
	p := personality.NewPersonality()
	p.Extraversion.Value = 100.0

	// Need
	companionshipNeed := 80.0 // High need

	// Run multiple trials since it's probabilistic
	initiations := 0
	trials := 1000

	for i := 0; i < trials; i++ {
		if ShouldInitiateConversation(ctx, p, companionshipNeed) {
			initiations++
		}
	}

	// Expected Probability:
	// Base 0.1 * Extraversion (2.0) * Relationship (1.25) * Need (1.8) = 0.45 (45%)
	// Allow margin of error
	if initiations < 350 || initiations > 550 {
		t.Errorf("Expected ~450 initiations, got %d", initiations)
	}

	// Test Proximity Fail
	ctx.Distance = 10.0
	if ShouldInitiateConversation(ctx, p, companionshipNeed) {
		t.Error("Should not initiate when distance > 5m")
	}
}
