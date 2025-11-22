package relationship

import (
	"math"
)

// CalculateDrift compares current behavior to baseline
// Returns the max drift across all traits and details
func CalculateDrift(baseline BehavioralProfile, recent BehavioralProfile) DriftMetrics {
	metrics := DriftMetrics{
		AffectedTraits: []string{},
	}

	maxDrift := 0.0

	// Helper to check drift for a trait
	checkTrait := func(name string, base, curr float64) {
		diff := curr - base
		absDiff := math.Abs(diff)

		if absDiff > maxDrift {
			maxDrift = absDiff
			// Direction: +1 if "improved" (subjective, but let's assume higher is "more of trait")
			// Prompt says: "+1 if personality improves by NPC's values".
			// Without NPC values, we'll just track direction of change relative to trait.
			// Let's store direction of the MAX drift.
			if diff > 0 {
				metrics.DriftDirection = 1
			} else {
				metrics.DriftDirection = -1
			}
		}

		if absDiff >= 0.3 {
			metrics.AffectedTraits = append(metrics.AffectedTraits, name)
		}
	}

	checkTrait("aggression", baseline.Aggression, recent.Aggression)
	checkTrait("generosity", baseline.Generosity, recent.Generosity)
	checkTrait("honesty", baseline.Honesty, recent.Honesty)
	checkTrait("sociability", baseline.Sociability, recent.Sociability)
	checkTrait("recklessness", baseline.Recklessness, recent.Recklessness)
	checkTrait("loyalty", baseline.Loyalty, recent.Loyalty)

	metrics.DriftScore = maxDrift
	metrics.DriftLevel = GetDriftLevel(maxDrift)

	return metrics
}

// GetDriftLevel returns the classification string
func GetDriftLevel(score float64) string {
	if score >= 0.7 {
		return "Severe"
	}
	if score >= 0.5 {
		return "Moderate"
	}
	if score >= 0.3 {
		return "Subtle"
	}
	return "None"
}
