package relationship

import (
	"fmt"
)

// Reaction represents the NPC's response to behavioral drift
type Reaction struct {
	Comment           string
	AffinityModifier  Affinity
	MemoryTrigger     bool
	MemoryEmotion     float64
	MemoryDescription string
}

// GenerateReaction creates a response based on drift metrics
func GenerateReaction(metrics DriftMetrics) Reaction {
	r := Reaction{}

	switch metrics.DriftLevel {
	case "Subtle":
		// 0.3-0.5: Concerned comments
		r.Comment = "You seem different lately."
		r.MemoryTrigger = true
		r.MemoryEmotion = 0.3
		r.MemoryDescription = fmt.Sprintf("Noticed subtle personality change: %v", metrics.AffectedTraits)

		// Modifier: +/- 5 affection depending on direction
		if metrics.DriftDirection > 0 {
			r.AffinityModifier.Affection = 5
		} else {
			r.AffinityModifier.Affection = -5
		}

	case "Moderate":
		// 0.5-0.7: Direct questioning
		r.Comment = "That's not like you. What's going on?"
		r.MemoryTrigger = true
		r.MemoryEmotion = 0.7
		r.MemoryDescription = fmt.Sprintf("Concerned about significant personality shift: %v", metrics.AffectedTraits)

		// Modifier: Trust -10, Fear +5 if negative
		if metrics.DriftDirection < 0 {
			r.AffinityModifier.Trust = -10
			r.AffinityModifier.Fear = 5
		} else {
			// Positive moderate drift? Maybe trust +5?
			// Prompt says "Positive drift (coward -> brave): affection += 15, trust += 10"
			// Let's apply the general modifier formula from prompt:
			// "affinityDelta = driftMagnitude * 50 * directionMultiplier"
			// Wait, prompt has specific rules for levels AND a general formula.
			// "Build relationship modifier based on drift: ... Calculate modifier: affinityDelta = driftMagnitude * 50 * directionMultiplier"
			// This seems to be the general rule for the *drift event itself*, separate from the reaction level?
			// Or does the reaction level define the *response* (comment/memory) and the formula defines the *affinity*?
			// Let's use the formula for affinity as it scales.
		}

	case "Severe":
		// 0.7+: Alarmed response
		r.Comment = "You're not yourself. Something is very wrong."
		r.MemoryTrigger = true
		r.MemoryEmotion = 1.0
		r.MemoryDescription = fmt.Sprintf("Alarmed by severe personality alteration: %v", metrics.AffectedTraits)

		// Specific penalties from prompt
		r.AffinityModifier.Trust = -25
	}

	// Apply general drift modifier formula if not overridden by specific penalties?
	// Prompt: "Build relationship modifier based on drift... Calculate modifier: affinityDelta = driftMagnitude * 50 * directionMultiplier"
	// Let's add this to the reaction modifier.

	magnitude := metrics.DriftScore
	direction := float64(metrics.DriftDirection)

	// Affection change
	affectionChange := int(magnitude * 50.0 * direction)
	r.AffinityModifier.Affection += affectionChange

	// Trust change (usually correlates with affection in this formula, but prompt says "trust -= 30" for negative)
	// Let's scale trust similarly
	trustChange := int(magnitude * 30.0 * direction)
	r.AffinityModifier.Trust += trustChange

	return r
}
