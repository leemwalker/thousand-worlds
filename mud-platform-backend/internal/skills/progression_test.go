package skills

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateXPNeeded(t *testing.T) {
	// Level 1 cost (0->1)
	// Our formula uses target level.
	// CalculateXPNeeded(1) = 100 * 1^1.5 = 100
	assert.Equal(t, 100.0, CalculateXPNeeded(1))

	// Level 2 cost (1->2)
	// 100 * 2^1.5 = 100 * 2.828 = 282.8
	assert.InDelta(t, 282.84, CalculateXPNeeded(2), 0.01)

	// Level 10 cost
	// 100 * 10^1.5 = 100 * 31.62 = 3162
	assert.InDelta(t, 3162.27, CalculateXPNeeded(10), 0.01)
}

func TestSkill_AddXP(t *testing.T) {
	skill := Skill{Name: "Test", Level: 0, XP: 0}

	// Add 50 XP (Not enough for L1)
	leveled, newLvl := skill.AddXP(50)
	assert.False(t, leveled)
	assert.Equal(t, 0, newLvl)
	assert.Equal(t, 50.0, skill.XP)

	// Add 60 XP (Total 110, enough for L1 cost 100)
	leveled, newLvl = skill.AddXP(60)
	assert.True(t, leveled)
	assert.Equal(t, 1, newLvl)
	assert.Equal(t, 10.0, skill.XP) // 110 - 100 = 10 remainder
}

func TestProgressionManager_GainXP(t *testing.T) {
	sheet := NewSkillSheet(uuid.New())
	pm := NewProgressionManager(sheet)

	// Gain XP
	incEvent, lvlEvent := pm.GainXP(SkillSlashing, 150)

	assert.NotNil(t, incEvent)
	assert.Equal(t, SkillSlashing, incEvent.SkillName)
	assert.Equal(t, 150.0, incEvent.XPGained)

	// Should level up from 0 to 1 (Cost 100)
	assert.NotNil(t, lvlEvent)
	assert.Equal(t, 1, lvlEvent.NewLevel)

	// Check sheet update
	assert.Equal(t, 1, sheet.Skills[SkillSlashing].Level)
}
