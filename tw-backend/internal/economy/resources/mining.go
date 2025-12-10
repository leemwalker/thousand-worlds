package resources

// CalculateMiningYield calculates yield specifically for mineral resources
// Applies additional modifiers for depth and ore concentration
func CalculateMiningYield(node *ResourceNode, deposit *MineralDeposit, request HarvestRequest) HarvestResult {
	// First calculate base harvest result (skill, tool, random)
	baseResult := CalculateHarvestYield(node, request)

	if !baseResult.Success {
		return baseResult
	}

	// Apply depth penalty
	// Deeper mines are harder to work in
	// Formula: 1.0 - (depth / 10000m) * 0.5
	// Surface (0m): 1.0 multiplier
	// 1000m: 0.95 multiplier
	// 5000m: 0.75 multiplier
	depthModifier := 1.0 - (deposit.Depth/10000.0)*0.5
	if depthModifier < 0.1 {
		depthModifier = 0.1 // Minimum 10% efficiency even at extreme depth
	}

	// Apply concentration bonus/penalty
	// Concentration is directly used as a multiplier
	// 1.0 (pure vein): 100% yield
	// 0.5 (poor vein): 50% yield
	concentrationModifier := deposit.Concentration

	// Calculate final yield
	finalYield := float64(baseResult.YieldAmount) * depthModifier * concentrationModifier

	return HarvestResult{
		Success:       true,
		YieldAmount:   int(finalYield),
		FailureReason: "",
	}
}
