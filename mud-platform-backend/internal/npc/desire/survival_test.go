package desire

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateSurvivalNeeds(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	ctx := Context{}

	// 1 Hour Pass
	UpdateSurvivalNeeds(dp, 1.0, ctx)

	assert.Equal(t, 1.0, dp.Needs[NeedHunger].Value)
	assert.Equal(t, 1.5, dp.Needs[NeedThirst].Value)
	assert.Equal(t, 1.0, dp.Needs[NeedSleep].Value)
}

func TestUpdateSurvivalNeeds_Sleeping(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	dp.Needs[NeedSleep].Value = 50.0
	ctx := Context{IsSleeping: true}

	// 1 Hour Sleep -> -10
	UpdateSurvivalNeeds(dp, 1.0, ctx)

	assert.Equal(t, 40.0, dp.Needs[NeedSleep].Value)
}

func TestApplySurvivalPenalties(t *testing.T) {
	dp := NewDesireProfile(uuid.New())

	// Starving
	dp.Needs[NeedHunger].Value = 90.0

	mods := ApplySurvivalPenalties(dp)
	assert.Equal(t, -5, mods.Might)

	// Exhausted + Starving
	dp.Needs[NeedSleep].Value = 95.0
	mods = ApplySurvivalPenalties(dp)
	// -5 (Hunger) + -12 (Sleep) = -17
	assert.Equal(t, -17, mods.Might)
}
