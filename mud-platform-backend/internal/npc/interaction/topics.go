package interaction

import (
	"mud-platform-backend/internal/npc/memory"
	"strings"
	"time"
)

// Topic represents a conversation subject
type Topic struct {
	Text   string
	Score  float64
	Memory *memory.Memory
}

// SelectTopic chooses the highest scoring topic
func SelectTopic(memories []memory.Memory, sharedMemories []memory.Memory) Topic {
	var bestTopic Topic
	bestTopic.Score = -1.0

	// Helper to calculate score
	calcScore := func(mem memory.Memory, isShared bool) float64 {
		// score = (emotionalWeight × 0.4) + (recency × 0.3) + (shared × 0.3)

		// Recency (0-1)
		hoursSince := time.Since(mem.Timestamp).Hours()
		recency := 1.0
		if hoursSince > 24 {
			recency = 0.0
		} else {
			recency = 1.0 - (hoursSince / 24.0)
		}

		sharedBonus := 0.0
		if isShared {
			sharedBonus = 1.0
		}

		return (mem.EmotionalWeight * 0.4) + (recency * 0.3) + (sharedBonus * 0.3)
	}

	// Check individual memories
	for _, mem := range memories {
		score := calcScore(mem, false)
		if score > bestTopic.Score {
			bestTopic = Topic{
				Text:   formatTopic(mem),
				Score:  score,
				Memory: &mem,
			}
		}
	}

	// Check shared memories (higher priority usually due to shared bonus)
	for _, mem := range sharedMemories {
		score := calcScore(mem, true)
		if score > bestTopic.Score {
			bestTopic = Topic{
				Text:   formatTopic(mem),
				Score:  score,
				Memory: &mem,
			}
		}
	}

	// Fallback if no memories
	if bestTopic.Score < 0 {
		bestTopic = Topic{
			Text:  "the weather",
			Score: 0.0,
		}
	}

	return bestTopic
}

func formatTopic(mem memory.Memory) string {
	// Simple formatter, in real system would use LLM or more complex logic
	if content, ok := mem.Content.(memory.EventContent); ok {
		return content.Description
	}
	return "something I saw"
}

// GenerateTopicStatement creates the dialogue line for the topic
func GenerateTopicStatement(topic Topic) string {
	template := TopicTemplates["memory"]
	return strings.Replace(template, "{topic}", topic.Text, 1)
}
