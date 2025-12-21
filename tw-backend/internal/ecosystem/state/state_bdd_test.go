package state_test

import (
	"testing"

	"tw-backend/internal/ecosystem/state"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// BDD Tests: Need System
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Need System Creation
// -----------------------------------------------------------------------------
// Given: Default NeedSystem
// When: Created
// Then: Should be usable
func TestBDD_NeedSystem_Creation(t *testing.T) {
	ns := state.NeedSystem{}
	assert.NotNil(t, &ns, "NeedSystem should be creatable")
}

// -----------------------------------------------------------------------------
// Scenario: Need Constants
// -----------------------------------------------------------------------------
// Given: Need system constants
// When: Examined
// Then: Should have reasonable values
func TestBDD_NeedSystem_Constants(t *testing.T) {
	assert.Equal(t, 100.0, state.MaxNeedValue, "Max need should be 100")
	assert.Equal(t, 0.0, state.MinNeedValue, "Min need should be 0")
	assert.Greater(t, state.ThresholdHungerCritical, 0.0, "Hunger threshold should be positive")
}

// -----------------------------------------------------------------------------
// Scenario: Living Entity State Tick
// -----------------------------------------------------------------------------
// Given: A living entity with initial needs
// When: NeedSystem.Tick is called
// Then: Needs should change appropriately
func TestBDD_NeedSystem_Tick(t *testing.T) {
	ns := state.NeedSystem{}
	entityState := &state.LivingEntityState{
		Needs: state.NeedState{
			Hunger:           0.0,
			Thirst:           0.0,
			Energy:           100.0,
			ReproductionUrge: 0.0,
		},
	}

	// Tick once
	ns.Tick(entityState, nil)

	// Hunger and thirst should increase
	assert.Greater(t, entityState.Needs.Hunger, 0.0, "Hunger should increase after tick")
	assert.Greater(t, entityState.Needs.Thirst, 0.0, "Thirst should increase after tick")
	// Energy should decrease
	assert.Less(t, entityState.Needs.Energy, 100.0, "Energy should decrease after tick")
}

// -----------------------------------------------------------------------------
// Scenario: Health Check
// -----------------------------------------------------------------------------
// Given: An entity with non-critical needs
// When: IsHealthy is called
// Then: Should return true
func TestBDD_NeedSystem_IsHealthy(t *testing.T) {
	ns := state.NeedSystem{}
	healthyState := &state.LivingEntityState{
		Needs: state.NeedState{
			Hunger: 20.0, // Well below critical (80)
			Thirst: 20.0, // Well below critical (85)
			Energy: 80.0, // Well above critical (10)
		},
	}

	assert.True(t, ns.IsHealthy(healthyState), "Entity with good needs should be healthy")
}

// -----------------------------------------------------------------------------
// Scenario: Unhealthy Due to Hunger
// -----------------------------------------------------------------------------
// Given: An entity with critical hunger
// When: IsHealthy is called
// Then: Should return false
func TestBDD_NeedSystem_UnhealthyHunger(t *testing.T) {
	ns := state.NeedSystem{}
	hungryState := &state.LivingEntityState{
		Needs: state.NeedState{
			Hunger: 90.0, // Above critical (80)
			Thirst: 20.0,
			Energy: 80.0,
		},
	}

	assert.False(t, ns.IsHealthy(hungryState), "Entity with critical hunger should not be healthy")
}

// -----------------------------------------------------------------------------
// Scenario: Unhealthy Due to Low Energy
// -----------------------------------------------------------------------------
// Given: An entity with critical energy
// When: IsHealthy is called
// Then: Should return false
func TestBDD_NeedSystem_UnhealthyEnergy(t *testing.T) {
	ns := state.NeedSystem{}
	exhaustedState := &state.LivingEntityState{
		Needs: state.NeedState{
			Hunger: 20.0,
			Thirst: 20.0,
			Energy: 5.0, // Below critical (10)
		},
	}

	assert.False(t, ns.IsHealthy(exhaustedState), "Entity with critical energy should not be healthy")
}

// -----------------------------------------------------------------------------
// Scenario: Multipliers Affect Tick Rate
// -----------------------------------------------------------------------------
// Given: An entity with custom multipliers
// When: Tick is called with multipliers
// Then: Needs should change at modified rate
func TestBDD_NeedSystem_Multipliers(t *testing.T) {
	ns := state.NeedSystem{}

	// Create two identical starting states
	state1 := &state.LivingEntityState{
		Needs: state.NeedState{Hunger: 0.0, Thirst: 0.0, Energy: 100.0},
	}
	state2 := &state.LivingEntityState{
		Needs: state.NeedState{Hunger: 0.0, Thirst: 0.0, Energy: 100.0},
	}

	// Tick one with no multipliers
	ns.Tick(state1, nil)

	// Tick other with 2x hunger multiplier
	ns.Tick(state2, map[string]float64{"hunger": 2.0})

	// State2 should have more hunger
	assert.Greater(t, state2.Needs.Hunger, state1.Needs.Hunger,
		"Higher multiplier should increase hunger faster")
}

// -----------------------------------------------------------------------------
// Scenario: Reproduction Urge Increases When Healthy
// -----------------------------------------------------------------------------
// Given: A healthy entity
// When: Tick is called
// Then: Reproduction urge should increase
func TestBDD_NeedSystem_ReproductionUrge(t *testing.T) {
	ns := state.NeedSystem{}
	healthyState := &state.LivingEntityState{
		Needs: state.NeedState{
			Hunger:           20.0,
			Thirst:           20.0,
			Energy:           80.0,
			ReproductionUrge: 0.0,
		},
	}

	ns.Tick(healthyState, nil)

	assert.Greater(t, healthyState.Needs.ReproductionUrge, 0.0,
		"Reproduction urge should increase for healthy entity")
}
