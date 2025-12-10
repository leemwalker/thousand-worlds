package skills

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetSynergyBonus(t *testing.T) {
	sheet := NewSkillSheet(uuid.New())

	// No bonus initially
	assert.Equal(t, 0, GetSynergyBonus(SkillSmithing, sheet))

	// Level up Mining to 51
	mining := sheet.Skills[SkillMining]
	mining.Level = 51
	sheet.Skills[SkillMining] = mining

	// Check Smithing bonus (Mining > 50)
	assert.Equal(t, 5, GetSynergyBonus(SkillSmithing, sheet))

	// Check Mining bonus (Smithing is 0)
	assert.Equal(t, 0, GetSynergyBonus(SkillMining, sheet))
}

func TestApplySoftCap(t *testing.T) {
	// Attribute 50 -> Soft Cap 75

	// Level 70 (Under Cap) -> Full XP
	assert.Equal(t, 100.0, ApplySoftCap(100.0, 70, 50))

	// Level 80 (Over Cap) -> Half XP
	assert.Equal(t, 50.0, ApplySoftCap(100.0, 80, 50))
}
