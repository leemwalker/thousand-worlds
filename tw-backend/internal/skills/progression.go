package skills

import (
	"math"
	"time"
)

const BaseXP = 100.0

// CalculateXPNeeded returns the total XP needed to reach the next level
// Formula: BaseXP * (currentLevel^1.5)
func CalculateXPNeeded(currentLevel int) float64 {
	if currentLevel == 0 {
		return BaseXP
	}
	return BaseXP * math.Pow(float64(currentLevel), 1.5)
}

// AddXP adds XP to a skill and handles leveling up
// Returns true if leveled up, and the new level
func (s *Skill) AddXP(amount float64) (bool, int) {
	s.XP += amount
	leveledUp := false

	// Check for level up
	// We loop in case multiple levels are gained at once (unlikely but possible with big XP drops)
	for {
		// Cost to reach Level N = Base * (N^1.5).

		cost := CalculateXPNeeded(s.Level + 1)

		if s.XP >= cost {
			s.XP -= cost
			s.Level++
			leveledUp = true
		} else {
			break
		}

		if s.Level >= 100 {
			s.Level = 100
			s.XP = 0 // Cap at 100
			break
		}
	}

	return leveledUp, s.Level
}

// ProgressionManager handles XP gain for a character
type ProgressionManager struct {
	Sheet *SkillSheet
}

func NewProgressionManager(sheet *SkillSheet) *ProgressionManager {
	return &ProgressionManager{Sheet: sheet}
}

func (pm *ProgressionManager) GainXP(skillName string, amount float64) (*SkillIncreasedEvent, *SkillLeveledUpEvent) {
	skill, exists := pm.Sheet.Skills[skillName]
	if !exists {
		return nil, nil
	}

	oldLevel := skill.Level
	leveledUp, newLevel := skill.AddXP(amount)
	pm.Sheet.Skills[skillName] = skill // Update map

	incEvent := &SkillIncreasedEvent{
		CharacterID: pm.Sheet.CharacterID,
		SkillName:   skillName,
		OldValue:    oldLevel,
		NewValue:    newLevel,
		XPGained:    amount,
		Timestamp:   time.Now(),
	}

	var lvlEvent *SkillLeveledUpEvent
	if leveledUp {
		lvlEvent = &SkillLeveledUpEvent{
			CharacterID: pm.Sheet.CharacterID,
			SkillName:   skillName,
			NewLevel:    newLevel,
			Timestamp:   time.Now(),
		}
	}

	return incEvent, lvlEvent
}
