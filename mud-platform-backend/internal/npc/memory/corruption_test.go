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
