package interview

import (
	"fmt"
	"strings"
)

// BuildInterviewPrompt constructs the prompt for the LLM
func BuildInterviewPrompt(state InterviewState, nextTopic Topic, history []ConversationTurn) string {
	// 1. Calculate progress
	totalQuestions := len(AllTopics)
	questionsAnswered := state.CurrentTopicIndex

	// 2. Format gathered info
	var existingAnswers strings.Builder
	for _, topic := range AllTopics {
		if ans, ok := state.Answers[topic.Name]; ok {
			existingAnswers.WriteString(fmt.Sprintf("- %s: %s\n", topic.Name, ans))
		}
	}

	// 3. Get last response
	lastResponse := ""
	if len(history) > 0 {
		lastResponse = history[len(history)-1].Answer
	}

	// 4. Construct template
	template := `You are a world-building assistant helping a player create a custom game world.

INTERVIEW PROGRESS:
%d / %d questions answered

INFORMATION GATHERED SO FAR:
%s

CURRENT CATEGORY: %s
CURRENT TOPIC: %s (%s)

Ask the next question about %s. Be conversational and enthusiastic.
If the player's previous answer was vague, ask a clarifying follow-up.
Keep questions concise (1-2 sentences).

Previous player response: "%s"
`

	return fmt.Sprintf(template,
		questionsAnswered, totalQuestions,
		existingAnswers.String(),
		nextTopic.Category,
		nextTopic.Name, nextTopic.Description,
		nextTopic.Name,
		lastResponse,
	)
}
