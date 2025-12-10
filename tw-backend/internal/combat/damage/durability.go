package damage

// DurabilityResult holds the outcome of durability loss
type DurabilityResult struct {
	CurrentDurability int
	IsBroken          bool
	DamageModifier    float64 // Multiplier for damage output (e.g., 0.9 for <50%)
}

// ReduceDurability reduces weapon durability based on action
func ReduceDurability(weapon *Weapon, amount int) DurabilityResult {
	weapon.Durability -= amount
	if weapon.Durability < 0 {
		weapon.Durability = 0
	}

	return GetDurabilityStatus(weapon)
}

// GetDurabilityStatus calculates the current status and modifiers
func GetDurabilityStatus(weapon *Weapon) DurabilityResult {
	if weapon.Durability == 0 {
		return DurabilityResult{
			CurrentDurability: 0,
			IsBroken:          true,
			DamageModifier:    0.0,
		}
	}

	percent := float64(weapon.Durability) / float64(weapon.MaxDurability)
	modifier := 1.0

	if percent < 0.25 {
		modifier = 0.75 // -25% damage
	} else if percent < 0.50 {
		modifier = 0.90 // -10% damage
	}

	return DurabilityResult{
		CurrentDurability: weapon.Durability,
		IsBroken:          false,
		DamageModifier:    modifier,
	}
}

// ReduceArmorDurability reduces armor durability (simple -1 per hit usually)
func ReduceArmorDurability(armor *Armor, amount int) {
	armor.Durability -= amount
	if armor.Durability < 0 {
		armor.Durability = 0
	}
}

// GetArmorEffectivenessModifier returns the multiplier for armor effectiveness based on durability
func GetArmorEffectivenessModifier(armor *Armor) float64 {
	if armor.MaxDurability == 0 {
		return 0
	}
	return float64(armor.Durability) / float64(armor.MaxDurability)
}
