package desire

import (
	"github.com/google/uuid"
)

// GetBestAction determines the action to fulfill a specific need
func GetBestAction(need *Need, npcID uuid.UUID) Action {
	action := Action{
		Source:   need.Name,
		Priority: need.Value, // Base priority, usually recalculated
	}

	switch need.Name {
	// Survival
	case NeedHunger:
		if need.Value >= 70 {
			action.Name = "Seek Food"
			action.Type = "seek_food"
		} else {
			action.Name = "Eat Snack"
			action.Type = "eat"
		}
	case NeedThirst:
		if need.Value >= 60 {
			action.Name = "Seek Water"
			action.Type = "seek_water"
		} else {
			action.Name = "Drink"
			action.Type = "drink"
		}
	case NeedSleep:
		if need.Value >= 75 {
			action.Name = "Seek Bed"
			action.Type = "seek_bed"
		} else {
			action.Name = "Rest"
			action.Type = "rest"
		}
	case NeedSafety:
		if need.Value >= 50 {
			action.Name = "Flee"
			action.Type = "flee"
		} else {
			action.Name = "Seek Shelter"
			action.Type = "seek_shelter"
		}

	// Social
	case NeedCompanionship:
		action.Name = "Seek Company"
		action.Type = "seek_company"
	case NeedConversation:
		action.Name = "Initiate Dialogue"
		action.Type = "talk"
	case NeedAffection:
		action.Name = "Give Gift"
		action.Type = "give_gift"

	// Achievement
	case NeedTaskCompletion:
		action.Name = "Work on Task"
		action.Type = "work"
	case NeedSkillImprovement:
		action.Name = "Practice Skill"
		action.Type = "practice"
	case NeedResourceAcquisition:
		action.Name = "Gather Resources"
		action.Type = "gather"

	// Pleasure
	case NeedCuriosity:
		action.Name = "Explore"
		action.Type = "explore"
	case NeedHedonism:
		action.Name = "Seek Entertainment"
		action.Type = "seek_fun"
	case NeedCreativity:
		action.Name = "Create Art"
		action.Type = "create"
	}

	return action
}
