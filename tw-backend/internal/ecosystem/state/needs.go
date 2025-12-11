package state

import (
	"math"
)

// NeedConstants defines thresholds and base rates
const (
	MaxNeedValue = 100.0
	MinNeedValue = 0.0

	// Thresholds where needs become critical actions
	ThresholdHungerCritical = 80.0 // When hunger is high, it's critical (assuming 0=Full, 100=Starving)
	ThresholdThirstCritical = 85.0
	ThresholdEnergyCritical = 10.0 // When energy is low, it's critical (assuming 100=Rested, 0=Exhausted)
	ThresholdMateReady      = 80.0

	// Base Decay Rates (per tick)
	BaseHungerRate       = 0.05
	BaseThirstRate       = 0.08
	BaseEnergyRate       = 0.03
	BaseReproductionRate = 0.2
)

// NeedSystem manages the calculation of needs
type NeedSystem struct{}

// Tick updates the needs for a single entity based on time passing
func (s *NeedSystem) Tick(state *LivingEntityState, multipliers map[string]float64) {
	// Apply base changes
	// Hunger increases (0 -> 100)
	hungerMult := getMultiplier(multipliers, "hunger", 1.0)
	state.Needs.Hunger = clamp(state.Needs.Hunger+(BaseHungerRate*hungerMult), MinNeedValue, MaxNeedValue)

	// Thirst increases (0 -> 100)
	thirstMult := getMultiplier(multipliers, "thirst", 1.0)
	state.Needs.Thirst = clamp(state.Needs.Thirst+(BaseThirstRate*thirstMult), MinNeedValue, MaxNeedValue)

	// Energy decreases (100 -> 0)
	energyMult := getMultiplier(multipliers, "energy", 1.0)
	state.Needs.Energy = clamp(state.Needs.Energy-(BaseEnergyRate*energyMult), MinNeedValue, MaxNeedValue)

	// Reproduction Urge increases if healthy (0 -> 100)
	if s.IsHealthy(state) {
		reproMult := getMultiplier(multipliers, "reproduction", 1.0)
		state.Needs.ReproductionUrge = clamp(state.Needs.ReproductionUrge+(BaseReproductionRate*reproMult), MinNeedValue, MaxNeedValue)
	}
}

// IsHealthy returns true if the entity is not in a critical state
func (s *NeedSystem) IsHealthy(state *LivingEntityState) bool {
	return state.Needs.Hunger < ThresholdHungerCritical &&
		state.Needs.Thirst < ThresholdThirstCritical &&
		state.Needs.Energy > ThresholdEnergyCritical
}

func getMultiplier(m map[string]float64, key string, def float64) float64 {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		return v
	}
	return def
}

func clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(max, val))
}
