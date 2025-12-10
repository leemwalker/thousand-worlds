package skills

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles business logic for skills
type Service struct {
	repo Repository
}

// NewService creates a new skills service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetSkillSheet retrieves the full skill sheet for a character
// It initializes a default sheet and overlays persistent data (XP)
func (s *Service) GetSkillSheet(ctx context.Context, characterID uuid.UUID) (*SkillSheet, error) {
	// 1. Create default sheet with all skills initialized to 0
	sheet := NewSkillSheet(characterID)

	// 2. Fetch stored skills (XP) from database
	storedSkills, err := s.repo.GetSkills(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored skills: %w", err)
	}

	// 3. Update sheet with stored values
	for _, storedSkill := range storedSkills {
		// Find the skill in the initialized sheet (to keep Category, etc.)
		if skill, exists := sheet.Skills[storedSkill.Name]; exists {
			// Update XP
			skill.XP = storedSkill.XP

			// Recalculate Level based on XP
			// We effectively "add" the XP to a 0-level skill to let the progression logic determine the level
			// Reset level to 0 first just to be safe (though it should be 0 from NewSkillSheet)
			skill.Level = 0
			// Calculate level from XP
			// We can use the existing AddXP logic, but passing 0 as 'amount' won't work if we want to set it absolute.
			// Ideally we should have a 'CalculateLevelFromXP' function.
			// Because AddXP is incremental, we need to simulate re-leveling.
			// Let's use internal logic of AddXP but applied to total XP.

			// Reset XP to 0 for calculation purposes? No, AddXP adds to existing.
			// Let's manually simulate level up loop:
			currentXP := skill.XP
			skill.XP = 0           // Reset for calculation
			skill.AddXP(currentXP) // Re-apply all XP to calculate level

			sheet.Skills[storedSkill.Name] = skill
		}
	}

	return sheet, nil
}

// GainXP adds XP to a skill and persists it
func (s *Service) GainXP(ctx context.Context, characterID uuid.UUID, skillName string, amount float64) (*SkillIncreasedEvent, *SkillLeveledUpEvent, error) {
	sheet, err := s.GetSkillSheet(ctx, characterID)
	if err != nil {
		return nil, nil, err
	}

	pm := NewProgressionManager(sheet)
	inc, lvl := pm.GainXP(skillName, amount)

	if inc != nil {
		// Persist the new total XP
		newTotalXP := sheet.Skills[skillName].XP
		if err := s.repo.UpdateSkill(ctx, characterID, skillName, newTotalXP); err != nil {
			return nil, nil, fmt.Errorf("failed to persist skill update: %w", err)
		}
	}

	return inc, lvl, nil
}
