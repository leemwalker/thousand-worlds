package item

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEquipmentManager_Equip(t *testing.T) {
	im := NewInventoryManager(100) // High capacity
	em := NewEquipmentManager(im)

	// Create equippable item
	sword := Item{
		ID:   uuid.New(),
		Name: "Iron Sword",
		Properties: ItemProperties{
			IsEquippable: true,
			Slot:         SlotMainHand,
		},
	}
	im.AddItem(sword)

	// Equip valid
	err := em.Equip(sword.ID, SlotMainHand)
	assert.NoError(t, err)
	assert.Equal(t, &sword, em.equipment.MainHand)
	assert.False(t, im.HasItem(sword.ID)) // Should be removed from inventory

	// Equip invalid slot
	helmet := Item{
		ID:   uuid.New(),
		Name: "Helmet",
		Properties: ItemProperties{
			IsEquippable: true,
			Slot:         SlotHead,
		},
	}
	im.AddItem(helmet)
	err = em.Equip(helmet.ID, SlotMainHand)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires head")

	// Equip occupied slot
	sword2 := Item{
		ID:   uuid.New(),
		Name: "Another Sword",
		Properties: ItemProperties{
			IsEquippable: true,
			Slot:         SlotMainHand,
		},
	}
	im.AddItem(sword2)
	err = em.Equip(sword2.ID, SlotMainHand)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already occupied")
}

func TestEquipmentManager_Unequip(t *testing.T) {
	im := NewInventoryManager(100)
	em := NewEquipmentManager(im)

	sword := Item{
		ID:     uuid.New(),
		Name:   "Iron Sword",
		Weight: 5.0,
		Properties: ItemProperties{
			IsEquippable: true,
			Slot:         SlotMainHand,
		},
	}
	im.AddItem(sword)
	em.Equip(sword.ID, SlotMainHand)

	// Unequip valid
	err := em.Unequip(SlotMainHand)
	assert.NoError(t, err)
	assert.Nil(t, em.equipment.MainHand)
	assert.True(t, im.HasItem(sword.ID))

	// Unequip empty
	err = em.Unequip(SlotOffHand)
	assert.Error(t, err)
}

func TestDurabilityManager(t *testing.T) {
	dm := NewDurabilityManager()

	// Weapon
	sword := Item{
		Name:          "Sword",
		Durability:    10,
		MaxDurability: 100,
		Properties: ItemProperties{
			DamageType: "slashing",
		},
	}

	// Degrade
	dm.Degrade(&sword, 5)
	assert.Equal(t, 5, sword.Durability)
	assert.False(t, dm.IsBroken(sword))
	assert.Equal(t, 1.0, dm.GetEffectiveness(sword))

	// Break
	dm.Degrade(&sword, 10)
	assert.Equal(t, 0, sword.Durability)
	assert.True(t, dm.IsBroken(sword))
	assert.Equal(t, 0.5, dm.GetEffectiveness(sword)) // 50% for broken weapon

	// Repair
	dm.Repair(&sword, 50)
	assert.Equal(t, 50, sword.Durability)
	assert.False(t, dm.IsBroken(sword))
}
