package item

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateMaxCarryWeight(t *testing.T) {
	// Might 10 -> 50kg
	assert.Equal(t, 50.0, CalculateMaxCarryWeight(10))
	// Might 50 -> 250kg
	assert.Equal(t, 250.0, CalculateMaxCarryWeight(50))
}

func TestInventoryManager_AddItem(t *testing.T) {
	im := NewInventoryManager(10) // 50kg limit
	item := Item{
		ID:     uuid.New(),
		Name:   "Heavy Rock",
		Weight: 40.0,
	}

	// Add valid item
	err := im.AddItem(item)
	assert.NoError(t, err)
	assert.Equal(t, 40.0, im.inventory.CurrentWeight)

	// Add item exceeding limit
	heavyItem := Item{
		ID:     uuid.New(),
		Name:   "Another Rock",
		Weight: 20.0,
	}
	err = im.AddItem(heavyItem)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds carry weight limit")
	assert.Equal(t, 40.0, im.inventory.CurrentWeight)
}

func TestInventoryManager_RemoveItem(t *testing.T) {
	im := NewInventoryManager(10)
	item := Item{
		ID:     uuid.New(),
		Name:   "Rock",
		Weight: 10.0,
	}
	im.AddItem(item)

	// Remove existing item
	removed, err := im.RemoveItem(item.ID)
	assert.NoError(t, err)
	assert.Equal(t, item.ID, removed.ID)
	assert.Equal(t, 0.0, im.inventory.CurrentWeight)

	// Remove non-existent item
	_, err = im.RemoveItem(uuid.New())
	assert.Error(t, err)
}
