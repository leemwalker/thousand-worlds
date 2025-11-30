package npc

import (
	"github.com/google/uuid"
)

// SpecializationTracker handles NPC proficiency in specific recipes
type SpecializationTracker struct {
	Proficiencies map[uuid.UUID]float64 // RecipeID -> Proficiency (0-100)
}

func NewSpecializationTracker() *SpecializationTracker {
	return &SpecializationTracker{
		Proficiencies: make(map[uuid.UUID]float64),
	}
}

// ImproveProficiency increases skill in a recipe
func (t *SpecializationTracker) ImproveProficiency(recipeID uuid.UUID, amount float64) {
	current := t.Proficiencies[recipeID]

	// Diminishing returns: harder to improve as you get better
	// Factor: (100 - current) / 100
	// At 0: factor 1.0
	// At 50: factor 0.5
	// At 90: factor 0.1
	factor := (100.0 - current) / 100.0
	actualGain := amount * factor

	// Minimum gain to ensure progress
	if actualGain < 0.1 {
		actualGain = 0.1
	}

	t.Proficiencies[recipeID] = current + actualGain
	if t.Proficiencies[recipeID] > 100.0 {
		t.Proficiencies[recipeID] = 100.0
	}
}

// GetProficiency returns the current proficiency for a recipe
func (t *SpecializationTracker) GetProficiency(recipeID uuid.UUID) float64 {
	return t.Proficiencies[recipeID]
}

// IsSpecialist returns true if proficiency is high enough
func (t *SpecializationTracker) IsSpecialist(recipeID uuid.UUID) bool {
	return t.Proficiencies[recipeID] >= 50.0
}

// GetCraftingBonuses returns speed and quality modifiers based on proficiency
func (t *SpecializationTracker) GetCraftingBonuses(recipeID uuid.UUID) (speedMod float64, qualityMod float64) {
	prof := t.Proficiencies[recipeID]

	// Speed: up to 50% faster (0.5 modifier means half time)
	// 0 prof -> 1.0
	// 100 prof -> 0.5
	speedMod = 1.0 - (prof / 200.0)

	// Quality: chance to improve quality tier
	// 0 prof -> 0.0
	// 100 prof -> 0.2 (20% bonus chance)
	qualityMod = prof / 500.0

	if t.IsSpecialist(recipeID) {
		speedMod -= 0.15   // Specialist bonus
		qualityMod += 0.10 // Specialist bonus
	}

	return speedMod, qualityMod
}
