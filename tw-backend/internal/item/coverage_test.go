package item

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCoverage_Durability(t *testing.T) {
	dm := NewDurabilityManager()

	// Test Repair Capping
	item := Item{Durability: 90, MaxDurability: 100}
	dm.Repair(&item, 20)
	assert.Equal(t, 100, item.Durability)

	// Test Broken Armor Effectiveness
	armor := Item{
		Durability: 0,
		Properties: ItemProperties{ArmorValue: 10},
	}
	assert.Equal(t, 0.0, dm.GetEffectiveness(armor))

	// Test Default Broken Effectiveness
	misc := Item{Durability: 0}
	assert.Equal(t, 0.0, dm.GetEffectiveness(misc))
}

func TestCoverage_EquipmentSlots(t *testing.T) {
	im := NewInventoryManager(100)
	em := NewEquipmentManager(im)
	item := Item{ID: uuid.New(), Name: "Test", Properties: ItemProperties{IsEquippable: true}}

	// Test all setSlotItem cases
	slots := []string{
		SlotMainHand, SlotOffHand, SlotHead, SlotChest,
		SlotLegs, SlotFeet, SlotNeck, SlotRing1, SlotRing2,
	}

	for _, slot := range slots {
		em.setSlotItem(slot, &item)
		assert.Equal(t, &item, em.getSlotItem(slot))
		em.setSlotItem(slot, nil)
		assert.Nil(t, em.getSlotItem(slot))
	}

	// Test default case
	assert.Nil(t, em.getSlotItem("invalid_slot"))
}

func TestCoverage_Commands_Errors(t *testing.T) {
	im := NewInventoryManager(10) // Low weight limit
	em := NewEquipmentManager(im)
	h := NewCommandHandler(im, em)
	charID := uuid.New()

	// Pickup Error (Overweight)
	heavyItem := Item{ID: uuid.New(), Weight: 100}
	_, err := h.Pickup(charID, heavyItem)
	assert.Error(t, err)

	// Drop Error (Not Found)
	_, err = h.Drop(charID, uuid.New(), 0, 0, 0)
	assert.Error(t, err)

	// Equip Error (Not Found)
	_, err = h.Equip(charID, uuid.New(), SlotMainHand)
	assert.Error(t, err)

	// Unequip Error (Empty)
	_, err = h.Unequip(charID, SlotMainHand)
	assert.Error(t, err)

	// Use Error (Not Found)
	_, err = h.Use(charID, uuid.New())
	assert.Error(t, err)
}

func TestCoverage_Inventory_GetItem_Error(t *testing.T) {
	im := NewInventoryManager(100)
	_, err := im.GetItem(uuid.New())
	assert.Error(t, err)
}
