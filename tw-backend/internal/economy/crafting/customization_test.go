package crafting

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCustomizeTechTree_Filtering(t *testing.T) {
	// Setup Nodes
	allNodes := map[TechLevel][]*TechNode{
		TechPrimitive: {
			{Name: "Fire"},
		},
		TechMedieval: {
			{Name: "Iron Working"},
		},
		TechIndustrial: {
			{Name: "Steam Engine"},
		},
	}

	config := &WorldConfiguration{
		WorldID:    uuid.New(),
		TechLevel:  TechMedieval,
		MagicLevel: "none",
	}

	tree := CustomizeTechTree(config, allNodes)

	assert.Equal(t, TechMedieval, tree.TechLevel)
	// Should include Primitive and Medieval, but NOT Industrial
	assert.Len(t, tree.Nodes, 2)

	hasFire := false
	hasIron := false
	hasSteam := false
	for _, n := range tree.Nodes {
		if n.Name == "Fire" {
			hasFire = true
		}
		if n.Name == "Iron Working" {
			hasIron = true
		}
		if n.Name == "Steam Engine" {
			hasSteam = true
		}
	}
	assert.True(t, hasFire)
	assert.True(t, hasIron)
	assert.False(t, hasSteam)
}

func TestCustomizeTechTree_MagicDominant(t *testing.T) {
	allNodes := map[TechLevel][]*TechNode{
		TechIndustrial: {
			{Name: "Steam Engine", Description: "Fuel based"},
			{Name: "Electricity", Description: "Electron based"},
		},
	}

	config := &WorldConfiguration{
		WorldID:    uuid.New(),
		TechLevel:  TechIndustrial,
		MagicLevel: "dominant",
	}

	tree := CustomizeTechTree(config, allNodes)

	for _, n := range tree.Nodes {
		if n.Name == "Mana Engine" {
			assert.Contains(t, n.Description, "mana")
		}
		if n.Name == "Arcane Power" {
			assert.Contains(t, n.Description, "crystal")
		}
	}
}

func TestCustomizeTechTree_MagicCommon(t *testing.T) {
	allNodes := map[TechLevel][]*TechNode{
		TechMedieval: {
			{Name: "Steel", ResearchTime: 100 * time.Minute},
		},
	}

	config := &WorldConfiguration{
		WorldID:    uuid.New(),
		TechLevel:  TechMedieval,
		MagicLevel: "common",
	}

	tree := CustomizeTechTree(config, allNodes)

	// Should be reduced by 25% => 75 minutes
	assert.Len(t, tree.Nodes, 1)
	assert.Equal(t, 75*time.Minute, tree.Nodes[0].ResearchTime)
}
