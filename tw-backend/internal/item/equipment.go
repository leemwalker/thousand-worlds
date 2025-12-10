package item

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// EquipmentManager handles equipment operations
type EquipmentManager struct {
	mu        sync.RWMutex
	equipment *Equipment
	inventory *InventoryManager
}

// NewEquipmentManager creates a new manager
func NewEquipmentManager(im *InventoryManager) *EquipmentManager {
	return &EquipmentManager{
		equipment: &Equipment{},
		inventory: im,
	}
}

// Equip moves an item from inventory to a slot
func (em *EquipmentManager) Equip(itemID uuid.UUID, slot string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// 1. Check if item exists in inventory
	item, err := em.inventory.GetItem(itemID)
	if err != nil {
		return err
	}

	// 2. Validate item properties for the slot
	if !item.Properties.IsEquippable {
		return fmt.Errorf("item is not equippable")
	}
	if item.Properties.Slot != slot {
		// Allow rings in either ring slot if generic "ring"
		if item.Properties.Slot == "ring" && (slot == SlotRing1 || slot == SlotRing2) {
			// valid
		} else {
			return fmt.Errorf("item cannot be equipped in slot %s (requires %s)", slot, item.Properties.Slot)
		}
	}

	// 3. Check if slot is occupied
	if em.getSlotItem(slot) != nil {
		return fmt.Errorf("slot %s is already occupied", slot)
	}

	// 4. Move item: Remove from inventory (logic handled by InventoryManager, but we need to be careful about lock ordering if we call it here)
	// Since we are holding em.mu, we should be careful. InventoryManager has its own mutex.
	// It's safer to release em.mu before calling inventory methods if they take time, but here they are fast.
	// However, to avoid deadlock, we should probably not hold em.mu while calling inventory.RemoveItem if inventory.RemoveItem could call back into EquipmentManager (unlikely).
	// But wait, RemoveItem modifies inventory state.

	// Let's actually remove it from inventory first.
	// We need to release the lock to call RemoveItem to avoid potential issues if we expand later,
	// but for now, let's just call it.
	// Actually, we need to ensure the item is still there.

	// Better approach:
	// 1. Remove from inventory.
	// 2. If successful, put in slot.
	// 3. If put in slot fails (shouldn't here), put back in inventory.

	// We need to release our lock to call inventory methods to be safe?
	// No, InventoryManager uses its own mutex. As long as InventoryManager doesn't call EquipmentManager, we are fine.

	_, err = em.inventory.RemoveItem(itemID)
	if err != nil {
		return err
	}

	// 5. Set slot
	em.setSlotItem(slot, &item)

	return nil
}

// Unequip moves an item from a slot to inventory
func (em *EquipmentManager) Unequip(slot string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// 1. Check if slot has item
	item := em.getSlotItem(slot)
	if item == nil {
		return fmt.Errorf("slot %s is empty", slot)
	}

	// 2. Add to inventory
	err := em.inventory.AddItem(*item)
	if err != nil {
		return fmt.Errorf("cannot unequip: %w", err)
	}

	// 3. Clear slot
	em.setSlotItem(slot, nil)

	return nil
}

// Helper to get item from slot
func (em *EquipmentManager) getSlotItem(slot string) *Item {
	switch slot {
	case SlotMainHand:
		return em.equipment.MainHand
	case SlotOffHand:
		return em.equipment.OffHand
	case SlotHead:
		return em.equipment.Head
	case SlotChest:
		return em.equipment.Chest
	case SlotLegs:
		return em.equipment.Legs
	case SlotFeet:
		return em.equipment.Feet
	case SlotNeck:
		return em.equipment.Neck
	case SlotRing1:
		return em.equipment.Ring1
	case SlotRing2:
		return em.equipment.Ring2
	default:
		return nil
	}
}

// Helper to set item in slot
func (em *EquipmentManager) setSlotItem(slot string, item *Item) {
	switch slot {
	case SlotMainHand:
		em.equipment.MainHand = item
	case SlotOffHand:
		em.equipment.OffHand = item
	case SlotHead:
		em.equipment.Head = item
	case SlotChest:
		em.equipment.Chest = item
	case SlotLegs:
		em.equipment.Legs = item
	case SlotFeet:
		em.equipment.Feet = item
	case SlotNeck:
		em.equipment.Neck = item
	case SlotRing1:
		em.equipment.Ring1 = item
	case SlotRing2:
		em.equipment.Ring2 = item
	}
}
