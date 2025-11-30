package crafting

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WorldConfiguration represents the tech/magic settings for a world
// This mirrors the structure from Phase 8.1
type WorldConfiguration struct {
	WorldID      uuid.UUID
	TechLevel    TechLevel
	MagicLevel   string // "none", "rare", "common", "dominant"
	UniqueAspect string
}

// CustomizeTechTree generates a world-specific tech tree based on configuration
func CustomizeTechTree(config *WorldConfiguration, allNodes map[TechLevel][]*TechNode) *TechTree {
	tree := &TechTree{
		TreeID:    uuid.New(),
		WorldID:   config.WorldID,
		Name:      fmt.Sprintf("Tech Tree for %s", config.WorldID),
		TechLevel: config.TechLevel,
		Nodes:     []*TechNode{},
		CreatedAt: time.Now(),
	}

	// 1. Filter nodes based on world tech level
	maxLevel := config.TechLevel
	levels := map[TechLevel]int{
		TechPrimitive:  1,
		TechMedieval:   2,
		TechIndustrial: 3,
		TechModern:     4,
		TechFuturistic: 5,
	}
	maxVal := levels[maxLevel]

	for level, nodes := range allNodes {
		if levels[level] <= maxVal {
			// Add nodes from this level
			// We need to deep copy nodes to modify them for this world if needed
			for _, node := range nodes {
				// Create a copy
				newNode := *node
				newNode.TreeID = tree.TreeID
				tree.Nodes = append(tree.Nodes, &newNode)
			}
		}
	}

	// 2. Apply magic interactions
	applyMagicInteractions(tree, config.MagicLevel)

	return tree
}

func applyMagicInteractions(tree *TechTree, magicLevel string) {
	switch magicLevel {
	case "dominant":
		// Magic replaces complex tech
		// For example, remove electricity and replace with mana
		replaceTechWithMagic(tree)
	case "common":
		// Magic enhances tech (Magitech)
		enhanceTechWithMagic(tree)
	case "rare":
		// Tech is dominant, magic is rare
	case "none":
		// Pure tech
	}
}

func replaceTechWithMagic(tree *TechTree) {
	// Example: Replace "Steam Engine" with "Mana Engine"
	// In a real implementation, this would be data-driven
	for _, node := range tree.Nodes {
		if node.Name == "Steam Engine" {
			node.Name = "Mana Engine"
			node.Description = "A device that converts raw mana into mechanical work."
			// Remove fuel requirements, add mana requirement
		}
		if node.Name == "Electricity" {
			node.Name = "Arcane Power"
			node.Description = "Channeling lightning energy through crystal conduits."
		}
	}
}

func enhanceTechWithMagic(tree *TechTree) {
	// Example: "Steel Production" becomes "Mithril Infusion" or similar
	// Or just reduce research times
	for _, node := range tree.Nodes {
		// Reduce research time by 25% due to magical aid
		node.ResearchTime = time.Duration(float64(node.ResearchTime) * 0.75)
	}
}
