package memory

import (
	"fmt"
	"strings"
)

// Common emotions to auto-tag
var emotionKeywords = []string{
	"joy", "happy", "excited",
	"anger", "angry", "furious",
	"fear", "scared", "terrified",
	"sadness", "sad", "grief",
	"surprise", "shocked",
	"disgust", "revolted",
}

// GenerateTags extracts tags from memory content
func GenerateTags(memory Memory) []string {
	tags := make(map[string]bool) // Use map for deduplication

	// Helper to add tags
	add := func(t string) {
		if t != "" {
			tags[strings.ToLower(t)] = true
		}
	}

	// 1. Tag Type
	add(memory.Type)

	// 2. Tag Content-specific fields
	switch content := memory.Content.(type) {
	case ObservationContent:
		add(fmt.Sprintf("location_%s", content.Location.WorldID))
		add(content.TimeOfDay)
		add(content.WeatherConditions)
		for _, entity := range content.EntitiesPresent {
			add(fmt.Sprintf("entity_%s", entity))
		}
		// Extract from event description
		extractKeywords(content.Event, add)

	case ConversationContent:
		add(fmt.Sprintf("location_%s", content.Location.WorldID))
		add(content.Outcome)
		for _, p := range content.Participants {
			add(fmt.Sprintf("participant_%s", p))
		}
		// Extract from dialogue
		for _, line := range content.Dialogue {
			extractKeywords(line.Text, add)
			add(line.Emotion)
		}

	case EventContent:
		add(fmt.Sprintf("location_%s", content.Location.WorldID))
		add(content.EventType)
		add(content.EmotionalResponse)
		for _, p := range content.Participants {
			add(fmt.Sprintf("participant_%s", p))
		}
		extractKeywords(content.Description, add)

	case RelationshipContent:
		add(content.RelationshipType)
		extractKeywords(content.FirstImpression, add)
	}

	// Convert map to slice
	result := make([]string, 0, len(tags))
	for t := range tags {
		result = append(result, t)
	}
	return result
}

func extractKeywords(text string, addFunc func(string)) {
	lower := strings.ToLower(text)
	for _, emotion := range emotionKeywords {
		if strings.Contains(lower, emotion) {
			addFunc(emotion)
		}
	}
}
