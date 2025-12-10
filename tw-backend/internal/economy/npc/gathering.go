package npc

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"mud-platform-backend/internal/economy/resources"

	"github.com/google/uuid"
)

// Errors
var (
	ErrNoSuitableResources = errors.New("no suitable resources found")
	ErrInventoryFull       = errors.New("inventory full")
)

// ResourceFinder interface for dependency injection
type ResourceFinder interface {
	FindNearbyNodes(location uuid.UUID, radius float64) ([]*resources.ResourceNode, error)
}

// Harvester interface for dependency injection
type Harvester interface {
	Harvest(nodeID uuid.UUID, skillLevel int, toolQuality float64) (*resources.HarvestResult, error)
}

// InventoryManager interface for dependency injection
type InventoryManager interface {
	Add(itemID uuid.UUID, quantity int) error
	IsFull() bool
	GetContents() map[uuid.UUID]int
}

// NPC represents the agent performing the gathering
// This is a simplified struct for the economy package; the full NPC struct would be in the entities package
type NPC struct {
	ID         uuid.UUID
	Location   uuid.UUID
	Occupation Occupation
	Skills     map[string]int
	Inventory  InventoryManager
	Desires    *Desires
	Equipment  *Equipment
}

type Desires struct {
	ResourceAcquisition float64
	TaskCompletion      float64
}

type Equipment struct {
	ToolQuality float64
}

// GatherResources executes the autonomous gathering loop
func (npc *NPC) GatherResources(ctx context.Context, finder ResourceFinder, harvester Harvester) error {
	// 1. Determine gathering need
	if npc.Desires.ResourceAcquisition < 50 {
		return nil // Not urgent enough
	}

	// 2. Find nearby harvestable resources
	nodes, err := finder.FindNearbyNodes(npc.Location, npc.Occupation.GatheringRadius)
	if err != nil {
		return fmt.Errorf("failed to find resources: %w", err)
	}

	// 3. Filter and prioritize
	target := npc.selectBestResource(nodes)
	if target == nil {
		return ErrNoSuitableResources
	}

	// 4. Path to resource (abstracted for this module)
	// In a full implementation, this would involve movement logic
	// For now, we assume they can reach it if it's in radius

	// 5. Harvest
	// Determine relevant skill
	skillLevel := 0
	// Simplified skill mapping - in reality would check resource type -> skill
	if len(npc.Occupation.PreferredSkills) > 0 {
		skillName := npc.Occupation.PreferredSkills[0]
		skillLevel = npc.Skills[skillName]
	}

	yield, err := harvester.Harvest(target.NodeID, skillLevel, npc.Equipment.ToolQuality)
	if err != nil {
		return fmt.Errorf("harvest failed: %w", err)
	}

	// 6. Add to inventory
	// Assuming the harvested item ID corresponds to the resource node ID or a lookup
	// For now, we'll use the node ID as a placeholder for the item ID
	if yield.Success && yield.YieldAmount > 0 {
		if err := npc.Inventory.Add(target.NodeID, yield.YieldAmount); err != nil {
			return ErrInventoryFull
		}
	}

	// 7. Decrease resource acquisition need
	npc.Desires.ResourceAcquisition -= 20.0
	if npc.Desires.ResourceAcquisition < 0 {
		npc.Desires.ResourceAcquisition = 0
	}

	return nil
}

func (npc *NPC) selectBestResource(nodes []*resources.ResourceNode) *resources.ResourceNode {
	var candidates []*resources.ResourceNode

	// Filter by occupation preference
	for _, node := range nodes {
		if npc.isResourcePreferred(node) {
			candidates = append(candidates, node)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by priority (simplified: just take the first one for now, or random)
	// Real logic would weigh distance vs value vs quantity
	// For this implementation, we'll sort by quantity descending as a proxy for value
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Quantity > candidates[j].Quantity
	})

	return candidates[0]
}

func (npc *NPC) isResourcePreferred(node *resources.ResourceNode) bool {
	category := node.GetCategory()

	// Check primary
	for _, t := range npc.Occupation.PrimaryResources {
		if t == category {
			return true
		}
	}
	// Check secondary
	for _, t := range npc.Occupation.SecondaryResources {
		if t == category {
			return true
		}
	}
	return false
}
