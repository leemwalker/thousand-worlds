package item

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeItemPickedUp          = "ItemPickedUp"
	EventTypeItemDropped           = "ItemDropped"
	EventTypeItemUsed              = "ItemUsed"
	EventTypeItemEquipped          = "ItemEquipped"
	EventTypeItemUnequipped        = "ItemUnequipped"
	EventTypeItemDurabilityChanged = "ItemDurabilityChanged"
)

type ItemPickedUpEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	ItemID      uuid.UUID `json:"item_id"`
	Weight      float64   `json:"weight"`
	Timestamp   time.Time `json:"timestamp"`
}

type ItemDroppedEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	ItemID      uuid.UUID `json:"item_id"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	Z           float64   `json:"z"`
	Timestamp   time.Time `json:"timestamp"`
}

type ItemUsedEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	ItemID      uuid.UUID `json:"item_id"`
	Effect      string    `json:"effect"`
	Timestamp   time.Time `json:"timestamp"`
}

type ItemEquippedEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	ItemID      uuid.UUID `json:"item_id"`
	Slot        string    `json:"slot"`
	Timestamp   time.Time `json:"timestamp"`
}

type ItemUnequippedEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	ItemID      uuid.UUID `json:"item_id"`
	Slot        string    `json:"slot"`
	Timestamp   time.Time `json:"timestamp"`
}

type ItemDurabilityChangedEvent struct {
	ItemID    uuid.UUID `json:"item_id"`
	OldValue  int       `json:"old_value"`
	NewValue  int       `json:"new_value"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}
