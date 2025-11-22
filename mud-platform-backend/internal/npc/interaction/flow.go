package interaction

import (
	"math/rand"
)

// GenerateGreeting selects a greeting based on affection
func GenerateGreeting(affection int) string {
	var templates []string
	if affection > 60 {
		templates = GreetingTemplates["high_affection"]
	} else if affection >= 20 {
		templates = GreetingTemplates["medium_affection"]
	} else {
		templates = GreetingTemplates["low_affection"]
	}

	return templates[rand.Intn(len(templates))]
}

// ProcessTurn advances the conversation state
func ProcessTurn(conv *Conversation, input string) string {
	// Simplified state machine for Phase 5.3
	switch conv.CurrentStage {
	case StageGreeting:
		conv.CurrentStage = StageTopic
		return "Greeting exchanged."
	case StageTopic:
		conv.CurrentStage = StageResponse
		return "Topic proposed."
	case StageResponse:
		conv.CurrentStage = StageEnd
		return "Response given."
	}
	return "Conversation ended."
}
