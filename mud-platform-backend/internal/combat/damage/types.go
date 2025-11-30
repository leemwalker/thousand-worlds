package damage

import "github.com/google/uuid"

// WeaponType represents the category of a weapon
type WeaponType string

const (
	WeaponSlashing    WeaponType = "slashing"
	WeaponPiercing    WeaponType = "piercing"
	WeaponBludgeoning WeaponType = "bludgeoning"
	WeaponRanged      WeaponType = "ranged"
)

// ArmorType represents the category of armor
type ArmorType string

const (
	ArmorLeather   ArmorType = "leather"
	ArmorChainMail ArmorType = "chain_mail"
	ArmorPlate     ArmorType = "plate"
)

// Weapon represents a combat weapon
type Weapon struct {
	WeaponID      uuid.UUID
	Name          string
	Type          WeaponType
	BaseDamage    int
	Durability    int
	MaxDurability int
	SkillRequired int // Minimum skill to use effectively
}

// Armor represents defensive gear
type Armor struct {
	ArmorID       uuid.UUID
	Name          string
	Type          ArmorType
	Durability    int
	MaxDurability int
}
