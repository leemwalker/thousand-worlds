package state

import (
	"testing"
	"tw-backend/internal/npc/genetics"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNeedSystem_Tick(t *testing.T) {
	sys := &NeedSystem{}

	// Setup
	state := &LivingEntityState{
		EntityID: uuid.New(),
		Needs: NeedState{
			Hunger:           0,
			Thirst:           0,
			Energy:           100,
			ReproductionUrge: 0,
		},
		DNA: genetics.NewDNA(),
	}

	// 1. Tick with default multipliers
	sys.Tick(state, nil)

	// Verify base decay
	assert.Equal(t, BaseHungerRate, state.Needs.Hunger, "Hunger should increase by base rate")
	assert.Equal(t, BaseThirstRate, state.Needs.Thirst, "Thirst should increase by base rate")
	assert.Equal(t, 100.0-BaseEnergyRate, state.Needs.Energy, "Energy should decrease by base rate")
	assert.Equal(t, BaseReproductionRate, state.Needs.ReproductionUrge, "Reproduction should increase by base rate")

	// 2. Tick with multipliers (e.g. Desert heat)
	state.Needs = NeedState{Hunger: 0, Thirst: 0, Energy: 100, ReproductionUrge: 0} // Reset
	multipliers := map[string]float64{
		"thirst": 2.0,
	}
	sys.Tick(state, multipliers)

	assert.Equal(t, BaseThirstRate*2.0, state.Needs.Thirst, "Thirst should increase by 2x base rate")
	assert.Equal(t, BaseHungerRate, state.Needs.Hunger, "Hunger should be unaffected")
}

func TestNeedSystem_IsHealthy(t *testing.T) {
	sys := &NeedSystem{}

	state := &LivingEntityState{
		Needs: NeedState{
			Hunger: 50,
			Thirst: 50,
			Energy: 50,
		},
	}
	assert.True(t, sys.IsHealthy(state))

	state.Needs.Hunger = ThresholdHungerCritical + 1
	assert.False(t, sys.IsHealthy(state))

	state.Needs.Hunger = 0
	state.Needs.Thirst = ThresholdThirstCritical + 1
	assert.False(t, sys.IsHealthy(state))

	state.Needs.Thirst = 0
	state.Needs.Energy = ThresholdEnergyCritical - 1
	assert.False(t, sys.IsHealthy(state))
}

func TestNeedSystem_Clamping(t *testing.T) {
	sys := &NeedSystem{}
	state := &LivingEntityState{
		Needs: NeedState{Hunger: 99.9, Energy: 0.1},
	}

	// Push over limits
	multipliers := map[string]float64{"hunger": 100.0, "energy": 100.0}
	sys.Tick(state, multipliers)

	assert.Equal(t, 100.0, state.Needs.Hunger)
	assert.Equal(t, 0.0, state.Needs.Energy)
}
