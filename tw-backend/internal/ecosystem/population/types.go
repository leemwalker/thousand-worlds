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
	DiseaseResistance float64 `json:"disease_resistance"` // 0.0 to 1.0 (immunity)

	// Appearance traits
	Covering    CoveringType    `json:"covering"`     // Body covering (fur, scales, feathers, etc.)
	FloraGrowth FloraGrowthType `json:"flora_growth"` // Growth type for plants (evergreen, deciduous, etc.)
	Display     float64         `json:"display"`      // Sexual display (0.0-1.0): bright colors, antlers, etc.
}

// DietType determines what a species primarily consumes
type DietType string

const (
	DietHerbivore      DietType = "herbivore"
	DietCarnivore      DietType = "carnivore"
	DietOmnivore       DietType = "omnivore"
	DietPhotosynthetic DietType = "photosynthetic"
)

// TrophicLevel represents position in the food chain
type TrophicLevel int

const (
	TrophicProducer          TrophicLevel = 1 // Flora - converts sunlight to biomass
	TrophicPrimaryConsumer   TrophicLevel = 2 // Herbivores - eat plants
	TrophicSecondaryConsumer TrophicLevel = 3 // Carnivores - eat herbivores
	TrophicApexPredator      TrophicLevel = 4 // Top predators - no natural predators
)

// Season represents the time of year affecting ecology
type Season int

const (
	SeasonSpring Season = 0 // Breeding season, food increasing
	SeasonSummer Season = 1 // Peak food availability, growth
	SeasonFall   Season = 2 // Food decreasing, migration trigger
	SeasonWinter Season = 3 // Low food, harsh conditions
)

// GetSeason returns the current season based on day of year (0-364)
func GetSeason(dayOfYear int) Season {
	day := dayOfYear % 365
	switch {
	case day < 91:
		return SeasonSpring
	case day < 182:
		return SeasonSummer
	case day < 273:
		return SeasonFall
	default:
		return SeasonWinter
	}
}

// GetSeasonFromYear extracts season from simulation year (each year has 4 seasons)
func GetSeasonFromYear(year int64) Season {
	return Season(year % 4)
}

// SeasonalFoodModifier returns food availability multiplier based on season and biome
func SeasonalFoodModifier(season Season, biomeType geography.BiomeType) float64 {
	// Base seasonal effects (temperate baseline)
	baseModifiers := map[Season]float64{
		SeasonSpring: 0.8, // Growing, but not peak
		SeasonSummer: 1.2, // Peak food
		SeasonFall:   1.0, // Harvest/abundance
		SeasonWinter: 0.5, // Scarcity
	}

	modifier := baseModifiers[season]

	// Adjust by biome type
	switch biomeType {
	case geography.BiomeTundra, geography.BiomeTaiga:
		// Extreme seasonal variation in polar regions
		if season == SeasonWinter {
			modifier *= 0.3 // Very harsh winters
		} else if season == SeasonSummer {
			modifier *= 1.3 // Intense but short growing season
		}
	case geography.BiomeDesert:
		// Desert has less seasonal variation, more about wet/dry
		modifier = 0.7 + (modifier-0.5)*0.3 // Compress range
	case geography.BiomeRainforest:
		// Tropical: minimal seasonal variation
		modifier = 0.9 + (modifier-0.5)*0.2 // Nearly constant
	case geography.BiomeOcean:
		// Ocean: moderate seasonal variation
		modifier = 0.8 + (modifier-0.5)*0.4
	case geography.BiomeAlpine:
		// Alpine: extreme, similar to tundra
		if season == SeasonWinter {
			modifier *= 0.4
		}
	}

	return modifier
}

// SeasonalBreedingModifier returns reproduction rate modifier based on season
func SeasonalBreedingModifier(season Season, biomeType geography.BiomeType) float64 {
	// Most species breed in spring/summer
	baseModifiers := map[Season]float64{
		SeasonSpring: 1.5, // Peak breeding
		SeasonSummer: 1.2, // Secondary breeding
		SeasonFall:   0.6, // Reduced
		SeasonWinter: 0.3, // Minimal
	}

	modifier := baseModifiers[season]

	// Tropical biomes have year-round breeding
	if biomeType == geography.BiomeRainforest {
		modifier = 0.9 + (modifier-0.5)*0.2 // Compress to near-constant
	}

	// Ocean species may have different timing
	if biomeType == geography.BiomeOcean {
		// Many marine species spawn in fall
		if season == SeasonFall {
			modifier = 1.3
		}
	}

	return modifier
}

// GetTrophicLevel returns the trophic level for a diet type
func GetTrophicLevel(diet DietType) TrophicLevel {
	switch diet {
	case DietPhotosynthetic:
		return TrophicProducer
	case DietHerbivore:
		return TrophicPrimaryConsumer
	case DietOmnivore:
		return TrophicSecondaryConsumer // Eats both plants and animals
	case DietCarnivore:
		return TrophicSecondaryConsumer
	default:
		return TrophicPrimaryConsumer
	}
}

// EnergyTransferEfficiency returns how much energy transfers up the food chain
// Based on the 10% rule - only ~10% of energy transfers to the next level
func EnergyTransferEfficiency(from, to TrophicLevel) float64 {
	if to <= from {
		return 0 // Can't transfer down or same level
	}
	// Each level up loses ~90% of energy
	levelDiff := int(to - from)
	efficiency := 1.0
	for i := 0; i < levelDiff; i++ {
		efficiency *= 0.10
	}
	return efficiency
}

// CalculateTrophicCapacity returns the maximum sustainable population at a trophic level
// Based on the ecological pyramid - each level can only support ~10% of the biomass below
// foodSupply is the total population/biomass of the level below
func CalculateTrophicCapacity(level TrophicLevel, foodSupply int64) int64 {
	if level == TrophicProducer {
		// Producers are limited by sunlight/nutrients, not lower levels
		return foodSupply // Passed as carrying capacity
	}
	// Each level up in the pyramid supports ~15% of the level below
	// (slightly higher than 10% to account for efficient predators)
	return int64(float64(foodSupply) * 0.15)
}

// SpeciesPopulation represents a species population within a biome
type SpeciesPopulation struct {
	SpeciesID          uuid.UUID       `json:"species_id"`
	Name               string          `json:"name"`                 // e.g., "Plains Deer", "Mountain Wolf"
	AncestorID         *uuid.UUID      `json:"ancestor_id"`          // Parent species (for lineage tracking)
	SymbiosisPartnerID *uuid.UUID      `json:"symbiosis_partner_id"` // Partner species for mutualism
	Count              int64           `json:"count"`                // Current total population
	JuvenileCount      int64           `json:"juvenile_count"`       // Pre-reproductive individuals
	AdultCount         int64           `json:"adult_count"`          // Reproductive adults
	Traits             EvolvableTraits `json:"traits"`               // Average traits for population
	TraitVariance      float64         `json:"trait_variance"`       // Genetic diversity (0.0 to 1.0)
	Diet               DietType        `json:"diet"`
	Generation         int64           `json:"generation"`   // Evolutionary generation
	CreatedYear        int64           `json:"created_year"` // Year this species evolved
}

// BiomePopulation tracks all species populations within a biome
type BiomePopulation struct {
	BiomeID          uuid.UUID                        `json:"biome_id"`
	BiomeType        geography.BiomeType              `json:"biome_type"`
	Species          map[uuid.UUID]*SpeciesPopulation `json:"species"`
	CarryingCapacity int64                            `json:"carrying_capacity"` // Max total population
	Fragmentation    float64                          `json:"fragmentation"`     // 0.0 = connected, 1.0 = isolated patches
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
		capacity = 2000 // Increased from 500 to support initial flora (desert is sparse but large)
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

	// Covering affects biome fitness
	coveringFitness := 0.0

	switch traits.Covering {
	case CoveringFur:
		// Insulation in cold climates
		if biomeType == geography.BiomeTundra || biomeType == geography.BiomeTaiga {
			coveringFitness += 0.2
		}
		if biomeType == geography.BiomeAlpine {
			coveringFitness += 0.15 // Thinner air needs more insulation
		}
		// Overheating in hot/humid climates
		if biomeType == geography.BiomeDesert {
			coveringFitness -= 0.15 // Dry heat worse
		}
		if biomeType == geography.BiomeRainforest {
			coveringFitness -= 0.1 // Humidity problematic
		}
		// Size interaction: large furred animals overheat more (square-cube law)
		if traits.Size > 5.0 && (biomeType == geography.BiomeDesert || biomeType == geography.BiomeRainforest) {
			coveringFitness -= 0.1 * (traits.Size - 5.0) / 5.0
		}

	case CoveringScales:
		// Water retention in arid environments
		if biomeType == geography.BiomeDesert {
			coveringFitness += 0.15
		}
		// Hydrodynamics in aquatic
		if biomeType == geography.BiomeOcean {
			coveringFitness += 0.1
		}
		// Less insulation in extreme cold
		if biomeType == geography.BiomeTundra {
			coveringFitness -= 0.1
		}

	case CoveringFeathers:
		// Best insulation-to-weight ratio
		if biomeType == geography.BiomeAlpine || biomeType == geography.BiomeTaiga {
			coveringFitness += 0.18
		}
		// Water resistance
		if biomeType == geography.BiomeRainforest {
			coveringFitness += 0.05
		}

	case CoveringShell:
		// Mobility penalty in all biomes
		coveringFitness -= 0.05
		// Desiccation resistance
		if biomeType == geography.BiomeDesert {
			coveringFitness += 0.1
		}

	case CoveringSkin:
		// Amphibian-like: needs moisture
		if biomeType == geography.BiomeRainforest || biomeType == geography.BiomeOcean {
			coveringFitness += 0.1
		}
		if biomeType == geography.BiomeDesert {
			coveringFitness -= 0.2 // Dries out
		}
	}

	fitness += coveringFitness

	// Clamp fitness to reasonable range
	if fitness < 0.5 {
		fitness = 0.5
	}
	if fitness > 1.5 {
		fitness = 1.5
	}

	return fitness
}
