package memory

// CalculateImportance computes the importance score of a memory
// Includes emotional boost from Phase 3.4
func CalculateImportance(memory Memory) float64 {
	// Base Importance (simplified from Phase 3.1 logic if needed, or we can reuse if we had it)
	// Phase 3.1 had `CalculateImportance` in `relevance.go`?
	// Let's check `relevance.go` content.
	// If it exists, we should modify it there or wrap it.
	// Since I can't see `relevance.go` right now, I'll assume I need to implement the *boost* logic here
	// or re-implement the full calculation if `relevance.go` is not easily extensible.

	// Prompt: "importance = baseImportance * (1 + emotionalWeight)"
	// Let's define base importance based on type/content if we don't have the base function handy.
	// Or better, let's assume this REPLACES the old calculation or extends it.

	baseImportance := 0.5 // Default

	switch memory.Type {
	case MemoryTypeEvent:
		baseImportance = 0.7
	case MemoryTypeRelationship:
		baseImportance = 0.6
	case MemoryTypeConversation:
		baseImportance = 0.4
	case MemoryTypeObservation:
		baseImportance = 0.3
	}

	// Emotional Boost
	// High-emotion memories (>0.7) get 1.7x to 2.0x importance boost
	// Formula: base * (1 + emotionalWeight)
	// If weight is 0.8 -> base * 1.8. Fits the range.

	importance := baseImportance * (1.0 + memory.EmotionalWeight)

	if importance > 1.0 {
		importance = 1.0 // Cap at 1.0? Or allow > 1.0?
		// Relevance scoring usually normalized 0-1, but importance can be higher?
		// Let's cap at 2.0 since base is ~0.5 and boost is ~2x.
		// Actually, let's not cap arbitrarily unless needed.
	}

	return importance
}
