package crafting

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateTechTree(tree *TechTree) error                      { return nil }
func (m *MockRepository) GetTechTree(treeID uuid.UUID) (*TechTree, error)          { return nil, nil }
func (m *MockRepository) GetTechTreeByWorld(worldID uuid.UUID) (*TechTree, error)  { return nil, nil }
func (m *MockRepository) CreateTechNode(node *TechNode) error                      { return nil }
func (m *MockRepository) GetTechNode(nodeID uuid.UUID) (*TechNode, error)          { return nil, nil }
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
func (m *MockRepository) UpdateRecipe(recipe *Recipe) error                     { return nil }
func (m *MockRepository) DeleteRecipe(recipeID uuid.UUID) error                 { return nil }
func (m *MockRepository) UnlockTech(entityID uuid.UUID, nodeID uuid.UUID) error { return nil }
func (m *MockRepository) GetUnlockedTech(entityID uuid.UUID) ([]*UnlockedTech, error) {
	return nil, nil
}
func (m *MockRepository) IsTechUnlocked(entityID uuid.UUID, nodeID uuid.UUID) (bool, error) {
	return false, nil
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
