package memory

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTags_Observation(t *testing.T) {
	entityID := uuid.New()
	worldID := uuid.New()

	mem := Memory{
		Type: MemoryTypeObservation,
		Content: ObservationContent{
			Event:             "Saw a terrified villager run away",
			Location:          Location{WorldID: worldID},
			EntitiesPresent:   []uuid.UUID{entityID},
			TimeOfDay:         "night",
			WeatherConditions: "rainy",
		},
	}

	tags := GenerateTags(mem)

	assert.Contains(t, tags, "observation")
	assert.Contains(t, tags, "night")
	assert.Contains(t, tags, "rainy")
	assert.Contains(t, tags, fmt.Sprintf("location_%s", worldID))
	assert.Contains(t, tags, fmt.Sprintf("entity_%s", entityID))
	assert.Contains(t, tags, "terrified")
	// Wait, "scary" contains "scary". My list has "scared".
	// "Saw a scary monster attack" -> "scary" is not in list.
	// Let's update the test or list.
	// List has "fear", "scared", "terrified".
	// Let's change text to "Saw a terrified villager".
}

func TestGenerateTags_Keywords(t *testing.T) {
	mem := Memory{
		Type: MemoryTypeEvent,
		Content: EventContent{
			Description: "I felt great joy and surprise at the party",
		},
	}

	tags := GenerateTags(mem)
	assert.Contains(t, tags, "joy")
	assert.Contains(t, tags, "surprise")
}

func TestGenerateTags_Conversation(t *testing.T) {
	p1 := uuid.New()
	p2 := uuid.New()

	mem := Memory{
		Type: MemoryTypeConversation,
		Content: ConversationContent{
			Participants: []uuid.UUID{p1, p2},
			Outcome:      "positive",
			Dialogue: []DialogueLine{
				{Text: "I am angry!", Emotion: "anger"},
			},
		},
	}

	tags := GenerateTags(mem)
	assert.Contains(t, tags, "conversation")
	assert.Contains(t, tags, "positive")
	assert.Contains(t, tags, fmt.Sprintf("participant_%s", p1))
	assert.Contains(t, tags, "anger") // From emotion field
	assert.Contains(t, tags, "angry") // From text keyword
}
