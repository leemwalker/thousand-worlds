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
