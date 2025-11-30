package crafting

import (
	"time"

	"github.com/google/uuid"
)

// TechLevel represents the technological era
type TechLevel string

const (
	TechPrimitive  TechLevel = "primitive"
	TechMedieval   TechLevel = "medieval"
	TechIndustrial TechLevel = "industrial"
	TechModern     TechLevel = "modern"
	TechFuturistic TechLevel = "futuristic"
)

// TechTree represents a complete technology tree for a world
type TechTree struct {
	TreeID    uuid.UUID
	WorldID   uuid.UUID
	Name      string
	TechLevel TechLevel
	Nodes     []*TechNode
	CreatedAt time.Time
}

// TechNode represents a single technology that can be researched
type TechNode struct {
	NodeID         uuid.UUID
	TreeID         uuid.UUID
	Name           string
	Description    string
	TechLevel      TechLevel
	Tier           int            // 1-4 within each tech level
	Prerequisites  []uuid.UUID    // IDs of required parent nodes
	UnlocksRecipes []uuid.UUID    // IDs of recipes unlocked by this node
	ResearchCost   map[string]int // Resource costs to unlock (Resource Name -> Quantity)
	ResearchTime   time.Duration
	IconPath       string
	CreatedAt      time.Time
}

// RecipeCategory defines the type of item produced
type RecipeCategory string

const (
	CategoryWeapon     RecipeCategory = "weapon"
	CategoryArmor      RecipeCategory = "armor"
	CategoryTool       RecipeCategory = "tool"
	CategoryConsumable RecipeCategory = "consumable"
	CategoryBuilding   RecipeCategory = "building"
	CategoryComponent  RecipeCategory = "component"
)

// Difficulty represents how hard a recipe is to execute
type Difficulty string

const (
	DifficultyTrivial    Difficulty = "trivial"
	DifficultyEasy       Difficulty = "easy"
	DifficultyMedium     Difficulty = "medium"
	DifficultyHard       Difficulty = "hard"
	DifficultyVeryHard   Difficulty = "very_hard"
	DifficultyMasterwork Difficulty = "masterwork"
)

// Recipe represents a formula for creating an item
type Recipe struct {
	RecipeID    uuid.UUID
	Name        string
	Description string
	Category    RecipeCategory
	TechNodeID  *uuid.UUID // nil if no tech requirement (basic recipes)

	// Inputs
	Ingredients     []Ingredient
	RequiredTool    *ToolRequirement
	RequiredStation *CraftingStation // forge, anvil, alchemy_table, workbench

	// Outputs
	Output     ItemOutput
	ByProducts []ItemOutput // Secondary products (e.g., leather scraps, slag)

	// Requirements
	RequiredSkill string // "smithing", "alchemy", "carpentry", etc.
	MinSkillLevel int
	CraftingTime  time.Duration
	SuccessRate   SuccessRateFormula

	// Quality
	QualityTiers []QualityTier

	// Economy
	BaseValue  int // Base economic value
	Difficulty Difficulty

	CreatedAt time.Time
}

// ItemQuality represents the quality level of an item
type ItemQuality int

const (
	QualityPoor       ItemQuality = 0
	QualityCommon     ItemQuality = 1
	QualityGood       ItemQuality = 2
	QualityExcellent  ItemQuality = 3
	QualityMasterwork ItemQuality = 4
)

// Ingredient represents a required resource for a recipe
type Ingredient struct {
	ResourceID uuid.UUID
	Quantity   int
	Quality    *ItemQuality // nil if any quality accepted
	Substitute []uuid.UUID  // Alternative resources (oak OR birch wood)
}

// ToolRequirement specifies a tool needed for crafting
type ToolRequirement struct {
	ToolType       string // "hammer", "saw", "chisel", "tongs"
	MinToolQuality ItemQuality
}

// CraftingStation specifies a facility needed for crafting
type CraftingStation struct {
	StationType    string // "forge", "anvil", "alchemy_table", "loom"
	MinStationTier int    // 1-5 (basic to legendary station quality)
}

// ItemOutput defines the result of a recipe
type ItemOutput struct {
	ItemID   uuid.UUID
	Quantity int
	Quality  ItemQuality // Determined by crafter skill
}

// SuccessRateFormula calculates the chance of successful crafting
type SuccessRateFormula struct {
	BaseRate        float64 // 0.0 to 1.0
	SkillModifier   float64 // Per skill point above minimum
	ToolModifier    float64 // Per tool quality level
	StationModifier float64 // Per station tier
}

// QualityTier defines probability of achieving a specific quality
type QualityTier struct {
	Name           string // "poor", "common", "good", "excellent", "masterwork"
	MinSkillLevel  int
	StatModifier   float64 // Multiplier for item stats
	Probability    float64 // Base chance at this tier
	SkillInfluence float64 // How much skill increases this tier's probability
}

// RecipeKnowledge tracks what an entity knows
type RecipeKnowledge struct {
	EntityID     uuid.UUID
	RecipeID     uuid.UUID
	Proficiency  float64 // 0-100, improves with use
	TimesUsed    int
	DiscoveredAt time.Time
	Source       string     // "research", "experiment", "taught", "found"
	TeacherID    *uuid.UUID // If taught, who taught it
}

// UnlockedTech tracks what tech nodes an entity has researched
type UnlockedTech struct {
	EntityID   uuid.UUID
	NodeID     uuid.UUID
	UnlockedAt time.Time
}

// CraftResult represents the outcome of a crafting attempt
type CraftResult struct {
	Success       bool
	Item          ItemOutput
	ByProducts    []ItemOutput
	FailureReason string
}
