package resources

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateResourceNodesFromMinerals(t *testing.T) {
	deposits := []*MineralDeposit{
		{
			DepositID:      uuid.New(),
			MineralType:    "iron_ore",
			LocationX:      100.0,
			LocationY:      200.0,
			Depth:          50.0,
			Quantity:       1000,
			Concentration:  0.75,
			SurfaceVisible: true,
		},
		{
			DepositID:      uuid.New(),
			MineralType:    "coal",
			LocationX:      150.0,
			LocationY:      250.0,
			Depth:          30.0,
			Quantity:       2000,
			Concentration:  0.85,
			SurfaceVisible: false,
		},
	}

	nodes, err := CreateResourceNodesFromMinerals(deposits)

	assert.NoError(t, err)
	assert.Len(t, nodes, 2)

	// Verify first node
	node1 := nodes[0]
	assert.Equal(t, "Iron Ore", node1.Name)
	assert.Equal(t, ResourceMineral, node1.Type)
	assert.Equal(t, deposits[0].DepositID, *node1.MineralDepositID)
	assert.Equal(t, deposits[0].LocationX, node1.LocationX)
	assert.Equal(t, deposits[0].LocationY, node1.LocationY)
	assert.Equal(t, deposits[0].Quantity, node1.Quantity)
	assert.Equal(t, deposits[0].Quantity, node1.MaxQuantity)
	assert.Equal(t, deposits[0].Depth, node1.Depth)
	assert.Equal(t, 0.0, node1.RegenRate) // Minerals don't regenerate
	assert.Equal(t, "mining", node1.RequiredSkill)
	assert.True(t, node1.MinSkillLevel > 0)

	// Verify second node
	node2 := nodes[1]
	assert.Equal(t, "Coal", node2.Name)
	assert.Equal(t, deposits[1].DepositID, *node2.MineralDepositID)
}

func TestMineralNoDuplication(t *testing.T) {
	// This test ensures we're creating references, not duplicating mineral data
	deposit := &MineralDeposit{
		DepositID:     uuid.New(),
		MineralType:   "gold_ore",
		LocationX:     300.0,
		LocationY:     400.0,
		Depth:         100.0,
		Quantity:      500,
		Concentration: 0.95,
	}

	nodes, err := CreateResourceNodesFromMinerals([]*MineralDeposit{deposit})

	assert.NoError(t, err)
	assert.Len(t, nodes, 1)

	node := nodes[0]
	// Verify it's a reference, not a copy
	assert.NotNil(t, node.MineralDepositID)
	assert.Equal(t, deposit.DepositID, *node.MineralDepositID)
	assert.Nil(t, node.SpeciesID) // Should not have species reference
}

func TestMineralInheritProperties(t *testing.T) {
	deposit := &MineralDeposit{
		DepositID:     uuid.New(),
		MineralType:   "diamond",
		LocationX:     500.0,
		LocationY:     600.0,
		Depth:         200.0,
		Quantity:      100,
		Concentration: 0.99,
	}

	nodes, err := CreateResourceNodesFromMinerals([]*MineralDeposit{deposit})

	assert.NoError(t, err)
	node := nodes[0]

	// Verify inherited properties
	assert.Equal(t, deposit.Quantity, node.Quantity, "Quantity should be inherited")
	assert.Equal(t, deposit.Quantity, node.MaxQuantity, "MaxQuantity should equal initial Quantity")
	assert.Equal(t, deposit.Depth, node.Depth, "Depth should be inherited")
	assert.Equal(t, deposit.LocationX, node.LocationX, "LocationX should be inherited")
	assert.Equal(t, deposit.LocationY, node.LocationY, "LocationY should be inherited")

	// Verify mineral-specific properties
	assert.Equal(t, 0.0, node.RegenRate, "Minerals should not regenerate")
	assert.Equal(t, "mining", node.RequiredSkill)
}

func TestMapMineralToResourceType(t *testing.T) {
	tests := []struct {
		mineralType  string
		expectedName string
	}{
		{"iron_ore", "Iron Ore"},
		{"coal", "Coal"},
		{"gold_ore", "Gold Ore"},
		{"copper_ore", "Copper Ore"},
		{"silver_ore", "Silver Ore"},
		{"diamond", "Diamond"},
		{"emerald", "Emerald"},
		{"ruby", "Ruby"},
		{"sapphire", "Sapphire"},
	}

	for _, tt := range tests {
		t.Run(tt.mineralType, func(t *testing.T) {
			name := MapMineralToResourceName(tt.mineralType)
			assert.Equal(t, tt.expectedName, name)
		})
	}
}

func TestDetermineSkillRequirement(t *testing.T) {
	tests := []struct {
		name          string
		depth         float64
		concentration float64
		expectedMin   int
		expectedMax   int
	}{
		{"Shallow high-grade", 10.0, 0.9, 0, 20},
		{"Medium depth", 50.0, 0.7, 10, 40},
		{"Deep low-grade", 150.0, 0.3, 60, 80},
		{"Very deep", 250.0, 0.5, 70, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill := DetermineSkillRequirement(tt.depth, tt.concentration)
			assert.GreaterOrEqual(t, skill, tt.expectedMin)
			assert.LessOrEqual(t, skill, tt.expectedMax)
		})
	}
}

func TestMineralRarity(t *testing.T) {
	tests := []struct {
		mineralType    string
		expectedRarity Rarity
	}{
		{"iron_ore", RarityCommon},
		{"coal", RarityCommon},
		{"copper_ore", RarityCommon},
		{"gold_ore", RarityUncommon},
		{"silver_ore", RarityUncommon},
		{"diamond", RarityRare},
		{"emerald", RarityRare},
		{"ruby", RarityVeryRare},
		{"sapphire", RarityVeryRare},
	}

	for _, tt := range tests {
		t.Run(tt.mineralType, func(t *testing.T) {
			rarity := DetermineMineralRarity(tt.mineralType)
			assert.Equal(t, tt.expectedRarity, rarity)
		})
	}
}
