package item

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// InventoryManager handles inventory operations
type InventoryManager struct {
	mu             sync.RWMutex
	inventory      *Inventory
	maxCarryWeight float64
}

// NewInventoryManager creates a new manager
func NewInventoryManager(might int) *InventoryManager {
	return &InventoryManager{
		inventory: &Inventory{
			Items: make(map[uuid.UUID]Item),
		},
		maxCarryWeight: CalculateMaxCarryWeight(might),
	}
}

// CalculateMaxCarryWeight returns the max weight in kg
func CalculateMaxCarryWeight(might int) float64 {
	return float64(might) * 5.0
}

// AddItem adds an item to the inventory if weight permits
func (im *InventoryManager) AddItem(item Item) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if im.inventory.CurrentWeight+item.Weight > im.maxCarryWeight {
		return fmt.Errorf("cannot pick up %s: exceeds carry weight limit", item.Name)
	}

	im.inventory.Items[item.ID] = item
	im.inventory.CurrentWeight += item.Weight
	return nil
}

// RemoveItem removes an item from the inventory
func (im *InventoryManager) RemoveItem(itemID uuid.UUID) (Item, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	item, exists := im.inventory.Items[itemID]
	if !exists {
		return Item{}, fmt.Errorf("item not found")
	}

	delete(im.inventory.Items, itemID)
	im.inventory.CurrentWeight -= item.Weight
	return item, nil
}

// GetItem returns an item from the inventory
func (im *InventoryManager) GetItem(itemID uuid.UUID) (Item, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	item, exists := im.inventory.Items[itemID]
	if !exists {
		return Item{}, fmt.Errorf("item not found")
	}
	return item, nil
}

// HasItem checks if the inventory contains an item
func (im *InventoryManager) HasItem(itemID uuid.UUID) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	_, exists := im.inventory.Items[itemID]
	return exists
}
