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
	Fertility  float64 `json:"fertility"`   // Reproduction rate multiplier (0.5 to 2.0)
	Lifespan   float64 `json:"lifespan"`    // Base lifespan in years
	Maturity   float64 `json:"maturity"`    // Age at sexual maturity in years (0.5 to 20)
	LitterSize float64 `json:"litter_size"` // Average offspring per reproduction (1 to 20)

	// Dietary traits
	CarnivoreTendency float64 `json:"carnivore_tendency"` // 0.0 (pure herbivore) to 1.0 (pure carnivore)
	VenomPotency      float64 `json:"venom_potency"`      // 0.0 to 1.0
	PoisonResistance  float64 `json:"poison_resistance"`  // 0.0 to 1.0

	// Appearance traits
	Covering    CoveringType    `json:"covering"`     // Body covering (fur, scales, feathers, etc.)
	FloraGrowth FloraGrowthType `json:"flora_growth"` // Growth type for plants (evergreen, deciduous, etc.)
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
			Fertility: 1.2, Lifespan: 10, Maturity: 1.0, LitterSize: 2.0,
			CarnivoreTendency: 0.0, VenomPotency: 0.0, PoisonResistance: 0.2,
			Covering: CoveringFur,
		}
	case DietCarnivore:
		return EvolvableTraits{
			Size: 3.0, Speed: 6.0, Strength: 5.0,
			Aggression: 0.8, Social: 0.5, Intelligence: 0.6,
			ColdResistance: 0.5, HeatResistance: 0.5, NightVision: 0.6, Camouflage: 0.5,
			Fertility: 0.8, Lifespan: 15, Maturity: 2.0, LitterSize: 3.0,
			CarnivoreTendency: 1.0, VenomPotency: 0.1, PoisonResistance: 0.3,
			Covering: CoveringFur,
		}
	case DietOmnivore:
		return EvolvableTraits{
			Size: 2.5, Speed: 4.0, Strength: 3.0,
			Aggression: 0.4, Social: 0.6, Intelligence: 0.7,
			ColdResistance: 0.5, HeatResistance: 0.5, NightVision: 0.4, Camouflage: 0.3,
			Fertility: 1.0, Lifespan: 12, Maturity: 1.5, LitterSize: 2.0,
			CarnivoreTendency: 0.5, VenomPotency: 0.0, PoisonResistance: 0.3,
			Covering: CoveringFur,
		}
	case DietPhotosynthetic:
		return EvolvableTraits{
			Size: 1.0, Speed: 0.0, Strength: 0.5,
			Aggression: 0.0, Social: 0.0, Intelligence: 0.0,
			ColdResistance: 0.3, HeatResistance: 0.5, NightVision: 0.0, Camouflage: 0.6,
			Fertility: 2.0, Lifespan: 50, Maturity: 0.5, LitterSize: 10.0,
			CarnivoreTendency: 0.0, VenomPotency: 0.0, PoisonResistance: 0.0,
			Covering: CoveringBark, FloraGrowth: FloraPerennial,
		}
	default:
		return EvolvableTraits{}
	}
}

// CalculateBiomeFitness returns a multiplier (0.5-1.5) based on how well traits match the biome
// Values > 1.0 mean the species is well-adapted, < 1.0 means poorly adapted
func CalculateBiomeFitness(traits EvolvableTraits, biomeType geography.BiomeType) float64 {
	fitness := 1.0

	switch biomeType {
	case geography.BiomeTundra, geography.BiomeAlpine:
		// Cold environments: ColdResistance helps, HeatResistance hurts
		fitness += (traits.ColdResistance - 0.5) * 0.4
		fitness -= (traits.HeatResistance - 0.5) * 0.2
		// Smaller size conserves heat
		if traits.Size < 3.0 {
			fitness += 0.1
		}

	case geography.BiomeDesert:
		// Hot dry environments: HeatResistance helps, ColdResistance hurts
		fitness += (traits.HeatResistance - 0.5) * 0.4
		fitness -= (traits.ColdResistance - 0.5) * 0.2
		// NightVision helps (nocturnal activity)
		fitness += traits.NightVision * 0.15
		// Large size is a disadvantage (water needs)
		if traits.Size > 5.0 {
			fitness -= 0.15
		}

	case geography.BiomeOcean:
		// Ocean: Size and speed help (swimming)
		fitness += traits.Speed * 0.03
		fitness += traits.Size * 0.02
		// Temperature resistance matters less in stable ocean temps
		// But camouflage helps avoid predators
		fitness += traits.Camouflage * 0.1

	case geography.BiomeRainforest:
		// Dense vegetation: Camouflage and intelligence help
		fitness += traits.Camouflage * 0.2
		fitness += traits.Intelligence * 0.15
		// Speed less useful in dense foliage
		if traits.Speed > 7.0 {
			fitness -= 0.1
		}
		// Heat resistance helps
		fitness += (traits.HeatResistance - 0.5) * 0.2

	case geography.BiomeGrassland:
		// Open terrain: Speed and social behavior help
		fitness += traits.Speed * 0.03
		fitness += traits.Social * 0.15
		// Camouflage less effective on open plains
		fitness -= (traits.Camouflage - 0.3) * 0.1

	case geography.BiomeTaiga:
		// Cold forests: Cold resistance, moderate size
		fitness += (traits.ColdResistance - 0.5) * 0.3
		fitness += traits.Camouflage * 0.1

	case geography.BiomeDeciduousForest:
		// Seasonal forest: Adaptability matters
		fitness += traits.Intelligence * 0.1
		fitness += traits.Camouflage * 0.1
	}

	// Clamp fitness to reasonable range
	if fitness < 0.5 {
		fitness = 0.5
	}
	if fitness > 1.5 {
		fitness = 1.5
	}

	return fitness
}
