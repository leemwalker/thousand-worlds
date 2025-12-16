// Package ecosystem provides effect application for turning point interventions.
package ecosystem

import (
	"github.com/google/uuid"
)

// EffectTarget represents what an effect can target
type EffectTarget string

const (
	EffectTargetSpecies EffectTarget = "species"
	EffectTargetWorld   EffectTarget = "world"
	EffectTargetRegion  EffectTarget = "region"
	EffectTargetBiome   EffectTarget = "biome"
)

// ActiveEffect tracks an effect currently applied to the simulation
type ActiveEffect struct {
	ID         uuid.UUID    `json:"id"`
	EffectType string       `json:"effect_type"`
	TargetType EffectTarget `json:"target_type"`
	TargetID   uuid.UUID    `json:"target_id"`
	Magnitude  float32      `json:"magnitude"`
	StartYear  int64        `json:"start_year"`
	Duration   int64        `json:"duration"` // 0 = permanent
	Data       string       `json:"data,omitempty"`
}

// IsExpired returns true if the effect has expired
func (ae *ActiveEffect) IsExpired(currentYear int64) bool {
	if ae.Duration == 0 {
		return false // Permanent
	}
	return currentYear >= ae.StartYear+ae.Duration
}

// EffectApplicator manages applying and tracking intervention effects
type EffectApplicator struct {
	WorldID       uuid.UUID      `json:"world_id"`
	CurrentYear   int64          `json:"current_year"`
	ActiveEffects []ActiveEffect `json:"active_effects"`
}

// NewEffectApplicator creates a new effect applicator for a world
func NewEffectApplicator(worldID uuid.UUID) *EffectApplicator {
	return &EffectApplicator{
		WorldID:       worldID,
		ActiveEffects: make([]ActiveEffect, 0),
	}
}

// ApplyEffect applies an intervention effect and tracks it
func (ea *EffectApplicator) ApplyEffect(effect InterventionEffect, targetID uuid.UUID) *ActiveEffect {
	active := ActiveEffect{
		ID:         uuid.New(),
		EffectType: effect.EffectType,
		TargetType: EffectTarget(effect.TargetType),
		TargetID:   targetID,
		Magnitude:  effect.Magnitude,
		StartYear:  ea.CurrentYear,
		Duration:   effect.Duration,
		Data:       effect.TargetTrait, // Store trait name or other data
	}
	ea.ActiveEffects = append(ea.ActiveEffects, active)
	return &active
}

// CleanupExpired removes expired effects
func (ea *EffectApplicator) CleanupExpired() int {
	activeCount := 0
	newEffects := make([]ActiveEffect, 0, len(ea.ActiveEffects))
	for _, effect := range ea.ActiveEffects {
		if !effect.IsExpired(ea.CurrentYear) {
			newEffects = append(newEffects, effect)
			activeCount++
		}
	}
	ea.ActiveEffects = newEffects
	return activeCount
}

// GetEffectsForTarget returns all active effects for a target
func (ea *EffectApplicator) GetEffectsForTarget(targetID uuid.UUID) []ActiveEffect {
	effects := make([]ActiveEffect, 0)
	for _, effect := range ea.ActiveEffects {
		if effect.TargetID == targetID && !effect.IsExpired(ea.CurrentYear) {
			effects = append(effects, effect)
		}
	}
	return effects
}

// GetTraitModifier returns the cumulative trait modifier for a species
func (ea *EffectApplicator) GetTraitModifier(speciesID uuid.UUID, traitName string) float32 {
	var modifier float32 = 0
	for _, effect := range ea.ActiveEffects {
		if effect.TargetID == speciesID &&
			effect.EffectType == "trait_boost" &&
			effect.Data == traitName &&
			!effect.IsExpired(ea.CurrentYear) {
			modifier += effect.Magnitude
		}
	}
	return modifier
}

// HasExtinctionImmunity checks if a species has extinction immunity
func (ea *EffectApplicator) HasExtinctionImmunity(speciesID uuid.UUID) bool {
	for _, effect := range ea.ActiveEffects {
		if effect.TargetID == speciesID &&
			effect.EffectType == "extinction_immunity" &&
			!effect.IsExpired(ea.CurrentYear) {
			return true
		}
	}
	return false
}

// GetMutationMultiplier returns the mutation rate multiplier for a species
func (ea *EffectApplicator) GetMutationMultiplier(speciesID uuid.UUID) float32 {
	var multiplier float32 = 1.0
	for _, effect := range ea.ActiveEffects {
		if effect.TargetID == speciesID &&
			effect.EffectType == "mutation_rate" &&
			!effect.IsExpired(ea.CurrentYear) {
			multiplier *= effect.Magnitude
		}
	}
	return multiplier
}

// GetPopulationModifier returns population modifier for a species
func (ea *EffectApplicator) GetPopulationModifier(speciesID uuid.UUID) float32 {
	var modifier float32 = 0
	for _, effect := range ea.ActiveEffects {
		if effect.TargetID == speciesID &&
			effect.EffectType == "population_modifier" &&
			!effect.IsExpired(ea.CurrentYear) {
			modifier += effect.Magnitude
		}
	}
	return modifier
}

// HasPower checks if a species has been granted a specific power
func (ea *EffectApplicator) HasPower(speciesID uuid.UUID, powerID string) bool {
	for _, effect := range ea.ActiveEffects {
		if effect.TargetID == speciesID &&
			effect.EffectType == "power_infusion" &&
			effect.Data == powerID &&
			!effect.IsExpired(ea.CurrentYear) {
			return true
		}
	}
	return false
}

// GetPowers returns all powers granted to a species
func (ea *EffectApplicator) GetPowers(speciesID uuid.UUID) []string {
	var powers []string
	for _, effect := range ea.ActiveEffects {
		if effect.TargetID == speciesID &&
			effect.EffectType == "power_infusion" &&
			!effect.IsExpired(ea.CurrentYear) {
			powers = append(powers, effect.Data)
		}
	}
	return powers
}

// ApplyTemperatureShift tracks a global temperature change
func (ea *EffectApplicator) ApplyTemperatureShift(magnitude float32, duration int64) *ActiveEffect {
	return ea.ApplyEffect(InterventionEffect{
		EffectType: "temperature_shift",
		TargetType: "world",
		Magnitude:  magnitude,
		Duration:   duration,
	}, ea.WorldID)
}

// GetTemperatureModifier returns the cumulative temperature shift
func (ea *EffectApplicator) GetTemperatureModifier() float32 {
	var modifier float32 = 0
	for _, effect := range ea.ActiveEffects {
		if effect.EffectType == "temperature_shift" &&
			!effect.IsExpired(ea.CurrentYear) {
			modifier += effect.Magnitude
		}
	}
	return modifier
}

// ApplyTectonicCollision tracks a tectonic collision event
func (ea *EffectApplicator) ApplyTectonicCollision(regionID uuid.UUID, magnitude float32) *ActiveEffect {
	return ea.ApplyEffect(InterventionEffect{
		EffectType: "tectonic_collision",
		TargetType: "region",
		Magnitude:  magnitude,
		Duration:   0, // Mountains are permanent
	}, regionID)
}

// ApplyCatastrophe triggers a catastrophic event
func (ea *EffectApplicator) ApplyCatastrophe(catastropheType string, regionID uuid.UUID, magnitude float32) *ActiveEffect {
	return ea.ApplyEffect(InterventionEffect{
		EffectType:  "trigger_catastrophe",
		TargetType:  "region",
		TargetTrait: catastropheType, // Store catastrophe type in Data field
		Magnitude:   magnitude,
		Duration:    100000, // Catastrophe effects last 100k years
	}, regionID)
}
