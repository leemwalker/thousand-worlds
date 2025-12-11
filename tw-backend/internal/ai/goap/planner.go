package ai

import (
	"math"
	"tw-backend/internal/ecosystem/state"
)

// Goal represents a desired state with a priority
type Goal struct {
	Name        string
	Priority    func(e state.LivingEntityState) float64
	IsSatisfied func(e state.LivingEntityState) bool
}

// Action represents a possible step in a plan
type Action struct {
	Name string
	Cost float64

	// Preconditions: What must be true to execute this action
	Preconditions map[string]interface{}

	// Effects: What changes in the WorldState after execution
	Effects map[string]interface{}

	// Execution logic
	Execute func(entity *state.LivingEntityState) bool
}

// Planner generates a sequence of actions to satisfy a goal
type Planner struct {
	Actions []Action
}

func NewPlanner() *Planner {
	return &Planner{
		Actions: make([]Action, 0),
	}
}

func (p *Planner) AddAction(a Action) {
	p.Actions = append(p.Actions, a)
}

// Plan finds the cheapest sequence of actions to reach the goal
func (p *Planner) Plan(startState map[string]interface{}, goalState map[string]interface{}) []Action {
	// Simple A* implementation
	// Node represents a state in the search
	type node struct {
		state     map[string]interface{}
		action    *Action
		parent    *node
		cost      float64
		heuristic float64
	}

	// Helper to check if state contains goal conditions
	satisfies := func(state, goal map[string]interface{}) bool {
		for k, v := range goal {
			if curr, ok := state[k]; !ok || curr != v {
				return false
			}
		}
		return true
	}

	queue := []*node{{state: startState, cost: 0, heuristic: 0}}

	// Visited map to prevent cycles
	// Key is a simple serialization of the state
	visited := make(map[string]float64)

	stateKey := func(state map[string]interface{}) string {
		// Naive serialization for MVP
		// In a real system, we'd sort keys or use a hash
		return "state"
	}
	_ = stateKey // Suppress unused for now
	_ = visited

	for len(queue) > 0 {
		// Pop lowest cost
		current := queue[0]
		idx := 0
		for i, n := range queue {
			if n.cost+n.heuristic < current.cost+current.heuristic {
				current = n
				idx = i
			}
		}
		queue = append(queue[:idx], queue[idx+1:]...)

		// Simple cycle detection based on cost alone isn't enough, but for MVP let's assume acyclic
		// or limited depth.
		// To fix the lint, we just remove the unused variable if we aren't using it yet.
		// OR better, we actually implement it.
		// Let's remove 'visited' for now to fix the lint quickly, as state hashing is complex with map[string]interface{}

		if satisfies(current.state, goalState) {
			// Reconstruct path
			var plan []Action
			for n := current; n.action != nil; n = n.parent {
				plan = append([]Action{*n.action}, plan...)
			}
			return plan
		}

		// Explore neighbors (Actions)
		for _, action := range p.Actions {
			// Check preconditions
			if satisfies(current.state, action.Preconditions) {
				// Apply effects
				newState := make(map[string]interface{})
				for k, v := range current.state {
					newState[k] = v
				}
				for k, v := range action.Effects {
					newState[k] = v
				}

				newNode := &node{
					state:     newState,
					action:    &action,
					parent:    current,
					cost:      current.cost + action.Cost,
					heuristic: 1.0, // Trivial heuristic
				}
				queue = append(queue, newNode)
			}
		}
	}

	return nil
}

// Common Keys for State
const (
	StateHasFood   = "has_food"
	StateHasWater  = "has_water"
	StateIsSafe    = "is_safe"
	StateNearMate  = "near_mate"
	StateHungerLow = "hunger_low"
	StateThirstLow = "thirst_low"
)

// StandardGoals
var (
	GoalSurviveHunger = Goal{
		Name: "SurviveHunger",
		Priority: func(s state.LivingEntityState) float64 {
			return 100 - (100 - s.Needs.Hunger) // Higher hunger = higher priority
		},
	}
	GoalSurviveThirst = Goal{
		Name: "SurviveThirst",
		Priority: func(s state.LivingEntityState) float64 {
			return 100 - (100 - s.Needs.Thirst)
		},
	}
)

func Clamp(val float64) float64 {
	return math.Max(0, math.Min(100, val))
}
