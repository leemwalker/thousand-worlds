package behaviortree

import "tw-backend/internal/ecosystem/state"

// Status represents the result of a behavior node execution
type Status int

const (
	StatusSuccess Status = iota
	StatusFailure
	StatusRunning
)

// Node is the interface for all behavior tree nodes
type Node interface {
	Tick(entity *state.LivingEntityState) Status
}

// Sequence runs children until one fails
type Sequence struct {
	Children []Node
}

func (s *Sequence) Tick(e *state.LivingEntityState) Status {
	for _, child := range s.Children {
		status := child.Tick(e)
		if status != StatusSuccess {
			return status
		}
	}
	return StatusSuccess
}

// Selector runs children until one succeeds
type Selector struct {
	Children []Node
}

func (s *Selector) Tick(e *state.LivingEntityState) Status {
	for _, child := range s.Children {
		status := child.Tick(e)
		if status != StatusFailure {
			return status
		}
	}
	return StatusFailure
}

// ActionNode executes a specific game action
type ActionNode struct {
	Action func(e *state.LivingEntityState) Status
}

func (a *ActionNode) Tick(e *state.LivingEntityState) Status {
	return a.Action(e)
}

// ConditionNode checks a predicate
type ConditionNode struct {
	Predicate func(e *state.LivingEntityState) bool
}

func (c *ConditionNode) Tick(e *state.LivingEntityState) Status {
	if c.Predicate(e) {
		return StatusSuccess
	}
	return StatusFailure
}
