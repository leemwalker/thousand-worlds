package crafting

import (
	"context"
	"testing"

	"tw-backend/internal/errors"
	"tw-backend/internal/worldentity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepo
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateTechTree(tree *TechTree) error { return m.Called(tree).Error(0) }
func (m *MockRepo) GetTechTree(treeID uuid.UUID) (*TechTree, error) {
	args := m.Called(treeID)
	return args.Get(0).(*TechTree), args.Error(1)
}
func (m *MockRepo) GetTechTreeByWorld(worldID uuid.UUID) (*TechTree, error) {
	args := m.Called(worldID)
	return args.Get(0).(*TechTree), args.Error(1)
}
func (m *MockRepo) CreateTechNode(node *TechNode) error { return m.Called(node).Error(0) }
func (m *MockRepo) GetTechNode(nodeID uuid.UUID) (*TechNode, error) {
	args := m.Called(nodeID)
	return args.Get(0).(*TechNode), args.Error(1)
}
func (m *MockRepo) GetTechNodesByTree(treeID uuid.UUID) ([]*TechNode, error) {
	args := m.Called(treeID)
	return args.Get(0).([]*TechNode), args.Error(1)
}
func (m *MockRepo) GetTechNodesByLevel(level TechLevel) ([]*TechNode, error) {
	args := m.Called(level)
	return args.Get(0).([]*TechNode), args.Error(1)
}
func (m *MockRepo) CreateRecipe(recipe *Recipe) error { return m.Called(recipe).Error(0) }
func (m *MockRepo) GetRecipe(recipeID uuid.UUID) (*Recipe, error) {
	args := m.Called(recipeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Recipe), args.Error(1)
}
func (m *MockRepo) GetRecipesByCategory(category RecipeCategory) ([]*Recipe, error) {
	args := m.Called(category)
	return args.Get(0).([]*Recipe), args.Error(1)
}
func (m *MockRepo) GetRecipesByTechLevel(techLevel TechLevel) ([]*Recipe, error) {
	args := m.Called(techLevel)
	return args.Get(0).([]*Recipe), args.Error(1)
}
func (m *MockRepo) GetRecipesByTechNode(nodeID uuid.UUID) ([]*Recipe, error) {
	args := m.Called(nodeID)
	return args.Get(0).([]*Recipe), args.Error(1)
}
func (m *MockRepo) GetRecipesBySkill(skill string, maxLevel int) ([]*Recipe, error) {
	args := m.Called(skill, maxLevel)
	return args.Get(0).([]*Recipe), args.Error(1)
}
func (m *MockRepo) UpdateRecipe(recipe *Recipe) error     { return m.Called(recipe).Error(0) }
func (m *MockRepo) DeleteRecipe(recipeID uuid.UUID) error { return m.Called(recipeID).Error(0) }
func (m *MockRepo) UnlockTech(entityID uuid.UUID, nodeID uuid.UUID) error {
	return m.Called(entityID, nodeID).Error(0)
}
func (m *MockRepo) GetUnlockedTech(entityID uuid.UUID) ([]*UnlockedTech, error) {
	args := m.Called(entityID)
	return args.Get(0).([]*UnlockedTech), args.Error(1)
}
func (m *MockRepo) IsTechUnlocked(entityID uuid.UUID, nodeID uuid.UUID) (bool, error) {
	args := m.Called(entityID, nodeID)
	return args.Bool(0), args.Error(1)
}
func (m *MockRepo) DiscoverRecipe(knowledge *RecipeKnowledge) error {
	return m.Called(knowledge).Error(0)
}
func (m *MockRepo) GetKnownRecipes(entityID uuid.UUID) ([]*Recipe, error) {
	args := m.Called(entityID)
	return args.Get(0).([]*Recipe), args.Error(1)
}
func (m *MockRepo) GetRecipeKnowledge(entityID uuid.UUID, recipeID uuid.UUID) (*RecipeKnowledge, error) {
	args := m.Called(entityID, recipeID)
	return args.Get(0).(*RecipeKnowledge), args.Error(1)
}
func (m *MockRepo) UpdateRecipeProficiency(entityID uuid.UUID, recipeID uuid.UUID, proficiency float64) error {
	return m.Called(entityID, recipeID, proficiency).Error(0)
}
func (m *MockRepo) SearchRecipes(query string, filters RecipeFilters) ([]*Recipe, error) {
	args := m.Called(query, filters)
	return args.Get(0).([]*Recipe), args.Error(1)
}

// MockInventory
type MockInventory struct {
	mock.Mock
}

func (m *MockInventory) RemoveItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int) error {
	return m.Called(ctx, charID, itemID, quantity).Error(0)
}
func (m *MockInventory) AddItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int, metadata map[string]interface{}) error {
	return m.Called(ctx, charID, itemID, quantity, metadata).Error(0)
}

func TestCraft_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockInv := new(MockInventory)
	svc := NewService(mockRepo, mockInv, nil)

	charID := uuid.New()
	recipeID := uuid.New()
	ingID := uuid.New()
	outID := uuid.New()

	recipe := &Recipe{
		RecipeID: recipeID,
		Ingredients: []Ingredient{
			{ResourceID: ingID, Quantity: 2},
		},
		Output: ItemOutput{ItemID: outID, Quantity: 1},
	}

	mockRepo.On("GetRecipe", recipeID).Return(recipe, nil)
	mockInv.On("RemoveItem", mock.Anything, charID, ingID, 2).Return(nil)
	mockInv.On("AddItem", mock.Anything, charID, outID, 1, mock.Anything).Return(nil)

	res, err := svc.Craft(context.Background(), charID, recipeID, nil)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, outID, res.Item.ItemID)

	mockRepo.AssertExpectations(t)
	mockInv.AssertExpectations(t)
}

func TestCraft_RecipeNotFound(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := NewService(mockRepo, nil, nil)

	id := uuid.New()
	mockRepo.On("GetRecipe", id).Return(nil, errors.NewNotFound("not found"))

	_, err := svc.Craft(context.Background(), uuid.New(), id, nil)
	require.Error(t, err)
}

func TestCraft_MissingIngredients(t *testing.T) {
	mockRepo := new(MockRepo)
	mockInv := new(MockInventory)
	svc := NewService(mockRepo, mockInv, nil)

	charID := uuid.New()
	recipeID := uuid.New()
	ingID := uuid.New()

	recipe := &Recipe{
		RecipeID: recipeID,
		Ingredients: []Ingredient{
			{ResourceID: ingID, Quantity: 1},
		},
	}

	mockRepo.On("GetRecipe", recipeID).Return(recipe, nil)
	mockInv.On("RemoveItem", mock.Anything, charID, ingID, 1).Return(errors.NewInvalidInput("missing"))

	_, err := svc.Craft(context.Background(), charID, recipeID, nil)
	require.Error(t, err)
}

func TestGetAvailableRecipes(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := NewService(mockRepo, nil, nil)
	charID := uuid.New()

	recipes := []*Recipe{{RecipeID: uuid.New()}}
	mockRepo.On("GetKnownRecipes", charID).Return(recipes, nil)

	res, err := svc.GetAvailableRecipes(context.Background(), charID)
	require.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestFindRecipeByName(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := NewService(mockRepo, nil, nil)

	recipes := []*Recipe{{Name: "Sword"}}
	mockRepo.On("SearchRecipes", "Sword", mock.Anything).Return(recipes, nil)

	res, err := svc.FindRecipeByName(context.Background(), "Sword")
	require.NoError(t, err)
	assert.Equal(t, "Sword", res.Name)
}

func TestFindRecipeByName_NotFound(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := NewService(mockRepo, nil, nil)

	mockRepo.On("SearchRecipes", "Excalibur", mock.Anything).Return([]*Recipe{}, nil)

	_, err := svc.FindRecipeByName(context.Background(), "Excalibur")
	require.Error(t, err)
}

func TestCraft_RequiredStation(t *testing.T) {
	mockRepo := new(MockRepo)
	mockInv := new(MockInventory)
	mockWorldEnt := new(MockWorldEntityService)
	svc := NewService(mockRepo, mockInv, mockWorldEnt)

	charID := uuid.New()
	recipeID := uuid.New()
	stationID := uuid.New()

	recipe := &Recipe{
		RecipeID:        recipeID,
		RequiredStation: &CraftingStation{StationType: "Forge", MinStationTier: 1},
		Ingredients:     []Ingredient{},
		Output:          ItemOutput{ItemID: uuid.New(), Quantity: 1},
	}

	// Case 1: Station missing
	mockRepo.On("GetRecipe", recipeID).Return(recipe, nil)
	_, err := svc.Craft(context.Background(), charID, recipeID, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "station required")

	// Case 2: Station valid
	stationEnt := &worldentity.WorldEntity{ID: stationID, Name: "My Forge"}
	mockRepo.On("GetRecipe", recipeID).Return(recipe, nil)
	mockWorldEnt.On("GetByID", mock.Anything, stationID).Return(stationEnt, nil)

	// Expect inventory calls
	mockInv.On("AddItem", mock.Anything, charID, recipe.Output.ItemID, recipe.Output.Quantity, mock.Anything).Return(nil)

	res, err := svc.Craft(context.Background(), charID, recipeID, &stationID)
	require.NoError(t, err)
	assert.True(t, res.Success)
}

func TestCraft_Byproducts(t *testing.T) {
	mockRepo := new(MockRepo)
	mockInv := new(MockInventory)
	svc := NewService(mockRepo, mockInv, nil)

	charID := uuid.New()
	recipeID := uuid.New()

	recipe := &Recipe{
		RecipeID:    recipeID,
		Ingredients: []Ingredient{},
		Output:      ItemOutput{ItemID: uuid.New(), Quantity: 1},
		ByProducts:  []ItemOutput{{ItemID: uuid.New(), Quantity: 2}},
	}

	mockRepo.On("GetRecipe", recipeID).Return(recipe, nil)
	mockInv.On("AddItem", mock.Anything, charID, recipe.Output.ItemID, recipe.Output.Quantity, mock.Anything).Return(nil)
	// Expect byproduct AddItem
	mockInv.On("AddItem", mock.Anything, charID, recipe.ByProducts[0].ItemID, recipe.ByProducts[0].Quantity, mock.Anything).Return(nil)

	res, err := svc.Craft(context.Background(), charID, recipeID, nil)
	require.NoError(t, err)
	assert.Len(t, res.ByProducts, 1)

	mockInv.AssertExpectations(t)
}

// MockWorldEntityService
type MockWorldEntityService struct{ mock.Mock }

func (m *MockWorldEntityService) GetByID(ctx context.Context, id uuid.UUID) (*worldentity.WorldEntity, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*worldentity.WorldEntity), args.Error(1)
}
