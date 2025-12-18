package inventory

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"tw-backend/internal/game/services/entity"
)

type Item struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Service struct {
	entityService *entity.Service
	repo          Repository
}

func NewService(entityService *entity.Service, repo Repository) *Service {
	return &Service{
		entityService: entityService,
		repo:          repo,
	}
}

// AddItem adds an item to a character's inventory
func (s *Service) AddItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int, metadata map[string]interface{}) error {
	return s.repo.AddItem(ctx, charID, itemID, quantity, metadata)
}

// RemoveItem removes an item from inventory by ID
func (s *Service) RemoveItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int) error {
	return s.repo.RemoveItem(ctx, charID, itemID, quantity)
}

// RemoveItemByName removes an item from inventory by name (first match)
func (s *Service) RemoveItemByName(ctx context.Context, charID uuid.UUID, itemName string) (Item, error) {
	// Logic needs to fetch inventory first to find ID by name
	// PROPOSAL: repository should handle "RemoveByName" or we fetch first?
	// For P0, let's fetch all and filter.
	items, err := s.repo.GetInventory(ctx, charID)
	if err != nil {
		return Item{}, err
	}

	for _, invItem := range items {
		// Assuming InventoryItem.Name is populated
		if invItem.Name == itemName { // Simple case insensitive in future
			err := s.repo.RemoveItem(ctx, charID, invItem.ItemID, 1)
			if err != nil {
				return Item{}, err
			}
			return Item{ID: invItem.ItemID, Name: invItem.Name, Description: invItem.Description}, nil
		}
	}

	return Item{}, fmt.Errorf("item '%s' not found in inventory", itemName)
}

// GetInventory returns all items for a character
func (s *Service) GetInventory(ctx context.Context, charID uuid.UUID) ([]InventoryItem, error) {
	return s.repo.GetInventory(ctx, charID)
}
