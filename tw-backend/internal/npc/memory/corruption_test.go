package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAndCorrupt_HighClarity(t *testing.T) {
	mem := Memory{Clarity: 1.0, Corrupted: false}
	CheckAndCorrupt(&mem)
	assert.False(t, mem.Corrupted)
}

func TestCheckAndCorrupt_AlreadyCorrupted(t *testing.T) {
	mem := Memory{Clarity: 0.1, Corrupted: true}
	CheckAndCorrupt(&mem)
	// Should remain corrupted, don't corrupt twice
	assert.True(t, mem.Corrupted)
}

func TestCorruptMemory_PreservesOriginal(t *testing.T) {
	original := ObservationContent{Event: "Clear Event"}
	mem := Memory{
		Clarity: 0.1,
		Content: original,
	}

	CorruptMemory(&mem)

	assert.True(t, mem.Corrupted)
	assert.NotNil(t, mem.OriginalContent)
	assert.Equal(t, original, mem.OriginalContent.(ObservationContent))

	// Verify content changed
	current := mem.Content.(ObservationContent)
	// Either location changed or event text changed
	changed := current.Event != original.Event || current.Location != original.Location
	assert.True(t, changed)
}

func TestCorruptMemory_ObservationContent(t *testing.T) {
	mem := Memory{
		Type: MemoryTypeObservation,
		Content: ObservationContent{
			Event:    "Original event",
			Location: Location{X: 100, Y: 200},
		},
	}

	CorruptMemory(&mem)
	assert.True(t, mem.Corrupted)
	assert.NotNil(t, mem.OriginalContent)
}

func TestCorruptMemory_ConversationContent(t *testing.T) {
	mem := Memory{
		Type:            MemoryTypeConversation,
		EmotionalWeight: 0.5,
		Content: ConversationContent{
			Outcome: "positive",
		},
	}

	CorruptMemory(&mem)
	assert.True(t, mem.Corrupted)
	assert.NotNil(t, mem.OriginalContent)
}

func TestCorruptMemory_EventContent(t *testing.T) {
	mem := Memory{
		Type: MemoryTypeEvent,
		Content: EventContent{
			Description: "Original description",
		},
	}

	CorruptMemory(&mem)
	assert.True(t, mem.Corrupted)
	assert.NotNil(t, mem.OriginalContent)

	// Event corruption should change description
	corrupted := mem.Content.(EventContent)
	assert.NotEqual(t, "Original description", corrupted.Description)
}

func TestCorruptMemory_RelationshipContent(t *testing.T) {
	mem := Memory{
		Type: MemoryTypeRelationship,
		Content: RelationshipContent{
			Affinity: 50,
		},
	}

	originalAffinity := 50
	CorruptMemory(&mem)
	assert.True(t, mem.Corrupted)
	assert.NotNil(t, mem.OriginalContent)

	// Relationship corruption should shift affinity
	corrupted := mem.Content.(RelationshipContent)
	// Affinity might have changed (could be +/- 10)
	// We can't predict exact value due to randomness, but verify it's valid
	assert.GreaterOrEqual(t, corrupted.Affinity, originalAffinity-10)
	assert.LessOrEqual(t, corrupted.Affinity, originalAffinity+10)
}
