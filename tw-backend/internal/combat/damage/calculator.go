package damage

import (
	"tw-backend/internal/character"
	"tw-backend/internal/combat/config"
)

// defaultConfig is the package-level combat configuration.
// Use SetConfig to change it for testing or runtime configuration.
var defaultConfig = config.Default()

// SetConfig sets the package-level combat configuration.
// This enables runtime configuration changes via SIGHUP or admin commands.
func SetConfig(cfg *config.CombatConfig) {
	defaultConfig = cfg
}

// GetConfig returns the current package-level combat configuration.
func GetConfig() *config.CombatConfig {
	return defaultConfig
}

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
	// Skill Modifier (uses config divisor)
	skillMod := 1.0 + (float64(weaponSkill) / defaultConfig.GetSkillDivisor())

	// Attribute Modifier (uses config divisors)
	var attrMod float64
	switch weapon.Type {
	case WeaponSlashing, WeaponBludgeoning:
		attrMod = 1.0 + (float64(attackerAttrs.Might) / defaultConfig.GetMightDivisor())
	case WeaponPiercing:
		attrMod = 1.0 + ((float64(attackerAttrs.Might) + float64(attackerAttrs.Agility)) / defaultConfig.GetMixedAttributeDivisor())
	case WeaponRanged:
		attrMod = 1.0 + (float64(attackerAttrs.Agility) / defaultConfig.GetAgilityDivisor())
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
	rollPercent := float64(roll) / defaultConfig.GetRollDivisor()
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
