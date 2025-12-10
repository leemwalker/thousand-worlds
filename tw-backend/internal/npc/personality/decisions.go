package personality

import (
	"math/rand"
)

// Decision Option Tags
const (
	TagNovel       = "novel"
	TagFamiliar    = "familiar"
	TagPlanned     = "planned"
	TagImpulsive   = "impulsive"
	TagSocial      = "social"
	TagSolitary    = "solitary"
	TagCooperative = "cooperative"
	TagCompetitive = "competitive"
	TagSafe        = "safe"
	TagRisky       = "risky"
)

// DecisionOption represents a choice available to the NPC
type DecisionOption struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// CalculateDecisionScore computes the score for an option based on personality and mood
func CalculateDecisionScore(option DecisionOption, p *Personality, m *Mood) float64 {
	score := 0.0

	// Helper to get effective trait value (Base + Mood Modifier)
	getTrait := func(traitName string, baseVal float64) float64 {
		val := baseVal
		if m != nil {
			if mod, ok := m.Modifiers[traitName]; ok {
				val += mod
			}
		}
		// Clamp 0-100
		if val > 100 {
			return 100
		}
		if val < 0 {
			return 0
		}
		return val
	}

	openness := getTrait(TraitOpenness, p.Openness.Value)
	conscientiousness := getTrait(TraitConscientiousness, p.Conscientiousness.Value)
	extraversion := getTrait(TraitExtraversion, p.Extraversion.Value)
	agreeableness := getTrait(TraitAgreeableness, p.Agreeableness.Value)
	neuroticism := getTrait(TraitNeuroticism, p.Neuroticism.Value)

	// Apply Modifiers based on Tags
	for _, tag := range option.Tags {
		switch tag {
		case TagNovel:
			// High Openness prefers novel
			if openness >= 70 {
				score += 20
			}
			if openness <= 30 {
				score += 5
			} // Low openness prefers familiar, but still small bonus? Prompt says "Low Openness: +20 to return home (familiar)"
		case TagFamiliar:
			if openness <= 30 {
				score += 20
			}
			if openness >= 70 {
				score += 5
			}

		case TagPlanned:
			// High Conscientiousness prefers planned
			if conscientiousness >= 70 {
				score += 25
			}
			if conscientiousness <= 30 {
				score -= 10
			}
		case TagImpulsive:
			if conscientiousness <= 30 {
				score += 20
			}
			if conscientiousness >= 70 {
				score -= 15
			}

		case TagSocial:
			// High Extraversion prefers social
			if extraversion >= 70 {
				score += 30
			}
			if extraversion <= 30 {
				score -= 10
			} // Dislikes social? Prompt says "Low: +25 to stay home"
		case TagSolitary:
			if extraversion <= 30 {
				score += 25
			}
			if extraversion >= 70 {
				score -= 10
			} // Dislikes solitary?

		case TagCooperative:
			// High Agreeableness prefers cooperative
			if agreeableness >= 70 {
				score += 20
			}
			if agreeableness <= 30 {
				score -= 10
			}
		case TagCompetitive: // or "Negotiate Hard"
			if agreeableness <= 30 {
				score += 20
			}
			if agreeableness >= 70 {
				score -= 10
			}

		case TagSafe:
			// High Neuroticism prefers safe
			if neuroticism >= 70 {
				score += 30
			}
			if neuroticism <= 30 {
				score -= 5
			} // "panic response" -5? Prompt: "Low: ... -5 to panic response"
		case TagRisky:
			if neuroticism <= 30 {
				score += 15
			}
			if neuroticism >= 70 {
				score -= 20
			}
		}
	}

	// Add Random Factor (0-20)
	score += rand.Float64() * 20.0

	return score
}
