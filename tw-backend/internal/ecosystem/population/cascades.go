// Package population provides extinction cascade mechanics for the ecosystem.
// This simulates how one extinction can trigger others through ecological dependencies.
package population

import (
	"math"

	"github.com/google/uuid"
)

// CascadeType represents the mechanism of extinction cascade
type CascadeType string

const (
	CascadeCoExtinction    CascadeType = "co_extinction"    // Obligate symbiont dies
	CascadePredatorLoss    CascadeType = "predator_loss"    // Prey loses predator control
	CascadeFoodLoss        CascadeType = "food_loss"        // Predator loses food source
	CascadePollinationLoss CascadeType = "pollination_loss" // Plant loses pollinator
	CascadeHabitatLoss     CascadeType = "habitat_loss"     // Ecosystem engineer dies
	CascadeCompetitorLoss  CascadeType = "competitor_loss"  // Competitive release
	CascadeKeystone        CascadeType = "keystone"         // Keystone species collapse
)

// EcologicalRole represents a species' role in the ecosystem
type EcologicalRole string

const (
	RoleApexPredator      EcologicalRole = "apex_predator"      // Top of food chain
	RoleMesopredator      EcologicalRole = "mesopredator"       // Mid-level predator
	RoleKeystone          EcologicalRole = "keystone"           // Disproportionate ecosystem effect
	RolePollinator        EcologicalRole = "pollinator"         // Plant reproduction
	RoleSeedDisperser     EcologicalRole = "seed_disperser"     // Plant distribution
	RoleDecomposer        EcologicalRole = "decomposer"         // Nutrient cycling
	RoleEcosystemEngineer EcologicalRole = "ecosystem_engineer" // Habitat modification
	RoleFoundationSpecies EcologicalRole = "foundation_species" // Habitat creation (e.g., corals, trees)
	RolePrimaryProducer   EcologicalRole = "primary_producer"   // Autotrophs
)

// EcologicalRelationship represents a dependency between species
type EcologicalRelationship struct {
	SourceSpeciesID uuid.UUID        `json:"source_species_id"` // The dependent species
	TargetSpeciesID uuid.UUID        `json:"target_species_id"` // The species being depended upon
	Type            RelationshipType `json:"relationship_type"`
	Strength        float32          `json:"strength"`    // 0.0-1.0, how critical the relationship is
	IsObligate      bool             `json:"is_obligate"` // True if source cannot survive without target
}

// RelationshipType describes ecological relationships
type RelationshipType string

const (
	RelationshipPredation    RelationshipType = "predation"    // Predator-prey
	RelationshipMutualism    RelationshipType = "mutualism"    // Both benefit
	RelationshipCommensalism RelationshipType = "commensalism" // One benefits, other neutral
	RelationshipParasitism   RelationshipType = "parasitism"   // One benefits, other harmed
	RelationshipCompetition  RelationshipType = "competition"  // Both negatively affected
	RelationshipHabitat      RelationshipType = "habitat"      // One provides habitat for other
)

// CascadeEvent represents a single cascade effect triggered by an extinction
type CascadeEvent struct {
	Year                int64       `json:"year"`
	TriggerSpeciesID    uuid.UUID   `json:"trigger_species_id"`
	TriggerSpeciesName  string      `json:"trigger_species_name"`
	AffectedSpeciesID   uuid.UUID   `json:"affected_species_id"`
	AffectedSpeciesName string      `json:"affected_species_name"`
	CascadeType         CascadeType `json:"cascade_type"`
	PopulationImpact    float32     `json:"population_impact"` // -1.0 (extinction) to +1.0 (boom)
	Description         string      `json:"description"`       // Lore-friendly description
}

// CascadeResult contains all cascade effects from an extinction
type CascadeResult struct {
	InitialExtinction    uuid.UUID             `json:"initial_extinction"`
	InitialName          string                `json:"initial_name"`
	Year                 int64                 `json:"year"`
	Events               []CascadeEvent        `json:"events"`
	SecondaryExtinctions []uuid.UUID           `json:"secondary_extinctions"`
	PopulationChanges    map[uuid.UUID]float32 `json:"population_changes"` // Species ID -> population multiplier
	TotalAffected        int                   `json:"total_affected"`
	CascadeGeneration    int                   `json:"cascade_generation"` // How many levels deep
}

// CascadeSimulator calculates extinction cascade effects
type CascadeSimulator struct {
	Relationships     []EcologicalRelationship       `json:"relationships"`
	SpeciesRoles      map[uuid.UUID][]EcologicalRole `json:"species_roles"`
	KeystoneSpecies   map[uuid.UUID]float32          `json:"keystone_species"` // Keystone importance (0-1)
	recentExtinctions map[uuid.UUID]int64            // Species ID -> year extinct
}

// NewCascadeSimulator creates a new cascade simulator
func NewCascadeSimulator() *CascadeSimulator {
	return &CascadeSimulator{
		Relationships:     make([]EcologicalRelationship, 0),
		SpeciesRoles:      make(map[uuid.UUID][]EcologicalRole),
		KeystoneSpecies:   make(map[uuid.UUID]float32),
		recentExtinctions: make(map[uuid.UUID]int64),
	}
}

// AddRelationship adds an ecological relationship between species
func (cs *CascadeSimulator) AddRelationship(rel EcologicalRelationship) {
	cs.Relationships = append(cs.Relationships, rel)
}

// SetSpeciesRole sets the ecological role(s) of a species
func (cs *CascadeSimulator) SetSpeciesRole(speciesID uuid.UUID, roles []EcologicalRole) {
	cs.SpeciesRoles[speciesID] = roles
}

// SetKeystoneImportance marks a species as keystone with given importance
func (cs *CascadeSimulator) SetKeystoneImportance(speciesID uuid.UUID, importance float32) {
	cs.KeystoneSpecies[speciesID] = importance

	// Also add keystone role
	roles := cs.SpeciesRoles[speciesID]
	hasKeystone := false
	for _, r := range roles {
		if r == RoleKeystone {
			hasKeystone = true
			break
		}
	}
	if !hasKeystone {
		cs.SpeciesRoles[speciesID] = append(roles, RoleKeystone)
	}
}

// CalculateCascade calculates all cascade effects from a species going extinct
func (cs *CascadeSimulator) CalculateCascade(
	extinctSpeciesID uuid.UUID,
	extinctSpeciesName string,
	year int64,
	maxGenerations int, // Limit cascade depth
) *CascadeResult {
	result := &CascadeResult{
		InitialExtinction:    extinctSpeciesID,
		InitialName:          extinctSpeciesName,
		Year:                 year,
		Events:               make([]CascadeEvent, 0),
		SecondaryExtinctions: make([]uuid.UUID, 0),
		PopulationChanges:    make(map[uuid.UUID]float32),
	}

	cs.recentExtinctions[extinctSpeciesID] = year

	// Track species that will go extinct this cascade
	toProcess := []uuid.UUID{extinctSpeciesID}
	processed := make(map[uuid.UUID]bool)
	generation := 0

	for len(toProcess) > 0 && generation < maxGenerations {
		currentBatch := toProcess
		toProcess = []uuid.UUID{}
		generation++

		for _, speciesID := range currentBatch {
			if processed[speciesID] {
				continue
			}
			processed[speciesID] = true

			// Find all species affected by this one dying
			affected := cs.findAffectedSpecies(speciesID)

			for _, effect := range affected {
				event := CascadeEvent{
					Year:               year,
					TriggerSpeciesID:   speciesID,
					TriggerSpeciesName: extinctSpeciesName, // Would need lookup in real impl
					AffectedSpeciesID:  effect.SpeciesID,
					CascadeType:        effect.CascadeType,
					PopulationImpact:   effect.Impact,
					Description:        effect.Description,
				}
				result.Events = append(result.Events, event)

				// Apply population change
				currentChange, exists := result.PopulationChanges[effect.SpeciesID]
				if !exists {
					currentChange = 1.0
				}
				result.PopulationChanges[effect.SpeciesID] = currentChange * (1 + effect.Impact)

				// Check for secondary extinction (population impact <= -0.9)
				if result.PopulationChanges[effect.SpeciesID] <= 0.1 {
					if !processed[effect.SpeciesID] {
						result.SecondaryExtinctions = append(result.SecondaryExtinctions, effect.SpeciesID)
						toProcess = append(toProcess, effect.SpeciesID)
						cs.recentExtinctions[effect.SpeciesID] = year
					}
				}
			}
		}

		result.CascadeGeneration = generation
	}

	result.TotalAffected = len(result.PopulationChanges)
	return result
}

// affectedInfo holds info about a species affected by extinction cascade
type affectedInfo struct {
	SpeciesID   uuid.UUID
	CascadeType CascadeType
	Impact      float32 // Negative = decline, positive = increase
	Description string
}

// findAffectedSpecies finds species affected by the given species going extinct
func (cs *CascadeSimulator) findAffectedSpecies(extinctID uuid.UUID) []affectedInfo {
	affected := make([]affectedInfo, 0)

	for _, rel := range cs.Relationships {
		// Check if this relationship involves the extinct species
		if rel.TargetSpeciesID == extinctID {
			// The source species depends on the extinct species
			impact, cascadeType, desc := cs.calculateDependencyImpact(rel)
			affected = append(affected, affectedInfo{
				SpeciesID:   rel.SourceSpeciesID,
				CascadeType: cascadeType,
				Impact:      impact,
				Description: desc,
			})
		} else if rel.SourceSpeciesID == extinctID {
			// The extinct species was doing something TO the target
			impact, cascadeType, desc := cs.calculateReleaseImpact(rel)
			affected = append(affected, affectedInfo{
				SpeciesID:   rel.TargetSpeciesID,
				CascadeType: cascadeType,
				Impact:      impact,
				Description: desc,
			})
		}
	}

	// Check if extinct species was keystone
	if importance, isKeystone := cs.KeystoneSpecies[extinctID]; isKeystone {
		// Keystone species affect many others
		affected = append(affected, cs.calculateKeystoneEffects(extinctID, importance)...)
	}

	return affected
}

// calculateDependencyImpact calculates impact when a dependency is lost
func (cs *CascadeSimulator) calculateDependencyImpact(rel EcologicalRelationship) (float32, CascadeType, string) {
	baseImpact := -rel.Strength

	switch rel.Type {
	case RelationshipPredation:
		// Predator loses food source
		if rel.IsObligate {
			return -1.0, CascadeFoodLoss, "obligate food source extinct"
		}
		return baseImpact * 0.7, CascadeFoodLoss, "food source reduced"

	case RelationshipMutualism:
		// Partner loses mutualist
		if rel.IsObligate {
			return -1.0, CascadeCoExtinction, "obligate symbiont extinct"
		}
		return baseImpact * 0.5, CascadeCoExtinction, "mutualist partner lost"

	case RelationshipHabitat:
		// Habitat provider gone
		if rel.IsObligate {
			return -1.0, CascadeHabitatLoss, "habitat destroyed"
		}
		return baseImpact * 0.8, CascadeHabitatLoss, "habitat reduced"

	case RelationshipCommensalism:
		// Commensal loses partner
		return baseImpact * 0.3, CascadeCoExtinction, "lost commensal relationship"

	default:
		return baseImpact * 0.2, CascadeCoExtinction, "ecological partner lost"
	}
}

// calculateReleaseImpact calculates impact when a species is released from pressure
func (cs *CascadeSimulator) calculateReleaseImpact(rel EcologicalRelationship) (float32, CascadeType, string) {
	baseImpact := rel.Strength // Positive - release from pressure

	switch rel.Type {
	case RelationshipPredation:
		// Prey released from predation pressure
		return baseImpact * 0.5, CascadePredatorLoss, "predator release - population surge"

	case RelationshipCompetition:
		// Competitor release
		return baseImpact * 0.4, CascadeCompetitorLoss, "competitive release - expansion into niche"

	case RelationshipParasitism:
		// Parasite gone - minor benefit
		return baseImpact * 0.2, CascadeCompetitorLoss, "parasite removed"

	default:
		return 0, "", ""
	}
}

// calculateKeystoneEffects calculates widespread effects of keystone species loss
func (cs *CascadeSimulator) calculateKeystoneEffects(keystoneID uuid.UUID, importance float32) []affectedInfo {
	effects := make([]affectedInfo, 0)

	// Keystone species affect many unrelated species
	// For each species in the ecosystem, there's a chance of negative impact
	for speciesID := range cs.SpeciesRoles {
		if speciesID == keystoneID {
			continue
		}

		// Check if already directly connected
		directlyConnected := false
		for _, rel := range cs.Relationships {
			if (rel.SourceSpeciesID == speciesID && rel.TargetSpeciesID == keystoneID) ||
				(rel.TargetSpeciesID == speciesID && rel.SourceSpeciesID == keystoneID) {
				directlyConnected = true
				break
			}
		}

		if !directlyConnected {
			// Indirect keystone effect
			impact := -importance * 0.3 // Up to 30% decline per keystone
			effects = append(effects, affectedInfo{
				SpeciesID:   speciesID,
				CascadeType: CascadeKeystone,
				Impact:      impact,
				Description: "keystone species collapse - ecosystem destabilization",
			})
		}
	}

	return effects
}

// IdentifyKeystoneSpecies identifies potential keystone species based on relationships
func (cs *CascadeSimulator) IdentifyKeystoneSpecies() map[uuid.UUID]float32 {
	keystones := make(map[uuid.UUID]float32)

	// Count how many species depend on each species
	dependencyCount := make(map[uuid.UUID]int)
	obligateDependents := make(map[uuid.UUID]int)

	for _, rel := range cs.Relationships {
		dependencyCount[rel.TargetSpeciesID]++
		if rel.IsObligate {
			obligateDependents[rel.TargetSpeciesID]++
		}
	}

	// Species with many dependents, especially obligate ones, are keystone
	for speciesID, count := range dependencyCount {
		importance := float32(count) / 10.0 // Normalize
		if importance > 1.0 {
			importance = 1.0
		}

		// Obligate dependents increase importance significantly
		obligateBonus := float32(obligateDependents[speciesID]) * 0.2
		importance = float32(math.Min(1.0, float64(importance+obligateBonus)))

		if importance >= 0.3 {
			keystones[speciesID] = importance
		}
	}

	// Check for ecosystem engineers and foundation species
	for speciesID, roles := range cs.SpeciesRoles {
		for _, role := range roles {
			if role == RoleEcosystemEngineer || role == RoleFoundationSpecies {
				current, exists := keystones[speciesID]
				if !exists {
					current = 0.4
				}
				keystones[speciesID] = float32(math.Min(1.0, float64(current+0.3)))
			}
		}
	}

	return keystones
}

// GetExtinctionRisk returns how vulnerable a species is to cascade extinction
func (cs *CascadeSimulator) GetExtinctionRisk(speciesID uuid.UUID) float32 {
	risk := float32(0)

	for _, rel := range cs.Relationships {
		if rel.SourceSpeciesID == speciesID {
			// This species depends on another
			if rel.IsObligate {
				risk += 0.3 // High risk from obligate dependency
			} else {
				risk += rel.Strength * 0.1
			}
		}
	}

	// Species with few connections are also at risk (specialists)
	connectionCount := 0
	for _, rel := range cs.Relationships {
		if rel.SourceSpeciesID == speciesID || rel.TargetSpeciesID == speciesID {
			connectionCount++
		}
	}
	if connectionCount <= 2 {
		risk += 0.2 // Specialist penalty
	}

	if risk > 1.0 {
		risk = 1.0
	}
	return risk
}

// GenerateExtinctionCause creates a lore-friendly extinction cause string
func GenerateExtinctionCause(cascadeType CascadeType, triggerName string) string {
	switch cascadeType {
	case CascadeCoExtinction:
		return "extinction of symbiotic partner " + triggerName
	case CascadeFoodLoss:
		return "starvation following loss of prey species " + triggerName
	case CascadePredatorLoss:
		return "population collapse from overpopulation after predator " + triggerName + " died"
	case CascadePollinationLoss:
		return "reproductive failure after pollinator " + triggerName + " vanished"
	case CascadeHabitatLoss:
		return "habitat destruction following extinction of " + triggerName
	case CascadeCompetitorLoss:
		return "ecosystem imbalance after competitor " + triggerName + " vanished"
	case CascadeKeystone:
		return "ecosystem collapse following loss of keystone species " + triggerName
	default:
		return "ecological cascade from extinction of " + triggerName
	}
}
