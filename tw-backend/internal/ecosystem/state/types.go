package state

import (
	"tw-backend/internal/npc/genetics"

	"github.com/google/uuid"
)

// Species represents the biological classification of an entity
type Species string

const (
	// Fauna (Desert)
	SpeciesLizard   Species = "lizard"
	SpeciesScorpion Species = "scorpion"
	SpeciesVulture  Species = "vulture"
	// Fauna (Forest)
	SpeciesDeer Species = "deer"
	SpeciesWolf Species = "wolf"
	SpeciesBear Species = "bear"
	// Fauna (Grassland)
	SpeciesRabbit Species = "rabbit"
	SpeciesHawk   Species = "hawk"
	SpeciesBison  Species = "bison"
	// Flora
	SpeciesCactus Species = "cactus"
	SpeciesFern   Species = "fern"
	SpeciesOak    Species = "oak"
	SpeciesGrass  Species = "grass"
	SpeciesKelp   Species = "kelp"
	// Precambrian (ancient life)
	SpeciesCyanobacteria Species = "cyanobacteria"
	SpeciesStromatolite  Species = "stromatolite"
	SpeciesEdiacaran     Species = "ediacaran"
	SpeciesDickinsonia   Species = "dickinsonia"
	SpeciesCharnia       Species = "charnia"
)

// DietType determines what an entity consumes
type DietType string

const (
	DietHerbivore      DietType = "herbivore"
	DietCarnivore      DietType = "carnivore"
	DietOmnivore       DietType = "omnivore"
	DietPhotosynthetic DietType = "photosynthetic"
)

// LivingEntityState wraps the core logic for a living thing
// This will be stored in Entity.Metadata OR managed separately and linked by ID
type LivingEntityState struct {
	EntityID   uuid.UUID    `json:"entity_id"`
	Species    Species      `json:"species"`
	Diet       DietType     `json:"diet"`
	Age        int64        `json:"age"` // In ticks
	Generation int          `json:"generation"`
	Needs      NeedState    `json:"needs"`
	DNA        genetics.DNA `json:"dna"`

	// Location info for game integration
	WorldID   uuid.UUID `json:"world_id"`
	PositionX float64   `json:"position_x"`
	PositionY float64   `json:"position_y"`

	// Decision History
	Logs []DecisionLog `json:"logs,omitempty"`

	// Lineage Tracking
	Parent1ID *uuid.UUID `json:"parent1_id,omitempty"`
	Parent2ID *uuid.UUID `json:"parent2_id,omitempty"`
}

// DecisionLog records an AI decision
type DecisionLog struct {
	Timestamp int64  `json:"timestamp"` // Unix timestamp or Tick count
	Action    string `json:"action"`    // e.g. "Eat", "Sleep"
	Reason    string `json:"reason"`    // e.g. "Hunger > 50"
}

// NeedState tracks current levels of diverse needs
type NeedState struct {
	Hunger           float64 `json:"hunger"`            // 0-100, 0 is full, 100 is starving
	Thirst           float64 `json:"thirst"`            // 0-100
	Energy           float64 `json:"energy"`            // 0-100, 100 is fully rested
	ReproductionUrge float64 `json:"reproduction_urge"` // 0-100, 100 is desperate
	Safety           float64 `json:"safety"`            // 0-100, 100 is safe
}
