package skills

import (
	"github.com/google/uuid"
)

// Skill Categories
const (
	CategoryCombat    = "Combat"
	CategoryCrafting  = "Crafting"
	CategoryGathering = "Gathering"
	CategoryUtility   = "Utility"
	CategorySocial    = "Social"
)

// Skill Names
const (
	// Combat
	SkillSlashing    = "Slashing"
	SkillPiercing    = "Piercing"
	SkillBludgeoning = "Bludgeoning"
	SkillDefense     = "Defense"
	SkillDodge       = "Dodge"

	// Crafting
	SkillSmithing  = "Smithing"
	SkillAlchemy   = "Alchemy"
	SkillCarpentry = "Carpentry"
	SkillTailoring = "Tailoring"
	SkillCooking   = "Cooking"

	// Gathering
	SkillMining    = "Mining"
	SkillHerbalism = "Herbalism"
	SkillLogging   = "Logging"
	SkillHunting   = "Hunting"
	SkillFishing   = "Fishing"

	// Utility
	SkillPerception = "Perception"
	SkillStealth    = "Stealth"
	SkillClimbing   = "Climbing"
	SkillSwimming   = "Swimming"
	SkillNavigation = "Navigation"

	// Social
	SkillPersuasion   = "Persuasion"
	SkillIntimidation = "Intimidation"
	SkillDeception    = "Deception"
	SkillBartering    = "Bartering"
)

// Difficulty Thresholds
const (
	DifficultyEasy     = 30
	DifficultyMedium   = 50
	DifficultyHard     = 70
	DifficultyVeryHard = 90
)

// Skill represents a single skill for a character
type Skill struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Level    int     `json:"level"` // 0-100
	XP       float64 `json:"xp"`
}

// SkillSheet holds all skills for a character
type SkillSheet struct {
	CharacterID uuid.UUID        `json:"character_id"`
	Skills      map[string]Skill `json:"skills"`
}

// NewSkillSheet creates a new skill sheet with all skills initialized to 0
func NewSkillSheet(charID uuid.UUID) *SkillSheet {
	sheet := &SkillSheet{
		CharacterID: charID,
		Skills:      make(map[string]Skill),
	}

	// Initialize all skills
	initSkill(sheet, SkillSlashing, CategoryCombat)
	initSkill(sheet, SkillPiercing, CategoryCombat)
	initSkill(sheet, SkillBludgeoning, CategoryCombat)
	initSkill(sheet, SkillDefense, CategoryCombat)
	initSkill(sheet, SkillDodge, CategoryCombat)

	initSkill(sheet, SkillSmithing, CategoryCrafting)
	initSkill(sheet, SkillAlchemy, CategoryCrafting)
	initSkill(sheet, SkillCarpentry, CategoryCrafting)
	initSkill(sheet, SkillTailoring, CategoryCrafting)
	initSkill(sheet, SkillCooking, CategoryCrafting)

	initSkill(sheet, SkillMining, CategoryGathering)
	initSkill(sheet, SkillHerbalism, CategoryGathering)
	initSkill(sheet, SkillLogging, CategoryGathering)
	initSkill(sheet, SkillHunting, CategoryGathering)
	initSkill(sheet, SkillFishing, CategoryGathering)

	initSkill(sheet, SkillPerception, CategoryUtility)
	initSkill(sheet, SkillStealth, CategoryUtility)
	initSkill(sheet, SkillClimbing, CategoryUtility)
	initSkill(sheet, SkillSwimming, CategoryUtility)
	initSkill(sheet, SkillNavigation, CategoryUtility)

	initSkill(sheet, SkillPersuasion, CategorySocial)
	initSkill(sheet, SkillIntimidation, CategorySocial)
	initSkill(sheet, SkillDeception, CategorySocial)
	initSkill(sheet, SkillBartering, CategorySocial)

	return sheet
}

func initSkill(sheet *SkillSheet, name, category string) {
	sheet.Skills[name] = Skill{
		Name:     name,
		Category: category,
		Level:    0,
		XP:       0,
	}
}
