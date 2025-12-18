package inventory

import (
	"fmt"
	"strings"

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
	// In-memory store for P0, TODO: Postgres persistence
	inventories map[uuid.UUID][]Item
}

func NewService(entityService *entity.Service) *Service {
	return &Service{
		entityService: entityService,
		inventories:   make(map[uuid.UUID][]Item),
	}
}

// AddItem adds an item to a character's inventory
func (s *Service) AddItem(charID uuid.UUID, item Item) {
	if s.inventories[charID] == nil {
		s.inventories[charID] = []Item{}
	}
	s.inventories[charID] = append(s.inventories[charID], item)
}

// RemoveItem removes an item from inventory by name (first match)
func (s *Service) RemoveItem(charID uuid.UUID, itemName string) (Item, error) {
	items := s.inventories[charID]
	for i, item := range items {
		if strings.EqualFold(item.Name, itemName) {
			// Found it
			s.inventories[charID] = append(items[:i], items[i+1:]...)
			return item, nil
		}
	}
	return Item{}, fmt.Errorf("item '%s' not found in inventory", itemName)
}

// GetInventory returns all items for a character
func (s *Service) GetInventory(charID uuid.UUID) []Item {
	if s.inventories[charID] == nil {
		return []Item{}
	}
	return s.inventories[charID]
}
