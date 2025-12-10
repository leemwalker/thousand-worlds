package crafting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// LoadRecipesFromFile loads recipe definitions from a JSON file
func LoadRecipesFromFile(path string) ([]*Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var recipes []*Recipe
	if err := json.Unmarshal(data, &recipes); err != nil {
		return nil, err
	}

	for _, recipe := range recipes {
		if recipe.RecipeID == uuid.Nil {
			recipe.RecipeID = uuid.New()
		}
	}

	return recipes, nil
}

// RecipeManager handles recipe operations
type RecipeManager struct {
	repo Repository
}

func NewRecipeManager(repo Repository) *RecipeManager {
	return &RecipeManager{repo: repo}
}

// GetCraftableRecipes returns recipes that can be crafted with available resources
func (m *RecipeManager) GetCraftableRecipes(entityID uuid.UUID, resources map[uuid.UUID]int) ([]*Recipe, error) {
	// 1. Get known recipes
	known, err := m.repo.GetKnownRecipes(entityID)
	if err != nil {
		return nil, err
	}

	var craftable []*Recipe

	// 2. Check ingredients for each recipe
	for _, recipe := range known {
		if canCraft(recipe, resources) {
			craftable = append(craftable, recipe)
		}
	}

	return craftable, nil
}

func canCraft(recipe *Recipe, resources map[uuid.UUID]int) bool {
	for _, ingredient := range recipe.Ingredients {
		// Check primary resource
		available := resources[ingredient.ResourceID]

		// Check substitutes if primary not enough
		if available < ingredient.Quantity {
			substituteAvailable := 0
			for _, subID := range ingredient.Substitute {
				substituteAvailable += resources[subID]
			}

			if available+substituteAvailable < ingredient.Quantity {
				return false
			}
		}
	}
	return true
}

// LoadAllRecipes loads all JSON files from the data directory
func LoadAllRecipes(dataDir string) ([]*Recipe, error) {
	var allRecipes []*Recipe

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			path := filepath.Join(dataDir, file.Name())
			recipes, err := LoadRecipesFromFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to load %s: %w", file.Name(), err)
			}
			allRecipes = append(allRecipes, recipes...)
		}
	}

	return allRecipes, nil
}
