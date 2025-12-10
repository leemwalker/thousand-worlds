package damage

import (
	"mud-platform-backend/internal/character"
)

// DamageResult holds the final calculation details
type DamageResult struct {
	FinalDamage int
	RawDamage   int
	IsCritical  bool
	IsFumble    bool
	Blocked     int
}

// CalculateDamage computes the final damage dealt
func CalculateDamage(
	attackerAttrs character.Attributes,
	weapon *Weapon,
	weaponSkill int,
	targetArmor *Armor,
	roll int,
	isHeavyAttack bool,
) DamageResult {
	// 1. Critical Check
	crit := CalculateCritical(roll, attackerAttrs.Cunning, isHeavyAttack)
	if crit.IsCriticalFailure {
		return DamageResult{IsFumble: true, FinalDamage: 0}
	}

	// 2. Base Modifiers
	// Skill Modifier
	skillMod := 1.0 + (float64(weaponSkill) / 200.0)

	// Attribute Modifier
	var attrMod float64
	switch weapon.Type {
	case WeaponSlashing, WeaponBludgeoning:
		attrMod = 1.0 + (float64(attackerAttrs.Might) / 200.0)
	case WeaponPiercing:
		attrMod = 1.0 + ((float64(attackerAttrs.Might) + float64(attackerAttrs.Agility)) / 400.0)
	case WeaponRanged:
		attrMod = 1.0 + (float64(attackerAttrs.Agility) / 200.0)
	default:
		attrMod = 1.0
	}

	// Durability Modifier
	durabilityStatus := GetDurabilityStatus(weapon)
	if durabilityStatus.IsBroken {
		return DamageResult{FinalDamage: 0}
	}

	// 3. Raw Damage Calculation
	// raw = base * skillMod * attrMod * (roll / 100) * critMult * durabilityMod
	rollPercent := float64(roll) / 100.0
	rawFloat := float64(weapon.BaseDamage) * skillMod * attrMod * rollPercent * crit.Multiplier * durabilityStatus.DamageModifier
	rawDamage := int(rawFloat)

	// 4. Armor Reduction
	reduction := 0.0
	if targetArmor != nil && targetArmor.Durability > 0 {
		if effectivenessMap, ok := ArmorEffectiveness[weapon.Type]; ok {
			if baseReduction, ok := effectivenessMap[targetArmor.Type]; ok {
				// Scale by armor durability
				durabilityScale := GetArmorEffectivenessModifier(targetArmor)
				effectiveReduction := baseReduction * durabilityScale

				// Apply critical ignore armor
				// If ignoreArmor is 0.5, we multiply reduction by (1 - 0.5) = 0.5
				reduction = effectiveReduction * (1.0 - crit.IgnoreArmor)
			}
		}
	}

	finalDamage := int(float64(rawDamage) * (1.0 - reduction))
	blocked := rawDamage - finalDamage

	return DamageResult{
		FinalDamage: finalDamage,
		RawDamage:   rawDamage,
		IsCritical:  crit.IsCritical,
		IsFumble:    false,
		Blocked:     blocked,
	}
}
