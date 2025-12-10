package resources

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreateResourceNodesFromMinerals creates ResourceNode references from Phase 8.2b mineral deposits
// This does NOT create new mineral deposits - it only creates references to existing ones
func CreateResourceNodesFromMinerals(deposits []*MineralDeposit) ([]*ResourceNode, error) {
	nodes := make([]*ResourceNode, 0, len(deposits))

	for _, deposit := range deposits {
		node := &ResourceNode{
			NodeID:           uuid.New(),
			Name:             MapMineralToResourceName(deposit.MineralType),
			Type:             ResourceMineral,
			Rarity:           DetermineMineralRarity(deposit.MineralType),
			LocationX:        deposit.LocationX,
			LocationY:        deposit.LocationY,
			LocationZ:        -deposit.Depth, // Negative Z for underground
			Quantity:         deposit.Quantity,
			MaxQuantity:      deposit.Quantity,
			RegenRate:        0.0, // Minerals don't regenerate
			RegenCooldown:    0,
			LastHarvested:    nil,
			BiomeAffinity:    []string{}, // Minerals determined by geology, not biome
			RequiredSkill:    "mining",
			MinSkillLevel:    DetermineSkillRequirement(deposit.Depth, deposit.Concentration),
			MineralDepositID: &deposit.DepositID, // Reference to Phase 8.2b deposit
			Depth:            deposit.Depth,
			SpeciesID:        nil,
			CreatedAt:        time.Now(),
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// MapMineralToResourceName converts mineral type codes to display names
func MapMineralToResourceName(mineralType string) string {
	// Convert snake_case to Title Case
	parts := strings.Split(mineralType, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// DetermineMineralRarity assigns rarity based on mineral type
func DetermineMineralRarity(mineralType string) Rarity {
	switch mineralType {
	case "iron_ore", "coal", "copper_ore", "stone", "tin_ore":
		return RarityCommon
	case "gold_ore", "silver_ore", "platinum_ore", "mithril_ore":
		return RarityUncommon
	case "diamond", "emerald", "adamantite_ore":
		return RarityRare
	case "ruby", "sapphire", "orichalcum_ore":
		return RarityVeryRare
	case "mythril", "starmetal", "void_crystal":
		return RarityLegendary
	default:
		return RarityCommon
	}
}

// DetermineSkillRequirement calculates required mining skill based on depth and ore grade
func DetermineSkillRequirement(depth, concentration float64) int {
	// Base skill from depth
	// 0-50m: 0-20 skill
	// 50-100m: 20-40 skill
	// 100-200m: 40-70 skill
	// 200m+: 70-100 skill
	depthFactor := depth / 3.0
	if depthFactor > 80 {
		depthFactor = 80
	}

	// Concentration modifier
	// High concentration (0.8-1.0): -10 skill
	// Medium concentration (0.5-0.8): no change
	// Low concentration (0.0-0.5): +10 skill
	concentrationModifier := 0
	if concentration >= 0.8 {
		concentrationModifier = -10
	} else if concentration < 0.5 {
		concentrationModifier = 10
	}

	skill := int(depthFactor) + concentrationModifier

	// Clamp to 0-100
	if skill < 0 {
		skill = 0
	}
	if skill > 100 {
		skill = 100
	}

	return skill
}
