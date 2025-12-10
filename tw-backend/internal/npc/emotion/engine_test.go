package emotion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeEvent_Threat(t *testing.T) {
	engine := NewEmotionEngine()
	ctx := EventContext{IsThreat: true}
	traits := PersonalityTraits{}

	profile, weight := engine.AnalyzeEvent(ctx, traits)

	assert.Equal(t, 0.95, weight)
	assert.Equal(t, 0.95, profile[Fear])
	assert.Equal(t, 0.5, profile[Anger])
}

func TestAnalyzeEvent_PersonalityModifiers(t *testing.T) {
	engine := NewEmotionEngine()
	ctx := EventContext{IsThreat: true}           // Base Fear 0.95
	traits := PersonalityTraits{Neuroticism: 1.0} // Fear +20%

	profile, _ := engine.AnalyzeEvent(ctx, traits)

	// 0.95 * 1.2 = 1.14 -> Capped at 1.0
	assert.Equal(t, 1.0, profile[Fear])
}

func TestCalculateComplexEmotions(t *testing.T) {
	engine := NewEmotionEngine()
	profile := EmotionProfile{
		Joy:      0.8,
		Surprise: 0.4,
	}

	engine.CalculateComplexEmotions(profile)

	// Anticipation = 0.8*0.5 + 0.4*0.5 = 0.4 + 0.2 = 0.6
	assert.InDelta(t, 0.6, profile[Anticipation], 0.0001)
}

func TestAnalyzeEvent_Gift(t *testing.T) {
	engine := NewEmotionEngine()
	ctx := EventContext{
		GiftValue:   100,
		WealthLevel: 1000, // Ratio 0.1
	}
	traits := PersonalityTraits{}

	_, weight := engine.AnalyzeEvent(ctx, traits)

	// Weight = 0.3 + (0.1 * 0.4) = 0.34
	assert.InDelta(t, 0.34, weight, 0.01)
}
