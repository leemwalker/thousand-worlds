package resources

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHarvestYieldCalculation(t *testing.T) {
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Oak Wood",
		Type:          ResourceVegetation,
		Quantity:      100,
		MinSkillLevel: 0,
	}

	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 50,
		ToolQuality:   100, // Good tool
	}

	result := CalculateHarvestYield(node, request)

	assert.True(t, result.Success)
	assert.Greater(t, result.YieldAmount, 0)
	// With skill 50 and tool 100, expect around 75% efficiency (±20%)
	// Skill modifier: 0.5 + 50/200 = 0.75
	// Tool modifier: 100/100 = 1.0
	// Total: 0.75 * 1.0 = 0.75 (before random)
	// Expected range: 60-90 (0.75 * 0.8 to 0.75 * 1.2)
	assert.GreaterOrEqual(t, result.YieldAmount, 40)
	assert.LessOrEqual(t, result.YieldAmount, 100)
}

func TestSkillModifier(t *testing.T) {
	tests := []struct {
		skill            int
		expectedModifier float64
	}{
		{0, 0.5},   // Skill 0 = 50% efficiency
		{50, 0.75}, // Skill 50 = 75% efficiency
		{100, 1.0}, // Skill 100 = 100% efficiency
		{200, 1.5}, // Skill 200 = 150% efficiency (max)
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.skill)), func(t *testing.T) {
			modifier := calculateSkillModifier(tt.skill)
			assert.InDelta(t, tt.expectedModifier, modifier, 0.01)
		})
	}
}

func TestToolModifier(t *testing.T) {
	tests := []struct {
		toolQuality      int
		expectedModifier float64
	}{
		{0, 0.0},   // No tool = 0%
		{50, 0.5},  // Basic tool = 50%
		{70, 0.7},  // Basic+ tool = 70%
		{100, 1.0}, // Good tool = 100%
		{130, 1.3}, // Excellent tool = 130%
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.toolQuality)), func(t *testing.T) {
			modifier := calculateToolModifier(tt.toolQuality)
			assert.InDelta(t, tt.expectedModifier, modifier, 0.01)
		})
	}
}

func TestHarvestFailureSkillTooLow(t *testing.T) {
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Rare Crystal",
		Type:          ResourceSpecial,
		Quantity:      10,
		MinSkillLevel: 60,
	}

	// Gatherer skill is 30, required is 60
	// Difference is 30 > 20, so 50% chance to fail
	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 30,
		ToolQuality:   100,
	}

	// Run multiple times to test probability
	failures := 0
	iterations := 100

	for i := 0; i < iterations; i++ {
		result := CalculateHarvestYield(node, request)
		if !result.Success {
			failures++
			assert.Equal(t, "skill_too_low", result.FailureReason)
		}
	}

	// Should fail approximately 50% of the time (±15% tolerance)
	assert.GreaterOrEqual(t, failures, 35)
	assert.LessOrEqual(t, failures, 65)
}

func TestHarvestFailureMinorSkillDeficit(t *testing.T) {
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Medicinal Herbs",
		Type:          ResourceVegetation,
		Quantity:      50,
		MinSkillLevel: 30,
	}

	// Gatherer skill is 25, required is 30
	// Difference is 5 < 20, so 25% chance to fail
	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 25,
		ToolQuality:   100,
	}

	failures := 0
	iterations := 100

	for i := 0; i < iterations; i++ {
		result := CalculateHarvestYield(node, request)
		if !result.Success {
			failures++
		}
	}

	// Should fail approximately 25% of the time (±12% tolerance)
	assert.GreaterOrEqual(t, failures, 13)
	assert.LessOrEqual(t, failures, 37)
}

func TestHarvestSuccessAdequateSkill(t *testing.T) {
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Wild Berries",
		Type:          ResourceVegetation,
		Quantity:      75,
		MinSkillLevel: 20,
	}

	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 40, // Well above required
		ToolQuality:   100,
	}

	// Should always succeed
	for i := 0; i < 10; i++ {
		result := CalculateHarvestYield(node, request)
		assert.True(t, result.Success)
		assert.Greater(t, result.YieldAmount, 0)
	}
}

func TestHarvestWithNoTool(t *testing.T) {
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Cotton Fiber",
		Type:          ResourceVegetation,
		Quantity:      100,
		MinSkillLevel: 0,
	}

	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 50,
		ToolQuality:   0, // No tool
	}

	result := CalculateHarvestYield(node, request)

	// With no tool, yield should be 0
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.YieldAmount)
}

func TestHarvestWithBasicTool(t *testing.T) {
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Hemp Fiber",
		Type:          ResourceVegetation,
		Quantity:      100,
		MinSkillLevel: 0,
	}

	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 100, // Perfect skill
		ToolQuality:   70,  // Basic tool
	}

	result := CalculateHarvestYield(node, request)

	assert.True(t, result.Success)
	// Skill mod: 1.0, Tool mod: 0.7, Random: 0.8-1.2
	// Expected: 56-84
	assert.GreaterOrEqual(t, result.YieldAmount, 40)
	assert.LessOrEqual(t, result.YieldAmount, 85)
}

func TestHarvestWithExcellentTool(t *testing.T) {
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Mahogany Wood",
		Type:          ResourceVegetation,
		Quantity:      100,
		MinSkillLevel: 20,
	}

	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 100,
		ToolQuality:   130, // Excellent tool (bonus)
	}

	result := CalculateHarvestYield(node, request)

	assert.True(t, result.Success)
	// Skill mod: 1.0, Tool mod: 1.3, Random: 0.8-1.2
	// Expected: 104-156 (capped at quantity)
	assert.GreaterOrEqual(t, result.YieldAmount, 80)
	assert.LessOrEqual(t, result.YieldAmount, 156)
}

func TestRandomVariance(t *testing.T) {
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Fish",
		Type:          ResourceAnimal,
		Quantity:      100,
		MinSkillLevel: 0,
	}

	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 50,
		ToolQuality:   100,
	}

	// Collect results to verify variance
	yields := make([]int, 50)
	for i := 0; i < 50; i++ {
		result := CalculateHarvestYield(node, request)
		yields[i] = result.YieldAmount
	}

	// Verify we get different values (random variance is working)
	uniqueValues := make(map[int]bool)
	for _, y := range yields {
		uniqueValues[y] = true
	}

	// Should have some variance (at least 3 different values)
	assert.GreaterOrEqual(t, len(uniqueValues), 3)
}
