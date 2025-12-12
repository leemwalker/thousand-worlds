package resources

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ResourceType categorizes resources by origin
type ResourceType string

const (
	ResourceMineral    ResourceType = "mineral"
	ResourceVegetation ResourceType = "vegetation"
	ResourceAnimal     ResourceType = "animal"
	ResourceSpecial    ResourceType = "special"
)

// Rarity defines how common a resource is
type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityUncommon  Rarity = "uncommon"
	RarityRare      Rarity = "rare"
	RarityVeryRare  Rarity = "very_rare"
	RarityLegendary Rarity = "legendary"
)

// ResourceCategory defines specific subtypes of resources for gathering priorities
type ResourceCategory string

const (
	CategoryGrain     ResourceCategory = "grain"
	CategoryVegetable ResourceCategory = "vegetable"
	CategoryFruit     ResourceCategory = "fruit"
	CategoryFiber     ResourceCategory = "fiber"
	CategoryHerb      ResourceCategory = "herb"
	CategoryFlower    ResourceCategory = "flower"
	CategoryMushroom  ResourceCategory = "mushroom"
	CategoryBerry     ResourceCategory = "berry"
	CategoryWood      ResourceCategory = "wood"
	CategoryResin     ResourceCategory = "resin"
	CategoryOre       ResourceCategory = "ore"
	CategoryCoal      ResourceCategory = "coal"
	CategoryGem       ResourceCategory = "gem"
	CategoryStone     ResourceCategory = "stone"
	CategoryMeat      ResourceCategory = "meat"
	CategoryHide      ResourceCategory = "hide"
	CategoryBone      ResourceCategory = "bone"
	CategoryFeather   ResourceCategory = "feather"
	CategoryUnknown   ResourceCategory = "unknown"
)

// ResourceNode represents a harvestable resource in the world
type ResourceNode struct {
	NodeID        uuid.UUID
	Name          string
	Type          ResourceType
	Rarity        Rarity
	LocationX     float64
	LocationY     float64
	LocationZ     float64
	Quantity      int
	MaxQuantity   int
	RegenRate     float64       // Units per day
	RegenCooldown time.Duration // Time before regen starts after harvest
	LastHarvested *time.Time
	BiomeAffinity []string // Biomes where this resource appears
	RequiredSkill string
	MinSkillLevel int

	// Mineral-specific (only populated for ResourceMineral type)
	MineralDepositID *uuid.UUID // Links to Phase 8.2b MineralDeposit
	Depth            float64    // Mining depth required (meters)

	// Animal-specific (only populated for ResourceAnimal type)
	SpeciesID *uuid.UUID // Links to Phase 8.4 Species

	CreatedAt time.Time
}

// GetCategory derives the specific category from the resource name and type
func (n *ResourceNode) GetCategory() ResourceCategory {
	// Simple keyword matching for now
	// In a real system, this might be a DB field or a lookup map
	name := strings.Title(n.Name) // Normalize case for checks, or just check both cases.
	// Better:
	name = n.Name

	switch n.Type {
	case ResourceMineral:
		if contains(name, "Ore") || contains(name, "Iron") || contains(name, "Copper") || contains(name, "Gold") {
			return CategoryOre
		}
		if contains(name, "Coal") {
			return CategoryCoal
		}
		if contains(name, "Gem") || contains(name, "Ruby") || contains(name, "Sapphire") || contains(name, "Crystal") {
			return CategoryGem
		}
		return CategoryStone

	case ResourceVegetation:
		if contains(name, "Wood") || contains(name, "Log") || contains(name, "Tree") {
			return CategoryWood
		}
		if contains(name, "Grain") || contains(name, "Wheat") || contains(name, "Barley") {
			return CategoryGrain
		}
		if contains(name, "Fiber") || contains(name, "Cotton") || contains(name, "Hemp") {
			return CategoryFiber
		}
		if contains(name, "Herb") || contains(name, "Root") || contains(name, "Leaf") {
			return CategoryHerb
		}
		if contains(name, "Berry") || contains(name, "Berries") {
			return CategoryBerry
		}
		if contains(name, "Mushroom") || contains(name, "Fungus") {
			return CategoryMushroom
		}
		if contains(name, "Flower") {
			return CategoryFlower
		}
		if contains(name, "Fruit") || contains(name, "Apple") {
			return CategoryFruit
		}
		if contains(name, "Resin") || contains(name, "Sap") {
			return CategoryResin
		}
		// Default vegetation
		return CategoryVegetable

	case ResourceAnimal:
		if contains(name, "Meat") {
			return CategoryMeat
		}
		if contains(name, "Hide") || contains(name, "Leather") || contains(name, "Pelt") || contains(name, "Fur") || contains(name, "Wool") {
			return CategoryHide
		}
		if contains(name, "Bone") || contains(name, "Horn") {
			return CategoryBone
		}
		if contains(name, "Feather") {
			return CategoryFeather
		}
		return CategoryMeat // Default animal product
	}

	return CategoryUnknown
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// MineralDeposit represents the Phase 8.2b mineral deposit structure
// This is for reference when querying existing deposits
type MineralDeposit struct {
	DepositID        uuid.UUID
	MineralType      string
	FormationType    string
	LocationX        float64
	LocationY        float64
	Depth            float64
	Quantity         int
	Concentration    float64
	VeinSize         string
	GeologicalAge    float64
	VeinShape        string
	VeinOrientationX float64
	VeinOrientationY float64
	VeinLength       float64
	VeinWidth        float64
	SurfaceVisible   bool
	RequiredDepth    float64
	CreatedAt        time.Time
}

// ResourceTemplate defines the blueprint for generating resource nodes
type ResourceTemplate struct {
	Name          string
	Type          ResourceType
	Rarity        Rarity
	MinQuantity   int
	MaxQuantity   int
	RegenRate     float64 // Units per day
	CooldownHours int
	RequiredSkill string
	MinSkillLevel int
	BiomeTypes    []string // Biomes where this resource can spawn
}

// HarvestRequest contains parameters for harvesting a resource
type HarvestRequest struct {
	NodeID        uuid.UUID
	GathererSkill int // 0-100
	ToolQuality   int // 0-130 (0=no tool, 50=basic, 70=basic, 100=good, 130=excellent)
}

// HarvestResult contains the outcome of a harvest attempt
type HarvestResult struct {
	Success       bool
	YieldAmount   int
	FailureReason string
}
