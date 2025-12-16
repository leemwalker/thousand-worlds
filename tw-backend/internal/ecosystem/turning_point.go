// Package ecosystem provides turning point mechanics for player intervention.
// Turning points pause the simulation and present choices that shape world evolution.
package ecosystem

import (
	"github.com/google/uuid"
)

// TurningPointTrigger describes what caused a turning point
type TurningPointTrigger string

const (
	TriggerInterval      TurningPointTrigger = "interval"       // Regular time-based
	TriggerExtinction    TurningPointTrigger = "extinction"     // Major extinction event
	TriggerSapience      TurningPointTrigger = "sapience"       // Sapience emergence
	TriggerClimateShift  TurningPointTrigger = "climate_shift"  // Major climate change
	TriggerTectonicEvent TurningPointTrigger = "tectonic_event" // Continental split/collision
	TriggerPandemic      TurningPointTrigger = "pandemic"       // Major disease outbreak
	TriggerMagicEvent    TurningPointTrigger = "magic_event"    // Magic-related occurrence
	TriggerPlayerRequest TurningPointTrigger = "player_request" // Player manually paused
	TriggerMilestone     TurningPointTrigger = "milestone"      // Species count, diversity, etc.
)

// InterventionType categorizes available interventions
type InterventionType string

const (
	InterventionNone       InterventionType = "none"       // Observe only
	InterventionNudge      InterventionType = "nudge"      // Subtle influence
	InterventionDirect     InterventionType = "direct"     // Direct action
	InterventionCataclysm  InterventionType = "cataclysm"  // Major disruption
	InterventionMagic      InterventionType = "magic"      // Magical intervention
	InterventionProtection InterventionType = "protection" // Shield from harm
	InterventionAccelerate InterventionType = "accelerate" // Speed up evolution
)

// Intervention represents a choice available at a turning point
type Intervention struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        InterventionType `json:"intervention_type"`
	Cost        int              `json:"cost"`     // Divine energy or similar resource
	Cooldown    int64            `json:"cooldown"` // Years before can be used again

	// Targeting
	TargetType string     `json:"target_type"` // "species", "biome", "region", "world"
	TargetID   *uuid.UUID `json:"target_id,omitempty"`

	// Effects
	Effects []InterventionEffect `json:"effects"`

	// Consequences (shown to player)
	ProsDescription string  `json:"pros_description"`
	ConsDescription string  `json:"cons_description"`
	RiskLevel       float32 `json:"risk_level"` // 0.0-1.0
}

// InterventionEffect describes a single effect of an intervention
type InterventionEffect struct {
	EffectType  string  `json:"effect_type"` // e.g., "population_modifier", "trait_boost"
	TargetType  string  `json:"target_type"` // "species", "world", "region", "biome"
	TargetTrait string  `json:"target_trait,omitempty"`
	Magnitude   float32 `json:"magnitude"` // Effect strength
	Duration    int64   `json:"duration"`  // Years the effect lasts (0 = permanent)
	Description string  `json:"description"`
}

// TurningPoint represents a moment where the player can intervene
type TurningPoint struct {
	ID             uuid.UUID           `json:"id"`
	WorldID        uuid.UUID           `json:"world_id"`
	Year           int64               `json:"year"`
	Trigger        TurningPointTrigger `json:"trigger"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
	DetailedReport string              `json:"detailed_report"` // Longer lore text

	// State snapshot
	TotalSpecies   int `json:"total_species"`
	ExtantSpecies  int `json:"extant_species"`
	ExtinctSpecies int `json:"extinct_species"`
	SapientSpecies int `json:"sapient_species"`

	// Available choices
	Interventions []Intervention `json:"interventions"`

	// Resolution
	IsResolved         bool       `json:"is_resolved"`
	ChosenIntervention *uuid.UUID `json:"chosen_intervention,omitempty"`
	ResolvedYear       int64      `json:"resolved_year,omitempty"`

	// For extinction events
	ExtinctionDetails *ExtinctionEventDetails `json:"extinction_details,omitempty"`

	// For sapience events
	SapienceDetails *SapienceEventDetails `json:"sapience_details,omitempty"`
}

// ExtinctionEventDetails provides context for extinction turning points
type ExtinctionEventDetails struct {
	Cause            string      `json:"cause"`
	Severity         float32     `json:"severity"` // 0.0-1.0
	SpeciesLost      int         `json:"species_lost"`
	BiomesAffected   []uuid.UUID `json:"biomes_affected"`
	KeystoneSpecies  []uuid.UUID `json:"keystone_species"`  // Flagged for potential intervention
	RecoveryEstimate int64       `json:"recovery_estimate"` // Years to recover
}

// SapienceEventDetails provides context for sapience turning points
type SapienceEventDetails struct {
	SpeciesID         uuid.UUID `json:"species_id"`
	SpeciesName       string    `json:"species_name"`
	IsMagicAssisted   bool      `json:"is_magic_assisted"`
	IntelligenceLevel float64   `json:"intelligence_level"`
	PopulationSize    int64     `json:"population_size"`
	PredictedPath     string    `json:"predicted_path"` // e.g., "technological", "magical", "spiritual"
}

// TurningPointManager handles turning point creation and resolution
type TurningPointManager struct {
	WorldID             uuid.UUID                   `json:"world_id"`
	CurrentYear         int64                       `json:"current_year"`
	TurningPoints       map[uuid.UUID]*TurningPoint `json:"turning_points"`
	PendingTurningPoint *uuid.UUID                  `json:"pending_turning_point,omitempty"`

	// Divine Energy - resource for interventions
	DivineEnergy      int   `json:"divine_energy"`       // Current energy pool
	EnergyPerInterval int64 `json:"energy_per_interval"` // Years per 1 energy (default 10000)
	LastEnergyYear    int64 `json:"last_energy_year"`    // Year of last energy accumulation

	// Configuration
	IntervalYears       int64   `json:"interval_years"` // Years between interval triggers
	LastIntervalYear    int64   `json:"last_interval_year"`
	ExtinctionThreshold float32 `json:"extinction_threshold"` // Species loss % to trigger

	// Available intervention templates
	InterventionTemplates []Intervention `json:"intervention_templates"`

	// Cooldowns for interventions
	Cooldowns map[string]int64 `json:"cooldowns"` // Intervention name -> year available
}

// NewTurningPointManager creates a new manager
func NewTurningPointManager(worldID uuid.UUID) *TurningPointManager {
	return &TurningPointManager{
		WorldID:               worldID,
		TurningPoints:         make(map[uuid.UUID]*TurningPoint),
		DivineEnergy:          10,      // Start with some energy
		EnergyPerInterval:     10000,   // 1 energy per 10k years
		IntervalYears:         1000000, // 1 million years default
		ExtinctionThreshold:   0.25,    // 25% species loss triggers event
		Cooldowns:             make(map[string]int64),
		InterventionTemplates: defaultInterventionTemplates(),
	}
}

// AccumulateEnergy adds Divine Energy based on years elapsed
// Called periodically during simulation
func (tpm *TurningPointManager) AccumulateEnergy(currentYear int64) {
	if tpm.EnergyPerInterval <= 0 {
		return
	}
	yearsElapsed := currentYear - tpm.LastEnergyYear
	energy := int(yearsElapsed / tpm.EnergyPerInterval)
	if energy > 0 {
		tpm.DivineEnergy += energy
		tpm.LastEnergyYear = currentYear - (yearsElapsed % tpm.EnergyPerInterval)
	}
}

// CanAfford returns true if there's enough Divine Energy for the intervention
func (tpm *TurningPointManager) CanAfford(cost int) bool {
	return tpm.DivineEnergy >= cost
}

// SpendEnergy deducts Divine Energy for an intervention
func (tpm *TurningPointManager) SpendEnergy(cost int) bool {
	if !tpm.CanAfford(cost) {
		return false
	}
	tpm.DivineEnergy -= cost
	return true
}

// defaultInterventionTemplates returns the standard interventions
func defaultInterventionTemplates() []Intervention {
	return []Intervention{
		{
			ID:              uuid.New(),
			Name:            "Observe Only",
			Description:     "Watch and let nature take its course",
			Type:            InterventionNone,
			Cost:            0,
			ProsDescription: "No interference with natural processes",
			ConsDescription: "No control over outcomes",
			RiskLevel:       0,
		},
		{
			ID:          uuid.New(),
			Name:        "Gentle Nudge",
			Description: "Subtly influence evolutionary pressure",
			Type:        InterventionNudge,
			Cost:        10,
			TargetType:  "species",
			Effects: []InterventionEffect{
				{EffectType: "trait_boost", Magnitude: 0.1, Duration: 100000, Description: "Slight trait enhancement"},
			},
			ProsDescription: "Gentle influence without major disruption",
			ConsDescription: "Effects may not be significant enough",
			RiskLevel:       0.1,
		},
		{
			ID:          uuid.New(),
			Name:        "Divine Protection",
			Description: "Shield a species from extinction",
			Type:        InterventionProtection,
			Cost:        50,
			Cooldown:    500000,
			TargetType:  "species",
			Effects: []InterventionEffect{
				{EffectType: "extinction_immunity", Magnitude: 1.0, Duration: 250000, Description: "Cannot go extinct"},
			},
			ProsDescription: "Guarantees species survival",
			ConsDescription: "Disrupts natural selection; high cost",
			RiskLevel:       0.3,
		},
		{
			ID:          uuid.New(),
			Name:        "Accelerate Evolution",
			Description: "Dramatically speed up mutation and adaptation",
			Type:        InterventionAccelerate,
			Cost:        30,
			Cooldown:    250000,
			TargetType:  "species",
			Effects: []InterventionEffect{
				{EffectType: "mutation_rate", Magnitude: 3.0, Duration: 50000, Description: "3x mutation rate"},
				{EffectType: "speciation_rate", Magnitude: 2.0, Duration: 50000, Description: "2x speciation chance"},
			},
			ProsDescription: "Rapid diversification and adaptation",
			ConsDescription: "May cause instability; unpredictable results",
			RiskLevel:       0.5,
		},
		{
			ID:          uuid.New(),
			Name:        "Targeted Cataclysm",
			Description: "Cause a localized extinction event",
			Type:        InterventionCataclysm,
			Cost:        40,
			Cooldown:    1000000,
			TargetType:  "biome",
			Effects: []InterventionEffect{
				{EffectType: "population_modifier", Magnitude: -0.7, Duration: 0, Description: "70% population loss"},
				{EffectType: "radiation_bonus", Magnitude: 3.0, Duration: 100000, Description: "Adaptive radiation follows"},
			},
			ProsDescription: "Clears ecological space for new evolution",
			ConsDescription: "Mass death; unpredictable cascade effects",
			RiskLevel:       0.8,
		},
		{
			ID:          uuid.New(),
			Name:        "Spark of Sapience",
			Description: "Uplift a species toward intelligence",
			Type:        InterventionMagic,
			Cost:        100,
			Cooldown:    5000000,
			TargetType:  "species",
			Effects: []InterventionEffect{
				{EffectType: "trait_boost", TargetTrait: "intelligence", Magnitude: 2.0, Duration: 0, Description: "Permanent intelligence boost"},
				{EffectType: "trait_boost", TargetTrait: "tool_use", Magnitude: 1.5, Duration: 0, Description: "Permanent tool use boost"},
			},
			ProsDescription: "Directly creates a sapient species",
			ConsDescription: "Extremely expensive; may disrupt natural path to sapience",
			RiskLevel:       0.4,
		},
		{
			ID:          uuid.New(),
			Name:        "Infuse Magic",
			Description: "Grant magical potential to a species",
			Type:        InterventionMagic,
			Cost:        75,
			Cooldown:    2000000,
			TargetType:  "species",
			Effects: []InterventionEffect{
				{EffectType: "trait_boost", TargetTrait: "magic_affinity", Magnitude: 5.0, Duration: 0, Description: "Major magic affinity boost"},
			},
			ProsDescription: "Opens path to magic-assisted sapience",
			ConsDescription: "Requires magic-enabled world; high cost",
			RiskLevel:       0.3,
		},
	}
}

// CheckForTurningPoint evaluates if a turning point should occur
func (tpm *TurningPointManager) CheckForTurningPoint(
	year int64,
	totalSpecies int,
	recentExtinctions int,
	newSapientSpecies []uuid.UUID,
	significantEvent string,
) *TurningPoint {
	tpm.CurrentYear = year

	// Don't trigger if one is already pending
	if tpm.PendingTurningPoint != nil {
		return nil
	}

	var trigger TurningPointTrigger
	var title, description string

	// Check for sapience emergence (highest priority)
	if len(newSapientSpecies) > 0 {
		trigger = TriggerSapience
		title = "Sapience Emerges"
		description = "A species has achieved sapience. The age of thought begins."
	} else if totalSpecies > 0 && float32(recentExtinctions)/float32(totalSpecies) >= tpm.ExtinctionThreshold {
		// Check for mass extinction
		trigger = TriggerExtinction
		title = "Mass Extinction"
		description = "A catastrophic extinction event threatens the biosphere."
	} else if year-tpm.LastIntervalYear >= tpm.IntervalYears {
		// Check for interval trigger
		trigger = TriggerInterval
		title = "Era Milestone"
		description = "A new era dawns. The world awaits your guidance."
		tpm.LastIntervalYear = year
	} else if significantEvent != "" {
		// Check for special events
		switch significantEvent {
		case "climate_shift":
			trigger = TriggerClimateShift
			title = "Climate Upheaval"
			description = "The world's climate undergoes dramatic transformation."
		case "tectonic_event":
			trigger = TriggerTectonicEvent
			title = "Continental Drift"
			description = "The land itself shifts, creating new barriers and bridges."
		case "pandemic":
			trigger = TriggerPandemic
			title = "The Great Plague"
			description = "Disease sweeps across populations."
		default:
			return nil
		}
	} else {
		return nil
	}

	// Create turning point
	tp := &TurningPoint{
		ID:             uuid.New(),
		WorldID:        tpm.WorldID,
		Year:           year,
		Trigger:        trigger,
		Title:          title,
		Description:    description,
		TotalSpecies:   totalSpecies,
		ExtantSpecies:  totalSpecies - recentExtinctions,
		ExtinctSpecies: recentExtinctions,
		Interventions:  tpm.getAvailableInterventions(trigger),
	}

	tpm.TurningPoints[tp.ID] = tp
	tpm.PendingTurningPoint = &tp.ID

	return tp
}

// getAvailableInterventions returns interventions available for a trigger type
func (tpm *TurningPointManager) getAvailableInterventions(trigger TurningPointTrigger) []Intervention {
	available := make([]Intervention, 0)

	for _, template := range tpm.InterventionTemplates {
		// Check cooldown
		if cooldownEnd, exists := tpm.Cooldowns[template.Name]; exists {
			if tpm.CurrentYear < cooldownEnd {
				continue // Still on cooldown
			}
		}

		// All triggers get "Observe Only"
		if template.Type == InterventionNone {
			available = append(available, template)
			continue
		}

		// Filter by trigger appropriateness
		switch trigger {
		case TriggerExtinction:
			if template.Type == InterventionProtection ||
				template.Type == InterventionNudge ||
				template.Type == InterventionAccelerate {
				available = append(available, template)
			}
		case TriggerSapience:
			if template.Type == InterventionNudge ||
				template.Type == InterventionMagic {
				available = append(available, template)
			}
		case TriggerInterval:
			// Most options available at interval
			available = append(available, template)
		default:
			if template.Type == InterventionNudge ||
				template.Type == InterventionProtection {
				available = append(available, template)
			}
		}
	}

	return available
}

// ResolveturningPoint applies a chosen intervention
func (tpm *TurningPointManager) ResolveTurningPoint(
	turningPointID uuid.UUID,
	interventionID uuid.UUID,
) *TurningPoint {
	tp, exists := tpm.TurningPoints[turningPointID]
	if !exists || tp.IsResolved {
		return nil
	}

	// Find chosen intervention
	var chosen *Intervention
	for i := range tp.Interventions {
		if tp.Interventions[i].ID == interventionID {
			chosen = &tp.Interventions[i]
			break
		}
	}

	if chosen == nil {
		return nil
	}

	// Apply cooldown
	if chosen.Cooldown > 0 {
		tpm.Cooldowns[chosen.Name] = tpm.CurrentYear + chosen.Cooldown
	}

	// Mark resolved
	tp.IsResolved = true
	tp.ChosenIntervention = &interventionID
	tp.ResolvedYear = tpm.CurrentYear
	tpm.PendingTurningPoint = nil

	return tp
}

// GetPendingTurningPoint returns the current pending turning point
func (tpm *TurningPointManager) GetPendingTurningPoint() *TurningPoint {
	if tpm.PendingTurningPoint == nil {
		return nil
	}
	return tpm.TurningPoints[*tpm.PendingTurningPoint]
}

// IsPaused returns true if simulation should be paused for player input
func (tpm *TurningPointManager) IsPaused() bool {
	return tpm.PendingTurningPoint != nil
}

// GetHistory returns all resolved turning points
func (tpm *TurningPointManager) GetHistory() []*TurningPoint {
	history := make([]*TurningPoint, 0)
	for _, tp := range tpm.TurningPoints {
		if tp.IsResolved {
			history = append(history, tp)
		}
	}
	return history
}
