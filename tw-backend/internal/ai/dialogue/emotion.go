package dialogue

import "strings"

func (s *DialogueService) inferEmotionalReaction(text string) (string, float64) {
	lower := strings.ToLower(text)

	// Keywords
	if strings.Contains(lower, "wonderful") || strings.Contains(lower, "happy") || strings.Contains(lower, "love") {
		return "joy", 0.7
	}
	if strings.Contains(lower, "damn") || strings.Contains(lower, "furious") || strings.Contains(lower, "unacceptable") {
		return "anger", 0.8
	}
	if strings.Contains(lower, "terrifying") || strings.Contains(lower, "worried") || strings.Contains(lower, "scared") {
		return "fear", 0.7
	}
	if strings.Contains(lower, "sad") || strings.Contains(lower, "unfortunate") {
		return "sadness", 0.6
	}

	// Punctuation
	if strings.Contains(text, "!!") {
		return "excited", 0.6
	}
	if strings.Contains(text, "??") {
		return "confused", 0.5
	}

	return "neutral", 0.1
}
