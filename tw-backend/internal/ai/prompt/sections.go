package prompt

import (
	"fmt"
	"tw-backend/internal/character"
	"tw-backend/internal/npc/memory"
	"tw-backend/internal/npc/personality"
	"tw-backend/internal/npc/relationship"
	"strings"
)

func buildPersonalitySection(p *personality.Personality) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("- Openness: %.1f/100\n", p.Openness.Value))
	sb.WriteString(fmt.Sprintf("- Conscientiousness: %.1f/100\n", p.Conscientiousness.Value))
	sb.WriteString(fmt.Sprintf("- Extraversion: %.1f/100\n", p.Extraversion.Value))
	sb.WriteString(fmt.Sprintf("- Agreeableness: %.1f/100\n", p.Agreeableness.Value))
	sb.WriteString(fmt.Sprintf("- Neuroticism: %.1f/100\n", p.Neuroticism.Value))
	return sb.String()
}

func buildStateSection(mood *personality.Mood, desire string, urgency float64, cond character.Attributes) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("- Mood: %s\n", mood.Type))
	sb.WriteString(fmt.Sprintf("- Top Desire: %s (urgency: %.1f/100)\n", desire, urgency))
	// Assuming Physical Condition is derived from Attributes roughly for now
	sb.WriteString(fmt.Sprintf("- Physical Condition: Vitality %d/100, Endurance %d/100", cond.Vitality, cond.Endurance))
	return sb.String()
}

func buildRelationshipSection(rel *relationship.Relationship) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("- Affection: %d/100\n", rel.CurrentAffinity.Affection))
	sb.WriteString(fmt.Sprintf("- Trust: %d/100\n", rel.CurrentAffinity.Trust))
	sb.WriteString(fmt.Sprintf("- Fear: %d/100\n", rel.CurrentAffinity.Fear))
	// Shared History could be summarized from RecentInteractions if needed, skipped for brevity in MVP
	return sb.String()
}

func buildMemoriesSection(memories []memory.Memory) string {
	if len(memories) == 0 {
		return "None"
	}
	var sb strings.Builder
	for _, mem := range memories {
		// Simple summary
		if content, ok := mem.Content.(memory.EventContent); ok {
			sb.WriteString(fmt.Sprintf("- %s (Emotional Weight: %.1f)\n", content.Description, mem.EmotionalWeight))
		} else if content, ok := mem.Content.(memory.ConversationContent); ok {
			sb.WriteString(fmt.Sprintf("- Conversation about %s (Outcome: %s)\n", "topic", content.Outcome))
		} else {
			sb.WriteString(fmt.Sprintf("- Memory of type %s\n", mem.Type))
		}
	}
	return sb.String()
}
