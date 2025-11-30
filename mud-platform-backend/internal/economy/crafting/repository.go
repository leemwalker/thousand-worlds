package crafting

import (
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for persisting and retrieving crafting data
type Repository interface {
	// Tech Tree Operations
	CreateTechTree(tree *TechTree) error
	GetTechTree(treeID uuid.UUID) (*TechTree, error)
	GetTechTreeByWorld(worldID uuid.UUID) (*TechTree, error)

	// Tech Node Operations
	CreateTechNode(node *TechNode) error
	GetTechNode(nodeID uuid.UUID) (*TechNode, error)
	GetTechNodesByTree(treeID uuid.UUID) ([]*TechNode, error)
	GetTechNodesByLevel(level TechLevel) ([]*TechNode, error)

	// Recipe Operations
	CreateRecipe(recipe *Recipe) error
	GetRecipe(recipeID uuid.UUID) (*Recipe, error)
	GetRecipesByCategory(category RecipeCategory) ([]*Recipe, error)
	GetRecipesByTechLevel(techLevel TechLevel) ([]*Recipe, error)
	GetRecipesByTechNode(nodeID uuid.UUID) ([]*Recipe, error)
	GetRecipesBySkill(skill string, maxLevel int) ([]*Recipe, error)
	UpdateRecipe(recipe *Recipe) error
	DeleteRecipe(recipeID uuid.UUID) error

	// Knowledge Operations
	UnlockTech(entityID uuid.UUID, nodeID uuid.UUID) error
	GetUnlockedTech(entityID uuid.UUID) ([]*UnlockedTech, error)
	IsTechUnlocked(entityID uuid.UUID, nodeID uuid.UUID) (bool, error)

	DiscoverRecipe(knowledge *RecipeKnowledge) error
	GetKnownRecipes(entityID uuid.UUID) ([]*Recipe, error)
	GetRecipeKnowledge(entityID uuid.UUID, recipeID uuid.UUID) (*RecipeKnowledge, error)
	UpdateRecipeProficiency(entityID uuid.UUID, recipeID uuid.UUID, proficiency float64) error

	// Search
	SearchRecipes(query string, filters RecipeFilters) ([]*Recipe, error)
}

// RecipeFilters defines criteria for searching recipes
type RecipeFilters struct {
	Category        *RecipeCategory
	TechLevel       *TechLevel
	MaxSkillLevel   *int
	MaxCraftTime    *time.Duration
	Difficulty      *Difficulty
	RequiredStation *string
}
