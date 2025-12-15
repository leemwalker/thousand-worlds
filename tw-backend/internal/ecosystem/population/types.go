package population

import (
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// EvolvableTraits represents the genetic characteristics that can evolve over generations.
// These traits influence both macro (population) and micro (individual) simulation.
type EvolvableTraits struct {
	// Physical traits
	Size     float64 `json:"size"`     // 0.1 (mouse) to 10.0 (elephant)
	Speed    float64 `json:"speed"`    // 0.1 to 10.0
	Strength float64 `json:"strength"` // 0.1 to 10.0

	// Behavioral traits
	Aggression   float64 `json:"aggression"`   // 0.0 (docile) to 1.0 (aggressive)
	Social       float64 `json:"social"`       // 0.0 (solitary) to 1.0 (pack animal)
	Intelligence float64 `json:"intelligence"` // 0.0 to 1.0

	// Survival traits
	ColdResistance float64 `json:"cold_resistance"` // 0.0 to 1.0
	HeatResistance float64 `json:"heat_resistance"` // 0.0 to 1.0
	NightVision    float64 `json:"night_vision"`    // 0.0 to 1.0
	Camouflage     float64 `json:"camouflage"`      // 0.0 to 1.0

	// Reproduction traits
	Fertility float64 `json:"fertility"` // Reproduction rate multiplier (0.5 to 2.0)
	Lifespan  float64 `json:"lifespan"`  // Base lifespan in years

	// Dietary traits
	CarnivoreTendency float64 `json:"carnivore_tendency"` // 0.0 (pure herbivore) to 1.0 (pure carnivore)
	VenomPotency      float64 `json:"venom_potency"`      // 0.0 to 1.0
	PoisonResistance  float64 `json:"poison_resistance"`  // 0.0 to 1.0
}

// DietType determines what a species primarily consumes
type DietType string

const (
	DietHerbivore      DietType = "herbivore"
	DietCarnivore      DietType = "carnivore"
	DietOmnivore       DietType = "omnivore"
	DietPhotosynthetic DietType = "photosynthetic"
)

// SpeciesPopulation represents a species population within a biome
type SpeciesPopulation struct {
	SpeciesID     uuid.UUID       `json:"species_id"`
	Name          string          `json:"name"`           // e.g., "Plains Deer", "Mountain Wolf"
	AncestorID    *uuid.UUID      `json:"ancestor_id"`    // Parent species (for lineage tracking)
	Count         int64           `json:"count"`          // Current population
	Traits        EvolvableTraits `json:"traits"`         // Average traits for population
	TraitVariance float64         `json:"trait_variance"` // Genetic diversity (0.0 to 1.0)
	Diet          DietType        `json:"diet"`
	Generation    int64           `json:"generation"`   // Evolutionary generation
	CreatedYear   int64           `json:"created_year"` // Year this species evolved
}

// BiomePopulation tracks all species populations within a biome
type BiomePopulation struct {
	BiomeID          uuid.UUID                        `json:"biome_id"`
	BiomeType        geography.BiomeType              `json:"biome_type"`
	Species          map[uuid.UUID]*SpeciesPopulation `json:"species"`
	CarryingCapacity int64                            `json:"carrying_capacity"` // Max total population
	YearsSimulated   int64                            `json:"years_simulated"`
}

// ExtinctSpecies records a species that has died out
type ExtinctSpecies struct {
	SpeciesID       uuid.UUID       `json:"species_id"`
	Name            string          `json:"name"`
	Traits          EvolvableTraits `json:"traits"`
	Diet            DietType        `json:"diet"`
	PeakPopulation  int64           `json:"peak_population"`
	ExistedFrom     int64           `json:"existed_from"`     // Year species emerged
	ExistedUntil    int64           `json:"existed_until"`    // Year species went extinct
	ExtinctionCause string          `json:"extinction_cause"` // e.g., "predation", "climate", "competition"
	FossilBiomes    []uuid.UUID     `json:"fossil_biomes"`    // Biomes where fossils can be found
}

// FossilRecord stores all extinct species for a world
type FossilRecord struct {
	WorldID uuid.UUID         `json:"world_id"`
	Extinct []*ExtinctSpecies `json:"extinct"`
}

// NewBiomePopulation creates a new biome population tracker
func NewBiomePopulation(biomeID uuid.UUID, biomeType geography.BiomeType) *BiomePopulation {
	// Carrying capacity based on biome type
	capacity := int64(1000)
	switch biomeType {
	case geography.BiomeRainforest:
		capacity = 5000
	case geography.BiomeGrassland:
		capacity = 3000
	case geography.BiomeDesert:
		capacity = 500
	case geography.BiomeTundra:
		capacity = 800
	case geography.BiomeOcean:
		capacity = 10000
	}

	return &BiomePopulation{
		BiomeID:          biomeID,
		BiomeType:        biomeType,
		Species:          make(map[uuid.UUID]*SpeciesPopulation),
		CarryingCapacity: capacity,
	}
}

// TotalPopulation returns the sum of all species populations
func (bp *BiomePopulation) TotalPopulation() int64 {
	var total int64
	for _, sp := range bp.Species {
		total += sp.Count
	}
	return total
}

// AddSpecies adds a new species to the biome
func (bp *BiomePopulation) AddSpecies(species *SpeciesPopulation) {
	bp.Species[species.SpeciesID] = species
}

// RemoveSpecies removes a species (extinction)
func (bp *BiomePopulation) RemoveSpecies(speciesID uuid.UUID) *SpeciesPopulation {
	species := bp.Species[speciesID]
	delete(bp.Species, speciesID)
	return species
}

// DefaultTraitsForDiet returns baseline traits for a diet type
func DefaultTraitsForDiet(diet DietType) EvolvableTraits {
	switch diet {
	case DietHerbivore:
		return EvolvableTraits{
			Size: 2.0, Speed: 5.0, Strength: 2.0,
			Aggression: 0.1, Social: 0.7, Intelligence: 0.3,
			ColdResistance: 0.5, HeatResistance: 0.5, NightVision: 0.3, Camouflage: 0.4,
			Fertility: 1.2, Lifespan: 10,
			CarnivoreTendency: 0.0, VenomPotency: 0.0, PoisonResistance: 0.2,
		}
	case DietCarnivore:
		return EvolvableTraits{
			Size: 3.0, Speed: 6.0, Strength: 5.0,
			Aggression: 0.8, Social: 0.5, Intelligence: 0.6,
			ColdResistance: 0.5, HeatResistance: 0.5, NightVision: 0.6, Camouflage: 0.5,
			Fertility: 0.8, Lifespan: 15,
			CarnivoreTendency: 1.0, VenomPotency: 0.1, PoisonResistance: 0.3,
		}
	case DietOmnivore:
		return EvolvableTraits{
			Size: 2.5, Speed: 4.0, Strength: 3.0,
			Aggression: 0.4, Social: 0.6, Intelligence: 0.7,
			ColdResistance: 0.5, HeatResistance: 0.5, NightVision: 0.4, Camouflage: 0.3,
			Fertility: 1.0, Lifespan: 12,
			CarnivoreTendency: 0.5, VenomPotency: 0.0, PoisonResistance: 0.3,
		}
	case DietPhotosynthetic:
		return EvolvableTraits{
			Size: 1.0, Speed: 0.0, Strength: 0.5,
			Aggression: 0.0, Social: 0.0, Intelligence: 0.0,
			ColdResistance: 0.3, HeatResistance: 0.5, NightVision: 0.0, Camouflage: 0.6,
			Fertility: 2.0, Lifespan: 50,
			CarnivoreTendency: 0.0, VenomPotency: 0.0, PoisonResistance: 0.0,
		}
	default:
		return EvolvableTraits{}
	}
}
