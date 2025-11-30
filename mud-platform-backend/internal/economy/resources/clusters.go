package resources

import (
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// CreateCluster generates additional nodes around a primary node
func CreateCluster(primaryNode *ResourceNode, seed int64) []*ResourceNode {
	r := rand.New(rand.NewSource(seed))

	// 60% chance to form a cluster
	if r.Float64() > 0.6 {
		return nil
	}

	numSecondary := 2 + r.Intn(4) // 2-5 additional nodes
	cluster := make([]*ResourceNode, numSecondary)

	for i := 0; i < numSecondary; i++ {
		// Place within 50-200m of primary
		angle := r.Float64() * 2 * math.Pi
		distance := 50.0 + r.Float64()*150.0

		offsetX := distance * math.Cos(angle)
		offsetY := distance * math.Sin(angle)

		// Create copy of primary node
		node := &ResourceNode{
			NodeID:           uuid.New(),
			Name:             primaryNode.Name,
			Type:             primaryNode.Type,
			Rarity:           primaryNode.Rarity,
			LocationX:        primaryNode.LocationX + offsetX,
			LocationY:        primaryNode.LocationY + offsetY,
			LocationZ:        primaryNode.LocationZ, // Same elevation roughly
			Quantity:         primaryNode.Quantity,
			MaxQuantity:      primaryNode.MaxQuantity,
			RegenRate:        primaryNode.RegenRate,
			RegenCooldown:    primaryNode.RegenCooldown,
			LastHarvested:    nil,
			BiomeAffinity:    make([]string, len(primaryNode.BiomeAffinity)),
			RequiredSkill:    primaryNode.RequiredSkill,
			MinSkillLevel:    primaryNode.MinSkillLevel,
			SpeciesID:        primaryNode.SpeciesID,
			MineralDepositID: nil, // Clusters don't share mineral deposit ID (minerals don't cluster this way)
			CreatedAt:        time.Now(),
		}

		copy(node.BiomeAffinity, primaryNode.BiomeAffinity)
		cluster[i] = node
	}

	return cluster
}

// CreateRichDeposit checks if a node should be upgraded to a rich source
func CreateRichDeposit(template ResourceTemplate, seed int64) *ResourceNode {
	r := rand.New(rand.NewSource(seed))

	// 5% chance for rich sources
	if r.Float64() > 0.05 {
		return nil
	}

	// Create rich node
	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          template.Name, // Could append "Rich " prefix if desired
		Type:          template.Type,
		Rarity:        RarityUncommon, // Always at least uncommon
		Quantity:      template.MaxQuantity * 3,
		MaxQuantity:   template.MaxQuantity * 3,
		RegenRate:     template.RegenRate * 2,
		RegenCooldown: time.Duration(template.CooldownHours) * time.Hour,
		RequiredSkill: template.RequiredSkill,
		MinSkillLevel: template.MinSkillLevel + 10, // Harder to harvest
		CreatedAt:     time.Now(),
	}

	// Ensure rarity is upgraded if it was common
	if template.Rarity != RarityCommon {
		node.Rarity = template.Rarity
	}

	return node
}
