package minerals

import (
	"math/rand"
)

// CalculateConcentration determines the ore grade
func CalculateConcentration(mineralType MineralType, context *TectonicContext) float64 {
	// Base concentration ranges by mineral rarity/type
	// We can infer rarity from BaseValue or hardcode for specific types
	var minBase, maxBase float64

	switch mineralType.Name {
	case "Iron", "Coal", "Limestone", "Salt", "Basalt", "Granite":
		minBase, maxBase = 0.6, 0.9
	case "Copper", "Tin", "Marble":
		minBase, maxBase = 0.4, 0.7
	case "Gold", "Silver", "Ruby", "Sapphire", "Emerald":
		minBase, maxBase = 0.1, 0.4
	case "Diamond", "Platinum":
		minBase, maxBase = 0.01, 0.1
	default:
		minBase, maxBase = 0.3, 0.6
	}

	baseConcentration := minBase + rand.Float64()*(maxBase-minBase)

	// Geological modifier
	geoMod := 1.0
	// Optimal conditions
	if (mineralType.FormationType == FormationIgneous && context.IsVolcanic) ||
		(mineralType.FormationType == FormationSedimentary && context.IsSedimentaryBasin) {
		geoMod = 1.0 + rand.Float64()*0.5 // 1.0 to 1.5
	} else {
		// Average to poor
		geoMod = 0.3 + rand.Float64()*0.7 // 0.3 to 1.0
	}

	// Random variation
	randomVar := 0.8 + rand.Float64()*0.4 // 0.8 to 1.2

	concentration := baseConcentration * geoMod * randomVar

	// Cap at 1.0
	if concentration > 1.0 {
		concentration = 1.0
	}
	return concentration
}

// GetBaseQuantity returns the base unit quantity for a given mineral and size
func GetBaseQuantity(mineralType MineralType, size VeinSize) int {
	base := 0
	switch size {
	case VeinSizeSmall:
		base = 1000
	case VeinSizeMedium:
		base = 5000
	case VeinSizeLarge:
		base = 20000
	case VeinSizeMassive:
		base = 100000
	}

	// Adjust for rarity
	// Rare minerals have smaller deposits even at "Large" size relative to common ones
	multiplier := 1.0
	switch mineralType.Name {
	case "Iron", "Coal", "Salt", "Limestone":
		multiplier = 1.0
	case "Copper":
		multiplier = 0.5
	case "Gold", "Silver":
		multiplier = 0.1
	case "Diamond", "Platinum", "Ruby", "Sapphire", "Emerald":
		multiplier = 0.01
	}

	return int(float64(base) * multiplier)
}
