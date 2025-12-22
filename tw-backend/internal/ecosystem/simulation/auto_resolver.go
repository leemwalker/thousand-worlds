// Package simulation provides auto-resolution for turning points in headless mode.
package simulation

import (
	"math/rand"

	"github.com/google/uuid"
)

// TurningPointInfo provides the data needed to resolve a turning point.
// This is a subset of ecosystem.TurningPoint to avoid import cycles.
type TurningPointInfo struct {
	ID            uuid.UUID
	Trigger       string // "extinction", "sapience", "interval", etc.
	Title         string
	Interventions []InterventionOption
}

// InterventionOption represents a choice available at a turning point.
type InterventionOption struct {
	ID        uuid.UUID
	Name      string
	Type      string // "none", "nudge", "direct", "catastrophe", etc.
	Cost      int
	RiskLevel float32
}

// AutoResolver automatically selects interventions for headless mode.
// Uses a seeded RNG for deterministic results given the same simulation seed.
type AutoResolver struct {
	seed int64
	rng  *rand.Rand
}

// NewAutoResolver creates a new auto-resolver with the given seed.
func NewAutoResolver(seed int64) *AutoResolver {
	return &AutoResolver{
		seed: seed,
		rng:  rand.New(rand.NewSource(seed)),
	}
}

// Resolve selects an intervention based on the turning point type and available options.
// The selection is deterministic given the same seed and turning point.
// Returns the index of the selected intervention, or -1 if none selected.
func (ar *AutoResolver) Resolve(tp TurningPointInfo) int {
	if len(tp.Interventions) == 0 {
		return -1 // No interventions available
	}

	// Strategy: Select based on trigger type and risk assessment
	switch tp.Trigger {
	case "extinction":
		// During extinction, prefer protective interventions
		return ar.selectByPreference(tp.Interventions, []string{"protection", "nudge", "none"})

	case "sapience":
		// For sapience events, prefer accelerating or nudging
		return ar.selectByPreference(tp.Interventions, []string{"accelerate", "nudge", "none"})

	case "interval":
		// Regular check-in: prefer low-risk or observing
		return ar.selectByPreference(tp.Interventions, []string{"none", "nudge"})

	default:
		// Unknown trigger: pick randomly with bias toward lower risk
		return ar.selectLowestRisk(tp.Interventions)
	}
}

// selectByPreference selects an intervention by type preference order.
// Falls back to random selection if no preferred type found.
func (ar *AutoResolver) selectByPreference(interventions []InterventionOption, preferences []string) int {
	for _, pref := range preferences {
		for i, intervention := range interventions {
			if intervention.Type == pref {
				return i
			}
		}
	}
	// No preferred type found, select randomly
	return ar.rng.Intn(len(interventions))
}

// selectLowestRisk selects the intervention with the lowest risk level.
func (ar *AutoResolver) selectLowestRisk(interventions []InterventionOption) int {
	if len(interventions) == 0 {
		return -1
	}

	lowestIdx := 0
	lowestRisk := interventions[0].RiskLevel

	for i, intervention := range interventions {
		if intervention.RiskLevel < lowestRisk {
			lowestRisk = intervention.RiskLevel
			lowestIdx = i
		}
	}

	return lowestIdx
}

// ResolveDeterministic always returns the same result for the same inputs.
// Useful for testing or when exact reproducibility is required.
func (ar *AutoResolver) ResolveDeterministic(tp TurningPointInfo) int {
	// Reset RNG to initial state for this turning point
	// Use turning point ID as additional seed component
	combinedSeed := ar.seed ^ int64(tp.ID[0])<<56 | int64(tp.ID[1])<<48 |
		int64(tp.ID[2])<<40 | int64(tp.ID[3])<<32 |
		int64(tp.ID[4])<<24 | int64(tp.ID[5])<<16 |
		int64(tp.ID[6])<<8 | int64(tp.ID[7])

	ar.rng = rand.New(rand.NewSource(combinedSeed))
	return ar.Resolve(tp)
}
