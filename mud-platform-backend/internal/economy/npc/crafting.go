package npc

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"mud-platform-backend/internal/economy/crafting"

	"github.com/google/uuid"
)

var (
	ErrNoIngredients     = errors.New("no ingredients for any recipe")
	ErrNoCraftingStation = errors.New("no suitable crafting station found")
	ErrCraftingFailed    = errors.New("crafting failed")
)

// RecipeManager interface for dependency injection
type RecipeManager interface {
	GetKnownRecipes(entityID uuid.UUID) ([]*crafting.Recipe, error)
	GetCraftableRecipes(entityID uuid.UUID, resources map[uuid.UUID]int) ([]*crafting.Recipe, error)
	ImproveProficiency(entityID, recipeID uuid.UUID, amount float64) error
}

// MarketDataProvider interface for dependency injection
type MarketDataProvider interface {
	GetAveragePrice(locationID, itemID uuid.UUID) float64
	GetLocalDemand(locationID, itemID uuid.UUID) int
}

// StationFinder interface for dependency injection
type StationFinder interface {
	FindNearestStation(location uuid.UUID, stationType string, minTier int) (*crafting.CraftingStation, error)
}

// Crafter interface for dependency injection
type Crafter interface {
	Craft(recipeID uuid.UUID, skillLevel int, toolQuality float64, stationTier int) (*crafting.CraftResult, error)
}

// CraftGoods executes the autonomous crafting loop
func (npc *NPC) CraftGoods(
	ctx context.Context,
	recipeMgr RecipeManager,
	market MarketDataProvider,
	stationFinder StationFinder,
	crafter Crafter,
) error {
	// 1. Check if crafting desire active
	// Assuming TaskCompletion is the relevant desire
	// In a real implementation, we'd have a more complex desire system
	if npc.Desires.TaskCompletion < 40 {
		return nil // Not motivated to craft
	}

	// 2. Get craftable recipes
	// We need to convert inventory to map[uuid.UUID]int
	// Assuming InventoryManager has a method for this, or we mock it
	resources := npc.Inventory.GetContents()

	craftable, err := recipeMgr.GetCraftableRecipes(npc.ID, resources)
	if err != nil {
		return fmt.Errorf("failed to get craftable recipes: %w", err)
	}

	if len(craftable) == 0 {
		return ErrNoIngredients
	}

	// 3. Prioritize by profit margin and demand
	bestRecipe := npc.selectBestRecipe(craftable, market)
	if bestRecipe == nil {
		return ErrNoIngredients
	}

	// 4. Find station
	if bestRecipe.RequiredStation != nil {
		_, err := stationFinder.FindNearestStation(
			npc.Location,
			bestRecipe.RequiredStation.StationType,
			bestRecipe.RequiredStation.MinStationTier,
		)
		if err != nil {
			return ErrNoCraftingStation
		}
		// In a real system, we'd move to the station here
	}

	// 5. Craft item
	skillLevel := 0
	if bestRecipe.RequiredSkill != "" {
		skillLevel = npc.Skills[bestRecipe.RequiredSkill]
	}

	// Determine station tier (simplified: assume we found a valid one)
	stationTier := 1
	if bestRecipe.RequiredStation != nil {
		stationTier = bestRecipe.RequiredStation.MinStationTier
	}

	result, err := crafter.Craft(bestRecipe.RecipeID, skillLevel, npc.Equipment.ToolQuality, stationTier)
	if err != nil {
		// Crafting failed
		// Improve proficiency slightly even on failure
		_ = recipeMgr.ImproveProficiency(npc.ID, bestRecipe.RecipeID, 0.5)
		return ErrCraftingFailed
	}

	// 6. Add to inventory
	if err := npc.Inventory.Add(result.Item.ItemID, result.Item.Quantity); err != nil {
		return ErrInventoryFull
	}

	// Handle byproducts
	for _, byproduct := range result.ByProducts {
		_ = npc.Inventory.Add(byproduct.ItemID, byproduct.Quantity)
	}

	// 7. Improve recipe proficiency
	// Calculate gain based on difficulty (simplified)
	gain := 1.0
	if bestRecipe.Difficulty == crafting.DifficultyHard {
		gain = 2.0
	}
	_ = recipeMgr.ImproveProficiency(npc.ID, bestRecipe.RecipeID, gain)

	// 8. Decrease task completion desire
	npc.Desires.TaskCompletion -= 30.0
	if npc.Desires.TaskCompletion < 0 {
		npc.Desires.TaskCompletion = 0
	}

	return nil
}

func (npc *NPC) selectBestRecipe(recipes []*crafting.Recipe, market MarketDataProvider) *crafting.Recipe {
	if len(recipes) == 0 {
		return nil
	}

	// Sort by potential profit (Price * Demand)
	// This is a heuristic: high price * high demand = high priority
	sort.Slice(recipes, func(i, j int) bool {
		scoreI := calculateScore(recipes[i], npc.Location, market)
		scoreJ := calculateScore(recipes[j], npc.Location, market)
		return scoreI > scoreJ
	})

	return recipes[0]
}

func calculateScore(recipe *crafting.Recipe, locationID uuid.UUID, market MarketDataProvider) float64 {
	price := market.GetAveragePrice(locationID, recipe.Output.ItemID)
	demand := float64(market.GetLocalDemand(locationID, recipe.Output.ItemID))

	// Avoid zero demand causing zero score if price is high
	if demand == 0 {
		demand = 0.5 // Baseline interest
	}

	return price * demand
}
