package relationship

// ApplyActionModifier updates affinity based on action type
// Modifiers respect bounds (-100 to +100)
func ApplyActionModifier(rel *Relationship, actionType string, value int) {
	delta := Affinity{}

	switch actionType {
	case "gift":
		// affection += giftValue / 10, trust += 5
		delta.Affection = value / 10
		delta.Trust = 5
	case "threat":
		// fear += 20, trust -= 10, affection -= 15
		delta.Fear = 20
		delta.Trust = -10
		delta.Affection = -15
	case "help":
		// affection += 10, trust += 8
		delta.Affection = 10
		delta.Trust = 8
	case "lie_caught":
		// trust -= 25, affection -= 10
		delta.Trust = -25
		delta.Affection = -10
	case "violence":
		// fear += 30, affection -= 40, trust -= 20
		delta.Fear = 30
		delta.Affection = -40
		delta.Trust = -20
	case "betrayal":
		// trust -= 50, affection -= 60, loyalty -= 0.3 (behavioral handled separately)
		delta.Trust = -50
		delta.Affection = -60
	case "support":
		// trust += 5, affection += 5
		delta.Trust = 5
		delta.Affection = 5
	}

	rel.CurrentAffinity.Affection = ClampAffinity(rel.CurrentAffinity.Affection + delta.Affection)
	rel.CurrentAffinity.Trust = ClampAffinity(rel.CurrentAffinity.Trust + delta.Trust)
	rel.CurrentAffinity.Fear = ClampAffinity(rel.CurrentAffinity.Fear + delta.Fear)
}

// ClampAffinity ensures value is between -100 and +100
func ClampAffinity(val int) int {
	if val > 100 {
		return 100
	}
	if val < -100 {
		return -100
	}
	return val
}
