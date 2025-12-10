package item

import (
	"github.com/google/uuid"
)

// Slot constants
const (
	SlotMainHand = "main_hand"
	SlotOffHand  = "off_hand"
	SlotHead     = "head"
	SlotChest    = "chest"
	SlotLegs     = "legs"
	SlotFeet     = "feet"
	SlotNeck     = "neck"
	SlotRing1    = "ring1"
	SlotRing2    = "ring2"
)

// ItemProperties holds dynamic properties for an item
type ItemProperties struct {
	IsEquippable bool              `json:"is_equippable"`
	Slot         string            `json:"slot,omitempty"`
	DamageType   string            `json:"damage_type,omitempty"`
	ArmorValue   int               `json:"armor_value,omitempty"`
	Effects      map[string]string `json:"effects,omitempty"`
}

// Item represents a distinct object in the game
type Item struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Weight        float64        `json:"weight"` // in kg
	StackSize     int            `json:"stack_size"`
	Durability    int            `json:"durability"` // 0-100
	MaxDurability int            `json:"max_durability"`
	Properties    ItemProperties `json:"properties"`
}

// Inventory represents a character's collection of items
type Inventory struct {
	Items         map[uuid.UUID]Item `json:"items"`
	CurrentWeight float64            `json:"current_weight"`
}

// Equipment represents the items currently equipped by a character
type Equipment struct {
	MainHand *Item `json:"main_hand,omitempty"`
	OffHand  *Item `json:"off_hand,omitempty"`
	Head     *Item `json:"head,omitempty"`
	Chest    *Item `json:"chest,omitempty"`
	Legs     *Item `json:"legs,omitempty"`
	Feet     *Item `json:"feet,omitempty"`
	Neck     *Item `json:"neck,omitempty"`
	Ring1    *Item `json:"ring1,omitempty"`
	Ring2    *Item `json:"ring2,omitempty"`
}
