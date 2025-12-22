// Package simulation provides the unified simulation engine.
// This package eliminates duplication between "world simulate" (headless)
// and "world runner" (interactive) by extracting shared simulation logic.
package simulation

import (
	"tw-backend/internal/ecosystem/pathogen"
	"tw-backend/internal/ecosystem/population"

	"github.com/google/uuid"
)

// SimulationMode determines how the engine handles turning points
type SimulationMode string

const (
	ModeHeadless    SimulationMode = "headless"    // Fast-forward, auto-resolve turning points
	ModeInteractive SimulationMode = "interactive" // Real-time, pause for player decisions
)

// SimulationSpeed controls how many simulation years per tick
type SimulationSpeed int

const (
	SpeedPaused SimulationSpeed = 0
	SpeedSlow   SimulationSpeed = 1
	SpeedNormal SimulationSpeed = 10
	SpeedFast   SimulationSpeed = 100
	SpeedTurbo  SimulationSpeed = 1000
)

// SimulationEvent represents notable occurrences during simulation
type SimulationEvent struct {
	Year        int64
	Type        string // "speciation", "extinction", "outbreak", "migration", etc.
	Description string
	Importance  int // 1-10
	SpeciesID   *uuid.UUID
}

// EventHandler is called when simulation events occur
type EventHandler func(event SimulationEvent)

// Subsystems holds optional subsystem references
type Subsystems struct {
	DiseaseSystem *pathogen.DiseaseSystem
	// GeologyManager *WorldGeology // Not imported to avoid cycle
}

// StepResult contains the outcome of a simulation step
type StepResult struct {
	YearsAdvanced  int64
	NewSpecies     int
	MigrationCount int64
	OutbreakCount  int
	GeologyUpdated bool
	Events         []SimulationEvent
}

// StepConfig configures a single simulation step
type StepConfig struct {
	SimulateLife     bool
	SimulateGeology  bool
	SimulateDiseases bool
	GlobalTempMod    float64
	EventHandler     EventHandler
}

// DefaultStepConfig returns a default configuration with all subsystems enabled
func DefaultStepConfig() StepConfig {
	return StepConfig{
		SimulateLife:     true,
		SimulateGeology:  true,
		SimulateDiseases: true,
		GlobalTempMod:    0.0,
	}
}

// SimulateYear advances the population simulation by one year.
// This is the PRIMARY shared function that both headless and interactive modes use.
func SimulateYear(popSim *population.PopulationSimulator, config StepConfig) {
	if config.SimulateLife && popSim != nil {
		popSim.SimulateYear()
	}
}

// ApplyPeriodicEvolution applies evolution effects every 1000 years.
// Returns true if evolution was applied this year.
func ApplyPeriodicEvolution(popSim *population.PopulationSimulator, config StepConfig) bool {
	if !config.SimulateLife || popSim == nil {
		return false
	}

	// Only run at 1000-year intervals, and not at year 0
	if popSim.CurrentYear == 0 || popSim.CurrentYear%1000 != 0 {
		return false
	}

	popSim.ApplyEvolution()
	popSim.ApplyCoEvolution()
	popSim.ApplyGeneticDrift()
	popSim.ApplySexualSelection()

	return true
}

// ApplyPeriodicSpeciation handles speciation, migration, oxygen levels every 10000 years.
// Returns StepResult with new species count and migration count.
func ApplyPeriodicSpeciation(popSim *population.PopulationSimulator, config StepConfig) (newSpecies int, migrants int64) {
	if !config.SimulateLife || popSim == nil {
		return 0, 0
	}

	if popSim.CurrentYear%10000 != 0 {
		return 0, 0
	}

	// Atmospheric oxygen
	popSim.UpdateOxygenLevel()
	popSim.ApplyOxygenEffects()

	// Speciation
	newSpecies = popSim.CheckSpeciation()

	// Migration
	migrants = popSim.ApplyMigrationCycle()

	return newSpecies, migrants
}

// ApplyDiseaseUpdate runs the disease system every 10000 years.
// Returns outbreak count.
func ApplyDiseaseUpdate(
	popSim *population.PopulationSimulator,
	diseaseSystem *pathogen.DiseaseSystem,
	config StepConfig,
) int {
	if !config.SimulateDiseases || !config.SimulateLife || popSim == nil || diseaseSystem == nil {
		return 0
	}

	if popSim.CurrentYear%10000 != 0 {
		return 0
	}

	outbreakCount := 0

	// Build species data for disease system
	speciesData := make(map[uuid.UUID]pathogen.SpeciesInfo)
	for _, biome := range popSim.Biomes {
		for _, sp := range biome.Species {
			if sp.Count > 0 {
				speciesData[sp.SpeciesID] = pathogen.SpeciesInfo{
					Population:        sp.Count,
					DiseaseResistance: float32(sp.Traits.DiseaseResistance),
					DietType:          string(sp.Diet),
					Density:           float64(sp.Count) / float64(biome.CarryingCapacity+1),
				}

				// Check for spontaneous outbreaks
				_, outbreak := diseaseSystem.CheckSpontaneousOutbreak(
					sp.SpeciesID, sp.Name, sp.Count,
					float64(sp.Count)/float64(biome.CarryingCapacity+1),
				)
				if outbreak != nil {
					outbreakCount++
				}
			}
		}
	}

	// Update all active outbreaks
	diseaseSystem.Update(popSim.CurrentYear, speciesData)

	return outbreakCount
}

// ShouldUpdateGeology returns true if geology should be updated this year
func ShouldUpdateGeology(currentYear int64, config StepConfig) bool {
	return config.SimulateGeology && currentYear%10000 == 0 && currentYear > 0
}

// Step executes a complete simulation step for the given number of years.
// This is the UNIFIED step function used by both headless and interactive modes.
func Step(
	popSim *population.PopulationSimulator,
	subsystems Subsystems,
	yearsToAdvance int64,
	config StepConfig,
) StepResult {
	result := StepResult{
		Events: make([]SimulationEvent, 0),
	}

	for i := int64(0); i < yearsToAdvance; i++ {
		// Core simulation
		SimulateYear(popSim, config)

		// Periodic evolution (every 1000 years)
		ApplyPeriodicEvolution(popSim, config)

		// Speciation & migration (every 10000 years)
		newSpecies, migrants := ApplyPeriodicSpeciation(popSim, config)
		if newSpecies > 0 {
			result.NewSpecies += newSpecies
			event := SimulationEvent{
				Year:        popSim.CurrentYear,
				Type:        "speciation",
				Description: "New species evolved",
				Importance:  7,
			}
			result.Events = append(result.Events, event)
			if config.EventHandler != nil {
				config.EventHandler(event)
			}
		}
		result.MigrationCount += migrants

		// Disease updates (every 10000 years)
		outbreaks := ApplyDiseaseUpdate(popSim, subsystems.DiseaseSystem, config)
		result.OutbreakCount += outbreaks

		// Geology check (caller handles actual geology update)
		if ShouldUpdateGeology(popSim.CurrentYear, config) {
			result.GeologyUpdated = true
		}

		result.YearsAdvanced++
	}

	return result
}
