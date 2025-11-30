package resources

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMiningYieldCalculation(t *testing.T) {
	// Setup base resource node
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Iron Ore",
		Type:          ResourceMineral,
		Quantity:      100,
		MinSkillLevel: 0,
		Depth:         100.0,
	}

	// Setup mineral deposit reference
	deposit := &MineralDeposit{
		DepositID:     uuid.New(),
		MineralType:   "iron_ore",
		Depth:         100.0,
		Concentration: 1.0, // 100% concentration for baseline
	}

	request := HarvestRequest{
		NodeID:        node.NodeID,
		GathererSkill: 100, // 100% efficiency
		ToolQuality:   100, // 100% efficiency
	}

	// Test baseline
	result := CalculateMiningYield(node, deposit, request)
	assert.True(t, result.Success)

	// Base yield (100) * Skill (1.0) * Tool (1.0) * Random (0.8-1.2)
	// Depth penalty: 1.0 - (100/10000)*0.5 = 0.995
	// Concentration: 1.0
	// Expected: ~99.5 (±20%)
	assert.Greater(t, result.YieldAmount, 70)
	assert.Less(t, result.YieldAmount, 130)
}

func TestMiningDepthPenalty(t *testing.T) {
	node := &ResourceNode{
		NodeID:   uuid.New(),
		Type:     ResourceMineral,
		Quantity: 100,
	}

	// Deep deposit (5000m)
	deposit := &MineralDeposit{
		Depth:         5000.0,
		Concentration: 1.0,
	}

	request := HarvestRequest{
		GathererSkill: 100,
		ToolQuality:   100,
	}

	result := CalculateMiningYield(node, deposit, request)

	// Depth penalty: 1.0 - (5000/10000)*0.5 = 0.75
	// Expected yield should be ~75% of base
	// Base range: 80-120
	// Expected range: 60-90
	assert.Greater(t, result.YieldAmount, 50)
	assert.Less(t, result.YieldAmount, 100)
}

func TestMiningConcentrationBonus(t *testing.T) {
	node := &ResourceNode{
		NodeID:   uuid.New(),
		Type:     ResourceMineral,
		Quantity: 100,
	}

	// Low concentration (0.5)
	deposit := &MineralDeposit{
		Depth:         0.0,
		Concentration: 0.5,
	}

	request := HarvestRequest{
		GathererSkill: 100,
		ToolQuality:   100,
	}

	result := CalculateMiningYield(node, deposit, request)

	// Concentration modifier: 0.5
	// Expected yield should be ~50% of base
	// Base range: 80-120
	// Expected range: 40-60
	assert.Greater(t, result.YieldAmount, 30)
	assert.Less(t, result.YieldAmount, 70)
}

func TestMiningCombinedModifiers(t *testing.T) {
	node := &ResourceNode{
		NodeID:   uuid.New(),
		Type:     ResourceMineral,
		Quantity: 100,
	}

	// Deep (2000m) and low concentration (0.6)
	deposit := &MineralDeposit{
		Depth:         2000.0,
		Concentration: 0.6,
	}

	request := HarvestRequest{
		GathererSkill: 100,
		ToolQuality:   100,
	}

	result := CalculateMiningYield(node, deposit, request)

	// Depth penalty: 1.0 - (2000/10000)*0.5 = 0.9
	// Concentration: 0.6
	// Combined: 0.9 * 0.6 = 0.54
	// Expected range: 43-65 (approx)
	assert.Greater(t, result.YieldAmount, 35)
	assert.Less(t, result.YieldAmount, 75)
}
