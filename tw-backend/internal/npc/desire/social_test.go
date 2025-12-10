package desire

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateSocialNeeds_Extravert(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	traits := PersonalityTraits{Extraversion: 1.0} // High Extraversion
	ctx := SocialContext{Alone: true}

	// 1 Hour Alone
	// Rate = 0.5 * (0.5 + 1.0) = 0.75
	UpdateSocialNeeds(dp, 1.0, traits, ctx)

	assert.Equal(t, 0.75, dp.Needs[NeedCompanionship].Value)
}

func TestUpdateSocialNeeds_Introvert(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	traits := PersonalityTraits{Extraversion: 0.0} // Low Extraversion
	ctx := SocialContext{Alone: true}

	// 1 Hour Alone
	// Rate = 0.5 * (0.5 + 0.0) = 0.25
	UpdateSocialNeeds(dp, 1.0, traits, ctx)

	assert.Equal(t, 0.25, dp.Needs[NeedCompanionship].Value)
}

func TestUpdateSocialNeeds_Conversation(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	traits := PersonalityTraits{Extraversion: 0.5}
	ctx := SocialContext{Talking: false}

	// 1 Hour Silence
	// Rate = 1.0 + 0.5 = 1.5
	UpdateSocialNeeds(dp, 1.0, traits, ctx)

	assert.Equal(t, 1.5, dp.Needs[NeedConversation].Value)
}

func TestApplySocialPenalties_HighCompanionship(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	dp.Needs[NeedCompanionship].Value = 85 // Very lonely

	mods := ApplySocialPenalties(dp)

	// Should apply -10 Presence penalty
	assert.Equal(t, -10, mods.Presence, "High companionship need should reduce Presence")
}

func TestApplySocialPenalties_LowCompanionship(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	dp.Needs[NeedCompanionship].Value = 50 // Not lonely

	mods := ApplySocialPenalties(dp)

	// Should apply no penalty
	assert.Equal(t, 0, mods.Presence, "Low companionship need should not affect Presence")
}

func TestApplySocialPenalties_Threshold(t *testing.T) {
	dp := NewDesireProfile(uuid.New())
	dp.Needs[NeedCompanionship].Value = 80 // Exactly at threshold

	mods := ApplySocialPenalties(dp)

	// Should apply penalty at threshold
	assert.Equal(t, -10, mods.Presence)
}
