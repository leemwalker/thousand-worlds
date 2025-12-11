package behaviortree

import (
	"tw-backend/internal/ecosystem/state"
)

// Standard Behavior Tree Library

func NewHerbivoreTree() Node {
	// Priority:
	// 1. Flee Danger (if logic exists)
	// 2. Eat if Hungry
	// 3. Sleep if Tired
	// 4. Wander (Default)

	return &Selector{
		Children: []Node{
			// Critical Needs
			&Sequence{
				Children: []Node{
					&ConditionNode{Predicate: func(e *state.LivingEntityState) bool {
						return e.Needs.Hunger > state.ThresholdHungerCritical
					}},
					&ActionNode{Action: ActionFindFood},
				},
			},
			&Sequence{
				Children: []Node{
					&ConditionNode{Predicate: func(e *state.LivingEntityState) bool {
						return e.Needs.Energy < state.ThresholdEnergyCritical
					}},
					&ActionNode{Action: ActionSleep},
				},
			},
			// Default
			&ActionNode{Action: ActionWander},
		},
	}
}

// Placeholder Actions
func ActionFindFood(e *state.LivingEntityState) Status {
	// Logic to find food would call pathfinding
	// For now just simulation
	e.Needs.Hunger -= 10
	if e.Needs.Hunger < 0 {
		e.Needs.Hunger = 0
	}
	e.AddLog("Eat", "Found food sources")
	return StatusSuccess
}

func ActionSleep(e *state.LivingEntityState) Status {
	e.Needs.Energy += 10
	if e.Needs.Energy > 100 {
		e.Needs.Energy = 100
	}
	e.AddLog("Sleep", "Resting to recover energy")
	return StatusSuccess
}

func ActionWander(e *state.LivingEntityState) Status {
	e.Needs.Energy -= 1
	// Log wandering less frequently? Or just log it.
	// For debugging, logging every tick might be spammy, but useful to see "Wandering..."
	// Let's only log if we haven't recently logged "Wander" to avoid spam
	if len(e.Logs) == 0 || e.Logs[len(e.Logs)-1].Action != "Wander" {
		e.AddLog("Wander", "Exploring environment")
	}
	return StatusSuccess
}

// NewFloraTree creates a behavior tree for plants (no movement, just growth)
func NewFloraTree() Node {
	return &Selector{
		Children: []Node{
			// Photosynthesize (always runs, increases energy/reduces hunger)
			&ActionNode{Action: ActionPhotosynthesize},
		},
	}
}

// ActionPhotosynthesize - plants gain energy from sunlight
func ActionPhotosynthesize(e *state.LivingEntityState) Status {
	// Plants slowly regenerate through photosynthesis
	// Reduce hunger (plants don't really "eat" but this represents nutrient absorption)
	e.Needs.Hunger -= 0.5
	if e.Needs.Hunger < 0 {
		e.Needs.Hunger = 0
	}

	// Gain energy
	e.Needs.Energy += 0.2
	if e.Needs.Energy > 100 {
		e.Needs.Energy = 100
	}

	// Reduce thirst slightly (water absorption)
	e.Needs.Thirst -= 0.3
	if e.Needs.Thirst < 0 {
		e.Needs.Thirst = 0
	}

	// Only log occasionally
	if len(e.Logs) == 0 || e.Logs[len(e.Logs)-1].Action != "Photosynthesize" {
		e.AddLog("Photosynthesize", "Absorbing sunlight")
	}

	return StatusSuccess
}
