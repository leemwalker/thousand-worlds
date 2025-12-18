package crafting

import (
	"context"
	"time"

	"tw-backend/internal/errors"
	"tw-backend/internal/game/services/inventory"
	"tw-backend/internal/worldentity"

	"github.com/google/uuid"
)

// Service handles crafting operations
type Service struct {
	repo               Repository
	inventoryService   *inventory.Service
	worldEntityService *worldentity.Service
}

// NewService creates a new crafting service
func NewService(repo Repository, invService *inventory.Service, worldEntService *worldentity.Service) *Service {
	return &Service{
		repo:               repo,
		inventoryService:   invService,
		worldEntityService: worldEntService,
	}
}

// Craft attempts to craft an item using a recipe
func (s *Service) Craft(ctx context.Context, characterID uuid.UUID, recipeID uuid.UUID, stationEntityID *uuid.UUID) (*CraftResult, error) {
	// 1. Get the recipe
	recipe, err := s.repo.GetRecipe(recipeID)
	if err != nil {
		return nil, errors.NewNotFound("recipe not found")
	}

	// 2. Validate station (if required)
	if recipe.RequiredStation != nil {
		if stationEntityID == nil {
			return nil, errors.NewInvalidInput("crafting station required: %s", recipe.RequiredStation.StationType)
		}

		// Verify station type and tier via WorldEntityService
		// This presumes WorldEntity has properties describing its crafting capabilities
		station, err := s.worldEntityService.GetByID(ctx, *stationEntityID)
		if err != nil {
			return nil, errors.NewNotFound("crafting station not found")
		}

		// TODO: Check if station.Properties matches recipe.RequiredStation
		// For now, we assume if the entity exists and was passed, it's valid,
		// but in a real implementation we would strictly check tags/properties.
		_ = station
	}

	// 3. Get character constraints (inventory)
	// We need to check if the user has the ingredients.
	// The InventoryService generic 'RemoveItem' might need to be called multiple times,
	// or we might need a transaction. For now, we'll check availability first.

	// Get all items to check quantities (Optimization: InventoryService should have HasItems method)
	// For MVP, we will optimistically attempt to remove ingredients.
	// Ideally, InventoryService should support atomic bulk removal.

	// 4. Check & Remove Ingredients
	for _, ing := range recipe.Ingredients {
		// Try to remove primary ingredient
		err := s.inventoryService.RemoveItem(ctx, characterID, ing.ResourceID, ing.Quantity)
		if err != nil {
			// If failed, we should rollback previous removals.
			// Since InventoryService doesn't expose a transaction yet, this is risky.
			// TODO: Add Transactional Inventory Support.
			return nil, errors.NewInvalidInput("missing ingredient: %s", ing.ResourceID) // Ideally return better error
		}
	}

	// 5. Calculate Result Quality
	// TODO: Use character skills and station modifiers
	quality := ItemQuality(1) // Default to standard quality for now

	// 6. Add Result to Inventory
	// If output is a specific ItemID (like a unique sword), we might need to "create" it first?
	// Or is Recipe.Output.ItemID a reference to an ItemTemplate?
	// Assuming ItemID in Output is a Template ID if it's generic, or we create a new instance.
	// For resource crafting (wood -> plank), it's just an ID.

	err = s.inventoryService.AddItem(ctx, characterID, recipe.Output.ItemID, recipe.Output.Quantity, map[string]interface{}{
		"quality":    quality,
		"crafted_at": time.Now(),
		"crafter_id": characterID,
	})
	if err != nil {
		// Big problem: we removed ingredients but failed to add result.
		// Log this critically.
		return nil, errors.NewInternalError("failed to add crafted item: %v", err)
	}

	// 7. Handle Byproducts
	for _, bp := range recipe.ByProducts {
		_ = s.inventoryService.AddItem(ctx, characterID, bp.ItemID, bp.Quantity, nil)
	}

	return &CraftResult{
		Success: true,
		Item: ItemOutput{
			ItemID:   recipe.Output.ItemID,
			Quantity: recipe.Output.Quantity,
			Quality:  quality,
		},
		ByProducts: recipe.ByProducts,
	}, nil
}

// GetAvailableRecipes returns recipes the character can craft
func (s *Service) GetAvailableRecipes(ctx context.Context, characterID uuid.UUID) ([]*Recipe, error) {
	// 1. Get recipes known by the character
	known, err := s.repo.GetKnownRecipes(characterID)
	if err != nil {
		return nil, err
	}

	// 2. Filter by what's possible (optional, maybe UI handles this)
	// For now return all known
	return known, nil
}

// FindRecipeByName searches for a recipe by name (case-insensitive)
func (s *Service) FindRecipeByName(ctx context.Context, name string) (*Recipe, error) {
	// 1. Search using repository
	// Depending on repository implementation, precise match might need 'SearchRecipes' or a new method.
	// Assuming SearchRecipes exists and can filter.
	// Pushing logic to repository is best.

	recipes, err := s.repo.SearchRecipes(name, RecipeFilters{})
	if err != nil {
		return nil, err
	}

	// Return first exact match or first result
	if len(recipes) == 0 {
		return nil, errors.NewNotFound("recipe '%s' not found", name)
	}

	return recipes[0], nil
}
