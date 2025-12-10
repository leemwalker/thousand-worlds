package desire

import (
	"sort"
)

// Priority represents a calculated need priority
type Priority struct {
	NeedName string
	Score    float64
}

// CalculatePriorities determines the most urgent needs
func CalculatePriorities(profile *DesireProfile, traits PersonalityTraits) []Priority {
	var priorities []Priority

	for _, need := range profile.Needs {
		score := CalculatePriorityScore(need, traits)
		priorities = append(priorities, Priority{NeedName: need.Name, Score: score})
	}

	// Sort descending
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i].Score > priorities[j].Score
	})

	return priorities
}

// CalculatePriorityScore computes urgency * personality * tier
func CalculatePriorityScore(need *Need, traits PersonalityTraits) float64 {
	// 1. Urgency (Raw Value)
	urgency := need.Value

	// 2. Tier Multiplier
	tierMult := 1.0
	switch need.Tier {
	case TierSurvival:
		tierMult = 4.0
	case TierSocial:
		tierMult = 2.0
	case TierAchievement:
		tierMult = 1.5
	case TierPleasure:
		tierMult = 1.0
	}

	// 3. Personality Weight
	// 0.5 to 2.0 based on relevant trait
	persWeight := 1.0
	switch need.Name {
	case NeedHunger, NeedThirst, NeedSleep, NeedSafety:
		// Neuroticism (anxious about needs)
		// 0.0 -> 0.5, 1.0 -> 2.0?
		// Let's map 0-1 to 0.5-1.5? Or 0.8-1.2?
		// Prompt: "affected by Neuroticism"
		// Let's use: 1.0 + (Neuroticism - 0.5) -> 0.5 to 1.5
		persWeight = 0.5 + traits.Neuroticism // 0.5 to 1.5
	case NeedCompanionship, NeedConversation, NeedAffection:
		// Extraversion
		persWeight = 0.5 + traits.Extraversion
	case NeedTaskCompletion, NeedSkillImprovement, NeedResourceAcquisition:
		// Conscientiousness
		persWeight = 0.5 + traits.Conscientiousness
	case NeedCuriosity, NeedCreativity:
		// Openness
		persWeight = 0.5 + traits.Openness
	case NeedHedonism:
		// Low Conscientiousness? High Extraversion?
		// Let's use Extraversion for now.
		persWeight = 0.5 + traits.Extraversion
	}

	return urgency * persWeight * tierMult
}

// ShouldInterrupt checks if a new priority should override the current one
func ShouldInterrupt(currentScore, candidateScore float64, isCriticalSurvival bool) bool {
	if isCriticalSurvival {
		return true
	}
	// Interruption threshold: new priority must be 2x current priority
	return candidateScore > (currentScore * 2.0)
}
