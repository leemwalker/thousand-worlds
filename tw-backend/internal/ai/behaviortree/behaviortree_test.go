package behaviortree

import (
	"testing"
	"tw-backend/internal/ecosystem/state"

	"github.com/stretchr/testify/assert"
)

// MockAction creates an action node that returns a specific status
func MockAction(s Status) *ActionNode {
	return &ActionNode{
		Action: func(e *state.LivingEntityState) Status {
			return s
		},
	}
}

func TestSequence(t *testing.T) {
	e := &state.LivingEntityState{}

	t.Run("All Success", func(t *testing.T) {
		seq := &Sequence{
			Children: []Node{
				MockAction(StatusSuccess),
				MockAction(StatusSuccess),
			},
		}
		assert.Equal(t, StatusSuccess, seq.Tick(e))
	})

	t.Run("First Fail", func(t *testing.T) {
		seq := &Sequence{
			Children: []Node{
				MockAction(StatusFailure),
				MockAction(StatusSuccess),
			},
		}
		assert.Equal(t, StatusFailure, seq.Tick(e))
	})

	t.Run("Running", func(t *testing.T) {
		seq := &Sequence{
			Children: []Node{
				MockAction(StatusSuccess),
				MockAction(StatusRunning),
				MockAction(StatusSuccess),
			},
		}
		assert.Equal(t, StatusRunning, seq.Tick(e))
	})
}

func TestSelector(t *testing.T) {
	e := &state.LivingEntityState{}

	t.Run("First Success", func(t *testing.T) {
		sel := &Selector{
			Children: []Node{
				MockAction(StatusSuccess),
				MockAction(StatusFailure),
			},
		}
		assert.Equal(t, StatusSuccess, sel.Tick(e))
	})

	t.Run("All Fail", func(t *testing.T) {
		sel := &Selector{
			Children: []Node{
				MockAction(StatusFailure),
				MockAction(StatusFailure),
			},
		}
		assert.Equal(t, StatusFailure, sel.Tick(e))
	})

	t.Run("Running", func(t *testing.T) {
		sel := &Selector{
			Children: []Node{
				MockAction(StatusFailure),
				MockAction(StatusRunning),
				MockAction(StatusSuccess),
			},
		}
		assert.Equal(t, StatusRunning, sel.Tick(e))
	})
}

func TestCondition(t *testing.T) {
	e := &state.LivingEntityState{}

	trueNode := &ConditionNode{Predicate: func(e *state.LivingEntityState) bool { return true }}
	falseNode := &ConditionNode{Predicate: func(e *state.LivingEntityState) bool { return false }}

	assert.Equal(t, StatusSuccess, trueNode.Tick(e))
	assert.Equal(t, StatusFailure, falseNode.Tick(e))
}

// -- Library Verification Tests --

func TestHerbivoreTree_Hungry(t *testing.T) {
	e := &state.LivingEntityState{
		Needs: state.NeedState{
			Hunger: state.ThresholdHungerCritical + 10,
			Energy: 100,
		},
	}

	tree := NewHerbivoreTree()
	status := tree.Tick(e)

	assert.Equal(t, StatusSuccess, status)
	assert.Contains(t, e.Logs[len(e.Logs)-1].Action, "Eat")
	assert.Less(t, e.Needs.Hunger, float64(state.ThresholdHungerCritical+10))
}

func TestHerbivoreTree_Tired(t *testing.T) {
	e := &state.LivingEntityState{
		Needs: state.NeedState{
			Hunger: 0,
			Energy: state.ThresholdEnergyCritical - 10,
		},
	}

	tree := NewHerbivoreTree()
	status := tree.Tick(e)

	assert.Equal(t, StatusSuccess, status)
	assert.Contains(t, e.Logs[len(e.Logs)-1].Action, "Sleep")
	assert.Greater(t, e.Needs.Energy, float64(state.ThresholdEnergyCritical-10))
}

func TestHerbivoreTree_Wander(t *testing.T) {
	e := &state.LivingEntityState{
		Needs: state.NeedState{
			Hunger: 0,
			Energy: 100,
		},
	}

	tree := NewHerbivoreTree()
	status := tree.Tick(e)

	assert.Equal(t, StatusSuccess, status)
	assert.Contains(t, e.Logs[len(e.Logs)-1].Action, "Wander")
}

func TestFloraTree(t *testing.T) {
	e := &state.LivingEntityState{
		Needs: state.NeedState{Hunger: 50, Energy: 50},
	}

	tree := NewFloraTree()
	status := tree.Tick(e)

	assert.Equal(t, StatusSuccess, status)
	assert.Contains(t, e.Logs[len(e.Logs)-1].Action, "Photosynthesize")
	assert.Less(t, e.Needs.Hunger, float64(50))
	assert.Greater(t, e.Needs.Energy, float64(50))
}
