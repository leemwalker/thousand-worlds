package relationship

// CalculateAverageBaseline computes the average profile from interaction history
func CalculateAverageBaseline(interactions []Interaction) BehavioralProfile {
	if len(interactions) == 0 {
		return BehavioralProfile{}
	}

	sum := BehavioralProfile{}
	count := float64(len(interactions))

	for _, i := range interactions {
		sum.Aggression += i.BehavioralContext.Aggression
		sum.Generosity += i.BehavioralContext.Generosity
		sum.Honesty += i.BehavioralContext.Honesty
		sum.Sociability += i.BehavioralContext.Sociability
		sum.Recklessness += i.BehavioralContext.Recklessness
		sum.Loyalty += i.BehavioralContext.Loyalty
	}

	return BehavioralProfile{
		Aggression:   sum.Aggression / count,
		Generosity:   sum.Generosity / count,
		Honesty:      sum.Honesty / count,
		Sociability:  sum.Sociability / count,
		Recklessness: sum.Recklessness / count,
		Loyalty:      sum.Loyalty / count,
	}
}

// UpdateBaseline updates the long-term baseline with new data
// Formula: new = (old * 0.9) + (recent * 0.1)
func UpdateBaseline(current BehavioralProfile, recent BehavioralProfile) BehavioralProfile {
	return BehavioralProfile{
		Aggression:   (current.Aggression * 0.9) + (recent.Aggression * 0.1),
		Generosity:   (current.Generosity * 0.9) + (recent.Generosity * 0.1),
		Honesty:      (current.Honesty * 0.9) + (recent.Honesty * 0.1),
		Sociability:  (current.Sociability * 0.9) + (recent.Sociability * 0.1),
		Recklessness: (current.Recklessness * 0.9) + (recent.Recklessness * 0.1),
		Loyalty:      (current.Loyalty * 0.9) + (recent.Loyalty * 0.1),
	}
}
