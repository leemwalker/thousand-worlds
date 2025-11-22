package item

// DurabilityManager handles item durability
type DurabilityManager struct{}

// NewDurabilityManager creates a new manager
func NewDurabilityManager() *DurabilityManager {
	return &DurabilityManager{}
}

// Degrade reduces item durability
func (dm *DurabilityManager) Degrade(item *Item, amount int) {
	item.Durability -= amount
	if item.Durability < 0 {
		item.Durability = 0
	}
}

// Repair restores item durability
func (dm *DurabilityManager) Repair(item *Item, amount int) {
	item.Durability += amount
	if item.Durability > item.MaxDurability {
		item.Durability = item.MaxDurability
	}
}

// IsBroken checks if item is broken
func (dm *DurabilityManager) IsBroken(item Item) bool {
	return item.Durability <= 0
}

// GetEffectiveness returns the effectiveness multiplier (1.0 normal, 0.5 broken weapon, 0.0 broken armor)
// This is a simplified logic; real logic might depend on item type
func (dm *DurabilityManager) GetEffectiveness(item Item) float64 {
	if item.Durability > 0 {
		return 1.0
	}

	// Broken logic
	// If it has damage type (weapon), 50% effectiveness
	if item.Properties.DamageType != "" {
		return 0.5
	}

	// If it has armor value (armor), 0% effectiveness
	if item.Properties.ArmorValue > 0 {
		return 0.0
	}

	return 0.0 // Default broken state
}
