package item

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CommandHandler handles high-level item commands
type CommandHandler struct {
	inventory *InventoryManager
	equipment *EquipmentManager
}

// NewCommandHandler creates a new handler
func NewCommandHandler(im *InventoryManager, em *EquipmentManager) *CommandHandler {
	return &CommandHandler{
		inventory: im,
		equipment: em,
	}
}

// Pickup adds an item to inventory and returns an event
func (h *CommandHandler) Pickup(charID uuid.UUID, item Item) (*ItemPickedUpEvent, error) {
	if err := h.inventory.AddItem(item); err != nil {
		return nil, err
	}

	return &ItemPickedUpEvent{
		CharacterID: charID,
		ItemID:      item.ID,
		Weight:      item.Weight,
		Timestamp:   time.Now(),
	}, nil
}

// Drop removes an item from inventory and returns an event
func (h *CommandHandler) Drop(charID uuid.UUID, itemID uuid.UUID, x, y, z float64) (*ItemDroppedEvent, error) {
	_, err := h.inventory.RemoveItem(itemID)
	if err != nil {
		return nil, err
	}

	return &ItemDroppedEvent{
		CharacterID: charID,
		ItemID:      itemID,
		X:           x,
		Y:           y,
		Z:           z,
		Timestamp:   time.Now(),
	}, nil
}

// Equip moves an item to a slot and returns an event
func (h *CommandHandler) Equip(charID uuid.UUID, itemID uuid.UUID, slot string) (*ItemEquippedEvent, error) {
	if err := h.equipment.Equip(itemID, slot); err != nil {
		return nil, err
	}

	return &ItemEquippedEvent{
		CharacterID: charID,
		ItemID:      itemID,
		Slot:        slot,
		Timestamp:   time.Now(),
	}, nil
}

// Unequip moves an item from a slot to inventory and returns an event
func (h *CommandHandler) Unequip(charID uuid.UUID, slot string) (*ItemUnequippedEvent, error) {
	// We need the item ID for the event, so let's peek at the slot first
	item := h.equipment.getSlotItem(slot)
	if item == nil {
		return nil, fmt.Errorf("slot %s is empty", slot)
	}
	itemID := item.ID

	if err := h.equipment.Unequip(slot); err != nil {
		return nil, err
	}

	return &ItemUnequippedEvent{
		CharacterID: charID,
		ItemID:      itemID,
		Slot:        slot,
		Timestamp:   time.Now(),
	}, nil
}

// Use consumes an item (placeholder logic) and returns an event
func (h *CommandHandler) Use(charID uuid.UUID, itemID uuid.UUID) (*ItemUsedEvent, error) {
	// For now, just remove it to simulate consumption if it's consumable
	// In a real system, we'd check properties
	item, err := h.inventory.GetItem(itemID)
	if err != nil {
		return nil, err
	}

	// Assume all used items are consumed for this phase
	_, err = h.inventory.RemoveItem(itemID)
	if err != nil {
		return nil, err
	}

	return &ItemUsedEvent{
		CharacterID: charID,
		ItemID:      itemID,
		Effect:      fmt.Sprintf("Used %s", item.Name),
		Timestamp:   time.Now(),
	}, nil
}
