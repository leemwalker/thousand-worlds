package interaction

import (
	"tw-backend/internal/npc/relationship"
)

// RelationshipUpdate holds changes to affinity
type RelationshipUpdate struct {
	AffectionDelta int
	TrustDelta     int
}

// ApplyInteractionOutcome calculates relationship changes
func ApplyInteractionOutcome(outcome string) RelationshipUpdate {
	update := RelationshipUpdate{}

	switch outcome {
	case OutcomePositive:
		// Positive: affection += 3, trust += 2
		update.AffectionDelta = 3
		update.TrustDelta = 2
	case OutcomeNegative:
		// Negative: affection -= 5, trust -= 3
		update.AffectionDelta = -5
		update.TrustDelta = -3
	case OutcomeNeutral:
		// Neutral: affection += 1
		update.AffectionDelta = 1
		update.TrustDelta = 0
	}

	return update
}

// UpdateRelationship applies the update to a relationship struct
func UpdateRelationship(rel *relationship.Relationship, update RelationshipUpdate) {
	rel.CurrentAffinity.Affection += update.AffectionDelta
	rel.CurrentAffinity.Trust += update.TrustDelta

	// Clamp values -100 to 100
	clamp(&rel.CurrentAffinity.Affection)
	clamp(&rel.CurrentAffinity.Trust)
}

func clamp(val *int) {
	if *val > 100 {
		*val = 100
	}
	if *val < -100 {
		*val = -100
	}
}
