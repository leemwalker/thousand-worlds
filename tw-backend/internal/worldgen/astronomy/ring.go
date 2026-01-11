// Package astronomy provides orbital mechanics and satellite calculations
// for world generation, including natural satellite generation and their
// physical effects on planetary systems.
package astronomy

import (
	"math"

	"github.com/google/uuid"
)

// RingComposition describes the material makeup of a planetary ring
type RingComposition string

const (
	RingCompositionIce   RingComposition = "ice"
	RingCompositionRock  RingComposition = "rock"
	RingCompositionDust  RingComposition = "dust"
	RingCompositionMixed RingComposition = "mixed"
)

// RingFormationStage represents the current stage of ring formation
type RingFormationStage string

const (
	RingStageChunks    RingFormationStage = "chunks"    // Moon just destroyed, large fragments
	RingStageDebris    RingFormationStage = "debris"    // Fragments spreading along orbit
	RingStageSpreading RingFormationStage = "spreading" // Debris flattening into ring plane
	RingStageStable    RingFormationStage = "stable"    // Ring reached equilibrium
)

// PlanetaryRing represents a ring structure orbiting a planet
type PlanetaryRing struct {
	ID           uuid.UUID          `json:"id"`
	InnerRadius  float64            `json:"inner_radius"` // meters from planet center
	OuterRadius  float64            `json:"outer_radius"` // meters from planet center
	Density      float64            `json:"density"`      // kg/m³
	Mass         float64            `json:"mass"`         // total mass in kg
	Composition  RingComposition    `json:"composition"`
	Stage        RingFormationStage `json:"stage"`
	OriginMoonID uuid.UUID          `json:"origin_moon_id"`
	FormedAtYear int64              `json:"formed_at_year"`
	Color        string             `json:"color"` // hex color for visualization
}

// RingSystem holds all rings orbiting a planet
type RingSystem struct {
	Rings []PlanetaryRing `json:"rings"`
}

// RingFormationTimescales in simulation years
const (
	// ChunksToDebrisYears is time for fragments to spread along orbit
	ChunksToDebrisYears = 1

	// DebrisToSpreadingYears is time for debris to begin flattening
	DebrisToSpreadingYears = 10

	// SpreadingToStableYears is time for ring to reach equilibrium
	SpreadingToStableYears = 100
)

// NewRingSystem creates an empty ring system
func NewRingSystem() *RingSystem {
	return &RingSystem{
		Rings: []PlanetaryRing{},
	}
}

// FormRingFromMoonDebris creates a new ring from destroyed moon debris.
// destructionSeverity ranges from 0 (partial) to 1 (complete destruction).
// Returns nil if insufficient debris for ring formation.
func FormRingFromMoonDebris(moon Satellite, destructionSeverity float64, currentYear int64) *PlanetaryRing {
	if destructionSeverity < 0.3 {
		// Not enough destruction for ring formation
		return nil
	}

	// Ring mass is fraction of moon based on destruction severity
	// Typically 1-10% of moon mass ends up in stable ring
	ringMassFraction := destructionSeverity * 0.1
	ringMass := moon.Mass * ringMassFraction

	// Ring forms at moon's orbital distance, spreading inward/outward
	// Inner edge: Roche limit (2.5 × planet radius ≈ 16 million meters for Earth)
	// Outer edge: Moon's original orbit + spread
	rocheLimit := RocheLimitFactor * EarthRadiusMeters

	innerRadius := math.Max(rocheLimit, moon.Distance*0.8)
	outerRadius := moon.Distance * 1.2

	// Determine composition based on moon properties
	composition := RingCompositionMixed
	if moon.Radius < MoonRadiusMeters*0.5 {
		composition = RingCompositionDust
	} else if moon.Mass > MoonMassKg*0.5 {
		composition = RingCompositionIce
	}

	return &PlanetaryRing{
		ID:           uuid.New(),
		InnerRadius:  innerRadius,
		OuterRadius:  outerRadius,
		Density:      500.0, // kg/m³, typical for icy rings
		Mass:         ringMass,
		Composition:  composition,
		Stage:        RingStageChunks,
		OriginMoonID: moon.ID,
		FormedAtYear: currentYear,
		Color:        "#D4C4A8", // Pale tan, like Saturn's rings
	}
}

// AddRing adds a ring to the system
func (rs *RingSystem) AddRing(ring *PlanetaryRing) {
	if ring != nil {
		rs.Rings = append(rs.Rings, *ring)
	}
}

// UpdateRingStages advances ring formation based on elapsed time.
// Rings can progress through multiple stages if enough time has passed.
func (rs *RingSystem) UpdateRingStages(currentYear int64) {
	for i := range rs.Rings {
		ring := &rs.Rings[i]
		age := currentYear - ring.FormedAtYear

		// Progress through all applicable stages based on age
		// This handles time jumps in simulation (e.g., jumping 100 years)
		if age >= SpreadingToStableYears {
			ring.Stage = RingStageStable
		} else if age >= DebrisToSpreadingYears {
			ring.Stage = RingStageSpreading
		} else if age >= ChunksToDebrisYears {
			ring.Stage = RingStageDebris
		}
		// Otherwise remains at RingStageChunks
	}
}

// GetVisibleRings returns rings that have progressed past the chunks stage
func (rs *RingSystem) GetVisibleRings() []PlanetaryRing {
	visible := []PlanetaryRing{}
	for _, ring := range rs.Rings {
		if ring.Stage != RingStageChunks {
			visible = append(visible, ring)
		}
	}
	return visible
}

// TotalRingMass returns combined mass of all rings
func (rs *RingSystem) TotalRingMass() float64 {
	var total float64
	for _, ring := range rs.Rings {
		total += ring.Mass
	}
	return total
}
