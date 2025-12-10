package resources

import (
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// PlaceResourcesInBiome generates renewable resources for a specific biome area
// area is in km²
func PlaceResourcesInBiome(biomeType string, area float64, seed int64) ([]*ResourceNode, error) {
	r := rand.New(rand.NewSource(seed))
	templates := GetResourceTemplatesForBiome(biomeType)

	var nodes []*ResourceNode

	// Calculate number of nodes based on density targets
	// Each template has implicit density based on how many templates exist for the biome
	// We'll iterate through templates and place them based on area

	for _, tmpl := range templates {
		// Determine density based on rarity
		// Common: 2-5 per km²
		// Uncommon: 0.5-2 per km²
		// Rare: 0.1-0.5 per km²
		// Very Rare: 0.01-0.1 per km²

		var minDensity, maxDensity float64

		switch tmpl.Rarity {
		case RarityCommon:
			minDensity, maxDensity = 1.0, 3.0
		case RarityUncommon:
			minDensity, maxDensity = 0.3, 1.0
		case RarityRare:
			minDensity, maxDensity = 0.05, 0.2
		default:
			minDensity, maxDensity = 0.01, 0.05
		}

		// Adjust for specific types
		if tmpl.Type == ResourceVegetation {
			// Vegetation is slightly denser
			minDensity *= 1.5
			maxDensity *= 1.5
		}

		// Calculate count for this area
		density := minDensity + r.Float64()*(maxDensity-minDensity)
		count := int(math.Round(density * area))

		for i := 0; i < count; i++ {
			// Create primary node
			node := createNodeFromTemplate(tmpl, r)

			// Check for rich deposit upgrade
			if richNode := CreateRichDeposit(tmpl, r.Int63()); richNode != nil {
				// Use rich node properties but keep location
				richNode.LocationX = node.LocationX
				richNode.LocationY = node.LocationY
				richNode.LocationZ = node.LocationZ
				richNode.BiomeAffinity = node.BiomeAffinity
				node = richNode
			}

			nodes = append(nodes, node)

			// Try to create cluster
			if cluster := CreateCluster(node, r.Int63()); cluster != nil {
				nodes = append(nodes, cluster...)
			}
		}
	}

	return nodes, nil
}

func createNodeFromTemplate(tmpl ResourceTemplate, r *rand.Rand) *ResourceNode {
	// Random location within the area (abstracted here as 0-1000m relative coords)
	// In a real system, this would be constrained by the actual biome polygon
	x := r.Float64() * 1000.0
	y := r.Float64() * 1000.0

	quantity := tmpl.MinQuantity + r.Intn(tmpl.MaxQuantity-tmpl.MinQuantity+1)

	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          tmpl.Name,
		Type:          tmpl.Type,
		Rarity:        tmpl.Rarity,
		LocationX:     x,
		LocationY:     y,
		LocationZ:     0, // Surface
		Quantity:      quantity,
		MaxQuantity:   tmpl.MaxQuantity, // Cap at max for template
		RegenRate:     tmpl.RegenRate,
		RegenCooldown: time.Duration(tmpl.CooldownHours) * time.Hour,
		LastHarvested: nil,
		BiomeAffinity: []string{}, // Will be set by caller if needed, or inferred
		RequiredSkill: tmpl.RequiredSkill,
		MinSkillLevel: tmpl.MinSkillLevel,
		CreatedAt:     time.Now(),
	}

	// Set biome affinity based on template
	// In reality, this function is called per biome, so we know the biome
	// But the template might be valid for multiple biomes
	// For now, we'll leave it empty or set it if passed

	return node
}
