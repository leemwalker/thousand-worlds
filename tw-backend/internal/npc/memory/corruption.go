package memory

import (
	"math/rand"
)

// CheckAndCorrupt applies corruption to a memory if clarity is low
// Corruption chance = 0.05 * (1 - clarity)
func CheckAndCorrupt(memory *Memory) {
	if memory.Corrupted {
		return // Already corrupted
	}

	chance := 0.05 * (1.0 - memory.Clarity)
	if chance <= 0 {
		return
	}

	if rand.Float64() < chance {
		CorruptMemory(memory)
	}
}

// CorruptMemory modifies the memory content and flags it as corrupted
func CorruptMemory(memory *Memory) {
	// Preserve original content if not already preserved
	if memory.OriginalContent == nil {
		memory.OriginalContent = memory.Content
	}

	memory.Corrupted = true

	// Apply specific corruption based on type
	switch c := memory.Content.(type) {
	case ObservationContent:
		// Location drift or detail loss
		if rand.Float64() < 0.5 {
			// Drift location slightly (mock implementation)
			c.Location.X += (rand.Float64() - 0.5) * 10
			c.Location.Y += (rand.Float64() - 0.5) * 10
		} else {
			// Vague details
			c.Event = "Something happened, but the details are fuzzy..."
		}
		memory.Content = c

	case ConversationContent:
		// Emotional shift or detail loss
		if rand.Float64() < 0.5 {
			c.Outcome = "uncertain"
		} else {
			// Shift emotional weight
			memory.EmotionalWeight += (rand.Float64() - 0.5) * 0.2 // +/- 0.1
			if memory.EmotionalWeight < 0 {
				memory.EmotionalWeight = 0
			}
			if memory.EmotionalWeight > 1 {
				memory.EmotionalWeight = 1
			}
		}
		memory.Content = c

	case EventContent:
		c.Description = "I remember this event, but the specifics escape me."
		memory.Content = c

	case RelationshipContent:
		// Shift affinity
		c.Affinity += int((rand.Float64() - 0.5) * 20) // +/- 10
		memory.Content = c
	}
}
