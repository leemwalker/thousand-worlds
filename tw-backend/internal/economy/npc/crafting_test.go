package npc

import (
	"context"
	"testing"

	"mud-platform-backend/internal/economy/crafting"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockRecipeManager struct {
	mock.Mock
}

func (m *MockRecipeManager) GetKnownRecipes(entityID uuid.UUID) ([]*crafting.Recipe, error) {
	args := m.Called(entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*crafting.Recipe), args.Error(1)
}

func (m *MockRecipeManager) GetCraftableRecipes(entityID uuid.UUID, resources map[uuid.UUID]int) ([]*crafting.Recipe, error) {
	args := m.Called(entityID, resources)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*crafting.Recipe), args.Error(1)
}

func (m *MockRecipeManager) ImproveProficiency(entityID, recipeID uuid.UUID, amount float64) error {
	args := m.Called(entityID, recipeID, amount)
	return args.Error(0)
}

type MockMarketDataProvider struct {
	mock.Mock
}

func (m *MockMarketDataProvider) GetAveragePrice(locationID, itemID uuid.UUID) float64 {
	args := m.Called(locationID, itemID)
	return args.Get(0).(float64)
}

func (m *MockMarketDataProvider) GetLocalDemand(locationID, itemID uuid.UUID) int {
	args := m.Called(locationID, itemID)
	return args.Int(0)
}

type MockStationFinder struct {
	mock.Mock
}

func (m *MockStationFinder) FindNearestStation(location uuid.UUID, stationType string, minTier int) (*crafting.CraftingStation, error) {
	args := m.Called(location, stationType, minTier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*crafting.CraftingStation), args.Error(1)
}

type MockCrafter struct {
	mock.Mock
}

func (m *MockCrafter) Craft(recipeID uuid.UUID, skillLevel int, toolQuality float64, stationTier int) (*crafting.CraftResult, error) {
	args := m.Called(recipeID, skillLevel, toolQuality, stationTier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*crafting.CraftResult), args.Error(1)
}

// MockInventoryManager extension for GetContents
func (m *MockInventoryManager) GetContents() map[uuid.UUID]int {
	args := m.Called()
	return args.Get(0).(map[uuid.UUID]int)
}

func TestCraftGoods(t *testing.T) {
	// Setup
	recipeMgr := new(MockRecipeManager)
	market := new(MockMarketDataProvider)
	stationFinder := new(MockStationFinder)
	crafter := new(MockCrafter)
	inventory := new(MockInventoryManager)

	npcID := uuid.New()
	locationID := uuid.New()

	npc := &NPC{
		ID:        npcID,
		Location:  locationID,
		Skills:    map[string]int{"smithing": 50},
		Inventory: inventory,
		Desires: &Desires{
			TaskCompletion: 80.0, // High need
		},
		Equipment: &Equipment{
			ToolQuality: 1.0,
		},
	}

	// Mock data
	recipeID1 := uuid.New()
	outputItemID1 := uuid.New()

	recipe1 := &crafting.Recipe{
		RecipeID:      recipeID1,
		Name:          "Iron Sword",
		RequiredSkill: "smithing",
		Output: crafting.ItemOutput{
			ItemID:   outputItemID1,
			Quantity: 1,
		},
		RequiredStation: &crafting.CraftingStation{
			StationType:    "anvil",
			MinStationTier: 1,
		},
		Difficulty: crafting.DifficultyMedium,
	}

	recipeID2 := uuid.New()
	outputItemID2 := uuid.New()

	recipe2 := &crafting.Recipe{
		RecipeID:      recipeID2,
		Name:          "Wooden Shield", // Worse option
		RequiredSkill: "smithing",
		Output: crafting.ItemOutput{
			ItemID:   outputItemID2,
			Quantity: 1,
		},
		RequiredStation: &crafting.CraftingStation{
			StationType:    "workbench",
			MinStationTier: 1,
		},
		Difficulty: crafting.DifficultyEasy,
	}

	// Expectation: Get contents
	inventory.On("GetContents").Return(map[uuid.UUID]int{})

	// Expectation: Get craftable recipes
	recipeMgr.On("GetCraftableRecipes", npcID, mock.Anything).Return([]*crafting.Recipe{recipe1, recipe2}, nil)

	// Expectation: Market data for prioritization
	// Recipe 1: 100 * 10 = 1000 score
	market.On("GetAveragePrice", locationID, outputItemID1).Return(100.0)
	market.On("GetLocalDemand", locationID, outputItemID1).Return(10)

	// Recipe 2: 50 * 5 = 250 score
	market.On("GetAveragePrice", locationID, outputItemID2).Return(50.0)
	market.On("GetLocalDemand", locationID, outputItemID2).Return(5)

	// Expectation: Find station (for the winner, recipe1)
	stationFinder.On("FindNearestStation", locationID, "anvil", 1).Return(&crafting.CraftingStation{}, nil)

	// Expectation: Craft (recipe1)
	result := &crafting.CraftResult{
		Success: true,
		Item: crafting.ItemOutput{
			ItemID:   outputItemID1,
			Quantity: 1,
		},
	}
	crafter.On("Craft", recipeID1, 50, 1.0, 1).Return(result, nil)

	// Expectation: Add to inventory
	inventory.On("Add", outputItemID1, 1).Return(nil)

	// Expectation: Improve proficiency
	recipeMgr.On("ImproveProficiency", npcID, recipeID1, 1.0).Return(nil)

	// Execute
	err := npc.CraftGoods(context.Background(), recipeMgr, market, stationFinder, crafter)

	// Verify
	assert.NoError(t, err)
	assert.Less(t, npc.Desires.TaskCompletion, 80.0, "Desire should decrease")

	recipeMgr.AssertExpectations(t)
	market.AssertExpectations(t)
	stationFinder.AssertExpectations(t)
	crafter.AssertExpectations(t)
	inventory.AssertExpectations(t)
}

func TestCraftGoods_LowDesire(t *testing.T) {
	npc := &NPC{
		Desires: &Desires{
			TaskCompletion: 20.0, // Low need
		},
	}

	err := npc.CraftGoods(context.Background(), nil, nil, nil, nil)
	assert.NoError(t, err)
}

func TestCraftGoods_NoIngredients(t *testing.T) {
	recipeMgr := new(MockRecipeManager)
	inventory := new(MockInventoryManager)

	npc := &NPC{
		ID:        uuid.New(),
		Inventory: inventory,
		Desires: &Desires{
			TaskCompletion: 80.0,
		},
	}

	inventory.On("GetContents").Return(map[uuid.UUID]int{})
	recipeMgr.On("GetCraftableRecipes", mock.Anything, mock.Anything).Return([]*crafting.Recipe{}, nil)

	err := npc.CraftGoods(context.Background(), recipeMgr, nil, nil, nil)
	assert.Equal(t, ErrNoIngredients, err)
}
