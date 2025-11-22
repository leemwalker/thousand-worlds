package skills

// SynergyMap defines which skills provide bonuses to others
// Key: Skill Name, Value: List of related skills
var SynergyMap = map[string][]string{
	SkillSmithing:   {SkillMining},
	SkillMining:     {SkillSmithing},
	SkillAlchemy:    {SkillHerbalism},
	SkillHerbalism:  {SkillAlchemy},
	SkillHunting:    {SkillStealth, SkillPerception},
	SkillStealth:    {SkillHunting},
	SkillPerception: {SkillHunting},
}

// GetSynergyBonus calculates the bonus from related skills
// +5 bonus if related skill > 50
func GetSynergyBonus(skillName string, sheet *SkillSheet) int {
	relatedSkills, ok := SynergyMap[skillName]
	if !ok {
		return 0
	}

	bonus := 0
	for _, relatedName := range relatedSkills {
		if relatedSkill, exists := sheet.Skills[relatedName]; exists {
			if relatedSkill.Level > 50 {
				bonus += 5
			}
		}
	}
	return bonus
}
