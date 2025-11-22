package desire

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePriorityScore(t *testing.T) {
	// Survival Need (Hunger 50)
	need := &Need{Name: NeedHunger, Value: 50.0, Tier: TierSurvival}
	traits := PersonalityTraits{Neuroticism: 0.5} // Weight 1.0

	// Score = 50 * 1.0 * 4.0 = 200
	score := CalculatePriorityScore(need, traits)
	assert.Equal(t, 200.0, score)

	// Social Need (Companionship 50)
	needSocial := &Need{Name: NeedCompanionship, Value: 50.0, Tier: TierSocial}
	traitsExt := PersonalityTraits{Extraversion: 1.0} // Weight 1.5

	// Score = 50 * 1.5 * 2.0 = 150
	scoreSocial := CalculatePriorityScore(needSocial, traitsExt)
	assert.Equal(t, 150.0, scoreSocial)
}

func TestShouldInterrupt(t *testing.T) {
	current := 100.0
	candidate := 150.0

	// Not 2x -> False
	assert.False(t, ShouldInterrupt(current, candidate, false))

	candidate = 201.0
	// > 2x -> True
	assert.True(t, ShouldInterrupt(current, candidate, false))

	// Critical -> True
	assert.True(t, ShouldInterrupt(current, 110.0, true))
}

func TestCalculatePriorities_Sorting(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	dp.Needs[NeedHunger].Value = 10.0        // Score ~40
	dp.Needs[NeedCompanionship].Value = 90.0 // Score ~180 (90 * 1.0 * 2.0)

	traits := PersonalityTraits{Extraversion: 0.5, Neuroticism: 0.5}

	priorities := CalculatePriorities(dp, traits)

	assert.Equal(t, NeedCompanionship, priorities[0].NeedName)
	assert.Equal(t, NeedHunger, priorities[1].NeedName) // Hunger is lower despite Tier 1 because value is low
}
