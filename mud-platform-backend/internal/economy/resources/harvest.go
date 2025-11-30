package resources

import (
	"math"
	"math/rand"
)

// CalculateHarvestYield calculates the amount of resources harvested
func CalculateHarvestYield(node *ResourceNode, request HarvestRequest) HarvestResult {
	// Check skill requirement
	skillDeficit := node.MinSkillLevel - request.GathererSkill

	if skillDeficit > 20 {
		// 50% chance to fail completely
		if rand.Float64() < 0.5 {
			return HarvestResult{
				Success:       false,
				YieldAmount:   0,
				FailureReason: "skill_too_low",
			}
		}
	} else if skillDeficit > 0 {
		// 25% chance to fail
		if rand.Float64() < 0.25 {
			return HarvestResult{
				Success:       false,
				YieldAmount:   0,
				FailureReason: "skill_insufficient",
			}
		}
	}

	// Calculate efficiency
	skillModifier := calculateSkillModifier(request.GathererSkill)
	toolModifier := calculateToolModifier(request.ToolQuality)
	randomFactor := 0.8 + rand.Float64()*0.4 // 0.8 to 1.2

	efficiency := skillModifier * toolModifier * randomFactor

	// Calculate base yield
	baseYield := float64(node.Quantity) * efficiency
	finalYield := int(math.Round(baseYield))

	// Ensure non-negative
	if finalYield < 0 {
		finalYield = 0
	}

	return HarvestResult{
		Success:       true,
		YieldAmount:   finalYield,
		FailureReason: "",
	}
}

// calculateSkillModifier returns the skill-based efficiency multiplier
// Formula: 0.5 + (skill / 200)
// Skill 0: 50% efficiency
// Skill 50: 75% efficiency
// Skill 100: 100% efficiency
// Skill 200: 150% efficiency
func calculateSkillModifier(skill int) float64 {
	return 0.5 + (float64(skill) / 200.0)
}

// calculateToolModifier returns the tool-based efficiency multiplier
// Formula: toolQuality / 100
// 0 (no tool): 0%
// 50 (basic): 50%
// 70 (basic+): 70%
// 100 (good): 100%
// 130 (excellent): 130%
func calculateToolModifier(toolQuality int) float64 {
	return float64(toolQuality) / 100.0
}
