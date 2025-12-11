package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanner_Plan(t *testing.T) {
	planner := NewPlanner()

	// Setup Actions for a simple survival scenario
	// Goal: Have Fire
	// Chain: Get Axe -> Chop Tree (Have Wood) -> Build Fire (Have Fire)

	planner.AddAction(Action{
		Name:          "GetAxe",
		Cost:          1,
		Preconditions: map[string]interface{}{},
		Effects:       map[string]interface{}{"has_axe": true},
	})

	planner.AddAction(Action{
		Name:          "ChopTree",
		Cost:          2,
		Preconditions: map[string]interface{}{"has_axe": true},
		Effects:       map[string]interface{}{"has_wood": true},
	})

	planner.AddAction(Action{
		Name:          "BuildFire",
		Cost:          1,
		Preconditions: map[string]interface{}{"has_wood": true},
		Effects:       map[string]interface{}{"has_fire": true},
	})

	// 1. Valid Plan
	startState := map[string]interface{}{
		"has_axe":  false,
		"has_wood": false,
		"has_fire": false,
	}
	goalState := map[string]interface{}{
		"has_fire": true,
	}

	plan := planner.Plan(startState, goalState)
	assert.NotNil(t, plan)
	assert.Len(t, plan, 3)
	assert.Equal(t, "GetAxe", plan[0].Name)
	assert.Equal(t, "ChopTree", plan[1].Name)
	assert.Equal(t, "BuildFire", plan[2].Name)

	// 2. Already Satisfied
	startStateSatisfied := map[string]interface{}{"has_fire": true}
	planSatisfied := planner.Plan(startStateSatisfied, goalState)
	assert.NotNil(t, planSatisfied)
	assert.Len(t, planSatisfied, 0, "Should return empty plan (or start directly met)")

	// 3. Impossible Plan
	goalImpossible := map[string]interface{}{"has_gold": true}
	planImpossible := planner.Plan(startState, goalImpossible)
	assert.Nil(t, planImpossible)
}
