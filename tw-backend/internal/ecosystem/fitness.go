package ecosystem

import (
	"tw-backend/internal/ecosystem/state"
)

// CalculateFitness determines how successful an entity is
func CalculateFitness(e *state.LivingEntityState, offspringCount int) float64 {
	// Formula:
	// (Age / MaxLifespan) * 0.3 +
	// (OffspringCount) * 0.4 +
	// (AverageNeedSatisfaction) * 0.2

	// MaxLifespan assumption: 1000 ticks
	ageScore := float64(e.Age) / 1000.0
	if ageScore > 1.0 {
		ageScore = 1.0
	}

	// Offspring Score (normalized, say max 10)
	reproScore := float64(offspringCount) / 10.0
	if reproScore > 1.0 {
		reproScore = 1.0
	}

	// Need Satisfaction (Average of inverse needs)
	// Hunger 0 = 100% satisfied
	avgNeeds := ((100 - e.Needs.Hunger) +
		(100 - e.Needs.Thirst) +
		e.Needs.Energy + // Energy is 100=Good
		(100 - e.Needs.Safety)) / 400.0

	return (ageScore * 0.3) + (reproScore * 0.4) + (avgNeeds * 0.3)
}
