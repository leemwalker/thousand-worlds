package resources

// GetResourceTemplatesForBiome returns the resource templates available in a given biome
func GetResourceTemplatesForBiome(biomeType string) []ResourceTemplate {
	switch biomeType {
	case "Deciduous Forest", "Taiga":
		return forestResources
	case "Grassland":
		return grasslandResources
	case "Desert":
		return desertResources
	case "Ocean":
		return oceanResources
	case "Tundra":
		return tundraResources
	case "Rainforest":
		return rainforestResources
	case "Alpine", "Mountain", "High Mountain":
		// Mountains have minerals from Phase 8.2b, minimal vegetation
		return mountainResources
	default:
		return []ResourceTemplate{}
	}
}

// Forest biome resources (10-20 wood/km², 5-10 herbs/km², 1-3 game/km²)
var forestResources = []ResourceTemplate{
	{
		Name:          "Oak Wood",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   100,
		MaxQuantity:   500,
		RegenRate:     2.0, // Slow-growing trees
		CooldownHours: 168, // 7 days
		RequiredSkill: "gathering",
		MinSkillLevel: 0,
	},
	{
		Name:          "Birch Wood",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   80,
		MaxQuantity:   400,
		RegenRate:     2.5,
		CooldownHours: 168,
		RequiredSkill: "gathering",
		MinSkillLevel: 5,
	},
	{
		Name:          "Maple Wood",
		Type:          ResourceVegetation,
		Rarity:        RarityUncommon,
		MinQuantity:   100,
		MaxQuantity:   300,
		RegenRate:     1.5,
		CooldownHours: 168,
		RequiredSkill: "gathering",
		MinSkillLevel: 15,
	},
	{
		Name:          "Medicinal Herbs",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   20,
		MaxQuantity:   100,
		RegenRate:     15.0, // Fast-growing
		CooldownHours: 24,   // 1 day
		RequiredSkill: "gathering",
		MinSkillLevel: 10,
	},
	{
		Name:          "Wild Berries",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   30,
		MaxQuantity:   150,
		RegenRate:     20.0,
		CooldownHours: 24,
		RequiredSkill: "gathering",
		MinSkillLevel: 0,
	},
	{
		Name:          "Deer Hide",
		Type:          ResourceAnimal,
		Rarity:        RarityUncommon,
		MinQuantity:   10,
		MaxQuantity:   50,
		RegenRate:     7.0, // Animal reproduction
		CooldownHours: 72,  // 3 days
		RequiredSkill: "hunting",
		MinSkillLevel: 20,
	},
	{
		Name:          "Rabbit Fur",
		Type:          ResourceAnimal,
		Rarity:        RarityCommon,
		MinQuantity:   15,
		MaxQuantity:   75,
		RegenRate:     10.0,
		CooldownHours: 48,
		RequiredSkill: "hunting",
		MinSkillLevel: 5,
	},
}

// Grassland biome resources (15-25 grains/km², 8-12 fibers/km², 5-10 grazing/km²)
var grasslandResources = []ResourceTemplate{
	{
		Name:          "Wild Grain",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   50,
		MaxQuantity:   200,
		RegenRate:     18.0,
		CooldownHours: 24,
		RequiredSkill: "gathering",
		MinSkillLevel: 0,
	},
	{
		Name:          "Cotton Fiber",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   40,
		MaxQuantity:   180,
		RegenRate:     12.0,
		CooldownHours: 48,
		RequiredSkill: "gathering",
		MinSkillLevel: 10,
	},
	{
		Name:          "Hemp Fiber",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   35,
		MaxQuantity:   160,
		RegenRate:     14.0,
		CooldownHours: 48,
		RequiredSkill: "gathering",
		MinSkillLevel: 8,
	},
	{
		Name:          "Bison Hide",
		Type:          ResourceAnimal,
		Rarity:        RarityCommon,
		MinQuantity:   20,
		MaxQuantity:   80,
		RegenRate:     6.0,
		CooldownHours: 72,
		RequiredSkill: "hunting",
		MinSkillLevel: 25,
	},
	{
		Name:          "Sheep Wool",
		Type:          ResourceAnimal,
		Rarity:        RarityCommon,
		MinQuantity:   25,
		MaxQuantity:   100,
		RegenRate:     8.0,
		CooldownHours: 48,
		RequiredSkill: "gathering",
		MinSkillLevel: 5,
	},
	{
		Name:          "Bird Feathers",
		Type:          ResourceAnimal,
		Rarity:        RarityUncommon,
		MinQuantity:   15,
		MaxQuantity:   60,
		RegenRate:     10.0,
		CooldownHours: 24,
		RequiredSkill: "gathering",
		MinSkillLevel: 12,
	},
}

// Desert biome resources (2-4 cacti/km², 0.3-0.8 crystals/km²)
var desertResources = []ResourceTemplate{
	{
		Name:          "Desert Cactus",
		Type:          ResourceVegetation,
		Rarity:        RarityUncommon,
		MinQuantity:   10,
		MaxQuantity:   50,
		RegenRate:     5.0,
		CooldownHours: 72,
		RequiredSkill: "gathering",
		MinSkillLevel: 15,
	},
	{
		Name:          "Aloe Vera",
		Type:          ResourceVegetation,
		Rarity:        RarityUncommon,
		MinQuantity:   8,
		MaxQuantity:   40,
		RegenRate:     6.0,
		CooldownHours: 48,
		RequiredSkill: "gathering",
		MinSkillLevel: 18,
	},
	{
		Name:          "Rare Crystal",
		Type:          ResourceSpecial,
		Rarity:        RarityRare,
		MinQuantity:   1,
		MaxQuantity:   10,
		RegenRate:     0.5, // Very slow
		CooldownHours: 168,
		RequiredSkill: "mining",
		MinSkillLevel: 40,
	},
	{
		Name:          "Scorpion Venom",
		Type:          ResourceAnimal,
		Rarity:        RarityRare,
		MinQuantity:   3,
		MaxQuantity:   15,
		RegenRate:     5.0,
		CooldownHours: 72,
		RequiredSkill: "hunting",
		MinSkillLevel: 35,
	},
}

// Ocean biome resources (20-30 fish/km², 10-15 kelp/km², 0.2-0.5 pearls/km²)
var oceanResources = []ResourceTemplate{
	{
		Name:          "Fish",
		Type:          ResourceAnimal,
		Rarity:        RarityCommon,
		MinQuantity:   50,
		MaxQuantity:   250,
		RegenRate:     10.0,
		CooldownHours: 24,
		RequiredSkill: "fishing",
		MinSkillLevel: 0,
	},
	{
		Name:          "Kelp",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   40,
		MaxQuantity:   200,
		RegenRate:     18.0,
		CooldownHours: 24,
		RequiredSkill: "gathering",
		MinSkillLevel: 5,
	},
	{
		Name:          "Seaweed",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   35,
		MaxQuantity:   180,
		RegenRate:     20.0,
		CooldownHours: 24,
		RequiredSkill: "gathering",
		MinSkillLevel: 0,
	},
	{
		Name:          "Pearl",
		Type:          ResourceSpecial,
		Rarity:        RarityRare,
		MinQuantity:   1,
		MaxQuantity:   5,
		RegenRate:     0.3,
		CooldownHours: 168,
		RequiredSkill: "gathering",
		MinSkillLevel: 50,
	},
	{
		Name:          "Coral",
		Type:          ResourceSpecial,
		Rarity:        RarityUncommon,
		MinQuantity:   5,
		MaxQuantity:   25,
		RegenRate:     1.0,
		CooldownHours: 168,
		RequiredSkill: "gathering",
		MinSkillLevel: 30,
	},
}

// Tundra biome resources (2-4 fur animals/km², 1-3 ice crystals/km²)
var tundraResources = []ResourceTemplate{
	{
		Name:          "Arctic Fox Fur",
		Type:          ResourceAnimal,
		Rarity:        RarityUncommon,
		MinQuantity:   10,
		MaxQuantity:   40,
		RegenRate:     6.0,
		CooldownHours: 72,
		RequiredSkill: "hunting",
		MinSkillLevel: 30,
	},
	{
		Name:          "Polar Bear Hide",
		Type:          ResourceAnimal,
		Rarity:        RarityRare,
		MinQuantity:   5,
		MaxQuantity:   20,
		RegenRate:     5.0,
		CooldownHours: 168,
		RequiredSkill: "hunting",
		MinSkillLevel: 60,
	},
	{
		Name:          "Ice Crystal",
		Type:          ResourceSpecial,
		Rarity:        RarityUncommon,
		MinQuantity:   3,
		MaxQuantity:   15,
		RegenRate:     2.0,
		CooldownHours: 72,
		RequiredSkill: "gathering",
		MinSkillLevel: 25,
	},
	{
		Name:          "Frozen Lichen",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   15,
		MaxQuantity:   60,
		RegenRate:     8.0,
		CooldownHours: 48,
		RequiredSkill: "gathering",
		MinSkillLevel: 10,
	},
}

// Rainforest biome resources (dense vegetation, exotic resources)
var rainforestResources = []ResourceTemplate{
	{
		Name:          "Mahogany Wood",
		Type:          ResourceVegetation,
		Rarity:        RarityUncommon,
		MinQuantity:   80,
		MaxQuantity:   350,
		RegenRate:     1.5,
		CooldownHours: 168,
		RequiredSkill: "gathering",
		MinSkillLevel: 20,
	},
	{
		Name:          "Rare Medicinal Plant",
		Type:          ResourceVegetation,
		Rarity:        RarityRare,
		MinQuantity:   10,
		MaxQuantity:   50,
		RegenRate:     10.0,
		CooldownHours: 48,
		RequiredSkill: "gathering",
		MinSkillLevel: 40,
	},
	{
		Name:          "Exotic Fruit",
		Type:          ResourceVegetation,
		Rarity:        RarityCommon,
		MinQuantity:   30,
		MaxQuantity:   120,
		RegenRate:     15.0,
		CooldownHours: 24,
		RequiredSkill: "gathering",
		MinSkillLevel: 5,
	},
	{
		Name:          "Jaguar Pelt",
		Type:          ResourceAnimal,
		Rarity:        RarityRare,
		MinQuantity:   5,
		MaxQuantity:   20,
		RegenRate:     6.0,
		CooldownHours: 168,
		RequiredSkill: "hunting",
		MinSkillLevel: 55,
	},
}

// Mountain biome resources (minerals from Phase 8.2b, minimal vegetation)
var mountainResources = []ResourceTemplate{
	{
		Name:          "Mountain Herbs",
		Type:          ResourceVegetation,
		Rarity:        RarityUncommon,
		MinQuantity:   8,
		MaxQuantity:   35,
		RegenRate:     5.0,
		CooldownHours: 72,
		RequiredSkill: "gathering",
		MinSkillLevel: 25,
	},
	{
		Name:          "Mountain Goat Hide",
		Type:          ResourceAnimal,
		Rarity:        RarityUncommon,
		MinQuantity:   10,
		MaxQuantity:   40,
		RegenRate:     6.0,
		CooldownHours: 72,
		RequiredSkill: "hunting",
		MinSkillLevel: 35,
	},
}
