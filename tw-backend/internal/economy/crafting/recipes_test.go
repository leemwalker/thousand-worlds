package crafting

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateTechTree(tree *TechTree) error                     { return nil }
func (m *MockRepository) GetTechTree(treeID uuid.UUID) (*TechTree, error)         { return nil, nil }
func (m *MockRepository) GetTechTreeByWorld(worldID uuid.UUID) (*TechTree, error) { return nil, nil }
func (m *MockRepository) CreateTechNode(node *TechNode) error                     { return nil }
func (m *MockRepository) GetTechNode(nodeID uuid.UUID) (*TechNode, error) {
	args := m.Called(nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TechNode), args.Error(1)
}
func (m *MockRepository) GetTechNodesByTree(treeID uuid.UUID) ([]*TechNode, error) { return nil, nil }
func (m *MockRepository) GetTechNodesByLevel(level TechLevel) ([]*TechNode, error) { return nil, nil }
func (m *MockRepository) CreateRecipe(recipe *Recipe) error                        { return nil }
func (m *MockRepository) GetRecipe(recipeID uuid.UUID) (*Recipe, error)            { return nil, nil }
func (m *MockRepository) GetRecipesByCategory(category RecipeCategory) ([]*Recipe, error) {
	return nil, nil
}
func (m *MockRepository) GetRecipesByTechLevel(techLevel TechLevel) ([]*Recipe, error) {
	return nil, nil
}
func (m *MockRepository) GetRecipesByTechNode(nodeID uuid.UUID) ([]*Recipe, error) { return nil, nil }
func (m *MockRepository) GetRecipesBySkill(skill string, maxLevel int) ([]*Recipe, error) {
	return nil, nil
}
func (m *MockRepository) UpdateRecipe(recipe *Recipe) error     { return nil }
func (m *MockRepository) DeleteRecipe(recipeID uuid.UUID) error { return nil }
func (m *MockRepository) UnlockTech(entityID uuid.UUID, nodeID uuid.UUID) error {
	args := m.Called(entityID, nodeID)
	return args.Error(0)
}
func (m *MockRepository) GetUnlockedTech(entityID uuid.UUID) ([]*UnlockedTech, error) {
	return nil, nil
}
func (m *MockRepository) IsTechUnlocked(entityID uuid.UUID, nodeID uuid.UUID) (bool, error) {
	args := m.Called(entityID, nodeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) DiscoverRecipe(knowledge *RecipeKnowledge) error {
	args := m.Called(knowledge)
	return args.Error(0)
}

func (m *MockRepository) GetKnownRecipes(entityID uuid.UUID) ([]*Recipe, error) {
	args := m.Called(entityID)
	return args.Get(0).([]*Recipe), args.Error(1)
}

func (m *MockRepository) GetRecipeKnowledge(entityID uuid.UUID, recipeID uuid.UUID) (*RecipeKnowledge, error) {
	args := m.Called(entityID, recipeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RecipeKnowledge), args.Error(1)
}

func (m *MockRepository) UpdateRecipeProficiency(entityID uuid.UUID, recipeID uuid.UUID, proficiency float64) error {
	return nil
}
func (m *MockRepository) SearchRecipes(query string, filters RecipeFilters) ([]*Recipe, error) {
	return nil, nil
}

func TestGetCraftableRecipes(t *testing.T) {
	// Setup
	repo := new(MockRepository)
	manager := NewRecipeManager(repo)

	entityID := uuid.New()
	woodID := uuid.New()
	stoneID := uuid.New()

	recipe := &Recipe{
		RecipeID: uuid.New(),
		Name:     "Stone Axe",
		Ingredients: []Ingredient{
			{ResourceID: woodID, Quantity: 1},
			{ResourceID: stoneID, Quantity: 1},
		},
	}

	repo.On("GetKnownRecipes", entityID).Return([]*Recipe{recipe}, nil)

	// Test 1: Insufficient resources
	resources := map[uuid.UUID]int{
		woodID: 1,
	}

	craftable, err := manager.GetCraftableRecipes(entityID, resources)
	assert.NoError(t, err)
	assert.Empty(t, craftable)

	// Test 2: Sufficient resources
	resources[stoneID] = 1

	craftable, err = manager.GetCraftableRecipes(entityID, resources)
	assert.NoError(t, err)
	assert.Len(t, craftable, 1)
	assert.Equal(t, "Stone Axe", craftable[0].Name)
}

func TestTeachRecipe(t *testing.T) {
	repo := new(MockRepository)
	manager := NewRecipeManager(repo)

	teacherID := uuid.New()
	studentID := uuid.New()
	recipeID := uuid.New()

	// Setup teacher knowledge
	knowledge := &RecipeKnowledge{
		EntityID:    teacherID,
		RecipeID:    recipeID,
		Proficiency: 100.0, // High proficiency
	}

	repo.On("GetRecipeKnowledge", teacherID, recipeID).Return(knowledge, nil)
	repo.On("DiscoverRecipe", mock.Anything).Return(nil)

	// Test success (high affection, high proficiency)
	err := manager.TeachRecipe(teacherID, studentID, recipeID, 80)
	assert.NoError(t, err)

	// Verify DiscoverRecipe was called
	repo.AssertCalled(t, "DiscoverRecipe", mock.MatchedBy(func(k *RecipeKnowledge) bool {
		return k.EntityID == studentID && k.RecipeID == recipeID && k.Source == "taught"
	}))
}

func TestTeachRecipeLowAffection(t *testing.T) {
	repo := new(MockRepository)
	manager := NewRecipeManager(repo)

	teacherID := uuid.New()
	studentID := uuid.New()
	recipeID := uuid.New()

	knowledge := &RecipeKnowledge{Proficiency: 100.0}
	repo.On("GetRecipeKnowledge", teacherID, recipeID).Return(knowledge, nil)

	// Test failure (low affection)
	err := manager.TeachRecipe(teacherID, studentID, recipeID, 20)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "relationship too low")
}

func TestExperimentWithIngredients(t *testing.T) {
	repo := new(MockRepository)
	manager := NewRecipeManager(repo)

	entityID := uuid.New()
	station := &CraftingStation{}

	// Test failure (random chance logic, but we can test inputs)
	// Since random is used without seed control in the source, we can't deterministically test success/fail
	// unless we mock rand or retry many times.
	// However, the function returns nil, nil for now as per source code.

	ingredients := []Ingredient{
		{ResourceID: uuid.New(), Quantity: 1},
	}

	recipe, err := manager.ExperimentWithIngredients(entityID, ingredients, station, 50)
	assert.NoError(t, err)
	assert.Nil(t, recipe) // Implementation currently returns nil, nil
}

func TestLoadRecipesFromFile(t *testing.T) {
	// Create temp file
	file, err := os.CreateTemp("", "recipes-*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	content := `[
		{"name": "Recipe 1", "ingredients": [{"resource_id": "00000000-0000-0000-0000-000000000001", "quantity": 1}]}
	]`
	file.WriteString(content)
	file.Close()

	recipes, err := LoadRecipesFromFile(file.Name())
	assert.NoError(t, err)
	assert.Len(t, recipes, 1)
	assert.Equal(t, "Recipe 1", recipes[0].Name)
	assert.NotEqual(t, uuid.Nil, recipes[0].RecipeID)
}

func TestGetCraftableRecipes_Substitutes(t *testing.T) {
	repo := new(MockRepository)
	manager := NewRecipeManager(repo)

	entityID := uuid.New()
	mainID := uuid.New()
	subID := uuid.New()

	recipe := &Recipe{
		RecipeID: uuid.New(),
		Name:     "Sub Test",
		Ingredients: []Ingredient{
			{
				ResourceID: mainID,
				Quantity:   10,
				Substitute: []uuid.UUID{subID},
			},
		},
	}

	repo.On("GetKnownRecipes", entityID).Return([]*Recipe{recipe}, nil)

	// Case 1: Enough main resource
	resources1 := map[uuid.UUID]int{mainID: 10}
	craftable1, err := manager.GetCraftableRecipes(entityID, resources1)
	assert.NoError(t, err)
	assert.Len(t, craftable1, 1)

	// Case 2: Mix of main and sub
	resources2 := map[uuid.UUID]int{mainID: 5, subID: 5}
	craftable2, err := manager.GetCraftableRecipes(entityID, resources2)
	assert.NoError(t, err)
	assert.Len(t, craftable2, 1)

	// Case 3: Only sub
	resources3 := map[uuid.UUID]int{subID: 10}
	craftable3, err := manager.GetCraftableRecipes(entityID, resources3)
	assert.NoError(t, err)
	assert.Len(t, craftable3, 1)

	// Case 4: Insufficient total
	resources4 := map[uuid.UUID]int{mainID: 4, subID: 5}
	craftable4, err := manager.GetCraftableRecipes(entityID, resources4)
	assert.NoError(t, err)
	assert.Empty(t, craftable4)
}
