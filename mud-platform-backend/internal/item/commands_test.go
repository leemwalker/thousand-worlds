package item

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCommandHandler(t *testing.T) {
	im := NewInventoryManager(100)
	em := NewEquipmentManager(im)
	h := NewCommandHandler(im, em)
	charID := uuid.New()

	// Pickup
	item := Item{ID: uuid.New(), Name: "Potion", Weight: 0.5}
	pickupEvent, err := h.Pickup(charID, item)
	assert.NoError(t, err)
	assert.Equal(t, item.ID, pickupEvent.ItemID)
	assert.True(t, im.HasItem(item.ID))

	// Use
	useEvent, err := h.Use(charID, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, item.ID, useEvent.ItemID)
	assert.False(t, im.HasItem(item.ID)) // Consumed

	// Pickup Sword
	sword := Item{
		ID:         uuid.New(),
		Name:       "Sword",
		Properties: ItemProperties{IsEquippable: true, Slot: SlotMainHand},
	}
	h.Pickup(charID, sword)

	// Equip
	equipEvent, err := h.Equip(charID, sword.ID, SlotMainHand)
	assert.NoError(t, err)
	assert.Equal(t, SlotMainHand, equipEvent.Slot)
	assert.Equal(t, &sword, em.equipment.MainHand)

	// Unequip
	unequipEvent, err := h.Unequip(charID, SlotMainHand)
	assert.NoError(t, err)
	assert.Equal(t, SlotMainHand, unequipEvent.Slot)
	assert.Nil(t, em.equipment.MainHand)
	assert.True(t, im.HasItem(sword.ID))

	// Drop
	dropEvent, err := h.Drop(charID, sword.ID, 1, 2, 3)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, dropEvent.X)
	assert.False(t, im.HasItem(sword.ID))
}
