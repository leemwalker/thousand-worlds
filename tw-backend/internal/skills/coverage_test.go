package skills

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCoverage_Checks(t *testing.T) {
	// Test PerformCheck (Randomized)
	// We can't assert exact values easily, but we can assert structure
	res := PerformCheck(50, 50, 50)
	assert.NotNil(t, res)
	assert.Equal(t, res.Total-res.Roll, 50+(50/5)) // Skill + Attr/5

	// Test PerformCheckWithSynergy
	resSyn := PerformCheckWithSynergy(50, 50, 10, 50)
	assert.NotNil(t, resSyn)
	assert.Equal(t, resSyn.Total-resSyn.Roll, 50+(50/5)+10)
}

func TestCoverage_QualityNames(t *testing.T) {
	assert.Equal(t, "Poor", GetQualityName(QualityPoor))
	assert.Equal(t, "Common", GetQualityName(QualityCommon))
	assert.Equal(t, "Fine", GetQualityName(QualityFine))
	assert.Equal(t, "Masterwork", GetQualityName(QualityMasterwork))
	assert.Equal(t, "Legendary", GetQualityName(QualityLegendary))
	assert.Equal(t, "Unknown", GetQualityName(999))
}

func TestCoverage_Progression_EdgeCases(t *testing.T) {
	// Test CalculateXPNeeded for Level 0
	assert.Equal(t, BaseXP, CalculateXPNeeded(0))

	// Test GainXP for non-existent skill
	pm := NewProgressionManager(NewSkillSheet(uuid.New())) // Valid UUID
	inc, lvl := pm.GainXP("NonExistentSkill", 100)
	assert.Nil(t, inc)
	assert.Nil(t, lvl)
}

func TestCoverage_Synergy_Missing(t *testing.T) {
	sheet := NewSkillSheet(uuid.New())
	// Skill with no synergy mapping
	assert.Equal(t, 0, GetSynergyBonus("UnknownSkill", sheet))
}
