package population

import (
	"fmt"
	"strings"

	"tw-backend/internal/worldgen/geography"
)

// CoveringType represents the body covering of a species
type CoveringType string

const (
	CoveringFur      CoveringType = "fur"
	CoveringScales   CoveringType = "scales"
	CoveringFeathers CoveringType = "feathers"
	CoveringShell    CoveringType = "shell"
	CoveringSkin     CoveringType = "skin"
	CoveringBark     CoveringType = "bark"
	CoveringNone     CoveringType = "none"
)

// FloraGrowthType represents the growth pattern of flora
type FloraGrowthType string

const (
	FloraEvergreen FloraGrowthType = "evergreen" // Keeps leaves year-round
	FloraDeciduous FloraGrowthType = "deciduous" // Drops leaves seasonally
	FloraAnnual    FloraGrowthType = "annual"    // Lives one season, reseeds
	FloraPerennial FloraGrowthType = "perennial" // Lives multiple years
	FloraSucculent FloraGrowthType = "succulent" // Stores water (desert adapted)
	FloraAquatic   FloraGrowthType = "aquatic"   // Lives in water
)

// GenerateSpeciesName creates a descriptive name from traits, diet, and biome
func GenerateSpeciesName(traits EvolvableTraits, diet DietType, biome geography.BiomeType) string {
	var parts []string

	if diet == DietPhotosynthetic {
		// Flora naming
		parts = generateFloraName(traits, biome)
	} else {
		// Fauna naming
		parts = generateFaunaName(traits, diet, biome)
	}

	return strings.Join(parts, " ")
}

// generateFloraName creates names for photosynthetic species
func generateFloraName(traits EvolvableTraits, biome geography.BiomeType) []string {
	var parts []string

	// Size descriptor
	switch {
	case traits.Size < 0.2:
		parts = append(parts, "Microscopic")
	case traits.Size < 0.5:
		parts = append(parts, "Tiny")
	case traits.Size < 1.0:
		parts = append(parts, "Small")
	case traits.Size < 3.0:
		// No size descriptor for medium
	case traits.Size < 5.0:
		parts = append(parts, "Tall")
	case traits.Size < 8.0:
		parts = append(parts, "Towering")
	default:
		parts = append(parts, "Giant")
	}

	// Growth type descriptor
	switch traits.FloraGrowth {
	case FloraEvergreen:
		parts = append(parts, "Evergreen")
	case FloraDeciduous:
		parts = append(parts, "Broadleaf")
	case FloraAnnual:
		parts = append(parts, "Seasonal")
	case FloraPerennial:
		parts = append(parts, "Hardy")
	case FloraSucculent:
		parts = append(parts, "Succulent")
	case FloraAquatic:
		if traits.Size < 0.2 {
			parts = append(parts, "Plankton")
			return parts // Early return for plankton
		}
		parts = append(parts, "Aquatic")
	default:
		// Default based on biome
		switch biome {
		case geography.BiomeTaiga, geography.BiomeAlpine:
			parts = append(parts, "Evergreen")
		case geography.BiomeDeciduousForest:
			parts = append(parts, "Broadleaf")
		case geography.BiomeDesert:
			parts = append(parts, "Succulent")
		case geography.BiomeOcean:
			if traits.Size < 0.2 {
				parts = append(parts, "Plankton")
				return parts
			}
			parts = append(parts, "Kelp")
			return parts
		case geography.BiomeRainforest:
			parts = append(parts, "Tropical")
		case geography.BiomeGrassland:
			parts = append(parts, "Prairie")
		default:
			parts = append(parts, "Wild")
		}
	}

	// Final type
	if traits.Size > 3.0 && traits.Covering == CoveringBark {
		parts = append(parts, "Tree")
	} else if traits.Size > 1.0 {
		parts = append(parts, "Shrub")
	} else if biome == geography.BiomeGrassland {
		parts = append(parts, "Grass")
	} else {
		parts = append(parts, "Plant")
	}

	return parts
}

// generateFaunaName creates names for animals
func generateFaunaName(traits EvolvableTraits, diet DietType, biome geography.BiomeType) []string {
	var parts []string

	// Size descriptor - more variety
	switch {
	case traits.Size < 0.3:
		parts = append(parts, pickFrom([]string{"Tiny", "Miniature", "Pygmy"}))
	case traits.Size < 1.0:
		parts = append(parts, pickFrom([]string{"Small", "Lesser", "Dwarf"}))
	case traits.Size < 2.0:
		// No size descriptor for medium - use other descriptors
	case traits.Size < 4.0:
		parts = append(parts, pickFrom([]string{"Large", "Greater", "Robust"}))
	case traits.Size < 7.0:
		parts = append(parts, pickFrom([]string{"Massive", "Grand", "Imperial"}))
	default:
		parts = append(parts, pickFrom([]string{"Giant", "Colossal", "Titanic"}))
	}

	// Behavior/trait descriptor (only one)
	descriptorAdded := false
	if traits.Speed > 7.0 && !descriptorAdded {
		parts = append(parts, pickFrom([]string{"Swift", "Dashing", "Fleet"}))
		descriptorAdded = true
	} else if traits.Intelligence > 0.7 && !descriptorAdded {
		parts = append(parts, pickFrom([]string{"Cunning", "Clever", "Sly"}))
		descriptorAdded = true
	} else if traits.Strength > 7.0 && !descriptorAdded {
		parts = append(parts, pickFrom([]string{"Mighty", "Powerful", "Stalwart"}))
		descriptorAdded = true
	} else if traits.Aggression > 0.7 && !descriptorAdded {
		parts = append(parts, pickFrom([]string{"Fierce", "Savage", "Brutal"}))
		descriptorAdded = true
	} else if traits.Camouflage > 0.7 && !descriptorAdded {
		parts = append(parts, pickFrom([]string{"Shadow", "Phantom", "Ghost"}))
		descriptorAdded = true
	}

	// Covering descriptor (only if no behavior descriptor)
	if !descriptorAdded || len(parts) < 2 {
		switch traits.Covering {
		case CoveringFur:
			parts = append(parts, pickFrom([]string{"Woolly", "Furred", "Shaggy"}))
		case CoveringScales:
			parts = append(parts, pickFrom([]string{"Scaled", "Plated", "Armored"}))
		case CoveringFeathers:
			parts = append(parts, pickFrom([]string{"Feathered", "Plumed", "Crested"}))
		case CoveringShell:
			parts = append(parts, pickFrom([]string{"Armored", "Carapaced", "Shelled"}))
		case CoveringSkin:
			parts = append(parts, pickFrom([]string{"Smooth", "Sleek", "Bare"}))
		}
	}

	// Creature type based on diet and traits
	switch diet {
	case DietHerbivore:
		parts = append(parts, getHerbivoreType(traits, biome))
	case DietCarnivore:
		parts = append(parts, getCarnivoreType(traits, biome))
	case DietOmnivore:
		parts = append(parts, getOmnivoreType(traits, biome))
	}

	return parts
}

// getHerbivoreType returns a creature type name based on traits
func getHerbivoreType(traits EvolvableTraits, biome geography.BiomeType) string {
	// Size and speed determine creature type
	if biome == geography.BiomeOcean {
		return pickFrom([]string{"Grazer", "Browser", "Drifter"})
	}

	if traits.Size > 5.0 {
		return pickFrom([]string{"Behemoth", "Titan", "Mammoth", "Auroch"})
	}
	if traits.Size > 3.0 {
		return pickFrom([]string{"Elk", "Stag", "Bison", "Ox"})
	}
	if traits.Speed > 6.0 {
		return pickFrom([]string{"Antelope", "Gazelle", "Sprinter", "Dasher"})
	}
	if traits.Social > 0.7 {
		return pickFrom([]string{"Grazer", "Browser", "Forager"})
	}

	return pickFrom([]string{"Browser", "Grazer", "Muncher", "Nibbler"})
}

// getCarnivoreType returns a predator type name based on traits
func getCarnivoreType(traits EvolvableTraits, biome geography.BiomeType) string {
	if biome == geography.BiomeOcean {
		if traits.Size > 5.0 {
			return pickFrom([]string{"Leviathan", "Kraken", "Behemoth"})
		}
		return pickFrom([]string{"Hunter", "Stalker", "Predator"})
	}

	if traits.VenomPotency > 0.5 {
		return pickFrom([]string{"Viper", "Fang", "Venomjaw"})
	}
	if traits.Size > 5.0 {
		if traits.Social > 0.5 {
			return pickFrom([]string{"Tyrant", "Apex", "Dominator"})
		}
		return pickFrom([]string{"Rex", "Terror", "Ravager"})
	}
	if traits.Speed > 6.0 {
		return pickFrom([]string{"Stalker", "Chaser", "Pursuer"})
	}
	if traits.Social > 0.7 {
		return pickFrom([]string{"Packwolf", "Reaver", "Raider"})
	}
	if traits.Intelligence > 0.7 {
		return pickFrom([]string{"Hunter", "Ambusher", "Lurker"})
	}

	return pickFrom([]string{"Predator", "Hunter", "Prowler", "Slayer"})
}

// getOmnivoreType returns a creature type for omnivores
func getOmnivoreType(traits EvolvableTraits, biome geography.BiomeType) string {
	if biome == geography.BiomeOcean {
		return pickFrom([]string{"Scavenger", "Opportunist", "Forager"})
	}

	if traits.Size > 4.0 {
		return pickFrom([]string{"Ursoid", "Brute", "Marauder"})
	}
	if traits.Social > 0.7 {
		return pickFrom([]string{"Troop", "Band", "Clan"})
	}
	if traits.Intelligence > 0.6 {
		return pickFrom([]string{"Trickster", "Scrounger", "Seeker"})
	}

	return pickFrom([]string{"Forager", "Roamer", "Wanderer", "Scavenger"})
}

// pickFrom returns a deterministic selection based on the first character's hash
func pickFrom(options []string) string {
	if len(options) == 0 {
		return ""
	}
	// Use a simple deterministic selection for consistency
	return options[0]
}

// DescribePopulation creates a natural language description of a population count
func DescribePopulation(count int64, name string) string {
	// Pluralize simple names (basic rules)
	plural := pluralize(name)

	switch {
	case count == 1:
		return fmt.Sprintf("A lone %s", name)
	case count <= 10:
		return fmt.Sprintf("A few %s", plural)
	case count <= 50:
		return fmt.Sprintf("A small herd of %s", name)
	case count <= 200:
		return fmt.Sprintf("A herd of %s", name)
	case count <= 500:
		return fmt.Sprintf("A large herd of %s", name)
	case count <= 1000:
		return fmt.Sprintf("A thriving swarm of %s", plural)
	case count <= 5000:
		return fmt.Sprintf("A massive forest of %s", name)
	default:
		return fmt.Sprintf("An ecosystem dominated by %s", name)
	}
}

// pluralize adds basic plural suffix to a name
func pluralize(name string) string {
	if len(name) == 0 {
		return name
	}
	lastChar := name[len(name)-1]
	switch lastChar {
	case 's', 'x', 'z':
		return name + "es"
	case 'y':
		if len(name) > 1 {
			return name[:len(name)-1] + "ies"
		}
		return name + "s"
	default:
		return name + "s"
	}
}

// GetCoveringForDiet returns a default covering type based on diet and biome
func GetCoveringForDiet(diet DietType, biome geography.BiomeType) CoveringType {
	switch diet {
	case DietPhotosynthetic:
		if biome == geography.BiomeOcean {
			return CoveringNone
		}
		return CoveringBark
	case DietHerbivore, DietOmnivore:
		switch biome {
		case geography.BiomeTundra, geography.BiomeAlpine, geography.BiomeTaiga:
			return CoveringFur
		case geography.BiomeDesert:
			return CoveringScales
		case geography.BiomeOcean:
			return CoveringScales
		default:
			return CoveringFur
		}
	case DietCarnivore:
		switch biome {
		case geography.BiomeTundra, geography.BiomeAlpine, geography.BiomeTaiga:
			return CoveringFur
		case geography.BiomeDesert:
			return CoveringScales
		case geography.BiomeOcean:
			return CoveringScales
		default:
			return CoveringFur
		}
	default:
		return CoveringSkin
	}
}

// GetFloraGrowthForBiome returns appropriate flora growth type for a biome
func GetFloraGrowthForBiome(biome geography.BiomeType) FloraGrowthType {
	switch biome {
	case geography.BiomeTaiga, geography.BiomeAlpine:
		return FloraEvergreen
	case geography.BiomeDeciduousForest:
		return FloraDeciduous
	case geography.BiomeDesert:
		return FloraSucculent
	case geography.BiomeOcean:
		return FloraAquatic
	case geography.BiomeGrassland:
		return FloraPerennial
	case geography.BiomeRainforest:
		return FloraEvergreen // Tropical broadleaf evergreen
	case geography.BiomeTundra:
		return FloraPerennial // Hardy tundra plants
	default:
		return FloraPerennial
	}
}
