package crafting

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// DiscoverRecipe unlocks a recipe for an entity
func (m *RecipeManager) DiscoverRecipe(entityID uuid.UUID, recipeID uuid.UUID, source string) error {
	knowledge := &RecipeKnowledge{
		EntityID:     entityID,
		RecipeID:     recipeID,
		Proficiency:  0.0,
		TimesUsed:    0,
		DiscoveredAt: time.Now(),
		Source:       source,
	}

	return m.repo.DiscoverRecipe(knowledge)
}

// ExperimentWithIngredients attempts to discover a recipe by combining ingredients
func (m *RecipeManager) ExperimentWithIngredients(
	entityID uuid.UUID,
	ingredients []Ingredient,
	station *CraftingStation,
	crafterSkill int,
) (*Recipe, error) {
	// 1. Calculate compatibility (simplified logic)
	// In a real system, this would check resource tags/properties
	compatibility := 0.5 // Base chance

	// 2. Calculate success chance
	// Skill 0-100 maps to 0.0-1.0 modifier
	skillMod := float64(crafterSkill) / 100.0
	successChance := compatibility * skillMod * 0.3 // Max 30% chance

	if rand.Float64() > successChance {
		return nil, nil // Failed
	}

	// 3. Find matching recipe
	// In reality, we'd look up if these ingredients match a hidden recipe
	// For now, we'll return a mock "discovered" recipe if successful
	// This requires querying the repo for recipes matching ingredients
	// We'll leave this as a placeholder for now

	return nil, nil
}

// TeachRecipe attempts to transfer recipe knowledge from teacher to student
func (m *RecipeManager) TeachRecipe(teacherID, studentID, recipeID uuid.UUID, relationshipAffection int) error {
	// 1. Validate teacher knows recipe
	teacherKnowledge, err := m.repo.GetRecipeKnowledge(teacherID, recipeID)
	if err != nil {
		return err
	}
	if teacherKnowledge == nil {
		return fmt.Errorf("teacher does not know recipe")
	}

	// 2. Check relationship
	if relationshipAffection < 40 {
		return fmt.Errorf("relationship too low")
	}

	// 3. Success check based on proficiency
	successChance := 0.7 + (teacherKnowledge.Proficiency / 200.0) // 70-100%

	if rand.Float64() > successChance {
		return fmt.Errorf("teaching failed")
	}

	// 4. Unlock for student
	return m.DiscoverRecipe(studentID, recipeID, "taught")
}
